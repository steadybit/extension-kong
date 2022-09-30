// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package services

import (
	"context"
	"encoding/json"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/routes"
	"github.com/steadybit/extension-kong/steadybit"
	"github.com/steadybit/extension-kong/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKongServices(t *testing.T) {
	WithTestContainers(t, []WithTestContainersCase{
		{
			Name: "Discover a single service",
			Test: testDiscoverServices,
		},
		{
			Name: "Discover a single route",
			Test: testDiscoverRoutes,
		},
		{
			Name: "Kong has no services by default",
			Test: testDiscoverNoServicesWhenNoneAreConfigured,
		},
		{
			Name: "prepare fails on missing service",
			Test: testPrepareFailsWhenServiceIsMissing,
		}, {
			Name: "prepare fails on unknown instance",
			Test: testPrepareFailsWhenInstanceIsUnknown,
		}, {
			Name: "prepare does not panic on broken JSON",
			Test: testPrepareNoPanicOnBrokenJson,
		}, {
			Name: "prepare configures disabled plugin",
			Test: testPrepareConfiguresDisabledPlugin,
		}, {
			Name: "prepare fails on unknown consumer",
			Test: testPrepareFailsOnUnknownConsumer,
		}, {
			Name: "prepare with a known consumer",
			Test: testPrepareWithConsumer,
		}, {
			Name: "prepare with a route",
			Test: testPrepareWithRoute,
		}, {
			Name: "start enables plugins",
			Test: testStartEnablesPlugin,
		}, {
			Name: "stop deletes plugins",
			Test: testStopDeletesPlugin,
		},
	})
}

func testDiscoverServices(t *testing.T, instance *config.Instance) {
	// Given
	configureService(t, instance, getTestService())

	// When
	targets := GetServiceTargets(instance)

	// Then
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "mockbin", target.Label)
	assert.Equal(t, []string{"https://mockbin.org:443/request"}, target.Attributes["kong.service.url"])
	assert.Equal(t, []string{"true"}, target.Attributes["kong.service.enabled"])
}

func testDiscoverRoutes(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	configureRoute(t, instance, getTestRoute(service))

	// When
	targets := routes.GetRouteTargets(instance)

	// Then
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "test", target.Label)
	assert.Equal(t, []string{"/products"}, target.Attributes["kong.route.path"])
}

func testDiscoverNoServicesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := GetServiceTargets(instance)
	assert.Empty(t, targets)
}

func testPrepareFailsWhenServiceIsMissing(t *testing.T, instance *config.Instance) {
	// Given
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {"unknown"},
			},
		},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)
	assert.Nil(t, state)
	assert.Contains(t, attackKitError.Title, "Failed to find service")
}

func testPrepareFailsWhenInstanceIsUnknown(t *testing.T, _ *config.Instance) {
	// Given
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {"unknown"},
				"kong.service.id":    {"unknown"},
			},
		},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)
	assert.Nil(t, state)
	assert.Contains(t, attackKitError.Title, "Failed to find a configured instance named")
}

func testPrepareNoPanicOnBrokenJson(t *testing.T, _ *config.Instance) {
	// When
	state, attackKitError := steadybit.PrepareRequestTermination([]byte{})
	assert.Nil(t, state)
	assert.Contains(t, attackKitError.Title, "Failed to parse request body")
}

func testPrepareConfiguresDisabledPlugin(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"status":  200,
			"message": "Hello from Kong extension",
		},
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)

	// Then
	assert.Nil(t, attackKitError)
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
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"status":   200,
			"message":  "Hello from Kong extension",
			"consumer": "unknown",
		},
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)

	// Then
	assert.Nil(t, state)
	assert.Contains(t, attackKitError.Title, "Failed to find consumer")
}

func testPrepareWithConsumer(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	consumer := configureConsumer(t, instance, getTestConsumer())
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"status":      200,
			"message":     "Hello from Kong extension",
			"consumer":    *consumer.Username,
			"body":        "some body",
			"contentType": "text/foobar",
			"trigger":     "banana",
		},
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}

	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)

	// Then
	assert.Nil(t, attackKitError)
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
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"status":      200,
			"message":     "Hello from Kong extension",
			"body":        "some body",
			"contentType": "text/foobar",
			"trigger":     "banana",
		},
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.route.id":      {*route.ID},
			},
		},
	}

	requestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	state, attackKitError := steadybit.PrepareRequestTermination(requestBodyJson)

	// Then
	assert.Nil(t, attackKitError)
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
	state := getSuccessfulPreparationState(t, instance)
	encodedState, err := utils.EncodeAttackState(state)
	require.NoError(t, err)
	startRequestBodyJson, err := json.Marshal(attack_kit_api.StartAttackRequestBody{
		State: encodedState,
	})
	require.NoError(t, err)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	newState, attackKitError := steadybit.StartRequestTermination(startRequestBodyJson)

	// Then
	assert.Nil(t, attackKitError)
	assert.Equal(t, instance.Name, newState.InstanceName)
	assert.Equal(t, state.PluginIds[0], newState.PluginIds[0])
	plugin, err := client.Plugins.Get(context.Background(), &state.PluginIds[0])
	require.NoError(t, err)
	assert.Equal(t, true, *plugin.Enabled)
	assert.NotNil(t, *plugin.Service.ID)
	assert.Equal(t, 200.0, plugin.Config["status_code"])
	assert.Equal(t, "Hello from Kong extension", plugin.Config["message"])
}

func getSuccessfulPreparationState(t *testing.T, instance *config.Instance) *steadybit.RequestTerminationState {
	service := configureService(t, instance, getTestService())
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"status":  200,
			"message": "Hello from Kong extension",
		},
		Target: attack_kit_api.Target{
			Attributes: map[string][]string{
				"kong.instance.name": {instance.Name},
				"kong.service.id":    {*service.ID},
			},
		},
	}
	prepareRequestBodyJson, err := json.Marshal(requestBody)
	require.NoError(t, err)

	state, attackKitError := steadybit.PrepareRequestTermination(prepareRequestBodyJson)
	require.Nil(t, attackKitError)
	return state
}

func getSuccessfulStartState(t *testing.T, instance *config.Instance) *steadybit.RequestTerminationState {
	prepareState := getSuccessfulPreparationState(t, instance)
	encodedState, err := utils.EncodeAttackState(prepareState)
	require.NoError(t, err)
	startRequestBodyJson, err := json.Marshal(attack_kit_api.StartAttackRequestBody{
		State: encodedState,
	})
	require.NoError(t, err)

	startState, attackKitError := steadybit.StartRequestTermination(startRequestBodyJson)
	assert.Nil(t, attackKitError)
	return startState
}

func testStopDeletesPlugin(t *testing.T, instance *config.Instance) {
	// Given
	state := getSuccessfulStartState(t, instance)
	encodedState, err := utils.EncodeAttackState(state)
	require.NoError(t, err)
	stopRequestBodyJson, err := json.Marshal(attack_kit_api.StartAttackRequestBody{
		State: encodedState,
	})
	require.NoError(t, err)

	client, err := instance.GetClient()
	require.NoError(t, err)

	// When
	attackKitError := steadybit.StopRequestTermination(stopRequestBodyJson)

	// Then
	assert.Nil(t, attackKitError)
	_, err = client.Plugins.Get(context.Background(), &state.PluginIds[0])
	assert.Error(t, err)
}

func getTestService() *kong.Service {
	return &kong.Service{
		Enabled:  discovery_kit_api.Ptr(true),
		Host:     discovery_kit_api.Ptr("mockbin.org"),
		Name:     discovery_kit_api.Ptr("mockbin"),
		Path:     discovery_kit_api.Ptr("/request"),
		Port:     discovery_kit_api.Ptr(443),
		Protocol: discovery_kit_api.Ptr("https"),
	}
}

func getTestRoute(service *kong.Service) *kong.Route {
	return &kong.Route{
		Name:    discovery_kit_api.Ptr("test"),
		Service: service,
		Hosts:   []*string{discovery_kit_api.Ptr("server1")},
		Paths:   []*string{discovery_kit_api.Ptr("/products")},
		Tags:    []*string{discovery_kit_api.Ptr("test")},
		Methods: []*string{utils.String("GET")},
	}
}

func configureService(t *testing.T, instance *config.Instance, service *kong.Service) *kong.Service {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdService, err := client.Services.Create(context.Background(), service)
	require.NoError(t, err)
	return createdService
}

func configureRoute(t *testing.T, instance *config.Instance, route *kong.Route) *kong.Route {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdRoute, err := client.Routes.Create(context.Background(), route)
	require.NoError(t, err)
	return createdRoute
}

func getTestConsumer() *kong.Consumer {
	return &kong.Consumer{
		Username: discovery_kit_api.Ptr("test-consumer"),
	}
}

func configureConsumer(t *testing.T, instance *config.Instance, consumer *kong.Consumer) *kong.Consumer {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdConsumer, err := client.Consumers.Create(context.Background(), consumer)
	require.NoError(t, err)
	return createdConsumer
}
