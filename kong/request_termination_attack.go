// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"fmt"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/utils"
)

type RequestTerminationAction struct {
}

type RequestTerminationState struct {
	PluginIds    []string
	InstanceName string
	ServiceId    string
	RouteId      string
}

func NewRequestTerminationAction() action_kit_sdk.Action[RequestTerminationState] {
	return RequestTerminationAction{}
}

var _ action_kit_sdk.Action[RequestTerminationState] = (*RequestTerminationAction)(nil)
var _ action_kit_sdk.ActionWithStop[RequestTerminationState] = (*RequestTerminationAction)(nil)

func (f RequestTerminationAction) NewEmptyState() RequestTerminationState {
	return RequestTerminationState{}
}

func (f RequestTerminationAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          "com.github.steadybit.extension_kong.routes.request_termination",
		Label:       "Terminate requests",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures for specific Kong routes.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(RouteIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: RouteTargetID,
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label:       "by route-id",
					Description: extutil.Ptr("Find route by id"),
					Query:       "kong.route.id=\"\"",
				},
				{
					Label:       "by route-name",
					Description: extutil.Ptr("Find route by name"),
					Query:       "kong.route.name=\"\"",
				},
			}),
		}),
		Category:    extutil.Ptr("network"),
		TimeControl: action_kit_api.External,
		Kind:        action_kit_api.Attack,
		Parameters: []action_kit_api.ActionParameter{
			{
				Label:        "Duration",
				Name:         "duration",
				Type:         "duration",
				Advanced:     extutil.Ptr(false),
				Required:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("30s"),
			},
			{
				Label:       "Consumer Username or ID",
				Name:        "consumer",
				Description: extutil.Ptr("You may optionally define for which Kong consumer the traffic should be impacted."),
				Type:        "string",
				Advanced:    extutil.Ptr(false),
				Required:    extutil.Ptr(false),
			},
			{
				Label:        "Message",
				Name:         "message",
				Type:         "string",
				Advanced:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("Error injected through the Steadybit Kong extension (through the request-termination Kong plugin)"),
			},
			{
				Label:       "Content-Type",
				Name:        "contentType",
				Description: extutil.Ptr("Content-Type response header to be returned for terminated requests."),
				Type:        "string",
				Advanced:    extutil.Ptr(true),
			},
			{
				Label:       "Body",
				Name:        "body",
				Description: extutil.Ptr("The raw response body to be returned for terminated requests. This is mutually exclusive with the message parameter. A body parameter takes precedence over the message parameter."),
				Type:        "string",
				Advanced:    extutil.Ptr(true),
			},
			{
				Label:        "HTTP status code",
				Name:         "status",
				Type:         "integer",
				Advanced:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("500"),
			},
			{
				Label:       "Trigger",
				Name:        "trigger",
				Type:        "string",
				Description: extutil.Ptr("When not set, the plugin always activates. When set to a string, the plugin will activate exclusively on requests containing either a header or a query parameter that is named the string."),
				Advanced:    extutil.Ptr(true),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f RequestTerminationAction) Prepare(_ context.Context, state *RequestTerminationState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	instanceName := findFirstValue(request.Target.Attributes, "kong.instance.name")
	if instanceName == nil {
		return nil, extension_kit.ToError("Missing target attribute 'kong.instance.name'", nil)
	}

	instance, err := config.FindInstanceByName(*instanceName)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", *instanceName), err)
	}

	requestedServiceId := findFirstValue(request.Target.Attributes, "kong.service.id")
	requestedRouteId := findFirstValue(request.Target.Attributes, "kong.route.id")
	if requestedServiceId == nil {
		return nil, extension_kit.ToError("Missing target attribute 'kong.service.id' required.", nil)
	}

	service, err := instance.FindService(requestedServiceId)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to find service '%s' within Kong", *requestedServiceId), err)
	}

	var route *kong.Route
	if requestedRouteId != nil {
		route, err = instance.FindRoute(service, requestedRouteId)
		if err != nil {
			return nil, extension_kit.ToError(fmt.Sprintf("Failed to find route '%s' within Kong", *requestedRouteId), err)
		}
	}

	var consumer *kong.Consumer = nil
	if request.Config["consumer"] != nil {
		configuredConsumer := request.Config["consumer"].(string)
		if len(configuredConsumer) > 0 {
			consumer, err = instance.FindConsumer(&configuredConsumer)
			if err != nil {
				return nil, extension_kit.ToError(fmt.Sprintf("Failed to find consumer '%s' within Kong", configuredConsumer), err)
			}
		}
	}

	kongConfig := kong.Configuration{
		"status_code": request.Config["status"].(int),
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
		Name:    extutil.Ptr("request-termination"),
		Enabled: extutil.Ptr(false),
		Tags: utils.Strings([]string{
			"created-by=steadybit",
		}),
		Service:  service,
		Route:    route,
		Consumer: consumer,
		Config:   kongConfig,
	})
	if err != nil {
		return nil, extension_kit.ToError("Failed to create plugin", err)
	}

	var serviceId string
	if service != nil {
		serviceId = *service.ID
	}
	var routeId string
	if route != nil {
		routeId = *route.ID
	}

	state.InstanceName = instance.Name
	state.ServiceId = serviceId
	state.RouteId = routeId
	state.PluginIds = []string{*plugin.ID}

	return nil, nil
}

func (f RequestTerminationAction) Start(_ context.Context, state *RequestTerminationState) (*action_kit_api.StartResult, error) {
	instance, err := config.FindInstanceByName(state.InstanceName)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err)
	}

	for _, pluginId := range state.PluginIds {
		// try to update first at route level
		if state.RouteId != "" {
			_, err = instance.UpdatePluginForRoute(&state.RouteId, &kong.Plugin{
				ID:      &pluginId,
				Enabled: extutil.Ptr(true),
			})
			if err != nil {
				return nil, extension_kit.ToError(fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s' at route level", pluginId), err)
			}
		} else if state.ServiceId != "" {
			_, err = instance.UpdatePluginForService(&state.ServiceId, &kong.Plugin{
				ID:      &pluginId,
				Enabled: extutil.Ptr(true),
			})
			if err != nil {
				return nil, extension_kit.ToError(fmt.Sprintf("Failed to enable plugin within Kong for plugin ID '%s' at service level", pluginId), err)
			}
		}
	}
	return nil, nil
}

func (f RequestTerminationAction) Stop(_ context.Context, state *RequestTerminationState) (*action_kit_api.StopResult, error) {
	instance, err := config.FindInstanceByName(state.InstanceName)
	if err != nil {
		return nil, extension_kit.ToError(fmt.Sprintf("Failed to find a configured instance named '%s'", state.InstanceName), err)
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
			return nil, extension_kit.ToError(fmt.Sprintf("Failed to delete plugin within Kong for plugin ID '%s' at %s level", pluginId, level), err)
		}
	}

	return nil, nil
}

func isDefinedString(v interface{}) bool {
	return v != nil && len(v.(string)) > 0
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
