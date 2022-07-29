// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"fmt"
	"github.com/kong/go-kong/kong"
	"github.com/mitchellh/mapstructure"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
)

type RequestTerminationState struct {
	Plugins      []*kong.Plugin
	InstanceName string
}

func describeRequestTermination(w http.ResponseWriter, _ *http.Request, _ []byte) {
	writeBody(w, attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_kong.request_termination",
		Label:       "Terminate requests",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures.",
		Version:     "1.1.0",
		Icon:        attack_kit_api.Ptr("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='64' height='64'%3E%3Cpath d='M20.986 50.552h11.662l6.055 7.54-1.04 2.568H22.596l.37-2.568-3.552-5.548zm8.238-33.765 6.33-.01L64 50.428l-2.2 10.23H49.61l.76-2.883-26.58-31.452zM40.518 3.34 53.68 13.758l-1.685 1.75 2.282 3.2v3.422l-6.563 5.386L36.68 14.39h-6.426l2.587-4.774zm-27.46 32.852 9.256-7.935L34.6 42.84l-3.5 5.342H19.782l-7.837 10.144-1.8 2.333H0V48.213l9.465-12.02z' fill='%23003459' fill-rule='evenodd'/%3E%3C/svg%3E"),
		TargetType:  "com.github.steadybit.extension_kong.service",
		Category:    "network",
		TimeControl: "EXTERNAL",
		Parameters: []attack_kit_api.AttackParameter{
			{
				Label:        "Duration",
				Name:         "duration",
				Type:         "duration",
				Advanced:     attack_kit_api.Ptr(false),
				Required:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("30s"),
			},
			{
				Label:       "Consumer Username or ID",
				Name:        "consumer",
				Description: attack_kit_api.Ptr("You may optionally define for which Kong consumer the traffic should be impacted."),
				Type:        "string",
				Advanced:    attack_kit_api.Ptr(false),
				Required:    attack_kit_api.Ptr(false),
			},
			{
				Label:        "Message",
				Name:         "message",
				Type:         "string",
				Advanced:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("Error injected through the Steadybit Kong extension (through the request-termination Kong plugin)"),
			},
			{
				Label:        "Content-Type",
				Name:         "content_type",
				Type:         "string",
				Advanced:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("Content type of the raw response configured with Body."),
			},
			{
				Label:        "Body",
				Name:         "body",
				Type:         "string",
				Advanced:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("The raw response body to send. This is mutually exclusive with the config.message field"),
			},
			{
				Label:        "HTTP status code",
				Name:         "status",
				Type:         "integer",
				Advanced:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("500"),
			},
			{
				Label:       "Trigger",
				Name:        "trigger",
				Type:        "string",
				Description: attack_kit_api.Ptr("When not set, the plugin always activates. When set to a string, the plugin will activate exclusively on requests containing either a header or a query parameter that is named the string."),
				Advanced:    attack_kit_api.Ptr(true),
			},
		},
		Prepare: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/attacks/request-termination/prepare",
		},
		Start: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/attacks/request-termination/start",
		},
		Stop: attack_kit_api.Ptr(attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/attacks/request-termination/stop",
		}),
	})
}

func prepareRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	var request attack_kit_api.PrepareAttackRequestBody
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
	if serviceId == nil {
		writeError(w, "Missing target attribute 'kong.service.id'", nil)
		return
	}

	service, err := instance.FindService(serviceId)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to find service '%s' within Kong", *serviceId), err)
		return
	}

	var consumer *kong.Consumer = nil
	if request.Config["consumer"] != nil {
		configuredConsumer := request.Config["consumer"].(string)
		if len(configuredConsumer) > 0 {
			consumer, err = instance.FindConsumer(&configuredConsumer)
			if err != nil {
				writeError(w, fmt.Sprintf("Failed to find consumer '%s' within Kong", configuredConsumer), err)
				return
			}
		}
	}

	if request.Config["message"] != nil && request.Config["body"] != nil {
		writeError(w, "You can't have a message and a body, please choose your return", err)
		return
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
			"status_code":  request.Config["status"].(float64),
			"message":      request.Config["message"].(string),
			"content_type": request.Config["content_type"].(string),
			"body":         request.Config["body"].(string),
			"trigger":      request.Config["trigger"].(string),
		},
	})
	if err != nil {
		writeError(w, "Failed to create plugin", err)
		return
	}

	err, encodedState := encodeAttackState(RequestTerminationState{
		InstanceName: instance.Name,
		Plugins: []*kong.Plugin{
			plugin,
		},
	})
	if err != nil {
		writeError(w, "Failed to encode attack state", err)
		return
	}

	writeBody(w, attack_kit_api.AttackStateAndMessages{
		State: encodedState,
	})
}

func decodeAttackState(attackState attack_kit_api.AttackState) (error, RequestTerminationState) {
	var result RequestTerminationState
	err := mapstructure.Decode(attackState, &result)
	return err, result
}

func encodeAttackState(attackState RequestTerminationState) (error, attack_kit_api.AttackState) {
	var result attack_kit_api.AttackState
	err := mapstructure.Decode(attackState, &result)
	return err, result
}

func startRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	var startAttackRequest attack_kit_api.StartAttackRequestBody
	err := json.Unmarshal(body, &startAttackRequest)
	if err != nil {
		writeError(w, "Failed to read request body", err)
		return
	}

	err, state := decodeAttackState(startAttackRequest.State)
	if err != nil {
		writeError(w, "Failed to decode attack state", err)
		return
	}

	instance, err := FindInstanceByName(state.InstanceName)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err)
		return
	}

	updatedPlugins := make([]*kong.Plugin, len(state.Plugins))
	for i, plugin := range state.Plugins {
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

	err, outputState := encodeAttackState(RequestTerminationState{
		InstanceName: instance.Name,
		Plugins:      updatedPlugins,
	})
	if err != nil {
		writeError(w, "Failed to encode attack state", err)
		return
	}

	writeBody(w, attack_kit_api.AttackStateAndMessages{
		State: outputState,
	})
}

func stopRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	w.Header().Set("Content-Type", "application/json")

	var stopAttackRequest attack_kit_api.StopAttackRequestBody
	err := json.Unmarshal(body, &stopAttackRequest)
	if err != nil {
		writeError(w, "Failed to read request body", err)
		return
	}

	err, state := decodeAttackState(stopAttackRequest.State)
	if err != nil {
		writeError(w, "Failed to decode attack state", err)
		return
	}

	instance, err := FindInstanceByName(state.InstanceName)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err)
		return
	}

	for _, plugin := range state.Plugins {
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
