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
	"testing"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/swan-go"
	"github.com/SWAN-community/swift-go"
	"github.com/google/uuid"
)

func TestDecrypt(t *testing.T) {
	s := getServices(t)
	g, err := s.owid.GetSigner(testDomain)
	if err != nil {
		t.Fatal(err)
	}
	m := &Request{}
	m.Email, err = swan.NewEmail(g, testEmail)
	if err != nil {
		t.Fatal(err)
	}
	m.RID, err = swan.NewIdentifier(g, "paf_browser_id", uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	m.Salt, err = swan.NewSaltFromString(g, testSalt)
	if err != nil {
		t.Fatal(err)
	}
	m.Pref, err = swan.NewPreferences(g, false)
	if err != nil {
		t.Fatal(err)
	}
	u := getTestUpdateURL(t, s, g, m)
	rr := common.HTTPTest(
		t,
		http.MethodGet,
		u,
		nil,
		swift.HandlerStore(s.swift,
			func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("malformed")
			}))
	if rr.Code != http.StatusOK {
		t.Fatal("status not ok")
	}
	r := common.ResponseAsStringTest(t, rr)
	t.Log(r)
}
