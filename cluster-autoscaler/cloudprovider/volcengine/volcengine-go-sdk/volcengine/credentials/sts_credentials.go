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

package credentials

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volc-sdk-golang/service/sts"
)

type StsAssumeRoleProvider struct {
	AccessKey       string
	SecurityKey     string
	RoleName        string
	AccountId       string
	Host            string
	Region          string
	Timeout         time.Duration
	DurationSeconds int
}

type StsAssumeRoleTime struct {
	CurrentTime string
	ExpiredTime string
}

func StsAssumeRole(p *StsAssumeRoleProvider) (*Credentials, *StsAssumeRoleTime, error) {
	ins := sts.NewInstance()
	if p.Region != "" {
		ins.SetRegion(p.Region)
	}
	if p.Host != "" {
		ins.SetHost(p.Host)
	}
	if p.Timeout > 0 {
		ins.Client.SetTimeout(p.Timeout)
	}

	ins.Client.SetAccessKey(p.AccessKey)
	ins.Client.SetSecretKey(p.SecurityKey)
	input := &sts.AssumeRoleRequest{
		DurationSeconds: p.DurationSeconds,
		RoleTrn:         fmt.Sprintf("trn:iam::%s:role/%s", p.AccountId, p.RoleName),
		RoleSessionName: uuid.New().String(),
	}
	output, statusCode, err := ins.AssumeRole(input)
	var reqId string
	if output != nil {
		reqId = output.ResponseMetadata.RequestId
	}
	if err != nil {
		return nil, nil, fmt.Errorf("AssumeRole error,httpcode is %v and reqId is %s error is %s", statusCode, reqId, err.Error())
	}
	if statusCode >= 300 || statusCode < 200 {
		return nil, nil, fmt.Errorf("AssumeRole error,httpcode is %v and reqId is %s", statusCode, reqId)
	}
	return NewCredentials(&StaticProvider{Value: Value{
			AccessKeyID:     output.Result.Credentials.AccessKeyId,
			SecretAccessKey: output.Result.Credentials.SecretAccessKey,
			SessionToken:    output.Result.Credentials.SessionToken,
		}}), &StsAssumeRoleTime{
			CurrentTime: output.Result.Credentials.CurrentTime,
			ExpiredTime: output.Result.Credentials.ExpiredTime,
		}, nil
}
