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

import "encoding/json"

type actionResult map[string]interface{}
type CommonResponse struct {
	*BaseResponse
	*actionResult
}

func NewCommonResponse() (response *CommonResponse) {
	response = &CommonResponse{
		BaseResponse: &BaseResponse{},
		actionResult: &actionResult{},
	}
	return
}

func (r *CommonResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, r.actionResult)
}

func (r *CommonResponse) GetBody() []byte {
	raw, _ := json.Marshal(r.actionResult)
	return raw
}
