// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/kong"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)

	exthealth.SetReady(false)
	exthealth.StartProbes(8085)

  // Most extensions require some form of configuration. These calls exist to parse and validate the
  // configuration obtained from environment variables.
  config.ParseConfiguration()
  config.ValidateConfiguration()

	kong.RegisterAttributeDescriptionHandlers()
	kong.RegisterServiceDiscoveryHandlers()
	action_kit_sdk.RegisterAction(kong.NewServiceRequestTerminationAction())
	kong.RegisterRouteDiscoveryHandlers()
	action_kit_sdk.RegisterAction(kong.NewRequestTerminationAction())

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

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8084,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList
	discovery_kit_api.DiscoveryList
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList: action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_api.DiscoveryList{
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
		},
	}
}
