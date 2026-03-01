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

const runningPodYaml = `
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

const oomPodYaml = `
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
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

func newPod(yaml string) (*corev1.Pod, error) {
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*corev1.Pod), nil
}

func newEvent(yaml string) (*corev1.Event, error) {
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*corev1.Event), nil
}

func TestOOMReceived(t *testing.T) {
	p1, err := newPod(runningPodYaml)
	assert.NoError(t, err)
	p2, err := newPod(oomPodYaml)
	assert.NoError(t, err)
	timestamp, err := time.Parse(time.RFC3339, "2018-02-23T13:38:48Z")
	assert.NoError(t, err)

	testCases := []struct {
		desc        string
		oldPod      *corev1.Pod
		newPod      *corev1.Pod
		wantOOMInfo OomInfo
	}{
		{
			desc:   "OK",
			oldPod: p1,
			newPod: p2,
			wantOOMInfo: OomInfo{
				ContainerID: model.ContainerID{
					ContainerName: "Name11",
					PodID: model.PodID{
						Namespace: "mockNamespace",
						PodName:   "Pod1",
					},
				},
				Memory:    model.ResourceAmount(int64(1024)),
				Timestamp: timestamp,
			},
		},
		{
			desc: "Old pod does not set memory requests",
			oldPod: func() *corev1.Pod {
				oldPod := p1.DeepCopy()
				oldPod.Spec.Containers[0].Resources.Requests = nil
				oldPod.Status.ContainerStatuses[0].Resources = nil
				return oldPod
			}(),
			newPod: p2,
			wantOOMInfo: OomInfo{
				ContainerID: model.ContainerID{
					ContainerName: "Name11",
					PodID: model.PodID{
						Namespace: "mockNamespace",
						PodName:   "Pod1",
					},
				},
				Memory:    model.ResourceAmount(int64(0)),
				Timestamp: timestamp,
			},
		},
		{
			desc: "Old pod also set memory request in containerStatus, prefer info from containerStatus",
			oldPod: func() *corev1.Pod {
				oldPod := p1.DeepCopy()
				oldPod.Status.ContainerStatuses[0].Resources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2048"),
					},
				}
				return oldPod
			}(),
			newPod: p2,
			wantOOMInfo: OomInfo{
				ContainerID: model.ContainerID{
					ContainerName: "Name11",
					PodID: model.PodID{
						Namespace: "mockNamespace",
						PodName:   "Pod1",
					},
				},
				Memory:    model.ResourceAmount(int64(2048)),
				Timestamp: timestamp,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			observer := NewObserver()
			observer.OnUpdate(tc.oldPod, tc.newPod)
			info := <-observer.observedOomsChannel
			assert.Equal(t, tc.wantOOMInfo, info)
		})
	}
}

func TestOOMStateAfterTerminatedState(t *testing.T) {
	p1, err := newPod(`
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
    state:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
`)
	assert.NoError(t, err)
	p2, err := newPod(oomPodYaml)
	assert.NoError(t, err)
	observer := NewObserver()
	observer.OnUpdate(p1, p2)

	// No OOM event should be sent if previous state was also "terminated".
	assert.Empty(t, observer.observedOomsChannel)
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
