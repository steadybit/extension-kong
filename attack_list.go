// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"encoding/json"
	"net/http"
)

func getAttackList(w http.ResponseWriter, _ *http.Request, _ []byte) {
	w.Header().Set("Content-Type", "application/json")

	attackList := AttackListResponse{
		Attacks: []EndpointRef{
			{
				"GET",
				"/attacks/request-termination",
			},
		},
	}

	json.NewEncoder(w).Encode(attackList)
}
