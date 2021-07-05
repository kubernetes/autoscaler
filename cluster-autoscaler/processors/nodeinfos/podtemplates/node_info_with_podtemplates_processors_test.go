/*
Copyright 2019 The Kubernetes Authors.

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

package podtemplates

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	scheduler_utils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
)

func Test_getNodeInfoWithPodTemplates(t *testing.T) {
	nodeName1 := "template-node-template-for-node-1"
	nodePod1 := newTestPod("bar", "foo", &apiv1.PodSpec{}, nodeName1)
	nodeInfo := newNodeInfo(nodeName1, 42, nodePod1)
	nodeInfoUnschedulable := newNodeInfo(nodeName1, 0, nodePod1)

	type args struct {
		baseNodeInfo *schedulerframework.NodeInfo
		podTemplates []*apiv1.PodTemplate
	}
	tests := []struct {
		name     string
		args     args
		wantFunc func() *schedulerframework.NodeInfo
		wantErr  bool
	}{
		{
			name: "nodeInfo should not be updated",
			args: args{
				baseNodeInfo: nodeInfo.Clone(),
				podTemplates: nil,
			},
			wantFunc: func() *schedulerframework.NodeInfo {
				return scheduler_utils.DeepCopyTemplateNode(nodeInfo, nodeInfoDeepCopySuffix)
			},
		},
		{
			name: "nodeInfo contains one additional Pod",
			args: args{
				baseNodeInfo: nodeInfo.Clone(),
				podTemplates: []*apiv1.PodTemplate{
					newPodTemplate("extra-ns", "extra-name", nil),
				},
			},
			wantFunc: func() *schedulerframework.NodeInfo {
				nodeInfo := scheduler_utils.DeepCopyTemplateNode(nodeInfo, nodeInfoDeepCopySuffix)
				nodeInfo.AddPod(newTestPod("extra-ns", "extra-name-pod", nil, nodeName1+"-podtemplate"))
				return nodeInfo
			},
		},
		{
			name: "nodeInfo unschedulable",
			args: args{
				baseNodeInfo: nodeInfoUnschedulable.Clone(),
				podTemplates: []*apiv1.PodTemplate{
					newPodTemplate("extra-ns", "extra-name", nil),
				},
			},
			wantFunc: func() *schedulerframework.NodeInfo {
				return scheduler_utils.DeepCopyTemplateNode(nodeInfoUnschedulable, nodeInfoDeepCopySuffix)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterSnapshot := simulator.NewBasicClusterSnapshot()
			predicateChecker, _ := simulator.NewTestPredicateChecker()

			got, err := getNodeInfoWithPodTemplates(tt.args.baseNodeInfo, tt.args.podTemplates, clusterSnapshot, predicateChecker)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeInfoWithPodTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := tt.wantFunc()

			got.Generation = want.Generation

			resetNodeInfoGeneratedFields(got)
			resetNodeInfoGeneratedFields(want)
			assert.EqualValues(t, want, got, "getNodeInfoWithPodTemplates wrong expected value")
		})
	}
}

func Test_nodeInfoWithPodTemplateProcessor_Process(t *testing.T) {
	namespace := "bar"
	podName := "foo-dfsdfds"
	nodeName1 := "template-node-for-template-node-1"
	wantNodeName1 := fmt.Sprintf("%s-%s", nodeName1, nodeInfoDeepCopySuffix)
	nodeNameCopyNode := "template-node-for-copy-node-0"

	tests := []struct {
		name                   string
		podTemplates           []*apiv1.PodTemplate
		podListerCreation      func(pts []*apiv1.PodTemplate) (v1lister.PodTemplateLister, error)
		nodeInfosForNodeGroups map[string]*schedulerframework.NodeInfo
		want                   map[string]*schedulerframework.NodeInfo
		wantErr                bool
	}{
		{
			name:              "1 pod added",
			podTemplates:      []*apiv1.PodTemplate{newPodTemplate(namespace, podName, nil)},
			podListerCreation: newTestDaemonSetLister,
			nodeInfosForNodeGroups: map[string]*schedulerframework.NodeInfo{
				nodeName1: newNodeInfo(nodeName1, 42),
			},
			want: map[string]*schedulerframework.NodeInfo{
				nodeName1: newNodeInfo(wantNodeName1, 42, newTestPod(namespace, fmt.Sprintf("%s-pod", podName), nil, wantNodeName1)),
			},
			wantErr: false,
		},
		{
			name:              "0 pod added: unschedulable node",
			podTemplates:      []*apiv1.PodTemplate{newPodTemplate(namespace, podName, nil)},
			podListerCreation: newTestDaemonSetLister,
			nodeInfosForNodeGroups: map[string]*schedulerframework.NodeInfo{
				nodeName1: newNodeInfo(nodeName1, 0),
			},
			want: map[string]*schedulerframework.NodeInfo{
				nodeName1: newNodeInfo(wantNodeName1, 0),
			},
			wantErr: false,
		},
		{
			name:              "0 pod added: real node",
			podTemplates:      []*apiv1.PodTemplate{newPodTemplate(namespace, podName, nil)},
			podListerCreation: newTestDaemonSetLister,
			nodeInfosForNodeGroups: map[string]*schedulerframework.NodeInfo{
				nodeNameCopyNode: newNodeInfo(nodeNameCopyNode, 0),
			},
			want: map[string]*schedulerframework.NodeInfo{
				nodeNameCopyNode: newNodeInfo(nodeNameCopyNode, 0),
			},
			wantErr: false,
		},

		{
			name:         "pod lister error",
			podTemplates: []*apiv1.PodTemplate{newPodTemplate(namespace, podName, nil)},
			podListerCreation: func(pts []*apiv1.PodTemplate) (v1lister.PodTemplateLister, error) {
				podLister := &podTemplateListerMock{}
				podLister.On("List").Return(pts, fmt.Errorf("unable to list"))
				return podLister, nil
			},
			nodeInfosForNodeGroups: map[string]*schedulerframework.NodeInfo{
				nodeName1: newNodeInfo(nodeName1, 0),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podLister, err := tt.podListerCreation(tt.podTemplates)
			assert.NoError(t, err, "err should be nil")

			ctx, cancelFunc := context.WithCancel(context.Background())

			p := &nodeInfoWithPodTemplateProcessor{
				podTemplateLister: podLister,
				ctx:               ctx,
				cancelFunc:        cancelFunc,
			}

			caCtx, err := newTestClusterAutoscalerContext()
			assert.NoError(t, err, "err should be nil")

			got, err := p.Process(caCtx, tt.nodeInfosForNodeGroups)

			resetNodeInfosGeneratedFields(got)
			resetNodeInfosGeneratedFields(tt.want)

			if (err != nil) != tt.wantErr {
				t.Errorf("nodeInfoWithPodTemplateProcessor.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got, "nodeInfoWithPodTemplateProcessor.Process() wrong expected value")
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

func newTestClusterAutoscalerContext() (*ca_context.AutoscalingContext, error) {
	predicateChecker, err := simulator.NewTestPredicateChecker()
	if err != nil {
		return nil, err
	}

	ctx := &ca_context.AutoscalingContext{
		PredicateChecker: predicateChecker,
	}
	return ctx, nil
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

func newNodeInfo(nodeName string, maxPod int64, pods ...*v1.Pod) *schedulerframework.NodeInfo {
	node1 := newNode(nodeName, maxPod)
	nodeInfo := schedulerframework.NewNodeInfo(pods...)
	nodeInfo.SetNode(node1)

	return nodeInfo
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

func resetNodeInfosGeneratedFields(nodeInfos map[string]*schedulerframework.NodeInfo) {
	for _, nodeInfo := range nodeInfos {
		resetNodeInfoGeneratedFields(nodeInfo)
	}
}

func resetNodeInfoGeneratedFields(nodeInfo *schedulerframework.NodeInfo) {
	nodeInfo.Generation = 0
	nodeInfo.Node().UID = ""
	for _, podInfo := range nodeInfo.Pods {
		podInfo.Pod.UID = ""
	}
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
