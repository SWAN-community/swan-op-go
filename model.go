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
	"time"

	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
)

// Extension to Model with information needed in a response.
type Response struct {
	swan.ModelResponse
}

// Extension to Model with information needed in a request.
type Request struct {
	swan.ModelRequest
}

// UnmarshalSwift adds the SWIFT pairs to the data model and also creates the
// SID from the email and salt if not contained in the SWIFT results.
func (m *Response) UnmarshalSwift(
	r *swift.Results,
	s *services,
	q *http.Request) error {
	err := m.ModelResponse.UnmarshalSwift(r)
	if err != nil {
		return err
	}
	m.setSID(s, q)
	if err != nil {
		return err
	}
	return nil
}

// setSID uses the Email and Salt to populate the SID data if they are present
// and contain valid data.
func (m *Response) setSID(s *services, r *http.Request) error {
	if m.Email != nil &&
		m.Email.Email != "" &&
		m.Salt != nil &&
		len(m.Salt.Salt) > 0 {
		g, err := s.owid.GetSigner(r.Host)
		if err != nil {
			return err
		}
		m.SID, err = swan.NewSID(g, m.Email, m.Salt)
		if err != nil {
			return err
		}
		m.SID.GetCookie().Expires = getExpires(s, m.SID.GetOWID())
	}
	return nil
}

// getExpires returns the expiry date for the OWID. Used with the cookie to
// tell the browser when it should be removed.
func getExpires(s *services, o *owid.OWID) time.Time {
	return o.TimeStamp.Add(time.Duration(s.config.DeleteDays) * time.Hour * 24)
}
