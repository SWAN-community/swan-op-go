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
	"github.com/SWAN-community/swift-go"
)

// handlerFetch returns a URL that can be used in the browser primary navigation
// to retrieve the most current data from the SWAN network. If no data is
// available default values are returned.
func handlerFetch(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		// Check caller is authorized to access SWAN.
		if !s.getAllowedHttp(w, r) {
			return
		}

		// Validate and set the return URL.
		err = swift.SetURL("returnUrl", "returnUrl", &r.Form)
		if err != nil {
			common.ReturnApplicationError(w, &common.HttpError{
				Error:   err,
				Message: "bad return url",
				Code:    http.StatusBadRequest})
			return
		}

		// Add the fields that need to be returned.
		r.Form.Set("rid>", "")
		r.Form.Set("pref>", "")
		r.Form.Set("email>", "")
		r.Form.Set("salt>", "")
		r.Form.Set("stop>", "")

		// Uses the SWIFT access node associated with this internet domain
		// to determine the URL to direct the browser to.
		u, err := createStorageOperationURL(s.swift, r, r.Form)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Return the response from the SWIFT layer.
		common.SendString(w, u)
	}
}
