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

package huaweicloudsdk

import (
	"net/http"
	"time"
)

// Config define the configs parameter of the provider client .
type Config struct {
	// Timeout specifies a time limit for requests made by this
	// Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	//
	// A Timeout of zero means no timeout.
	Timeout time.Duration `default:"10000000000"`
	// Transport specifies the mechanism by which individual
	// HTTP requests are made.
	// If nil, DefaultTransport is used.
	HttpTransport *http.Transport `default:""`
	//AutoRetry         bool            `default:"true"`
	//MaxRetryTime      int             `default:"3"`
	//UserAgent         string          `default:""`
	//EnableAsync       bool            `default:"false"`
	//MaxTaskQueueSize  int             `default:"1000"`
	//GoRoutinePoolSize int             `default:"5"`
}

//NewConfig return Config instance with its default tag.
func NewConfig() (config *Config) {
	config = &Config{}
	InitStructWithDefaultTag(config)
	return
}

// WithTimeout customizes the http connection timeout.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithHttpTransport customizes the http transport.
func (c *Config) WithHttpTransport(httpTransport *http.Transport) *Config {
	c.HttpTransport = httpTransport
	return c
}

/*
func (c *Config) WithAutoRetry(isAutoRetry bool) *Config {
	c.AutoRetry = isAutoRetry
	return c
}

func (c *Config) WithMaxRetryTime(maxRetryTime int) *Config {
	c.MaxRetryTime = maxRetryTime
	return c
}

func (c *Config) WithEnableAsync(isEnableAsync bool) *Config {
	c.EnableAsync = isEnableAsync
	return c
}

func (c *Config) WithMaxTaskQueueSize(maxTaskQueueSize int) *Config {
	c.MaxTaskQueueSize = maxTaskQueueSize
	return c
}

func (c *Config) WithGoRoutinePoolSize(goRoutinePoolSize int) *Config {
	c.GoRoutinePoolSize = goRoutinePoolSize
	return c
}
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

*/
