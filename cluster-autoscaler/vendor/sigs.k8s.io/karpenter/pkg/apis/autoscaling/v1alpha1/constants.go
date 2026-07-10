/*
Copyright The Kubernetes Authors.

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

package v1alpha1

import "math"

// Constants shared across the CapacityBuffer controller, the provisioner, and
// any downstream consumer (disruption, metrics). Mirrors upstream Cluster
// Autoscaler constants at
// k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/constants.go
const (
	ActiveProvisioningStrategy = "buffer.x-k8s.io/active-capacity"

	// Condition types written to CapacityBuffer status.
	ReadyForProvisioningCondition = "ReadyForProvisioning"
	ProvisioningCondition         = "Provisioning"
	LimitedByQuotasCondition      = "LimitedByQuotas"

	// Supported scalableRef kinds.
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindReplicaSet  = "ReplicaSet"

	// FakePodAnnotationKey marks a virtual pod constructed from a CapacityBuffer.
	FakePodAnnotationKey   = "karpenter.sh/capacity-buffer-fake-pod"
	FakePodAnnotationValue = "true"

	// BufferNameLabel records which CapacityBuffer a virtual pod belongs to.
	BufferNameLabel = "karpenter.sh/capacity-buffer-name"

	// BufferNamespaceLabel records the namespace of the CapacityBuffer a virtual pod belongs to.
	BufferNamespaceLabel = "karpenter.sh/capacity-buffer-namespace"

	// VirtualPodPriority is the priority stamped onto virtual buffer pods so that
	// future preemption / disruption logic can identify them as low-value.
	// NOTE: Karpenter's scheduler currently sorts the queue by resource size,
	// not priority, so this value does not affect scheduling order today.
	VirtualPodPriority int32 = math.MinInt32
)
