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

package protocol

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
)

// HostPrefixHandlerName is the handler name for the host prefix request
// handler.
const HostPrefixHandlerName = "volcenginesdk.endpoint.HostPrefixHandler"

// NewHostPrefixHandler constructs a build handler
func NewHostPrefixHandler(prefix string, labelsFn func() map[string]string) request.NamedHandler {
	builder := HostPrefixBuilder{
		Prefix:   prefix,
		LabelsFn: labelsFn,
	}

	return request.NamedHandler{
		Name: HostPrefixHandlerName,
		Fn:   builder.Build,
	}
}

// HostPrefixBuilder provides the request handler to expand and prepend
// the host prefix into the operation's request endpoint host.
type HostPrefixBuilder struct {
	Prefix   string
	LabelsFn func() map[string]string
}

// Build updates the passed in Request with the HostPrefix template expanded.
func (h HostPrefixBuilder) Build(r *request.Request) {
	//if volcengine.BoolValue(r.Config.DisableEndpointHostPrefix) {
	//	return
	//}

	var labels map[string]string
	if h.LabelsFn != nil {
		labels = h.LabelsFn()
	}

	prefix := h.Prefix
	for name, value := range labels {
		prefix = strings.Replace(prefix, "{"+name+"}", value, -1)
	}

	r.HTTPRequest.URL.Host = prefix + r.HTTPRequest.URL.Host
	if len(r.HTTPRequest.Host) > 0 {
		r.HTTPRequest.Host = prefix + r.HTTPRequest.Host
	}
}
