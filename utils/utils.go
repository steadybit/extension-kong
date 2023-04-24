// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package utils

func Strings(ss []string) []*string {
	r := make([]*string, len(ss))
	for i := range ss {
		r[i] = &ss[i]
	}
	return r
}
