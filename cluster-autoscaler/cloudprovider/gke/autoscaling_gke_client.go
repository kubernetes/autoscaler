/*
Copyright 2018 The Kubernetes Authors.

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

package gke

import (
	"flag"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

var (
	// GkeAPIEndpoint overrides default GKE API endpoint for testing.
	// This flag is outside main as it's only useful for test/development.
	GkeAPIEndpoint = flag.String("gke-api-endpoint", "", "GKE API endpoint address. This flag is used by developers only. Users shouldn't change this flag.")
)

const (
	defaultOperationWaitTimeout  = 120 * time.Second
	defaultOperationPollInterval = 1 * time.Second
)

const (
	clusterPathPrefix   = "projects/%s/locations/%s/clusters/%s"
	nodePoolPathPrefix  = "projects/%s/locations/%s/clusters/%s/nodePools/%%s"
	operationPathPrefix = "projects/%s/locations/%s/operations/%%s"
)

// AutoscalingGkeClient is used for communicating with GKE API.
type AutoscalingGkeClient interface {
	// reading cluster state
	GetCluster() (Cluster, error)

	// modifying cluster state
	DeleteNodePool(string) error
	CreateNodePool(*GkeMig) error
}

// Cluster contains cluster's fields we want to use.
type Cluster struct {
	Locations       []string
	NodePools       []NodePool
	ResourceLimiter *cloudprovider.ResourceLimiter
}

// NodePool contains node pool's fields we want to use.
type NodePool struct {
	Name              string
	InstanceGroupUrls []string
	Autoscaled        bool
	MinNodeCount      int64
	MaxNodeCount      int64
	Autoprovisioned   bool
}
