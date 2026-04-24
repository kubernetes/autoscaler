/*
Copyright 2018 The Kubernetes Authors.

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

package oom

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
}

// podYamlHeader is the common metadata/spec prefix for all pod fixtures in
// this file. Every YAML below appends only the containerStatus fields that
// distinguish the scenario.
const podYamlHeader = `
apiVersion: v1
kind: Pod
metadata:
  name: Pod1
  namespace: mockNamespace
spec:
  containers:
  - name: Name11
    resources:
      requests:
        memory: "1024"
status:
  containerStatuses:
  - name: Name11
`

const (
	runningPodYaml = podYamlHeader

	oomPodYaml = podYamlHeader + `
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

	runningBeforeOOMRestartYaml = podYamlHeader + `
    restartCount: 0
    state:
      running:
        startedAt: 2018-02-23T13:00:00Z
`

	runningAfterOOMRestartYaml = podYamlHeader + `
    restartCount: 1
    state:
      running:
        startedAt: 2018-02-23T13:38:50Z
    lastState:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

	runningAfterNonOOMRestartYaml = podYamlHeader + `
    restartCount: 1
    state:
      running:
        startedAt: 2018-02-23T13:38:50Z
    lastState:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: Error
`

	terminatedBeforeFastRestartYaml = podYamlHeader + `
    restartCount: 0
    state:
      terminated:
        finishedAt: 2018-02-23T13:00:00Z
        reason: Error
`

	terminatedAfterFastRestartYaml = podYamlHeader + `
    restartCount: 1
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

	terminatedNonOOMYaml = podYamlHeader + `
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: Error
`

	terminatedNoReasonYaml = podYamlHeader + `
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
`

	waitingCrashLoopOOMYaml = podYamlHeader + `
    restartCount: 3
    state:
      waiting:
        reason: CrashLoopBackOff
    lastState:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

	runningAfterCrashLoopYaml = podYamlHeader + `
    restartCount: 4
    state:
      running:
        startedAt: 2018-02-23T13:40:00Z
    lastState:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`
)

func mustNewPod(t *testing.T, yaml string) *corev1.Pod {
	t.Helper()
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		t.Fatalf("failed to parse pod YAML: %v", err)
	}
	return obj.(*corev1.Pod)
}

func newEvent(yaml string) (*corev1.Event, error) {
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*corev1.Event), nil
}

func TestOOMObserverOnUpdate(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2018-02-23T13:38:48Z")
	assert.NoError(t, err)

	containerID := model.ContainerID{
		ContainerName: "Name11",
		PodID: model.PodID{
			Namespace: "mockNamespace",
			PodName:   "Pod1",
		},
	}
	wantOOM := func(memory int64) *OomInfo {
		return &OomInfo{
			ContainerID: containerID,
			Memory:      model.ResourceAmount(memory),
			Timestamp:   timestamp,
		}
	}
	setContainerStatusMemory := func(quantity string) func(*corev1.Pod) {
		return func(pod *corev1.Pod) {
			pod.Status.ContainerStatuses[0].Resources = &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse(quantity),
				},
			}
		}
	}

	testCases := []struct {
		desc        string
		oldPodYaml  string
		newPodYaml  string
		mutateOld   func(*corev1.Pod)
		mutateNew   func(*corev1.Pod)
		wantOOMInfo *OomInfo // nil => expect no OOM event
	}{
		{
			desc:        "Running -> Terminated OOMKilled records a new OOM",
			oldPodYaml:  runningPodYaml,
			newPodYaml:  oomPodYaml,
			wantOOMInfo: wantOOM(1024),
		},
		{
			desc:       "isNewOOM reads zero memory when new pod has no requests",
			oldPodYaml: runningPodYaml,
			newPodYaml: oomPodYaml,
			mutateNew: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Resources.Requests = nil
				pod.Status.ContainerStatuses[0].Resources = nil
			},
			wantOOMInfo: wantOOM(0),
		},
		{
			desc:        "isNewOOM prefers containerStatus.resources over spec.resources",
			oldPodYaml:  runningPodYaml,
			newPodYaml:  oomPodYaml,
			mutateNew:   setContainerStatusMemory("2048"),
			wantOOMInfo: wantOOM(2048),
		},
		{
			desc:        "Running -> Running with OOM lastState records a previous OOM",
			oldPodYaml:  runningBeforeOOMRestartYaml,
			newPodYaml:  runningAfterOOMRestartYaml,
			wantOOMInfo: wantOOM(1024),
		},
		{
			desc:        "isPreviousOOM reads resources from oldPod, not newPod",
			oldPodYaml:  runningBeforeOOMRestartYaml,
			newPodYaml:  runningAfterOOMRestartYaml,
			mutateOld:   setContainerStatusMemory("2048"),
			mutateNew:   setContainerStatusMemory("4096"),
			wantOOMInfo: wantOOM(2048),
		},
		{
			desc:        "Terminated non-OOM -> Terminated OOMKilled with restart records a new OOM",
			oldPodYaml:  terminatedBeforeFastRestartYaml,
			newPodYaml:  terminatedAfterFastRestartYaml,
			wantOOMInfo: wantOOM(1024),
		},
		{
			desc:       "Running -> Terminated with non-OOM reason is ignored",
			oldPodYaml: runningPodYaml,
			newPodYaml: terminatedNonOOMYaml,
		},
		{
			desc:       "Running -> Running with non-OOM lastState is ignored",
			oldPodYaml: runningPodYaml,
			newPodYaml: runningAfterNonOOMRestartYaml,
		},
		{
			desc:       "Waiting(CrashLoopBackOff) -> Running with OOM lastState is not double-counted",
			oldPodYaml: waitingCrashLoopOOMYaml,
			newPodYaml: runningAfterCrashLoopYaml,
		},
		{
			desc:       "Informer relist emitting identical Running state is not double-counted",
			oldPodYaml: runningAfterOOMRestartYaml,
			newPodYaml: runningAfterOOMRestartYaml,
		},
		{
			desc:       "Terminated -> Terminated OOMKilled without restart is ignored",
			oldPodYaml: terminatedNoReasonYaml,
			newPodYaml: oomPodYaml,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			oldPod := mustNewPod(t, tc.oldPodYaml)
			newPod := mustNewPod(t, tc.newPodYaml)
			if tc.mutateOld != nil {
				tc.mutateOld(oldPod)
			}
			if tc.mutateNew != nil {
				tc.mutateNew(newPod)
			}
			observer := NewObserver()
			observer.OnUpdate(oldPod, newPod)
			if tc.wantOOMInfo == nil {
				assert.Empty(t, observer.observedOomsChannel)
				return
			}
			info := <-observer.observedOomsChannel
			assert.Equal(t, *tc.wantOOMInfo, info)
		})
	}
}

func TestParseEvictionEvent(t *testing.T) {
	parseTimestamp := func(str string) time.Time {
		timestamp, err := time.Parse(time.RFC3339, "2018-02-23T13:38:48Z")
		assert.NoError(t, err)
		return timestamp.UTC()
	}
	parseResources := func(str string) model.ResourceAmount {
		memory, err := resource.ParseQuantity(str)
		assert.NoError(t, err)
		return model.ResourceAmount(memory.Value())
	}

	toContainerID := func(namespace, pod, container string) model.ContainerID {
		return model.ContainerID{
			PodID: model.PodID{
				PodName:   pod,
				Namespace: namespace,
			},
			ContainerName: container,
		}
	}

	testCases := []struct {
		event   string
		oomInfo []OomInfo
	}{
		{
			event: `
apiVersion: v1
kind: Event
metadata:
  annotations:
    offending_containers: test-container
    offending_containers_usage: 1024Ki
    starved_resource: memory
  creationTimestamp: 2018-02-23T13:38:48Z
involvedObject:
  apiVersion: v1
  kind: Pod
  name: pod1
  namespace: test-namespace
reason: Evicted
`,
			oomInfo: []OomInfo{
				{
					Timestamp:   parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:      parseResources("1024Ki"),
					ContainerID: toContainerID("test-namespace", "pod1", "test-container"),
				},
			},
		},
		{
			event: `
apiVersion: v1
kind: Event
metadata:
  annotations:
    offending_containers: test-container,other-container
    offending_containers_usage: 1024Ki,2048Ki
    starved_resource: memory,memory
  creationTimestamp: 2018-02-23T13:38:48Z
involvedObject:
  apiVersion: v1
  kind: Pod
  name: pod1
  namespace: test-namespace
reason: Evicted
`,
			oomInfo: []OomInfo{
				{
					Timestamp:   parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:      parseResources("1024Ki"),
					ContainerID: toContainerID("test-namespace", "pod1", "test-container"),
				},
				{
					Timestamp:   parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:      parseResources("2048Ki"),
					ContainerID: toContainerID("test-namespace", "pod1", "other-container"),
				},
			},
		},
		{
			event: `
apiVersion: v1
kind: Event
metadata:
  annotations:
    offending_containers: test-container,other-container
    offending_containers_usage: 1024Ki,2048Ki
    starved_resource: memory,evictable                       # invalid resource skipped
  creationTimestamp: 2018-02-23T13:38:48Z
involvedObject:
  apiVersion: v1
  kind: Pod
  name: pod1
  namespace: test-namespace
reason: Evicted
`,
			oomInfo: []OomInfo{
				{
					Timestamp:   parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:      parseResources("1024Ki"),
					ContainerID: toContainerID("test-namespace", "pod1", "test-container"),
				},
			},
		},
		{
			event: `
apiVersion: v1
kind: Event
metadata:
  annotations:
    offending_containers: test-container,other-container
    offending_containers_usage: 1024Ki,2048Ki
    starved_resource: memory                              # missing resource invalids all event
  creationTimestamp: 2018-02-23T13:38:48Z
involvedObject:
  apiVersion: v1
  kind: Pod
  name: pod1
  namespace: test-namespace
reason: Evicted
`,
			oomInfo: []OomInfo{},
		},
	}

	for _, tc := range testCases {
		event, err := newEvent(tc.event)
		assert.NoError(t, err)
		assert.NotNil(t, event)

		oomInfoArray := parseEvictionEvent(event)
		assert.Equal(t, tc.oomInfo, oomInfoArray)
	}
}
