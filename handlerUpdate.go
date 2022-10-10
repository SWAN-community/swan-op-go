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
	"fmt"
	"net/http"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
)

// handlerUpdate returns a URL that can be used in the browser primary
// navigation to update the SWAN network data with the values provided in the
// form parameters.
func handlerUpdate(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check caller is authorized to access SWAN.
		if !s.getAllowedHttp(w, r) {
			return
		}

		// Turn the incoming request into a model.
		m := swan.ModelRequestFromHttpRequest(r, w)
		if m == nil {
			return
		}

		// Valid that the data in the model is correct if not in debug mode.
		// This check needs to be by passed for tests where the relevant OWID
		// signers are not available via HTTP.
		if !s.config.Debug && !m.Verify(w, s.config.Scheme) {
			return
		}

		// As this is an update operation do not use the home node alone.
		r.Form.Set("useHomeNode", "false")

		// Validate and set the return URL.
		err := swift.SetURL("returnUrl", "returnUrl", &r.Form)
		if err != nil {
			common.ReturnApplicationError(w, &common.HttpError{
				Error:   err,
				Message: "bad return url",
				Code:    http.StatusBadRequest})
			return
		}

		// Validate that the SWAN values provided are valid OWIDs and then set
		// the values. If the RID is not provided created a new one to use if
		// a value does not exist already.
		if m.RID != nil && m.RID.IsSigned() {
			// Use the > sign to indicate the newest value should be used.
			b, err := m.RID.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("rid>%s",
				getExpire(s, m.RID.GetOWID())),
				string(b))
		} else {
			rid, err := createRID(s, r)
			if err != nil {
				common.ReturnServerError(w, err)
			}

			// Use the < sign to indicate the oldest, or existing value should
			// be used. This new one should be used only if others don't already
			// exist.
			b, err := rid.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("rid<%s",
				getExpire(s, rid.GetOWID())),
				string(b))
		}
		if m.Pref != nil && m.Pref.IsSigned() {
			// Use the > sign to indicate the newest value should be used.
			b, err := m.Pref.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("pref>%s",
				getExpire(s, m.Pref.GetOWID())),
				string(b))
		}
		if m.Email != nil && m.Email.IsSigned() {
			// Use the > sign to indicate the newest value should be used.
			b, err := m.Email.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("email>%s",
				getExpire(s, m.Email.GetOWID())),
				string(b))
		}
		if m.Salt != nil && m.Salt.IsSigned() {
			// Use the > sign to indicate the newest value should be used.
			b, err := m.Salt.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("salt>%s",
				getExpire(s, m.Salt.GetOWID())),
				string(b))
		}
		if m.Stop != nil {
			// Use the + sign to indicate duplicate values should be merged.
			t := s.config.DeleteDate().Format("2006-01-02")
			for _, v := range m.Stop.Value {
				r.Form.Add(fmt.Sprintf("stop+%s", t), v)
			}
		}

		// Uses the SWIFT access node associated with this internet domain
		// to determine the URL to direct the browser to.
		u, err := createStorageOperationURL(s.swift, r, r.Form)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Return the URL from the SWIFT layer.
		common.SendString(w, u)
	}
}

// getExpire returns a string representation of the date when the value will
// be removed from SWIFT. The OWID's created date is added to the operator's
// delete days value.
func getExpire(s *services, o *owid.OWID) string {
	return o.GetExpires(s.config.DeleteDays).Format("2006-01-02")
}
