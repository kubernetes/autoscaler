/*
Copyright 2016 The Kubernetes Authors.

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

package alicloud

import (
	"github.com/denverdino/aliyungo/metadata"
	"github.com/golang/glog"
	"os"
)

const (
	ACCESS_KEY_ID     = "ACCESS_KEY_ID"
	ACCESS_KEY_SECRET = "ACCESS_KEY_SECRET"
	REGION_ID         = "REGION_ID"
)

type CloudConfig struct {
	RegionId        string
	AccessKeyID     string
	AccessKeySecret string
	STSEnabled      bool
}

func (cc *CloudConfig) IsValid() bool {
	if cc.AccessKeyID == "" {
		cc.AccessKeyID = os.Getenv(ACCESS_KEY_ID)
	}

	if cc.AccessKeySecret == "" {
		cc.AccessKeySecret = os.Getenv(ACCESS_KEY_SECRET)
	}

	if cc.RegionId == "" {
		cc.RegionId = os.Getenv(REGION_ID)
	}

	if cc.RegionId == "" || cc.AccessKeyID == "" || cc.AccessKeySecret == "" {
		glog.V(5).Infof("Failed to get AccessKeyId:%s,AccessKeySecret:%s,RegionId:%s from CloudConfig and Env\n", cc.AccessKeyID, cc.AccessKeySecret, cc.RegionId)
		glog.V(5).Infof("Try to use sts token in metadata instead.\n")
		if cc.ValidateSTSToken() == true && cc.GetRegion() != "" {
			//if CA is working on ECS with valid role name, use sts token instead.
			cc.STSEnabled = true
			return true
		}
	} else {
		cc.STSEnabled = false
		return true
	}
	return false
}

func (cc *CloudConfig) ValidateSTSToken() bool {
	m := metadata.NewMetaData(nil)
	r, err := m.RoleName()
	if err != nil || r == "" {
		glog.Warningf("The role name %s is not valid and error is %v", r, err)
		return false
	}
	return true
}

func (cc *CloudConfig) GetSTSToken() (metadata.RoleAuth, error) {
	m := metadata.NewMetaData(nil)
	r, err := m.RoleName()
	if err != nil {
		return metadata.RoleAuth{}, err
	}
	auth, err := m.RamRoleToken(r)
	if err != nil {
		return metadata.RoleAuth{}, err
	}
	return auth, nil
}

func (cc *CloudConfig) GetRegion() string {
	if cc.RegionId != "" {
		return cc.RegionId
	}
	m := metadata.NewMetaData(nil)
	r, err := m.Region()
	if err != nil {
		glog.Errorf("Failed to get RegionId from metadata.Because of %s\n", err.Error())
	}
	return r
}
