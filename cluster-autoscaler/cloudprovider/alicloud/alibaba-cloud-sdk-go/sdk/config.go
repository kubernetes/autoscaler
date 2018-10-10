/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/utils"
	"net/http"
	"time"
)

// Config is SDK client options
type Config struct {
	AutoRetry         bool            `default:"true"`
	MaxRetryTime      int             `default:"3"`
	UserAgent         string          `default:""`
	Debug             bool            `default:"false"`
	Timeout           time.Duration   `default:"10000000000"`
	HttpTransport     *http.Transport `default:""`
	EnableAsync       bool            `default:"false"`
	MaxTaskQueueSize  int             `default:"1000"`
	GoRoutinePoolSize int             `default:"5"`
	Scheme            string          `default:"HTTP"`
}

// NewConfig returns client config
func NewConfig() (config *Config) {
	config = &Config{}
	utils.InitStructWithDefaultTag(config)
	return
}

// WithTimeout set client timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithAutoRetry set client with retry
func (c *Config) WithAutoRetry(isAutoRetry bool) *Config {
	c.AutoRetry = isAutoRetry
	return c
}

// WithMaxRetryTime set client with max retry times
func (c *Config) WithMaxRetryTime(maxRetryTime int) *Config {
	c.MaxRetryTime = maxRetryTime
	return c
}

// WithUserAgent set client user agent
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// WithHttpTransport set client custom http transport
func (c *Config) WithHttpTransport(httpTransport *http.Transport) *Config {
	c.HttpTransport = httpTransport
	return c
}

// WithEnableAsync enable client async option
func (c *Config) WithEnableAsync(isEnableAsync bool) *Config {
	c.EnableAsync = isEnableAsync
	return c
}

// WithMaxTaskQueueSize set client max task queue size
func (c *Config) WithMaxTaskQueueSize(maxTaskQueueSize int) *Config {
	c.MaxTaskQueueSize = maxTaskQueueSize
	return c
}

// WithGoRoutinePoolSize set client go routine pool size
func (c *Config) WithGoRoutinePoolSize(goRoutinePoolSize int) *Config {
	c.GoRoutinePoolSize = goRoutinePoolSize
	return c
}

// WithDebug set client debug mode
func (c *Config) WithDebug(isDebug bool) *Config {
	c.Debug = isDebug
	return c
}
