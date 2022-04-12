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
	systemConfigPath = "/etc/ovh.conf"
	userConfigPath   = "/.ovh.conf" // prefixed with homeDir
	localConfigPath  = "./ovh.conf"
)

// loadConfig loads client configuration from params, environments or configuration
// files (by order of decreasing precedence).
//
// loadConfig will check OVH_CONSUMER_KEY, OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET
// and OVH_ENDPOINT environment variables. If any is present, it will take precedence
// over any configuration from file.
//
// Configuration files are ini files. They share the same format as python-ovh,
// node-ovh, php-ovh and all other wrappers. If any wrapper is configured, all
// can re-use the same configuration. loadConfig will check for configuration in:
//
// - ./ovh.conf
// - $HOME/.ovh.conf
// - /etc/ovh.conf
//
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
