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

package sts

import (
	"encoding/json"
	"net/url"
	"strconv"
)

func (p *STS) commonHandler(api string, query url.Values, resp interface{}) (int, error) {
	respBody, statusCode, err := p.Client.Query(api, query)
	if err != nil {
		return statusCode, err
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return statusCode, err
	}
	return statusCode, nil
}

func (p *STS) AssumeRole(req *AssumeRoleRequest) (*AssumeRoleResp, int, error) {
	query := url.Values{}
	resp := new(AssumeRoleResp)

	if req.DurationSeconds > 0 {
		query.Set("DurationSeconds", strconv.Itoa(req.DurationSeconds))
	}

	if req.Policy != "" {
		query.Set("Policy", req.Policy)
	}

	query.Set("RoleTrn", req.RoleTrn)
	query.Set("RoleSessionName", req.RoleSessionName)

	statusCode, err := p.commonHandler("AssumeRole", query, resp)
	if err != nil {
		return nil, statusCode, err
	}
	return resp, statusCode, nil
}
