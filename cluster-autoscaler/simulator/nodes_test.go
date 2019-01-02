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

package simulator

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/kubernetes/pkg/kubelet/types"

	"github.com/stretchr/testify/assert"
)

func TestRequiredPodsForNode(t *testing.T) {
	nodeName := "node1"
	pod1 := apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
			SelfLink:  "pod1",
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}
	// Manifest pod.
	pod2 := apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "kube-system",
			SelfLink:  "pod2",
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "something",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: nodeName,
		},
	}

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{&pod1, &pod2})
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	pods, err := GetRequiredPodsForNode(nodeName, registry)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
	assert.Equal(t, "pod2", pods[0].Name)
}
