/*
Copyright 2016 The Kubernetes Authors.

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

package kamatera

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"strings"
)

func mockKamateraServerName() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func mockServerConfig(namePrefix string, tags []string) ServerConfig {
	return ServerConfig{
		NamePrefix:     namePrefix,
		Password:       "",
		SshKey:         "",
		Datacenter:     "IL",
		Image:          "ubuntu_server_18.04_64-bit",
		Cpu:            "1A",
		Ram:            "1024",
		Disks:          []string{"size=10"},
		Dailybackup:    false,
		Managed:        false,
		Networks:       []string{"name=wan,ip=auto"},
		BillingCycle:   "hourly",
		MonthlyPackage: "",
		ScriptFile:     "#!/bin/bash",
		UserdataFile:   "",
		Tags:           tags,
	}
}

type kamateraClientMock struct {
	mock.Mock
}

func (c *kamateraClientMock) SetBaseURL(baseURL string) {
	c.Called(baseURL)
}

func (c *kamateraClientMock) ListServers(ctx context.Context, instances map[string]*Instance) ([]Server, error) {
	args := c.Called(ctx, instances)
	return args.Get(0).([]Server), args.Error(1)
}

func (c *kamateraClientMock) CreateServers(ctx context.Context, count int, config ServerConfig) ([]Server, error) {
	args := c.Called(ctx, count, config)
	return args.Get(0).([]Server), args.Error(1)
}

func (c *kamateraClientMock) DeleteServer(ctx context.Context, id string) error {
	args := c.Called(ctx, id)
	return args.Error(0)
}
