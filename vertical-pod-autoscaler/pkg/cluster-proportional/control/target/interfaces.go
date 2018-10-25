package target

import "k8s.io/api/core/v1"

type Interface interface {
	// ReadClusterState gets the current state of the cluster (summary statistics)
	ReadClusterState() (*ClusterStats, error)
}

type ClusterStats struct {
	NodeCount          int
	NodeSumAllocatable v1.ResourceList
}
