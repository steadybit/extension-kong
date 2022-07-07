// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"fmt"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
)

type RequestTerminationConfig struct {
	Message  string `json:"message"`
	Status   int    `json:"status"`
	Consumer string `json:"consumer"`
}

type RequestTerminationState struct {
	Plugins  []*kong.Plugin `json:"plugins"`
	Instance Instance       `json:"instance"`
}

func describeRequestTermination(w http.ResponseWriter, _ *http.Request, _ []byte) {
	writeBody(w, DescribeAttackResponse{
		Id:          "com.github.steadybit.extension_kong.request_termination",
		Label:       "Terminate requests in Kong",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures.",
		Version:     "1.0.0",
		Target:      "com.github.steadybit.extension_kong.service",
		Category:    "network",
		TimeControl: "EXTERNAL",
		Parameters: []AttackParameter{
			{
				Label:    "Duration",
				Name:     "duration",
				Type:     "duration",
				Advanced: false,
				Required: true,
			},
			{
				Label:       "Consumer Username or ID",
				Name:        "consumer",
				Description: "You may optionally define for which Kong consumer the traffic should be impacted.",
				Type:        "string",
				Advanced:    false,
				Required:    false,
			},
			{
				Label:        "Message",
				Name:         "message",
				Type:         "string",
				Advanced:     true,
				DefaultValue: "Error injected through the Steadybit Kong extension (through the request-termination Kong plugin)",
			},
			{
				Label:        "HTTP status code",
				Name:         "status",
				Type:         "integer",
				Advanced:     true,
				DefaultValue: "500",
			},
		},
		Prepare: EndpointRef{
			"POST",
			"/attacks/request-termination/prepare",
		},
		Start: EndpointRef{
			"POST",
			"/attacks/request-termination/start",
		},
		Stop: EndpointRef{
			"POST",
			"/attacks/request-termination/stop",
		},
	})
}

func prepareRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request PrepareAttackRequest[RequestTerminationConfig]
	err := json.Unmarshal(body, &request)
	if err != nil {
		writeError(w, "Failed to read request body", err)
		return
	}

	instanceName := findFirstValue(request.Target.Attributes, "kong.instance.name")
	if instanceName == nil {
		writeError(w, "Missing target attribute 'kong.instance.name'", nil)
		return
	}

	instance, err := FindInstanceByName(*instanceName)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to find a configured instance named '%s'", instanceName), err)
		return
	}

	serviceId := findFirstValue(request.Target.Attributes, "kong.service.id")
	if instanceName == nil {
		writeError(w, "Missing target attribute 'kong.service.id'", nil)
		return
	}

	service, err := instance.FindService(serviceId)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to find service '%s' within Kong", *serviceId), err)
		return
	}

	var consumer *kong.Consumer = nil
	if len(request.Config.Consumer) > 0 {
		consumer, err = instance.FindConsumer(&request.Config.Consumer)
		if err != nil {
			writeError(w, fmt.Sprintf("Failed to find consumer '%s' within Kong", request.Config.Consumer), err)
			return
		}
	}

	plugin, err := instance.CreatePlugin(&kong.Plugin{
		Name:    utils.String("request-termination"),
		Enabled: utils.Bool(false),
		Tags: utils.Strings([]string{
			"created-by=steadybit",
		}),
		Service:  service,
		Consumer: consumer,
		Config: kong.Configuration{
			"status_code": request.Config.Status,
			"message":     request.Config.Message,
		},
	})
	if err != nil {
		writeError(w, "Failed to create plugin", err)
		return
	}

	writeBody(w, PrepareAttackResponse[RequestTerminationState]{
		State: RequestTerminationState{
			Instance: *instance,
			Plugins: []*kong.Plugin{
				plugin,
			},
		},
	})
}

func startRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	var startAttackRequest StartAttackRequest[RequestTerminationState]
	err := json.Unmarshal(body, &startAttackRequest)
	if err != nil {
		writeError(w, "Failed to read request body", err)
		return
	}

	instance := startAttackRequest.State.Instance
	updatedPlugins := make([]*kong.Plugin, len(startAttackRequest.State.Plugins))
	for i, plugin := range startAttackRequest.State.Plugins {
		updatedPlugin, err := instance.UpdatePlugin(&kong.Plugin{
			ID:      plugin.ID,
			Enabled: utils.Bool(true),
		})
		updatedPlugins[i] = updatedPlugin
		if err != nil {
			writeError(w, fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s'", *plugin.ID), err)
			return
		}
	}

	writeBody(w, StartAttackResponse[RequestTerminationState]{
		State: RequestTerminationState{
			Instance: instance,
			Plugins:  updatedPlugins,
		},
	})
}

func stopRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	w.Header().Set("Content-Type", "application/json")

	var stopAttackRequest StopAttackRequest[RequestTerminationState]
	err := json.Unmarshal(body, &stopAttackRequest)
	if err != nil {
		writeError(w, "Failed to read request body", err)
		return
	}

	instance := stopAttackRequest.State.Instance
	for _, plugin := range stopAttackRequest.State.Plugins {
		err := instance.DeletePlugin(plugin.ID)
		if err != nil {
			writeError(w, fmt.Sprintf("Failed to delete plugin within Kong for plugin ID '%s'", *plugin.ID), err)
			return
		}
	}
}

func findFirstValue(attributes map[string][]string, key string) *string {
	if attributes == nil {
		return nil
	}
	if len(attributes[key]) == 0 {
		return nil
	}
	return &attributes[key][0]
}
