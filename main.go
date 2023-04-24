// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/kong"
	"github.com/steadybit/extension-kong/utils"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()

	exthealth.SetReady(false)
	exthealth.StartProbes(8085)

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
	action_kit_sdk.InstallSignalHandler()
	exthealth.SetReady(true)
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
				Method: "GET",
				Path:   kong.ServiceAttackEndpoint,
			},
			{
				Method: "GET",
				Path:   kong.RouteAttackEndpoint,
			},
		},
		Discoveries: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   kong.ServiceDiscoveryEndpoint,
			},
			{
				Method: "GET",
				Path:   kong.RouteDiscoveryEndpoint,
			},
		},
		TargetTypes: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   kong.ServiceDiscoveryEndpoint + "/target-description",
			},
			{
				Method: "GET",
				Path:   kong.RouteDiscoveryEndpoint + "/target-description",
			},
		},
		TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
			{
				Method: "GET",
				Path:   "/kong/attribute-descriptions",
			},
		},
	}
}
