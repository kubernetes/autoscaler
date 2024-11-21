package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Annotations that are labeled into the ReplicaSets that are part of an experiment
const (
	ExperimentNameAnnotationKey         = "experiment.argoproj.io/name"
	ExperimentTemplateNameAnnotationKey = "experiment.argoproj.io/template-name"
)

// Experiment is a specification for an Experiment resource
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:path=experiments,shortName=exp
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Experiment status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time since resource was created"
type Experiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ExperimentSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status ExperimentStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ExperimentSpec is the spec for a Experiment resource
type ExperimentSpec struct {
	// Templates are a list of PodSpecs that define the ReplicaSets that should be run during an experiment.
	// +patchMergeKey=name
	// +patchStrategy=merge
	Templates []TemplateSpec `json:"templates" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,1,rep,name=templates"`
	// Duration the amount of time for the experiment to run as a duration string (e.g. 30s, 5m, 1h).
	// If omitted, the experiment will run indefinitely, stopped either via termination, or a failed analysis run.
	// +optional
	Duration DurationString `json:"duration,omitempty" protobuf:"bytes,2,opt,name=duration,casttype=DurationString"`
	// ProgressDeadlineSeconds The maximum time in seconds for a experiment to
	// make progress before it is considered to be failed. Argo Rollouts will
	// continue to process failed experiments and a condition with a
	// ProgressDeadlineExceeded reason will be surfaced in the experiment status.
	// Defaults to 600s.
	// +optional
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,3,opt,name=progressDeadlineSeconds"`
	// Terminate is used to prematurely stop the experiment
	Terminate bool `json:"terminate,omitempty" protobuf:"varint,4,opt,name=terminate"`
	// Analyses references AnalysisTemplates to run during the experiment
	// +patchMergeKey=name
	// +patchStrategy=merge
	Analyses []ExperimentAnalysisTemplateRef `json:"analyses,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,5,rep,name=analyses"`
	// ScaleDownDelaySeconds adds a delay before scaling down the Experiment.
	// If omitted, the Experiment waits 30 seconds before scaling down.
	// A minimum of 30 seconds is recommended to ensure IP table propagation across the nodes in
	// a cluster. See https://github.com/argoproj/argo-rollouts/issues/19#issuecomment-476329960 for
	// more information
	// +optional
	ScaleDownDelaySeconds *int32 `json:"scaleDownDelaySeconds,omitempty" protobuf:"varint,6,opt,name=scaleDownDelaySeconds"`
	// DryRun object contains the settings for running the analysis in Dry-Run mode
	// +patchMergeKey=metricName
	// +patchStrategy=merge
	// +optional
	DryRun []DryRun `json:"dryRun,omitempty" patchStrategy:"merge" patchMergeKey:"metricName" protobuf:"bytes,7,rep,name=dryRun"`
	// MeasurementRetention object contains the settings for retaining the number of measurements during the analysis
	// +patchMergeKey=metricName
	// +patchStrategy=merge
	// +optional
	MeasurementRetention []MeasurementRetention `json:"measurementRetention,omitempty" patchStrategy:"merge" patchMergeKey:"metricName" protobuf:"bytes,8,rep,name=measurementRetention"`
}

type TemplateSpec struct {
	// Name of the template used to identity replicaset running for this experiment
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`
	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,3,opt,name=minReadySeconds"`
	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this experiment.
	// It must match the pod template's labels. Each selector must be unique to the other selectors in the other templates
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,4,opt,name=selector"`
	// Template describes the pods that will be created.
	Template corev1.PodTemplateSpec `json:"template" protobuf:"bytes,5,opt,name=template"`
	// TemplateService describes how a service should be generated for template
	Service *TemplateService `json:"service,omitempty" protobuf:"bytes,6,opt,name=service"`
}

type TemplateService struct {
	// Name of the service generated by the experiment
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

type TemplateStatusCode string

const (
	TemplateStatusProgressing TemplateStatusCode = "Progressing"
	TemplateStatusRunning     TemplateStatusCode = "Running"
	TemplateStatusSuccessful  TemplateStatusCode = "Successful"
	TemplateStatusFailed      TemplateStatusCode = "Failed"
	TemplateStatusError       TemplateStatusCode = "Error"
)

func (ts TemplateStatusCode) Completed() bool {
	switch ts {
	case TemplateStatusSuccessful, TemplateStatusFailed, TemplateStatusError:
		return true
	}
	return false
}

// TemplateStatus is the status of a specific template of an Experiment
type TemplateStatus struct {
	// Name of the template used to identity which hash to compare to the hash
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Total number of non-terminated pods targeted by this experiment (their labels match the selector).
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`
	// Total number of non-terminated pods targeted by this experiment that have the desired template spec.
	UpdatedReplicas int32 `json:"updatedReplicas" protobuf:"varint,3,opt,name=updatedReplicas"`
	// Total number of ready pods targeted by this experiment.
	ReadyReplicas int32 `json:"readyReplicas" protobuf:"varint,4,opt,name=readyReplicas"`
	// Total number of available pods (ready for at least minReadySeconds) targeted by this experiment.
	AvailableReplicas int32 `json:"availableReplicas" protobuf:"varint,5,opt,name=availableReplicas"`
	// CollisionCount count of hash collisions for the Experiment. The Experiment controller uses this
	// field as a collision avoidance mechanism when it needs to create the name for the
	// newest ReplicaSet.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,6,opt,name=collisionCount"`
	// Phase is the status of the ReplicaSet associated with the template
	Status TemplateStatusCode `json:"status,omitempty" protobuf:"bytes,7,opt,name=status,casttype=TemplateStatusCode"`
	// Message is a message explaining the current status
	Message string `json:"message,omitempty" protobuf:"bytes,8,opt,name=message"`
	// LastTransitionTime is the last time the replicaset transitioned, which resets the countdown
	// on the ProgressDeadlineSeconds check.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,9,opt,name=lastTransitionTime"`
	// ServiceName is the name of the service which corresponds to this experiment
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,10,opt,name=serviceName"`
	// PodTemplateHash is the value of the Replicas' PodTemplateHash
	PodTemplateHash string `json:"podTemplateHash,omitempty" protobuf:"bytes,11,opt,name=podTemplateHash"`
}

// ExperimentStatus is the status for a Experiment resource
type ExperimentStatus struct {
	// Phase is the status of the experiment. Takes into consideration ReplicaSet degradations and
	// AnalysisRun statuses
	Phase AnalysisPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=AnalysisPhase"`
	// Message is an explanation for the current status
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// TemplateStatuses holds the ReplicaSet related statuses for individual templates
	// +optional
	TemplateStatuses []TemplateStatus `json:"templateStatuses,omitempty" protobuf:"bytes,3,rep,name=templateStatuses"`
	// AvailableAt the time when all the templates become healthy and the experiment should start tracking the time to
	// run for the duration of specificed in the spec.
	// +optional
	AvailableAt *metav1.Time `json:"availableAt,omitempty" protobuf:"bytes,4,opt,name=availableAt"`
	// Conditions a list of conditions a experiment can have.
	// +optional
	Conditions []ExperimentCondition `json:"conditions,omitempty" protobuf:"bytes,5,rep,name=conditions"`
	// AnalysisRuns tracks the status of AnalysisRuns associated with this Experiment
	// +optional
	AnalysisRuns []ExperimentAnalysisRunStatus `json:"analysisRuns,omitempty" protobuf:"bytes,6,rep,name=analysisRuns"`
}

// ExperimentConditionType defines the conditions of Experiment
type ExperimentConditionType string

// These are valid conditions of a experiment.
const (
	// InvalidExperimentSpec means the experiment has an invalid spec and will not progress until
	// the spec is fixed.
	InvalidExperimentSpec ExperimentConditionType = "InvalidSpec"
	// ExperimentCompleted means the experiment is available, ie. the active service is pointing at a
	// replicaset with the required replicas up and running for at least minReadySeconds.
	ExperimentCompleted ExperimentConditionType = "Completed"
	// ExperimentProgressing means the experiment is progressing. Progress for a experiment is
	// considered when a new replica set is created or adopted, when pods scale
	// up or old pods scale down, or when the services are updated. Progress is not estimated
	// for paused experiment.
	ExperimentProgressing ExperimentConditionType = "Progressing"
	// ExperimentRunning means that an experiment has reached the desired state and is running for the duration
	// specified in the spec
	ExperimentRunning ExperimentConditionType = "Running"
	// ExperimentReplicaFailure ReplicaFailure is added in a experiment when one of its pods
	// fails to be created or deleted.
	ExperimentReplicaFailure ExperimentConditionType = "ReplicaFailure"
)

// ExperimentCondition describes the state of a experiment at a certain point.
type ExperimentCondition struct {
	// Type of deployment condition.
	Type ExperimentConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ExperimentConditionType"`
	// Phase of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime" protobuf:"bytes,3,opt,name=lastUpdateTime"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message" protobuf:"bytes,6,opt,name=message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExperimentList is a list of Experiment resources
type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	Items []Experiment `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type ExperimentAnalysisTemplateRef struct {
	// Name is the name of the analysis
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// TemplateName reference of the AnalysisTemplate name used by the Experiment to create the run
	TemplateName string `json:"templateName" protobuf:"bytes,2,opt,name=templateName"`
	// Whether to look for the templateName at cluster scope or namespace scope
	// +optional
	ClusterScope bool `json:"clusterScope,omitempty" protobuf:"varint,3,opt,name=clusterScope"`
	// Args are the arguments that will be added to the AnalysisRuns
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Args []Argument `json:"args,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,4,rep,name=args"`
	// RequiredForCompletion blocks the Experiment from completing until the analysis has completed
	RequiredForCompletion bool `json:"requiredForCompletion,omitempty" protobuf:"varint,5,opt,name=requiredForCompletion"`
}

type ExperimentAnalysisRunStatus struct {
	// Name is the name of the analysis
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// AnalysisRun is the name of the AnalysisRun
	AnalysisRun string `json:"analysisRun" protobuf:"bytes,2,opt,name=analysisRun"`
	// Phase is the status of the AnalysisRun
	Phase AnalysisPhase `json:"phase" protobuf:"bytes,3,opt,name=phase,casttype=AnalysisPhase"`
	// Message is a message explaining the current status
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}
