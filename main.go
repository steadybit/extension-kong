// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/services"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	utils.RegisterHttpHandler("/", utils.GetterAsHandler(getExtensionDescription))

	services.RegisterServiceDiscoveryHandlers()
	services.RegisterServiceAttackHandlers()

	port := 8084
	log.Info().Msgf("Starting Test extension server on port %d. Get started via /", port)
	log.Info().Msgf("Starting with configuration:")
	for _, instance := range config.Instances {
		if instance.IsAuthenticated() {
			log.Info().Msgf("  %s: %s (authenticated with %s header)", instance.Name, instance.BaseUrl, instance.HeaderKey)
		} else {
			log.Info().Msgf("  %s: %s", instance.Name, instance.BaseUrl)
		}
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	log.Fatal().Err(err).Msgf("Failed to start HTTP server")
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
				"/service/attack/request-termination",
			},
		},
		Discoveries: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/service/discovery",
			},
		},
		TargetTypes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/service/discovery/target-description",
			},
		},
		TargetAttributes: []discovery_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/service/discovery/attribute-descriptions",
			},
		},
	}
}
