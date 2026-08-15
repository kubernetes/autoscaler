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
)

// ServerRequestType represents the type of request to be made to the Kamatera API
type ServerRequestType string

const (
	// ServerRequestPoweroff power off the server
	ServerRequestPoweroff ServerRequestType = "poweroff"
	// ServerRequestPoweron power on the server
	ServerRequestPoweron ServerRequestType = "poweron"
)

// CommandStatus represents the status of a command executed on Kamatera API
type CommandStatus int

const (
	// CommandStatusPending the command is still pending
	CommandStatusPending CommandStatus = 1
	// CommandStatusComplete the command is complete
	CommandStatusComplete CommandStatus = 2
	// CommandStatusError   the command ended with an error
	CommandStatusError CommandStatus = 3
)

// kamateraAPIClient is the interface used to call kamatera API
type kamateraAPIClient interface {
	ListServers(ctx context.Context, instances map[string]*Instance, namePrefix string, providerIDPrefix string) ([]Server, error)
	StartServerTerminate(ctx context.Context, name string, force bool) (string, error)
	StartServerRequest(ctx context.Context, requestType ServerRequestType, name string) (string, error)
	StartCreateServers(ctx context.Context, count int, config ServerConfig) (map[string]string, error)
	getServerTags(ctx context.Context, serverName string, instances map[string]*Instance, providerIDPrefix string) ([]string, error)
	getCommandStatus(ctx context.Context, commandID string) (CommandStatus, error)
}

// buildKamateraAPIClient returns the struct ready to perform calls to kamatera API
func buildKamateraAPIClient(clientId string, secret string, url string) kamateraAPIClient {
	client := NewKamateraApiClientRest(clientId, secret, url)
	return &client
}
