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

package controller

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	balancerapi "k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	fakebalancer "k8s.io/autoscaler/balancer/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/balancer/pkg/client/informers/externalversions"
	coretesting "k8s.io/client-go/testing"
	"k8s.io/klog/v2"

	"k8s.io/client-go/kubernetes/fake"
)

type testContext struct {
	statusInfo      *BalancerStatusInfo
	balancerError   *BalancerError
	input           []balancerapi.Balancer
	core            *fakeCore
	controller      *Controller
	stop            chan struct{}
	balancerUpdates chan balancerapi.Balancer
	events          chan v1.Event
}

type fakeCore struct {
	sync.Mutex
	received    *balancerapi.Balancer
	calls       int32
	testContext *testContext
}

func (f *fakeCore) ProcessBalancer(balancer *balancerapi.Balancer, now time.Time) (*BalancerStatusInfo, *BalancerError) {
	defer f.Unlock()
	f.Lock()
	f.received = balancer
	f.calls++
	return f.testContext.statusInfo, f.testContext.balancerError
}

func (f *fakeCore) IsSynced() bool {
	return true
}

func prepareTest(balancer *balancerapi.Balancer, info *BalancerStatusInfo, err *BalancerError, updateOk bool) *testContext {
	tc := &testContext{
		input:           []balancerapi.Balancer{*balancer},
		statusInfo:      info,
		balancerError:   err,
		stop:            make(chan struct{}),
		balancerUpdates: make(chan balancerapi.Balancer, 1000),
		events:          make(chan v1.Event, 1000),
	}

	balancerclient := &fakebalancer.Clientset{}
	balancerclient.AddReactor("list", "balancers",
		func(action coretesting.Action) (handled bool, ret runtime.Object, err error) {
			obj := &balancerapi.BalancerList{}
			obj.Items = tc.input
			klog.Infof("List balancers")
			return true, obj, nil
		})

	balancerclient.AddReactor("update", "balancers",
		func(action coretesting.Action) (handled bool, ret runtime.Object, err error) {
			obj := action.(coretesting.UpdateAction).GetObject().(*balancerapi.Balancer)
			if updateOk {
				tc.balancerUpdates <- *obj.DeepCopy()
				klog.Infof("Updated balancers")
				return true, obj, nil
			}
			return true, obj, fmt.Errorf("Access denied")
		})

	balancerInformerFactory := externalversions.NewSharedInformerFactory(balancerclient, 0)
	informer := balancerInformerFactory.Balancer().V1alpha1().Balancers()

	fakeEvents := fake.Clientset{}
	fakeEvents.AddReactor("create", "events",
		func(action coretesting.Action) (handled bool, ret runtime.Object, err error) {
			obj := action.(coretesting.CreateAction).GetObject().(*v1.Event)
			tc.events <- *obj.DeepCopy()
			klog.Info("Published event")
			return true, obj, nil
		})

	fc := &fakeCore{
		testContext: tc,
	}

	tc.controller = NewController(balancerclient, informer, fakeEvents.CoreV1().Events(""), fc, time.Second)
	balancerInformerFactory.Start(tc.stop)

	return tc
}

func poolBalancer(ch chan balancerapi.Balancer, maxDuration time.Duration) *balancerapi.Balancer {
	select {
	case balancer := <-ch:
		return &balancer
	case <-time.After(maxDuration):
		return nil
	}
}

func poolEvents(ch chan v1.Event, maxDuration time.Duration) *v1.Event {
	select {
	case event := <-ch:
		return &event
	case <-time.After(maxDuration):
		return nil
	}
}

func TestController(t *testing.T) {
	testCases := []struct {
		name string

		// input
		balancer *balancerapi.Balancer
		info     *BalancerStatusInfo
		err      *BalancerError

		// expectations:
		statusReplicas  int32
		conditionType   string
		conditionStatus metav1.ConditionStatus
		expectedEvent   string
		updateFailed    bool
	}{
		{
			name:            "all fine",
			balancer:        newBalancer(5),
			info:            &BalancerStatusInfo{replicasObserved: 10},
			err:             nil,
			statusReplicas:  10,
			conditionType:   balancerapi.BalancerConditionRunning,
			conditionStatus: metav1.ConditionTrue,
		},
		{
			name:            "early error",
			balancer:        newBalancer(5),
			info:            nil,
			err:             newBalancerError(ScaleSubresourcePolling, errors.New("Booom")),
			conditionType:   balancerapi.BalancerConditionRunning,
			conditionStatus: metav1.ConditionFalse,
			expectedEvent:   "UnableToBalance",
		},
		{
			name: "error overwrite",
			balancer: func() *balancerapi.Balancer {
				b := newBalancer(5)
				setConditionsBasedOnError(b, nil, time.Now())
				return b
			}(),
			info:            nil,
			err:             newBalancerError(ScaleSubresourcePolling, errors.New("Booom")),
			conditionType:   balancerapi.BalancerConditionRunning,
			conditionStatus: metav1.ConditionFalse,
			expectedEvent:   "UnableToBalance",
		},
		{
			name:            "late error",
			balancer:        newBalancer(5),
			info:            &BalancerStatusInfo{replicasObserved: 7},
			err:             newBalancerError(ReplicaCountSetting, errors.New("Booom")),
			statusReplicas:  7,
			conditionType:   balancerapi.BalancerConditionRunning,
			conditionStatus: metav1.ConditionFalse,
			expectedEvent:   "UnableToBalance",
		},
		{
			name:          "balancer update error",
			balancer:      newBalancer(5),
			info:          &BalancerStatusInfo{replicasObserved: 7},
			err:           nil,
			expectedEvent: "StatusNotUpdated",
			updateFailed:  true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i+1, tc.name), func(t *testing.T) {
			testContext := prepareTest(
				tc.balancer,
				tc.info,
				tc.err,
				!tc.updateFailed,
			)
			go testContext.controller.Run(1, testContext.stop)

			if !tc.updateFailed {
				// On a very overloaded test machine this may be flaky.
				// Each test case is expected to end up in < 0.1s. The machine needs
				// to be >> 100x slower than usual in order for that to fail.
				balancer := poolBalancer(testContext.balancerUpdates, time.Second*10)
				close(testContext.stop)

				assert.NotNil(t, balancer)
				assert.Equal(t, tc.statusReplicas, balancer.Status.Replicas)
				found := false
				for _, c := range balancer.Status.Conditions {
					if c.Type == tc.conditionType {
						found = true
						assert.Equal(t, tc.conditionStatus, c.Status)
					}
				}
				assert.True(t, found)
			}

			if len(tc.expectedEvent) == 0 {
				// This is not guaranteed to always catch errors. We expect no events to
				// be published, we wait a bit here, but after 0.1s after the balancer is
				// processed we expect that no unexpected events will arrive.
				assert.Nil(t, poolEvents(testContext.events, time.Millisecond*100))
			} else {
				// On a very overloaded test machine this may be flaky.
				// Each test case is expected to end up in < 0.1s. The machine needs
				// to be >> 100x slower than usual in order for that to fail.
				event := poolEvents(testContext.events, time.Second*10)
				assert.Equal(t, tc.expectedEvent, event.Reason)
			}
		})
	}
}
