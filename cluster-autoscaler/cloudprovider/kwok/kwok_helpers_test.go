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

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

const multipleNodes = `
apiVersion: v1
kind: Node
metadata:
  annotations:
    kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
    node.alpha.kubernetes.io/ttl: "0"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2023-05-31T04:39:16Z"
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: kind-control-plane
    kwok-nodegroup: control-plane
    kubernetes.io/os: linux
    node-role.kubernetes.io/control-plane: ""
    node.kubernetes.io/exclude-from-external-load-balancers: ""
  name: kind-control-plane
  resourceVersion: "603"
  uid: 86716ec7-3071-4091-b055-77b4361d1dca
spec:
  podCIDR: 10.244.0.0/24
  podCIDRs:
  - 10.244.0.0/24
  providerID: kind://docker/kind/kind-control-plane
  taints:
  - effect: NoSchedule
    key: node-role.kubernetes.io/control-plane
status:
  addresses:
  - address: 172.18.0.2
    type: InternalIP
  - address: kind-control-plane
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
  conditions:
  - lastHeartbeatTime: "2023-05-31T04:40:29Z"
    lastTransitionTime: "2023-05-31T04:39:13Z"
    message: kubelet has sufficient memory available
    reason: KubeletHasSufficientMemory
    status: "False"
    type: MemoryPressure
  - lastHeartbeatTime: "2023-05-31T04:40:29Z"
    lastTransitionTime: "2023-05-31T04:39:13Z"
    message: kubelet has no disk pressure
    reason: KubeletHasNoDiskPressure
    status: "False"
    type: DiskPressure
  - lastHeartbeatTime: "2023-05-31T04:40:29Z"
    lastTransitionTime: "2023-05-31T04:39:13Z"
    message: kubelet has sufficient PID available
    reason: KubeletHasSufficientPID
    status: "False"
    type: PIDPressure
  - lastHeartbeatTime: "2023-05-31T04:40:29Z"
    lastTransitionTime: "2023-05-31T04:39:46Z"
    message: kubelet is posting ready status
    reason: KubeletReady
    status: "True"
    type: Ready
  daemonEndpoints:
    kubeletEndpoint:
      Port: 10250
  images:
  - names:
    - registry.k8s.io/etcd:3.5.6-0
    sizeBytes: 102542580
  - names:
    - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
    - registry.k8s.io/kube-apiserver:v1.26.3
    sizeBytes: 80392681
  - names:
    - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
    - registry.k8s.io/kube-controller-manager:v1.26.3
    sizeBytes: 68538487
  - names:
    - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
    - registry.k8s.io/kube-proxy:v1.26.3
    sizeBytes: 67217404
  - names:
    - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
    - registry.k8s.io/kube-scheduler:v1.26.3
    sizeBytes: 57761399
  - names:
    - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    sizeBytes: 27726335
  - names:
    - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
    - docker.io/kindest/local-path-provisioner@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
    sizeBytes: 18664669
  - names:
    - registry.k8s.io/coredns/coredns:v1.9.3
    sizeBytes: 14837849
  - names:
    - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
    sizeBytes: 3052037
  - names:
    - registry.k8s.io/pause:3.7
    sizeBytes: 311278
  nodeInfo:
    architecture: amd64
    bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
    containerRuntimeVersion: containerd://1.6.19-46-g941215f49
    kernelVersion: 5.15.0-72-generic
    kubeProxyVersion: v1.26.3
    kubeletVersion: v1.26.3
    machineID: 96f8c8b8c8ae4600a3654341f207586e
    operatingSystem: linux
    osImage: Ubuntu
    systemUUID: 111aa932-7f99-4bef-aaf7-36aa7fb9b012
---

apiVersion: v1
kind: Node
metadata:
  annotations:
    kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
    node.alpha.kubernetes.io/ttl: "0"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2023-05-31T04:39:57Z"
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: kind-worker
    kwok-nodegroup: kind-worker
    kubernetes.io/os: linux
  name: kind-worker
  resourceVersion: "577"
  uid: 2ac0eb71-e5cf-4708-bbbf-476e8f19842b
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
  conditions:
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has sufficient memory available
    reason: KubeletHasSufficientMemory
    status: "False"
    type: MemoryPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has no disk pressure
    reason: KubeletHasNoDiskPressure
    status: "False"
    type: DiskPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has sufficient PID available
    reason: KubeletHasSufficientPID
    status: "False"
    type: PIDPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:40:05Z"
    message: kubelet is posting ready status
    reason: KubeletReady
    status: "True"
    type: Ready
  daemonEndpoints:
    kubeletEndpoint:
      Port: 10250
  images:
  - names:
    - registry.k8s.io/etcd:3.5.6-0
    sizeBytes: 102542580
  - names:
    - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
    - registry.k8s.io/kube-apiserver:v1.26.3
    sizeBytes: 80392681
  - names:
    - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
    - registry.k8s.io/kube-controller-manager:v1.26.3
    sizeBytes: 68538487
  - names:
    - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
    - registry.k8s.io/kube-proxy:v1.26.3
    sizeBytes: 67217404
  - names:
    - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
    - registry.k8s.io/kube-scheduler:v1.26.3
    sizeBytes: 57761399
  - names:
    - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    sizeBytes: 27726335
  - names:
    - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
    sizeBytes: 18664669
  - names:
    - registry.k8s.io/coredns/coredns:v1.9.3
    sizeBytes: 14837849
  - names:
    - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
    sizeBytes: 3052037
  - names:
    - registry.k8s.io/pause:3.7
    sizeBytes: 311278
  nodeInfo:
    architecture: amd64
    bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
    containerRuntimeVersion: containerd://1.6.19-46-g941215f49
    kernelVersion: 5.15.0-72-generic
    kubeProxyVersion: v1.26.3
    kubeletVersion: v1.26.3
    machineID: a98a13ff474d476294935341f1ba9816
    operatingSystem: linux
    osImage: Ubuntu
    systemUUID: 5f3c1af8-a385-4776-85e4-73d7f4252b44
`

const nodeList = `
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
      node.alpha.kubernetes.io/ttl: "0"
      volumes.kubernetes.io/controller-managed-attach-detach: "true"
    creationTimestamp: "2023-05-31T04:39:16Z"
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kind-control-plane
      kwok-nodegroup: control-plane
      kubernetes.io/os: linux
      node-role.kubernetes.io/control-plane: ""
      node.kubernetes.io/exclude-from-external-load-balancers: ""
    name: kind-control-plane
    resourceVersion: "506"
    uid: 86716ec7-3071-4091-b055-77b4361d1dca
  spec:
    podCIDR: 10.244.0.0/24
    podCIDRs:
    - 10.244.0.0/24
    providerID: kind://docker/kind/kind-control-plane
    taints:
    - effect: NoSchedule
      key: node-role.kubernetes.io/control-plane
  status:
    addresses:
    - address: 172.18.0.2
      type: InternalIP
    - address: kind-control-plane
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
    conditions:
    - lastHeartbeatTime: "2023-05-31T04:39:58Z"
      lastTransitionTime: "2023-05-31T04:39:13Z"
      message: kubelet has sufficient memory available
      reason: KubeletHasSufficientMemory
      status: "False"
      type: MemoryPressure
    - lastHeartbeatTime: "2023-05-31T04:39:58Z"
      lastTransitionTime: "2023-05-31T04:39:13Z"
      message: kubelet has no disk pressure
      reason: KubeletHasNoDiskPressure
      status: "False"
      type: DiskPressure
    - lastHeartbeatTime: "2023-05-31T04:39:58Z"
      lastTransitionTime: "2023-05-31T04:39:13Z"
      message: kubelet has sufficient PID available
      reason: KubeletHasSufficientPID
      status: "False"
      type: PIDPressure
    - lastHeartbeatTime: "2023-05-31T04:39:58Z"
      lastTransitionTime: "2023-05-31T04:39:46Z"
      message: kubelet is posting ready status
      reason: KubeletReady
      status: "True"
      type: Ready
    daemonEndpoints:
      kubeletEndpoint:
        Port: 10250
    images:
    - names:
      - registry.k8s.io/etcd:3.5.6-0
      sizeBytes: 102542580
    - names:
      - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
      - registry.k8s.io/kube-apiserver:v1.26.3
      sizeBytes: 80392681
    - names:
      - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
      - registry.k8s.io/kube-controller-manager:v1.26.3
      sizeBytes: 68538487
    - names:
      - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
      - registry.k8s.io/kube-proxy:v1.26.3
      sizeBytes: 67217404
    - names:
      - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
      - registry.k8s.io/kube-scheduler:v1.26.3
      sizeBytes: 57761399
    - names:
      - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      sizeBytes: 27726335
    - names:
      - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
      sizeBytes: 18664669
    - names:
      - registry.k8s.io/coredns/coredns:v1.9.3
      sizeBytes: 14837849
    - names:
      - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
      sizeBytes: 3052037
    - names:
      - registry.k8s.io/pause:3.7
      sizeBytes: 311278
    nodeInfo:
      architecture: amd64
      bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
      containerRuntimeVersion: containerd://1.6.19-46-g941215f49
      kernelVersion: 5.15.0-72-generic
      kubeProxyVersion: v1.26.3
      kubeletVersion: v1.26.3
      machineID: 96f8c8b8c8ae4600a3654341f207586e
      operatingSystem: linux
      osImage: Ubuntu
      systemUUID: 111aa932-7f99-4bef-aaf7-36aa7fb9b012
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
      node.alpha.kubernetes.io/ttl: "0"
      volumes.kubernetes.io/controller-managed-attach-detach: "true"
    creationTimestamp: "2023-05-31T04:39:57Z"
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kind-worker
      kwok-nodegroup: kind-worker
      kubernetes.io/os: linux
    name: kind-worker
    resourceVersion: "577"
    uid: 2ac0eb71-e5cf-4708-bbbf-476e8f19842b
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
    conditions:
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has sufficient memory available
      reason: KubeletHasSufficientMemory
      status: "False"
      type: MemoryPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has no disk pressure
      reason: KubeletHasNoDiskPressure
      status: "False"
      type: DiskPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has sufficient PID available
      reason: KubeletHasSufficientPID
      status: "False"
      type: PIDPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:40:05Z"
      message: kubelet is posting ready status
      reason: KubeletReady
      status: "True"
      type: Ready
    daemonEndpoints:
      kubeletEndpoint:
        Port: 10250
    images:
    - names:
      - registry.k8s.io/etcd:3.5.6-0
      sizeBytes: 102542580
    - names:
      - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
      - registry.k8s.io/kube-apiserver:v1.26.3
      sizeBytes: 80392681
    - names:
      - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
      - registry.k8s.io/kube-controller-manager:v1.26.3
      sizeBytes: 68538487
    - names:
      - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
      - registry.k8s.io/kube-proxy:v1.26.3
      sizeBytes: 67217404
    - names:
      - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
      - registry.k8s.io/kube-scheduler:v1.26.3
      sizeBytes: 57761399
    - names:
      - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      sizeBytes: 27726335
    - names:
      - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
      sizeBytes: 18664669
    - names:
      - registry.k8s.io/coredns/coredns:v1.9.3
      sizeBytes: 14837849
    - names:
      - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
      sizeBytes: 3052037
    - names:
      - registry.k8s.io/pause:3.7
      sizeBytes: 311278
    nodeInfo:
      architecture: amd64
      bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
      containerRuntimeVersion: containerd://1.6.19-46-g941215f49
      kernelVersion: 5.15.0-72-generic
      kubeProxyVersion: v1.26.3
      kubeletVersion: v1.26.3
      machineID: a98a13ff474d476294935341f1ba9816
      operatingSystem: linux
      osImage: Ubuntu
      systemUUID: 5f3c1af8-a385-4776-85e4-73d7f4252b44
kind: List
metadata:
  resourceVersion: ""
`

const wrongIndentation = `
apiVersion: v1
  items:
  - apiVersion: v1
# everything below should be in-line with apiVersion above
  kind: Node
metadata:
  annotations:
    kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
    node.alpha.kubernetes.io/ttl: "0"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2023-05-31T04:39:57Z"
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: kind-worker
    kwok-nodegroup: kind-worker
    kubernetes.io/os: linux
  name: kind-worker
  resourceVersion: "577"
  uid: 2ac0eb71-e5cf-4708-bbbf-476e8f19842b
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
  conditions:
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has sufficient memory available
    reason: KubeletHasSufficientMemory
    status: "False"
    type: MemoryPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has no disk pressure
    reason: KubeletHasNoDiskPressure
    status: "False"
    type: DiskPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:39:57Z"
    message: kubelet has sufficient PID available
    reason: KubeletHasSufficientPID
    status: "False"
    type: PIDPressure
  - lastHeartbeatTime: "2023-05-31T04:40:17Z"
    lastTransitionTime: "2023-05-31T04:40:05Z"
    message: kubelet is posting ready status
    reason: KubeletReady
    status: "True"
    type: Ready
  daemonEndpoints:
    kubeletEndpoint:
      Port: 10250
  images:
  - names:
    - registry.k8s.io/etcd:3.5.6-0
    sizeBytes: 102542580
  - names:
    - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
    - registry.k8s.io/kube-apiserver:v1.26.3
    sizeBytes: 80392681
  - names:
    - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
    - registry.k8s.io/kube-controller-manager:v1.26.3
    sizeBytes: 68538487
  - names:
    - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
    - registry.k8s.io/kube-proxy:v1.26.3
    sizeBytes: 67217404
  - names:
    - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
    - registry.k8s.io/kube-scheduler:v1.26.3
    sizeBytes: 57761399
  - names:
    - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
    sizeBytes: 27726335
  - names:
    - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
    sizeBytes: 18664669
  - names:
    - registry.k8s.io/coredns/coredns:v1.9.3
    sizeBytes: 14837849
  - names:
    - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
    sizeBytes: 3052037
  - names:
    - registry.k8s.io/pause:3.7
    sizeBytes: 311278
  nodeInfo:
    architecture: amd64
    bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
    containerRuntimeVersion: containerd://1.6.19-46-g941215f49
    kernelVersion: 5.15.0-72-generic
    kubeProxyVersion: v1.26.3
    kubeletVersion: v1.26.3
    machineID: a98a13ff474d476294935341f1ba9816
    operatingSystem: linux
    osImage: Ubuntu 22.04.2 LTS
    systemUUID: 5f3c1af8-a385-4776-85e4-73d7f4252b44
kind: List
metadata:
  resourceVersion: ""
`

const noGPULabel = `
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
      node.alpha.kubernetes.io/ttl: "0"
      volumes.kubernetes.io/controller-managed-attach-detach: "true"
    creationTimestamp: "2023-05-31T04:39:57Z"
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kind-worker
      kwok-nodegroup: kind-worker
      kubernetes.io/os: linux
    name: kind-worker
    resourceVersion: "577"
    uid: 2ac0eb71-e5cf-4708-bbbf-476e8f19842b
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
    conditions:
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has sufficient memory available
      reason: KubeletHasSufficientMemory
      status: "False"
      type: MemoryPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has no disk pressure
      reason: KubeletHasNoDiskPressure
      status: "False"
      type: DiskPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:39:57Z"
      message: kubelet has sufficient PID available
      reason: KubeletHasSufficientPID
      status: "False"
      type: PIDPressure
    - lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:40:05Z"
      message: kubelet is posting ready status
      reason: KubeletReady
      status: "True"
      type: Ready
    daemonEndpoints:
      kubeletEndpoint:
        Port: 10250
    images:
    - names:
      - registry.k8s.io/etcd:3.5.6-0
      sizeBytes: 102542580
    - names:
      - docker.io/library/import-2023-03-30@sha256:ba097b515c8c40689733c0f19de377e9bf8995964b7d7150c2045f3dfd166657
      - registry.k8s.io/kube-apiserver:v1.26.3
      sizeBytes: 80392681
    - names:
      - docker.io/library/import-2023-03-30@sha256:8dbb345de79d1c44f59a7895da702a5f71997ae72aea056609445c397b0c10dc
      - registry.k8s.io/kube-controller-manager:v1.26.3
      sizeBytes: 68538487
    - names:
      - docker.io/library/import-2023-03-30@sha256:44db4d50a5f9c8efbac0d37ea974d1c0419a5928f90748d3d491a041a00c20b5
      - registry.k8s.io/kube-proxy:v1.26.3
      sizeBytes: 67217404
    - names:
      - docker.io/library/import-2023-03-30@sha256:3dd2337f70af979c7362b5e52bbdfcb3a5fd39c78d94d02145150cd2db86ba39
      - registry.k8s.io/kube-scheduler:v1.26.3
      sizeBytes: 57761399
    - names:
      - docker.io/kindest/kindnetd:v20230330-48f316cd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      - docker.io/kindest/kindnetd@sha256:c19d6362a6a928139820761475a38c24c0cf84d507b9ddf414a078cf627497af
      sizeBytes: 27726335
    - names:
      - docker.io/kindest/local-path-provisioner:v0.0.23-kind.0@sha256:f2d0a02831ff3a03cf51343226670d5060623b43a4cfc4808bd0875b2c4b9501
      sizeBytes: 18664669
    - names:
      - registry.k8s.io/coredns/coredns:v1.9.3
      sizeBytes: 14837849
    - names:
      - docker.io/kindest/local-path-helper:v20230330-48f316cd@sha256:135203f2441f916fb13dad1561d27f60a6f11f50ec288b01a7d2ee9947c36270
      sizeBytes: 3052037
    - names:
      - registry.k8s.io/pause:3.7
      sizeBytes: 311278
    nodeInfo:
      architecture: amd64
      bootID: 2d71b318-5d07-4de2-9e61-2da28cf5bbf0
      containerRuntimeVersion: containerd://1.6.19-46-g941215f49
      kernelVersion: 5.15.0-72-generic
      kubeProxyVersion: v1.26.3
      kubeletVersion: v1.26.3
      machineID: a98a13ff474d476294935341f1ba9816
      operatingSystem: linux
      osImage: Ubuntu 22.04.2 LTS
      systemUUID: 5f3c1af8-a385-4776-85e4-73d7f4252b44
kind: List
metadata:
  resourceVersion: ""
`

func TestLoadNodeTemplatesFromConfigMap(t *testing.T) {
	var testTemplatesMap = map[string]string{
		"wrongIndentation":         wrongIndentation,
		defaultTemplatesConfigName: testTemplates,
		"multipleNodes":            multipleNodes,
		"nodeList":                 nodeList,
	}

	testTemplateName := defaultTemplatesConfigName

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		if getAction.GetName() == defaultConfigName {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					configKey: testConfig,
				},
			}, nil
		}

		if testTemplatesMap[testTemplateName] != "" {
			return true, &apiv1.ConfigMap{
				Data: map[string]string{
					templatesKey: testTemplatesMap[testTemplateName],
				},
			}, nil
		}

		return true, nil, errors.NewNotFound(apiv1.Resource("configmaps"), "whatever")
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	os.Setenv("POD_NAMESPACE", "kube-system")

	kwokConfig, err := LoadConfigFile(fakeClient)
	assert.Nil(t, err)

	// happy path
	testTemplateName = defaultTemplatesConfigName
	nos, err := LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, fakeClient)
	assert.Nil(t, err)
	assert.NotEmpty(t, nos)
	assert.Greater(t, len(nos), 0)

	testTemplateName = "wrongIndentation"
	nos, err = LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, fakeClient)
	assert.Error(t, err)
	assert.Empty(t, nos)
	assert.Equal(t, len(nos), 0)

	// multiple nodes is something like []*Node{node1, node2, node3, ...}
	testTemplateName = "multipleNodes"
	nos, err = LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, fakeClient)
	assert.Nil(t, err)
	assert.NotEmpty(t, nos)
	assert.Greater(t, len(nos), 0)

	// node list is something like []*List{Items:[]*Node{node1, node2, node3, ...}}
	testTemplateName = "nodeList"
	nos, err = LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, fakeClient)
	assert.Nil(t, err)
	assert.NotEmpty(t, nos)
	assert.Greater(t, len(nos), 0)

	// fake client which returns configmap with wrong key
	fakeClient = &fake.Clientset{}
	fakeClient.Fake.AddReactor("get", "configmaps", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.ConfigMap{
			Data: map[string]string{
				"foo": testTemplatesMap[testTemplateName],
			},
		}, nil
	})

	fakeClient.Fake.AddReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)

		if getAction == nil {
			return false, nil, nil
		}

		return true, &apiv1.NodeList{Items: []apiv1.Node{}}, errors.NewNotFound(apiv1.Resource("nodes"), "whatever")
	})

	// throw error if configmap data key is not `templates`
	nos, err = LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, fakeClient)
	assert.Error(t, err)
	assert.Empty(t, nos)
	assert.Equal(t, len(nos), 0)
}
