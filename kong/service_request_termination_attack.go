// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
)

type ServiceTerminationAction struct {
}

func NewServiceRequestTerminationAction() action_kit_sdk.Action[RequestTerminationState] {
	return ServiceTerminationAction{}
}

var _ action_kit_sdk.Action[RequestTerminationState] = (*ServiceTerminationAction)(nil)
var _ action_kit_sdk.ActionWithStop[RequestTerminationState] = (*ServiceTerminationAction)(nil)

func (f ServiceTerminationAction) NewEmptyState() RequestTerminationState {
	return RequestTerminationState{}
}

func (f ServiceTerminationAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          "com.steadybit.extension_kong.request_termination",
		Label:       "Terminate requests",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures at Kong service level.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(ServiceIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType: ServiceTargetId,
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label:       "route-id",
					Description: extutil.Ptr("Find service by id"),
					Query:       "kong.service.id=\"\"",
				},
				{
					Label:       "route-name",
					Description: extutil.Ptr("Find service by name"),
					Query:       "kong.service.name=\"\"",
				},
			}),
		}),
		Technology:  extutil.Ptr("Kong"),
		Kind:        action_kit_api.Attack,
		TimeControl: action_kit_api.TimeControlExternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Label:        "Duration",
				Name:         "duration",
				Type:         action_kit_api.ActionParameterTypeDuration,
				Advanced:     extutil.Ptr(false),
				Required:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("30s"),
			},
			{
				Label:       "Consumer Username or ID",
				Name:        "consumer",
				Description: extutil.Ptr("You may optionally define for which Kong consumer the traffic should be impacted."),
				Type:        action_kit_api.ActionParameterTypeString,
				Advanced:    extutil.Ptr(false),
				Required:    extutil.Ptr(false),
			},
			{
				Label:        "Message",
				Name:         "message",
				Type:         action_kit_api.ActionParameterTypeString,
				Advanced:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("Error injected through the Steadybit Kong extension (through the request-termination Kong plugin)"),
			},
			{
				Label:       "Content-Type",
				Name:        "contentType",
				Description: extutil.Ptr("Content-Type response header to be returned for terminated requests."),
				Type:        action_kit_api.ActionParameterTypeString,
				Advanced:    extutil.Ptr(true),
			},
			{
				Label:       "Body",
				Name:        "body",
				Description: extutil.Ptr("The raw response body to be returned for terminated requests. This is mutually exclusive with the message parameter. A body parameter takes precedence over the message parameter."),
				Type:        action_kit_api.ActionParameterTypeString,
				Advanced:    extutil.Ptr(true),
			},
			{
				Label:        "HTTP status code",
				Name:         "status",
				Type:         action_kit_api.ActionParameterTypeInteger,
				Advanced:     extutil.Ptr(true),
				DefaultValue: extutil.Ptr("500"),
			},
			{
				Label:       "Trigger",
				Name:        "trigger",
				Type:        action_kit_api.ActionParameterTypeString,
				Description: extutil.Ptr("When not set, the plugin always activates. When set to a string, the plugin will activate exclusively on requests containing either a header or a query parameter that is named the string."),
				Advanced:    extutil.Ptr(true),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f ServiceTerminationAction) Prepare(ctx context.Context, state *RequestTerminationState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	return NewRequestTerminationAction().Prepare(ctx, state, request)
}

func (f ServiceTerminationAction) Start(ctx context.Context, state *RequestTerminationState) (*action_kit_api.StartResult, error) {
	return NewRequestTerminationAction().Start(ctx, state)
}

func (f ServiceTerminationAction) Stop(ctx context.Context, state *RequestTerminationState) (*action_kit_api.StopResult, error) {
	return NewRequestTerminationAction().(action_kit_sdk.ActionWithStop[RequestTerminationState]).Stop(ctx, state)
}
