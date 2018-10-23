/*
Copyright 2018 The Kubernetes Authors.

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

package responses

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"net/http"
	"strings"
)

// AcsResponse interface
type AcsResponse interface {
	IsSuccess() bool
	GetHttpStatus() int
	GetHttpHeaders() map[string][]string
	GetHttpContentString() string
	GetHttpContentBytes() []byte
	GetOriginHttpResponse() *http.Response
	parseFromHttpResponse(httpResponse *http.Response) error
}

// Unmarshal return response body
func Unmarshal(response AcsResponse, httpResponse *http.Response, format string) (err error) {
	err = response.parseFromHttpResponse(httpResponse)
	if err != nil {
		return
	}
	if !response.IsSuccess() {
		err = errors.NewServerError(response.GetHttpStatus(), response.GetHttpContentString(), "")
		return
	}
	if _, isCommonResponse := response.(CommonResponse); isCommonResponse {
		// common response need not unmarshal
		return
	}

	if len(response.GetHttpContentBytes()) == 0 {
		return
	}

	if strings.ToUpper(format) == "JSON" {
		initJsonParserOnce()
		err = jsonParser.Unmarshal(response.GetHttpContentBytes(), response)
		if err != nil {
			err = errors.NewClientError(errors.JsonUnmarshalErrorCode, errors.JsonUnmarshalErrorMessage, err)
		}
	} else if strings.ToUpper(format) == "XML" {
		err = xml.Unmarshal(response.GetHttpContentBytes(), response)
	}
	return
}

// BaseResponse wrap originHttpResponse
type BaseResponse struct {
	httpStatus         int
	httpHeaders        map[string][]string
	httpContentString  string
	httpContentBytes   []byte
	originHttpResponse *http.Response
}

// GetHttpStatus returns httpStatus
func (baseResponse *BaseResponse) GetHttpStatus() int {
	return baseResponse.httpStatus
}

// GetHttpHeaders returns httpHeaders
func (baseResponse *BaseResponse) GetHttpHeaders() map[string][]string {
	return baseResponse.httpHeaders
}

// GetHttpContentString returns httpContentString
func (baseResponse *BaseResponse) GetHttpContentString() string {
	return baseResponse.httpContentString
}

// GetHttpContentBytes returns httpContentBytes
func (baseResponse *BaseResponse) GetHttpContentBytes() []byte {
	return baseResponse.httpContentBytes
}

// GetOriginHttpResponse returns originHttpResponse
func (baseResponse *BaseResponse) GetOriginHttpResponse() *http.Response {
	return baseResponse.originHttpResponse
}

// IsSuccess checks weather httpStatus is 200 or not
func (baseResponse *BaseResponse) IsSuccess() bool {
	if baseResponse.GetHttpStatus() >= 200 && baseResponse.GetHttpStatus() < 300 {
		return true
	}

	return false
}

func (baseResponse *BaseResponse) parseFromHttpResponse(httpResponse *http.Response) (err error) {
	defer httpResponse.Body.Close()
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return
	}
	baseResponse.httpStatus = httpResponse.StatusCode
	baseResponse.httpHeaders = httpResponse.Header
	baseResponse.httpContentBytes = body
	baseResponse.httpContentString = string(body)
	baseResponse.originHttpResponse = httpResponse
	return
}

// String returns base response content
func (baseResponse *BaseResponse) String() string {
	resultBuilder := bytes.Buffer{}
	// statusCode
	resultBuilder.WriteString("\n")
	resultBuilder.WriteString(fmt.Sprintf("%s %s\n", baseResponse.originHttpResponse.Proto, baseResponse.originHttpResponse.Status))
	// httpHeaders
	//resultBuilder.WriteString("Headers:\n")
	for key, value := range baseResponse.httpHeaders {
		resultBuilder.WriteString(key + ": " + strings.Join(value, ";") + "\n")
	}
	resultBuilder.WriteString("\n")
	// content
	//resultBuilder.WriteString("Content:\n")
	resultBuilder.WriteString(baseResponse.httpContentString + "\n")
	return resultBuilder.String()
}

// CommonResponse wrap base response
type CommonResponse struct {
	*BaseResponse
}

// NewCommonResponse return common response
func NewCommonResponse() (response *CommonResponse) {
	return &CommonResponse{
		BaseResponse: &BaseResponse{},
	}
}
