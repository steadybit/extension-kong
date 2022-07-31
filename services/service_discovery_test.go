// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package services

import (
	"context"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetup(t *testing.T) {
	WithTestContainers(t, []WithTestContainersCase{
		{
			Name: "Discover a single service",
			Test: testDiscoverServices,
		},
		{
			Name: "Kong has no services by default",
			Test: testDiscoverNoServicesWhenNoneAreConfigured,
		},
	})
}

func testDiscoverServices(t *testing.T, instance *config.Instance) {
	configureService(t, instance, &kong.Service{
		Enabled:  discovery_kit_api.Ptr(true),
		Host:     discovery_kit_api.Ptr("mockbin.org"),
		Name:     discovery_kit_api.Ptr("mockbin"),
		Path:     discovery_kit_api.Ptr("/request"),
		Port:     discovery_kit_api.Ptr(443),
		Protocol: discovery_kit_api.Ptr("https"),
	})

	targets := GetServiceTargets(instance)
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "mockbin", target.Label)
	assert.Equal(t, []string{"https://mockbin.org:443/request"}, target.Attributes["kong.service.url"])
	assert.Equal(t, []string{"true"}, target.Attributes["kong.service.enabled"])
}

func testDiscoverNoServicesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := GetServiceTargets(instance)
	assert.Empty(t, targets)
}

func configureService(t *testing.T, instance *config.Instance, service *kong.Service) *kong.Service {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdService, err := client.Services.Create(context.Background(), service)
	require.NoError(t, err)
	return createdService
}
