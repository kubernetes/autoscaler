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

package volcengine

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/service/ecs"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/session"
)

// EcsService represents the ECS interfaces
type EcsService interface {
	GetInstanceTypeById(instanceTypeId string) (*ecs.InstanceTypeForDescribeInstanceTypesOutput, error)
}

type ecsService struct {
	ecsClient *ecs.ECS
}

// GetInstanceTypeById returns instance type info by given instance type id
func (e *ecsService) GetInstanceTypeById(instanceTypeId string) (*ecs.InstanceTypeForDescribeInstanceTypesOutput, error) {
	resp, err := e.ecsClient.DescribeInstanceTypes(&ecs.DescribeInstanceTypesInput{
		InstanceTypeIds: volcengine.StringSlice([]string{instanceTypeId}),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.InstanceTypes) == 0 || volcengine.StringValue(resp.InstanceTypes[0].InstanceTypeId) != instanceTypeId {
		return nil, fmt.Errorf("instance type %s not found", instanceTypeId)
	}
	return resp.InstanceTypes[0], nil
}

func newEcsService(cloudConfig *cloudConfig) EcsService {
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(cloudConfig.getAccessKey(), cloudConfig.getSecretKey(), "")).
		WithRegion(cloudConfig.getRegion()).
		WithEndpoint(cloudConfig.getEndpoint())
	sess, _ := session.NewSession(config)
	client := ecs.New(sess)
	return &ecsService{
		ecsClient: client,
	}
}
