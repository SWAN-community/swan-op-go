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
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
)

func TestResponse(t *testing.T) {

	// Setup the services.
	s := getServices(t)
	g, err := s.owid.GetSigner(testDomain)
	if err != nil {
		t.Fatal(err)
	}

	// Create the data to use with the test.
	rid, err := createRID(s, &http.Request{Host: testDomain})
	if err != nil {
		t.Fatal(err)
	}
	rid64, err := rid.MarshalBase64()
	if err != nil {
		t.Fatal(err)
	}
	email, err := swan.NewEmail(g, testEmail)
	if err != nil {
		t.Fatal(err)
	}
	email64, err := email.MarshalBase64()
	if err != nil {
		t.Fatal(err)
	}
	salt, err := swan.NewSaltFromString(g, testSalt)
	if err != nil {
		t.Fatal(err)
	}
	salt64, err := salt.MarshalBase64()
	if err != nil {
		t.Fatal(err)
	}
	pref, err := swan.NewPreferences(g, true)
	if err != nil {
		t.Fatal(err)
	}
	pref64, err := pref.MarshalBase64()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("pass", func(t *testing.T) {
		m := &Response{}
		r := &swift.Results{}
		responseAddPair(r, "rid", [][]byte{rid64})
		responseAddPair(r, "pref", [][]byte{pref64})
		responseAddPair(r, "salt", [][]byte{salt64})
		responseAddPair(r, "email", [][]byte{email64})
		err := m.UnmarshalSwift(r)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(m.RID.OWID.Signature, rid.OWID.Signature) {
			t.Fatal("rid")
		}
		if !bytes.Equal(m.Pref.OWID.Signature, pref.OWID.Signature) {
			t.Fatal("pref")
		}
		if !bytes.Equal(m.Email.OWID.Signature, email.OWID.Signature) {
			t.Fatal("email")
		}
		if !bytes.Equal(m.Salt.OWID.Signature, salt.OWID.Signature) {
			t.Fatal("salt")
		}
		err = m.setSID(s, &http.Request{Host: testDomain})
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetValidity(s)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(b))
	})
}

func responseAddPair(r *swift.Results, key string, value [][]byte) {
	t := time.Now().UTC()
	p := &swift.Pair{}
	p.SetKey(key)
	p.SetValues(value)
	p.SetCreated(t)
	p.SetExpires(t.Add(time.Hour))
	r.Pairs = append(r.Pairs, p)
}
