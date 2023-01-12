// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/kong"
	"github.com/steadybit/extension-kong/utils"
)

func main() {
	extlogging.InitZeroLog()

	utils.RegisterHttpHandler("/", utils.GetterAsHandler(getExtensionDescription))

	kong.RegisterAttributeDescriptionHandlers()
	kong.RegisterServiceDiscoveryHandlers()
	kong.RegisterRouteDiscoveryHandlers()
	kong.RegisterServiceAttackHandlers()
	kong.RegisterRouteAttackHandlers()

	log.Log().Msgf("Starting with configuration:")
	for _, instance := range config.Instances {
		if instance.IsAuthenticated() {
			log.Log().Msgf("  %s: %s (authenticated with %s header)", instance.Name, instance.BaseUrl, instance.HeaderKey)
		} else {
			log.Log().Msgf("  %s: %s", instance.Name, instance.BaseUrl)
		}
	}
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8084,
	})
}

type ExtensionListResponse struct {
	Discoveries      []discovery_kit_api.DescribingEndpointReference `json:"discoveries"`
	TargetTypes      []discovery_kit_api.DescribingEndpointReference `json:"targetTypes"`
	TargetAttributes []discovery_kit_api.DescribingEndpointReference `json:"targetAttributes"`
	Attacks          []attack_kit_api.DescribingEndpointReference    `json:"attacks"`
}

func getExtensionDescription() ExtensionListResponse {
	return ExtensionListResponse{
		Attacks: []attack_kit_api.DescribingEndpointReference{
			{
				"GET",
				kong.ServiceAttackEndpoint,
			},
			{
				"GET",
				kong.RouteAttackEndpoint,
			},
		},
		Discoveries: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				kong.ServiceDiscoveryEndpoint,
			},
			{
				"GET",
				kong.RouteDiscoveryEndpoint,
			},
		},
		TargetTypes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				kong.ServiceDiscoveryEndpoint + "/target-description",
			},
			{
				"GET",
				kong.RouteDiscoveryEndpoint + "/target-description",
			},
		},
		TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/kong/attribute-descriptions",
			},
		},
	}
}
