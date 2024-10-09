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
)

type Options struct {
	Transport *http.Transport

	// PoolSize is the max number of connections that will be created by the connection pool.
	// Default: 8
	PoolSize int

	// RecycleThreshold is the lowest value of any X-Ms-Ratelimit-Remaining-* header that
	// can be seen before the associated connection will be re-established.
	// Default: 100
	RecycleThreshold int64

	// MinReqsBeforeRecycle is a safeguard to prevent frequent connection churn in the unlikely event
	// that a connections lands on an ARM instance that already has a depleted rate limiting quota.
	// Default: 10
	MinReqsBeforeRecycle int64
}

// New wraps a transport to provide smart connection pooling and client-side load balancing.
func New(ctx context.Context, opts Options) http.RoundTripper {
	if opts.Transport == nil {
		opts.Transport = http.DefaultTransport.(*http.Transport)
	}

	if opts.PoolSize == 0 {
		opts.PoolSize = 8
	}

	if opts.RecycleThreshold == 0 {
		opts.RecycleThreshold = 100
	}
	if opts.MinReqsBeforeRecycle == 0 {
		opts.MinReqsBeforeRecycle = 10
	}

	return NewHostScopedTransport(ctx, func() *transportChannPool {
		return newtransportChannPool(opts.PoolSize, func() Transport {
			return &ClosableTransport{
				Transport: opts.Transport.Clone(),
			}
		}, &KillBeforeThrottledPolicy{opts.RecycleThreshold})
	})
}
