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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/swift-go"
)

// The time format to use when adding the validation time to the response.
const ValidationTimeFormat = "2006-01-02T15:04:05Z07:00"

// handlerDecryptRawAsJSON returns the original data held in the the operation.
// Used by user interfaces to get the operations details for dispaly, or to
// continue a storage operation after time has passed waiting for the user.
// This method should never be used for passing for purposes other than for
// users editing their data.
func handlerDecryptRawAsJSON(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Create the SWAN Operator model from the information in the request
		// which represents the results of a SWIFT operation.
		m := createResponseModel(s, w, r)
		if m == nil {
			return
		}

		// Set the validity for the response now that the final data is present.
		err := m.SetValidity(s.config.DeleteDays)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Turn the model into a JSON string.
		j, err := json.Marshal(m)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Send the JSON string.
		common.SendJS(w, j)
	}
}

// handlerDecryptAsJSON turns the the "encrypted" parameter into an array of key
// value pairs where the value is encoded as an OWID using the credentials of
// the SWAN Operator.
// If the timestamp of the data provided has expired then an error is returned.
// The Email value is converted to a hashed version before being returned.
func handlerDecryptAsJSON(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Create the SWAN Operator model from the information in the request
		// which represents the results of a SWIFT operation.
		m := createResponseModel(s, w, r)
		if m == nil {
			return
		}

		// IMPORTANT: The caller is not allowed access to the raw Email or Salt.
		// Remove these values now that the SID is created.
		m.Email = nil
		m.Salt = nil

		// Set the validity for the response now that the final data is present.
		err := m.SetValidity(s.config.DeleteDays)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Turn the model into a JSON string.
		j, err := json.Marshal(m)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Send the JSON string.
		common.SendJS(w, j)
	}
}

func createResponseModel(
	s *services,
	w http.ResponseWriter,
	r *http.Request) *Response {

	// Get the SWIFT results from the request.
	o := getResults(s, w, r)
	if o == nil {
		return nil
	}

	// Turn the SWIFT results into a SWAN Operator model.
	m := &Response{}
	err := m.UnmarshalSwift(o, s, r)
	if err != nil {
		common.ReturnServerError(w, err)
		return nil
	}

	// If there is no SID then set the SID.
	if m.SID == nil {
		err = m.setSID(s, r)
		if err != nil {
			common.ReturnServerError(w, fmt.Errorf("set sid: %w", err))
			return nil
		}
	}

	// If there is no RID then create a new RID.
	if m.RID == nil {
		err = m.newRID(s, r)
		if err != nil {
			common.ReturnServerError(w, fmt.Errorf("new rid: %w", err))
			return nil
		}
	}

	// Set the expiry time on the cookies based on the configuration for the
	// operators where DeleteDays contains the number of days that should elapse
	// before the data is automatically removed.
	e := s.config.DeleteDays
	if m.RID != nil {
		m.RID.GetCookie().Expires = m.RID.GetOWID().GetExpires(e)
	}
	if m.Pref != nil {
		m.Pref.GetCookie().Expires = m.Pref.GetOWID().GetExpires(e)
	}
	if m.Email != nil {
		m.Email.GetCookie().Expires = m.Email.GetOWID().GetExpires(e)
	}
	if m.Salt != nil {
		m.Salt.GetCookie().Expires = m.Salt.GetOWID().GetExpires(e)
	}
	if m.SID != nil {

		// The SID expiry time should be passed on the earliest of the email
		// or the salt when they are available. If they are not available then
		// the number of days since the creation time should be used.
		var n time.Time
		if m.Email != nil && m.Salt != nil {
			a := m.Email.GetCookie().Expires
			b := m.Salt.GetCookie().Expires
			n = a
			if b.Before(n) {
				n = b
			}
		} else {
			n = m.SID.GetOWID().GetExpires(e)
		}
		m.SID.GetCookie().Expires = n
	}

	return m
}

// getResults validates access, unpacks the results and validates the timestamp.
func getResults(
	s *services,
	w http.ResponseWriter,
	r *http.Request) *swift.Results {

	// Check caller can access.
	if !s.getAllowedHttp(w, r) {
		return nil
	}

	// Get the SWIFT results from the request.
	o := getSWIFTResults(s, w, r)
	if o == nil {
		return nil
	}

	// Validate that the timestamp has not expired.
	if !o.IsTimeStampValid() {
		common.ReturnApplicationError(w, &common.HttpError{
			Message: "data expired and can no longer be used",
			Code:    http.StatusBadRequest})
		return nil
	}

	return o
}

// Check that the encrypted parameter is present and if so decodes and decrypts
// it to return the SWIFT results. If there is an error then the method will be
// responsible for handling the response.
func getSWIFTResults(
	s *services,
	w http.ResponseWriter,
	r *http.Request) *swift.Results {

	// Validate that the encrypted parameter is present.
	v := r.Form.Get("encrypted")
	if v == "" {
		common.ReturnApplicationError(w, &common.HttpError{
			Message: "missing 'encrypted' parameter",
			Code:    http.StatusBadRequest})
		return nil
	}

	// Decode the query string to form the byte array.
	d, err := base64.RawURLEncoding.DecodeString(v)
	if err != nil {
		common.ReturnApplicationError(w, &common.HttpError{
			Message: "not url encoded base64 string",
			Code:    http.StatusBadRequest})
		return nil
	}

	// Decrypt the string with the access node.
	o, err := decryptAndDecode(s.swift, r.Host, d)
	if err != nil {
		common.ReturnServerError(w, err)
		return nil
	}

	return o
}
