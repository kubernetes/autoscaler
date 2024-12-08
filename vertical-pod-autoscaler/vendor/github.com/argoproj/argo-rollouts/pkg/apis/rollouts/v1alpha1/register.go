package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rollouts "github.com/argoproj/argo-rollouts/pkg/apis/rollouts"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: rollouts.Group, Version: "v1alpha1"}

var (
	// GroupVersionResource for all rollout types
	RolloutGVR                 = SchemeGroupVersion.WithResource("rollouts")
	AnalysisRunGVR             = SchemeGroupVersion.WithResource("analysisruns")
	AnalysisTemplateGVR        = SchemeGroupVersion.WithResource("analysistemplates")
	ClusterAnalysisTemplateGVR = SchemeGroupVersion.WithResource("clusteranalysistemplates")
	ExperimentGVR              = SchemeGroupVersion.WithResource("experiments")
)

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Rollout{},
		&RolloutList{},
		&Experiment{},
		&ExperimentList{},
		&AnalysisTemplate{},
		&AnalysisTemplateList{},
		&ClusterAnalysisTemplate{},
		&ClusterAnalysisTemplateList{},
		&AnalysisRun{},
		&AnalysisRunList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
