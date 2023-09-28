// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kong/config"
	"net/http"
	"strconv"
	"strings"
)

const ServiceDiscoveryEndpoint = "/kong/service/discovery"

func RegisterServiceDiscoveryHandlers() {
	exthttp.RegisterHttpHandler(ServiceDiscoveryEndpoint, exthttp.GetterAsHandler(getServiceDiscoveryDescription))
	exthttp.RegisterHttpHandler(ServiceDiscoveryEndpoint+"/target-description", exthttp.GetterAsHandler(getServiceTargetDescription))
	exthttp.RegisterHttpHandler(ServiceDiscoveryEndpoint+"/discovered-services", getServiceDiscoveryResults)
}

func getServiceDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         ServiceTargetId,
		RestrictTo: discovery_kit_api.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         ServiceDiscoveryEndpoint + "/discovered-services",
			CallInterval: discovery_kit_api.Ptr("300s"),
		},
	}
}

func getServiceTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       ServiceTargetId,
		Label:    discovery_kit_api.PluralLabel{One: "Kong service", Other: "Kong services"},
		Category: discovery_kit_api.Ptr("API gateway"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     discovery_kit_api.Ptr(ServiceIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "kong.service.name"},
				{Attribute: "kong.service.url"},
				{Attribute: "kong.instance.name"},
				{Attribute: "kong.service.tag"},
				{Attribute: "kong.service.enabled"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "kong.service.name",
					Direction: "ASC",
				},
			},
		},
	}
}

func getServiceDiscoveryResults(w http.ResponseWriter, _ *http.Request, _ []byte) {
	var targets = make([]discovery_kit_api.Target, 0, 100)
	for _, instance := range config.Instances {
		targets = append(targets, GetServiceTargets(&instance)...)
	}
	exthttp.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
}

func GetServiceTargets(instance *config.Instance) []discovery_kit_api.Target {
	services, err := instance.GetServices()
	if err != nil {
		log.Err(err).Msgf("Failed to get services from Kong instance %s (%s)", instance.Name, instance.BaseUrl)
		return []discovery_kit_api.Target{}
	}

	targets := make([]discovery_kit_api.Target, len(services))
	for i, service := range services {

		attributes := make(map[string][]string)
		attributes["kong.instance.name"] = []string{instance.Name}
		if service.ID != nil {
			attributes["kong.service.id"] = []string{*service.ID}
		}
		if service.Name != nil {
			attributes["kong.service.name"] = []string{*service.Name}
			attributes["steadybit.label"] = []string{*service.Name}
		}
		var url strings.Builder
		if service.Protocol != nil {
			attributes["kong.service.protocol"] = []string{*service.Protocol}
			fmt.Fprintf(&url, "%s://", *service.Protocol)
		}
		if service.Host != nil {
			attributes["kong.service.path"] = []string{*service.Host}
			fmt.Fprintf(&url, "%s", *service.Host)
		}
		if service.Port != nil {
			attributes["kong.service.port"] = []string{strconv.Itoa(*service.Port)}
			fmt.Fprintf(&url, ":%d", *service.Port)
		}
		if service.Path != nil {
			attributes["kong.service.path"] = []string{*service.Path}
			fmt.Fprintf(&url, "%s", *service.Path)
		}
		if service.URL != nil {
			attributes["kong.service.url"] = []string{*service.URL}
		} else if url.Len() > 0 {
			attributes["kong.service.url"] = []string{url.String()}
		}
		if service.Enabled != nil {
			attributes["kong.service.enabled"] = []string{strconv.FormatBool(*service.Enabled)}
		}
		for _, tag := range service.Tags {
			attributes["kong.service.tag"] = append(attributes["kong.service.tag"], *tag)
		}

		targets[i] = discovery_kit_api.Target{
			Id:         fmt.Sprintf("%s-%s", instance.Name, *service.ID),
			Label:      *service.Name,
			TargetType: ServiceTargetId,
			Attributes: attributes,
		}
	}
	return discovery_kit_commons.ApplyAttributeExcludes(targets, config.Config.DiscoveryAttributesExcludesService)
}
