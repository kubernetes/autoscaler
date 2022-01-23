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

package tencentcloud

import (
	as "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	cvm "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"k8s.io/klog/v2"
)

// MockCloudService mock cloud service
type MockCloudService struct {
	CloudService
}

// NewCloudMockService creates an instance of caching MockCloudService
func NewCloudMockService(cvmClient *cvm.Client, vpcClient *vpc.Client, asClient *as.Client, tkeClient *tke.Client) CloudService {
	return &MockCloudService{
		CloudService: NewCloudService(cvmClient, vpcClient, asClient, tkeClient),
	}
}

// DeleteInstances remove instances of specified ASG.
func (mcs *MockCloudService) DeleteInstances(asg Asg, instances []string) error {
	klog.Infof("Delete Instances %v from %s", instances, asg.TencentcloudRef().String())
	return nil
}

// ResizeAsg set the target size of ASG.
func (mcs *MockCloudService) ResizeAsg(ref TcRef, size uint64) error {
	klog.Infof("Resize %s to %d", ref.String(), size)
	return nil
}
