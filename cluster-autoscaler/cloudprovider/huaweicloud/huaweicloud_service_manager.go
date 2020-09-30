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

// ScalingGroupInstance is a dummy struct for ScalingGroupInstance. And it will be replaced later.
type ScalingGroupInstance struct {
	InstanceId *string `json:"instance_id,omitempty"`
}

// ElasticCloudServerService represents the elastic cloud server interfaces.
// It should contains all request against elastic cloud server service.
type ElasticCloudServerService interface {
	// DeleteServers deletes a group of server by ID.
	DeleteServers(serverIDs []string) error
}

// AutoScalingService represents the auto scaling service interfaces.
// It should contains all request against auto scaling service.
type AutoScalingService interface {
	// GetDesireInstanceNumber gets the desire instance number of specific auto scaling group.
	GetDesireInstanceNumber(groupID string) (int, error)

	// GetInstances gets the instances in an auto scaling group.
	GetInstances(groupID string) ([]ScalingGroupInstance, error)

	// IncreaseSizeInstance increases the instance number of specific auto scaling group.
	// The delta should be non-negative.
	// IncreaseSizeInstance wait until instance number is updated.
	IncreaseSizeInstance(groupID string, delta int) error
}

// CloudServiceManager represents the cloud service interfaces.
// It should contains all requests against cloud services.
type CloudServiceManager interface {
	// ElasticCloudServerService represents the elastic cloud server interfaces.
	ElasticCloudServerService

	// AutoScalingService represents the auto scaling service interfaces.
	AutoScalingService
}
