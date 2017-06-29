// +build !linux

/*
Copyright 2017 The Kubernetes Authors.

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

// Dummy implementation. Real one should be built on linux.

package kubemark

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

type KubemarkCloudProvider struct{}

func BuildKubemarkCloudProvider(kubemarkController *kubemark.KubemarkController, specs []string) (*KubemarkCloudProvider, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (kubemark *KubemarkCloudProvider) Name() string { return "" }

func (kubemark *KubemarkCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return []cloudProvider.NodeGroup{}
}

func (kubemark *KubemarkCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (kubemark *KubemarkCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (kubemark *KubemarkCloudProvider) GetAvilableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

func (kubemark *KubemarkCloudProvider) NewNodeGroup(name string, machineType string, labels map[string]string, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

type FakeNodeGroup struct{}

func (f *FakeNodeGroup) Id() string                            { return "" }
func (f *FakeNodeGroup) MinSize() int                          { return 0 }
func (f *FakeNodeGroup) MaxSize() int                          { return 0 }
func (f *FakeNodeGroup) Debug() string                         { return "" }
func (f *FakeNodeGroup) Nodes() ([]string, error)              { return []string{}, cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) DeleteNodes(nodes []*apiv1.Node) error { return cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) IncreaseSize(delta int) error          { return cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) TargetSize() (int, error)              { return 0, cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) DecreaseTargetSize(delta int) error    { return cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
func (f *FakeNodeGroup) Exist() (bool, error) { return true, nil }
func (f *FakeNodeGroup) Create() error        { return cloudprovider.ErrNotImplemented }
func (f *FakeNodeGroup) Delete() error        { return cloudprovider.ErrNotImplemented }
