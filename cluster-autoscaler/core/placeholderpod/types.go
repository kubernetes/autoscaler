package placeholderpod

import (
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// Tolerations is a slice of k8s tolerations
type Tolerations []apiv1.Toleration

// Spec is a specification of a placeholder pod
type Spec struct {
	// MilliCPU is the amount of CPU cpu requested for the virtual primary container of a placeholder pod
	MilliCPU int64
	// Memory is the amount of memory requested for the virtual primary container of a placeholder pod
	Memory int64
	// NodeStickiness represents what kind of nodes this placeholder pod wants to be scheduled on
	NodeStickiness NodeStickiness
}

// NodeStickiness represents what kind of nodes this placeholder pod wants to be scheduled on
type NodeStickiness struct {
	// NodeSelector is the contents of the `spec.nodeSelector` field of a placeholder pod
	NodeSelector map[string]string
	// PodAntiAffinityRequiredTerms is the contents of the `spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution` field of a placeholder pod
	PodAntiAffinityRequiredTerms []apiv1.PodAffinityTerm
	// NodeAffinityRequiredTerms is the contents of the `spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution` field of a placeholder pod
	NodeAffinityRequiredTerms *apiv1.NodeSelector
}

// ReplicaSet is a definition for a set of equivalent placeholder pods
type ReplicaSet struct {
	// Count is the number of replicas to be created
	Count int64
	// Name is the name of this set which will also be included in the names of placeholder pods
	Name string
	// PodSpec is the spec of placeholder pods created by this command
	PodSpec Spec
}

// New creates a new placeholder specification
func New(milliCPU int64, memory int64, nodeStickiness NodeStickiness) Spec {
	return Spec{
		MilliCPU:       milliCPU,
		Memory:         memory,
		NodeStickiness: nodeStickiness,
	}
}
