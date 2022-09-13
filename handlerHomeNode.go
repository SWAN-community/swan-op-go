/* ****************************************************************************
 * Copyright 2020 51 Degrees Mobile Experts Limited (51degrees.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 * ***************************************************************************/

package swanop

import (
	"net/http"

	"github.com/SWAN-community/common-go"
)

// handlerHomeNode returns the home node as a string.
func handlerHomeNode(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		// Check caller is authorized to access SWAN.
		if !s.getAllowedHttp(w, r) {
			return
		}

		// Get the home for the requesting browser.
		n, err := s.swift.GetHomeNode(r)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Return the response from the SWIFT layer.
		common.SendString(w, n.Domain())
	}
}
