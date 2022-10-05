// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kong/utils"
)

const (
	RouteAttackEndpoint = "/kong/route/attack/request-termination"
)

func RegisterRouteAttackHandlers() {
	utils.RegisterHttpHandler(RouteAttackEndpoint, utils.GetterAsHandler(getRouteRequestTerminationAttackDescription))
	utils.RegisterHttpHandler(RouteAttackEndpoint+"/prepare", prepareRequestTermination)
	utils.RegisterHttpHandler(RouteAttackEndpoint+"/start", startRequestTermination)
	utils.RegisterHttpHandler(RouteAttackEndpoint+"/stop", stopRequestTermination)
}

func getRouteRequestTerminationAttackDescription() attack_kit_api.AttackDescription {
	return attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_kong.routes.request_termination",
		Label:       "Terminate requests",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures for specific Kong routes.",
		Version:     "1.1.1",
		Icon:        attack_kit_api.Ptr(RouteIcon),
		TargetType:  RouteTargetID,
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
				Label:       "Content-Type",
				Name:        "contentType",
				Description: attack_kit_api.Ptr("Content-Type response header to be returned for terminated requests."),
				Type:        "string",
				Advanced:    attack_kit_api.Ptr(true),
			},
			{
				Label:       "Body",
				Name:        "body",
				Description: attack_kit_api.Ptr("The raw response body to be returned for terminated requests. This is mutually exclusive with the message parameter. A body parameter takes precedence over the message parameter."),
				Type:        "string",
				Advanced:    attack_kit_api.Ptr(true),
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
			Path:   RouteAttackEndpoint + "/prepare",
		},
		Start: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   RouteAttackEndpoint + "/start",
		},
		Stop: attack_kit_api.Ptr(attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   RouteAttackEndpoint + "/stop",
		}),
	}
}
