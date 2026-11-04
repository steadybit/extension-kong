// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package kong

import (
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
)

func TestParameterGroups(t *testing.T) {
	for name, desc := range map[string]action_kit_api.ActionDescription{
		"route":   RequestTerminationAction{}.Describe(),
		"service": ServiceTerminationAction{}.Describe(),
	} {
		t.Run(name, func(t *testing.T) {
			var order []string
			members := map[string][]string{}
			for _, p := range desc.Parameters {
				order = append(order, p.Name)
				assert.Nil(t, p.Order, "parameter %s: the list order decides", p.Name)
				if p.Group != nil {
					members[*p.Group] = append(members[*p.Group], p.Name)
				} else if p.Advanced != nil {
					// No advanced parameter is left alone in "Additional Settings" on platforms with groups.
					assert.False(t, *p.Advanced, "parameter %s", p.Name)
				}
			}
			assert.Equal(t, []string{"duration", "status", "message", "contentType", "body", "consumer", "trigger"}, order)
			assert.Equal(t, map[string][]string{
				groupResponseDetails: {"message", "contentType", "body"},
				groupTrafficFilter:   {"consumer", "trigger"},
			}, members)
			assert.Len(t, *desc.ParameterGroups, 1)
			assert.Equal(t, groupTrafficFilter, (*desc.ParameterGroups)[0].Name)
		})
	}
}
