/*
Copyright 2024 The Kubernetes Authors.

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

package testsnapshot

import (
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// testFailer is an abstraction that covers both *testing.T and *testing.B.
type testFailer interface {
	Fatalf(format string, args ...any)
}

// NewTestSnapshotOrDie returns an instance of ClusterSnapshot that can be used in tests.
func NewTestSnapshotOrDie(t testFailer) clustersnapshot.ClusterSnapshot {
	snapshot, _, err := NewTestSnapshotAndHandle()
	if err != nil {
		t.Fatalf("NewTestSnapshotOrDie: couldn't create test ClusterSnapshot: %v", err)
	}
	return snapshot
}

// NewCustomTestSnapshotOrDie returns an instance of ClusterSnapshot with a specific ClusterSnapshotStore that can be used in tests.
func NewCustomTestSnapshotOrDie(t testFailer, snapshotStore clustersnapshot.ClusterSnapshotStore) clustersnapshot.ClusterSnapshot {
	result, _, err := NewCustomTestSnapshotAndHandle(snapshotStore)
	if err != nil {
		t.Fatalf("NewCustomTestSnapshotOrDie: couldn't create test ClusterSnapshot: %v", err)
	}
	return result
}

// NewTestSnapshotAndHandle returns an instance of ClusterSnapshot and a framework handle that can be used in tests.
func NewTestSnapshotAndHandle() (clustersnapshot.ClusterSnapshot, *framework.Handle, error) {
	return NewCustomTestSnapshotAndHandle(store.NewBasicSnapshotStore())
}

// NewCustomTestSnapshotAndHandle returns an instance of ClusterSnapshot with a specific ClusterSnapshotStore that can be used in tests.
func NewCustomTestSnapshotAndHandle(snapshotStore clustersnapshot.ClusterSnapshotStore) (clustersnapshot.ClusterSnapshot, *framework.Handle, error) {
	testFwHandle, err := framework.NewTestFrameworkHandle()
	if err != nil {
		return nil, nil, err
	}
	return predicate.NewPredicateSnapshot(snapshotStore, testFwHandle, true), testFwHandle, nil
}
