// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"github.com/steadybit/extension-kong/v2/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testDiscoverServices(t *testing.T, instance *config.Instance) {
	// Given
	configureService(t, instance, getTestService())
	config.Config.DiscoveryAttributesExcludesService = []string{"kong.service.id"}
	// When
	targets := getServiceTargets(instance)

	// Then
	assert.NotEmpty(t, targets)
	target := targets[0]
	assert.Equal(t, "mockbin", target.Label)
	assert.Equal(t, []string{"https://mockbin.org:443/request"}, target.Attributes["kong.service.url"])
	assert.Equal(t, []string{"true"}, target.Attributes["kong.service.enabled"])
	assert.NotContains(t, target.Attributes, "kong.service.id")
}

func testDiscoverNoServicesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := getServiceTargets(instance)
	assert.Empty(t, targets)
}
