/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"fmt"
	"strings"
)

// Use variables for easier test overload
var (
	systemConfigPath = "/etc/vke.conf"
	userConfigPath   = "/.vke.conf" // prefixed with homeDir
	localConfigPath  = "./vke.conf"
)

func (c *Client) loadConfig(endpointName string) error {
	// Load real endpoint URL by name. If endpoint contains a '/', consider it as a URL
	if strings.Contains(endpointName, "/") {
		c.endpoint = endpointName
	} else {
		c.endpoint = Endpoints[endpointName]
	}

	// If we still have no valid endpoint, AppKey or AppSecret, return an error
	if c.endpoint == "" {
		return fmt.Errorf("unknown endpoint '%s', consider checking 'Endpoints' list of using an URL", endpointName)
	}
	if c.AppKey == "" {
		return fmt.Errorf("missing application key, please check your configuration or consult the documentation to create one")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("missing application secret, please check your configuration or consult the documentation to create one")
	}

	return nil
}
