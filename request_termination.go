// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"net/http"
)

type RequestTerminationConfig struct {
	Instance string `json:"instance"`
	Service  string `json:"service"`
	Consumer string `json:"consumer"`
}

type RequestTerminationState struct {
	PluginIds    []string
	Instance     Instance
	ServiceName  string
	ConsumerName string
}

func describeRequestTermination(w http.ResponseWriter, _ *http.Request, _ []byte) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DescribeAttackResponse{
		Id:          "com.github.steadybit.extension_kong.request_termination",
		Label:       "Terminate requests in Kong",
		Description: "Leverage the Kong request-termination plugin to inject HTTP failures.",
		Version:     "1.0.0",
		// TODO can we get rid of this target?
		Target:      "container",
		Category:    "network",
		TimeControl: "EXTERNAL",
		Parameters: []AttackParameter{
			{
				Label:    "Duration",
				Name:     "duration",
				Type:     "duration",
				Advanced: false,
				Required: true,
			},
			{
				Label:    "Instance",
				Name:     "instance",
				Type:     "string",
				Advanced: false,
				Required: true,
			},
			{
				Label:    "Service Name or ID",
				Name:     "service",
				Type:     "string",
				Advanced: false,
				Required: true,
			},
			{
				Label:       "Consumer Username or ID",
				Name:        "consumer",
				Description: "You may optionally define for which Kong consumer the traffic should be impacted.",
				Type:        "string",
				Advanced:    false,
				Required:    false,
			},
			{
				Label:        "Message",
				Name:         "message",
				Type:         "string",
				Advanced:     true,
				DefaultValue: "Error injected through the Steadybit Kong extension (through the request-termination Kong plugin)",
			},
			{
				Label:        "HTTP status code",
				Name:         "status",
				Type:         "integer",
				Advanced:     true,
				DefaultValue: "500",
			},
		},
		Prepare: EndpointRef{
			"POST",
			"/attacks/request-termination/prepare",
		},
		Start: EndpointRef{
			"POST",
			"/attacks/request-termination/start",
		},
		Stop: EndpointRef{
			"POST",
			"/attacks/request-termination/stop",
		},
	})
}

func prepareRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	w.Header().Set("Content-Type", "application/json")

	var prepareAttackRequest PrepareAttackRequest[RequestTerminationConfig]
	err := json.Unmarshal(body, &prepareAttackRequest)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(ErrorResponse{
			Title:  "Failed to read request body",
			Detail: err.Error(),
		})
		return
	}

	instanceIndex := slices.IndexFunc(Instances, func(i Instance) bool { return i.Name == prepareAttackRequest.Config.Instance })
	if instanceIndex < 0 {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(ErrorResponse{
			Title: fmt.Sprintf("Failed to find a configured instance named '%s'", prepareAttackRequest.Config.Instance),
		})
		return
	}

	// todo find services and consumers

	json.NewEncoder(w).Encode(PrepareAttackResponse[RequestTerminationState]{
		State: RequestTerminationState{
			ServiceName:  "42",
			ConsumerName: "43",
			Instance:     Instances[instanceIndex],
		},
	})
}

func startRequestTermination(w http.ResponseWriter, _ *http.Request, body []byte) {
	//w.Header().Set("Content-Type", "application/json")
	//
	//var startAttackRequest StartAttackRequest
	//err := json.Unmarshal(body, &startAttackRequest)
	//if err != nil {
	//	w.WriteHeader(500)
	//	json.NewEncoder(w).Encode(ErrorResponse{
	//		Title:  "Failed to read request body",
	//		Detail: err.Error(),
	//	})
	//	return
	//}
	//
	//InfoLogger.Printf("Starting rollout restart attack for %s\n", startAttackRequest)
	//
	//cmd := exec.Command("kubectl",
	//	"rollout",
	//	"restart",
	//	"--namespace",
	//	startAttackRequest.State.Namespace,
	//	fmt.Sprintf("deployment/%s", startAttackRequest.State.Deployment))
	//cmdOut, cmdErr := cmd.CombinedOutput()
	//if cmdErr != nil {
	//	ErrorLogger.Printf("Failed to execute rollout restart %s: %s", cmdErr, cmdOut)
	//	w.WriteHeader(500)
	//	json.NewEncoder(w).Encode(ErrorResponse{
	//		Title:  fmt.Sprintf("Failed to execute rollout restart %s: %s", cmdErr, cmdOut),
	//		Detail: cmdErr.Error(),
	//	})
	//	return
	//}
	//
	//json.NewEncoder(w).Encode(StartAttackResponse{
	//	State: startAttackRequest.State,
	//})
}

func stopRequestTermination(_ http.ResponseWriter, _ *http.Request, _ []byte) {
}
