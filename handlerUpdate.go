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
	"strings"

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

		// Get the time when the data should be deleted.
		t := s.config.DeleteDate().Format("2006-01-02")

		// Validate that the SWAN values provided are valid OWIDs and then set
		// the values. If the RID is not provided created a new one to use if
		// a value does not exist already.
		if r.Form.Get("rid") != "" {
			err = validateRID(s, r.FormValue("rid"), "rid")
			if err != nil {
				common.ReturnApplicationError(w, &common.HttpError{
					Error:   err,
					Message: "bad rid",
					Code:    http.StatusBadRequest})
				return
			}

			// Use the > sign to indicate the newest value should be used.
			r.Form.Set(fmt.Sprintf("rid>%s", t), r.Form.Get("rid"))
			r.Form.Del("rid")
		} else {
			rid := createRID(s, w, r)
			if rid == nil {
				return
			}

			// Use the < sign to indicate the oldest, or existing value should
			// be used.
			v, err := rid.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
			r.Form.Set(fmt.Sprintf("rid<%s", t), string(v))
		}
		if r.Form.Get("pref") != "" {
			err = validatePref(s, r.FormValue("pref"), "pref")
			if err != nil {
				common.ReturnApplicationError(w, &common.HttpError{
					Error:   err,
					Message: "bad pref",
					Code:    http.StatusBadRequest})
				return
			}
			r.Form.Set(fmt.Sprintf("pref>%s", t), r.Form.Get("pref"))
			r.Form.Del("pref")
		}
		if r.Form.Get("email") != "" {
			err = validateEmail(s, r.FormValue("email"), "email")
			if err != nil {
				common.ReturnApplicationError(w, &common.HttpError{
					Error:   err,
					Message: "bad email",
					Code:    http.StatusBadRequest})
				return
			}
			r.Form.Set(fmt.Sprintf("email>%s", t), r.Form.Get("email"))
			r.Form.Del("email")
		}
		if r.Form.Get("salt") != "" {
			err = validateSalt(s, r.FormValue("salt"), "salt")
			if err != nil {
				common.ReturnApplicationError(w, &common.HttpError{
					Error:   err,
					Message: "bad salt",
					Code:    http.StatusBadRequest})
				return
			}
			r.Form.Set(fmt.Sprintf("salt>%s", t), r.Form.Get("salt"))
			r.Form.Del("salt")
		}
		if r.Form.Get("stop") != "" {
			r.Form.Set(fmt.Sprintf("stop+%s", t), r.Form.Get("stop"))
			r.Form.Del("stop")
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

// validateOWID validates that the OWID is correct if the domain is not
// localhost. Localhost is always allowed to enable debugging.
func validateOWID(s *services, k string, o *owid.OWID) error {
	if !strings.EqualFold(o.Domain, "localhost") {
		b, err := o.Verify(s.config.Scheme)
		if err != nil {
			return err
		}
		if !b {
			return fmt.Errorf("'%s' not a verified OWID", k)
		}
	}
	return nil
}

func validateRID(s *services, k string, v string) error {
	i, err := swan.IdentifierUnmarshalBase64([]byte(v))
	if err != nil {
		return err
	}
	return validateOWID(s, k, i.Base.OWID)
}

func validatePref(s *services, k string, v string) error {
	p, err := swan.PreferencesUnmarshalBase64([]byte(v))
	if err != nil {
		return err
	}
	return validateOWID(s, k, p.OWID)
}

func validateEmail(s *services, k string, v string) error {
	e, err := swan.EmailUnmarshalBase64([]byte(v))
	if err != nil {
		return err
	}
	return validateOWID(s, k, e.OWID)
}

func validateSalt(s *services, k string, v string) error {
	t, err := swan.SaltUnmarshalBase64([]byte(v))
	if err != nil {
		return err
	}
	return validateOWID(s, k, t.OWID)
}
