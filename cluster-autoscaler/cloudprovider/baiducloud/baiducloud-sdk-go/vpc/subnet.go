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
	"bytes"
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
)

// Subnet define subnet of vpc
type Subnet struct {
	SubnetID    string `json:"subnetId"`
	Name        string `json:"name"`
	ZoneName    string `json:"zoneName"`
	Cidr        string `json:"cidr"`
	VpcID       string `json:"vpcId"`
	SubnetType  string `json:"subnetType"`
	Description string `json:"description"`
}

// CreateSubnetArgs define args create a subnet
type CreateSubnetArgs struct {
	Name        string `json:"name"`
	ZoneName    string `json:"zoneName"`
	Cidr        string `json:"cidr"`
	VpcID       string `json:"vpcId"`
	SubnetType  string `json:"subnetType,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateSubnetResponse define response of creating a subnet
type CreateSubnetResponse struct {
	SubnetID string `json:"subnetId"`
}

// ListSubnetResponse json
type ListSubnetResponse struct {
	Marker      string    `json:"marker"`
	IsTruncated bool      `json:"isTruncated"`
	NextMarker  string    `json:"nextMarker"`
	MaxKeys     int       `json:"maxKeys"`
	Subnets     []*Subnet `json:"subnets"`
}

// DescribeSubnetResponse json
type DescribeSubnetResponse struct {
	Subnet *Subnet `json:"subnet"`
}

// CreateSubnet 在VPC中创建子网
func (c *Client) CreateSubnet(args *CreateSubnetArgs) (string, error) {
	if args == nil {
		return "", fmt.Errorf("CreateSubnet failed: CreateSubnetArgs is nil")
	}
	params := map[string]string{
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	req, err := bce.NewRequest("POST", c.GetURL("v1/subnet", params), bytes.NewBuffer(postContent))
	if err != nil {
		return "", err
	}
	resp, err := c.SendRequest(req, nil)
	if err != nil {
		return "", err
	}
	bodyContent, err := resp.GetBodyContent()
	if err != nil {
		return "", err
	}

	var createSubnetResponse *CreateSubnetResponse
	err = json.Unmarshal(bodyContent, &createSubnetResponse)
	if err != nil {
		return "", err
	}
	return createSubnetResponse.SubnetID, nil
}

// ListSubnet 查询指定VPC的所有子网列表信息
func (c *Client) ListSubnet(params map[string]string) ([]*Subnet, error) {
	req, err := bce.NewRequest("GET", c.GetURL("v1/subnet", params), nil)
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

	var listSubnetResponse *ListSubnetResponse
	err = json.Unmarshal(bodyContent, &listSubnetResponse)
	if err != nil {
		return nil, err
	}
	return listSubnetResponse.Subnets, nil
}

// DescribeSubnet 查询指定子网的详细信息
func (c *Client) DescribeSubnet(subnetId string) (*Subnet, error) {
	if len(subnetId) == 0 {
		return nil, fmt.Errorf("DescribeSubnet failed, subnetId must not be empty")
	}
	req, err := bce.NewRequest("GET", c.GetURL("v1/subnet/"+subnetId, nil), nil)
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

	var describeSubnetResponse *DescribeSubnetResponse
	err = json.Unmarshal(bodyContent, &describeSubnetResponse)
	if err != nil {
		return nil, err
	}
	return describeSubnetResponse.Subnet, nil
}
