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
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/metadata"
	"k8s.io/klog/v2"
)

const (
	accessKeyId          = "ACCESS_KEY_ID"
	accessKeySecret      = "ACCESS_KEY_SECRET"
	oidcProviderARN      = "ALIBABA_CLOUD_OIDC_PROVIDER_ARN"
	oldOidcProviderARN   = "ALICLOUD_OIDC_PROVIDER_ARN"
	oidcTokenFilePath    = "ALIBABA_CLOUD_OIDC_TOKEN_FILE"
	oldOidcTokenFilePath = "ALICLOUD_OIDC_TOKEN_FILE_PATH"
	roleARN              = "ALIBABA_CLOUD_ROLE_ARN"
	oldRoleARN           = "ALICLOUD_ROLE_ARN"
	roleSessionName      = "ALIBABA_CLOUD_SESSION_NAME"
	oldRoleSessionName   = "ALICLOUD_SESSION_NAME"
	regionId             = "REGION_ID"
)

type cloudConfig struct {
	RegionId          string
	AccessKeyID       string
	AccessKeySecret   string
	OIDCProviderARN   string
	OIDCTokenFilePath string
	RoleARN           string
	RoleSessionName   string
	RRSAEnabled       bool
	STSEnabled        bool
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

	if cc.OIDCProviderARN == "" {
		cc.OIDCProviderARN = firstNotEmpty(os.Getenv(oidcProviderARN), os.Getenv(oldOidcProviderARN))
	}

	if cc.OIDCTokenFilePath == "" {
		cc.OIDCTokenFilePath = firstNotEmpty(os.Getenv(oidcTokenFilePath), os.Getenv(oldOidcTokenFilePath))
	}

	if cc.RoleARN == "" {
		cc.RoleARN = firstNotEmpty(os.Getenv(roleARN), os.Getenv(oldRoleARN))
	}

	if cc.RoleSessionName == "" {
		cc.RoleSessionName = firstNotEmpty(os.Getenv(roleSessionName), os.Getenv(oldRoleSessionName))
	}

	if cc.RegionId != "" && cc.AccessKeyID != "" && cc.AccessKeySecret != "" {
		klog.V(2).Info("Using AccessKey authentication")
		return true
	} else if cc.RegionId != "" && cc.OIDCProviderARN != "" && cc.OIDCTokenFilePath != "" && cc.RoleARN != "" && cc.RoleSessionName != "" {
		klog.V(2).Info("Using RRSA authentication")
		cc.RRSAEnabled = true
		return true
	} else {
		klog.V(5).Infof("Failed to get AccessKeyId:%s,RegionId:%s from cloudConfig and Env\n", cc.AccessKeyID, cc.RegionId)
		klog.V(5).Infof("Failed to get OIDCProviderARN:%s,OIDCTokenFilePath:%s,RoleARN:%s,RoleSessionName:%s,RegionId:%s from cloudConfig and Env\n", cc.OIDCProviderARN, cc.OIDCTokenFilePath, cc.RoleARN, cc.RoleSessionName, cc.RegionId)
		klog.V(5).Infof("Try to use sts token in metadata instead.\n")
		if cc.validateSTSToken() && cc.getRegion() != "" {
			//if CA is working on ECS with valid role name, use sts token instead.
			cc.STSEnabled = true
			return true
		}
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

// firstNotEmpty returns the first non-empty string from the input list.
// If all strings are empty or no arguments are provided, it returns an empty string.
func firstNotEmpty(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}

	return ""
}
