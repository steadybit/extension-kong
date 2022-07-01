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

type AttackListResponse struct {
	Attacks []EndpointRef `json:"attacks"`
}

type AttackParameter struct {
	Label        string `json:"label"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	Advanced     bool   `json:"advanced"`
	Order        int    `json:"order"`
	DefaultValue string `json:"defaultValue"`
}

type DescribeAttackResponse struct {
	Id          string            `json:"id"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Category    string            `json:"category"`
	Target      string            `json:"target"`
	TimeControl string            `json:"timeControl"`
	Parameters  []AttackParameter `json:"parameters"`
	Prepare     EndpointRef       `json:"prepare"`
	Start       EndpointRef       `json:"start"`
	Stop        EndpointRef       `json:"stop"`
}

type PrepareAttackRequest[T any] struct {
	Config T `json:"config"`
}

type PrepareAttackResponse[T any] struct {
	State T `json:"state"`
}

type StartAttackRequest[T any] struct {
	State T `json:"state"`
}

type StartAttackResponse[T any] struct {
	State T `json:"state"`
}

type StopAttackRequest[T any] struct {
	State T `json:"state"`
}
