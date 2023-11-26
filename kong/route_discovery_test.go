// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"github.com/steadybit/extension-kong/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testDiscoverRoutes(t *testing.T, instance *config.Instance) {
	// Given
	service := configureService(t, instance, getTestService())
	configureRoute(t, instance, getTestRoute(service))
	config.Config.DiscoveryAttributesExcludesRoute = []string{"kong.service.id"}

	// When
	targets := getRouteTargets(instance)

	// Then
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "test", target.Label)
	assert.Equal(t, []string{"/products"}, target.Attributes["kong.route.path"])
	assert.Equal(t, []string{"mockbin"}, target.Attributes["kong.service.name"])
	assert.NotContains(t, target.Attributes, "kong.service.id")
}

func testDiscoverNoRoutesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := getRouteTargets(instance)
	assert.Empty(t, targets)
}
