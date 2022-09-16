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
	"net/http"
	"os"
	"testing"

	"github.com/SWAN-community/access-go"
	"github.com/SWAN-community/owid-go"
)

// Test values.
const (
	testNetwork   = "swan"
	testEmail     = "test@" + testDomain
	testSalt      = "1234"
	testDomain    = "not.a.valid.domain"
	testName      = "Secure Web Addressability Network"
	testTermsUrl  = "https://" + testDomain
	testAccessKey = "A"
	testSettings  = `{
		"message": "Test Message",
		"title": "Test Title",
		"storageOperationTimeout": 30000,
		"homeNodeTimeout": 86400,
		"backgroundColor": "#f5f5f5",
		"messageColor": "darkslategray",
		"progressColor": "darkgreen",
		"scheme": "http",
		"nodeCount": 10,
		"debug": false,
		"alivePollingSeconds": 0,
		"storageManagerRefreshMinutes": 1,
		"maxStores": 100}`

	// SWIFT files don't have to contain encryption keys. Therefore just adding
	// the configuration file to the test folder is all that is needed to make
	// SWIFT work for SWAN Operator tests.
	testSwiftFile = `{
		"` + testDomain + `": {
			"cookieDomain": "` + testDomain + `",
			"created": "2022-01-01T00:00:00.0Z",
			"domain": "` + testDomain + `",
			"expires": "2052-01-01T00:00:00.0Z",
			"network": "` + testNetwork + `",
			"role": 0,
			"scrambler": "",
			"secrets": [],
			"starts": "2022-01-01T00:00:00.0Z"
		},
		"store.` + testDomain + `": {
			"cookieDomain": "store.` + testDomain + `",
			"created": "2022-01-01T00:00:00.0Z",
			"domain": "store.` + testDomain + `",
			"expires": "2052-01-01T00:00:00.0Z",
			"network": "` + testNetwork + `",
			"role": 1,
			"scrambler": "",
			"secrets": [],
			"starts": "2022-01-01T00:00:00.0Z"
		}
	}`
)

// getServices creates a temporary settings file and then adds OWID and SWIFT
// nodes to the settings to enable a simple SWAN Operator for testing purposes.
// The OWID signer must include private and public keys. If these were embedded
// into the source code then warnings will be triggered by GitHub concerning the
// presence of cryptographic information. Therefore the OWID signer is
// registered for each test using the test helper method RegisterTestSigner.
func getServices(t *testing.T) *services {
	s := newServices(
		tempSettings(t),
		access.NewFixed([]string{testAccessKey}))

	// Register the test domain as a valid OWID signer.
	owid.RegisterTestSigner(
		t,
		s.owid,
		testDomain,
		http.MethodGet,
		testName,
		testTermsUrl)
	g, err := s.owid.GetSigner(testDomain)
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		t.Fatalf("no signer for domain '%s'", testDomain)
	}
	if g.Domain != testDomain {
		t.Fatal("domain mismatch")
	}
	n, err := s.swift.GetAccessNodeForHost(testDomain)
	if err != nil {
		t.Fatal(err)
	}
	if n == nil {
		t.Fatal("no SWIFT node")
	}
	return s
}

// Creates a temporary settings file for the configuration and the associated
// test SWIFT nodes.
func tempSettings(t *testing.T) string {
	m := t.TempDir()
	f := m + "\\appsettings.test.json"
	s := m + "\\swift.json"

	// Create the content for the appsettings files.
	var c map[string]interface{}
	err := json.Unmarshal([]byte(testSettings), &c)
	if err != nil {
		t.Fatal(err)
	}
	c["debug"] = true
	c["owidFile"] = m + "\\owid.json"
	c["swiftFile"] = s
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	// Write the appsettings file.
	err = os.WriteFile(f, b, 0666)
	if err != nil {
		t.Fatal(err)
	}

	// Write the SWIFT note data file.
	err = os.WriteFile(s, []byte(testSwiftFile), 0666)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
