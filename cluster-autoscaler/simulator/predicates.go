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

	apiv1 "k8s.io/api/core/v1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	informers "k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	"k8s.io/kubernetes/pkg/scheduler/factory"

	// We need to import provider to initialize default scheduler.
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"

	"github.com/golang/glog"
)

const (
	// We want to disable affinity predicate for performance reasons if no pod
	// requires it.
	affinityPredicateName = "MatchInterPodAffinity"
)

type predicateInfo struct {
	name      string
	predicate algorithm.FitPredicate
}

// PredicateChecker checks whether all required predicates pass for given Pod and Node.
type PredicateChecker struct {
	predicates                []predicateInfo
	predicateMetadataProducer algorithm.PredicateMetadataProducer
	enableAffinityPredicate   bool
}

// There are no const arrays in Go, this is meant to be used as a const.
var priorityPredicates = []string{"PodFitsResources", "GeneralPredicates", "PodToleratesNodeTaints"}

func init() {
	// This results in filtering out some predicate functions registered by defaults.init() method.
	// In scheduler this method is run from app.runCommand().
	// We also need to call it in CA to have simulation behaviour consistent with scheduler.
	// Note: the logic of method is conditional and depends on feature gates enabled. To have same
	//       behaviour in CA and scheduler both need to be run with same set of feature gates.
	algorithmprovider.ApplyFeatureGates()
}

// NewPredicateChecker builds PredicateChecker.
func NewPredicateChecker(kubeClient kube_client.Interface, stop <-chan struct{}) (*PredicateChecker, error) {
	provider, err := factory.GetAlgorithmProvider(factory.DefaultProvider)
	if err != nil {
		return nil, err
	}
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)

	schedulerConfigFactory := factory.NewConfigFactory(&factory.ConfigFactoryArgs{
		SchedulerName:                  "cluster-autoscaler",
		Client:                         kubeClient,
		NodeInformer:                   informerFactory.Core().V1().Nodes(),
		PodInformer:                    informerFactory.Core().V1().Pods(),
		PvInformer:                     informerFactory.Core().V1().PersistentVolumes(),
		PvcInformer:                    informerFactory.Core().V1().PersistentVolumeClaims(),
		ReplicationControllerInformer:  informerFactory.Core().V1().ReplicationControllers(),
		ReplicaSetInformer:             informerFactory.Apps().V1().ReplicaSets(),
		StatefulSetInformer:            informerFactory.Apps().V1().StatefulSets(),
		ServiceInformer:                informerFactory.Core().V1().Services(),
		PdbInformer:                    informerFactory.Policy().V1beta1().PodDisruptionBudgets(),
		StorageClassInformer:           informerFactory.Storage().V1().StorageClasses(),
		HardPodAffinitySymmetricWeight: apiv1.DefaultHardPodAffinitySymmetricWeight,
	})
	informerFactory.Start(stop)

	metadataProducer, err := schedulerConfigFactory.GetPredicateMetadataProducer()
	if err != nil {
		return nil, err
	}
	predicateMap, err := schedulerConfigFactory.GetPredicates(provider.FitPredicateKeys)
	predicateMap["ready"] = isNodeReadyAndSchedulablePredicate
	if err != nil {
		return nil, err
	}
	// We always want to have PodFitsResources as a first predicate we run
	// as this is cheap to check and it should be enough to fail predicates
	// in most of our simulations (especially binpacking).
	if _, found := predicateMap["PodFitsResources"]; !found {
		predicateMap["PodFitsResources"] = predicates.PodFitsResources
	}

	predicateList := make([]predicateInfo, 0)
	for _, predicateName := range priorityPredicates {
		if predicate, found := predicateMap[predicateName]; found {
			predicateList = append(predicateList, predicateInfo{name: predicateName, predicate: predicate})
			delete(predicateMap, predicateName)
		}
	}

	for predicateName, predicate := range predicateMap {
		predicateList = append(predicateList, predicateInfo{name: predicateName, predicate: predicate})
	}

	for _, predInfo := range predicateList {
		glog.V(1).Infof("Using predicate %s", predInfo.name)
	}

	// TODO: Verify that run is not needed anymore.
	// schedulerConfigFactory.Run()

	return &PredicateChecker{
		predicates:                predicateList,
		predicateMetadataProducer: metadataProducer,
		enableAffinityPredicate:   true,
	}, nil
}

func isNodeReadyAndSchedulablePredicate(pod *apiv1.Pod, meta algorithm.PredicateMetadata, nodeInfo *schedulercache.NodeInfo) (bool,
	[]algorithm.PredicateFailureReason, error) {
	ready := kube_util.IsNodeReadyAndSchedulable(nodeInfo.Node())
	if !ready {
		return false, []algorithm.PredicateFailureReason{predicates.NewFailureReason("node is unready")}, nil
	}
	return true, []algorithm.PredicateFailureReason{}, nil
}

// NewTestPredicateChecker builds test version of PredicateChecker.
func NewTestPredicateChecker() *PredicateChecker {
	return &PredicateChecker{
		predicates: []predicateInfo{
			{name: "default", predicate: predicates.GeneralPredicates},
			{name: "ready", predicate: isNodeReadyAndSchedulablePredicate},
		},
		predicateMetadataProducer: func(_ *apiv1.Pod, _ map[string]*schedulercache.NodeInfo) algorithm.PredicateMetadata {
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
func (p *PredicateChecker) GetPredicateMetadata(pod *apiv1.Pod, nodeInfos map[string]*schedulercache.NodeInfo) algorithm.PredicateMetadata {
	// Skip precomputation if affinity predicate is disabled - it's not worth it performance-wise.
	if !p.enableAffinityPredicate {
		return nil
	}
	return p.predicateMetadataProducer(pod, nodeInfos)
}

// FitsAny checks if the given pod can be place on any of the given nodes.
func (p *PredicateChecker) FitsAny(pod *apiv1.Pod, nodeInfos map[string]*schedulercache.NodeInfo) (string, error) {
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
	failureReasons []algorithm.PredicateFailureReason
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
func NewPredicateError(name string, err error, reasons []string, originalReasons []algorithm.PredicateFailureReason) *PredicateError {
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
func (pe *PredicateError) OriginalReasons() []algorithm.PredicateFailureReason {
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
func (p *PredicateChecker) CheckPredicates(pod *apiv1.Pod, predicateMetadata algorithm.PredicateMetadata, nodeInfo *schedulercache.NodeInfo) *PredicateError {
	for _, predInfo := range p.predicates {
		// Skip affinity predicate if it has been disabled.
		if !p.enableAffinityPredicate && predInfo.name == affinityPredicateName {
			continue
		}

		match, failureReasons, err := predInfo.predicate(pod, predicateMetadata, nodeInfo)

		if err != nil || !match {
			return &PredicateError{
				predicateName:  predInfo.name,
				failureReasons: failureReasons,
				err:            err,
			}
		}
	}
	return nil
}
