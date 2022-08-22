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

// kamateraAPIClient is the interface used to call kamatera API
type kamateraAPIClient interface {
	ListServers(ctx context.Context, instances map[string]*Instance) ([]Server, error)
	DeleteServer(ctx context.Context, name string) error
	CreateServers(ctx context.Context, count int, config ServerConfig) ([]Server, error)
}

// buildKamateraAPIClient returns the struct ready to perform calls to kamatera API
func buildKamateraAPIClient(clientId string, secret string, url string) kamateraAPIClient {
	client := NewKamateraApiClientRest(clientId, secret, url)
	return &client
}
