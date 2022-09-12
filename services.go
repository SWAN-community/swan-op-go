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
	"github.com/SWAN-community/access-go"
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swift-go"
)

// HTTP headers that if present indicate a request is probably from a web
// browser.
var invalidHTTPHeaders = []string{
	"Accept",
	"Accept-Language",
	"Cookie"}

// Services references all the information needed for every method.
type services struct {
	config Configuration
	swift  *swift.Services // Services used by the SWIFT network
	owid   *owid.Services  // Services used for OWID creation and verification
	access access.Access   // Instance of access service
}

// newServices a set of services to use with SWAN. These provide defaults via
// the configuration parameter, and access to persistent storage via the store
// parameter.
func newServices(settingsFile string, swanAccess access.Access) *services {
	var swiftStores []swift.Store
	var owidStore owid.Store

	// Use the file provided to get the SWIFT settings.
	swiftConfig := swift.NewConfig(settingsFile)
	err := swiftConfig.Validate()
	if err != nil {
		panic(err)
	}

	// Use the file provided to get the OWID settings.
	owidConfig := owid.NewConfig(settingsFile)

	// Link to the SWIFT storage.
	swiftStores = swift.NewStore(swiftConfig)

	swiftStoreSvc := swift.NewStorageService(swiftConfig, swiftStores...)

	// Link to the OWID storage.
	owidStore = owid.NewStore(&owidConfig)

	// Get the default browser detector.
	b, err := swift.NewBrowserRegexes()
	if err != nil {
		panic(err)
	}

	// Create the swan configuration.
	c := newConfig(settingsFile)

	// Return the services.
	return &services{
		c,
		swift.NewServices(swiftConfig, swiftStoreSvc, swanAccess, b),
		owid.NewServices(&owidConfig, owidStore, swanAccess),
		swanAccess}
}
