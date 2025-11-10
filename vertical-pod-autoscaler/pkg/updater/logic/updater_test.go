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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	restriction "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/restriction"
	utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
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
		shouldInPlaceFail     bool
		expectFetchCalls      bool
		expectedEvictionCount int
		expectedInPlacedCount int
		canEvict              bool
		canInPlaceUpdate      utils.InPlaceDecision
		isCPUBoostTest        bool
	}{
		{
			name:                  "with Auto mode",
			updateMode:            vpa_types.UpdateModeAuto,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 5,
			expectedInPlacedCount: 0,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
		},
		{
			name:                  "with Initial mode",
			updateMode:            vpa_types.UpdateModeInitial,
			shouldInPlaceFail:     false,
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 0,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
		},
		{
			name:                  "with Off mode",
			updateMode:            vpa_types.UpdateModeOff,
			shouldInPlaceFail:     false,
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 0,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
		},
		{
			name:                  "with InPlaceOrRecreate mode expecting in-place updates",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 5,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
		},
		{
			name:                  "with InPlaceOrRecreate mode expecting fallback to evictions",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 5,
			expectedInPlacedCount: 0,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceEvict,
		},
		{
			name:                  "with InPlaceOrRecreate mode expecting no evictions or in-place",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 0,
			canEvict:              false,
			canInPlaceUpdate:      utils.InPlaceDeferred,
		},
		{
			name:                  "with InPlaceOrRecreate mode and failed in-place update",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     true,
			expectFetchCalls:      true,
			expectedEvictionCount: 5, // All pods should be evicted after in-place update fails
			expectedInPlacedCount: 5, // All pods attempt in-place update first
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
		},
		{
			name:                  "with InPlaceOrRecreate mode and unboost",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 5,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
			isCPUBoostTest:        true,
		},
		{
			name:                  "with Recreate mode and unboost",
			updateMode:            vpa_types.UpdateModeRecreate,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 5,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
			isCPUBoostTest:        true,
		},
		{
			name:                  "with Auto mode and unboost",
			updateMode:            vpa_types.UpdateModeAuto,
			shouldInPlaceFail:     false,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 5,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
			isCPUBoostTest:        true,
		},
		{
			name:                  "with InPlaceOrRecreate mode and unboost and In-place fails",
			updateMode:            vpa_types.UpdateModeInPlaceOrRecreate,
			shouldInPlaceFail:     true,
			expectFetchCalls:      true,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 5,
			canEvict:              true,
			canInPlaceUpdate:      utils.InPlaceApproved,
			isCPUBoostTest:        true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRunOnceBase(
				t,
				tc.updateMode,
				tc.shouldInPlaceFail,
				newFakeValidator(true),
				tc.expectFetchCalls,
				tc.expectedEvictionCount,
				tc.expectedInPlacedCount,
				tc.canInPlaceUpdate,
				tc.isCPUBoostTest,
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
		expectedInPlacedCount int
	}{
		{
			name:                  "with valid status",
			statusValidator:       newFakeValidator(true),
			expectFetchCalls:      true,
			expectedEvictionCount: 5,
			expectedInPlacedCount: 0,
		},
		{
			name:                  "with invalid status",
			statusValidator:       newFakeValidator(false),
			expectFetchCalls:      false,
			expectedEvictionCount: 0,
			expectedInPlacedCount: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testRunOnceBase(
				t,
				vpa_types.UpdateModeAuto,
				false,
				tc.statusValidator,
				tc.expectFetchCalls,
				tc.expectedEvictionCount,
				tc.expectedInPlacedCount,
				utils.InPlaceApproved,
				false,
			)
		})
	}
}

func testRunOnceBase(
	t *testing.T,
	updateMode vpa_types.UpdateMode,
	shouldInPlaceFail bool,
	statusValidator status.Validator,
	expectFetchCalls bool,
	expectedEvictionCount int,
	expectedInPlacedCount int,
	canInPlaceUpdate utils.InPlaceDecision,
	isCPUBoostTest bool,
) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, true)
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
	inplace := &test.PodsInPlaceRestrictionMock{}

	vpaObj := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G").
		WithTargetRef(&v1.CrossVersionObjectReference{
			Kind:       rc.Kind,
			Name:       rc.Name,
			APIVersion: rc.APIVersion,
		}).
		Get()

	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()

		pods[i].Labels = labels
		if isCPUBoostTest {
			pods[i].Annotations = map[string]string{
				"startup-cpu-boost": "",
			}
			pods[i].Status.Conditions = []apiv1.PodCondition{
				{
					Type:   apiv1.PodReady,
					Status: apiv1.ConditionTrue,
				},
			}
		}

		if !isCPUBoostTest {
			inplace.On("CanInPlaceUpdate", pods[i]).Return(canInPlaceUpdate)
			eviction.On("CanEvict", pods[i]).Return(true)
		} else {
			inplace.On("CanUnboost", pods[i], vpaObj).Return(isCPUBoostTest)
		}
		if shouldInPlaceFail {
			inplace.On("InPlaceUpdate", pods[i], nil).Return(fmt.Errorf("in-place update failed"))
		} else {
			inplace.On("InPlaceUpdate", pods[i], nil).Return(nil)
		}

		eviction.On("Evict", pods[i], nil).Return(nil)
	}

	factory := &restriction.FakePodsRestrictionFactory{
		Eviction: eviction,
		InPlace:  inplace,
	}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}

	podLister := &test.PodListerMock{}
	podLister.On("List").Return(pods, nil)

	vpaObj.Spec.UpdatePolicy = &vpa_types.PodUpdatePolicy{UpdateMode: &updateMode}
	if isCPUBoostTest {
		cpuStartupBoost := &vpa_types.GenericStartupBoost{
			Type:     vpa_types.FactorStartupBoostType,
			Duration: &metav1.Duration{Duration: 1 * time.Minute},
		}
		vpaObj.Spec.StartupBoost = &vpa_types.StartupBoost{
			CPU: cpuStartupBoost,
		}
	}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()

	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		restrictionFactory:           factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
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
	inplace.AssertNumberOfCalls(t, "InPlaceUpdate", expectedInPlacedCount)
}

func TestRunOnceNotingToProcess(t *testing.T) {
	eviction := &test.PodsEvictionRestrictionMock{}
	inplace := &test.PodsInPlaceRestrictionMock{}
	factory := &restriction.FakePodsRestrictionFactory{
		Eviction: eviction,
		InPlace:  inplace,
	}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	podLister := &test.PodListerMock{}
	vpaLister.On("List").Return(nil, nil).Once()

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		restrictionFactory:           factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
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
	inplace := &test.PodsInPlaceRestrictionMock{}
	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()

		pods[i].Labels = labels
		eviction.On("CanEvict", pods[i]).Return(true)
		eviction.On("Evict", pods[i], nil).Return(nil)
	}

	factory := &restriction.FakePodsRestrictionFactory{
		Eviction: eviction,
		InPlace:  inplace,
	}
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
		WithTargetRef(targetRef).
		Get()

	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()

	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		restrictionFactory:           factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
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
	eviction.AssertNumberOfCalls(t, "InPlaceUpdate", 0)
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

func TestLogDeprecationWarnings(t *testing.T) {
	tests := []struct {
		name             string
		updateMode       *vpa_types.UpdateMode
		updatePolicy     *vpa_types.PodUpdatePolicy
		shouldLogWarning bool
	}{
		{
			name:             "Auto mode should trigger deprecation warning logic",
			updateMode:       &[]vpa_types.UpdateMode{vpa_types.UpdateModeAuto}[0],
			shouldLogWarning: true,
		},
		{
			name:             "Recreate mode should not trigger warning logic",
			updateMode:       &[]vpa_types.UpdateMode{vpa_types.UpdateModeRecreate}[0],
			shouldLogWarning: false,
		},
		{
			name:             "Initial mode should not trigger warning logic",
			updateMode:       &[]vpa_types.UpdateMode{vpa_types.UpdateModeInitial}[0],
			shouldLogWarning: false,
		},
		{
			name:             "nil update policy should not trigger warning logic",
			updatePolicy:     nil,
			shouldLogWarning: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var updatePolicy *vpa_types.PodUpdatePolicy
			if tc.updatePolicy != nil {
				updatePolicy = tc.updatePolicy
			} else if tc.updateMode != nil {
				updatePolicy = &vpa_types.PodUpdatePolicy{
					UpdateMode: tc.updateMode,
				}
			}

			vpa := &vpa_types.VerticalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vpa",
					Namespace: "default",
				},
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: updatePolicy,
				},
			}

			shouldLogWarning := vpa.Spec.UpdatePolicy != nil &&
				vpa.Spec.UpdatePolicy.UpdateMode != nil &&
				*vpa.Spec.UpdatePolicy.UpdateMode == vpa_types.UpdateModeAuto

			assert.Equal(t, tc.shouldLogWarning, shouldLogWarning,
				"Expected shouldLogWarning=%v for test case %s", tc.shouldLogWarning, tc.name)

			// Call the function to ensure it doesn't panic
			logDeprecationWarnings(vpa)
		})
	}
}
func TestRunOnce_AutoUnboostThenEvict(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, true)
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
		ObjectMeta: metav1.ObjectMeta{Name: "rc", Namespace: "default"},
		Spec:       apiv1.ReplicationControllerSpec{Replicas: &replicas},
	}
	pods := make([]*apiv1.Pod, livePods)
	vpaObj := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G").
		WithTargetRef(&v1.CrossVersionObjectReference{Kind: rc.Kind, Name: rc.Name, APIVersion: rc.APIVersion}).
		WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, "1m").
		Get()

	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()
		pods[i].Labels = labels
	}

	eviction := &test.PodsEvictionRestrictionMock{}
	inplace := &test.PodsInPlaceRestrictionMock{}
	factory := &restriction.FakePodsRestrictionFactory{Eviction: eviction, InPlace: inplace}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	podLister := &test.PodListerMock{}
	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		restrictionFactory:           factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		selectorFetcher:              mockSelectorFetcher,
		controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
		useAdmissionControllerStatus: true,
		statusValidator:              newFakeValidator(true),
		priorityProcessor:            priority.NewProcessor(),
	}

	// Cycle 1: Unboost the cpu
	for i := range pods {
		pods[i].Annotations = map[string]string{"startup-cpu-boost": ""}
		pods[i].Status.Conditions = []apiv1.PodCondition{
			{
				Type:   apiv1.PodReady,
				Status: apiv1.ConditionTrue,
			},
		}
		inplace.On("CanUnboost", pods[i], vpaObj).Return(true).Once()
		inplace.On("InPlaceUpdate", pods[i], nil).Return(nil)
	}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()
	podLister.On("List").Return(pods, nil).Once()
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater.RunOnce(context.Background())
	inplace.AssertNumberOfCalls(t, "InPlaceUpdate", 5)
	inplace.AssertNumberOfCalls(t, "CanUnboost", 5)
	eviction.AssertNumberOfCalls(t, "Evict", 0)

	// Cycle 2: Regular patch which will lead to eviction
	for i := range pods {
		pods[i].Annotations = nil
		inplace.On("CanUnboost", pods[i], vpaObj).Return(false).Once()
		eviction.On("CanEvict", pods[i]).Return(true)
		eviction.On("Evict", pods[i], nil).Return(nil)
	}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()
	podLister.On("List").Return(pods, nil).Once()
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater.RunOnce(context.Background())
	inplace.AssertNumberOfCalls(t, "InPlaceUpdate", 5) // all 5 from previous run only
	inplace.AssertNumberOfCalls(t, "CanUnboost", 5)    // all 5 from previous run only
	eviction.AssertNumberOfCalls(t, "Evict", 5)
}

func TestRunOnce_AutoUnboostThenInPlace(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, true)

	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, true)
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
		ObjectMeta: metav1.ObjectMeta{Name: "rc", Namespace: "default"},
		Spec:       apiv1.ReplicationControllerSpec{Replicas: &replicas},
	}
	pods := make([]*apiv1.Pod, livePods)
	vpaObj := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G").
		WithTargetRef(&v1.CrossVersionObjectReference{Kind: rc.Kind, Name: rc.Name, APIVersion: rc.APIVersion}).
		WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, "1m").
		Get()

	for i := range pods {
		pods[i] = test.Pod().WithName("test_"+strconv.Itoa(i)).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
			Get()
		pods[i].Labels = labels
	}

	eviction := &test.PodsEvictionRestrictionMock{}
	inplace := &test.PodsInPlaceRestrictionMock{}
	factory := &restriction.FakePodsRestrictionFactory{Eviction: eviction, InPlace: inplace}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	podLister := &test.PodListerMock{}
	mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

	updater := &updater{
		vpaLister:                    vpaLister,
		podLister:                    podLister,
		restrictionFactory:           factory,
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		selectorFetcher:              mockSelectorFetcher,
		controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
		useAdmissionControllerStatus: true,
		statusValidator:              newFakeValidator(true),
		priorityProcessor:            priority.NewProcessor(),
	}

	// Cycle 1: Unboost the cpu
	for i := range pods {
		pods[i].Annotations = map[string]string{"startup-cpu-boost": ""}
		pods[i].Status.Conditions = []apiv1.PodCondition{
			{
				Type:   apiv1.PodReady,
				Status: apiv1.ConditionTrue,
			},
		}
		inplace.On("CanUnboost", pods[i], vpaObj).Return(true).Once()
		inplace.On("InPlaceUpdate", pods[i], nil).Return(nil)
	}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()
	podLister.On("List").Return(pods, nil).Once()
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater.RunOnce(context.Background())
	inplace.AssertNumberOfCalls(t, "InPlaceUpdate", 5)
	inplace.AssertNumberOfCalls(t, "CanUnboost", 5)
	eviction.AssertNumberOfCalls(t, "Evict", 0)

	// Cycle 2: Regular patch which will lead to eviction
	for i := range pods {
		pods[i].Annotations = nil
		inplace.On("CanInPlaceUpdate", pods[i]).Return(utils.InPlaceApproved)
		inplace.On("InPlaceUpdate", pods[i], nil).Return(nil)
	}
	vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()
	podLister.On("List").Return(pods, nil).Once()
	mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(selector, nil)

	updater.RunOnce(context.Background())
	inplace.AssertNumberOfCalls(t, "InPlaceUpdate", 10)
	inplace.AssertNumberOfCalls(t, "CanUnboost", 5) // all 5 from previous run only
	eviction.AssertNumberOfCalls(t, "Evict", 0)
}
