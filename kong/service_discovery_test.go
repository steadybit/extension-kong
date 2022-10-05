// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"github.com/steadybit/extension-kong/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func testDiscoverNoServicesWhenNoneAreConfigured(t *testing.T, instance *config.Instance) {
	targets := GetServiceTargets(instance)
	assert.Empty(t, targets)
}
