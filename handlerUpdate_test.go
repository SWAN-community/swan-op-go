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
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swan-go"
)

func TestUpdate(t *testing.T) {
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
	getTestUpdateURL(t, s, g, m)
}

func getTestUpdateURL(
	t *testing.T,
	s *services,
	g *owid.Signer,
	m *Request) *url.URL {
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(fmt.Sprintf(
		"%s://%s/swan/api/v1/update",
		s.config.Scheme,
		testDomain))
	if err != nil {
		t.Fatal(err)
	}
	q := &url.Values{}
	q.Set("accessKey", testAccessKey)
	q.Set("returnUrl", s.config.Scheme+"://"+testDomain)
	u.RawQuery = q.Encode()
	rr := common.HTTPTest(
		t,
		http.MethodPost,
		u,
		bytes.NewReader(b),
		handlerUpdate(s))
	if rr.Code != http.StatusOK {
		t.Fatal("status not ok")
	}
	r, err := url.ParseRequestURI(common.ResponseAsStringTest(t, rr))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
