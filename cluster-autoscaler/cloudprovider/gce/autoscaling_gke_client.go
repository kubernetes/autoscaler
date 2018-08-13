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

package gce

import (
	"flag"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// This flag is outside main as it's only useful for test/development.
var (
	GkeAPIEndpoint = flag.String("gke-api-endpoint", "", "GKE API endpoint address. This flag is used by developers only. Users shouldn't change this flag.")
)

const (
	clusterPathPrefix   = "projects/%s/locations/%s/clusters/%s"
	nodePoolsPathPrefix = "projects/%s/locations/%s/clusters/%s/nodePools/"
	nodePoolPathPrefix  = "projects/%s/locations/%s/clusters/%s/nodePools/%%s"
	operationPathPrefix = "projects/%s/locations/%s/operations/%%s"
)

// AutoscalingGkeClient is used for communicating with GKE API.
type AutoscalingGkeClient interface {
	// reading cluster state
	FetchLocations() ([]string, error)
	FetchNodePools() ([]NodePool, error)
	FetchResourceLimits() (*cloudprovider.ResourceLimiter, error)

	// modifying cluster state
	DeleteNodePool(string) error
	CreateNodePool(Mig) error
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
