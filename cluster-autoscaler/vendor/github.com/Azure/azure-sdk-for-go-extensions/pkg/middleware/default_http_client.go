/*
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

package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/Azure/go-armbalancer"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/net/http2"
)

var (
	// defaultHTTPClient is configured with the defaultRoundTripper
	defaultHTTPClient *http.Client
	// defaultTransport is a pre-configured *http.Transport for http/2
	defaultTransport *http.Transport
	// defaultRoundTripper wraps the defaultTransport with arm balancer and otel propagation
	defaultRoundTripper http.RoundTripper
)

// DefaultHTTPClient returns a shared http client, and transport leveraging armbalancer for
// clientside loadbalancing, so we can leverage HTTP/2, and not get throttled by arm at the instance level.
func DefaultHTTPClient() *http.Client {
	return defaultHTTPClient
}

func init() {
	defaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// We call configureHttp2TransportPing() in the package init to ensure that our defaultTransport is always configured
	// with the http2 additional settings that work around the issue described here:
	// https://github.com/golang/go/issues/59690
	// azure sdk related issue is here:
	// https://github.com/Azure/azure-sdk-for-go/issues/21346#issuecomment-1699665586
	configureHttp2TransportPing(defaultTransport)
	defaultRoundTripper = armbalancer.New(armbalancer.Options{
		// PoolSize is the number of clientside http/2 persistent connections
		// we want to have configured in our transport. Note, that without clientside loadbalancing
		// with arm, HTTP/2 Will force persistent connection to stick to a single arm instance, and will
		// result in a substantial amount of throttling
		PoolSize:  100,
		Transport: defaultTransport,
	})

	defaultHTTPClient = &http.Client{
		Transport: otelhttp.NewTransport(
			defaultRoundTripper,
			otelhttp.WithPropagators(propagation.TraceContext{}),
		),
	}
}

// configureHttp2Transport ensures that our defaultTransport is configured
// with the http2 additional settings that work around the issue described here:
// https://github.com/golang/go/issues/59690
// azure sdk related issue is here:
// https://github.com/Azure/azure-sdk-for-go/issues/21346#issuecomment-1699665586
// This is called by the package init to ensure that our defaultTransport is always configured
// you should not call this anywhere else.
func configureHttp2TransportPing(tr *http.Transport) {
	// http2Transport holds a reference to the default transport and configures "h2" middlewares that
	// will use the below settings, making the standard http.Transport behave correctly for dropped connections
	http2Transport, err := http2.ConfigureTransports(tr)
	if err != nil {
		// by initializing in init(), we know it is only called once.
		// this errors if called twice.
		panic(err)
	}
	// we give 10s to the server to respond to the ping. if no response is received,
	// the transport will close the connection, so that the next request will open a new connection, and not
	// hit a context deadline exceeded error.
	http2Transport.PingTimeout = 10 * time.Second
	// if no frame is received for 30s, the transport will issue a ping health check to the server.
	http2Transport.ReadIdleTimeout = 30 * time.Second
}
