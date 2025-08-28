/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package latencytracker

import (
	"sync"
	"testing"
	"time"
)

func TestUpdateStateWithUnneededList_AddsNewNodes(t *testing.T) {
	tracker := NewNodeLatencyTracker()
	now := time.Now()
	node := NodeInfo{Name: "node1", UnneededSince: now, Threshold: 5 * time.Minute}

	tracker.UpdateStateWithUnneededList([]NodeInfo{node}, now)

	tracker.Lock()
	defer tracker.Unlock()
	if _, ok := tracker.nodes["node1"]; !ok {
		t.Errorf("expected node1 to be tracked, but was not")
	}
}

func TestUpdateStateWithUnneededList_DoesNotDuplicate(t *testing.T) {
	tracker := NewNodeLatencyTracker()
	now := time.Now()
	node := NodeInfo{Name: "node1", UnneededSince: now, Threshold: 5 * time.Minute}

	tracker.UpdateStateWithUnneededList([]NodeInfo{node}, now)
	tracker.UpdateStateWithUnneededList([]NodeInfo{node}, now.Add(time.Minute))

	tracker.Lock()
	defer tracker.Unlock()
	if len(tracker.nodes) != 1 {
		t.Errorf("expected 1 tracked node, got %d", len(tracker.nodes))
	}
}

func TestObserveDeletion_RemovesNode(t *testing.T) {
	tracker := NewNodeLatencyTracker()
	now := time.Now()
	node := NodeInfo{
		Name:          "node1",
		UnneededSince: now.Add(-10 * time.Minute),
		Threshold:     5 * time.Minute,
	}
	tracker.UpdateStateWithUnneededList([]NodeInfo{node}, now)

	tracker.ObserveDeletion("node1", now)

	tracker.Lock()
	defer tracker.Unlock()
	if _, ok := tracker.nodes["node1"]; ok {
		t.Errorf("expected node1 removed after ObserveDeletion")
	}
}

func TestObserveDeletion_NoOpIfNodeNotTracked(t *testing.T) {
	tracker := NewNodeLatencyTracker()
	now := time.Now()

	tracker.ObserveDeletion("node1", now)

	tracker.Lock()
	defer tracker.Unlock()
	if len(tracker.nodes) != 0 {
		t.Errorf("expected no nodes tracked, got %d", len(tracker.nodes))
	}
}

func TestConcurrentUpdatesAndDeletions(t *testing.T) {
	tracker := NewNodeLatencyTracker()
	now := time.Now()

	node := NodeInfo{
		Name:          "node1",
		UnneededSince: now,
		Threshold:     2 * time.Minute,
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				tracker.UpdateStateWithUnneededList([]NodeInfo{node}, time.Now())
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				tracker.ObserveDeletion("node1", time.Now())
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()

	tracker.Lock()
	defer tracker.Unlock()
	if len(tracker.nodes) > 1 {
		t.Errorf("expected at most 1 tracked node, got %d", len(tracker.nodes))
	}
}
