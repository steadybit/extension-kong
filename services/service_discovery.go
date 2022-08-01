// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package services

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kong/config"
	"github.com/steadybit/extension-kong/utils"
	"net/http"
	"strconv"
	"strings"
)

func RegisterServiceDiscoveryHandlers() {
	utils.RegisterHttpHandler("/service/discovery", utils.GetterAsHandler(getServiceDiscoveryDescription))
	utils.RegisterHttpHandler("/service/discovery/target-description", utils.GetterAsHandler(getServiceTargetDescription))
	utils.RegisterHttpHandler("/service/discovery/attribute-descriptions", utils.GetterAsHandler(getServiceAttributeDescriptions))
	utils.RegisterHttpHandler("/service/discovery/discovered-services", getServiceDiscoveryResults)
}

func getServiceDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         serviceTargetId,
		RestrictTo: discovery_kit_api.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/service/discovery/discovered-services",
			CallInterval: discovery_kit_api.Ptr("30s"),
		},
	}
}

func getServiceTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       serviceTargetId,
		Label:    discovery_kit_api.PluralLabel{One: "Kong service", Other: "Kong services"},
		Category: discovery_kit_api.Ptr("API gateway"),
		Version:  "1.1.1",
		Icon:     discovery_kit_api.Ptr(serviceIcon),
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

func getServiceAttributeDescriptions() discovery_kit_api.AttributeDescriptions {
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
			},
		},
	}
}

func getServiceDiscoveryResults(w http.ResponseWriter, _ *http.Request, _ []byte) {
	var targets = make([]discovery_kit_api.Target, 0, 100)
	for _, instance := range config.Instances {
		targets = append(targets, GetServiceTargets(&instance)...)
	}
	utils.WriteBody(w, discovery_kit_api.DiscoveredTargets{Targets: targets})
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
			TargetType: "com.github.steadybit.extension_kong.service",
			Attributes: attributes,
		}
	}
	return targets
}
