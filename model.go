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

// UnmarshalSwift adds the SWIFT pairs to the data model and also creates SID
// and RID values if not provided from SWIFT.
func (m *Response) UnmarshalSwift(
	r *swift.Results,
	s *services,
	q *http.Request) error {
	err := m.ModelResponse.UnmarshalSwift(r)
	if err != nil {
		return err
	}
	m.newRID(s, q)
	if err != nil {
		return err
	}
	m.setSID(s, q)
	if err != nil {
		return err
	}
	return nil
}

// newRID sets a new RID in the model and return true if successful, otherwise
// false. All the HTTP errors are handled by the implementation.
func (m *Response) newRID(s *services, r *http.Request) error {
	var err error
	m.RID, err = createRID(s, r)
	if err != nil {
		return err
	}
	m.RID.Cookie.Expires = m.RID.GetOWID().TimeStamp.Add(
		time.Duration(s.config.DeleteDays) * time.Hour * 24)
	return nil
}

// setSID uses the Email and Salt to populate the SID data.
func (m *Response) setSID(s *services, r *http.Request) error {
	if m.Email.Email != "" && m.Salt.Salt != nil {
		g, err := s.owid.GetSigner(r.Host)
		if err != nil {
			return err
		}
		m.SID, err = swan.NewSID(g, m.Email, m.Salt)
		if err != nil {
			return err
		}
		m.SID.GetCookie().Expires = m.SID.GetOWID().TimeStamp.Add(
			time.Duration(s.config.DeleteDays) * time.Hour * 24)
	}
	return nil
}
