/*
Copyright 2021 The Kubernetes Authors.

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

package errors

import (
	"fmt"
)

type TencentCloudSDKError struct {
	Code      string
	Message   string
	RequestId string
}

func (e *TencentCloudSDKError) Error() string {
	if e.RequestId == "" {
		return fmt.Sprintf("[TencentCloudSDKError] Code=%s, Message=%s", e.Code, e.Message)
	}
	return fmt.Sprintf("[TencentCloudSDKError] Code=%s, Message=%s, RequestId=%s", e.Code, e.Message, e.RequestId)
}

func NewTencentCloudSDKError(code, message, requestId string) error {
	return &TencentCloudSDKError{
		Code:      code,
		Message:   message,
		RequestId: requestId,
	}
}

func (e *TencentCloudSDKError) GetCode() string {
	return e.Code
}

func (e *TencentCloudSDKError) GetMessage() string {
	return e.Message
}

func (e *TencentCloudSDKError) GetRequestId() string {
	return e.RequestId
}
