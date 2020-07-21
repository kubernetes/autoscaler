/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/auth/aksk"
)

// CloudConfig is the cloud config file for huaweicloud.
type CloudConfig struct {
	Global struct {
		IdentityEndpoint string `gcfg:"identity-endpoint"` // example: "https://iam.cn-north-4.myhuaweicloud.com/v3.0"
		ProjectID        string `gcfg:"project-id"`
		AccessKey        string `gcfg:"access-key"`
		SecretKey        string `gcfg:"secret-key"`
		Cloud            string `gcfg:"cloud"`     // example: "huaweicloud"
		Region           string `gcfg:"region"`    // example: "cn-north-4"
		DomainID         string `gcfg:"domain-id"` // The ACCOUNT ID. example: "a0e8ff63c0fb4fd49cc2dbdf1dea14e2"
	}
}

// toAKSKOptions creates and returns a new instance of type aksk.AKSKOptions
func toAKSKOptions(cfg CloudConfig) aksk.AKSKOptions {
	opts := aksk.AKSKOptions{
		IdentityEndpoint: cfg.Global.IdentityEndpoint,
		ProjectID:        cfg.Global.ProjectID,
		AccessKey:        cfg.Global.AccessKey,
		SecretKey:        cfg.Global.SecretKey,
		Cloud:            cfg.Global.Cloud,
		Region:           cfg.Global.Region,
		DomainID:         cfg.Global.DomainID,
	}
	return opts
}
