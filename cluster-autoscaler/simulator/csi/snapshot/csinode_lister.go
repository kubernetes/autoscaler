package snapshot

import storagev1 "k8s.io/api/storage/v1"

type SnapshotCSINodeLister struct {
	snapshot *Snapshot
}

func (s SnapshotCSINodeLister) List() ([]*storagev1.CSINode, error) {
	return s.snapshot.listCSINodes(), nil
}

func (s SnapshotCSINodeLister) Get(name string) (*storagev1.CSINode, error) {
	return s.snapshot.Get(name)
}
