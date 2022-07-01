// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func loggingMiddleware(next func(w http.ResponseWriter, r *http.Request, body []byte)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

func main() {
	http.Handle("/attacks", loggingMiddleware(getAttackList))
	http.Handle("/attacks/request-termination", loggingMiddleware(describeRequestTermination))
	http.Handle("/attacks/request-termination/prepare", loggingMiddleware(prepareRequestTermination))
	http.Handle("/attacks/request-termination/start", loggingMiddleware(startRequestTermination))
	http.Handle("/attacks/request-termination/stop", loggingMiddleware(stopRequestTermination))

	port := 8084
	InfoLogger.Printf("Starting kong extension server on port %d. Get started via /attacks\n", port)

	InfoLogger.Printf("Starting with configuration %s\n", Instances)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
