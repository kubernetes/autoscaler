/*
Copyright 2021 The Kubernetes Authors.

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

package podtemplate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func Test_podTemplateProcessor_GetDaemonSetPodsFromPodTemplateForNode(t *testing.T) {
	type fields struct {
		podTemplateLister v1lister.PodTemplateLister
	}
	type args struct {
		baseNodeInfo  *schedulerframework.NodeInfo
		ignoredTaints taints.TaintKeySet
	}

	namespace := "bar"
	podName := "foo-dfsdfds"
	nodeName1 := "template-node-template-for-node-1"
	nodePod1 := newTestPod("bar", "foo", &apiv1.PodSpec{}, nodeName1)
	nodeInfo := newNodeInfo(nodeName1, 42, nodePod1)

	podTemplate1 := newPodTemplate(namespace, podName, nil)

	nodeInfoUnschedulable := newNodeInfo(nodeName1, 0, nodePod1)

	tests := []struct {
		name              string
		fields            fields
		args              args
		podListerCreation func() (v1lister.PodTemplateLister, error)
		want              []*apiv1.Pod
		wantErr           bool
	}{
		{
			name:   "No PodTemplate present in the cluster",
			fields: fields{},
			args: args{
				baseNodeInfo: nodeInfo.Clone(),
			},
			podListerCreation: func() (v1lister.PodTemplateLister, error) {
				return newTestDaemonSetLister(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "One PodTemplate, that match the node",
			fields: fields{},
			args: args{
				baseNodeInfo: nodeInfo.Clone(),
			},
			podListerCreation: func() (v1lister.PodTemplateLister, error) {
				tpls := []*apiv1.PodTemplate{podTemplate1}
				return newTestDaemonSetLister(tpls)
			},
			want:    []*apiv1.Pod{newPod(podTemplate1, nodeName1)},
			wantErr: false,
		},
		{
			name:   "Unschedulable Node",
			fields: fields{},
			args: args{
				baseNodeInfo: nodeInfoUnschedulable.Clone(),
			},
			podListerCreation: func() (v1lister.PodTemplateLister, error) {
				tpls := []*apiv1.PodTemplate{podTemplate1}
				return newTestDaemonSetLister(tpls)
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister := tt.fields.podTemplateLister
			if tt.podListerCreation != nil {
				var err error
				lister, err = tt.podListerCreation()
				assert.NoError(t, err, "err should be nil")
			}
			p := &podTemplateProcessor{
				podTemplateLister: lister,
			}

			predicateChecker, _ := predicatechecker.NewTestPredicateChecker()

			got, err := p.GetDaemonSetPodsFromPodTemplateForNode(tt.args.baseNodeInfo, predicateChecker, tt.args.ignoredTaints)
			if (err != nil) != tt.wantErr {
				t.Errorf("podTemplateProcessor.GetDaemonSetPodsFromPodTemplateForNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got, "podTemplateProcessor.GetDaemonSetPodsFromPodTemplateForNode() wrong expected value")
		})
	}
}

func newTestDaemonSetLister(pts []*apiv1.PodTemplate) (v1lister.PodTemplateLister, error) {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, pt := range pts {
		err := store.Add(pt)
		if err != nil {
			return nil, fmt.Errorf("Error adding object to cache: %v", err)
		}
	}
	return v1lister.NewPodTemplateLister(store), nil
}

type podTemplateListerMock struct {
	mock.Mock
}

// List lists all PodTemplates in the indexer.
func (p *podTemplateListerMock) List(selector labels.Selector) (ret []*apiv1.PodTemplate, err error) {
	args := p.Called()
	return args.Get(0).([]*apiv1.PodTemplate), args.Error(1)
}

// PodTemplates returns an object that can list and get PodTemplates.
func (p *podTemplateListerMock) PodTemplates(namespace string) v1lister.PodTemplateNamespaceLister {
	args := p.Called(namespace)
	return args.Get(0).(v1lister.PodTemplateNamespaceLister)
}

func newNode(name string, maxPods int64) *apiv1.Node {
	// define a resource list to allow pod scheduling on this node.
	newResourceList := apiv1.ResourceList{
		apiv1.ResourcePods: *resource.NewQuantity(maxPods, resource.DecimalSI),
	}
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"kubernetes.io/hostname": name,
			},
		},
		Status: apiv1.NodeStatus{
			Allocatable: newResourceList,
		},
	}
}

func newNodeInfo(nodeName string, maxPod int64, pods ...*apiv1.Pod) *schedulerframework.NodeInfo {
	node1 := newNode(nodeName, maxPod)
	nodeInfo := schedulerframework.NewNodeInfo(pods...)
	nodeInfo.SetNode(node1)

	return nodeInfo
}

func newTestPod(namespace, name string, spec *apiv1.PodSpec, nodeName string) *apiv1.Pod {
	newPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	newPod.Namespace = namespace
	newPod.Name = name
	if spec != nil {
		newPod.Spec = *spec
	}
	newPod.Spec.NodeName = nodeName
	return newPod
}

func newPodTemplate(namespace, name string, spec *apiv1.PodTemplateSpec) *apiv1.PodTemplate {
	pt := &apiv1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if spec != nil {
		pt.Template = *spec
	}

	return pt
}
