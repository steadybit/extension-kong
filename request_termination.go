// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"net/http"
	url2 "net/url"
)

type RequestTerminationConfig struct {
	Message  string `json:"message"`
	Status   int    `json:"status"`
	Instance string `json:"instance"`
	Service  string `json:"service"`
	Consumer string `json:"consumer"`
}

type RequestTerminationState struct {
	PluginIds []string
	Instance  Instance
	Service   Service
	Consumer  Consumer
}

type Service struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Consumer struct {
	Id       string `json:"id"`
	Username string `json:"username"`
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
	instance := Instances[instanceIndex]

	serviceNameOrId := prepareAttackRequest.Config.Service
	service, err := findService(instance, serviceNameOrId)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(ErrorResponse{
			Title:  fmt.Sprintf("Failed to find service '%s' within Kong", serviceNameOrId),
			Detail: err.Error(),
		})
		return
	}

	consumer := Consumer{}
	consumerNameOrId := prepareAttackRequest.Config.Consumer
	if len(consumerNameOrId) > 0 {
		consumer, err = findConsumer(instance, consumerNameOrId)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(ErrorResponse{
				Title:  fmt.Sprintf("Failed to find consumer '%s' within Kong", consumerNameOrId),
				Detail: err.Error(),
			})
			return
		}
	}

	pluginId, err := definePlugins(prepareAttackRequest.Config, instance, service, consumer)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(ErrorResponse{
			Title:  "Failed to define plugin within Kong",
			Detail: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(PrepareAttackResponse[RequestTerminationState]{
		State: RequestTerminationState{
			Service:  service,
			Consumer: consumer,
			Instance: instance,
			PluginIds: []string{
				pluginId,
			},
		},
	})
}

func findService(instance Instance, nameOrId string) (Service, error) {
	service := &Service{}

	url := fmt.Sprintf("%s/services/%s", instance.Origin, url2.PathEscape(nameOrId))
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add("Accept", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return *service, err
	}

	if response.StatusCode != 200 {
		return *service, errors.New(fmt.Sprintf("Kong service endpoint 'GET %s' responded with unexpected status code '%d'", url, response.StatusCode))
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return *service, err
	}

	err = json.Unmarshal(content, service)
	if err != nil {
		return *service, err
	}

	return *service, nil
}

func findConsumer(instance Instance, nameOrId string) (Consumer, error) {
	consumer := &Consumer{}

	url := fmt.Sprintf("%s/consumers/%s", instance.Origin, url2.PathEscape(nameOrId))
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add("Accept", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return *consumer, err
	}

	if response.StatusCode != 200 {
		return *consumer, errors.New(fmt.Sprintf("Kong consumer endpoint 'GET %s' responded with unexpected status code '%d'", url, response.StatusCode))
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return *consumer, err
	}

	err = json.Unmarshal(content, consumer)
	if err != nil {
		return *consumer, err
	}

	return *consumer, nil
}

type PluginDefinitionRequestBody struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Service struct {
		Id string `json:"id"`
	} `json:"service"`
	Consumer struct {
		Id string `json:"id,omitempty"`
	} `json:"consumer,omitempty"`
	Config struct {
		StatusCode int    `json:"status_code,omitempty"`
		Message    string `json:"message,omitempty"`
	} `json:"config,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type Plugin struct {
	Id string `json:"id"`
}

func definePlugins(config RequestTerminationConfig, instance Instance, service Service, consumer Consumer) (string, error) {
	req := PluginDefinitionRequestBody{
		Name:    "request-termination",
		Enabled: false,
		Tags: []string{
			"created-by=steadybit",
		},
	}
	req.Service.Id = service.Id
	req.Config.StatusCode = config.Status
	req.Config.Message = config.Message

	if len(consumer.Id) > 0 {
		req.Consumer.Id = consumer.Id
	}

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(req)
	url := fmt.Sprintf("%s/services/%s/plugins", instance.Origin, url2.PathEscape(service.Id))
	request, _ := http.NewRequest(http.MethodPost, url, body)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 201 {
		return "", errors.New(fmt.Sprintf("Kong plugin definition endpoint 'POST %s' responded with unexpected status code '%d'", url, response.StatusCode))
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	resp := &Plugin{}
	err = json.Unmarshal(content, resp)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
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
