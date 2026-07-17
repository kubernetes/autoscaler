/*
Copyright 2025 The Kubernetes Authors.

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

package snapshot

import (
	storagev1 "k8s.io/api/storage/v1"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

// SnapshotCSINodeLister provides access to CSI nodes within a snapshot.
type SnapshotCSINodeLister struct {
	snapshot *Snapshot
}

var _ schedulerinterface.CSINodeLister = SnapshotCSINodeLister{}

// List returns all CSI nodes in the snapshot.
func (s SnapshotCSINodeLister) List() ([]*storagev1.CSINode, error) {
	return s.snapshot.listCSINodes(), nil
}

// Get retrieves a CSI node by name from the snapshot.
func (s SnapshotCSINodeLister) Get(name string) (*storagev1.CSINode, error) {
	return s.snapshot.Get(name)
}
