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
	"github.com/google/go-cmp/cmp"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
)

// CloneTestSnapshot creates a deep copy of the provided Snapshot.
// This function is intended for testing purposes only.
func CloneTestSnapshot(snapshot *Snapshot) *Snapshot {
	cloneString := func(s string) string { return s }
	cloneCSINode := func(cn *storagev1.CSINode) *storagev1.CSINode { return cn.DeepCopy() }

	csiNodes := common.ClonePatchSet(snapshot.csiNodes, cloneString, cloneCSINode)

	return &Snapshot{
		csiNodes: csiNodes,
	}
}

// snapshotFlattenedComparer returns a cmp.Option that provides a custom comparer function
// for comparing two *Snapshot objects based on their underlying data maps, it doesn't
// compare the underlying patchsets, instead flattened objects are compared.
// This function is intended for testing purposes only.
func snapshotFlattenedComparer() cmp.Option {
	return cmp.Comparer(func(a, b *Snapshot) bool {
		if a == nil || b == nil {
			return a == b
		}

		csiNodesEqual := cmp.Equal(a.csiNodes.AsMap(), b.csiNodes.AsMap())

		return csiNodesEqual
	})
}
