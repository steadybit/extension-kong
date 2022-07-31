// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/services"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	http.Handle("/", panicRecovery(logRequest(rootHandler)))

	services.RegisterServiceDiscoveryHandlers()
	services.RegisterServiceAttackHandlers()

	port := 8084
	log.Info().Msgf("Starting Kong extension server on port %d. Get started via /\n", port)
	log.Info().Msgf("Starting with configuration:\n")
	for _, instance := range config.Instances {
		if instance.IsAuthenticated() {
			log.Info().Msgf("  %s: %s (authenticated with %s header)", instance.Name, instance.BaseUrl, instance.HeaderKey)
		} else {
			log.Info().Msgf("  %s: %s", instance.Name, instance.BaseUrl)
		}
	}
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type ExtensionListResponse struct {
	Discoveries      []discovery_kit_api.DescribingEndpointReference `json:"discoveries"`
	TargetTypes      []discovery_kit_api.DescribingEndpointReference `json:"targetTypes"`
	TargetAttributes []discovery_kit_api.DescribingEndpointReference `json:"targetAttributes"`
	Attacks          []attack_kit_api.DescribingEndpointReference    `json:"attacks"`
}

func rootHandler(w http.ResponseWriter, request *http.Request, _ []byte) {
	if request.URL.Path != "/" {
		w.WriteHeader(404)
		return
	}

	writeBody(w, ExtensionListResponse{
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
	})
}

func panicRecovery(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("Panic: %v\n %s", err, string(debug.Stack()))
				writeError(w, "Internal Server Error", nil)
			}
		}()
		next(w, r)
	}
}

func logRequest(next func(w http.ResponseWriter, r *http.Request, body []byte)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, bodyReadErr := ioutil.ReadAll(r.Body)
		if bodyReadErr != nil {
			http.Error(w, bodyReadErr.Error(), http.StatusBadRequest)
			return
		}

		if len(body) > 0 {
			log.Info().Msgf("%s %s with body %s", r.Method, r.URL, body)
		} else {
			log.Info().Msgf("%s %s", r.Method, r.URL)
		}

		next(w, r, body)
	}
}

func writeError(w http.ResponseWriter, title string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	var response discovery_kit_api.DiscoveryKitError
	if err != nil {
		response = discovery_kit_api.DiscoveryKitError{Title: title, Detail: discovery_kit_api.Ptr(err.Error())}
	} else {
		response = discovery_kit_api.DiscoveryKitError{Title: title}
	}
	json.NewEncoder(w).Encode(response)
}

func writeBody(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}
