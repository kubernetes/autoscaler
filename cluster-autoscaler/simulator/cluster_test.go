/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestReservation(t *testing.T) {
	pod := buildPod("p1", 100, 200000)
	pod2 := &kube_api.Pod{
		Spec: kube_api.PodSpec{
			Containers: []kube_api.Container{
				{
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{},
					},
				},
			},
		},
	}
	nodeInfo := schedulercache.NewNodeInfo(pod, pod, pod2)

	node := &kube_api.Node{
		ObjectMeta: kube_api.ObjectMeta{
			Name: "node1",
		},
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
				kube_api.ResourceMemory: *resource.NewQuantity(2000000, resource.DecimalSI),
			},
		},
	}
	reservation, err := calculateReservation(node, nodeInfo)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, reservation, 0.01)

	node2 := &kube_api.Node{
		ObjectMeta: kube_api.ObjectMeta{
			Name: "node2",
		},
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourceCPU: *resource.NewMilliQuantity(2000, resource.DecimalSI),
			},
		},
	}
	_, err = calculateReservation(node2, nodeInfo)
	assert.Error(t, err)
}

func TestFindPlaceAllOk(t *testing.T) {
	pod1 := buildPod("p1", 300, 500000)
	new1 := buildPod("p2", 600, 500000)
	new2 := buildPod("p3", 500, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1": schedulercache.NewNodeInfo(pod1),
		"n2": schedulercache.NewNodeInfo(),
	}
	node1 := buildNode("n1", 1000, 2000000)
	node2 := buildNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	err := findPlaceFor(
		[]*kube_api.Pod{new1, new2},
		[]*kube_api.Node{node1, node2},
		nodeInfos)
	assert.NoError(t, err)
}

func TestFindPlaceAllBas(t *testing.T) {
	pod1 := buildPod("p1", 300, 500000)
	new1 := buildPod("p2", 600, 500000)
	new2 := buildPod("p3", 500, 500000)
	new3 := buildPod("p4", 700, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1": schedulercache.NewNodeInfo(pod1),
		"n2": schedulercache.NewNodeInfo(),
	}
	node1 := buildNode("n1", 1000, 2000000)
	node2 := buildNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	err := findPlaceFor(
		[]*kube_api.Pod{new1, new2, new3},
		[]*kube_api.Node{node1, node2},
		nodeInfos)
	assert.Error(t, err)
}

func TestFindNone(t *testing.T) {
	pod1 := buildPod("p1", 300, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1": schedulercache.NewNodeInfo(pod1),
		"n2": schedulercache.NewNodeInfo(),
	}
	node1 := buildNode("n1", 1000, 2000000)
	node2 := buildNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	err := findPlaceFor(
		[]*kube_api.Pod{},
		[]*kube_api.Node{node1, node2},
		nodeInfos)
	assert.NoError(t, err)
}

func buildPod(name string, cpu int64, mem int64) *kube_api.Pod {
	return &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
		Spec: kube_api.PodSpec{
			Containers: []kube_api.Container{
				{
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{
							kube_api.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.DecimalSI),
							kube_api.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
						},
					},
				},
			},
		},
	}
}

func buildNode(name string, cpu int64, mem int64) *kube_api.Node {
	return &kube_api.Node{
		ObjectMeta: kube_api.ObjectMeta{
			Name: name,
		},
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.DecimalSI),
				kube_api.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
				kube_api.ResourcePods:   *resource.NewQuantity(100, resource.DecimalSI),
			},
		},
	}
}
