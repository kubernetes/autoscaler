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

package etag

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type Etag struct {
	ETag azcore.ETag `json:"etag,omitempty"`
}

func AppendEtag(req *policy.Request) (*http.Response, error) {
	if req.Raw().Method == http.MethodPatch || req.Raw().Method == http.MethodPost || req.Raw().Method == http.MethodPut {
		body, err := io.ReadAll(req.Body())
		if err != nil {
			return nil, err
		}
		var etag Etag
		err = json.Unmarshal(body, &etag)
		if err != nil {
			return nil, err
		}
		if etag.ETag != "" {
			req.Raw().Header.Set("If-Match", string(etag.ETag))
		}
		if err = req.RewindBody(); err != nil {
			return nil, err
		}
	}

	return req.Next()
}
