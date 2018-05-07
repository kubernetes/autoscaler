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

package core

import (
	"fmt"
	"sort"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"strconv"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
)

func TestFindUnneededNodes(t *testing.T) {
	p1 := BuildTestPod("p1", 100, 0)
	p1.Spec.NodeName = "n1"

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p2 := BuildTestPod("p2", 300, 0)
	p2.Spec.NodeName = "n2"
	p2.OwnerReferences = ownerRef

	p3 := BuildTestPod("p3", 400, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.NodeName = "n3"

	p4 := BuildTestPod("p4", 2000, 0)
	p4.OwnerReferences = ownerRef
	p4.Spec.NodeName = "n4"

	p5 := BuildTestPod("p5", 100, 0)
	p5.OwnerReferences = ownerRef
	p5.Spec.NodeName = "n5"

	p6 := BuildTestPod("p6", 500, 0)
	p6.OwnerReferences = ownerRef
	p6.Spec.NodeName = "n7"

	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n3 := BuildTestNode("n3", 1000, 10)
	n4 := BuildTestNode("n4", 10000, 10)
	n5 := BuildTestNode("n5", 1000, 10)
	n5.Annotations = map[string]string{
		ScaleDownDisabledKey: "true",
	}
	n6 := BuildTestNode("n6", 1000, 10)
	n7 := BuildTestNode("n7", 0, 10) // Node without utilization
	n8 := BuildTestNode("n8", 1000, 10)
	n8.Spec.Taints = []apiv1.Taint{{Key: deletetaint.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}
	n9 := BuildTestNode("n9", 1000, 10)
	n9.Spec.Taints = []apiv1.Taint{{Key: deletetaint.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-60, 10)}}

	SetNodeReadyState(n1, true, time.Time{})
	SetNodeReadyState(n2, true, time.Time{})
	SetNodeReadyState(n3, true, time.Time{})
	SetNodeReadyState(n4, true, time.Time{})
	SetNodeReadyState(n5, true, time.Time{})
	SetNodeReadyState(n6, true, time.Time{})
	SetNodeReadyState(n7, true, time.Time{})
	SetNodeReadyState(n8, true, time.Time{})
	SetNodeReadyState(n9, true, time.Time{})

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)
	provider.AddNode("ng1", n4)
	provider.AddNode("ng1", n5)
	provider.AddNode("ng1", n7)
	provider.AddNode("ng1", n8)
	provider.AddNode("ng1", n9)

	context := AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		LogRecorder:          fakeLogRecorder,
		CloudProvider:        provider,
	}

	sd := NewScaleDown(&context)
	sd.UpdateUnneededNodes([]*apiv1.Node{n1, n2, n3, n4, n5, n7, n8, n9}, []*apiv1.Node{n1, n2, n3, n4, n5, n6, n7, n8, n9},
		[]*apiv1.Pod{p1, p2, p3, p4, p5, p6}, time.Now(), nil)

	assert.Equal(t, 3, len(sd.unneededNodes))
	addTime, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	addTime, found = sd.unneededNodes["n7"]
	assert.True(t, found)
	addTime, found = sd.unneededNodes["n8"]
	assert.True(t, found)
	assert.Contains(t, sd.podLocationHints, p2.Namespace+"/"+p2.Name)
	assert.Equal(t, 6, len(sd.nodeUtilizationMap))

	sd.unremovableNodes = make(map[string]time.Time)
	sd.unneededNodes["n1"] = time.Now()
	sd.UpdateUnneededNodes([]*apiv1.Node{n1, n2, n3, n4}, []*apiv1.Node{n1, n2, n3, n4}, []*apiv1.Pod{p1, p2, p3, p4}, time.Now(), nil)
	sd.unremovableNodes = make(map[string]time.Time)

	assert.Equal(t, 1, len(sd.unneededNodes))
	addTime2, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	assert.Equal(t, addTime, addTime2)
	assert.Equal(t, 4, len(sd.nodeUtilizationMap))

	sd.unremovableNodes = make(map[string]time.Time)
	sd.UpdateUnneededNodes([]*apiv1.Node{n1, n2, n3, n4}, []*apiv1.Node{n1, n3, n4}, []*apiv1.Pod{p1, p2, p3, p4}, time.Now(), nil)
	assert.Equal(t, 0, len(sd.unneededNodes))

	// Node n1 is unneeded, but should be skipped because it has just recently been found to be unremovable
	sd.UpdateUnneededNodes([]*apiv1.Node{n1}, []*apiv1.Node{n1}, []*apiv1.Pod{}, time.Now(), nil)
	assert.Equal(t, 0, len(sd.unneededNodes))
	// Verify that no other nodes are in unremovable map.
	assert.Equal(t, 1, len(sd.unremovableNodes))

	// But it should be checked after timeout
	sd.UpdateUnneededNodes([]*apiv1.Node{n1}, []*apiv1.Node{n1}, []*apiv1.Pod{}, time.Now().Add(UnremovableNodeRecheckTimeout+time.Second), nil)
	assert.Equal(t, 1, len(sd.unneededNodes))
	// Verify that nodes that are no longer unremovable are removed.
	assert.Equal(t, 0, len(sd.unremovableNodes))
}

func TestFindUnneededMaxCandidates(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 2)

	numNodes := 100
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)

	numCandidates := 30

	context := AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold:    0.35,
			ScaleDownNonEmptyCandidatesCount: numCandidates,
			ScaleDownCandidatesPoolRatio:     1,
			ScaleDownCandidatesPoolMinCount:  1000,
		},
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		LogRecorder:          fakeLogRecorder,
		CloudProvider:        provider,
	}
	sd := NewScaleDown(&context)

	sd.UpdateUnneededNodes(nodes, nodes, pods, time.Now(), nil)
	assert.Equal(t, numCandidates, len(sd.unneededNodes))
	// Simulate one of the unneeded nodes got deleted
	deleted := sd.unneededNodesList[len(sd.unneededNodesList)-1]
	for i, node := range nodes {
		if node.Name == deleted.Name {
			// Move pod away from the node
			var newNode int
			if i >= 1 {
				newNode = i - 1
			} else {
				newNode = i + 1
			}
			pods[i].Spec.NodeName = nodes[newNode].Name
			nodes[i] = nodes[len(nodes)-1]
			nodes[len(nodes)-1] = nil
			nodes = nodes[:len(nodes)-1]
			break
		}
	}

	sd.UpdateUnneededNodes(nodes, nodes, pods, time.Now(), nil)
	// Check that the deleted node was replaced
	assert.Equal(t, numCandidates, len(sd.unneededNodes))
	assert.NotContains(t, sd.unneededNodes, deleted)
}

func TestFindUnneededEmptyNodes(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 100)

	// 30 empty nodes and 70 heavily underutilized.
	numNodes := 100
	numEmpty := 30
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes-numEmpty; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)

	numCandidates := 30

	context := AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold:    0.35,
			ScaleDownNonEmptyCandidatesCount: numCandidates,
			ScaleDownCandidatesPoolRatio:     1.0,
			ScaleDownCandidatesPoolMinCount:  1000,
		},
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		LogRecorder:          fakeLogRecorder,
		CloudProvider:        provider,
	}
	sd := NewScaleDown(&context)

	sd.UpdateUnneededNodes(nodes, nodes, pods, time.Now(), nil)
	for _, node := range sd.unneededNodesList {
		t.Log(node.Name)
	}
	assert.Equal(t, numEmpty+numCandidates, len(sd.unneededNodes))
}

func TestFindUnneededNodePool(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 100)

	numNodes := 100
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)

	numCandidates := 30

	context := AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold:    0.35,
			ScaleDownNonEmptyCandidatesCount: numCandidates,
			ScaleDownCandidatesPoolRatio:     0.1,
			ScaleDownCandidatesPoolMinCount:  10,
		},
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		LogRecorder:          fakeLogRecorder,
		CloudProvider:        provider,
	}
	sd := NewScaleDown(&context)

	sd.UpdateUnneededNodes(nodes, nodes, pods, time.Now(), nil)
	assert.NotEmpty(t, sd.unneededNodes)
}

func TestDeleteNode(t *testing.T) {
	// common parameters
	nothingReturned := "Nothing returned"
	nodeDeleteFailedFunc :=
		func(string, string) error {
			return fmt.Errorf("won't remove node")
		}
	podNotFoundFunc :=
		func(action core.Action) (bool, runtime.Object, error) {
			return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
		}

	// scenarios
	testScenarios := []struct {
		name              string
		pods              []string
		drainSuccess      bool
		nodeDeleteSuccess bool
		expectedDeletion  bool
	}{
		{
			name:              "successful attempt to delete node with pods",
			pods:              []string{"p1", "p2"},
			drainSuccess:      true,
			nodeDeleteSuccess: true,
			expectedDeletion:  true,
		},
		/* Temporarily disabled as it takes several minutes due to hardcoded timeout.
		* TODO(aleksandra-malinowska): move MaxPodEvictionTime to AutoscalingContext.
		{
			name:              "failed on drain",
			pods:              []string{"p1", "p2"},
			drainSuccess:      false,
			nodeDeleteSuccess: true,
			expectedDeletion:  false,
		},
		*/
		{
			name:              "failed on node delete",
			pods:              []string{"p1", "p2"},
			drainSuccess:      true,
			nodeDeleteSuccess: false,
			expectedDeletion:  false,
		},
		{
			name:              "successful attempt to delete empty node",
			pods:              []string{},
			drainSuccess:      true,
			nodeDeleteSuccess: true,
			expectedDeletion:  true,
		},
		{
			name:              "failed attempt to delete empty node",
			pods:              []string{},
			drainSuccess:      true,
			nodeDeleteSuccess: false,
			expectedDeletion:  false,
		},
	}

	for _, scenario := range testScenarios {
		// run each scenario as an independent test
		t.Run(scenario.name, func(t *testing.T) {
			// set up test channels
			updatedNodes := make(chan string, 10)
			deletedNodes := make(chan string, 10)
			deletedPods := make(chan string, 10)

			// set up test data
			n1 := BuildTestNode("n1", 1000, 1000)
			SetNodeReadyState(n1, true, time.Time{})
			pods := make([]*apiv1.Pod, len(scenario.pods))
			for i, podName := range scenario.pods {
				pod := BuildTestPod(podName, 100, 0)
				pods[i] = pod
			}

			// set up fake provider
			deleteNodeHandler := nodeDeleteFailedFunc
			if scenario.nodeDeleteSuccess {
				deleteNodeHandler =
					func(nodeGroup string, node string) error {
						deletedNodes <- node
						return nil
					}
			}
			provider := testprovider.NewTestCloudProvider(nil, deleteNodeHandler)
			provider.AddNodeGroup("ng1", 1, 100, 100)
			provider.AddNode("ng1", n1)

			// set up fake client
			fakeClient := &fake.Clientset{}
			fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				return true, n1, nil
			})
			fakeClient.Fake.AddReactor("update", "nodes",
				func(action core.Action) (bool, runtime.Object, error) {
					update := action.(core.UpdateAction)
					obj := update.GetObject().(*apiv1.Node)
					taints := make([]string, 0, len(obj.Spec.Taints))
					for _, taint := range obj.Spec.Taints {
						taints = append(taints, taint.Key)
					}
					updatedNodes <- fmt.Sprintf("%s-%s", obj.Name, taints)
					return true, obj, nil
				})
			fakeClient.Fake.AddReactor("create", "pods",
				func(action core.Action) (bool, runtime.Object, error) {
					if !scenario.drainSuccess {
						return true, nil, fmt.Errorf("won't evict")
					}
					createAction := action.(core.CreateAction)
					if createAction == nil {
						return false, nil, nil
					}
					eviction := createAction.GetObject().(*policyv1.Eviction)
					if eviction == nil {
						return false, nil, nil
					}
					deletedPods <- eviction.Name
					return true, nil, nil
				})
			fakeClient.Fake.AddReactor("get", "pods", podNotFoundFunc)

			// set up fake recorders
			fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
			fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)

			// build context
			context := &AutoscalingContext{
				AutoscalingOptions:   AutoscalingOptions{},
				ClientSet:            fakeClient,
				Recorder:             fakeRecorder,
				LogRecorder:          fakeLogRecorder,
				CloudProvider:        provider,
				ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
			}

			// attempt delete
			err := deleteNode(context, n1, pods)

			// verify
			if scenario.expectedDeletion {
				assert.NoError(t, err)
				assert.Equal(t, n1.Name, getStringFromChan(deletedNodes))
			} else {
				assert.NotNil(t, err)
			}
			assert.Equal(t, nothingReturned, getStringFromChanImmediately(deletedNodes))

			taintedUpdate := fmt.Sprintf("%s-%s", n1.Name, []string{deletetaint.ToBeDeletedTaint})
			assert.Equal(t, taintedUpdate, getStringFromChan(updatedNodes))
			if !scenario.expectedDeletion {
				untaintedUpdate := fmt.Sprintf("%s-%s", n1.Name, []string{})
				assert.Equal(t, untaintedUpdate, getStringFromChan(updatedNodes))
			}
			assert.Equal(t, nothingReturned, getStringFromChanImmediately(updatedNodes))
		})
	}
}

func TestDrainNode(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
}

func TestDrainNodeWithRescheduled(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p2Rescheduled := BuildTestPod("p2", 300, 0)
	p2Rescheduled.Spec.NodeName = "n2"
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if getAction == nil {
			return false, nil, nil
		}
		if getAction.GetName() == "p2" {
			return true, p2Rescheduled, nil
		}
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
}

func TestDrainNodeWithRetries(t *testing.T) {
	deletedPods := make(chan string, 10)
	// Simulate pdb of size 1, by making them goroutine succeed sequentially
	// and fail/retry before they can proceed.
	ticket := make(chan bool, 1)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p3 := BuildTestPod("p3", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		select {
		case <-ticket:
			deletedPods <- eviction.Name
			return true, nil, nil
		default:
			select {
			case ticket <- true:
			default:
			}
			return true, nil, fmt.Errorf("Too many concurrent evictions")
		}
	})
	err := drainNode(n1, []*apiv1.Pod{p1, p2, p3}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	deleted = append(deleted, getStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
	assert.Equal(t, p3.Name, deleted[2])
}

func TestScaleDown(t *testing.T) {
	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	deletedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/extensions/v1beta1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p1 := BuildTestPod("p1", 100, 0)
	p1.OwnerReferences = GenerateOwnerReferences(job.Name, "Job", "extensions/v1beta1", "")

	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		deleteAction := action.(core.DeleteAction)
		deletedPods <- deleteAction.GetName()
		return true, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			MaxGracefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		LogRecorder:          fakeLogRecorder,
	}
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2},
		[]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, time.Now().Add(-5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, nil, time.Now())
	waitForDeleteToFinish(t, scaleDown)
	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNodeDeleteStarted, result)
	assert.Equal(t, n1.Name, getStringFromChan(deletedNodes))
	assert.Equal(t, n1.Name, getStringFromChan(updatedNodes))
}

func waitForDeleteToFinish(t *testing.T, sd *ScaleDown) {
	for start := time.Now(); time.Since(start) < 20*time.Second; time.Sleep(100 * time.Millisecond) {
		if !sd.nodeDeleteStatus.IsDeleteInProgress() {
			return
		}
	}
	t.Fatalf("Node delete not finished")
}

// this IGNORES duplicates
func assertEqualSet(t *testing.T, a []string, b []string) {
	assertSubset(t, a, b)
	assertSubset(t, b, a)
}

// this IGNORES duplicates
func assertSubset(t *testing.T, a []string, b []string) {
	for _, x := range a {
		found := false
		for _, y := range b {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Failed to find %s (from %s) in %v", x, a, b)
		}
	}
}

var defaultScaleDownOptions = AutoscalingOptions{
	ScaleDownUtilizationThreshold: 0.5,
	ScaleDownUnneededTime:         time.Minute,
	MaxGracefulTerminationSec:     60,
	MaxEmptyBulkDelete:            10,
	MinCoresTotal:                 0,
	MinMemoryTotal:                0,
	MaxCoresTotal:                 config.DefaultMaxClusterCores,
	MaxMemoryTotal:                config.DefaultMaxClusterMemory,
}

func TestScaleDownEmptyMultipleNodeGroups(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1_1", 1000, 1000, true, "ng1"},
			{"n1_2", 1000, 1000, true, "ng1"},
			{"n2_1", 1000, 1000, true, "ng2"},
			{"n2_2", 1000, 1000, true, "ng2"},
		},
		options:            defaultScaleDownOptions,
		expectedScaleDowns: []string{"n1_1", "n2_1"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptySingleNodeGroup(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 1000, 1000, true, "ng1"},
			{"n2", 1000, 1000, true, "ng1"},
		},
		options:            defaultScaleDownOptions,
		expectedScaleDowns: []string{"n1"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinCoresLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinCoresTotal = 2
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000, true, "ng1"},
			{"n2", 1000, 1000, true, "ng1"},
		},
		options:            options,
		expectedScaleDowns: []string{"n2"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinMemoryLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinMemoryTotal = 4000
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000 * MB, true, "ng1"},
			{"n2", 1000, 1000 * MB, true, "ng1"},
			{"n3", 1000, 1000 * MB, true, "ng1"},
			{"n4", 1000, 3000 * MB, true, "ng1"},
		},
		options:            options,
		expectedScaleDowns: []string{"n1", "n2"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGroupSizeLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000, true, "ng1"},
		},
		options:            options,
		expectedScaleDowns: []string{},
	}
	simpleScaleDownEmpty(t, config)
}
func simpleScaleDownEmpty(t *testing.T, config *scaleTestConfig) {
	updatedNodes := make(chan string, 10)
	deletedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	nodes := make([]*apiv1.Node, len(config.nodes))
	nodesMap := make(map[string]*apiv1.Node)
	groups := make(map[string][]*apiv1.Node)
	for i, n := range config.nodes {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		SetNodeReadyState(node, n.ready, time.Time{})
		nodesMap[n.name] = node
		nodes[i] = node
		groups[n.group] = append(groups[n.group], node)
	}

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if node, found := nodesMap[getAction.GetName()]; found {
			return true, node, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())

	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})

	for name, nodesInGroup := range groups {
		provider.AddNodeGroup(name, 1, 10, len(nodesInGroup))
		for _, n := range nodesInGroup {
			provider.AddNode(name, n)
		}
	}

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: config.options.MinCoresTotal, cloudprovider.ResourceNameMemory: config.options.MinMemoryTotal},
		map[string]int64{cloudprovider.ResourceNameCores: config.options.MaxCoresTotal, cloudprovider.ResourceNameMemory: config.options.MaxMemoryTotal})
	provider.SetResourceLimiter(resourceLimiter)

	assert.NotNil(t, provider)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions:   config.options,
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		LogRecorder:          fakeLogRecorder,
	}
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes(nodes,
		nodes, []*apiv1.Pod{}, time.Now().Add(-5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown(nodes, []*apiv1.Pod{}, nil, time.Now())
	waitForDeleteToFinish(t, scaleDown)
	// This helps to verify that TryToScaleDown doesn't attempt to remove anything
	// after delete in progress status is gone.
	close(deletedNodes)

	assert.NoError(t, err)
	var expectedScaleDownResult ScaleDownResult
	if len(config.expectedScaleDowns) > 0 {
		expectedScaleDownResult = ScaleDownNodeDeleted
	} else {
		expectedScaleDownResult = ScaleDownNoUnneeded
	}
	assert.Equal(t, expectedScaleDownResult, result)

	// Check the channel (and make sure there isn't more than there should be).
	// Report only up to 10 extra nodes found.
	deleted := make([]string, 0, len(config.expectedScaleDowns)+10)
	for i := 0; i < len(config.expectedScaleDowns)+10; i++ {
		d := getStringFromChanImmediately(deletedNodes)
		if d == "" { // a closed channel yields empty value
			break
		}
		deleted = append(deleted, d)
	}

	assertEqualSet(t, config.expectedScaleDowns, deleted)
}

func TestNoScaleDownUnready(t *testing.T) {
	fakeClient := &fake.Clientset{}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Minute))
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p2 := BuildTestPod("p2", 800, 0)
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.Fatalf("Unexpected deletion of %s", node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Hour,
			MaxGracefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		LogRecorder:          fakeLogRecorder,
	}

	// N1 is unready so it requires a bigger unneeded time.
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2},
		[]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, time.Now().Add(-5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, nil, time.Now())
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNoUnneeded, result)

	deletedNodes := make(chan string, 10)

	provider = testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Hour))
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	// N1 has been unready for 2 hours, ok to delete.
	context.CloudProvider = provider
	scaleDown = NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Node{n1, n2},
		[]*apiv1.Pod{p2}, time.Now().Add(-2*time.Hour), nil)
	result, err = scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p2}, nil, time.Now())
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNodeDeleteStarted, result)
	assert.Equal(t, n1.Name, getStringFromChan(deletedNodes))
}

func TestScaleDownNoMove(t *testing.T) {
	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/extensions/v1beta1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	// N2 is unready so no pods can be moved there.
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, false, time.Time{})

	p1 := BuildTestPod("p1", 100, 0)
	p1.OwnerReferences = GenerateOwnerReferences(job.Name, "Job", "extensions/v1beta1", "")

	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.FailNow()
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Hour,
			MaxGracefulTerminationSec:     60,
		},
		PredicateChecker:     simulator.NewTestPredicateChecker(),
		CloudProvider:        provider,
		ClientSet:            fakeClient,
		Recorder:             fakeRecorder,
		ClusterStateRegistry: clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder),
		LogRecorder:          fakeLogRecorder,
	}
	scaleDown := NewScaleDown(context)
	scaleDown.UpdateUnneededNodes([]*apiv1.Node{n1, n2}, []*apiv1.Node{n1, n2},
		[]*apiv1.Pod{p1, p2}, time.Now().Add(5*time.Minute), nil)
	result, err := scaleDown.TryToScaleDown([]*apiv1.Node{n1, n2}, []*apiv1.Pod{p1, p2}, nil, time.Now())
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, ScaleDownNoUnneeded, result)
}

func getStringFromChan(c chan string) string {
	select {
	case val := <-c:
		return val
	case <-time.After(10 * time.Second):
		return "Nothing returned"
	}
}

func getStringFromChanImmediately(c chan string) string {
	select {
	case val := <-c:
		return val
	default:
		return "Nothing returned"
	}
}

func TestCleanToBeDeleted(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n2.Spec.Taints = []apiv1.Taint{{Key: deletetaint.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("Wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		switch obj.Name {
		case n1.Name:
			n1 = obj
		case n2.Name:
			n2 = obj
		}
		return true, obj, nil
	})
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)

	cleanToBeDeleted([]*apiv1.Node{n1, n2}, fakeClient, fakeRecorder)

	assert.Equal(t, 0, len(n1.Spec.Taints))
	assert.Equal(t, 0, len(n2.Spec.Taints))
}

func TestCleanUpNodeAutoprovisionedGroups(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil,
		nil, func(id string) error {
			if id == "ng2" {
				return nil
			}
			return fmt.Errorf("Node group %s shouldn't be deleted", id)
		},
		nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddAutoprovisionedNodeGroup("ng2", 0, 10, 0, "mt1")
	provider.AddAutoprovisionedNodeGroup("ng3", 0, 10, 1, "mt1")
	provider.AddNode("ng3", n1)
	assert.NotNil(t, provider)

	fakeClient := &fake.Clientset{}
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient)
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", fakeRecorder, false)
	assert.NoError(t, cleanUpNodeAutoprovisionedGroups(provider, fakeLogRecorder))
}

func TestCalculateCoresAndMemoryTotal(t *testing.T) {
	nodeConfigs := []nodeConfig{
		{"n1", 2000, 7500 * MB, true, "ng1"},
		{"n2", 2000, 7500 * MB, true, "ng1"},
		{"n3", 2000, 7500 * MB, true, "ng1"},
		{"n4", 12000, 8000 * MB, true, "ng1"},
		{"n5", 16000, 7500 * MB, true, "ng1"},
		{"n6", 8000, 6000 * MB, true, "ng1"},
		{"n7", 6000, 16000 * MB, true, "ng1"},
	}
	nodes := make([]*apiv1.Node, len(nodeConfigs))
	for i, n := range nodeConfigs {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		SetNodeReadyState(node, n.ready, time.Now())
		nodes[i] = node
	}

	nodes[6].Spec.Taints = []apiv1.Taint{
		{
			Key:    deletetaint.ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}

	coresTotal, memoryTotal := calculateCoresAndMemoryTotal(nodes, time.Now())

	assert.Equal(t, int64(42), coresTotal)
	assert.Equal(t, int64(44000), memoryTotal)
}

func TestFilterOutMasters(t *testing.T) {
	nodeConfigs := []nodeConfig{
		{"n1", 2000, 4000, false, "ng1"},
		{"n2", 2000, 4000, true, "ng2"},
		{"n3", 2000, 8000, true, ""}, // real master
		{"n4", 1000, 2000, true, "ng3"},
		{"n5", 1000, 2000, true, "ng3"},
		{"n6", 2000, 8000, true, ""}, // same machine type, no node group, no api server
		{"n7", 2000, 8000, true, ""}, // real master
	}
	nodes := make([]*apiv1.Node, len(nodeConfigs))
	for i, n := range nodeConfigs {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		SetNodeReadyState(node, n.ready, time.Now())
		nodes[i] = node
	}

	BuildTestPodWithExtra := func(name, namespace, node string, labels map[string]string) *apiv1.Pod {
		pod := BuildTestPod(name, 100, 200)
		pod.Spec.NodeName = node
		pod.Namespace = namespace
		pod.Labels = labels
		return pod
	}

	pods := []*apiv1.Pod{
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n2", map[string]string{}),                                          // without label
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "fake-kube-system", "n6", map[string]string{"component": "kube-apiserver"}),        // wrong namespace
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n3", map[string]string{"component": "kube-apiserver"}),             // real api server
		BuildTestPodWithExtra("hidden-name", "kube-system", "n7", map[string]string{"component": "kube-apiserver"}),                                  // also a real api server
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n1", map[string]string{"component": "kube-apiserver-dev"}),         // wrong label
		BuildTestPodWithExtra("custom-deployment", "custom", "n5", map[string]string{"component": "custom-component", "custom-key": "custom-value"}), // unrelated pod
	}

	withoutMasters := filterOutMasters(nodes, pods)

	withoutMastersNames := make([]string, len(withoutMasters))
	for i, n := range withoutMasters {
		withoutMastersNames[i] = n.Name
	}
	assertEqualSet(t, []string{"n1", "n2", "n4", "n5", "n6"}, withoutMastersNames)
}
