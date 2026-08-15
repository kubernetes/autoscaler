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

package blb

import (
	"bytes"
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
)

// BackendServer json
type BackendServer struct {
	InstanceId string `json:"instanceId"`
	Weight     int    `json:"weight,omitempty"`
}

// BackendServerStatus json
type BackendServerStatus struct {
	InstanceId string `json:"instanceId"`
	Weight     int    `json:"weight"`
	Status     string `json:"status"`
}

// AddBackendServersArgs json
type AddBackendServersArgs struct {
	LoadBalancerId    string          `json:"-"`
	BackendServerList []BackendServer `json:"backendServerList"`
}

func (args *AddBackendServersArgs) validate() error {
	if args == nil {
		return fmt.Errorf("AddBackendServersArgs need args")
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("AddBackendServersArgs need LoadBalancerId")
	}
	if args.BackendServerList == nil {
		return fmt.Errorf("UpdateUDPListener need BackendServerList")
	}

	return nil
}

// AddBackendServers add backend server
func (c *Client) AddBackendServers(args *AddBackendServersArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req, err := bce.NewRequest("POST", c.GetURL("v1/blb"+"/"+args.LoadBalancerId+"/backendserver", params), bytes.NewBuffer(postContent))
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// DescribeBackendServersArgs json
type DescribeBackendServersArgs struct {
	LoadBalancerId string `json:"-"`
}

// DescribeBackendServersResponse json
type DescribeBackendServersResponse struct {
	Marker            string          `json:"marker"`
	IsTruncated       bool            `json:"isTruncated"`
	NextMarker        string          `json:"nextMarker"`
	MaxKeys           int             `json:"maxKeys"`
	BackendServerList []BackendServer `json:"backendServerList"`
}

func (args *DescribeBackendServersArgs) validate() error {
	if args == nil {
		return fmt.Errorf("DescribeBackendServersArgs need args")
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("DescribeBackendServersArgs need LoadBalancerId")
	}
	return nil
}

// DescribeBackendServers describe backend server
func (c *Client) DescribeBackendServers(args *DescribeBackendServersArgs) ([]BackendServer, error) {
	err := args.validate()
	if err != nil {
		return nil, err
	}
	req, err := bce.NewRequest("GET", c.GetURL("v1/blb"+"/"+args.LoadBalancerId+"/backendserver", nil), nil)
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
	var blbsResp *DescribeBackendServersResponse
	err = json.Unmarshal(bodyContent, &blbsResp)

	if err != nil {
		return nil, err
	}
	return blbsResp.BackendServerList, nil

}

// UpdateBackendServersArgs json
type UpdateBackendServersArgs struct {
	LoadBalancerId    string          `json:"-"`
	BackendServerList []BackendServer `json:"backendServerList"`
}

func (args *UpdateBackendServersArgs) validate() error {
	if args == nil {
		return fmt.Errorf("UpdateBackendServersArgs need args")
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("UpdateBackendServersArgs need LoadBalancerId")
	}
	if len(args.BackendServerList) == 0 {
		return fmt.Errorf("UpdateBackendServersArgs need BackendServerList")
	}
	return nil
}

// UpdateBackendServers update  BackendServers
func (c *Client) UpdateBackendServers(args *UpdateBackendServersArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"update":      "",
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req, err := bce.NewRequest("PUT", c.GetURL("v1/blb"+"/"+args.LoadBalancerId+"/backendserver", params), bytes.NewBuffer(postContent))
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// RemoveBackendServersArgs json
type RemoveBackendServersArgs struct {
	LoadBalancerId    string   `json:"-"`
	BackendServerList []string `json:"backendServerList"`
}

func (args *RemoveBackendServersArgs) validate() error {
	if args == nil {
		return fmt.Errorf("UpdateBackendServersArgs need args")
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("UpdateBackendServersArgs need LoadBalancerId")
	}
	if len(args.BackendServerList) == 0 {
		return fmt.Errorf("UpdateBackendServersArgs need BackendServerList")
	}
	return nil
}

// RemoveBackendServers remove a BackendServers
func (c *Client) RemoveBackendServers(args *RemoveBackendServersArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req, err := bce.NewRequest("PUT", c.GetURL("v1/blb"+"/"+args.LoadBalancerId+"/backendserver", params), bytes.NewBuffer(postContent))
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}
