// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kong/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func testPrepareFailsWhenServiceIsMissing(t *testing.T, instance *config.Instance) {
	// Given
	requestBody := action_kit_api.PrepareActionRequestBody{
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {"unknown"},
			},
		},
	}

	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to find service")
}

func testPrepareFailsWhenInstanceIsUnknown(t *testing.T, _ *config.Instance) {
	// Given
	requestBody := action_kit_api.PrepareActionRequestBody{
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {"unknown"},
				"kong.service.id":    {"unknown"},
			},
		},
	}
	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to find a configured instance named")
}

func testPrepareConfiguresDisabledPlugin(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	requestBody := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"status":  200,
			"message": "Hello from Kong extension",
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}

	client, err := instance.GetClient()
	require.NoError(t, err)

	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, instance.Name, state.InstanceName)
	plugin, err := client.Plugins.Get(context.Background(), &state.PluginIds[0])
	require.NoError(t, err)
	assert.Equal(t, "request-termination", *plugin.Name)
	assert.Equal(t, false, *plugin.Enabled)
	assert.Equal(t, *service.ID, *plugin.Service.ID)
	assert.Nil(t, plugin.Consumer)
	assert.Equal(t, 200.0, plugin.Config["status_code"])
	assert.Nil(t, plugin.Config["trigger"])
	assert.Nil(t, plugin.Config["body"])
	assert.Nil(t, plugin.Config["content_type"])
	assert.Equal(t, "Hello from Kong extension", plugin.Config["message"])
}

func testPrepareFailsOnUnknownConsumer(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	requestBody := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"status":   200,
			"message":  "Hello from Kong extension",
			"consumer": "unknown",
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}
	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to find consumer")
}

func testPrepareWithConsumer(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	consumer := configureConsumer(t, instance, getTestConsumer())
	requestBody := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"status":      200,
			"message":     "Hello from Kong extension",
			"consumer":    *consumer.Username,
			"body":        "some body",
			"contentType": "text/foobar",
			"trigger":     "banana",
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}

	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, result)
	assert.Nil(t, err)
	plugin, err := client.Plugins.Get(context.Background(), &state.PluginIds[0])
	require.NoError(t, err)
	assert.Equal(t, *consumer.ID, *plugin.Consumer.ID)
	assert.Equal(t, "banana", plugin.Config["trigger"])
	assert.Equal(t, "some body", plugin.Config["body"])
	assert.Nil(t, plugin.Config["message"])
	assert.Equal(t, "text/foobar", plugin.Config["content_type"])
}

func testPrepareWithRoute(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	route := configureRoute(t, instance, getTestRoute(service))
	requestBody := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"status":      200,
			"message":     "Hello from Kong extension",
			"body":        "some body",
			"contentType": "text/foobar",
			"trigger":     "banana",
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.route.id":      {*route.ID},
				"kong.service.id":    {*service.ID},
			},
		},
	}

	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, err)
	assert.Nil(t, result)
	plugin, err := client.Plugins.Get(context.Background(), &state.PluginIds[0])
	require.NoError(t, err)
	assert.Equal(t, *route.ID, *plugin.Route.ID)
	assert.Equal(t, "banana", plugin.Config["trigger"])
	assert.Equal(t, "some body", plugin.Config["body"])
	assert.Nil(t, plugin.Config["message"])
	assert.Equal(t, "text/foobar", plugin.Config["content_type"])
}

func testStartEnablesPlugin(t *testing.T, instance *config.Instance) {
	// Given
	action := NewRequestTerminationAction()
	state := getSuccessfulPreparationState(t, instance)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	result, err := action.Start(context.TODO(), state)

	// Then
	assert.Nil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, instance.Name, state.InstanceName)
	assert.Equal(t, state.PluginIds[0], state.PluginIds[0])
	plugin, err := client.Plugins.Get(context.Background(), &state.PluginIds[0])
	require.NoError(t, err)
	assert.Equal(t, true, *plugin.Enabled)
	assert.NotNil(t, *plugin.Service.ID)
	assert.Equal(t, 200.0, plugin.Config["status_code"])
	assert.Equal(t, "Hello from Kong extension", plugin.Config["message"])
}

func getSuccessfulPreparationState(t *testing.T, instance *config.Instance) *RequestTerminationState {
	service := configureService(t, instance, getTestService())
	requestBody := action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"status":  200,
			"message": "Hello from Kong extension",
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}
	action := NewRequestTerminationAction()
	state := action.NewEmptyState()

	result, err := action.Prepare(context.TODO(), &state, requestBody)
	require.Nil(t, result)
	require.Nil(t, err)
	return &state
}

func getSuccessfulStartState(t *testing.T, instance *config.Instance) *RequestTerminationState {
	state := getSuccessfulPreparationState(t, instance)
	action := NewRequestTerminationAction()
	result, err := action.Start(context.TODO(), state)
	require.Nil(t, result)
	require.Nil(t, err)
	return state
}

func testStopDeletesPlugin(t *testing.T, instance *config.Instance) {
	// Given
	state := getSuccessfulStartState(t, instance)

	action := NewRequestTerminationAction().(action_kit_sdk.ActionWithStop[RequestTerminationState])

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	result, err := action.Stop(context.TODO(), state)

	// Then
	assert.Nil(t, result)
	assert.Nil(t, err)
	_, err = client.Plugins.Get(context.Background(), &state.PluginIds[0])
	assert.Error(t, err)
}
