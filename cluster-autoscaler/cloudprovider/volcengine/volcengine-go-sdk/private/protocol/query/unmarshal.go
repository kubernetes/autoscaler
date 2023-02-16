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

package query

// This File is modify from https://github.com/aws/aws-sdk-go/blob/main/private/protocol/query/unmarshal.go

import (
	"encoding/xml"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/private/protocol/xml/xmlutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
)

// UnmarshalHandler is a named request handler for unmarshaling volcenginequery protocol requests
var UnmarshalHandler = request.NamedHandler{Name: "awssdk.volcenginequery.Unmarshal", Fn: Unmarshal}

// UnmarshalMetaHandler is a named request handler for unmarshaling volcenginequery protocol request metadata
var UnmarshalMetaHandler = request.NamedHandler{Name: "awssdk.volcenginequery.UnmarshalMeta", Fn: UnmarshalMeta}

// Unmarshal unmarshals a response for an VOLCSTACK Query service.
func Unmarshal(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	if r.DataFilled() {
		decoder := xml.NewDecoder(r.HTTPResponse.Body)
		err := xmlutil.UnmarshalXML(r.Data, decoder, r.Operation.Name+"Result")
		if err != nil {
			r.Error = volcengineerr.NewRequestFailure(
				volcengineerr.New(request.ErrCodeSerialization, "failed decoding Query response", err),
				r.HTTPResponse.StatusCode,
				r.RequestID,
			)
			return
		}
	}
}

// UnmarshalMeta unmarshals header response values for an VOLCSTACK Query service.
func UnmarshalMeta(r *request.Request) {
	r.RequestID = r.HTTPResponse.Header.Get("X-Top-Requestid")
}
