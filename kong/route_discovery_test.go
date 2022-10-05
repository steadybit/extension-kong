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

	// When
	targets := GetRouteTargets(instance)

	// Then
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "test", target.Label)
	assert.Equal(t, []string{"/products"}, target.Attributes["kong.route.path"])
	assert.Equal(t, []string{"mockbin"}, target.Attributes["kong.service.name"])
}

func testDiscoverNoRoutesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := GetRouteTargets(instance)
	assert.Empty(t, targets)
}
