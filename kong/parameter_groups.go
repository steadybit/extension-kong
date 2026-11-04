// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package kong

import "github.com/steadybit/action-kit/go/action_kit_api/v2"

const (
	groupResponseDetails = "Response Details"
	groupTrafficFilter   = "Traffic Filter"
)

var (
	routeTrafficFilterGroup = action_kit_api.ParameterGroup{
		Name:        groupTrafficFilter,
		Description: "Leave unchanged to affect all requests to the selected routes.",
	}
	serviceTrafficFilterGroup = action_kit_api.ParameterGroup{
		Name:        groupTrafficFilter,
		Description: "Leave unchanged to affect all requests to the selected services.",
	}
)
