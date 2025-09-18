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
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

var testConfigs = map[string]string{
	defaultConfigName:          testConfig,
	"without-kwok":             withoutKwok,
	"with-static-kwok-release": withStaticKwokRelease,
	"skip-kwok-install":        skipKwokInstall,
}

// with node templates from configmap
const testConfig = `
apiVersion: v1alpha1
readNodesFrom: configmap # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "kwok-nodegroup"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
kwok: {}
`

// with node templates from configmap
const testConfigSkipTaint = `
apiVersion: v1alpha1
readNodesFrom: configmap # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "kwok-nodegroup"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  skipTaint: true
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
kwok: {}
`
const testConfigDynamicTemplates = `
apiVersion: v1alpha1
readNodesFrom: cluster # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "kwok-nodegroup"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
kwok: {}
`

const testConfigDynamicTemplatesSkipTaint = `
apiVersion: v1alpha1
readNodesFrom: cluster # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "kwok-nodegroup"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  skipTaint: true
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
kwok: {}
`

const withoutKwok = `
apiVersion: v1alpha1
readNodesFrom: configmap # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "node.kubernetes.io/instance-type"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
`

const withStaticKwokRelease = `
apiVersion: v1alpha1
readNodesFrom: configmap # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "node.kubernetes.io/instance-type"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
kwok:
  release: "v0.2.1"
configmap:
  name: kwok-provider-templates
`

const skipKwokInstall = `
apiVersion: v1alpha1
readNodesFrom: configmap # possible values: [cluster,configmap]
nodegroups:
  # to specify how to group nodes into a nodegroup
  # e.g., you want to treat nodes with same instance type as a nodegroup
  # node1: m5.xlarge
  # node2: c5.xlarge
  # node3: m5.xlarge
  # nodegroup1: [node1,node3]
  # nodegroup2: [node2]
  fromNodeLabelKey: "node.kubernetes.io/instance-type"
  # you can either specify fromNodeLabelKey OR fromNodeAnnotationKey
  # (both are not allowed)
  # fromNodeAnnotationKey: "eks.amazonaws.com/nodegroup"
nodes:
  gpuConfig:
    # to tell kwok provider what label should be considered as GPU label
    gpuLabelKey: "k8s.amazonaws.com/accelerator"
    availableGPUTypes:
      "nvidia-tesla-k80": {}
      "nvidia-tesla-p100": {}
configmap:
  name: kwok-provider-templates
kwok:
  skipInstall: true
`

func TestLoadConfigFile(t *testing.T) {
	defer func() {
		os.Unsetenv("KWOK_PROVIDER_CONFIGMAP")
	}()

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		cmName := getConfigMapName()
		if getAction.GetName() == cmName {
			return true, &v1.ConfigMap{
				Data: map[string]string{
					configKey: testConfigs[cmName],
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(v1.Resource("configmaps"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	kwokConfig, err := LoadConfigFile(fakeClient)
	assert.Nil(t, err)
	assert.NotNil(t, kwokConfig)
	assert.NotNil(t, kwokConfig.status)
	assert.NotEmpty(t, kwokConfig.status.gpuLabel)

	os.Setenv("KWOK_PROVIDER_CONFIGMAP", "without-kwok")
	kwokConfig, err = LoadConfigFile(fakeClient)
	assert.Nil(t, err)
	assert.NotNil(t, kwokConfig)
	assert.NotNil(t, kwokConfig.status)
	assert.NotEmpty(t, kwokConfig.status.gpuLabel)

	os.Setenv("KWOK_PROVIDER_CONFIGMAP", "with-static-kwok-release")
	kwokConfig, err = LoadConfigFile(fakeClient)
	assert.Nil(t, err)
	assert.NotNil(t, kwokConfig)
	assert.NotNil(t, kwokConfig.status)
	assert.NotEmpty(t, kwokConfig.status.gpuLabel)

	os.Setenv("KWOK_PROVIDER_CONFIGMAP", "skip-kwok-install")
	kwokConfig, err = LoadConfigFile(fakeClient)
	assert.Nil(t, err)
	assert.NotNil(t, kwokConfig)
	assert.NotNil(t, kwokConfig.status)
	assert.NotEmpty(t, kwokConfig.status.gpuLabel)
}
