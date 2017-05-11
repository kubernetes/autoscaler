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
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	core "k8s.io/client-go/testing"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
)

func TestNewAutoscalerStatic(t *testing.T) {
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
	kubeEventRecorder := kube_util.CreateEventRecorder(fakeClient)
	opts := AutoscalerOptions{
		ConfigFetcherOptions: dynamic.ConfigFetcherOptions{
			ConfigMapName: "",
		},
	}
	predicateChecker := simulator.NewTestPredicateChecker()
	listerRegistry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil)
	a := NewAutoscaler(opts, predicateChecker, fakeClient, kubeEventRecorder, listerRegistry)
	assert.IsType(t, &StaticAutoscaler{}, a)
}

func TestNewAutoscalerDynamic(t *testing.T) {
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
	kubeEventRecorder := kube_util.CreateEventRecorder(fakeClient)
	opts := AutoscalerOptions{
		ConfigFetcherOptions: dynamic.ConfigFetcherOptions{
			ConfigMapName: "testconfigmap",
		},
	}
	predicateChecker := simulator.NewTestPredicateChecker()
	listerRegistry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil)
	a := NewAutoscaler(opts, predicateChecker, fakeClient, kubeEventRecorder, listerRegistry)
	assert.IsType(t, &DynamicAutoscaler{}, a)
}
