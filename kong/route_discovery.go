// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
	"strings"
)

const RouteDiscoveryEndpoint = "/kong/route/discovery"

func RegisterRouteDiscoveryHandlers() {
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint, utils.GetterAsHandler(getRouteDiscoveryDescription))
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint+"/target-description", utils.GetterAsHandler(getRouteTargetDescription))
	utils.RegisterHttpHandler(RouteDiscoveryEndpoint+"/discovered-routes", getRouteDiscoveryResults)
}

func getRouteDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         RouteTargetID,
		RestrictTo: discovery_kit_api.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         RouteDiscoveryEndpoint + "/discovered-routes",
			CallInterval: discovery_kit_api.Ptr("600s"),
		},
	}
}

func getRouteTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       RouteTargetID,
		Label:    discovery_kit_api.PluralLabel{One: "Kong route", Other: "Kong routes"},
		Category: discovery_kit_api.Ptr("API gateway"),
		Version:  "1.1.1",
		Icon:     discovery_kit_api.Ptr(RouteIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "kong.instance.name"},
				{Attribute: "kong.route.name"},
				{Attribute: "kong.route.id"},
				{Attribute: "kong.service.name"},
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

func getRouteDiscoveryResults(w http.ResponseWriter, _ *http.Request, _ []byte) {
	var targets = make([]discovery_kit_api.Target, 0, 1000)
	for _, instance := range config.Instances {
		targets = append(targets, GetRouteTargets(&instance)...)
	}
	utils.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
}

func GetRouteTargets(instance *config.Instance) []discovery_kit_api.Target {
	services, err := instance.GetServices()
	if err != nil {
		log.Err(err).Msgf("Failed to get services from Kong instance %s (%s)", instance.Name, instance.BaseUrl)
		return []discovery_kit_api.Target{}
	}

	targets := make([]discovery_kit_api.Target, 0, len(services)*10)
	for _, service := range services {
		routes, _, err := instance.GetRoutesForService(service.ID)
		if err != nil {
			log.Err(err).Msgf("Failed to get routes from Kong instance %s (%s) for service %s (%s)", instance.Name, instance.BaseUrl, *service.Name, *service.ID)
			continue
		}

		for _, route := range routes {
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
			attributes["kong.service.id"] = []string{*service.ID}
			attributes["kong.service.name"] = []string{*service.Name}
			fmt.Fprintf(&url, "%s://", *service.Name)

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

			if route.ID != nil && route.Name != nil {
				targets = append(targets, discovery_kit_api.Target{
					Id:         fmt.Sprintf("%s-%s", instance.Name, *route.ID),
					Label:      *route.Name,
					TargetType: RouteTargetID,
					Attributes: attributes,
				})
			}

		}
	}
	return targets
}
