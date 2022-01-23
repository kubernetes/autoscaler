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

package common

import (
	"encoding/json"
	"fmt"

	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

const (
	octetStream = "application/octet-stream"
)

type actionParameters map[string]interface{}

type CommonRequest struct {
	*BaseRequest
	// custom header, may be overwritten
	header map[string]string
	actionParameters
}

func NewCommonRequest(service, version, action string) (request *CommonRequest) {
	request = &CommonRequest{
		BaseRequest:      &BaseRequest{},
		actionParameters: actionParameters{},
	}
	request.Init().WithApiInfo(service, version, action)
	return
}

// SetActionParameters set common request's actionParameters to your data.
// note: your data Must be a json-formatted string or byte array or map[string]interface{}
// note: you could not call SetActionParameters and SetOctetStreamParameters at once
func (cr *CommonRequest) SetActionParameters(data interface{}) error {
	if data == nil {
		return nil
	}
	switch data.(type) {
	case []byte:
		if err := json.Unmarshal(data.([]byte), &cr.actionParameters); err != nil {
			msg := fmt.Sprintf("Fail to parse contents %s to json, because: %s", data.([]byte), err)
			return tcerr.NewTencentCloudSDKError("ClientError.ParseJsonError", msg, "")
		}
	case string:
		if err := json.Unmarshal([]byte(data.(string)), &cr.actionParameters); err != nil {
			msg := fmt.Sprintf("Fail to parse contents %s to json, because: %s", data.(string), err)
			return tcerr.NewTencentCloudSDKError("ClientError.ParseJsonError", msg, "")
		}
	case map[string]interface{}:
		cr.actionParameters = data.(map[string]interface{})
	default:
		msg := fmt.Sprintf("Invalid data type:%T, must be one of the following: []byte, string, map[string]interface{}", data)
		return tcerr.NewTencentCloudSDKError("ClientError.InvalidParameter", msg, "")
	}
	return nil
}

func (cr *CommonRequest) IsOctetStream() bool {
	v, ok := cr.header["Content-Type"]
	if !ok || v != octetStream {
		return false
	}
	value, ok := cr.actionParameters["OctetStreamBody"]
	if !ok {
		return false
	}
	_, ok = value.([]byte)
	if !ok {
		return false
	}
	return true
}

func (cr *CommonRequest) SetHeader(header map[string]string) {
	if header == nil {
		return
	}
	cr.header = header
}

func (cr *CommonRequest) GetHeader() map[string]string {
	return cr.header
}

// SetOctetStreamParameters set request body to your data, and set head Content-Type to application/octet-stream
// note: you could not call SetActionParameters and SetOctetStreamParameters on the same request
func (cr *CommonRequest) SetOctetStreamParameters(header map[string]string, body []byte) {
	parameter := map[string]interface{}{}
	if header == nil {
		header = map[string]string{}
	}
	header["Content-Type"] = octetStream
	cr.header = header
	parameter["OctetStreamBody"] = body
	cr.actionParameters = parameter
}

func (cr *CommonRequest) GetOctetStreamBody() []byte {
	if cr.IsOctetStream() {
		return cr.actionParameters["OctetStreamBody"].([]byte)
	} else {
		return nil
	}
}

func (cr *CommonRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(cr.actionParameters)
}
