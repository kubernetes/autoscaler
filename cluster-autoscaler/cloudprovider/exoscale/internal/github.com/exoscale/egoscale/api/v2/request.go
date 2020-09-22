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

package v2

import (
	"context"
	"fmt"
	"net/http"
)

const APIPrefix = "v2.alpha"

const defaultReqEndpointEnv = "api"

// ReqEndpoint represents an Exoscale API request endpoint.
type ReqEndpoint struct {
	env  string
	zone string
}

// NewReqEndpoint returns a new Exoscale API request endpoint from an environment and zone.
func NewReqEndpoint(env, zone string) ReqEndpoint {
	var re = ReqEndpoint{
		env:  env,
		zone: zone,
	}

	if re.env == "" {
		re.env = defaultReqEndpointEnv
	}

	return re
}

// Env returns the Exoscale API endpoint environment.
func (r *ReqEndpoint) Env() string {
	return r.env
}

// Zone returns the Exoscale API endpoint zone.
func (r *ReqEndpoint) Zone() string {
	return r.zone
}

// Host returns the Exoscale API endpoint host FQDN.
func (r *ReqEndpoint) Host() string {
	return fmt.Sprintf("%s-%s.exoscale.com", r.env, r.zone)
}

// WithEndpoint returns an augmented context instance containing the Exoscale endpoint to send
// the request to.
func WithEndpoint(ctx context.Context, endpoint ReqEndpoint) context.Context {
	return context.WithValue(ctx, ReqEndpoint{}, endpoint)
}

// WithZone is a shorthand to WithEndpoint where only the endpoint zone has to be specified.
// If a request endpoint is already set in the specified context instance, the value currently
// set for the environment will be reused.
func WithZone(ctx context.Context, zone string) context.Context {
	var env string

	if v, ok := ctx.Value(ReqEndpoint{}).(ReqEndpoint); ok {
		env = v.Env()
	}

	return WithEndpoint(ctx, NewReqEndpoint(env, zone))
}

// SetEndpointFromContext is an HTTP client request interceptor that overrides the "Host" header
// with information from a request endpoint optionally set in the context instance. If none is
// found, the request is left untouched.
func SetEndpointFromContext(ctx context.Context, req *http.Request) error {
	if v, ok := ctx.Value(ReqEndpoint{}).(ReqEndpoint); ok {
		req.Host = v.Host()
		req.URL.Host = v.Host()
	}

	return nil
}
