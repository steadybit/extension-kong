// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

type EndpointRef struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ErrorResponse struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type ExtensionListResponse struct {
	Attacks          []EndpointRef `json:"attacks"`
	Discoveries      []EndpointRef `json:"discoveries"`
	TargetTypes      []EndpointRef `json:"targetTypes"`
	TargetAttributes []EndpointRef `json:"targetAttributes"`
}

type DescribeTargetTypeResponse struct {
	Id       string      `json:"id"`
	Version  string      `json:"version"`
	Label    PluralLabel `json:"label"`
	Category string      `json:"category"`
	Icon     string      `json:"icon"`
	Table    Table       `json:"table"`
}

type PluralLabel struct {
	One   string `json:"one"`
	Other string `json:"other"`
}

type Table struct {
	Columns []Column `json:"columns"`
	OrderBy []Order  `json:"orderBy"`
}

type Column struct {
	Attribute          string `json:"attribute"`
	FallbackAttributes string `json:"fallbackAttributes"`
}

type Order struct {
	Attribute string `json:"attribute"`
	Direction string `json:"direction"`
}

type DescribeDiscoveryResponse struct {
	Id         string                       `json:"id"`
	Discover   EndpointRefWithCallInternval `json:"discover"`
	RestrictTo string                       `json:"restrictTo"`
}

type EndpointRefWithCallInternval struct {
	EndpointRef
	CallInterval string `json:"callInterval"`
}

type DescribeTargetAttributeResponse struct {
	Attributes []TargetAttributeDescription `json:"attributes"`
}

type TargetAttributeDescription struct {
	Attribute string      `json:"attribute"`
	Label     PluralLabel `json:"label"`
}

type DiscoverResponse struct {
	Targets []DiscoverTarget `json:"targets"`
}

type DiscoverTarget struct {
	Id         string              `json:"id"`
	Label      string              `json:"label"`
	TargetType string              `json:"targetType"`
	Attributes map[string][]string `json:"attributes"`
}
