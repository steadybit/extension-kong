// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import "testing"

func TestWithTestContainers(t *testing.T) {
	WithTestContainers(t, []WithTestContainersCase{
		{
			Name: "prepare fails on missing service",
			Test: testPrepareFailsWhenServiceIsMissing,
		}, {
			Name: "prepare fails on unknown instance",
			Test: testPrepareFailsWhenInstanceIsUnknown,
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

		{
			Name: "Discover a single route",
			Test: testDiscoverRoutes,
		},
		{
			Name: "Kong has no routes by default",
			Test: testDiscoverNoRoutesWhenNoneAreConfigured,
		},

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
