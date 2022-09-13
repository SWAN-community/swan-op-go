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
	"time"

	"github.com/SWAN-community/config-go"
)

// Configuration maps to the appsettings.json settings file.
type Configuration struct {
	config.Base `mapstructure:",squash"`
	// Seconds until the value provided should be revalidated
	RevalidateSeconds int `mapstructure:"revalidateSeconds"`
	// The number of days after which the data will automatically be removed
	// from SWAN and will need to be provided again by the user.
	DeleteDays int `mapstructure:"deleteDays"`
}

// RevalidateSecondsDuration in seconds as a time.Duration
func (c *Configuration) RevalidateSecondsDuration() time.Duration {
	return time.Duration(c.RevalidateSeconds) * time.Second
}

// newConfig a new instance of configuration from the file provided.
func newConfig(file string) Configuration {
	var c Configuration
	err := config.LoadConfig([]string{"."}, file, &c)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Set defaults if they're not provided in the settings.
	if c.DeleteDays == 0 {
		c.DeleteDays = 90
	}
	if c.RevalidateSeconds == 0 {
		c.RevalidateSeconds = int(time.Hour.Seconds())
	}
	return c
}

// Gets the delete date for the SWAN data. This is the data after which the
// date will be removed from the network. Users will have to re-enter the data
// after this time.
func (c *Configuration) DeleteDate() time.Time {
	return time.Now().UTC().AddDate(0, 0, c.DeleteDays)
}
