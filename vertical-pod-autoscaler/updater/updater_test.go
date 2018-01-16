/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/apimock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/eviction"
	"k8s.io/kubernetes/pkg/api/testapi"
)

func TestRunOnce(t *testing.T) {
	replicas := int32(5)
	livePods := 5
	labels := map[string]string{"app": "testingApp"}
	selector := "app = testingApp"
	containerName := "container1"
	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}
	pods := make([]*apiv1.Pod, livePods)
	eviction := &utils.PodsEvictionRestrictionMock{}

	recommender := &utils.RecommenderMock{}
	rec := utils.Recommendation(containerName, "2", "200M")

	for i := range pods {
		pods[i] = utils.BuildTestPod("test"+string(i), containerName, "1", "100M", &rc.ObjectMeta, &rc.TypeMeta)
		pods[i].Spec.NodeSelector = labels
		eviction.On("CanEvict", pods[i]).Return(true)
		eviction.On("Evict", pods[i]).Return(nil)
		recommender.On("Get", &pods[i].Spec).Return(rec, nil)
	}

	factory := &fakeEvictFactory{eviction}
	vpaLister := &utils.VerticalPodAutoscalerListerMock{}
	podLister := &utils.PodListerMock{}
	podLister.On("List").Return(pods, nil)

	vpaObj := utils.BuildTestVerticalPodAutoscaler(containerName, "1", "3", "100M", "1G", selector)
	vpaLister.On("List").Return([]*apimock.VerticalPodAutoscaler{vpaObj}, nil).Once()

	updater := &updater{
		vpaLister:        vpaLister,
		podLister:        podLister,
		recommender:      recommender,
		evictionFactrory: factory,
	}

	updater.RunOnce()
	eviction.AssertNumberOfCalls(t, "Evict", 5)
}

func TestRunOnceNotingToProcess(t *testing.T) {
	recommender := &utils.RecommenderMock{}
	eviction := &utils.PodsEvictionRestrictionMock{}
	factory := &fakeEvictFactory{eviction}
	vpaLister := &utils.VerticalPodAutoscalerListerMock{}
	podLister := &utils.PodListerMock{}
	podLister.On("List").Return(nil, nil).Once()
	vpaLister.On("List").Return(nil, nil).Once()

	updater := &updater{
		vpaLister:        vpaLister,
		podLister:        podLister,
		recommender:      recommender,
		evictionFactrory: factory,
	}
	updater.RunOnce()
}

type fakeEvictFactory struct {
	evict eviction.PodsEvictionRestriction
}

func (f fakeEvictFactory) NewPodsEvictionRestriction(pods []*apiv1.Pod) eviction.PodsEvictionRestriction {
	return f.evict
}
