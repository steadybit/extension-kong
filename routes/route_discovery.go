// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package routes

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/common"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
	"strings"
)

const RouteDiscoveryEndpoint = "/kong/route/discovery"

func RegisterRouteDiscoveryHandlers() {
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint, utils.GetterAsHandler(getRouteDiscoveryDescription))
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint+"/target-description", utils.GetterAsHandler(getRouteTargetDescription))
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint+"/attribute-descriptions", utils.GetterAsHandler(getRouteAttributeDescriptions))
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint+"/discovered-routes", getRouteDiscoveryResults)
}

func getRouteDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         common.RouteTargetID,
		RestrictTo: discovery_kit_api.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         RouteDiscoveryEndpoint + "/discovered-routes",
			CallInterval: discovery_kit_api.Ptr("30s"),
		},
	}
}

func getRouteTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       common.RouteTargetID,
		Label:    discovery_kit_api.PluralLabel{One: "Kong route", Other: "Kong routes"},
		Category: discovery_kit_api.Ptr("API gateway"),
		Version:  "1.1.1",
		Icon:     discovery_kit_api.Ptr(common.RouteIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "kong.instance.name"},
				{Attribute: "kong.route.name"},
				{Attribute: "kong.route.id"},
				{Attribute: "kong.route.service.name"},
				{Attribute: "kong.route.tag"},
				{Attribute: "kong.route.host"},
				{Attribute: "kong.route.method"},
				{Attribute: "kong.route.path"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "kong.route.name",
					Direction: "ASC",
				},
			},
		},
	}
}

func getRouteAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{
		Attributes: []discovery_kit_api.AttributeDescription{
			{
				Attribute: "kong.instance.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong instance name",
					Other: "Kong instance names",
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
			},
			{
				Attribute: "kong.route.service.name",
				Label: discovery_kit_api.PluralLabel{
					One:   "Kong route service name",
					Other: "Kong route service names",
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

func getRouteDiscoveryResults(w http.ResponseWriter, _ *http.Request, _ []byte) {
	var targets = make([]discovery_kit_api.Target, 0, 1000)
	for _, instance := range config.Instances {
		targets = append(targets, GetRouteTargets(&instance)...)
	}
	utils.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
}

func GetRouteTargets(instance *config.Instance) []discovery_kit_api.Target {
	routes, err := instance.GetRoutes()
	if err != nil {
		log.Err(err).Msgf("Failed to get routes from Kong instance %s (%s)", instance.Name, instance.BaseUrl)
		return []discovery_kit_api.Target{}
	}

	targets := make([]discovery_kit_api.Target, len(routes))
	for i, route := range routes {

		attributes := make(map[string][]string)
		attributes["kong.instance.name"] = []string{instance.Name}
		if route.ID != nil {
			attributes["kong.route.id"] = []string{*route.ID}
		}
		if route.Name != nil {
			attributes["kong.route.name"] = []string{*route.Name}
			attributes["steadybit.label"] = []string{*route.Name}
		}
		var url strings.Builder
		if route.Service != nil && route.Service.Name != nil {
			attributes["kong.route.service.name"] = []string{*route.Service.Name}
			fmt.Fprintf(&url, "%s://", *route.Service.Name)
		}
		for _, path := range route.Paths {
			attributes["kong.route.path"] = append(attributes["kong.route.path"], *path)
		}
		for _, host := range route.Hosts {
			attributes["kong.route.host"] = append(attributes["kong.route.host"], *host)
		}
		for _, tag := range route.Tags {
			attributes["kong.route.tag"] = append(attributes["kong.route.tag"], *tag)
		}
		for _, method := range route.Methods {
			attributes["kong.route.method"] = append(attributes["kong.route.method"], *method)
		}

		targets[i] = discovery_kit_api.Target{
			Id:         fmt.Sprintf("%s-%s", instance.Name, *route.ID),
			Label:      *route.Name,
			TargetType: "com.github.steadybit.extension_kong.route",
			Attributes: attributes,
		}
	}
	return targets
}
