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

package eip

import (
	"bytes"
	"encoding/json"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
)

// Eip  json
type Eip struct {
	Name            string          `json:"name"`
	Eip             string          `json:"eip"`
	Status          string          `json:"status"`
	EipInstanceType EipInstanceType `json:"eipInstanceType"`
	InstanceType    InstanceType    `json:"instanceType"`
	InstanceId      string          `json:"instanceId"`
	ShareGroupId    string          `json:"shareGroupId"`
	BandwidthInMbps int             `json:"bandwidthInMbps"`
	PaymentTiming   string          `json:"paymentTiming"`
	BillingMethod   string          `json:"billingMethod"`
	CreateTime      string          `json:"createTime"`
	ExpireTime      string          `json:"expireTime"`
}

// Billing  json
type Billing struct {
	PaymentTiming string `json:"paymentTiming"`
	BillingMethod string `json:"billingMethod"`
}

// Reservation  json
type Reservation struct {
	ReservationLength   int    `json:"reservationLength"`
	ReservationTimeUnit string `json:"reservationTimeUnit"`
}

// CreateEipArgs  json
type CreateEipArgs struct {
	//  公网带宽，单位为Mbps。
	// 对于prepay以及bandwidth类型的EIP，限制为为1~200之间的整数，
	// 对于traffic类型的EIP，限制为1~1000之前的整数。
	BandwidthInMbps int      `json:"bandwidthInMbps"`
	Billing         *Billing `json:"billing"`
	Name            string   `json:"name,omitempty"`
}

// CreateEipResponse  json
type CreateEipResponse struct {
	Ip string `json:"eip"`
}

// InstanceType  json
type InstanceType string

const (
	// BCC instance
	BCC InstanceType = "BCC"
	// BLB instance
	BLB InstanceType = "BLB"
)

const (
	// PAYMENTTIMING_PREPAID eip
	PAYMENTTIMING_PREPAID string = "Prepaid"
	// PAYMENTTIMING_POSTPAID eip
	PAYMENTTIMING_POSTPAID string = "Postpaid"
	// BILLINGMETHOD_BYTRAFFIC eip
	BILLINGMETHOD_BYTRAFFIC string = "ByTraffic"
	// BILLINGMETHOD_BYBANDWIDTH eip
	BILLINGMETHOD_BYBANDWIDTH string = "ByBandwidth"
)

// EipInstanceType type
type EipInstanceType string

const (
	// NORMAL type
	NORMAL EipInstanceType = "normal"
	// SHARED type
	SHARED EipInstanceType = "shared"
)

func (args *CreateEipArgs) validate() error {
	if args == nil {
		return fmt.Errorf("CreateEipArgs need args")
	}
	if args.BandwidthInMbps == 0 {
		return fmt.Errorf("CreateEipArgs need BandwidthInMbps")
	}
	if args.Billing == nil {
		return fmt.Errorf("CreateEipArgs need Billing")
	}
	return nil
}

// CreateEip create a eip
func (c *Client) CreateEip(args *CreateEipArgs) (string, error) {
	err := args.validate()
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	req, err := bce.NewRequest("POST", c.GetURL("v1/eip", params), bytes.NewBuffer(postContent))
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
	var blbsResp *CreateEipResponse
	err = json.Unmarshal(bodyContent, &blbsResp)

	if err != nil {
		return "", err
	}
	return blbsResp.Ip, nil

}

// ResizeEipArgs json
type ResizeEipArgs struct {
	BandwidthInMbps int    `json:"newBandwidthInMbps"`
	Ip              string `json:"-"`
}

func (args *ResizeEipArgs) validate() error {
	if args == nil {
		return fmt.Errorf("ResizeEipArgs need args")
	}
	if args.Ip == "" {
		return fmt.Errorf("ResizeEipArgs need ip")
	}
	if args.BandwidthInMbps == 0 {
		return fmt.Errorf("ResizeEipArgs need BandwidthInMbps")
	}
	return nil
}

// ResizeEip resize a eip
func (c *Client) ResizeEip(args *ResizeEipArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"resize":      "",
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return err
	}
	// url := "http://" + Endpoint[c.GetRegion()] + "/v1/eip" + "/" + args.Ip + "?" + "resize&" + "clientToken=" + c.GenerateClientToken()
	// req, err := bce.NewRequest("PUT", url, bytes.NewBuffer(postContent))
	req, err := bce.NewRequest("PUT", c.GetURL("v1/eip"+"/"+args.Ip, params), bytes.NewBuffer(postContent))
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil

}

// BindEipArgs  json
type BindEipArgs struct {
	Ip           string       `json:"-"`
	InstanceType InstanceType `json:"instanceType"`
	InstanceId   string       `json:"instanceId"`
}

func (args *BindEipArgs) validate() error {
	if args == nil {
		return fmt.Errorf("BindEip need args")
	}
	if args.Ip == "" {
		return fmt.Errorf("BindEip need ip")
	}
	if args.InstanceType == "" {
		return fmt.Errorf("BindEip need InstanceType")
	}
	if args.InstanceId == "" {
		return fmt.Errorf("BindEip need InstanceId")
	}
	return nil
}

// BindEip bind a eip
func (c *Client) BindEip(args *BindEipArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"bind":        "",
		"clientToken": c.GenerateClientToken(),
	}
	postContent, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req, err := bce.NewRequest("PUT", c.GetURL("v1/eip"+"/"+args.Ip, params), bytes.NewBuffer(postContent))
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil

}

// EipArgs json
type EipArgs struct {
	Ip string `json:"-"`
}

func (args *EipArgs) validate() error {
	if args == nil {
		return fmt.Errorf("EipArgs need args")
	}
	if args.Ip == "" {
		return fmt.Errorf("EipArgs need ip")
	}
	return nil
}

// UnbindEip unbind a eip
func (c *Client) UnbindEip(args *EipArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"unbind":      "",
		"clientToken": c.GenerateClientToken(),
	}

	req, err := bce.NewRequest("PUT", c.GetURL("v1/eip"+"/"+args.Ip, params), nil)
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteEip delete a eip
func (c *Client) DeleteEip(args *EipArgs) error {
	err := args.validate()
	if err != nil {
		return err
	}
	params := map[string]string{
		"clientToken": c.GenerateClientToken(),
	}
	req, err := bce.NewRequest("DELETE", c.GetURL("v1/eip"+"/"+args.Ip, params), nil)
	if err != nil {
		return err
	}
	_, err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetEipsArgs json
type GetEipsArgs struct {
	Ip           string       `json:"-"`
	InstanceType InstanceType `json:"instanceType"`
	InstanceId   string       `json:"instanceId"`
}

// GetEipsResponse json
type GetEipsResponse struct {
	EipList     []Eip  `json:"eipList"`
	Marker      string `json:"marker"`
	IsTruncated bool   `json:"isTruncated"`
	NextMarker  string `json:"nextMarker"`
	MaxKeys     int    `json:"maxKeys"`
}

// GetEips get eips
func (c *Client) GetEips(args *GetEipsArgs) ([]Eip, error) {
	if args == nil {
		args = &GetEipsArgs{}
	}
	params := map[string]string{
		"eip":          args.Ip,
		"instanceType": string(args.InstanceType),
		"instanceId":   args.InstanceId,
	}
	req, err := bce.NewRequest("GET", c.GetURL("v1/eip", params), nil)
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
	var blbsResp *GetEipsResponse
	err = json.Unmarshal(bodyContent, &blbsResp)

	if err != nil {
		return nil, err
	}
	return blbsResp.EipList, nil
}
