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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	hpav1 "k8s.io/api/autoscaling/v1"
	hpa "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	balancerapi "k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type podListerMock struct {
	pods []*v1.Pod
}

type podNamespaceListerMock struct {
	pods []*v1.Pod
}

func (p *podListerMock) List(selector labels.Selector) (ret []*v1.Pod, err error) {
	result := make([]*v1.Pod, 0)
	for _, pod := range p.pods {
		if selector.Matches(labels.Set(pod.Labels)) {
			result = append(result, pod)
		}
	}
	return result, nil
}

func (p *podListerMock) Pods(namespace string) corelisters.PodNamespaceLister {
	filtered := make([]*v1.Pod, 0)
	for _, pod := range p.pods {
		if pod.Namespace == namespace {
			filtered = append(filtered, pod)
		}
	}
	return &podNamespaceListerMock{
		pods: filtered,
	}
}

func (p *podNamespaceListerMock) List(selector labels.Selector) (ret []*v1.Pod, err error) {
	result := make([]*v1.Pod, 0)
	for _, pod := range p.pods {
		if selector.Matches(labels.Set(pod.Labels)) {
			result = append(result, pod)
		}
	}
	return result, nil
}

func (p *podNamespaceListerMock) Get(name string) (*v1.Pod, error) {
	for _, pod := range p.pods {
		if pod.Name == name {
			return pod, nil
		}
	}
	return nil, errors.New("Not found")
}

func newPod(namespace, name string, phase v1.PodPhase, createTime time.Time, podLabels string) *v1.Pod {
	parsedLabels, err := labels.ConvertSelectorToLabelsMap(podLabels)
	if err != nil {
		panic(err)
	}
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.NewTime(createTime),
			Labels:            parsedLabels,
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}

func newTarget(name string) balancerapi.BalancerTarget {
	return balancerapi.BalancerTarget{
		Name: name,
		ScaleTargetRef: hpa.CrossVersionObjectReference{
			Name:       name,
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
	}
}

func newScale(name string, replicas int32) *hpav1.Scale {
	return &hpav1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: hpav1.ScaleSpec{
			Replicas: replicas,
		},
		Status: hpav1.ScaleStatus{
			Replicas: replicas,
			Selector: "run=" + name,
		},
	}
}

func newBalancer(replicas int32) *balancerapi.Balancer {
	return &balancerapi.Balancer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "balancer",
		},
		Spec: balancerapi.BalancerSpec{
			Targets: []balancerapi.BalancerTarget{
				newTarget("a"),
				newTarget("b"),
			},
			Replicas: replicas,
			Selector: metav1.LabelSelector{MatchLabels: map[string]string{"service": "nginx"}},
			Policy: balancerapi.BalancerPolicy{
				PolicyName: balancerapi.ProportionalPolicyName,
				Proportions: &balancerapi.ProportionalPolicy{
					TargetProportions: map[string]int32{"a": 30, "b": 70},
				},
				Fallback: &balancerapi.FallbackPolicy{
					StartupTimeoutSeconds: 60,
				},
			},
		},
	}
}

func TestProcessBalancer(t *testing.T) {

	tests := []struct {
		name               string
		pods               []*v1.Pod
		balancer           *balancerapi.Balancer
		scales             []*hpav1.Scale
		noChange           bool
		expected           map[string]int32
		balancerPhaseError BalancerPhase
	}{
		{
			name:     "No pods, 10 replicas, 30/70",
			pods:     []*v1.Pod{},
			balancer: newBalancer(10),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 0),
			},
			expected: map[string]int32{"a": 3, "b": 7},
		},
		{
			name: "With pods, 10 replicas, 30/70",
			pods: []*v1.Pod{
				newPod("default", "a1", v1.PodRunning, time.Now(), "run=a,service=nginx"),
				newPod("default", "a2", v1.PodRunning, time.Now(), "run=a,service=nginx"),
				newPod("default", "a3", v1.PodRunning, time.Now(), "run=a,service=nginx"),
			},
			balancer: newBalancer(10),
			scales: []*hpav1.Scale{
				newScale("a", 3),
				newScale("b", 0),
			},
			expected: map[string]int32{"a": 3, "b": 7},
		},
		{
			name: "With pods, wrong selector",
			pods: []*v1.Pod{
				newPod("default", "a1", v1.PodRunning, time.Now(), "run=a,service=nginx"),
			},
			balancer: func() *balancerapi.Balancer {
				b := newBalancer(10)
				b.Spec.Selector.MatchLabels["xx"] = "yy"
				return b
			}(),
			scales: []*hpav1.Scale{
				newScale("a", 1),
				newScale("b", 0),
			},
			balancerPhaseError: PodLabelsChecking,
		},
		{
			name: "Without pods, priority",
			pods: []*v1.Pod{},
			balancer: func() *balancerapi.Balancer {
				b := newBalancer(10)
				b.Spec.Policy.PolicyName = balancerapi.PriorityPolicyName
				b.Spec.Policy.Proportions = nil
				b.Spec.Policy.Priorities = &balancerapi.PriorityPolicy{
					TargetOrder: []string{"a", "b"},
				}
				return b
			}(),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 0),
			},
			expected: map[string]int32{"a": 10, "b": 0},
		},
		{
			name:     "No pods, 0 replicas, 30/70",
			pods:     []*v1.Pod{},
			balancer: newBalancer(0),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 0),
			},
			noChange: true,
		},
		{
			name: "With pods, 1 replica, 30/70, with fallback",
			pods: []*v1.Pod{
				newPod("default", "b1", v1.PodPending, time.Now().Add(-time.Hour), "run=b,service=nginx"),
			},
			balancer: newBalancer(1),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 1),
			},
			expected: map[string]int32{"a": 1, "b": 1},
		},
		{
			name: "With pods, 1 replica, 30/70, with young pending",
			pods: []*v1.Pod{
				newPod("default", "b1", v1.PodPending, time.Now().Add(-time.Second*20), "run=b,service=nginx"),
			},
			balancer: newBalancer(1),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 0),
			},
			expected: map[string]int32{"a": 0, "b": 1},
		},
		{
			name: "With pods, 1 replica, 30/70, with medium pending (fallback)",
			pods: []*v1.Pod{
				newPod("default", "b1", v1.PodPending, time.Now().Add(-time.Second*61), "run=b,service=nginx"),
			},
			balancer: newBalancer(1),
			scales: []*hpav1.Scale{
				newScale("a", 0),
				newScale("b", 0),
			},
			expected: map[string]int32{"a": 1, "b": 1},
		},
		{
			name:               "Wrong targets",
			pods:               []*v1.Pod{},
			balancer:           newBalancer(0),
			scales:             []*hpav1.Scale{},
			balancerPhaseError: ScaleSubresourcePolling,
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			scaleClient := scaleClientMock{
				scales: map[string]*hpav1.Scale{},
			}
			for _, s := range tc.scales {
				scaleClient.scales[scalesMockKey(s.Namespace, newTarget(s.Name).ScaleTargetRef)] = s
			}

			podLister := podListerMock{
				pods: tc.pods,
			}

			core := newCoreForTests(&scaleClient, &podLister)
			statusInfo, errorsInfo := core.ProcessBalancer(tc.balancer, time.Now())

			if tc.balancerPhaseError != "" {
				assert.True(t, statusInfo == nil || statusInfo.updated == false)
				assert.Equal(t, tc.balancerPhaseError, errorsInfo.phase)
			}
			if tc.balancerPhaseError == "" {
				assert.Equal(t, !tc.noChange, statusInfo.updated)
			}
			if tc.expected != nil {
				for k, v := range tc.expected {
					key := scalesMockKey("default", newTarget(k).ScaleTargetRef)
					replicas := scaleClient.scales[key].Spec.Replicas
					assert.Equal(t, v, replicas, "replica count for "+key)
				}
			}
		})
	}
}
