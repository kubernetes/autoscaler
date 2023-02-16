/*
Copyright 2023 The Kubernetes Authors.

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

package volcengine

import (
	"os"

	"k8s.io/klog/v2"
)

const (
	regionId  = "REGION_ID"
	accessKey = "ACCESS_KEY"
	secretKey = "SECRET_KEY"
	endpoint  = "ENDPOINT"

	defaultEndpoint = "open.volcengineapi.com"
)

type cloudConfig struct {
	regionId  string
	accessKey string
	secretKey string
	endpoint  string
}

func (c *cloudConfig) getRegion() string {
	return c.regionId
}

func (c *cloudConfig) getAccessKey() string {
	return c.accessKey
}

func (c *cloudConfig) getSecretKey() string {
	return c.secretKey
}

func (c *cloudConfig) getEndpoint() string {
	return c.endpoint
}

func (c *cloudConfig) validate() bool {
	if c.regionId == "" {
		c.regionId = os.Getenv(regionId)
	}

	if c.accessKey == "" {
		c.accessKey = os.Getenv(accessKey)
	}

	if c.secretKey == "" {
		c.secretKey = os.Getenv(secretKey)
	}

	if c.endpoint == "" {
		c.endpoint = os.Getenv(endpoint)
	}

	if c.endpoint == "" {
		c.endpoint = defaultEndpoint
	}

	if c.regionId == "" || c.accessKey == "" || c.secretKey == "" || c.endpoint == "" {
		klog.V(5).Infof("Failed to get RegionId:%s,AccessKey:%s,SecretKey:%s,Endpoint:%s from cloudConfig and Env\n", c.regionId, c.accessKey, c.secretKey, c.endpoint)
		return false
	}

	return true
}
