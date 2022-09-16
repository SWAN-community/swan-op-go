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
	"net/url"
	"testing"

	"github.com/SWAN-community/common-go"
)

// TestFetch confirms that a URL is returned. It does not test that the URL can
// be used with a SWIFT node to retrieve encrypted data.
func TestFetch(t *testing.T) {
	getTestFetchURL(t, getServices(t))
}

func getTestFetchURL(t *testing.T, s *services) *url.URL {
	u, err := url.Parse(fmt.Sprintf(
		"%s://%s/swan/api/v1/fetch",
		s.config.Scheme,
		testDomain))
	if err != nil {
		t.Fatal(err)
	}
	q := url.Values{}
	q.Set("accessKey", testAccessKey)
	q.Set("returnUrl", s.config.Scheme+"://"+testDomain)
	u.RawQuery = q.Encode()
	rr := common.HTTPTest(
		t,
		http.MethodGet,
		u,
		nil,
		handlerFetch(s))
	if rr.Code != http.StatusOK {
		t.Fatal("status not ok")
	}
	i, err := url.ParseRequestURI(common.ResponseAsStringTest(t, rr))
	if err != nil {
		t.Fatal(err)
	}
	return i
}
