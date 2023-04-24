// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

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

type RequestTerminationState struct {
	PluginIds    []string
	InstanceName string
	ServiceId    string
	RouteId      string
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

	requestedServiceId := findFirstValue(request.Target.Attributes, "kong.service.id")
	requestedRouteId := findFirstValue(request.Target.Attributes, "kong.route.id")
	if requestedServiceId == nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Missing target attribute 'kong.service.id' required.", nil))
	}

	service, err := instance.FindService(requestedServiceId)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find service '%s' within Kong", *requestedServiceId), err))
	}

	var route *kong.Route
	if requestedRouteId != nil {
		route, err = instance.FindRoute(service, requestedRouteId)
		if err != nil {
			return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find route '%s' within Kong", *requestedRouteId), err))
		}
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

	kongConfig := kong.Configuration{
		"status_code": request.Config["status"].(float64),
	}

	if isDefinedString(request.Config["body"]) {
		kongConfig["body"] = request.Config["body"]
	} else if isDefinedString(request.Config["message"]) {
		kongConfig["message"] = request.Config["message"]
	}

	if isDefinedString(request.Config["contentType"]) {
		kongConfig["content_type"] = request.Config["contentType"]
	}

	if isDefinedString(request.Config["trigger"]) {
		kongConfig["trigger"] = request.Config["trigger"]
	}

	plugin, err := instance.CreatePluginAtAnyLevel(&kong.Plugin{
		Name:    utils.String("request-termination"),
		Enabled: utils.Bool(false),
		Tags: utils.Strings([]string{
			"created-by=steadybit",
		}),
		Service:  service,
		Route:    route,
		Consumer: consumer,
		Config:   kongConfig,
	})
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to create plugin", err))
	}

	var serviceId string
	if service != nil {
		serviceId = *service.ID
	}
	var routeId string
	if route != nil {
		routeId = *route.ID
	}

	return attack_kit_api.Ptr(RequestTerminationState{
		InstanceName: instance.Name,
		ServiceId:    serviceId,
		RouteId:      routeId,
		PluginIds:    []string{*plugin.ID},
	}), nil
}

func isDefinedString(v interface{}) bool {
	return v != nil && len(v.(string)) > 0
}

func decodeAttackState(attackState attack_kit_api.AttackState) (RequestTerminationState, error) {
	var result RequestTerminationState
	err := mapstructure.Decode(attackState, &result)
	return result, err
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
		// try to update first at route level
		if state.RouteId != "" {
			_, err = instance.UpdatePluginForRoute(&state.RouteId, &kong.Plugin{
				ID:      &pluginId,
				Enabled: utils.Bool(true),
			})
			if err != nil {
				return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s' at route level", pluginId), err))
			}
		} else if state.ServiceId != "" {
			_, err = instance.UpdatePluginForService(&state.ServiceId, &kong.Plugin{
				ID:      &pluginId,
				Enabled: utils.Bool(true),
			})
			if err != nil {
				return nil, attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s' at service level", pluginId), err))
			}
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

	state, err := decodeAttackState(stopAttackRequest.State)
	if err != nil {
		return attack_kit_api.Ptr(utils.ToError("Failed to decode attack state", err))
	}

	instance, err := config.FindInstanceByName(state.InstanceName)
	if err != nil {
		return attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err))
	}

	for _, pluginId := range state.PluginIds {
		level := "service"
		if state.RouteId != "" {
			err = instance.DeletePluginForRoute(&state.RouteId, &pluginId)
			level = "route"
		} else if state.ServiceId != "" {
			err = instance.DeletePluginForService(&state.ServiceId, &pluginId)
		}
		if err != nil {
			return attack_kit_api.Ptr(utils.ToError(fmt.Sprintf("Failed to delete plugin within Kong for plugin ID '%s' at %s level", pluginId, level), err))
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
