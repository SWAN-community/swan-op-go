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

	"github.com/SWAN-community/swan-go"

	"github.com/google/uuid"
)

// createRID returns a swan.Identifier with a new unique RID.
func createRID(s *services, r *http.Request) (*swan.Identifier, error) {
	g, err := s.owid.GetSigner(r.Host)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("host '%s' unknown signer", r.Host)
	}
	i, err := swan.NewIdentifier(g, "paf_browser_id", uuid.New())
	if err != nil {
		return nil, err
	}
	return i, nil
}
