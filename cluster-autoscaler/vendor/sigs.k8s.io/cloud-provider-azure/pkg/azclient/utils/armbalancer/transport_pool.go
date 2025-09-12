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
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type transportChannPool struct {
	sync.WaitGroup
	capacity            chan struct{}
	pool                chan Transport
	transportFactory    func() Transport
	transportDropPolicy []TransportDropPolicy
}

type TransportDropPolicy interface {
	ShouldDropTransport(header http.Header) bool
}

type TransportDropPolicyFunc func(header http.Header) bool

func (function TransportDropPolicyFunc) ShouldDropTransport(header http.Header) bool {
	if function == nil {
		return false
	}
	return function(header)
}

type Transport interface {
	http.RoundTripper
	ForceClose() error
}

func newtransportChannPool(size int, transportFactory func() Transport, dropPolicy ...TransportDropPolicy) *transportChannPool {
	if size <= 0 {
		return nil
	}
	pool := &transportChannPool{
		capacity:            make(chan struct{}, size),
		pool:                make(chan Transport, size),
		transportFactory:    transportFactory,
		transportDropPolicy: dropPolicy,
	}
	return pool
}

func (pool *transportChannPool) Run(ctx context.Context) error {
CLEANUP:
	for {
		select {
		case <-ctx.Done():
			break CLEANUP
		case pool.capacity <- struct{}{}:
			pool.pool <- pool.transportFactory()
		}
	}

	//cleanup
	close(pool.capacity) // no more transport is added. consumers will be released if channel is closed.
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		pool.Wait()      // wait for transport recycle loop
		close(pool.pool) // no more transport is added consumers will released if channel is closed.
		return nil
	})
	for transport := range pool.pool {
		transport := transport
		errGroup.Go(transport.ForceClose)
	}
	return errGroup.Wait() // close all of transports in pool
}

func (pool *transportChannPool) RoundTrip(req *http.Request) (*http.Response, error) {
	transport, err := pool.selectTransport(req)
	if err != nil {
		return nil, err
	}
	resp, err := transport.RoundTrip(req)
	var header http.Header
	if resp != nil {
		header = resp.Header.Clone()
	}
	pool.Add(1)
	go pool.recycleTransport(transport, header)
	return resp, err
}

func (pool *transportChannPool) selectTransport(req *http.Request) (Transport, error) {
	for {
		var t Transport
		var ok bool
		select {
		case t, ok = <-pool.pool:
			if !ok {
				return nil, http.ErrServerClosed
			}
			return t, nil
		case <-req.Context().Done():
			return nil, http.ErrServerClosed
		}
	}
}

func (pool *transportChannPool) recycleTransport(t Transport, header http.Header) {
	defer pool.Done()
	for _, policy := range pool.transportDropPolicy {
		if policy.ShouldDropTransport(header) {
			t.ForceClose()  // drop the transport
			<-pool.capacity // notify pool to create new transport
			return
		}
	}
	pool.pool <- t
}

func (pool *transportChannPool) ForceClose() error {
	close(pool.pool)
	return nil
}

type KillBeforeThrottledPolicy struct {
	RecycleThreshold int64
}

func (policy *KillBeforeThrottledPolicy) ShouldDropTransport(header http.Header) bool {
	for key, vals := range header {
		if !strings.HasPrefix(key, "X-Ms-Ratelimit-Remaining-") {
			continue
		}
		n, err := strconv.ParseInt(vals[0], 10, 0)
		if err != nil {
			continue
		}
		if n < policy.RecycleThreshold {
			return true
		}
	}
	return false
}
