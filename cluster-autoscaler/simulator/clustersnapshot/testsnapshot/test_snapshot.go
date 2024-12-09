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

// NewTestSnapshot returns an instance of ClusterSnapshot that can be used in tests.
func NewTestSnapshot() (clustersnapshot.ClusterSnapshot, error) {
	testFwHandle, err := framework.NewTestFrameworkHandle()
	if err != nil {
		return nil, err
	}
	return predicate.NewPredicateSnapshot(store.NewBasicSnapshotStore(), testFwHandle, true), nil
}

// NewTestSnapshotOrDie returns an instance of ClusterSnapshot that can be used in tests.
func NewTestSnapshotOrDie(t testFailer) clustersnapshot.ClusterSnapshot {
	snapshot, err := NewTestSnapshot()
	if err != nil {
		t.Fatalf("NewTestSnapshotOrDie: couldn't create test ClusterSnapshot: %v", err)
	}
	return snapshot
}

// NewCustomTestSnapshot returns an instance of ClusterSnapshot with a specific ClusterSnapshotStore that can be used in tests.
func NewCustomTestSnapshot(snapshotStore clustersnapshot.ClusterSnapshotStore) (clustersnapshot.ClusterSnapshot, error) {
	testFwHandle, err := framework.NewTestFrameworkHandle()
	if err != nil {
		return nil, err
	}
	return predicate.NewPredicateSnapshot(snapshotStore, testFwHandle, true), nil
}

// NewCustomTestSnapshotOrDie returns an instance of ClusterSnapshot with a specific ClusterSnapshotStore that can be used in tests.
func NewCustomTestSnapshotOrDie(t testFailer, snapshotStore clustersnapshot.ClusterSnapshotStore) clustersnapshot.ClusterSnapshot {
	result, err := NewCustomTestSnapshot(snapshotStore)
	if err != nil {
		t.Fatalf("NewCustomTestSnapshotOrDie: couldn't create test ClusterSnapshot: %v", err)
	}
	return result
}
