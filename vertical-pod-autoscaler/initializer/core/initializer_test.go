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

package core

import (
	"testing"

	"k8s.io/autoscaler/vertical-pod-autoscaler/apimock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils"
	recommender "k8s.io/autoscaler/vertical-pod-autoscaler/recommender_mock"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateResourceRequests(t *testing.T) {
	type testCase struct {
		pod            *apiv1.Pod
		vpa            *apimock.VerticalPodAutoscaler
		recommender    recommender.CachingRecommender
		expectedAction bool
		expectedMem    string
		expectedCPU    string
	}
	containerName := "container1"
	labels := map[string]string{"app": "testingApp"}
	vpa := utils.BuildTestVerticalPodAutoscaler(containerName, "1", "3", "100M", "1G", "app = testingApp")

	recommender := &utils.RecommenderMock{}
	rec := utils.Recommendation(containerName, "2", "200M")

	uninitialized := utils.BuildTestPod("test_uninitialized", containerName, "1", "100M", nil, nil)
	uninitialized.ObjectMeta.Labels = labels
	uninitialized.ObjectMeta.Initializers = &metav1.Initializers{
		Pending: []metav1.Initializer{{Name: VPAInitializerName}},
	}
	recommender.On("Get", &uninitialized.Spec).Return(rec, nil)

	initialized := utils.BuildTestPod("test_initialized", containerName, "1", "100M", nil, nil)
	initialized.ObjectMeta.Labels = labels
	recommender.On("Get", &initialized.Spec).Return(rec, nil)

	mismatchedVPA := utils.BuildTestVerticalPodAutoscaler(containerName, "1", "3", "100M", "1G", "app = differentApp")

	testCases := []testCase{{
		pod:            uninitialized,
		recommender:    recommender,
		vpa:            vpa,
		expectedAction: true,
		expectedMem:    "200M",
		expectedCPU:    "2",
	}, {
		pod:            initialized,
		recommender:    recommender,
		vpa:            vpa,
		expectedAction: false,
	}, {
		pod:            uninitialized,
		recommender:    recommender,
		vpa:            mismatchedVPA,
		expectedAction: false,
	}}
	for _, tc := range testCases {
		vpaLister := &utils.VerticalPodAutoscalerListerMock{}
		vpaLister.On("List").Return([]*apimock.VerticalPodAutoscaler{tc.vpa}, nil)

		podList := apiv1.PodList{Items: []apiv1.Pod{*tc.pod}}
		testClient := fake.NewSimpleClientset(&podList)

		initializer := &initializer{
			recommender: tc.recommender,
			vpaLister:   vpaLister,
			client:      testClient,
		}

		initializer.updateResourceRequests(tc.pod)

		if tc.expectedAction {
			assert.Equal(t, 1, len(testClient.Actions()))
			updated := testClient.Actions()[0].(core.UpdateAction).GetObject().(*apiv1.Pod)
			assert.Equal(t, tc.pod.ObjectMeta.Name, updated.ObjectMeta.Name)
			assert.Equal(t, resource.MustParse(tc.expectedMem), *updated.Spec.Containers[0].Resources.Requests.Memory())
			assert.Equal(t, resource.MustParse(tc.expectedCPU), *updated.Spec.Containers[0].Resources.Requests.Cpu())
		}
	}
}
