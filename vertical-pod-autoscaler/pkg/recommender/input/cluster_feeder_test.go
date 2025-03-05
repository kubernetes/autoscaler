/*
Copyright 2019 The Kubernetes Authors.

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

package input

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	core "k8s.io/client-go/testing"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	fakeautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type fakeControllerFetcher struct {
	key *controllerfetcher.ControllerKeyWithAPIVersion
	err error
}

func (f *fakeControllerFetcher) FindTopMostWellKnownOrScalable(_ context.Context, _ *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

var (
	recommenderName                     = "name"
	empty                               = ""
	unsupportedConditionTextFromFetcher = "Cannot read targetRef. Reason: targetRef not defined"
	unsupportedConditionNoExtraText     = "Cannot read targetRef"
	unsupportedConditionNoTargetRef     = "Cannot read targetRef"
	unsupportedConditionMudaMudaMuda    = "Error checking if target is a topmost well-known or scalable controller: muda muda muda"
	unsupportedTargetRefHasParent       = "The targetRef controller has a parent but it should point to a topmost well-known or scalable controller"
)

const (
	kind         = "dodokind"
	name1        = "dotaro"
	name2        = "doseph"
	namespace    = "testNamespace"
	apiVersion   = "stardust"
	testGcPeriod = time.Minute
)

// NewClusterState returns a new clusterState with no pods.
func NewFakeClusterState(vpas map[model.VpaID]*model.Vpa, pods map[model.PodID]*model.PodState) *fakeClusterState {
	return &fakeClusterState{
		stubbedVPAs:  vpas,
		stubbedPods:  pods,
		addedSamples: make(map[model.ContainerID][]*model.ContainerUsageSampleWithKey),
	}
}

type fakeClusterState struct {
	model.ClusterState
	addedPods    []model.PodID
	addedSamples map[model.ContainerID][]*model.ContainerUsageSampleWithKey
	stubbedVPAs  map[model.VpaID]*model.Vpa
	stubbedPods  map[model.PodID]*model.PodState
}

func (cs *fakeClusterState) AddSample(sample *model.ContainerUsageSampleWithKey) error {
	samplesForContainer := cs.addedSamples[sample.Container]
	cs.addedSamples[sample.Container] = append(samplesForContainer, sample)
	return nil
}

func (cs *fakeClusterState) AddOrUpdatePod(podID model.PodID, _ labels.Set, _ v1.PodPhase) {
	cs.addedPods = append(cs.addedPods, podID)
}

func (cs *fakeClusterState) Pods() map[model.PodID]*model.PodState {
	return cs.stubbedPods
}

func (cs *fakeClusterState) VPAs() map[model.VpaID]*model.Vpa {
	return cs.stubbedVPAs
}

func (cs *fakeClusterState) StateMapSize() int {
	return 0
}

func TestLoadVPAs(t *testing.T) {

	type testCase struct {
		name                                string
		selector                            labels.Selector
		fetchSelectorError                  error
		targetRef                           *autoscalingv1.CrossVersionObjectReference
		topMostWellKnownOrScalableKey       *controllerfetcher.ControllerKeyWithAPIVersion
		findTopMostWellKnownOrScalableError error
		expectedSelector                    labels.Selector
		expectedConfigUnsupported           *string
		expectedConfigDeprecated            *string
		expectedVpaFetch                    bool
		recommenderName                     *string
		recommender                         string
	}

	testCases := []testCase{
		{
			name:                      "no selector",
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionTextFromFetcher,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:                      "also no selector but no error",
			selector:                  nil,
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoExtraText,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "targetRef selector",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:                      "no targetRef",
			selector:                  parseLabelSelector("app = test"),
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "can't decide if top-level-ref",
			selector:           nil,
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedConditionNoTargetRef,
			expectedVpaFetch:          true,
		},
		{
			name:               "non-top-level targetRef",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name2,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedTargetRefHasParent,
			expectedVpaFetch:          true,
		},
		{
			name:               "error checking if top-level-ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "doestar",
				Name:       "doseph-doestar",
				APIVersion: "taxonomy",
			},
			expectedConfigUnsupported:           &unsupportedConditionMudaMudaMuda,
			expectedVpaFetch:                    true,
			findTopMostWellKnownOrScalableError: fmt.Errorf("muda muda muda"),
		},
		{
			name:               "top-level target ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   parseLabelSelector("app = test"),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "no recommenderName",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          false,
			recommenderName:           &empty,
		},
		{
			name:               "recommenderName doesn't match recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          false,
			recommenderName:           &recommenderName,
			recommender:               "other",
		},
		{
			name:               "recommenderName matches recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
			recommenderName:           &recommenderName,
			recommender:               recommenderName,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vpaBuilder := test.VerticalPodAutoscaler().WithName("testVpa").WithContainer("container").WithNamespace("testNamespace").WithTargetRef(tc.targetRef)
			if tc.recommender != "" {
				vpaBuilder = vpaBuilder.WithRecommender(tc.recommender)
			}
			vpa := vpaBuilder.Get()
			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpa}, nil)

			targetSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
			clusterState := model.NewClusterState(testGcPeriod)

			clusterStateFeeder := clusterStateFeeder{
				vpaLister:       vpaLister,
				clusterState:    clusterState,
				selectorFetcher: targetSelectorFetcher,
				controllerFetcher: &fakeControllerFetcher{
					key: tc.topMostWellKnownOrScalableKey,
					err: tc.findTopMostWellKnownOrScalableError,
				},
			}
			if tc.recommenderName == nil {
				clusterStateFeeder.recommenderName = DefaultRecommenderName
			} else {
				clusterStateFeeder.recommenderName = *tc.recommenderName
			}

			if tc.expectedVpaFetch {
				targetSelectorFetcher.EXPECT().Fetch(vpa).Return(tc.selector, tc.fetchSelectorError)
			}
			clusterStateFeeder.LoadVPAs(context.Background())

			vpaID := model.VpaID{
				Namespace: vpa.Namespace,
				VpaName:   vpa.Name,
			}

			if !tc.expectedVpaFetch {
				assert.NotContains(t, clusterState.VPAs(), vpaID)
				return
			}
			assert.Contains(t, clusterState.VPAs(), vpaID)
			storedVpa := clusterState.VPAs()[vpaID]
			if tc.expectedSelector != nil {
				assert.NotNil(t, storedVpa.PodSelector)
				assert.Equal(t, tc.expectedSelector.String(), storedVpa.PodSelector.String())
			} else {
				assert.Nil(t, storedVpa.PodSelector)
			}

			if tc.expectedConfigDeprecated != nil {
				assert.Contains(t, storedVpa.Conditions, vpa_types.ConfigDeprecated)
				assert.Equal(t, *tc.expectedConfigDeprecated, storedVpa.Conditions[vpa_types.ConfigDeprecated].Message)
			} else {
				assert.NotContains(t, storedVpa.Conditions, vpa_types.ConfigDeprecated)
			}

			if tc.expectedConfigUnsupported != nil {
				assert.Contains(t, storedVpa.Conditions, vpa_types.ConfigUnsupported)
				assert.Equal(t, *tc.expectedConfigUnsupported, storedVpa.Conditions[vpa_types.ConfigUnsupported].Message)
			} else {
				assert.NotContains(t, storedVpa.Conditions, vpa_types.ConfigUnsupported)
			}

		})
	}
}

type testSpecClient struct {
	pods []*spec.BasicPodSpec
}

func (c *testSpecClient) GetPodSpecs() ([]*spec.BasicPodSpec, error) {
	return c.pods, nil
}

func makeTestSpecClient(podLabels []map[string]string) spec.SpecClient {
	pods := make([]*spec.BasicPodSpec, len(podLabels))
	for i, l := range podLabels {
		pods[i] = &spec.BasicPodSpec{
			ID:        model.PodID{Namespace: "default", PodName: fmt.Sprintf("pod-%d", i)},
			PodLabels: l,
		}
	}
	return &testSpecClient{
		pods: pods,
	}
}

func newTestContainerSpec(podID model.PodID, containerName string, milicores int, memory int64) spec.BasicContainerSpec {
	containerID := model.ContainerID{
		PodID:         podID,
		ContainerName: containerName,
	}
	requestedResources := model.Resources{
		model.ResourceCPU:    model.ResourceAmount(milicores),
		model.ResourceMemory: model.ResourceAmount(memory),
	}
	return spec.BasicContainerSpec{
		ID:      containerID,
		Image:   containerName + "Image",
		Request: requestedResources,
	}
}

func newTestPodSpec(podId model.PodID, containerSpecs []spec.BasicContainerSpec, initContainerSpecs []spec.BasicContainerSpec) *spec.BasicPodSpec {
	return &spec.BasicPodSpec{
		ID:             podId,
		PodLabels:      map[string]string{podId.PodName + "LabelKey": podId.PodName + "LabelValue"},
		Containers:     containerSpecs,
		InitContainers: initContainerSpecs,
	}
}

func TestClusterStateFeeder_LoadPods_ContainerTracking(t *testing.T) {
	podWithoutInitContainersID := model.PodID{Namespace: "default", PodName: "PodWithoutInitContainers"}
	containerSpecs := []spec.BasicContainerSpec{
		newTestContainerSpec(podWithoutInitContainersID, "container1", 500, 512*1024*1024),
		newTestContainerSpec(podWithoutInitContainersID, "container2", 1000, 1024*1024*1024),
	}
	podWithoutInitContainers := newTestPodSpec(podWithoutInitContainersID, containerSpecs, nil)

	podWithInitContainersID := model.PodID{Namespace: "default", PodName: "PodWithInitContainers"}
	containerSpecs2 := []spec.BasicContainerSpec{
		newTestContainerSpec(podWithInitContainersID, "container1", 2000, 2048*1024*1024),
	}
	initContainerSpecs2 := []spec.BasicContainerSpec{
		newTestContainerSpec(podWithInitContainersID, "init1", 40, 128*1024*1024),
		newTestContainerSpec(podWithInitContainersID, "init2", 100, 256*1024*1024),
	}
	podWithInitContainers := newTestPodSpec(podWithInitContainersID, containerSpecs2, initContainerSpecs2)

	client := &testSpecClient{pods: []*spec.BasicPodSpec{podWithoutInitContainers, podWithInitContainers}}

	clusterState := model.NewClusterState(testGcPeriod)

	feeder := clusterStateFeeder{
		specClient:     client,
		memorySaveMode: false,
		clusterState:   clusterState,
	}

	feeder.LoadPods()

	assert.Equal(t, len(feeder.clusterState.Pods()), 2)
	assert.Equal(t, len(feeder.clusterState.Pods()[podWithInitContainersID].Containers), 1)
	assert.Equal(t, len(feeder.clusterState.Pods()[podWithInitContainersID].InitContainers), 2)
	assert.Equal(t, len(feeder.clusterState.Pods()[podWithoutInitContainersID].Containers), 2)
	assert.Equal(t, len(feeder.clusterState.Pods()[podWithoutInitContainersID].InitContainers), 0)

}

func TestClusterStateFeeder_LoadPods_MemorySaverMode(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		VPALabelSelectors []string
		PodLabels         []map[string]string
		TrackedPods       int
	}{
		{
			Name:              "simple",
			VPALabelSelectors: []string{"name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod"},
				{"type": "stateful"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "multiple",
			VPALabelSelectors: []string{"name=vpa-pod,type=stateful"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod", "type": "stateful"},
				{"type": "stateful"},
				{"name": "vpa-pod"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "no matches",
			VPALabelSelectors: []string{"name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "non-vpa-pod", "type": "stateful"},
			},
			TrackedPods: 0,
		},
		{
			Name:              "set based",
			VPALabelSelectors: []string{"environment in (staging, qa),name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod", "environment": "staging"},
				{"name": "vpa-pod", "environment": "production"},
				{"name": "non-vpa-pod", "environment": "staging"},
				{"name": "non-vpa-pod", "environment": "production"},
			},
			TrackedPods: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			vpas := make(map[model.VpaID]*model.Vpa)
			for i, selector := range tc.VPALabelSelectors {
				vpaLabel, err := labels.Parse(selector)
				assert.NoError(t, err)
				key := model.VpaID{VpaName: fmt.Sprintf("test-vpa-%d", i), Namespace: "default"}
				vpas[key] = &model.Vpa{PodSelector: vpaLabel}
			}
			clusterState := NewFakeClusterState(vpas, nil)

			feeder := clusterStateFeeder{
				specClient:     makeTestSpecClient(tc.PodLabels),
				memorySaveMode: true,
				clusterState:   clusterState,
			}

			feeder.LoadPods()
			assert.Len(t, clusterState.addedPods, tc.TrackedPods, "number of pods is not %d", tc.TrackedPods)

			clusterState = NewFakeClusterState(vpas, nil)

			feeder = clusterStateFeeder{
				specClient:     makeTestSpecClient(tc.PodLabels),
				memorySaveMode: false,
				clusterState:   clusterState,
			}

			feeder.LoadPods()
			assert.Len(t, clusterState.addedPods, len(tc.PodLabels), "number of pods is not %d", len(tc.PodLabels))
		})
	}
}

func newContainerMetricsSnapshot(id model.ContainerID, cpuUsage int64, memUsage int64) (*metrics.ContainerMetricsSnapshot, []*model.ContainerUsageSampleWithKey) {
	snapshotTimestamp := time.Now()
	snapshotWindow := time.Duration(1234)
	snapshot := &metrics.ContainerMetricsSnapshot{
		ID:             id,
		SnapshotTime:   snapshotTimestamp,
		SnapshotWindow: snapshotWindow,
		Usage: model.Resources{
			model.ResourceCPU:    model.ResourceAmount(cpuUsage),
			model.ResourceMemory: model.ResourceAmount(memUsage),
		},
	}
	samples := []*model.ContainerUsageSampleWithKey{
		{
			Container: id,
			ContainerUsageSample: model.ContainerUsageSample{
				MeasureStart: snapshotTimestamp,
				Resource:     model.ResourceCPU,
				Usage:        model.ResourceAmount(cpuUsage),
			},
		},
		{
			Container: id,
			ContainerUsageSample: model.ContainerUsageSample{
				MeasureStart: snapshotTimestamp,
				Resource:     model.ResourceMemory,
				Usage:        model.ResourceAmount(memUsage),
			},
		},
	}
	return snapshot, samples
}

type fakeMetricsClient struct {
	snapshots []*metrics.ContainerMetricsSnapshot
}

func (m fakeMetricsClient) GetContainersMetrics() ([]*metrics.ContainerMetricsSnapshot, error) {
	return m.snapshots, nil
}

func TestClusterStateFeeder_LoadRealTimeMetrics(t *testing.T) {
	namespaceName := "test-namespace"
	podID := model.PodID{Namespace: namespaceName, PodName: "Pod"}
	regularContainer1 := model.ContainerID{PodID: podID, ContainerName: "Container1"}
	regularContainer2 := model.ContainerID{PodID: podID, ContainerName: "Container2"}
	initContainer := model.ContainerID{PodID: podID, ContainerName: "InitContainer"}

	pods := map[model.PodID]*model.PodState{
		podID: {ID: podID,
			Containers: map[string]*model.ContainerState{
				"Container1": {},
				"Container2": {},
			},
			InitContainers: []string{
				"InitContainer",
			}},
	}

	var containerMetricsSnapshots []*metrics.ContainerMetricsSnapshot

	regularContainer1MetricsSnapshot, regularContainer1UsageSamples := newContainerMetricsSnapshot(regularContainer1, 100, 1024)
	containerMetricsSnapshots = append(containerMetricsSnapshots, regularContainer1MetricsSnapshot)
	regularContainer2MetricsSnapshot, regularContainer2UsageSamples := newContainerMetricsSnapshot(regularContainer2, 200, 2048)
	containerMetricsSnapshots = append(containerMetricsSnapshots, regularContainer2MetricsSnapshot)
	initContainer1MetricsSnapshots, _ := newContainerMetricsSnapshot(initContainer, 300, 3072)
	containerMetricsSnapshots = append(containerMetricsSnapshots, initContainer1MetricsSnapshots)

	clusterState := NewFakeClusterState(nil, pods)

	feeder := clusterStateFeeder{
		memorySaveMode: false,
		clusterState:   clusterState,
		metricsClient:  fakeMetricsClient{snapshots: containerMetricsSnapshots},
	}

	feeder.LoadRealTimeMetrics()

	assert.Equal(t, 2, len(clusterState.addedSamples))

	samplesForContainer1 := clusterState.addedSamples[regularContainer1]
	assert.Contains(t, samplesForContainer1, regularContainer1UsageSamples[0])
	assert.Contains(t, samplesForContainer1, regularContainer1UsageSamples[1])

	samplesForContainer2 := clusterState.addedSamples[regularContainer2]
	assert.Contains(t, samplesForContainer2, regularContainer2UsageSamples[0])
	assert.Contains(t, samplesForContainer2, regularContainer2UsageSamples[1])
}

type fakeHistoryProvider struct {
	history map[model.PodID]*history.PodHistory
	err     error
}

func (fhp *fakeHistoryProvider) GetClusterHistory() (map[model.PodID]*history.PodHistory, error) {
	return fhp.history, fhp.err
}

func TestClusterStateFeeder_InitFromHistoryProvider(t *testing.T) {
	pod1 := model.PodID{
		Namespace: "ns",
		PodName:   "a-pod",
	}
	memAmount := model.ResourceAmount(128 * 1024 * 1024)
	t0 := time.Date(2021, time.August, 30, 10, 21, 0, 0, time.UTC)
	containerCpu := "containerCpu"
	containerMem := "containerMem"
	pod1History := history.PodHistory{
		LastLabels: map[string]string{},
		LastSeen:   t0,
		Samples: map[string][]model.ContainerUsageSample{
			containerCpu: {
				{
					MeasureStart: t0,
					Usage:        10,
					Resource:     model.ResourceCPU,
				},
			},
			containerMem: {
				{
					MeasureStart: t0,
					Usage:        memAmount,
					Resource:     model.ResourceMemory,
				},
			},
		},
	}
	provider := fakeHistoryProvider{
		history: map[model.PodID]*history.PodHistory{
			pod1: &pod1History,
		},
	}

	clusterState := model.NewClusterState(testGcPeriod)
	feeder := clusterStateFeeder{
		clusterState: clusterState,
	}
	feeder.InitFromHistoryProvider(&provider)
	if !assert.Contains(t, feeder.clusterState.Pods(), pod1) {
		return
	}
	pod1State := feeder.clusterState.Pods()[pod1]
	if !assert.Contains(t, pod1State.Containers, containerCpu) {
		return
	}
	containerState := pod1State.Containers[containerCpu]
	if !assert.NotNil(t, containerState) {
		return
	}
	assert.Equal(t, t0, containerState.LastCPUSampleStart)
	if !assert.Contains(t, pod1State.Containers, containerMem) {
		return
	}
	containerState = pod1State.Containers[containerMem]
	if !assert.NotNil(t, containerState) {
		return
	}
	assert.Equal(t, memAmount, containerState.GetMaxMemoryPeak())
}

func TestFilterVPAs(t *testing.T) {
	recommenderName := "test-recommender"
	defaultRecommenderName := "default-recommender"

	vpa1 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: defaultRecommenderName},
			},
		},
	}
	vpa2 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: recommenderName},
			},
		},
	}
	vpa3 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: "another-recommender"},
			},
		},
	}

	allVpaCRDs := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2, vpa3}

	feeder := &clusterStateFeeder{
		recommenderName: recommenderName,
	}

	// Set expected results
	expectedResult := []*vpa_types.VerticalPodAutoscaler{vpa2}

	// Run the filterVPAs function
	result := filterVPAs(feeder, allVpaCRDs)

	if len(result) != len(expectedResult) {
		t.Fatalf("expected %d VPAs, got %d", len(expectedResult), len(result))
	}

	assert.ElementsMatch(t, expectedResult, result)
}

func TestFilterVPAsIgnoreNamespaces(t *testing.T) {

	vpa1 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace1",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa2 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace2",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa3 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ignore1",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa4 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ignore2",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}

	allVpaCRDs := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2, vpa3, vpa4}

	feeder := &clusterStateFeeder{
		recommenderName:   DefaultRecommenderName,
		ignoredNamespaces: []string{"ignore1", "ignore2"},
	}

	// Set expected results
	expectedResult := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2}

	// Run the filterVPAs function
	result := filterVPAs(feeder, allVpaCRDs)

	if len(result) != len(expectedResult) {
		t.Fatalf("expected %d VPAs, got %d", len(expectedResult), len(result))
	}

	assert.ElementsMatch(t, expectedResult, result)
}

func TestCanCleanupCheckpoints(t *testing.T) {
	client := fake.NewSimpleClientset()

	_, err := client.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "testNamespace"}}, metav1.CreateOptions{})
	assert.NoError(t, err)

	vpaBuilder := test.VerticalPodAutoscaler().WithContainer("container").WithNamespace("testNamespace").WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
		Kind:       kind,
		Name:       name1,
		APIVersion: apiVersion,
	})

	balanced := vpaBuilder.WithRecommender("balanced").WithName("balanced").Get()
	performance := vpaBuilder.WithRecommender("performance").WithName("performance").Get()
	savings := vpaBuilder.WithRecommender("savings").WithName("savings").Get()
	defaultVpa := vpaBuilder.WithRecommender("default").WithName("default").Get()

	vpas := []*vpa_types.VerticalPodAutoscaler{balanced, performance, savings, defaultVpa}
	vpaLister := &test.VerticalPodAutoscalerListerMock{}
	vpaLister.On("List").Return(vpas, nil)

	checkpoints := &vpa_types.VerticalPodAutoscalerCheckpointList{
		Items: []vpa_types.VerticalPodAutoscalerCheckpoint{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "testNamespace",
					Name:      "nonExistentVPA",
				},
				Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
					VPAObjectName: "nonExistentVPA",
				},
			},
		},
	}

	for _, vpa := range vpas {
		checkpoints.Items = append(checkpoints.Items, vpa_types.VerticalPodAutoscalerCheckpoint{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vpa.Namespace,
				Name:      vpa.Name,
			},
			Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
				VPAObjectName: vpa.Name,
			},
		})
	}

	checkpointClient := &fakeautoscalingv1.FakeAutoscalingV1{Fake: &core.Fake{}}
	checkpointClient.Fake.AddReactor("list", "verticalpodautoscalercheckpoints", func(action core.Action) (bool, runtime.Object, error) {
		return true, checkpoints, nil
	})

	deletedCheckpoints := []string{}
	checkpointClient.Fake.AddReactor("delete", "verticalpodautoscalercheckpoints", func(action core.Action) (bool, runtime.Object, error) {
		deleteAction := action.(core.DeleteAction)
		deletedCheckpoints = append(deletedCheckpoints, deleteAction.GetName())

		return true, nil, nil
	})

	feeder := clusterStateFeeder{
		coreClient:          client.CoreV1(),
		vpaLister:           vpaLister,
		vpaCheckpointClient: checkpointClient,
		clusterState:        model.NewClusterState(testGcPeriod),
		recommenderName:     "default",
	}

	feeder.GarbageCollectCheckpoints()

	assert.Contains(t, deletedCheckpoints, "nonExistentVPA")

	for _, vpa := range vpas {
		assert.NotContains(t, deletedCheckpoints, vpa.Name)
	}
}
