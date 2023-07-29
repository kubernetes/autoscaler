package v1alpha1

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=mpa
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.updatePolicy.updateMode"
// +kubebuilder:printcolumn:name="CPU",type="string",JSONPath=".status.recommendation.containerRecommendations[0].target.cpu"
// +kubebuilder:printcolumn:name="Mem",type="string",JSONPath=".status.recommendation.containerRecommendations[0].target.memory"
// +kubebuilder:printcolumn:name="Provided",type="string",JSONPath=".status.conditions[?(@.type=='RecommendationProvided')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MultidimPodAutoscaler is the configuration for a multidimensional pod autoscaler,
// which automatically manages pod resources and number of replicas based on historical and
// real-time resource utilization as well as workload performance.
type MultidimPodAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	Spec MultidimPodAutoscalerSpec `json:"spec"`

	// Current information about the autoscaler.
	// +optional
	Status MultidimPodAutoscalerStatus `json:"status,omitempty"`
}

// MultidimPodAutoscalerSpec is the specification of the behavior of the autoscaler.
type MultidimPodAutoscalerSpec struct {
	// ScaleTargetRef points to the controller managing the set of pods for the autoscaler to
	// control, e.g., Deployment, StatefulSet. MultidimPodAutoscaler can be targeted at controller
	// implementing scale subresource (the pod set is retrieved from the controller's ScaleStatus
	// or some well known controllers (e.g., for DaemonSet the pod set is read from the
	// controller's spec). If MultidimPodAutoscaler cannot use specified target it will report
	// the ConfigUnsupported condition.
	ScaleTargetRef *autoscaling.CrossVersionObjectReference `json:"scaleTargetRef"`

	// Describes the rules on how changes are applied to the pods.
	// If not specified, all fields in the `PodUpdatePolicy` are set to their default values.
	// +optional
	UpdatePolicy *PodUpdatePolicy `json:"updatePolicy,omitempty"`

	// Contains the specifications about the metric type and target in terms of resource
	// utilization or workload performance. See the individual metric source types for
	// more information about how each type of metric must respond.
	// +listType=atomic
	// +optional
	Metrics []autoscalingv2.MetricSpec `json:"metrics,omitempty"`

	// Describes the constraints for the number of replicas.
	Constraints *HorizontalScalingConstraints `json:"constraints,omitempty"`
	// Controls how the VPA autoscaler computes recommended resources.
	// The resource policy is also used to set constraints on the recommendations for individual
	// containers. If not specified, the autoscaler computes recommended resources for all
	// containers in the pod, without additional constraints.
	// +optional
	ResourcePolicy *vpa.PodResourcePolicy `json:"resourcePolicy,omitempty"`

	// Recommender responsible for generating recommendation for the set of pods and the deployment.
	// List should be empty (then the default recommender will be used) or contain exactly one
	// recommender.
	// +optional
	Recommenders []*MultidimPodAutoscalerRecommenderSelector `json:"recommenders,omitempty"`
}

// Describes the current status of a multidimensional pod autoscaler
type MultidimPodAutoscalerStatus struct {
	// Last time the MultidimPodAutoscaler scaled the number of pods and resizes containers;
	// Used by the autoscaler to control how often scaling operations are performed.
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	// Current number of replicas of pods managed by this autoscaler.
	CurrentReplicas int32 `json:"currentReplicas"`

	// Desired number of replicas of pods managed by this autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas"`

	// The most recently computed amount of resources for each controlled pod recommended by the
	// autoscaler.
	// +optional
	Recommendation *vpa.RecommendedPodResources `json:"recommendation,omitempty"`

	// The last read state of the metrics used by this autoscaler.
	// +listType=atomic
	// +optional
	CurrentMetrics []autoscalingv2.MetricStatus `json:"currentMetrics"`

	// Conditions is the set of conditions required for this autoscaler to scale its target, and
	// indicates whether or not those conditions are met.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MultidimPodAutoscalerCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// PodUpdatePolicy describes the rules on how changes are applied to the pods.
type PodUpdatePolicy struct {
	// Controls when autoscaler applies changes to the pod resources.
	// The default is 'Auto'.
	// +optional
	UpdateMode *vpa.UpdateMode `json:"updateMode,omitempty"`
}

// HorizontalScalingConstraints describes the constraints for horizontal scaling.
type HorizontalScalingConstraints struct {
	// Lower limit for the number of pods that can be set by the autoscaler, default 1.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// Upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than
	// MinReplicas.
	MaxReplicas *int32 `json:"maxReplicas"`
	// Behavior configures the scaling behavior of the target in both Up and Down direction
	// (scaleUp and scaleDown fields respectively).
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// MultidimPodAutoscalerRecommenderSelector points to a specific Multidimensional Pod Autoscaler
// recommender.
// In the future it might pass parameters to the recommender.
type MultidimPodAutoscalerRecommenderSelector struct {
	// Name of the recommender responsible for generating recommendation for this object.
	Name string `json:"name"`
}

// MultidimPodAutoscalerCondition describes the state of a MultidimPodAutoscaler at a certain point.
type MultidimPodAutoscalerCondition struct {
	// type describes the current condition
	Type MultidimPodAutoscalerConditionType `json:"type"`
	// status is the status of the condition (True, False, Unknown)
	Status v1.ConditionStatus `json:"status"`
	// lastTransitionTime is the last time the condition transitioned from one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// reason is the reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// message is a human-readable explanation containing details about the transition
	// +optional
	Message string `json:"message,omitempty"`
}

// MultidimPodAutoscalerConditionType are the valid conditions of a MultidimPodAutoscaler.
type MultidimPodAutoscalerConditionType string

var (
	// RecommendationProvided indicates whether the MPA recommender was able to give a
	// recommendation.
	RecommendationProvided MultidimPodAutoscalerConditionType = "RecommendationProvided"
	// LowConfidence indicates whether the MPA recommender has low confidence in the recommendation
	// for some of containers.
	LowConfidence MultidimPodAutoscalerConditionType = "LowConfidence"
	// NoPodsMatched indicates that label selector used with MPA object didn't match any pods.
	NoPodsMatched MultidimPodAutoscalerConditionType = "NoPodsMatched"
	// FetchingHistory indicates that MPA recommender is in the process of loading additional
	// history samples.
	FetchingHistory MultidimPodAutoscalerConditionType = "FetchingHistory"
	// ConfigDeprecated indicates that this MPA configuration is deprecated and will stop being
	// supported soon.
	ConfigDeprecated MultidimPodAutoscalerConditionType = "ConfigDeprecated"
	// ConfigUnsupported indicates that this MPA configuration is unsupported and recommendations
	// will not be provided for it.
	ConfigUnsupported MultidimPodAutoscalerConditionType = "ConfigUnsupported"
	// ScalingActive indicates that the MPA controller is able to scale if necessary, i.e.,
	// it is correctly configured, can fetch the desired metrics, and isn't disabled.
	ScalingActive MultidimPodAutoscalerConditionType = "ScalingActive"
	// AbleToScale indicates a lack of transient issues which prevent scaling from occurring,
	// such as being in a backoff window, or being unable to access/update the target scale.
	AbleToScale MultidimPodAutoscalerConditionType = "AbleToScale"
	// ScalingLimited indicates that the calculated scale based on metrics would be above or
	// below the range for the MPA, and has thus been capped.
	ScalingLimited MultidimPodAutoscalerConditionType = "ScalingLimited"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultidimPodAutoscalerList is a list of MultidimPodAutoscaler objects.
type MultidimPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata"`

	// items is the list of Multidimensional Pod Autoscaler objects.
	Items []MultidimPodAutoscaler `json:"items"`
}
