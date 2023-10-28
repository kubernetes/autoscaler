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

package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/profile"
)

// Client defined a tencentcloud client
type Client interface {
	Send(ctx context.Context, req tchttp.Request, res tchttp.Response) error
}

// NewClient new a Client
func NewClient(credential common.CredentialIface, region string, cpf *profile.ClientProfile) Client {
	c := &client{}
	c.Init(region).
		WithCredential(credential).
		WithProfile(cpf)
	return c
}

type client struct {
	common.Client
}

// Send a http request
func (c *client) Send(ctx context.Context, req tchttp.Request, resp tchttp.Response) error {
	if req == nil || resp == nil {
		return errors.New("req & resp are not allowed to be empty")
	}
	if c.GetCredential() == nil {
		return errors.New("Send require credential")
	}

	start := time.Now()
	err := c.Client.Send(req, resp)

	// 上报指标
	duration := time.Since(start)
	metrics.RegisterCloudAPIInvoked(req.GetService(), req.GetAction(), err)

	// 打印日志
	responseBytes, _ := json.Marshal(resp)
	requestBytes, _ := json.Marshal(req)
	klog.V(4).Infof("\"invoke cloud api\" Host=%s Service=%s Action=%s Duration=%s Request=%s Response=%s",
		req.GetDomain(), req.GetService(), req.GetAction(), duration.String(), string(requestBytes), string(responseBytes))

	return err
}
