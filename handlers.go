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

	"github.com/SWAN-community/access-go"
	"github.com/SWAN-community/owid-go"
	"github.com/SWAN-community/swift-go"
)

// AddHandlers adds swift and owid end points configured from the JSON file
// provided.
// settingsFile path to the file that contains the configuration settings.
// swanAccess an authorization instance used to valid requests.
// malformedHandler if SWAN can't handle the request the handler to use instead.
func AddHandlers(
	settingsFile string,
	swanAccess access.Access,
	malformedHandler func(w http.ResponseWriter, r *http.Request)) error {

	// Create the new set of services.
	s := newServices(settingsFile, swanAccess)

	// Add the SWIFT handlers.
	swift.AddHandlers(s.swift, malformedHandler)

	// Add the OWID handlers.
	owid.AddHandlers(s.owid)

	// Add the SWAN handlers.
	http.HandleFunc("/swan/api/v1/fetch", handlerFetch(s))
	http.HandleFunc("/swan/api/v1/update", handlerUpdate(s))
	http.HandleFunc("/swan/api/v1/stop", handlerStop(s))
	http.HandleFunc("/swan/api/v1/home-node", handlerHomeNode(s))
	http.HandleFunc("/swan/api/v1/decrypt", handlerDecryptAsJSON(s))
	http.HandleFunc("/swan/api/v1/decrypt-raw", handlerDecryptRawAsJSON(s))
	http.HandleFunc("/swan/api/v1/create-rid", handlerCreateRID(s))
	http.HandleFunc("/health", handlerHealth(s))
	return nil
}
