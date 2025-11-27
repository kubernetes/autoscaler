package snapshot

import storagev1 "k8s.io/api/storage/v1"

// SnapshotCSINodeLister provides access to CSI nodes within a snapshot.
type SnapshotCSINodeLister struct {
	snapshot *Snapshot
}

// List returns all CSI nodes in the snapshot.
func (s SnapshotCSINodeLister) List() ([]*storagev1.CSINode, error) {
	return s.snapshot.listCSINodes(), nil
}

// Get retrieves a CSI node by name from the snapshot.
func (s SnapshotCSINodeLister) Get(name string) (*storagev1.CSINode, error) {
	return s.snapshot.Get(name)
}
