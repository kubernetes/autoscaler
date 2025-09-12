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

package armbalancer

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func NewHostScopedTransport(ctx context.Context, transportFactory func() *transportChannPool) Transport {
	ctx, cancelFn := context.WithCancel(ctx)
	transport := &hostScopedTransport{
		ctx:              ctx,
		cancelFn:         cancelFn,
		transportMap:     sync.Map{},
		transportFactory: transportFactory,
	}
	transport.serverGrp.SetLimit(-1)
	return transport
}

type hostScopedTransport struct {
	ctx              context.Context
	cancelFn         context.CancelFunc
	transportMap     sync.Map
	transportFactory func() *transportChannPool
	serverGrp        errgroup.Group
}

func (hostScopedTransport *hostScopedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transportRaw, ok := hostScopedTransport.transportMap.Load(strings.ToLower(req.Host))
	if !ok {
		transportPool := hostScopedTransport.transportFactory()
		hostScopedTransport.serverGrp.Go(func() error { return transportPool.Run(hostScopedTransport.ctx) })
		hostScopedTransport.transportMap.Store(strings.ToLower(req.Host), transportPool)
		transportRaw = transportPool
	}
	transport := transportRaw.(Transport)
	return transport.RoundTrip(req)
}

func (hostScopedTransport *hostScopedTransport) ForceClose() error {
	hostScopedTransport.cancelFn()
	return hostScopedTransport.serverGrp.Wait()
}
