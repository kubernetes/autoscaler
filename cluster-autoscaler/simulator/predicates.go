/*
Copyright 2016 The Kubernetes Authors.

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

package simulator

import (
	"fmt"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	informers "k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/factory"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	// We need to import provider to initialize default scheduler.
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"

	"k8s.io/klog"
)

const (
	// We want to disable affinity predicate for performance reasons if no pod
	// requires it.
	affinityPredicateName = "MatchInterPodAffinity"
)

var (
	// initMutex is used for guarding static initialization.
	staticInitMutex sync.Mutex
	// statiInitHappened denotes if static initialization happened.
	staticInitDone bool
)

// PredicateInfo assigns a name to a predicate
type PredicateInfo struct {
	Name      string
	Predicate predicates.FitPredicate
}

// PredicateChecker checks whether all required predicates pass for given Pod and Node.
type PredicateChecker struct {
	predicates                []PredicateInfo
	predicateMetadataProducer predicates.PredicateMetadataProducer
	enableAffinityPredicate   bool
}

// We run some predicates first as they are cheap to check and they should be enough
// to fail predicates in most of our simulations (especially binpacking).
// There are no const arrays in Go, this is meant to be used as a const.
var priorityPredicates = []string{"PodFitsResources", "PodToleratesNodeTaints", "GeneralPredicates", "ready"}

func staticInitIfNeeded() {
	staticInitMutex.Lock()
	defer staticInitMutex.Unlock()

	if staticInitDone {
		return
	}

	// This results in filtering out some predicate functions registered by defaults.init() method.
	// In scheduler this method is run from app.runCommand().
	// We also need to call it in CA to have simulation behaviour consistent with scheduler.
	// Note: the logic of method is conditional and depends on feature gates enabled. To have same
	//       behaviour in CA and scheduler both need to be run with same set of feature gates.
	algorithmprovider.ApplyFeatureGates()
	staticInitDone = true
}

// NoOpEventRecorder is a noop implementation of EventRecorder
type NoOpEventRecorder struct{}

// Event is a noop method implementation
func (NoOpEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
}

// Eventf is a noop method implementation
func (NoOpEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}

// PastEventf is a noop method implementation
func (NoOpEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}

// AnnotatedEventf is a noop method implementation
func (NoOpEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker(kubeClient kube_client.Interface, stop <-chan struct{}) (*PredicateChecker, error) {
	staticInitIfNeeded()

	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	algorithmProvider := factory.DefaultProvider

	// Set up the configurator which can create schedulers from configs.
	nodeInformer := informerFactory.Core().V1().Nodes()
	podInformer := informerFactory.Core().V1().Pods()
	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
	replicationControllerInformer := informerFactory.Core().V1().ReplicationControllers()
	replicaSetInformer := informerFactory.Apps().V1().ReplicaSets()
	statefulSetInformer := informerFactory.Apps().V1().StatefulSets()
	serviceInformer := informerFactory.Core().V1().Services()
	pdbInformer := informerFactory.Policy().V1beta1().PodDisruptionBudgets()
	storageClassInformer := informerFactory.Storage().V1().StorageClasses()
	configurator := factory.NewConfigFactory(&factory.ConfigFactoryArgs{
		SchedulerName:                  apiv1.DefaultSchedulerName,
		Client:                         kubeClient,
		NodeInformer:                   nodeInformer,
		PodInformer:                    podInformer,
		PvInformer:                     pvInformer,
		PvcInformer:                    pvcInformer,
		ReplicationControllerInformer:  replicationControllerInformer,
		ReplicaSetInformer:             replicaSetInformer,
		StatefulSetInformer:            statefulSetInformer,
		ServiceInformer:                serviceInformer,
		PdbInformer:                    pdbInformer,
		StorageClassInformer:           storageClassInformer,
		HardPodAffinitySymmetricWeight: apiv1.DefaultHardPodAffinitySymmetricWeight,
		DisablePreemption:              false,
		PercentageOfNodesToScore:       schedulerapi.DefaultPercentageOfNodesToScore,
		BindTimeoutSeconds:             scheduler.BindTimeoutSeconds,
	})
	// Create the config from a named algorithm provider.
	config, err := configurator.CreateFromProvider(algorithmProvider)
	if err != nil {
		return nil, fmt.Errorf("couldn't create scheduler using provider %q: %v", algorithmProvider, err)
	}
	// Additional tweaks to the config produced by the configurator.
	config.Recorder = NoOpEventRecorder{}
	config.DisablePreemption = false
	config.StopEverything = stop

	// Create the scheduler.
	sched := scheduler.NewFromConfig(config)

	scheduler.AddAllEventHandlers(sched, apiv1.DefaultSchedulerName,
		nodeInformer, podInformer, pvInformer, pvcInformer, serviceInformer, storageClassInformer)

	predicateMap := map[string]predicates.FitPredicate{}
	for predicateName, predicateFunc := range sched.Config().Algorithm.Predicates() {
		predicateMap[predicateName] = predicateFunc
	}
	// We want to make sure that some predicates are present to run them first
	// as they are cheap to check and they should be enough to fail predicates
	// in most of our simulations (especially binpacking).
	predicateMap["ready"] = IsNodeReadyAndSchedulablePredicate
	if _, found := predicateMap["PodFitsResources"]; !found {
		predicateMap["PodFitsResources"] = predicates.PodFitsResources
	}
	if _, found := predicateMap["PodToleratesNodeTaints"]; !found {
		predicateMap["PodToleratesNodeTaints"] = predicates.PodToleratesNodeTaints
	}

	predicateList := make([]PredicateInfo, 0, len(predicateMap))
	for _, predicateName := range priorityPredicates {
		if predicate, found := predicateMap[predicateName]; found {
			predicateList = append(predicateList, PredicateInfo{Name: predicateName, Predicate: predicate})
			delete(predicateMap, predicateName)
		}
	}

	for predicateName, predicate := range predicateMap {
		predicateList = append(predicateList, PredicateInfo{Name: predicateName, Predicate: predicate})
	}

	for _, predInfo := range predicateList {
		klog.V(1).Infof("Using predicate %s", predInfo.Name)
	}

	informerFactory.Start(stop)

	metadataProducer, err := configurator.GetPredicateMetadataProducer()
	if err != nil {
		return nil, fmt.Errorf("could not obtain predicateMetadataProducer; %v", err.Error())
	}

	return &PredicateChecker{
		predicates:                predicateList,
		predicateMetadataProducer: metadataProducer,
		enableAffinityPredicate:   true,
	}, nil
}

// IsNodeReadyAndSchedulablePredicate checks if node is ready.
func IsNodeReadyAndSchedulablePredicate(pod *apiv1.Pod, meta predicates.PredicateMetadata, nodeInfo *schedulernodeinfo.NodeInfo) (bool,
	[]predicates.PredicateFailureReason, error) {
	ready := kube_util.IsNodeReadyAndSchedulable(nodeInfo.Node())
	if !ready {
		return false, []predicates.PredicateFailureReason{predicates.NewFailureReason("node is unready")}, nil
	}
	return true, []predicates.PredicateFailureReason{}, nil
}

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() *PredicateChecker {
	return &PredicateChecker{
		predicates: []PredicateInfo{
			{Name: "default", Predicate: predicates.GeneralPredicates},
			{Name: "ready", Predicate: IsNodeReadyAndSchedulablePredicate},
		},
		predicateMetadataProducer: func(_ *apiv1.Pod, _ map[string]*schedulernodeinfo.NodeInfo) predicates.PredicateMetadata {
			return nil
		},
	}
}

// NewCustomTestPredicateChecker builds test version of PredicateChecker with additional predicates.
// Helps with benchmarking different ordering of predicates.
func NewCustomTestPredicateChecker(predicateInfos []PredicateInfo) *PredicateChecker {
	return &PredicateChecker{
		predicates: predicateInfos,
		predicateMetadataProducer: func(_ *apiv1.Pod, _ map[string]*schedulernodeinfo.NodeInfo) predicates.PredicateMetadata {
			return nil
		},
	}
}

// SetAffinityPredicateEnabled can be used to enable or disable checking MatchInterPodAffinity
// predicate. This will cause incorrect CA behavior if there is at least a single pod in
// cluster using affinity/antiaffinity. However, checking affinity predicate is extremely
// costly even if no pod is using it, so it may be worth disabling it in such situation.
func (p *PredicateChecker) SetAffinityPredicateEnabled(enable bool) {
	p.enableAffinityPredicate = enable
}

// IsAffinityPredicateEnabled checks if affinity predicate is enabled.
func (p *PredicateChecker) IsAffinityPredicateEnabled() bool {
	return p.enableAffinityPredicate
}

// GetPredicateMetadata precomputes some information useful for running predicates on a given pod in a given state
// of the cluster (represented by nodeInfos map). Passing the result of this function to CheckPredicates can significantly
// improve the performance of running predicates, especially MatchInterPodAffinity predicate. However, calculating
// predicateMetadata is also quite expensive, so it's not always the best option to run this method.
// Please refer to https://github.com/kubernetes/autoscaler/issues/257 for more details.
func (p *PredicateChecker) GetPredicateMetadata(pod *apiv1.Pod, nodeInfos map[string]*schedulernodeinfo.NodeInfo) predicates.PredicateMetadata {
	// Skip precomputation if affinity predicate is disabled - it's not worth it performance-wise.
	if !p.enableAffinityPredicate {
		return nil
	}
	return p.predicateMetadataProducer(pod, nodeInfos)
}

// FitsAny checks if the given pod can be place on any of the given nodes.
func (p *PredicateChecker) FitsAny(pod *apiv1.Pod, nodeInfos map[string]*schedulernodeinfo.NodeInfo) (string, error) {
	for name, nodeInfo := range nodeInfos {
		// Be sure that the node is schedulable.
		if nodeInfo.Node().Spec.Unschedulable {
			continue
		}
		if err := p.CheckPredicates(pod, nil, nodeInfo); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot put pod %s on any node", pod.Name)
}

// PredicateError implements error, preserving the original error information from scheduler predicate.
type PredicateError struct {
	predicateName  string
	failureReasons []predicates.PredicateFailureReason
	err            error

	reasons []string
	message string
}

// Error returns a predefined error message.
func (pe *PredicateError) Error() string {
	if pe.message != "" {
		return pe.message
	}
	// Don't generate verbose message.
	return "Predicates failed"
}

// VerboseError generates verbose error message if it isn't yet done.
// We're running a ton of predicates and more often than not we only care whether
// they pass or not and don't care for a reason. Turns out formatting nice error
// messages gets very expensive, so we only do it if explicitly requested.
func (pe *PredicateError) VerboseError() string {
	if pe.message != "" {
		return pe.message
	}
	// Generate verbose message.
	if pe.err != nil {
		pe.message = fmt.Sprintf("%s predicate error: %v", pe.predicateName, pe.err)
		return pe.message
	}
	pe.message = fmt.Sprintf("%s predicate mismatch, reason: %s", pe.predicateName, strings.Join(pe.Reasons(), ", "))
	return pe.message
}

// NewPredicateError creates a new predicate error from error and reasons.
func NewPredicateError(name string, err error, reasons []string, originalReasons []predicates.PredicateFailureReason) *PredicateError {
	return &PredicateError{
		predicateName:  name,
		err:            err,
		reasons:        reasons,
		failureReasons: originalReasons,
	}
}

// Reasons returns original failure reasons from failed predicate as a slice of strings.
func (pe *PredicateError) Reasons() []string {
	if pe.reasons != nil {
		return pe.reasons
	}
	pe.reasons = make([]string, len(pe.failureReasons))
	for i, reason := range pe.failureReasons {
		pe.reasons[i] = reason.GetReason()
	}
	return pe.reasons
}

// OriginalReasons returns original failure reasons from failed predicate as a slice of PredicateFailureReason.
func (pe *PredicateError) OriginalReasons() []predicates.PredicateFailureReason {
	return pe.failureReasons
}

// PredicateName returns the name of failed predicate.
func (pe *PredicateError) PredicateName() string {
	return pe.predicateName
}

// CheckPredicates checks if the given pod can be placed on the given node.
// To improve performance predicateMetadata can be calculated using GetPredicateMetadata
// method and passed to CheckPredicates, however, this may lead to incorrect results if
// it was calculated using NodeInfo map representing different cluster state and the
// performance gains of CheckPredicates won't always offset the cost of GetPredicateMetadata.
// Alternatively you can pass nil as predicateMetadata.
func (p *PredicateChecker) CheckPredicates(pod *apiv1.Pod, predicateMetadata predicates.PredicateMetadata, nodeInfo *schedulernodeinfo.NodeInfo) *PredicateError {
	for _, predInfo := range p.predicates {
		// Skip affinity predicate if it has been disabled.
		if !p.enableAffinityPredicate && predInfo.Name == affinityPredicateName {
			continue
		}

		match, failureReasons, err := predInfo.Predicate(pod, predicateMetadata, nodeInfo)

		if err != nil || !match {
			return &PredicateError{
				predicateName:  predInfo.Name,
				failureReasons: failureReasons,
				err:            err,
			}
		}
	}
	return nil
}
