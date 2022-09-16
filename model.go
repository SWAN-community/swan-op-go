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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
)

type Validity struct {
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
}

type Identifier struct {
	swan.Identifier
	Validity
}

type Preferences struct {
	swan.Preferences
	Validity
}

type Email struct {
	swan.Email
	Validity
}

type Salt struct {
	swan.Salt
	Validity
}

type ByteArray struct {
	swan.ByteArray
	Validity
}

type StringArray struct {
	Value []string
	Validity
}

type entry struct {
	key      string
	validity *Validity
	owid     *owid.OWID
}

// Model used when request or responding with SWAN data.
type Model struct {
	RID   *Identifier  `json:"rid,omitempty"`
	Pref  *Preferences `json:"pref,omitempty"`
	Email *Email       `json:"email,omitempty"`
	Salt  *Salt        `json:"salt,omitempty"`
	Stop  *StringArray `json:"stop,omitempty"`
	State []string     `json:"state,omitempty"`
}

// Extension to Model with information needed in a response.
type Response struct {
	Model
	SID *ByteArray `json:"sid,omitempty"`
	Val Validity   `json:"val,omitempty"`
}

// Extension to Model with information needed in a request.
type Request struct {
	Model
}

func (m *Request) UnmarshalRequest(r *http.Request) error {
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Response) UnmarshalSwift(r *swift.Results) error {

	// Set the fields that are also fields in the SWIFT results.
	m.State = r.State

	// Unpack or copy the SWIFT key value pairs that the model knows about.
	for _, v := range r.Pairs {
		switch v.Key() {
		case "rid":
			m.RID = &Identifier{}
			err := m.RID.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		case "email":
			m.Email = &Email{}
			err := m.Email.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		case "salt":
			m.Salt = &Salt{}
			err := m.Salt.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		case "pref":
			m.Pref = &Preferences{}
			err := m.Pref.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		case "sid":
			m.SID = &ByteArray{}
			err := m.SID.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		case "stop":
			m.Stop = &StringArray{}
			err := m.Stop.UnmarshalSwift(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Model) Verify(s *services) error {
	for _, v := range m.getEntries() {
		ok, err := v.owid.Verify(s.config.Scheme)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s invalid", v.key)
		}
	}
	return nil
}

func (v *Validity) UnmarshalSwiftValidity(p *swift.Pair) error {
	v.Created = p.Created()
	v.Expires = p.Expires()
	return nil
}

func (s *StringArray) UnmarshalSwift(p *swift.Pair) error {
	s.Value = make([]string, 0, len(p.Values()))
	for _, v := range p.Values() {
		if len(v) > 0 {
			s.Value = append(s.Value, string(v))
		}
	}
	return s.UnmarshalSwiftValidity(p)
}

func (e *Email) UnmarshalSwift(p *swift.Pair) error {
	if len(p.Values()) == 0 {
		return nil
	}
	err := e.UnmarshalBase64(p.Values()[0])
	if err != nil {
		return err
	}
	return e.UnmarshalSwiftValidity(p)
}

func (i *Identifier) UnmarshalSwift(p *swift.Pair) error {
	if len(p.Values()) == 0 {
		return nil
	}
	err := i.UnmarshalBase64(p.Values()[0])
	if err != nil {
		return err
	}
	return i.UnmarshalSwiftValidity(p)
}

func (f *Preferences) UnmarshalSwift(p *swift.Pair) error {
	if len(p.Values()) == 0 {
		return nil
	}
	err := f.UnmarshalBase64(p.Values()[0])
	if err != nil {
		return err
	}
	return f.UnmarshalSwiftValidity(p)
}

func (s *Salt) UnmarshalSwift(p *swift.Pair) error {
	if len(p.Values()) == 0 {
		return nil
	}
	err := s.UnmarshalBase64(p.Values()[0])
	if err != nil {
		return err
	}
	return s.UnmarshalSwiftValidity(p)
}

func (b *ByteArray) UnmarshalSwift(p *swift.Pair) error {
	if len(p.Values()) == 0 {
		return nil
	}
	err := b.UnmarshalBase64(p.Values()[0])
	if err != nil {
		return err
	}
	return b.UnmarshalSwiftValidity(p)
}

func (m *Response) getEntries() []*entry {
	i := m.Model.getEntries()
	if m.SID != nil {
		i = append(i, &entry{
			key:      "sid",
			owid:     m.SID.GetOWID(),
			validity: &m.SID.Validity})
	}
	return i
}

func (m *Model) getEntries() []*entry {
	i := make([]*entry, 0, 5)
	if m.Email != nil {
		i = append(i, &entry{
			key:      "email",
			owid:     m.Email.GetOWID(),
			validity: &m.Email.Validity})
	}
	if m.Pref != nil {
		i = append(i, &entry{
			key:      "pref",
			owid:     m.Pref.GetOWID(),
			validity: &m.Pref.Validity})
	}
	if m.Salt != nil {
		i = append(i, &entry{
			key:      "salt",
			owid:     m.Salt.GetOWID(),
			validity: &m.Salt.Validity})
	}
	if m.RID != nil {
		i = append(i, &entry{
			key:      "rid",
			owid:     m.RID.GetOWID(),
			validity: &m.RID.Validity})
	}
	return i
}

// SetValidity sets the created and expires times. This is used by the caller to
// indicate when they should recheck the returned data with the SWAN network
// for updates.
func (m *Response) SetValidity(s *services) error {
	m.Val.Created = time.Now().UTC()
	m.Val.Expires = m.Val.Created.Add(s.config.RevalidateSecondsDuration())
	for _, v := range m.getEntries() {
		if v.validity.Expires.Before(m.Val.Expires) {
			m.Val.Expires = v.validity.Expires
		}
	}
	return nil
}

// newRID sets a new RID in the model and return true if successful, otherwise
// false. All the HTTP errors are handled by the implementation.
func (m *Model) newRID(s *services, r *http.Request) error {
	i, err := createRID(s, r)
	if err != nil {
		return err
	}
	m.RID = &Identifier{}
	m.RID.Identifier = *i
	m.RID.Created = i.OWID.TimeStamp
	m.RID.Expires = i.OWID.TimeStamp.Add(
		time.Duration(s.config.DeleteDays) * time.Hour * 24)
	return nil
}

// setSID uses the Email and Salt to populate the SID data.
func (m *Response) setSID(s *services, r *http.Request) error {
	if m.Email.Email.Email != "" && m.Salt.Salt.Salt != nil {
		g, err := s.owid.GetSigner(r.Host)
		if err != nil {
			return err
		}
		i, err := swan.NewSID(g, &m.Email.Email, &m.Salt.Salt)
		if err != nil {
			return err
		}
		m.SID = &ByteArray{}
		m.SID.ByteArray = *i
		m.SID.Created = i.OWID.TimeStamp
		m.SID.Expires = i.OWID.TimeStamp.Add(
			time.Duration(s.config.DeleteDays) * time.Hour * 24)
	}
	return nil
}