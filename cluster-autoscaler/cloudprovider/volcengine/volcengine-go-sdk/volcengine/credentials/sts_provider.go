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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volc-sdk-golang/service/sts"
)

type StsValue StsAssumeRoleProvider

type StsProvider struct {
	Expiry
	StsValue
}

func (s *StsProvider) Retrieve() (Value, error) {
	ins := sts.NewInstance()
	if s.Region != "" {
		ins.SetRegion(s.Region)
	}
	if s.Host != "" {
		ins.SetHost(s.Host)
	}
	if s.Timeout > 0 {
		ins.Client.SetTimeout(s.Timeout)
	}
	if s.DurationSeconds < 900 {
		return Value{}, fmt.Errorf("DurationSeconds must greater than 900 seconds ")
	}

	ins.Client.SetAccessKey(s.AccessKey)
	ins.Client.SetSecretKey(s.SecurityKey)
	input := &sts.AssumeRoleRequest{
		DurationSeconds: s.DurationSeconds,
		RoleTrn:         fmt.Sprintf("trn:iam::%s:role/%s", s.AccountId, s.RoleName),
		RoleSessionName: uuid.New().String(),
	}
	t := time.Now().Add(time.Duration(s.DurationSeconds-60) * time.Second)
	output, _, err := ins.AssumeRole(input)
	if err != nil || output.ResponseMetadata.Error != nil {
		if err == nil {
			bb, _err := json.Marshal(output.ResponseMetadata.Error)
			if _err != nil {
				return Value{}, _err
			}
			return Value{}, fmt.Errorf(string(bb))
		}
		return Value{}, err
	}
	v := Value{
		AccessKeyID:     output.Result.Credentials.AccessKeyId,
		SecretAccessKey: output.Result.Credentials.SecretAccessKey,
		SessionToken:    output.Result.Credentials.SessionToken,
		ProviderName:    "StsProvider",
	}
	s.SetExpiration(t, 0)
	return v, nil
}

func (s *StsProvider) IsExpired() bool {
	return s.Expiry.IsExpired()
}

func NewStsCredentials(value StsValue) *Credentials {

	p := &StsProvider{
		StsValue: value,
		Expiry:   Expiry{},
	}
	return NewExpireAbleCredentials(p)
}
