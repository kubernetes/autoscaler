/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	acontext "k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

func TestDaemonSetEvictionForEmptyNodes(t *testing.T) {
	testScenarios := []struct {
		name                  string
		dsPods                []string
		evictionTimeoutExceed bool
		dsEvictionTimeout     time.Duration
		evictionSuccess       bool
		err                   error
		evictByDefault        bool
		extraAnnotationValue  map[string]string
		expectNotEvicted      map[string]struct{}
		podPriorities         []int32
	}{
		{
			name:              "Successful attempt to evict DaemonSet pods",
			dsPods:            []string{"d1", "d2"},
			dsEvictionTimeout: 5000 * time.Millisecond,
			evictionSuccess:   true,
			evictByDefault:    true,
		},
		{
			name:                 "Evict single pod due to annotation",
			dsPods:               []string{"d1", "d2"},
			dsEvictionTimeout:    5000 * time.Millisecond,
			evictionSuccess:      true,
			extraAnnotationValue: map[string]string{"d1": "true"},
			expectNotEvicted:     map[string]struct{}{"d2": {}},
		},
		{
			name:                 "Don't evict single pod due to annotation",
			dsPods:               []string{"d1", "d2"},
			dsEvictionTimeout:    5000 * time.Millisecond,
			evictionSuccess:      true,
			evictByDefault:       true,
			extraAnnotationValue: map[string]string{"d1": "false"},
			expectNotEvicted:     map[string]struct{}{"d1": {}},
		},
	}

	for _, scenario := range testScenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			options := config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUtilizationThreshold: 0.5,
					ScaleDownUnneededTime:         time.Minute,
				},
				MaxGracefulTerminationSec:      1,
				DaemonSetEvictionForEmptyNodes: scenario.evictByDefault,
				MaxPodEvictionTime:             scenario.dsEvictionTimeout,
			}
			deletedPods := make(chan string, len(scenario.dsPods)+2)
			waitBetweenRetries := 10 * time.Millisecond

			fakeClient := &fake.Clientset{}
			n1 := BuildTestNode("n1", 1000, 1000)
			SetNodeReadyState(n1, true, time.Time{})
			dsPods := make([]*apiv1.Pod, len(scenario.dsPods))
			for i, dsName := range scenario.dsPods {
				ds := BuildTestPod(dsName, 100, 0, WithDSController())
				ds.Spec.NodeName = "n1"
				if v, ok := scenario.extraAnnotationValue[dsName]; ok {
					ds.Annotations[daemonset.EnableDsEvictionKey] = v
				}
				dsPods[i] = ds
			}

			fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
				createAction := action.(core.CreateAction)
				if createAction == nil {
					return false, nil, nil
				}
				eviction := createAction.GetObject().(*policyv1beta1.Eviction)
				if eviction == nil {
					return false, nil, nil
				}
				if scenario.evictionTimeoutExceed {
					time.Sleep(10 * scenario.dsEvictionTimeout)
				}
				if !scenario.evictionSuccess {
					return true, nil, fmt.Errorf("fail to evict the pod")
				}
				deletedPods <- eviction.Name
				return true, nil, nil
			})
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroup("ng1", 1, 10, 1)
			provider.AddNode("ng1", n1)
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil)

			context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
			assert.NoError(t, err)

			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, []*apiv1.Node{n1}, dsPods)

			drainConfig := SingleRuleDrainConfig(context.MaxGracefulTerminationSec)
			evictor := Evictor{
				EvictionRetryTime:                waitBetweenRetries,
				shutdownGracePeriodByPodPriority: drainConfig,
			}
			nodeInfo, err := context.ClusterSnapshot.GetNodeInfo(n1.Name)
			assert.NoError(t, err)
			_, err = evictor.EvictDaemonSetPods(&context, nodeInfo)
			if scenario.err != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), scenario.err.Error())
				return
			}
			assert.Nil(t, err)
			var expectEvicted []string
			for _, p := range scenario.dsPods {
				if _, found := scenario.expectNotEvicted[p]; found {
					continue
				}
				expectEvicted = append(expectEvicted, p)
			}
			deleted := make([]string, len(expectEvicted))
			for i := 0; i < len(expectEvicted); i++ {
				deleted[i] = utils.GetStringFromChan(deletedPods)
			}
			assert.ElementsMatch(t, deleted, expectEvicted)
		})
	}
}

func TestDrainNodeWithPods(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 300, 0, WithNodeName(n1.Name))
	d1 := BuildTestPod("d1", 150, 0, WithNodeName(n1.Name), WithDSController())

	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec:         20,
		MaxPodEvictionTime:                5 * time.Second,
		DaemonSetEvictionForOccupiedNodes: true,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              DefaultPodEvictionHeadroom,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, d1})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	_, err = evictor.DrainNode(&ctx, nodeInfo)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))

	sort.Strings(deleted)
	assert.Equal(t, d1.Name, deleted[0])
	assert.Equal(t, p1.Name, deleted[1])
	assert.Equal(t, p2.Name, deleted[2])
}

func TestDrainNodeWithPodsWithRescheduled(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 300, 0, WithNodeName(n1.Name))
	p2Rescheduled := BuildTestPod("p2", 300, 0)
	p2Rescheduled.Spec.NodeName = "n2"

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
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec: 20,
		MaxPodEvictionTime:        5 * time.Second,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              DefaultPodEvictionHeadroom,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	_, err = evictor.DrainNode(&ctx, nodeInfo)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
}

func TestDrainNodeWithPodsWithRetries(t *testing.T) {
	deletedPods := make(chan string, 10)
	// Simulate pdb of size 1 by making the 'eviction' goroutine:
	// - read from (at first empty) channel
	// - if it's empty, fail and write to it, then retry
	// - succeed on successful read.
	ticket := make(chan bool, 1)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 300, 0, WithNodeName(n1.Name))
	p3 := BuildTestPod("p3", 300, 0, WithNodeName(n1.Name))
	d1 := BuildTestPod("d1", 150, 0, WithDSController(), WithNodeName(n1.Name))

	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
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
			return true, nil, fmt.Errorf("too many concurrent evictions")
		}
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec:         20,
		MaxPodEvictionTime:                5 * time.Second,
		DaemonSetEvictionForOccupiedNodes: true,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              DefaultPodEvictionHeadroom,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, p3, d1})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	_, err = evictor.DrainNode(&ctx, nodeInfo)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, d1.Name, deleted[0])
	assert.Equal(t, p1.Name, deleted[1])
	assert.Equal(t, p2.Name, deleted[2])
	assert.Equal(t, p3.Name, deleted[3])
}

func TestDrainNodeWithPodsDaemonSetEvictionFailure(t *testing.T) {
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 300, 0, WithNodeName(n1.Name))
	d1 := BuildTestPod("d1", 150, 0, WithDSController(), WithNodeName(n1.Name))
	d2 := BuildTestPod("d2", 250, 0, WithDSController(), WithNodeName(n1.Name))

	e1 := fmt.Errorf("eviction_error: d1")
	e2 := fmt.Errorf("eviction_error: d2")

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		if eviction.Name == "d1" {
			return true, nil, e1
		}
		if eviction.Name == "d2" {
			return true, nil, e2
		}
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec: 20,
		MaxPodEvictionTime:        0 * time.Second,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              DefaultPodEvictionHeadroom,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, d1, d2})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	evictionResults, err := evictor.DrainNode(&ctx, nodeInfo)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(evictionResults))
	assert.Equal(t, p1, evictionResults["p1"].Pod)
	assert.Equal(t, p2, evictionResults["p2"].Pod)
	assert.NoError(t, evictionResults["p1"].Err)
	assert.NoError(t, evictionResults["p2"].Err)
	assert.False(t, evictionResults["p1"].TimedOut)
	assert.False(t, evictionResults["p2"].TimedOut)
	assert.True(t, evictionResults["p1"].WasEvictionSuccessful())
	assert.True(t, evictionResults["p2"].WasEvictionSuccessful())
}

func TestDrainNodeWithPodsEvictionFailure(t *testing.T) {
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 100, 0, WithNodeName(n1.Name))
	p3 := BuildTestPod("p3", 100, 0, WithNodeName(n1.Name))
	p4 := BuildTestPod("p4", 100, 0, WithNodeName(n1.Name))
	e2 := fmt.Errorf("eviction_error: p2")
	e4 := fmt.Errorf("eviction_error: p4")
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}

		if eviction.Name == "p2" {
			return true, nil, e2
		}
		if eviction.Name == "p4" {
			return true, nil, e4
		}
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec: 20,
		MaxPodEvictionTime:        0 * time.Second,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)
	r := evRegister{}
	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              DefaultPodEvictionHeadroom,
		evictionRegister:                 &r,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, p3, p4})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	evictionResults, err := evictor.DrainNode(&ctx, nodeInfo)
	assert.Error(t, err)
	assert.Equal(t, 4, len(evictionResults))
	assert.Equal(t, *p1, *evictionResults["p1"].Pod)
	assert.Equal(t, *p2, *evictionResults["p2"].Pod)
	assert.Equal(t, *p3, *evictionResults["p3"].Pod)
	assert.Equal(t, *p4, *evictionResults["p4"].Pod)
	assert.NoError(t, evictionResults["p1"].Err)
	assert.Contains(t, evictionResults["p2"].Err.Error(), e2.Error())
	assert.NoError(t, evictionResults["p3"].Err)
	assert.Contains(t, evictionResults["p4"].Err.Error(), e4.Error())
	assert.False(t, evictionResults["p1"].TimedOut)
	assert.True(t, evictionResults["p2"].TimedOut)
	assert.False(t, evictionResults["p3"].TimedOut)
	assert.True(t, evictionResults["p4"].TimedOut)
	assert.True(t, evictionResults["p1"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p2"].WasEvictionSuccessful())
	assert.True(t, evictionResults["p3"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p4"].WasEvictionSuccessful())
	assert.Contains(t, r.pods, p1, p3)
}

func TestDrainWithPodsNodeDisappearanceFailure(t *testing.T) {
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 100, 0, WithNodeName(n1.Name))
	p3 := BuildTestPod("p3", 100, 0, WithNodeName(n1.Name))
	p4 := BuildTestPod("p4", 100, 0, WithNodeName(n1.Name))
	e2 := fmt.Errorf("disappearance_error: p2")
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if getAction == nil {
			return false, nil, nil
		}
		if getAction.GetName() == "p2" {
			return true, nil, e2
		}
		if getAction.GetName() == "p4" {
			return true, nil, nil
		}
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec: 0,
		MaxPodEvictionTime:        0 * time.Second,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	evictor := Evictor{
		EvictionRetryTime:                0,
		PodEvictionHeadroom:              0,
		shutdownGracePeriodByPodPriority: legacyFlagDrainConfig,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, p3, p4})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	evictionResults, err := evictor.DrainNode(&ctx, nodeInfo)
	assert.Error(t, err)
	assert.Equal(t, 4, len(evictionResults))
	assert.Equal(t, *p1, *evictionResults["p1"].Pod)
	assert.Equal(t, *p2, *evictionResults["p2"].Pod)
	assert.Equal(t, *p3, *evictionResults["p3"].Pod)
	assert.Equal(t, *p4, *evictionResults["p4"].Pod)
	assert.NoError(t, evictionResults["p1"].Err)
	assert.Contains(t, evictionResults["p2"].Err.Error(), e2.Error())
	assert.NoError(t, evictionResults["p3"].Err)
	assert.NoError(t, evictionResults["p4"].Err)
	assert.False(t, evictionResults["p1"].TimedOut)
	assert.True(t, evictionResults["p2"].TimedOut)
	assert.False(t, evictionResults["p3"].TimedOut)
	assert.True(t, evictionResults["p4"].TimedOut)
	assert.True(t, evictionResults["p1"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p2"].WasEvictionSuccessful())
	assert.True(t, evictionResults["p3"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p4"].WasEvictionSuccessful())
}

func TestPodsToEvict(t *testing.T) {
	for tn, tc := range map[string]struct {
		pods               []*apiv1.Pod
		nodeNameOverwrite  string
		dsEvictionDisabled bool
		wantDsPods         []*apiv1.Pod
		wantNonDsPods      []*apiv1.Pod
	}{
		"no pods": {
			pods:          []*apiv1.Pod{},
			wantDsPods:    []*apiv1.Pod{},
			wantNonDsPods: []*apiv1.Pod{},
		},
		"mirror pods are never returned": {
			pods:          []*apiv1.Pod{mirrorPod("pod-1"), mirrorPod("pod-2")},
			wantDsPods:    []*apiv1.Pod{},
			wantNonDsPods: []*apiv1.Pod{},
		},
		"non-DS pods are correctly returned": {
			pods:          []*apiv1.Pod{regularPod("pod-1"), regularPod("pod-2")},
			wantDsPods:    []*apiv1.Pod{},
			wantNonDsPods: []*apiv1.Pod{regularPod("pod-1"), regularPod("pod-2")},
		},
		"DS pods are correctly returned when DS eviction is enabled": {
			pods:          []*apiv1.Pod{dsPod("pod-1", false), dsPod("pod-2", false)},
			wantDsPods:    []*apiv1.Pod{dsPod("pod-1", false), dsPod("pod-2", false)},
			wantNonDsPods: []*apiv1.Pod{},
		},
		"DS pods are not returned when DS eviction is disabled and the pods are not marked as evictable": {
			dsEvictionDisabled: true,
			pods:               []*apiv1.Pod{dsPod("pod-1", false), dsPod("pod-2", false)},
			wantDsPods:         []*apiv1.Pod{},
			wantNonDsPods:      []*apiv1.Pod{},
		},
		"DS pods are correctly returned when DS eviction is disabled, but the pods are marked as evictable": {
			dsEvictionDisabled: true,
			pods:               []*apiv1.Pod{dsPod("pod-1", true), dsPod("pod-2", false), dsPod("pod-3", true)},
			wantDsPods:         []*apiv1.Pod{dsPod("pod-1", true), dsPod("pod-3", true)},
			wantNonDsPods:      []*apiv1.Pod{},
		},
		"all pod kinds are correctly handled together": {
			pods: []*apiv1.Pod{
				dsPod("ds-pod-1", false), dsPod("ds-pod-2", false),
				regularPod("regular-pod-1"), regularPod("regular-pod-2"),
				mirrorPod("mirror-pod-1"), mirrorPod("mirror-pod-2"),
			},
			wantDsPods:    []*apiv1.Pod{dsPod("ds-pod-1", false), dsPod("ds-pod-2", false)},
			wantNonDsPods: []*apiv1.Pod{regularPod("regular-pod-1"), regularPod("regular-pod-2")},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			snapshot := clustersnapshot.NewBasicClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true)
			node := BuildTestNode("test-node", 1000, 1000)
			err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, tc.pods...))
			if err != nil {
				t.Errorf("AddNodeWithPods unexpected error: %v", err)
			}
			ctx := &acontext.AutoscalingContext{
				ClusterSnapshot: snapshot,
				AutoscalingOptions: config.AutoscalingOptions{
					DaemonSetEvictionForOccupiedNodes: !tc.dsEvictionDisabled,
				},
			}
			nodeName := "test-node"
			if tc.nodeNameOverwrite != "" {
				nodeName = tc.nodeNameOverwrite
			}
			nodeInfo, err := snapshot.GetNodeInfo(nodeName)
			if err != nil {
				t.Fatalf("GetNodeInfo() unexpected error: %v", err)
			}
			gotDsPods, gotNonDsPods := podsToEvict(nodeInfo, ctx.DaemonSetEvictionForOccupiedNodes)
			if diff := cmp.Diff(tc.wantDsPods, gotDsPods, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("podsToEvict dsPods diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantNonDsPods, gotNonDsPods, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("podsToEvict nonDsPods diff (-want +got):\n%s", diff)
			}
		})
	}
}

func regularPod(name string) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func mirrorPod(name string) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				types.ConfigMirrorAnnotationKey: "some-key",
			},
		},
	}
}

func dsPod(name string, evictable bool) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: GenerateOwnerReferences(name+"-ds", "DaemonSet", "apps/v1", "some-uid"),
		},
	}
	if evictable {
		pod.Annotations = map[string]string{daemonset.EnableDsEvictionKey: "true"}
	}
	return pod
}

type evRegister struct {
	sync.Mutex
	pods []*apiv1.Pod
}

func (eR *evRegister) RegisterEviction(pod *apiv1.Pod) {
	eR.Lock()
	defer eR.Unlock()
	eR.pods = append(eR.pods, pod)
}
