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

package clientset

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bcc"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/blb"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/eip"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/vpc"
)

// Interface contains all methods of clients
type Interface interface {
	Bcc() *bcc.Client
	Blb() *blb.Client
	Eip() *eip.Client
	Vpc() *vpc.Client
}

// Clientset contains the clients for groups.
type Clientset struct {
	BccClient *bcc.Client
	BlbClient *blb.Client
	EipClient *eip.Client
	VpcClient *vpc.Client
}

// Bcc retrieves the BccClient
func (c *Clientset) Bcc() *bcc.Client {
	if c == nil {
		return nil
	}
	return c.BccClient
}

// Blb retrieves the BccClient
func (c *Clientset) Blb() *blb.Client {
	if c == nil {
		return nil
	}
	return c.BlbClient
}

// Eip retrieves the BccClient
func (c *Clientset) Eip() *eip.Client {
	if c == nil {
		return nil
	}
	return c.EipClient
}

// Vpc retrieves the VpcClient
func (c *Clientset) Vpc() *vpc.Client {
	if c == nil {
		return nil
	}
	return c.VpcClient
}

// NewFromConfig create a new Clientset for the given config.
func NewFromConfig(cfg *bce.Config) (*Clientset, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	var cs Clientset
	cs.BccClient = bcc.NewClient(cfg)
	cs.BlbClient = blb.NewBLBClient(cfg)
	cs.EipClient = eip.NewEIPClient(cfg)
	cs.VpcClient = vpc.NewVPCClient(cfg)
	return &cs, nil
}
