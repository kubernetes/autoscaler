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

package alicloud

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/metadata"
	klog "k8s.io/klog/v2"
	"os"
)

const (
	accessKeyId     = "ACCESS_KEY_ID"
	accessKeySecret = "ACCESS_KEY_SECRET"
	regionId        = "REGION_ID"
)

type cloudConfig struct {
	RegionId        string
	AccessKeyID     string
	AccessKeySecret string
	STSEnabled      bool
}

func (cc *cloudConfig) isValid() bool {
	if cc.AccessKeyID == "" {
		cc.AccessKeyID = os.Getenv(accessKeyId)
	}

	if cc.AccessKeySecret == "" {
		cc.AccessKeySecret = os.Getenv(accessKeySecret)
	}

	if cc.RegionId == "" {
		cc.RegionId = os.Getenv(regionId)
	}

	if cc.RegionId == "" || cc.AccessKeyID == "" || cc.AccessKeySecret == "" {
		klog.V(5).Infof("Failed to get AccessKeyId:%s,AccessKeySecret:%s,RegionId:%s from cloudConfig and Env\n", cc.AccessKeyID, cc.AccessKeySecret, cc.RegionId)
		klog.V(5).Infof("Try to use sts token in metadata instead.\n")
		if cc.validateSTSToken() == true && cc.getRegion() != "" {
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

func (cc *cloudConfig) validateSTSToken() bool {
	m := metadata.NewMetaData(nil)
	r, err := m.RoleName()
	if err != nil || r == "" {
		klog.Warningf("The role name %s is not valid and error is %v", r, err)
		return false
	}
	return true
}

func (cc *cloudConfig) getSTSToken() (metadata.RoleAuth, error) {
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

func (cc *cloudConfig) getRegion() string {
	if cc.RegionId != "" {
		return cc.RegionId
	}
	m := metadata.NewMetaData(nil)
	r, err := m.Region()
	if err != nil {
		klog.Errorf("Failed to get RegionId from metadata.Because of %s\n", err.Error())
	}
	return r
}
