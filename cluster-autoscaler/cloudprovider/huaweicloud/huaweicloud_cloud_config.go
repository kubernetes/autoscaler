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
	"fmt"
	"io"
	"os"

	"gopkg.in/gcfg.v1"
	huaweicloudsdkbasic "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	huaweicloudsdkconfig "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	huaweicloudsdkas "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	huaweicloudsdkecs "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
)

// CloudConfig is the cloud config file for huaweicloud.
type CloudConfig struct {
	Global struct {
		ECSEndpoint string `gcfg:"ecs-endpoint"`
		ASEndpoint  string `gcfg:"as-endpoint"`
		ProjectID   string `gcfg:"project-id"`
		AccessKey   string `gcfg:"access-key"`
		SecretKey   string `gcfg:"secret-key"`
	}
}

func (c *CloudConfig) getECSClient() *huaweicloudsdkecs.EcsClient {
	// There are two types of services provided by HUAWEI CLOUD according to scope:
	// - Regional services: most of services belong to this classification, such as ECS.
	// - Global services: such as IAM, TMS, EPS.
	// For Regional services' authentication, projectId is required.
	// For global services' authentication, domainId is required.
	// More details please refer to:
	// https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/0281b9734f0f95ed5565729e54d96e9820262426/README.md#use-go-sdk
	credentials := huaweicloudsdkbasic.NewCredentialsBuilder().
		WithAk(c.Global.AccessKey).
		WithSk(c.Global.SecretKey).
		WithProjectId(c.Global.ProjectID).
		Build()

	client := huaweicloudsdkecs.EcsClientBuilder().
		WithEndpoint(c.Global.ECSEndpoint).
		WithCredential(credentials).
		WithHttpConfig(huaweicloudsdkconfig.DefaultHttpConfig()).
		Build()

	return huaweicloudsdkecs.NewEcsClient(client)
}

func (c *CloudConfig) getASClient() *huaweicloudsdkas.AsClient {
	credentials := huaweicloudsdkbasic.NewCredentialsBuilder().
		WithAk(c.Global.AccessKey).
		WithSk(c.Global.SecretKey).
		WithProjectId(c.Global.ProjectID).
		Build()

	client := huaweicloudsdkas.AsClientBuilder().
		WithEndpoint(c.Global.ASEndpoint).
		WithCredential(credentials).
		WithHttpConfig(huaweicloudsdkconfig.DefaultHttpConfig()).
		Build()

	return huaweicloudsdkas.NewAsClient(client)
}

func (c *CloudConfig) validate() error {
	if len(c.Global.ECSEndpoint) == 0 {
		return fmt.Errorf("ECS endpoint missing from cloud configuration")
	}

	if len(c.Global.ASEndpoint) == 0 {
		return fmt.Errorf("AS endpoint missing from cloud configuration")
	}

	if len(c.Global.AccessKey) == 0 {
		return fmt.Errorf("access key missing from cloud coniguration")
	}

	if len(c.Global.SecretKey) == 0 {
		return fmt.Errorf("secret key missing from cloud configuration")
	}

	if len(c.Global.ProjectID) == 0 {
		return fmt.Errorf("project id missing from cloud configuration")
	}

	return nil
}

func readConf(confFile string) (*CloudConfig, error) {
	var conf io.ReadCloser
	conf, err := os.Open(confFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %s, error: %v", confFile, err)
	}
	defer conf.Close()

	var cloudConfig CloudConfig
	if err := gcfg.ReadInto(&cloudConfig, conf); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %s, error: %v", confFile, err)
	}

	return &cloudConfig, nil
}
