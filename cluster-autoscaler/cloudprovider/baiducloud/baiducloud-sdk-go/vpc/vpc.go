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

package vpc

import (
	"encoding/json"
	"strconv"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
)

// Vpc type define
type Vpc struct {
	VpcID       string `json:"vpcId"`
	Name        string `json:"name"`
	CIDR        string `json:"cidr"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
}

// ShowVpc define ShowVpcModel
type ShowVpc struct {
	VpcID       string   `json:"vpcId"`
	Name        string   `json:"name"`
	CIDR        string   `json:"cidr"`
	Description string   `json:"description"`
	IsDefault   bool     `json:"isDefault"`
	Subnets     []Subnet `json:"subnets"`
}

// CreateVpcArgs define args for creating vpc
type CreateVpcArgs struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Cidr        string `json:"cidr"`
}

// CreateVpcResponse define response
type CreateVpcResponse struct {
	VpcID string `json:"vpcId"`
}

// ListVpcArgs args
type ListVpcArgs struct {
	IsDefault bool `json:"isDefault"`
}

// ListVpcResponse define list vpc response
type ListVpcResponse struct {
	Vpcs []Vpc `json:"vpcs"`
}

// ListVpc list all vpcs
// https://cloud.baidu.com/doc/VPC/API.html#.E6.9F.A5.E8.AF.A2VPC.E5.88.97.E8.A1.A8
func (c *Client) ListVpc(args *ListVpcArgs) ([]Vpc, error) {
	if args == nil {
		args = &ListVpcArgs{}
	}
	params := map[string]string{
		"isDefault": strconv.FormatBool(args.IsDefault),
	}
	req, err := bce.NewRequest("GET", c.GetURL("v1/vpc", params), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.SendRequest(req, nil)
	if err != nil {
		return nil, err
	}
	bodyContent, err := resp.GetBodyContent()
	if err != nil {
		return nil, err
	}
	var lvResp *ListVpcResponse
	err = json.Unmarshal(bodyContent, &lvResp)
	if err != nil {
		return nil, err
	}
	return lvResp.Vpcs, nil
}
