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

package bcc

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
)

// Endpoint contains all endpoints of Baidu Cloud BCC.
var Endpoint = map[string]string{
	"bj": "bcc.bj.baidubce.com",
	"gz": "bcc.gz.baidubce.com",
	"su": "bcc.su.baidubce.com",
	"bd": "bcc.bd.baidubce.com",
	"hk": "bcc.hkg.baidubce.com",
}

// Client is the BCC client implemention for Baidu Cloud BCC API.
type Client struct {
	*bce.Client
}

// NewClient returns client for BCC
func NewClient(config *bce.Config) *Client {
	bceClient := bce.NewClient(config)
	return &Client{bceClient}
}

// GetURL generates the full URL of http request for Baidu Cloud BCC API.
func (c *Client) GetURL(objectKey string, params map[string]string) string {
	host := c.Endpoint
	if host == "" {
		host = Endpoint[c.GetRegion()]
	}
	uriPath := objectKey
	return c.Client.GetURL(host, uriPath, params)
}
