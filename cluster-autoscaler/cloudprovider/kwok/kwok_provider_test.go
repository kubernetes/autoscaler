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

package kwok

import (
	"os"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	v1lister "k8s.io/client-go/listers/core/v1"
	core "k8s.io/client-go/testing"
)

func TestNodeGroups(t *testing.T) {
	fakeClient := &fake.Clientset{}
	var nodesFrom string
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						configKey: testConfig,
					},
				}, nil
			}

			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplates,
				},
			}, nil

		}

		if getAction.GetName() == defaultTemplatesConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						templatesKey: testTemplates,
					},
				}, nil
			}
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	t.Run("use template nodes from the configmap", func(t *testing.T) {
		nodesFrom = "configmap"
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)

		ngs := p.NodeGroups()
		assert.NotNil(t, ngs)
		assert.NotEmpty(t, ngs)
		assert.Len(t, ngs, 1)
		assert.Contains(t, ngs[0].Id(), "kind-worker")
	})

	t.Run("use template nodes from the cluster (aka get them using kube client)", func(t *testing.T) {
		nodesFrom = "cluster"
		fakeNode := &apiv1.Node{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Node",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
				Labels: map[string]string{
					"kwok-nodegroup": "kind-worker",
				},
			},
			Spec: apiv1.NodeSpec{},
		}
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{
			"kind-worker": fakeNode,
		})

		fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
			getAction := action.(core.ListAction)

			if getAction == nil {
				return false, nil, nil
			}

			return true, &apiv1.NodeList{Items: []apiv1.Node{*fakeNode}}, nil
		})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)

		ngs := p.NodeGroups()
		assert.NotNil(t, ngs)
		assert.NotEmpty(t, ngs)
		assert.Len(t, ngs, 1)
		assert.Contains(t, ngs[0].Id(), "kind-worker")
	})
}

func TestRefresh(t *testing.T) {
	fakeClient := &fake.Clientset{}
	var nodesFrom string
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						configKey: testConfig,
					},
				}, nil
			}

			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplates,
				},
			}, nil

		}

		if getAction.GetName() == defaultTemplatesConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						templatesKey: testTemplates,
					},
				}, nil
			}
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	t.Run("refresh nodegroup target size", func(t *testing.T) {
		nodesFrom = "configmap"
		ngName := "kind-worker"
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{
			"node1": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Labels: map[string]string{
						"kwok-nodegroup": ngName,
					},
				},
			},
			"node2": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "node2",
					Labels: map[string]string{
						"kwok-nodegroup": ngName,
					},
				},
			},
			"node3": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "node3",
					Labels: map[string]string{
						"kwok-nodegroup": ngName,
					},
				},
			},
			"node4": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "node4",
				},
			},
		})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)

		err = p.Refresh()
		assert.Nil(t, err)
		for _, ng := range p.NodeGroups() {
			if ng.Id() == ngName {
				targetSize, err := ng.TargetSize()
				assert.Nil(t, err)
				assert.Equal(t, 3, targetSize)
			}
		}
	})
}

func TestGetResourceLimiter(t *testing.T) {
	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		}

		if getAction.GetName() == defaultTemplatesConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplates,
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

	ko := &kwokOptions{
		kubeClient:      fakeClient,
		autoscalingOpts: &config.AutoscalingOptions{},
		discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
		resourceLimiter: cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
			map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
		allNodesLister: fakeNodeLister,
		ngNodeListerFn: testNodeLister,
	}

	p, err := BuildKwokProvider(ko)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// usual case
	cp, err := p.GetResourceLimiter()
	assert.Nil(t, err)
	assert.NotNil(t, cp)

	// resource limiter is nil
	ko.resourceLimiter = nil
	p, err = BuildKwokProvider(ko)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	cp, err = p.GetResourceLimiter()
	assert.Nil(t, err)
	assert.Nil(t, cp)

}

func TestGetAvailableGPUTypes(t *testing.T) {
	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		}

		if getAction.GetName() == defaultTemplatesConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplates,
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

	ko := &kwokOptions{
		kubeClient:      fakeClient,
		autoscalingOpts: &config.AutoscalingOptions{},
		discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
		resourceLimiter: cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
			map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
		allNodesLister: fakeNodeLister,
		ngNodeListerFn: testNodeLister,
	}

	p, err := BuildKwokProvider(ko)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// usual case
	l := p.GetAvailableGPUTypes()
	assert.NotNil(t, l)
	assert.Equal(t, map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {}}, l)

	// kwok provider config is nil
	kwokProviderConfigBackup := p.config
	p.config = nil
	l = p.GetAvailableGPUTypes()
	assert.Empty(t, l)

	// kwok provider config.status is nil
	p.config = kwokProviderConfigBackup
	statusBackup := p.config.status
	p.config.status = nil
	l = p.GetAvailableGPUTypes()
	assert.Empty(t, l)
	p.config.status = statusBackup
}

func TestGetNodeGpuConfig(t *testing.T) {
	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		}

		if getAction.GetName() == defaultTemplatesConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplates,
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

	ko := &kwokOptions{
		kubeClient:      fakeClient,
		autoscalingOpts: &config.AutoscalingOptions{},
		discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
		resourceLimiter: cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
			map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
		allNodesLister: fakeNodeLister,
		ngNodeListerFn: testNodeLister,
	}

	p, err := BuildKwokProvider(ko)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	nodeWithGPU := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"k8s.amazonaws.com/accelerator": "nvidia-tesla-k80",
			},
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				gpu.ResourceNvidiaGPU: resource.MustParse("2Gi"),
			},
		},
	}
	l := p.GetNodeGpuConfig(nodeWithGPU)
	assert.NotNil(t, l)
	assert.Equal(t, "k8s.amazonaws.com/accelerator", l.Label)
	assert.Equal(t, gpu.ResourceNvidiaGPU, string(l.ExtendedResourceName))
	assert.Equal(t, "nvidia-tesla-k80", l.Type)

	nodeWithNoAllocatableGPU := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"k8s.amazonaws.com/accelerator": "nvidia-tesla-k80",
			},
		},
	}
	l = p.GetNodeGpuConfig(nodeWithNoAllocatableGPU)
	assert.NotNil(t, l)
	assert.Equal(t, "k8s.amazonaws.com/accelerator", l.Label)
	assert.Equal(t, gpu.ResourceNvidiaGPU, string(l.ExtendedResourceName))
	assert.Equal(t, "nvidia-tesla-k80", l.Type)

	nodeWithNoGPULabel := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				gpu.ResourceNvidiaGPU: resource.MustParse("2Gi"),
			},
		},
	}
	l = p.GetNodeGpuConfig(nodeWithNoGPULabel)
	assert.NotNil(t, l)
	assert.Equal(t, "k8s.amazonaws.com/accelerator", l.Label)
	assert.Equal(t, gpu.ResourceNvidiaGPU, string(l.ExtendedResourceName))
	assert.Equal(t, "", l.Type)

}

func TestGPULabel(t *testing.T) {
	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		}

		if getAction.GetName() == defaultTemplatesConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplates,
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

	ko := &kwokOptions{
		kubeClient:      fakeClient,
		autoscalingOpts: &config.AutoscalingOptions{},
		discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
		resourceLimiter: cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
			map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
		allNodesLister: fakeNodeLister,
		ngNodeListerFn: testNodeLister,
	}

	p, err := BuildKwokProvider(ko)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// usual case
	l := p.GPULabel()
	assert.Equal(t, "k8s.amazonaws.com/accelerator", l)

	// kwok provider config is nil
	kwokProviderConfigBackup := p.config
	p.config = nil
	l = p.GPULabel()
	assert.Empty(t, l)

	// kwok provider config.status is nil
	p.config = kwokProviderConfigBackup
	statusBackup := p.config.status
	p.config.status = nil
	l = p.GPULabel()
	assert.Empty(t, l)
	p.config.status = statusBackup

}

func TestNodeGroupForNode(t *testing.T) {
	fakeClient := &fake.Clientset{}
	var nodesFrom string
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						configKey: testConfig,
					},
				}, nil
			}

			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplates,
				},
			}, nil

		}

		if getAction.GetName() == defaultTemplatesConfigName {
			if nodesFrom == "configmap" {
				return true, &apiv1.ConfigMap{
					Data: map[string]string{
						templatesKey: testTemplates,
					},
				}, nil
			}
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	t.Run("use template nodes from the configmap", func(t *testing.T) {
		nodesFrom = "configmap"
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Len(t, p.nodeGroups, 1)
		assert.Contains(t, p.nodeGroups[0].Id(), "kind-worker")

		testNode := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"kubernetes.io/hostname":        "kind-worker",
					"k8s.amazonaws.com/accelerator": "nvidia-tesla-k80",
					"kwok-nodegroup":                "kind-worker",
				},
				Name: "kind-worker",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: "kwok:kind-worker-m24xz",
			},
		}
		ng, err := p.NodeGroupForNode(testNode)
		assert.NoError(t, err)
		assert.NotNil(t, ng)
		assert.Contains(t, ng.Id(), "kind-worker")
	})

	t.Run("use template nodes from the cluster (aka get them using kube client)", func(t *testing.T) {
		nodesFrom = "cluster"
		fakeNode := &apiv1.Node{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Node",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
				Labels: map[string]string{
					"kwok-nodegroup": "kind-worker",
				},
			},
			Spec: apiv1.NodeSpec{},
		}
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{
			"kind-worker": fakeNode,
		})

		fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
			getAction := action.(core.ListAction)

			if getAction == nil {
				return false, nil, nil
			}

			return true, &apiv1.NodeList{Items: []apiv1.Node{*fakeNode}}, nil
		})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Len(t, p.nodeGroups, 1)
		assert.Contains(t, p.nodeGroups[0].Id(), "kind-worker")

		testNode := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"kubernetes.io/hostname":        "kind-worker",
					"k8s.amazonaws.com/accelerator": "nvidia-tesla-k80",
					"kwok-nodegroup":                "kind-worker",
				},
				Name: "kind-worker",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: "kwok:kind-worker-m24xz",
			},
		}
		ng, err := p.NodeGroupForNode(testNode)
		assert.NoError(t, err)
		assert.NotNil(t, ng)
		assert.Contains(t, ng.Id(), "kind-worker")
	})

	t.Run("empty nodegroup name for node", func(t *testing.T) {
		nodesFrom = "configmap"
		fakeNodeLister := newTestAllNodeLister(map[string]*apiv1.Node{})

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: testNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Len(t, p.nodeGroups, 1)
		assert.Contains(t, p.nodeGroups[0].Id(), "kind-worker")

		testNode := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-without-labels",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: "kwok:random-instance-id",
			},
		}
		ng, err := p.NodeGroupForNode(testNode)
		assert.NoError(t, err)
		assert.Nil(t, ng)
	})

}

func TestBuildKwokProvider(t *testing.T) {
	defer func() {
		os.Unsetenv("KWOK_PROVIDER_CONFIGMAP")
	}()

	fakeClient := &fake.Clientset{}

	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		switch getAction.GetName() {
		case defaultConfigName:
			// for nodesFrom: configmap
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		case "testConfigDynamicTemplatesSkipTaint":
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplatesSkipTaint,
				},
			}, nil
		case "testConfigDynamicTemplates":
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplates,
				},
			}, nil

		case "testConfigSkipTaint":
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigSkipTaint,
				},
			}, nil

		case defaultTemplatesConfigName:
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplatesMinimal,
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeNode1 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Annotations: map[string]string{
				NGNameAnnotation: "ng1",
			},
			Labels: map[string]string{
				"kwok-nodegroup": "ng1",
			},
		},
	}

	fakeNode2 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			Annotations: map[string]string{
				NGNameAnnotation: "ng2",
			},
			Labels: map[string]string{
				"kwok-nodegroup": "ng2",
			},
		},
	}

	fakeNode3 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "node3",
			Annotations: map[string]string{},
			// not a node that should be managed by kwok provider
			Labels: map[string]string{},
		},
	}

	fakeNodes := make(map[string]*apiv1.Node)
	fakeNodeLister := newTestAllNodeLister(fakeNodes)

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		listAction := action.(core.ListAction)
		if listAction == nil {
			return false, nil, nil
		}

		nodes := []apiv1.Node{}
		for _, node := range fakeNodes {
			nodes = append(nodes, *node)
		}
		return true, &apiv1.NodeList{Items: nodes}, nil
	})

	ko := &kwokOptions{
		kubeClient:      fakeClient,
		autoscalingOpts: &config.AutoscalingOptions{},
		discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
		resourceLimiter: cloudprovider.NewResourceLimiter(
			map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
			map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
		allNodesLister: fakeNodeLister,
		ngNodeListerFn: testNodeLister,
	}

	os.Setenv("POD_NAMESPACE", "kube-system")

	t.Run("(don't skip adding taint) use template nodes from the cluster (aka get them using kube client)", func(t *testing.T) {
		// use template nodes from the cluster (aka get them using kube client)
		fakeNodes = map[string]*apiv1.Node{fakeNode1.Name: &fakeNode1, fakeNode2.Name: &fakeNode2, fakeNode3.Name: &fakeNode3}
		fakeNodeLister.setNodesMap(fakeNodes)
		os.Setenv("KWOK_PROVIDER_CONFIGMAP", "testConfigDynamicTemplates")

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.nodeGroups)
		assert.Len(t, p.nodeGroups, 2)
		assert.NotNil(t, p.kubeClient)
		assert.NotNil(t, p.resourceLimiter)
		assert.NotNil(t, p.config)

		for i := range p.nodeGroups {
			assert.NotNil(t, p.nodeGroups[i].nodeTemplate)
			assert.Equal(t, kwokProviderTaint(), p.nodeGroups[i].nodeTemplate.Spec.Taints[0])
		}
	})

	t.Run("(skip adding taint) use template nodes from the cluster (aka get them using kube client)", func(t *testing.T) {
		// use template nodes from the cluster (aka get them using kube client)
		fakeNodes = map[string]*apiv1.Node{fakeNode1.Name: &fakeNode1, fakeNode2.Name: &fakeNode2, fakeNode3.Name: &fakeNode3}
		fakeNodeLister.setNodesMap(fakeNodes)
		os.Setenv("KWOK_PROVIDER_CONFIGMAP", "testConfigDynamicTemplatesSkipTaint")

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.nodeGroups)
		assert.Len(t, p.nodeGroups, 2)
		assert.NotNil(t, p.kubeClient)
		assert.NotNil(t, p.resourceLimiter)
		assert.NotNil(t, p.config)

		for i := range p.nodeGroups {
			assert.NotNil(t, p.nodeGroups[i].nodeTemplate)
			assert.Empty(t, p.nodeGroups[i].nodeTemplate.Spec.Taints)
		}
	})

	t.Run("(don't skip adding taint) use template nodes from the configmap", func(t *testing.T) {
		// use template nodes from the configmap
		fakeNodes = map[string]*apiv1.Node{}

		nos, err := LoadNodeTemplatesFromConfigMap(defaultTemplatesConfigName, fakeClient)
		assert.NoError(t, err)
		assert.NotEmpty(t, nos)

		for i := range nos {
			fakeNodes[nos[i].GetName()] = nos[i]
		}
		fakeNodeLister = newTestAllNodeLister(fakeNodes)

		// fallback to default configmap name
		os.Unsetenv("KWOK_PROVIDER_CONFIGMAP")

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.nodeGroups)
		assert.Len(t, p.nodeGroups, 2)
		assert.NotNil(t, p.kubeClient)
		assert.NotNil(t, p.resourceLimiter)
		assert.NotNil(t, p.config)

		for i := range p.nodeGroups {
			assert.NotNil(t, p.nodeGroups[i].nodeTemplate)
			assert.Equal(t, kwokProviderTaint(), p.nodeGroups[i].nodeTemplate.Spec.Taints[0])
		}
	})

	t.Run("(skip adding taint) use template nodes from the configmap", func(t *testing.T) {
		// use template nodes from the configmap
		fakeNodes = map[string]*apiv1.Node{}

		nos, err := LoadNodeTemplatesFromConfigMap(defaultTemplatesConfigName, fakeClient)
		assert.NoError(t, err)
		assert.NotEmpty(t, nos)

		for i := range nos {
			fakeNodes[nos[i].GetName()] = nos[i]
		}

		os.Setenv("KWOK_PROVIDER_CONFIGMAP", "testConfigSkipTaint")

		fakeNodeLister = newTestAllNodeLister(fakeNodes)

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotNil(t, p.nodeGroups)
		assert.Len(t, p.nodeGroups, 2)
		assert.NotNil(t, p.kubeClient)
		assert.NotNil(t, p.resourceLimiter)
		assert.NotNil(t, p.config)

		for i := range p.nodeGroups {
			assert.NotNil(t, p.nodeGroups[i].nodeTemplate)
			assert.Empty(t, p.nodeGroups[i].nodeTemplate.Spec.Taints)
		}
	})
}

func TestCleanup(t *testing.T) {

	defer func() {
		os.Unsetenv("KWOK_PROVIDER_CONFIGMAP")
	}()

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		switch getAction.GetName() {
		case defaultConfigName:
			// for nodesFrom: configmap
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		case "testConfigDynamicTemplates":
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigDynamicTemplates,
				},
			}, nil
		case defaultTemplatesConfigName:
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplatesMinimal,
				},
			}, nil

		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeNode1 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Annotations: map[string]string{
				NGNameAnnotation: "ng1",
			},
			Labels: map[string]string{
				"kwok-nodegroup": "ng1",
			},
		},
	}

	fakeNode2 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			Annotations: map[string]string{
				NGNameAnnotation: "ng2",
			},
			Labels: map[string]string{
				"kwok-nodegroup": "ng2",
			},
		},
	}

	fakeNode3 := apiv1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "node3",
			Annotations: map[string]string{},
			// not a node that should be managed by kwok provider
			Labels: map[string]string{},
		},
	}

	fakeNodes := make(map[string]*apiv1.Node)
	fakeNodeLister := newTestAllNodeLister(fakeNodes)

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		listAction := action.(core.ListAction)
		if listAction == nil {
			return false, nil, nil
		}

		nodes := []apiv1.Node{}
		for _, node := range fakeNodes {
			nodes = append(nodes, *node)
		}
		return true, &apiv1.NodeList{Items: nodes}, nil
	})

	fakeClient.Fake.AddReactor("delete", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		deleteAction := action.(core.DeleteAction)

		if deleteAction == nil {
			return false, nil, nil
		}

		if fakeNodes[deleteAction.GetName()] != nil {
			delete(fakeNodes, deleteAction.GetName())
		}

		fakeNodeLister.setNodesMap(fakeNodes)

		return false, nil, errors.NewNotFound(apiv1.Resource("nodes"), deleteAction.GetName())

	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	t.Run("use template nodes from the cluster (aka get them using kube client)", func(t *testing.T) {
		// use template nodes from the cluster (aka get them using kube client)
		fakeNodes = map[string]*apiv1.Node{fakeNode1.Name: &fakeNode1, fakeNode2.Name: &fakeNode2, fakeNode3.Name: &fakeNode3}
		fakeNodeLister.setNodesMap(fakeNodes)
		os.Setenv("KWOK_PROVIDER_CONFIGMAP", "testConfigDynamicTemplates")

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: kube_util.NewNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotEmpty(t, p.nodeGroups)

		err = p.Cleanup()
		assert.NoError(t, err)
		nodeList, err := fakeNodeLister.List(labels.NewSelector())
		assert.NoError(t, err)
		assert.Len(t, nodeList, 1)
		assert.Equal(t, fakeNode3, *nodeList[0])
	})

	t.Run("use template nodes from the configmap", func(t *testing.T) {
		// use template nodes from the configmap
		fakeNodes = map[string]*apiv1.Node{}

		nos, err := LoadNodeTemplatesFromConfigMap(defaultTemplatesConfigName, fakeClient)
		assert.NoError(t, err)
		assert.NotEmpty(t, nos)

		for i := range nos {
			fakeNodes[nos[i].GetName()] = nos[i]
		}

		// fallback to default configmap name
		os.Unsetenv("KWOK_PROVIDER_CONFIGMAP")
		fakeNodeLister = newTestAllNodeLister(fakeNodes)

		ko := &kwokOptions{
			kubeClient:      fakeClient,
			autoscalingOpts: &config.AutoscalingOptions{},
			discoveryOpts:   &cloudprovider.NodeGroupDiscoveryOptions{},
			resourceLimiter: cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000}),
			allNodesLister: fakeNodeLister,
			ngNodeListerFn: kube_util.NewNodeLister,
		}

		p, err := BuildKwokProvider(ko)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.NotEmpty(t, p.nodeGroups)

		err = p.Cleanup()
		assert.NoError(t, err)
		nodeList, err := fakeNodeLister.List(labels.NewSelector())
		assert.NoError(t, err)
		assert.Len(t, nodeList, 1)
		assert.Equal(t, fakeNode3, *nodeList[0])
	})

}

func testNodeLister(lister v1lister.NodeLister, filter func(*apiv1.Node) bool) kube_util.NodeLister {
	return kube_util.NewTestNodeLister(nil)
}

// fakeAllNodeLister implements v1lister.NodeLister interface
type fakeAllNodeLister struct {
	nodesMap map[string]*apiv1.Node
}

func newTestAllNodeLister(nodesMap map[string]*apiv1.Node) *fakeAllNodeLister {
	return &fakeAllNodeLister{nodesMap: nodesMap}
}

func (f *fakeAllNodeLister) List(_ labels.Selector) (ret []*apiv1.Node, err error) {
	n := []*apiv1.Node{}

	for _, node := range f.nodesMap {
		n = append(n, node)
	}

	return n, nil
}

func (f *fakeAllNodeLister) Get(name string) (*apiv1.Node, error) {
	if f.nodesMap[name] == nil {
		return nil, errors.NewNotFound(apiv1.Resource("nodes"), name)
	}
	return f.nodesMap[name], nil
}

func (f *fakeAllNodeLister) setNodesMap(nodesMap map[string]*apiv1.Node) {
	f.nodesMap = nodesMap
}
