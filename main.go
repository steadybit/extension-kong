// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
)

func main() {
	http.Handle("/", panicRecovery(logRequest(rootHandler)))
	http.Handle("/attacks/request-termination", panicRecovery(logRequest(describeRequestTermination)))
	http.Handle("/attacks/request-termination/prepare", panicRecovery(logRequest(prepareRequestTermination)))
	http.Handle("/attacks/request-termination/start", panicRecovery(logRequest(startRequestTermination)))
	http.Handle("/attacks/request-termination/stop", panicRecovery(logRequest(stopRequestTermination)))
	http.Handle("/discoveries/services", panicRecovery(logRequest(describeServices)))
	http.Handle("/discoveries/services/type", panicRecovery(logRequest(describeServiceType)))
	http.Handle("/discoveries/services/type/attributes", panicRecovery(logRequest(describeKongTypeAttributes)))
	http.Handle("/discoveries/services/discover", panicRecovery(logRequest(discoverServices)))

	port := 8084
	InfoLogger.Printf("Starting kong extension server on port %d. Get started via /\n", port)
	InfoLogger.Printf("Starting with configuration %s\n", Instances)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func rootHandler(w http.ResponseWriter, request *http.Request, _ []byte) {
	if request.URL.Path != "/" {
		w.WriteHeader(404)
		return
	}

	writeBody(w, ExtensionListResponse{
		Attacks: []EndpointRef{
			{
				"GET",
				"/attacks/request-termination",
			},
		},
		Discoveries: []EndpointRef{
			{
				"GET",
				"/discoveries/services",
			},
		},
		TargetTypes: []EndpointRef{
			{
				"GET",
				"/discoveries/services/type",
			},
		},
		TargetAttributes: []EndpointRef{
			{
				"GET",
				"/discoveries/services/type/attributes",
			},
		},
	})
}

func panicRecovery(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ErrorLogger.Printf("Panic: %v\n %s", err, string(debug.Stack()))
				writeError(w, "Internal Server Error", nil)
			}
		}()
		next(w, r)
	}
}

func logRequest(next func(w http.ResponseWriter, r *http.Request, body []byte)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, bodyReadErr := ioutil.ReadAll(r.Body)
		if bodyReadErr != nil {
			http.Error(w, bodyReadErr.Error(), http.StatusBadRequest)
			return
		}

		if len(body) > 0 {
			InfoLogger.Printf("%s %s with body %s", r.Method, r.URL, body)
		} else {
			InfoLogger.Printf("%s %s", r.Method, r.URL)
		}

		next(w, r, body)
	}
}

func writeError(w http.ResponseWriter, title string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	var response ErrorResponse
	if err != nil {
		response = ErrorResponse{Title: title, Detail: err.Error()}
	} else {
		response = ErrorResponse{Title: title}
	}
	json.NewEncoder(w).Encode(response)
}

func writeBody(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}
