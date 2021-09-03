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
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swift-go"
	"net/http"
	"net/url"
)

// handlerUpdate returns a URL that can be used in the browser primary
// navigation to update the SWAN network data with the values provided in the
// form parameters.
func handlerUpdate(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check caller is authorized to access SWAN.
		if s.getAccessAllowed(w, r) == false {
			return
		}

		// As this is an update operation do not use the home node alone.
		r.Form.Set("useHomeNode", "false")

		// Validate and set the return URL.
		err := swift.SetURL("returnUrl", "returnUrl", &r.Form)
		if err != nil {
			returnAPIError(&s.config, w, err, http.StatusBadRequest)
			return
		}

		// Get the time when the data should be deleted.
		t := s.config.DeleteDate().Format("2006-01-02")

		// Validate that the SWAN values provided are valid OWIDs and then set
		// the values. If the SWID is not provided created a new one to use if
		// a value does not exist already.
		if r.Form.Get("swid") != "" {
			err = validateOWID(s, &r.Form, "swid")
			if err != nil {
				returnAPIError(&s.config, w, err, http.StatusBadRequest)
				return
			}

			// Use the > sign to indicate the newest value should be used.
			r.Form.Set(fmt.Sprintf("swid>%s", t), r.Form.Get("swid"))
			r.Form.Del("swid")
		} else {
			swid, err := createSWID(s, r)
			if err != nil {
				returnServerError(&s.config, w, err)
				return
			}

			// Use the < sign to indicate the oldest, or existing value should
			// be used.
			r.Form.Set(fmt.Sprintf("swid<%s", t), swid.AsString())
		}
		if r.Form.Get("pref") != "" {
			err = validateOWID(s, &r.Form, "pref")
			if err != nil {
				returnAPIError(&s.config, w, err, http.StatusBadRequest)
				return
			}
			r.Form.Set(fmt.Sprintf("pref>%s", t), r.Form.Get("pref"))
			r.Form.Del("pref")
		}
		if r.Form.Get("email") != "" {
			err = validateOWID(s, &r.Form, "email")
			if err != nil {
				returnAPIError(&s.config, w, err, http.StatusBadRequest)
				return
			}
			r.Form.Set(fmt.Sprintf("email>%s", t), r.Form.Get("email"))
			r.Form.Del("email")
		}
		if r.Form.Get("salt") != "" {
			err = validateOWID(s, &r.Form, "salt")
			if err != nil {
				returnAPIError(&s.config, w, err, http.StatusBadRequest)
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
			returnAPIError(&s.config, w, err, http.StatusBadRequest)
			return
		}

		// Return the URL from the SWIFT layer.
		sendResponse(s, w, "text/plain; charset=utf-8", []byte(u))
	}
}

func validateOWID(s *services, q *url.Values, k string) error {
	o, err := owid.FromForm(q, k)
	if err != nil {
		return err
	}
	b, err := o.Verify(s.config.Scheme)
	if err != nil {
		return err
	}
	if b == false {
		return fmt.Errorf("'%s' not a verified OWID", k)
	}
	return nil
}
