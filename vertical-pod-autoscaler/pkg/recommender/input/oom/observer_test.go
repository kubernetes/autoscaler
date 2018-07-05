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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	_ "k8s.io/kubernetes/pkg/apis/core/install"       //to decode yaml
	_ "k8s.io/kubernetes/pkg/apis/extensions/install" //to decode yaml
)

const pod1Yaml = `
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
    restartCount: 0
`

const pod2Yaml = `
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
    restartCount: 1
    lastState:
      terminated:
        finishedAt: 2018-02-23T13:38:48Z
        reason: OOMKilled
`

func newPod(yaml string) (*v1.Pod, error) {
	decode := legacyscheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*v1.Pod), nil
}

func newEvent(yaml string) (*v1.Event, error) {
	decode := legacyscheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*v1.Event), nil
}

func TestOOMReceived(t *testing.T) {
	p1, err := newPod(pod1Yaml)
	assert.NoError(t, err)
	p2, err := newPod(pod2Yaml)
	assert.NoError(t, err)
	observer := NewObserver()
	go observer.OnUpdate(p1, p2)

	info := <-observer.ObservedOomsChannel
	assert.Equal(t, "mockNamespace", info.Namespace)
	assert.Equal(t, "Pod1", info.Pod)
	assert.Equal(t, "Name11", info.Container)
	assert.Equal(t, int64(1024), info.Memory.Value())
	timestamp, err := time.Parse(time.RFC3339, "2018-02-23T13:38:48Z")
	assert.NoError(t, err)
	assert.Equal(t, timestamp.Unix(), info.Timestamp.Unix())
}

func TestParseEvictionEvent(t *testing.T) {
	parseTimestamp := func(str string) time.Time {
		timestamp, err := time.Parse(time.RFC3339, "2018-02-23T13:38:48Z")
		assert.NoError(t, err)
		return timestamp.UTC()
	}
	parseResources := func(str string) resource.Quantity {
		memory, err := resource.ParseQuantity(str)
		assert.NoError(t, err)
		return memory
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
					Timestamp: parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:    parseResources("1024Ki"),
					Pod:       "pod1",
					Container: "test-container",
					Namespace: "test-namespace",
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
					Timestamp: parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:    parseResources("1024Ki"),
					Pod:       "pod1",
					Container: "test-container",
					Namespace: "test-namespace",
				},
				{
					Timestamp: parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:    parseResources("2048Ki"),
					Pod:       "pod1",
					Container: "other-container",
					Namespace: "test-namespace",
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
					Timestamp: parseTimestamp("2018-02-23T13:38:48Z "),
					Memory:    parseResources("1024Ki"),
					Pod:       "pod1",
					Container: "test-container",
					Namespace: "test-namespace",
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
