/*
Copyright 2021 The Kubernetes Authors.

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

package debuggingsnapshot

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

func TestBasicSnapshotRequest(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	snapshotter := NewDebuggingSnapshotter(true)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Pod1",
		},
		Spec: v1.PodSpec{
			NodeName: "testNode",
		},
	}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testNode",
		},
	}
	nodeInfo := framework.NewTestNodeInfo(node, pod)

	var nodeGroups []*framework.NodeInfo
	nodeGroups = append(nodeGroups, nodeInfo)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	go func() {
		snapshotter.ResponseHandler(w, req)
		wg.Done()
	}()

	for !snapshotter.IsDataCollectionAllowed() {
		snapshotter.StartDataCollection()
	}
	snapshotter.SetClusterNodes(nodeGroups)
	snapshotter.Flush()

	wg.Wait()
	resp := w.Result()
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, int64(0), resp.ContentLength)
}

func TestFlushWithoutData(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	snapshotter := NewDebuggingSnapshotter(true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	go func() {
		snapshotter.ResponseHandler(w, req)
		wg.Done()
	}()

	for !snapshotter.IsDataCollectionAllowed() {
		snapshotter.StartDataCollection()
	}
	snapshotter.Flush()

	wg.Wait()
	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Greater(t, int64(0), resp.ContentLength)
}

func TestRequestTerminationOnShutdown(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	snapshotter := NewDebuggingSnapshotter(true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	go func() {
		snapshotter.ResponseHandler(w, req)
		wg.Done()
	}()

	for !snapshotter.IsDataCollectionAllowed() {
		snapshotter.StartDataCollection()
	}

	go snapshotter.Cleanup()
	wg.Wait()

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestRejectParallelRequest(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	snapshotter := NewDebuggingSnapshotter(true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	go func() {
		snapshotter.ResponseHandler(w, req)
		wg.Done()
	}()

	for !snapshotter.IsDataCollectionAllowed() {
		snapshotter.StartDataCollection()
	}

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	snapshotter.ResponseHandler(w1, req1)
	assert.Equal(t, http.StatusTooManyRequests, w1.Code)

	snapshotter.SetClusterNodes(nil)
	snapshotter.Flush()
	wg.Wait()

	assert.Equal(t, http.StatusOK, w.Code)
}
