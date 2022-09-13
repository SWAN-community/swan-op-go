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
	"strings"
	"time"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
)

// Seperator used for an array of string values.
const listSeparator = " "

// The time format to use when adding the validation time to the response.
const ValidationTimeFormat = "2006-01-02T15:04:05Z07:00"

// handlerDecryptRawAsJSON returns the original data held in the the operation.
// Used by user interfaces to get the operations details for dispaly, or to
// continue a storage operation after time has passed waiting for the user.
// This method should never be used for passing for purposes other than for
// users editing their data.
func handlerDecryptRawAsJSON(s *services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the SWIFT results from the request.
		o := getResults(s, w, r)
		if o == nil {
			return
		}

		// Create a map of key value pairs.
		p := make(map[string]interface{})

		// Unpack or copy the SWIFT key value pairs to the map.
		for _, v := range o.Pairs() {
			switch v.Key() {
			case "rid":
				// RID does not get the OWID removed. It's is copied.
				i, err := pairToIdentifier(v)
				if err != nil {
					common.ReturnServerError(w, err)
					return
				}
				if i != nil {
					p[v.Key()], err = i.MarshalBase64()
					if err != nil {
						common.ReturnServerError(w, err)
						return
					}
				}
			case "email":
				// Email is unpacked so that the original value can be
				// displayed.
				e, err := pairToEmail(v)
				if err != nil {
					common.ReturnServerError(w, err)
					return
				}
				if e != nil {
					p[v.Key()] = e.Email
				}
			case "salt":
				// Salt is unpacked so that the email can be hashed. The payload
				// of the OWID is the salt as a base 64 string originally
				// returned from the salt-js JavaScript.
				s, err := pairToSalt(v)
				if err != nil {
					common.ReturnServerError(w, err)
					return
				}
				if s != nil {
					p[v.Key()] = fmt.Sprintf("%d", s.Salt)
				} else {
					p[v.Key()] = ""
				}
			case "pref":
				// Allow preferences are unpacked so that the original value can
				// be displayed.
				f, err := pairToPreferences(v)
				if err != nil {
					common.ReturnServerError(w, err)
					return
				}
				if f != nil {
					p[v.Key()] = f.Data.UseBrowsingForPersonalization
				}
			}
		}

		// If there is no valid RID create a new one.
		if p["rid"] == nil {
			var err error
			o := createRID(s, w, r)
			if o == nil {
				return
			}
			p["rid"], err = o.MarshalBase64()
			if err != nil {
				common.ReturnServerError(w, err)
				return
			}
		}

		// Set the values needed by the UIP to continue the operation.
		p["title"] = o.HTML.Title
		p["backgroundColor"] = o.HTML.BackgroundColor
		p["messageColor"] = o.HTML.MessageColor
		p["progressColor"] = o.HTML.ProgressColor
		p["message"] = o.HTML.Message
		p["state"] = o.State()

		// Turn the map of Raw SWAN data into a JSON string.
		j, err := json.Marshal(p)
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

		// Get the SWIFT results from the request.
		o := getResults(s, w, r)
		if o == nil {
			return
		}

		// Copy the key value pairs from SWIFT to SWAN. This is needed to
		// turn the email into a SID, and to convert the stopped domains from
		// byte arrays to a single string.
		v, err := convertPairs(s, r, o.Map())
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Turn the SWAN Pairs into a JSON string.
		j, err := json.Marshal(v)
		if err != nil {
			common.ReturnServerError(w, err)
			return
		}

		// Send the JSON string.
		common.SendJS(w, j)
	}
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

// verifyOWID confirms that the OWID provided has a valid signature.
func verifyOWID(s *services, o *owid.OWID) error {
	b, err := o.Verify(s.config.Scheme)
	if err != nil {
		return err
	}
	if !b {
		return fmt.Errorf("OWID failed verification")
	}
	return nil
}

// Copy the SWIFT results to the SWAN pairs. If the key is the email then it
// will be converted to a SID. An additional pair is written to contain the
// validation time in UTC. An error is returned if the SWIFT results are
// not usable.
func convertPairs(
	s *services,
	r *http.Request,
	p map[string]*swift.Pair) ([]*swan.Pair, error) {
	var m time.Time
	w := make([]*swan.Pair, 0, len(p)+1)

	for _, v := range p {
		// Turn the raw SWAN data into the SWAN data ready for readonly use.
		switch v.Key() {
		case "email":
			n := p["salt"]
			if n != nil && len(v.Values()) > 0 && len(n.Values()) > 0 {
				// Verify email
				e, err := pairToEmail(v)
				if err != nil {
					return nil, err
				}
				if s.config.Debug {
					err = verifyOWID(s, e.OWID)
					if err != nil {
						return nil, err
					}
				}
				// Verify salt
				t, err := pairToSalt(n)
				if err != nil {
					return nil, err
				}
				if s.config.Debug {
					err = verifyOWID(s, t.OWID)
					if err != nil {
						return nil, err
					}
				}
				// Get the SID from the email and the salt
				s, err := getSID(s, r, v, e, t)
				if err != nil {
					return nil, err
				}
				w = append(w, s)
			}
		case "salt":
			// Don't do anything with salt as we have used it when
			// creating the SID.
		case "pref":
			if len(v.Values()) > 0 {
				if s.config.Debug {
					f, err := pairToPreferences(v)
					if err != nil {
						return nil, err
					}
					err = verifyOWID(s, f.OWID)
					if err != nil {
						return nil, err
					}
				}
				w = append(w, copyValue(v))
			}
		case "rid":
			if len(v.Values()) > 0 {
				if s.config.Debug {
					i, err := pairToIdentifier(v)
					if err != nil {
						return nil, err
					}
					err = verifyOWID(s, i.OWID)
					if err != nil {
						return nil, err
					}
				}
				w = append(w, copyValue(v))
			}
		case "stop":
			s, err := getStopped(v)
			if err != nil {
				return nil, err
			}
			w = append(w, s)
		default:
			w = append(w, copyValue(v))
		}
	}

	// Find the expiry date furthest in the future. This will be used to set the
	// val pair to indicate the caller when they should recheck the network.
	for _, v := range w {
		if m.Before(v.Expires) {
			m = v.Expires
		}
	}

	// Add a final pair to indicate when the caller should revalidate the
	// SWAN data with the network. This is recommended for the caller, but not
	// compulsory.
	t := time.Now()
	e := t.Add(s.config.RevalidateSecondsDuration()).Format(
		ValidationTimeFormat)
	w = append(w, &swan.Pair{
		Key:     "val",
		Created: t,
		Expires: m,
		Value:   e,
	})
	return w, nil
}

// Converts the array of stopped values into a single string seperated by the
// listSeparator.
func getStopped(p *swift.Pair) (*swan.Pair, error) {
	s := make([]string, 0, len(p.Values()))
	for _, v := range p.Values() {
		if len(v) > 0 {
			s = append(s, string(v))
		}
	}
	return &swan.Pair{
		Key:     p.Key(),
		Created: p.Created(),
		Expires: p.Expires(),
		Value:   strings.Join(s, listSeparator)}, nil
}

// copyValue turns the SWIFT key and SWAN value into a SWAN pair.
func copyValue(p *swift.Pair) *swan.Pair {
	return &swan.Pair{
		Key:     p.Key(),
		Created: p.Created(),
		Expires: p.Expires(),
		Value:   p.Value()}
}

// getSID turns the email address that is contained in the Value OWID into
// a hashed version in a new OWID with this SWAN Operator as the creator.
func getSID(
	s *services,
	r *http.Request,
	emailPair *swift.Pair,
	email *swan.Email,
	salt *swan.Salt) (*swan.Pair, error) {
	v := &swan.Pair{
		Key:     "sid",
		Created: emailPair.Created(),
		Expires: emailPair.Expires()}
	if email != nil &&
		salt != nil {
		g, err := s.owid.GetSigner(r.Host)
		if err != nil {
			return nil, err
		}
		d, err := swan.NewSID(g, email, salt)
		if err != nil {
			return nil, err
		}
		b, err := d.MarshalBase64()
		if err != nil {
			return nil, err
		}
		v.Value = string(b)
	}
	return v, nil
}

func pairToEmail(pair *swift.Pair) (*swan.Email, error) {
	var e swan.Email
	if len(pair.Values()) == 0 {
		return nil, fmt.Errorf("no values")
	}
	err := e.UnmarshalBinary(pair.Values()[0])
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func pairToIdentifier(pair *swift.Pair) (*swan.Identifier, error) {
	var i swan.Identifier
	if len(pair.Values()) == 0 {
		return nil, fmt.Errorf("no values")
	}
	err := i.UnmarshalBinary(pair.Values()[0])
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func pairToPreferences(pair *swift.Pair) (*swan.Preferences, error) {
	var p swan.Preferences
	if len(pair.Values()) == 0 {
		return nil, fmt.Errorf("no values")
	}
	err := p.UnmarshalBinary(pair.Values()[0])
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func pairToSalt(pair *swift.Pair) (*swan.Salt, error) {
	var s swan.Salt
	if len(pair.Values()) == 0 {
		return nil, fmt.Errorf("no values")
	}
	err := s.UnmarshalBinary(pair.Values()[0])
	if err != nil {
		return nil, err
	}
	return &s, nil
}
