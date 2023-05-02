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
	"io"
	"io/ioutil"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
)

// UnmarshalDiscardBodyHandler is a named request handler to empty and close a response's volcenginebody
var UnmarshalDiscardBodyHandler = request.NamedHandler{Name: "volcenginesdk.shared.UnmarshalDiscardBody", Fn: UnmarshalDiscardBody}

// UnmarshalDiscardBody is a request handler to empty a response's volcenginebody and closing it.
func UnmarshalDiscardBody(r *request.Request) {
	if r.HTTPResponse == nil || r.HTTPResponse.Body == nil {
		return
	}

	io.Copy(ioutil.Discard, r.HTTPResponse.Body)
	r.HTTPResponse.Body.Close()
}
