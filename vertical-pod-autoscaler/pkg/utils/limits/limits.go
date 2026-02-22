package limits

import apiv1 "k8s.io/api/core/v1"

// GlobalMaxAllowed holds global maximum caps for recommendations.
// The VPA-level maxAllowed takes precedence.
type GlobalMaxAllowed struct {
	// Container-level maximums apply to per-Container recommendations.
	Container apiv1.ResourceList
	// Pod-level maximums apply to the Pod-level recommendations
	Pod apiv1.ResourceList
}

// GlobalMinAllowed holds global minimum caps for recommendations.
// The VPA-level minAllowed takes precedence.
type GlobalMinAllowed struct {
	// Pod-level minimums apply to the Pod-level recommendations
	Pod apiv1.ResourceList
}
