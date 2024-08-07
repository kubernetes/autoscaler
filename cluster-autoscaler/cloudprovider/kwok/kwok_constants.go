/*
Copyright 2023 The Kubernetes Authors.

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

package kwok

const (
	// ProviderName is the cloud provider name for kwok
	ProviderName = "kwok"

	//NGNameAnnotation is the annotation kwok provider uses to track the nodegroups
	NGNameAnnotation = "cluster-autoscaler.kwok.nodegroup/name"
	// NGMinSizeAnnotation is annotation on template nodes which specify min size of the nodegroup
	NGMinSizeAnnotation = "cluster-autoscaler.kwok.nodegroup/min-count"
	// NGMaxSizeAnnotation is annotation on template nodes which specify max size of the nodegroup
	NGMaxSizeAnnotation = "cluster-autoscaler.kwok.nodegroup/max-count"
	// NGDesiredSizeAnnotation is annotation on template nodes which specify desired size of the nodegroup
	NGDesiredSizeAnnotation = "cluster-autoscaler.kwok.nodegroup/desired-count"

	// KwokManagedAnnotation is the default annotation
	// that kwok manages to decide if it should manage
	// a node it sees in the cluster
	KwokManagedAnnotation = "kwok.x-k8s.io/node"

	groupNodesByAnnotation = "annotation"
	groupNodesByLabel      = "label"

	// // GPULabel is the label added to nodes with GPU resource.
	// GPULabel = "cloud.google.com/gke-accelerator"

	// for kwok provider config
	nodeTemplatesFromConfigMap = "configmap"
	nodeTemplatesFromCluster   = "cluster"
)

const testTemplates = `
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    annotations: {}
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kind-worker
      kwok-nodegroup: kind-worker
      kubernetes.io/os: linux
      k8s.amazonaws.com/accelerator: "nvidia-tesla-k80"
    name: kind-worker
  spec:
    podCIDR: 10.244.2.0/24
    podCIDRs:
    - 10.244.2.0/24
    providerID: kind://docker/kind/kind-worker
  status:
    addresses:
    - address: 172.18.0.3
      type: InternalIP
    - address: kind-worker
      type: Hostname
    allocatable:
      cpu: "12"
      ephemeral-storage: 959786032Ki
      hugepages-1Gi: "0"
      hugepages-2Mi: "0"
      memory: 32781516Ki
      pods: "110"
    capacity:
      cpu: "12"
      ephemeral-storage: 959786032Ki
      hugepages-1Gi: "0"
      hugepages-2Mi: "0"
      memory: 32781516Ki
      pods: "110"
- apiVersion: v1
  kind: Node
  metadata:
    annotations: {}
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kind-worker-2
      kubernetes.io/os: linux
      k8s.amazonaws.com/accelerator: "nvidia-tesla-k80"
    name: kind-worker-2
  spec:
    podCIDR: 10.244.2.0/24
    podCIDRs:
    - 10.244.2.0/24
    providerID: kind://docker/kind/kind-worker-2
  status:
    addresses:
    - address: 172.18.0.3
      type: InternalIP
    - address: kind-worker-2
      type: Hostname
    allocatable:
      cpu: "12"
      ephemeral-storage: 959786032Ki
      hugepages-1Gi: "0"
      hugepages-2Mi: "0"
      memory: 32781516Ki
      pods: "110"
    capacity:
      cpu: "12"
      ephemeral-storage: 959786032Ki
      hugepages-1Gi: "0"
      hugepages-2Mi: "0"
      memory: 32781516Ki
      pods: "110"
kind: List
metadata:
  resourceVersion: ""
`

// yaml version of fakeNode1, fakeNode2 and fakeNode3
const testTemplatesMinimal = `
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      cluster-autoscaler.kwok.nodegroup/name: ng1
    labels:
      kwok-nodegroup: ng1
    name: node1
  spec: {}
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      cluster-autoscaler.kwok.nodegroup/name: ng2
    labels:
      kwok-nodegroup: ng2
    name: node2
  spec: {}
- apiVersion: v1
  kind: Node
  metadata:
    annotations: {}
    labels: {}
    name: node3
  spec: {}
kind: List
metadata:
  resourceVersion: ""
`
