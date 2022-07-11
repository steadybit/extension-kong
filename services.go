// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func describeServiceType(w http.ResponseWriter, _ *http.Request, _ []byte) {
	writeBody(w, DescribeTargetTypeResponse{
		Id:       "com.github.steadybit.extension_kong.service",
		Label:    PluralLabel{One: "Kong Service", Other: "Kong Services"},
		Category: "API gateway",
		Version:  "1.1.0",
		Icon:     "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='64' height='64'%3E%3Cpath d='M20.986 50.552h11.662l6.055 7.54-1.04 2.568H22.596l.37-2.568-3.552-5.548zm8.238-33.765 6.33-.01L64 50.428l-2.2 10.23H49.61l.76-2.883-26.58-31.452zM40.518 3.34 53.68 13.758l-1.685 1.75 2.282 3.2v3.422l-6.563 5.386L36.68 14.39h-6.426l2.587-4.774zm-27.46 32.852 9.256-7.935L34.6 42.84l-3.5 5.342H19.782l-7.837 10.144-1.8 2.333H0V48.213l9.465-12.02z' fill='%23003459' fill-rule='evenodd'/%3E%3C/svg%3E",
		Table: Table{
			Columns: []Column{
				{Attribute: "kong.service.name"},
				{Attribute: "kong.service.url"},
				{Attribute: "kong.instance.name"},
				{Attribute: "kong.service.tag"},
				{Attribute: "kong.service.enabled"},
			},
			OrderBy: []Order{
				{
					Attribute: "kong.service.name",
					Direction: "ASC",
				},
			},
		},
	})
}

func describeKongTypeAttributes(w http.ResponseWriter, _ *http.Request, _ []byte) {
	writeBody(w, DescribeTargetAttributeResponse{
		Attributes: []TargetAttributeDescription{
			{
				Attribute: "kong.instance.name",
				Label: PluralLabel{
					One:   "Kong Instance Name",
					Other: "Kong Instance Names",
				},
			}, {
				Attribute: "kong.service.name",
				Label: PluralLabel{
					One:   "Kong Service Name",
					Other: "Kong Service Names",
				},
			}, {
				Attribute: "kong.service.id",
				Label: PluralLabel{
					One:   "Kong Service Id",
					Other: "Kong Service Ids",
				},
			},
			{
				Attribute: "kong.service.url",
				Label: PluralLabel{
					One:   "Kong Service URL",
					Other: "Kong Service URLs",
				},
			}, {
				Attribute: "kong.service.tag",
				Label: PluralLabel{
					One:   "Kong Service Tag",
					Other: "Kong Service Tags",
				},
			}, {
				Attribute: "kong.service.enabled",
				Label: PluralLabel{
					One:   "Kong Service Enabled",
					Other: "Kong Service Enabled",
				},
			},
		},
	})
}

func describeServices(w http.ResponseWriter, _ *http.Request, _ []byte) {
	writeBody(w, DescribeDiscoveryResponse{
		Id:         "com.github.steadybit.extension_kong.service",
		RestrictTo: "LEADER",
		Discover: EndpointRefWithCallInternval{
			EndpointRef: EndpointRef{
				Method: "GET",
				Path:   "/discoveries/services/discover",
			},
			CallInterval: "30s",
		},
	})
}

func discoverServices(w http.ResponseWriter, _ *http.Request, _ []byte) {
	var targets = []DiscoverTarget{}
	for _, instance := range Instances {
		targets = append(targets, getServiceTargets(&instance)...)
	}
	writeBody(w, DiscoverResponse{Targets: targets})
}

func getServiceTargets(instance *Instance) []DiscoverTarget {
	services, err := instance.GetServices()
	if err != nil {
		ErrorLogger.Printf("Failed to get services from Kong instance %s (%s): %s", instance.Name, instance.BaseUrl, err)
		return []DiscoverTarget{}
	}

	targets := make([]DiscoverTarget, len(services))
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

		targets[i] = DiscoverTarget{
			Id:         fmt.Sprintf("%s-%s", instance.Name, *service.ID),
			Label:      *service.Name,
			TargetType: "com.github.steadybit.extension_kong.service",
			Attributes: attributes,
		}
	}
	return targets
}
