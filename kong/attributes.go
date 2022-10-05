// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/utils"
)

func RegisterAttributeDescriptionHandlers() {
	utils.RegisterHttpHandler("/kong/attribute-descriptions", utils.GetterAsHandler(getAttributeDescriptions))
}

func getAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{
		Attributes: []discovery_kit_api.AttributeDescription{
			{
				Attribute: "kong.instance.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong instance name",
					Other: "Kong instance names",
				},
			}, {
				Attribute: "kong.service.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong service name",
					Other: "Kong service names",
				},
			}, {
				Attribute: "kong.service.id",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong service ID",
					Other: "Kong service IDs",
				},
			},
			{
				Attribute: "kong.service.url",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong service URL",
					Other: "Kong service URLs",
				},
			}, {
				Attribute: "kong.service.tag",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong service tag",
					Other: "Kong service tags",
				},
			}, {
				Attribute: "kong.service.enabled",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong service enabled",
					Other: "Kong service enabled",
				},
			}, {
				Attribute: "kong.route.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route name",
					Other: "Kong route names",
				},
			}, {
				Attribute: "kong.route.id",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route ID",
					Other: "Kong route IDs",
				},
			}, {
				Attribute: "kong.route.tag",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route tag",
					Other: "Kong route tags",
				},
			}, {
				Attribute: "kong.route.host",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route host",
					Other: "Kong route hosts",
				},
			}, {
				Attribute: "kong.route.method",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route method",
					Other: "Kong route methods",
				},
			}, {
				Attribute: "kong.route.path",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route path",
					Other: "Kong route paths",
				},
			},
		},
	}
}
