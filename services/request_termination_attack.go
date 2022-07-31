// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package services

import (
	"encoding/json"
	"fmt"
	"github.com/kong/go-kong/kong"
	"github.com/mitchellh/mapstructure"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
)

func RegisterServiceAttackHandlers() {
	utils.RegisterHttpHandler("/service/attack/request-termination", utils.GetterAsHandler(getServiceRequestTerminationAttackDescription))
	utils.RegisterHttpHandler("/service/attack/request-termination/prepare", prepareRequestTermination)
	utils.RegisterHttpHandler("/service/attack/request-termination/start", startRequestTermination)
	utils.RegisterHttpHandler("/service/attack/request-termination/stop", stopRequestTermination)
}

func getServiceRequestTerminationAttackDescription() attack_kit_api.AttackDescription {
	return attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_kong.request_termination",
		Label:       "Terminate requests",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures.",
		Version:     "1.1.0",
		Icon:        attack_kit_api.Ptr(serviceIcon),
		TargetType:  serviceTargetId,
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
				Label:        "HTTP status code",
				Name:         "status",
				Type:         "integer",
				Advanced:     attack_kit_api.Ptr(true),
				DefaultValue: attack_kit_api.Ptr("500"),
			},
		},
		Prepare: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/service/attack/request-termination/prepare",
		},
		Start: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/service/attack/request-termination/start",
		},
		Stop: attack_kit_api.Ptr(attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/service/attack/request-termination/stop",
		}),
	}
}

type RequestTerminationState struct {
	PluginIds    []string
	InstanceName string
}

func prepareRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	state, err := PrepareRequestTermination(body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func PrepareRequestTermination(body []byte) (*RequestTerminationState, *attack_kit_api.AttackKitError) {
	var request attack_kit_api.PrepareAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	instanceName := findFirstValue(request.Target.Attributes, "kong.instance.name")
	if instanceName == nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Missing target attribute 'kong.instance.name'", nil))
	}

	instance, err := config.FindInstanceByName(*instanceName)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", *instanceName), err))
	}

	serviceId := findFirstValue(request.Target.Attributes, "kong.service.id")
	if serviceId == nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Missing target attribute 'kong.service.id'", nil))
	}

	service, err := instance.FindService(serviceId)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find service '%s' within Kong", *serviceId), err))
	}

	var consumer *kong.Consumer = nil
	if request.Config["consumer"] != nil {
		configuredConsumer := request.Config["consumer"].(string)
		if len(configuredConsumer) > 0 {
			consumer, err = instance.FindConsumer(&configuredConsumer)
			if err != nil {
				return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find consumer '%s' within Kong", configuredConsumer), err))
			}
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
			"status_code": request.Config["status"].(float64),
			"message":     request.Config["message"].(string),
		},
	})
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to create plugin", err))
	}

	return attack_kit_api.Ptr(RequestTerminationState{
		InstanceName: instance.Name,
		PluginIds:    []string{*plugin.ID},
	}), nil
}

func decodeAttackState(attackState attack_kit_api.AttackState) (error, RequestTerminationState) {
	var result RequestTerminationState
	err := mapstructure.Decode(attackState, &result)
	return err, result
}

func startRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	state, err := StartRequestTermination(body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func StartRequestTermination(body []byte) (*RequestTerminationState, *attack_kit_api.AttackKitError) {
	var request attack_kit_api.StartAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	var state RequestTerminationState
	err = utils.DecodeAttackState(request.State, &state)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse attack state", err))
	}

	instance, err := config.FindInstanceByName(state.InstanceName)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err))
	}

	for _, pluginId := range state.PluginIds {
		_, err := instance.UpdatePlugin(&kong.Plugin{
			ID:      &pluginId,
			Enabled: utils.Bool(true),
		})
		if err != nil {
			return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s'", pluginId), err))
		}
	}

	return &state, nil
}

func stopRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	err := StopRequestTermination(body)
	if err != nil {
		utils.WriteError(w, *err)
	}
}

func StopRequestTermination(body []byte) *attack_kit_api.AttackKitError {
	var stopAttackRequest attack_kit_api.StopAttackRequestBody
	err := json.Unmarshal(body, &stopAttackRequest)
	if err != nil {
		return attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	err, state := decodeAttackState(stopAttackRequest.State)
	if err != nil {
		return attack_kit_api.Ptr(utils.ToError("Failed to decode attack state", err))
	}

	instance, err := config.FindInstanceByName(state.InstanceName)
	if err != nil {
		return attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err))
	}

	for _, pluginId := range state.PluginIds {
		err := instance.DeletePlugin(&pluginId)
		if err != nil {
			return attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to delete plugin within Kong for plugin ID '%s'", pluginId), err))
		}
	}

	return nil
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
