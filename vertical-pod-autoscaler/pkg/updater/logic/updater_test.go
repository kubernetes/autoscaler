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

package logic

import (
	"context"
	"strconv"
	"testing"
	"time"

	"golang.org/x/time/rate"
	v1 "k8s.io/api/autoscaling/v1"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/eviction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestRunOnce_Mode(t *testing.T) {
	tests := []struct {
		name                  string
		updateMode            vpa_types.UpdateMode
		expectFetchCalls      bool
		expectedEvictionCount int
	}{
		{
			name:                  "with Auto mode",
			updateMode:            vpa_types.UpdateModeAuto,
			expectFetchCalls:      true,
			expectedEvictionCount: 5,
		},
		{
			name:                  "with Initial mode",
			updateMode:            vpa_types.UpdateModeInitial,
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
		},
		{
			name:                  "with Off mode",
			updateMode:            vpa_types.UpdateModeOff,
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRunOnceBase(
				t,
				tc.updateMode,
				newFakeValidator(true),
				tc.expectFetchCalls,
				tc.expectedEvictionCount,
			)
		})
	}
}

func TestRunOnce_Status(t *testing.T) {
	tests := []struct {
		name                  string
		statusValidator       status.Validator
		expectFetchCalls      bool
		expectedEvictionCount int
	}{
		{
			name:                  "with valid status",
			statusValidator:       newFakeValidator(true),
			expectFetchCalls:      true,
			expectedEvictionCount: 5,
		},
		{
			name:                  "with invalid status",
			statusValidator:       newFakeValidator(false),
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRunOnceBase(
				t,
				vpa_types.UpdateModeAuto,
				tc.statusValidator,
				tc.expectFetchCalls,
				tc.expectedEvictionCount,
			)
		})
	}
}

func testRunOnceBase(
	t *testing.T,
	updateMode vpa_types.UpdateMode,
	statusValidator status.Validator,
	expectFetchCalls bool,
	expectedEvictionCount int,
) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	replicas := int32(5)
	livePods := 5
	labels := map[string]string{"app": "testingApp"}
	selector := parseLabelSelector("app = testingApp")
	containerName := "container1"
	rc := apiv1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}
	pods := make([]*apiv1.Pod, livePods)
	eviction := &test.PodsEvictionRestrictionMock{}

	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()

		pods[i].Labels = labels
		// We will test in-place separately, but we do need to account for these calls
		eviction.On("CanInPlaceUpdate", pods[i]).Return(false)
		eviction.On("IsInPlaceUpdating", pods[i]).Return(false)

		eviction.On("CanEvict", pods[i]).Return(true)
		eviction.On("Evict", pods[i], nil).Return(nil)
	}

	factory := &fakeEvictFactory{eviction}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}

	podLister := &test.PodListerMock{}
	podLister.On("List").Return(pods, nil)
	targetRef := &v1.CrossVersionObjectReference{
		Kind:       rc.Kind,
		Name:       rc.Name,
		APIVersion: rc.APIVersion,
	}
	vpaObj := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G").
		WithTargetRef(targetRef).Get()

	vpaObj.Spec.UpdatePolicy = &vpa_types.PodUpdatePolicy{UpdateMode: &updateMode}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()

	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		evictionFactory:              factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		selectorFetcher:              mockSelectorFetcher,
		controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
		useAdmissionControllerStatus: true,
		statusValidator:              statusValidator,
		priorityProcessor:            priority.NewProcessor(),
	}

	if expectFetchCalls {
		mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)
	}
	updater.RunOnce(context.Background())
	eviction.AssertNumberOfCalls(t, "Evict", expectedEvictionCount)
}

func TestRunOnceNotingToProcess(t *testing.T) {
	eviction := &test.PodsEvictionRestrictionMock{}
	factory := &fakeEvictFactory{eviction}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	podLister := &test.PodListerMock{}
	vpaLister.On("List").Return(nil, nil).Once()

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		evictionFactory:              factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		useAdmissionControllerStatus: true,
		statusValidator:              newFakeValidator(true),
	}
	updater.RunOnce(context.Background())
}

func TestGetRateLimiter(t *testing.T) {
	cases := []struct {
		rateLimit       float64
		rateLimitBurst  int
		expectedLimiter *rate.Limiter
	}{
		{0.0, 1, rate.NewLimiter(rate.Inf, 0)},
		{-1.0, 2, rate.NewLimiter(rate.Inf, 0)},
		{10.0, 3, rate.NewLimiter(rate.Limit(10), 3)},
	}
	for _, tc := range cases {
		limiter := getRateLimiter(tc.rateLimit, tc.rateLimitBurst)
		assert.Equal(t, tc.expectedLimiter.Burst(), limiter.Burst())
		assert.InDelta(t, float64(tc.expectedLimiter.Limit()), float64(limiter.Limit()), 1e-6)
	}
}

type fakeEvictFactory struct {
	evict eviction.PodsEvictionRestriction
}

func (f fakeEvictFactory) NewPodsEvictionRestriction(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) eviction.PodsEvictionRestriction {
	return f.evict
}

type fakeValidator struct {
	isValid bool
}

func newFakeValidator(isValid bool) status.Validator {
	return &fakeValidator{isValid}
}

func (f *fakeValidator) IsStatusValid(ctx context.Context, statusTimeout time.Duration) (bool, error) {
	return f.isValid, nil
}

func TestRunOnceIgnoreNamespaceMatchingPods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	replicas := int32(5)
	livePods := 5
	labels := map[string]string{"app": "testingApp"}
	selector := parseLabelSelector("app = testingApp")

	containerName := "container1"
	rc := apiv1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	}
	pods := make([]*apiv1.Pod, livePods)
	eviction := &test.PodsEvictionRestrictionMock{}

	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()

		pods[i].Labels = labels
		eviction.On("CanEvict", pods[i]).Return(true)
		eviction.On("Evict", pods[i], nil).Return(nil)
	}

	factory := &fakeEvictFactory{eviction}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}

	podLister := &test.PodListerMock{}
	podLister.On("List").Return(pods, nil)
	targetRef := &v1.CrossVersionObjectReference{
		Kind:       rc.Kind,
		Name:       rc.Name,
		APIVersion: rc.APIVersion,
	}
	vpaObj := test.VerticalPodAutoscaler().
		WithNamespace("default").
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G").
		WithTargetRef(targetRef).Get()

	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()

	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		evictionFactory:              factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		selectorFetcher:              mockSelectorFetcher,
		controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
		useAdmissionControllerStatus: true,
		priorityProcessor:            priority.NewProcessor(),
		ignoredNamespaces:            []string{"not-default"},
		statusValidator:              newFakeValidator(true),
	}

	updater.RunOnce(context.Background())
	eviction.AssertNumberOfCalls(t, "Evict", 5)
}

func TestRunOnceIgnoreNamespaceMatching(t *testing.T) {
	eviction := &test.PodsEvictionRestrictionMock{}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	vpaObj := test.VerticalPodAutoscaler().
		WithNamespace("default").
		WithContainer("container").Get()

	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()

	updater := &updater{
		vpaLister:         vpaLister,
		ignoredNamespaces: []string{"default"},
	}

	updater.RunOnce(context.Background())
	eviction.AssertNumberOfCalls(t, "Evict", 0)
}

func TestNewEventRecorder(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	er := newEventRecorder(fakeClient)

	maxRetries := 5
	retryDelay := 100 * time.Millisecond
	contextTimeout := 5 * time.Second

	testCases := []struct {
		reason  string
		object  runtime.Object
		message string
	}{
		{
			reason:  "EvictedPod",
			object:  &apiv1.Pod{},
			message: "Evicted pod",
		},
		{
			reason:  "EvictedPod",
			object:  &vpa_types.VerticalPodAutoscaler{},
			message: "Evicted pod",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.reason, func(t *testing.T) {
			er.Event(tc.object, apiv1.EventTypeNormal, tc.reason, tc.message)

			var events *apiv1.EventList
			var err error
			// Add delay for fake client to catch up due to be being asynchronous
			for i := 0; i < maxRetries; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
				defer cancel()
				events, err = fakeClient.CoreV1().Events("default").List(ctx, metav1.ListOptions{})
				if err == nil && len(events.Items) > 0 {
					break
				}
				time.Sleep(retryDelay)
			}

			assert.NoError(t, err, "should be able to list events")
			assert.Equal(t, 1, len(events.Items), "should have exactly 1 event")

			event := events.Items[0]
			assert.Equal(t, tc.reason, event.Reason)
			assert.Equal(t, tc.message, event.Message)
			assert.Equal(t, apiv1.EventTypeNormal, event.Type)
			assert.Equal(t, "vpa-updater", event.Source.Component)
		})
	}
}
