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
	"os"
	"testing"

	"github.com/SWAN-community/access-go"
	"github.com/SWAN-community/owid-go"
)

// Test values.
const (
	testDomain    = "swan.community"
	testName      = "Secure Web Addressability Network"
	testTermsUrl  = "https://swan.community"
	testAccessKey = "A"
	testSettings  = `{
		"message": "Test Message",
		"title": "Test Title",
		"storageOperationTimeout": 30,
		"homeNodeTimeout": 86400,
		"backgroundColor": "#f5f5f5",
		"messageColor": "darkslategray",
		"progressColor": "darkgreen",
		"scheme": "http",
		"nodeCount": 10,
		"debug": false,
		"alivePollingSeconds": 2,
		"storageManagerRefreshMinutes": 1,
		"maxStores": 100}`
)

// getServices creates a temporary settings file and then adds OWID and SWIFT
// nodes to the settings to enable a simple SWAN Operator for testing purposes.
func getServices(t *testing.T) *services {
	s := newServices(
		tempSettings(t),
		access.NewFixed([]string{testAccessKey}))
	owid.RegisterTestSigner(
		t,
		s.owid,
		testDomain,
		"GET",
		testName,
		testTermsUrl)
	g, err := s.owid.GetSigner(testDomain)
	if err != nil {
		t.Fatal(err)
	}
	if g.Domain != testDomain {
		t.Fatal("domain mismatch")
	}
	return s
}

// Creates a temporary settings file for the configuration.
func tempSettings(t *testing.T) string {
	m := t.TempDir()
	f := m + "\\appsettings.test.json"
	var c map[string]interface{}
	err := json.Unmarshal([]byte(testSettings), &c)
	if err != nil {
		t.Fatal(err)
	}
	c["owidFile"] = m + "\\owid.json"
	c["swiftFile"] = m + "\\swift.json"
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(f, b, 0666)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
