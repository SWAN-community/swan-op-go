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
	"net/url"
	"testing"

	"github.com/SWAN-community/common-go"
	"github.com/SWAN-community/swan-go"
)

func TestCreateRID(t *testing.T) {
	s := getServices(t)
	values := url.Values{}
	values.Set("accessKey", testAccessKey)
	rr := common.HTTPTest(
		t,
		"GET",
		testDomain,
		"/swan/api/v1/create-rid",
		values,
		handlerCreateRID(s))
	if rr.Code != http.StatusOK {
		t.Fatal("status not ok")
	}
	i := &swan.Identifier{}
	err := i.UnmarshalBinary(common.ResponseAsByteArrayTest(t, rr))
	if err != nil {
		t.Fatal(err)
	}
	g, err := s.owid.GetSigner(testDomain)
	if err != nil {
		t.Fatal(err)
	}
	v, err := g.Verify(i.OWID)
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Fatal("verification failed")
	}
}
