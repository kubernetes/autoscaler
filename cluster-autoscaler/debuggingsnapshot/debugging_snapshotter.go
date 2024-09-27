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
	"context"
	"net/http"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// DebuggingSnapshotterState is the type for the debugging snapshot State machine
// The states guide the workflow of the snapshot.
type DebuggingSnapshotterState int

// DebuggingSnapshotterState help navigate the different workflows of the snapshot capture.
const (
	// SNAPSHOTTER_DISABLED is when debuggingSnapshot is disabled on the cluster and no action can be taken
	SNAPSHOTTER_DISABLED DebuggingSnapshotterState = iota + 1
	// LISTENING is set when snapshotter is enabled on the cluster and is ready to listen to a
	// snapshot request. Used by ResponseHandler to wait on to listen to request
	LISTENING
	// TRIGGER_ENABLED is set by ResponseHandler if a valid snapshot request is received
	// it states that a snapshot request needs to be processed
	TRIGGER_ENABLED
	// START_DATA_COLLECTION is used to synchronise the collection of data.
	// Since the trigger is an asynchronous process, data collection could be started mid-loop
	// leading to incomplete data. So setter methods wait for START_DATA_COLLECTION before collecting data
	// which is set at the start of the next loop after receiving the trigger
	START_DATA_COLLECTION
	// DATA_COLLECTED is set by setter func (also used by setter func for data collection)
	// This is set to let Flush know that at least some data collected and there isn't
	// an error State leading to no data collection
	DATA_COLLECTED
)

// DebuggingSnapshotterImpl is the impl for DebuggingSnapshotter
type DebuggingSnapshotterImpl struct {
	// State captures the internal state of the snapshotter
	State *DebuggingSnapshotterState
	// DebuggingSnapshot is the data bean for the snapshot
	DebuggingSnapshot DebuggingSnapshot
	// Mutex is the synchronisation used to the methods/states in the critical section
	Mutex *sync.Mutex
	// Trigger is the channel on which the Response Handler waits on to know
	// when there is data to be flushed back to the channel from the Snapshot
	Trigger chan struct{}
	// CancelRequest is the cancel function for the snapshot request. It is used to
	// terminate any ongoing request when CA is shutting down
	CancelRequest context.CancelFunc
}

// DebuggingSnapshotter is the interface for debugging snapshot
type DebuggingSnapshotter interface {

	// StartDataCollection will check the State(s) and enable data
	// collection for the loop if applicable
	StartDataCollection()
	// SetClusterNodes is a setter to capture all the ClusterNode
	SetClusterNodes([]*framework.NodeInfo)
	// SetUnscheduledPodsCanBeScheduled is a setter for all pods which are unscheduled
	// but they can be scheduled. i.e. pods which aren't triggering scale-up
	SetUnscheduledPodsCanBeScheduled([]*v1.Pod)
	// SetTemplateNodes is a setter for all the TemplateNodes present in the cluster
	// incl. templates for which there are no nodes
	SetTemplateNodes(map[string]*framework.NodeInfo)
	// ResponseHandler is the http response handler to manage incoming requests
	ResponseHandler(http.ResponseWriter, *http.Request)
	// IsDataCollectionAllowed checks the internal State of the snapshotter
	// to find if data can be collected. This can be used before preprocessing
	// for the snapshot
	IsDataCollectionAllowed() bool
	// Flush triggers the flushing of the snapshot
	Flush()
	// Cleanup clears the internal data beans of the snapshot, readying for next request
	Cleanup()
}

// NewDebuggingSnapshotter returns a new instance of DebuggingSnapshotter
func NewDebuggingSnapshotter(isDebuggerEnabled bool) DebuggingSnapshotter {
	state := SNAPSHOTTER_DISABLED
	if isDebuggerEnabled {
		klog.Infof("Debugging Snapshot is enabled")
		state = LISTENING
	}
	return &DebuggingSnapshotterImpl{
		State:             &state,
		Mutex:             &sync.Mutex{},
		DebuggingSnapshot: &DebuggingSnapshotImpl{},
		Trigger:           make(chan struct{}, 1),
	}
}

// ResponseHandler is the impl for request handler
func (d *DebuggingSnapshotterImpl) ResponseHandler(w http.ResponseWriter, r *http.Request) {

	d.Mutex.Lock()
	// checks if the handler is in the correct State to accept a new snapshot request
	if *d.State != LISTENING {
		defer d.Mutex.Unlock()
		klog.Errorf("Debugging Snapshot is currently being processed. Another snapshot can't be processed")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Another debugging snapshot request is being processed. Concurrent requests not supported"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	d.CancelRequest = cancel

	klog.Infof("Received a new snapshot, that is accepted")
	// set the State to trigger enabled, to allow workflow to collect data
	*d.State = TRIGGER_ENABLED
	d.Mutex.Unlock()

	select {
	case <-d.Trigger:
		d.Mutex.Lock()
		d.DebuggingSnapshot.SetEndTimestamp(time.Now().In(time.UTC))
		body, isErrorMessage := d.DebuggingSnapshot.GetOutputBytes()
		if isErrorMessage {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write(body)

		// reset the debugging State to receive a new snapshot request
		*d.State = LISTENING
		d.CancelRequest = nil
		d.DebuggingSnapshot.Cleanup()

		d.Mutex.Unlock()
	case <-ctx.Done():
		d.Mutex.Lock()
		klog.Infof("Received terminate trigger, aborting ongoing snapshot request")
		w.WriteHeader(http.StatusServiceUnavailable)

		d.DebuggingSnapshot.Cleanup()
		*d.State = LISTENING
		d.CancelRequest = nil
		select {
		case <-d.Trigger:
		default:
		}
		d.Mutex.Unlock()
	}
}

// IsDataCollectionAllowed encapsulate the check to know if data collection is currently active
// This should be used by setters and by any function that is contingent on data collection State
// before doing extra processing.
// e.g. If you want to pre-process a particular State in cloud-provider for snapshot
// you should check this func in the loop before doing that extra processing
func (d *DebuggingSnapshotterImpl) IsDataCollectionAllowed() bool {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	return d.IsDataCollectionAllowedNoLock()
}

// IsDataCollectionAllowedNoLock encapsulated the check to know if data collection is currently active
// The need for NoLock implementation is for cases when the caller funcs have procured the lock
// for a single transactional execution
func (d *DebuggingSnapshotterImpl) IsDataCollectionAllowedNoLock() bool {
	return *d.State == DATA_COLLECTED || *d.State == START_DATA_COLLECTION
}

// StartDataCollection changes the State when the trigger has been enabled
// to start data collection. To be done at the start of the runLoop to allow for consistency
// as the trigger can be called mid-loop leading to partial data collection
func (d *DebuggingSnapshotterImpl) StartDataCollection() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if *d.State == TRIGGER_ENABLED {
		*d.State = START_DATA_COLLECTION
		klog.Infof("Trigger Enabled for Debugging Snapshot, starting data collection")
		d.DebuggingSnapshot.SetStartTimestamp(time.Now().In(time.UTC))
	}
}

// Flush is the impl for DebuggingSnapshotter.Flush
// It checks if any data has been collected or data collection failed
func (d *DebuggingSnapshotterImpl) Flush() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	// Case where Data Collection was started but no data was collected, needs to
	// be stated as an error and reset to pre-trigger State
	if *d.State == START_DATA_COLLECTION {
		klog.Errorf("No data was collected for the snapshot in this loop. So no snapshot can be generated.")
		d.DebuggingSnapshot.SetErrorMessage("Unable to collect any data")
		d.Trigger <- struct{}{}
		return
	}

	if *d.State == DATA_COLLECTED {
		d.Trigger <- struct{}{}
	}
}

// SetClusterNodes is the setter for Node Group Info
// All filtering/prettifying of data should be done here.
func (d *DebuggingSnapshotterImpl) SetClusterNodes(nodeInfos []*framework.NodeInfo) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if !d.IsDataCollectionAllowedNoLock() {
		return
	}
	klog.V(4).Infof("NodeGroupInfo is being set for the debugging snapshot")
	d.DebuggingSnapshot.SetClusterNodes(nodeInfos)
	*d.State = DATA_COLLECTED
}

// SetUnscheduledPodsCanBeScheduled is the setter for UnscheduledPodsCanBeScheduled
func (d *DebuggingSnapshotterImpl) SetUnscheduledPodsCanBeScheduled(podList []*v1.Pod) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if !d.IsDataCollectionAllowedNoLock() {
		return
	}
	klog.V(4).Infof("UnscheduledPodsCanBeScheduled is being set for the debugging snapshot")
	d.DebuggingSnapshot.SetUnscheduledPodsCanBeScheduled(podList)
	*d.State = DATA_COLLECTED
}

// SetTemplateNodes is the setter for TemplateNodes
func (d *DebuggingSnapshotterImpl) SetTemplateNodes(templates map[string]*framework.NodeInfo) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if !d.IsDataCollectionAllowedNoLock() {
		return
	}
	klog.V(4).Infof("TemplateNodes is being set for the debugging snapshot")
	d.DebuggingSnapshot.SetTemplateNodes(templates)
}

// Cleanup clears the internal data sets of the cluster
func (d *DebuggingSnapshotterImpl) Cleanup() {
	if d.CancelRequest != nil {
		d.CancelRequest()
	}
}
