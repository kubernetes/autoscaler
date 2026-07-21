/*
Copyright 2017 The Kubernetes Authors.

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

package routines

import (
	"context"
	"slices"
	"sync"
	"time"

	"k8s.io/klog/v2"

	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	scopeutil "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/scope"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender duties followed by update of recommendations in VPA objects.
	RunOnce()
	// GetClusterState returns ClusterState used by Recommender
	GetClusterState() model.ClusterState
	// GetClusterStateFeeder returns ClusterStateFeeder used by Recommender
	GetClusterStateFeeder() input.ClusterStateFeeder
	// UpdateVPAs computes recommendations and sends VPAs status updates to API Server
	UpdateVPAs()
	// MaintainCheckpoints stores current checkpoints in API Server and garbage collect old ones
	// MaintainCheckpoints writes checkpoints for at least `update-worker-count` number of VPAs.
	// Checkpoints are written until ctx permits or all checkpoints are written.
	MaintainCheckpoints(ctx context.Context)
}

type recommender struct {
	clusterState                  model.ClusterState
	clusterStateFeeder            input.ClusterStateFeeder
	checkpointWriter              checkpoint.CheckpointWriter
	checkpointsGCInterval         time.Duration
	checkpointsWriteTimeout       time.Duration
	controllerFetcher             controllerfetcher.ControllerFetcher
	lastCheckpointGC              time.Time
	vpaClient                     vpa_api.VerticalPodAutoscalersGetter
	podResourceRecommender        logic.PodResourceRecommender
	recommendationFormat          logic.RecommendationFormat
	useCheckpoints                bool
	lastAggregateContainerStateGC time.Time
	recommendationPostProcessor   []RecommendationPostProcessor
	updateWorkerCount             int
	// scopedRecommendationCache stores final scoped recommendations keyed by VPA ID and
	// recommendation generation. It lets us skip grouped recomputation when no inputs changed.
	scopedRecommendationCache      map[model.VpaID]scopedRecommendationCacheEntry
	scopedRecommendationCacheMutex sync.RWMutex
	// observedStatusMu guards reads/writes of ObservedVPAs status fields. UpdateVPAs may run
	// concurrently (and workers process shared observed objects), so in-memory status sync
	// must not race with status comparisons.
	observedStatusMu sync.Mutex
}

type scopedRecommendationCacheEntry struct {
	generation     uint64
	recommendation *vpaautoscalingv1.RecommendedPodResources
	groups         []vpaautoscalingv1.RecommendedPodResourcesGroup
}

func (r *recommender) GetClusterState() model.ClusterState {
	return r.clusterState
}

func (r *recommender) GetClusterStateFeeder() input.ClusterStateFeeder {
	return r.clusterStateFeeder
}

func (r *recommender) getScopedRecommendationCache(vpaID model.VpaID, generation uint64) (*vpaautoscalingv1.RecommendedPodResources, []vpaautoscalingv1.RecommendedPodResourcesGroup, bool) {
	r.scopedRecommendationCacheMutex.RLock()
	defer r.scopedRecommendationCacheMutex.RUnlock()
	entry, found := r.scopedRecommendationCache[vpaID]
	if !found || entry.generation != generation {
		return nil, nil, false
	}
	return entry.recommendation, entry.groups, true
}

func (r *recommender) setScopedRecommendationCache(vpaID model.VpaID, generation uint64, recommendation *vpaautoscalingv1.RecommendedPodResources, groups []vpaautoscalingv1.RecommendedPodResourcesGroup) {
	r.scopedRecommendationCacheMutex.Lock()
	defer r.scopedRecommendationCacheMutex.Unlock()
	if r.scopedRecommendationCache == nil {
		r.scopedRecommendationCache = make(map[model.VpaID]scopedRecommendationCacheEntry)
	}
	r.scopedRecommendationCache[vpaID] = scopedRecommendationCacheEntry{
		generation:     generation,
		recommendation: recommendation,
		groups:         groups,
	}
}

func processVPAUpdate(r *recommender, vpa *model.Vpa, observedVpa *vpaautoscalingv1.VerticalPodAutoscaler) {
	had := vpa.HasRecommendation()
	isScopedDaemonSet := observedVpa.Spec.TargetRef != nil && scopeutil.IsScopedDaemonSet(observedVpa.Spec.TargetRef.Kind, string(observedVpa.Spec.Scope))
	var recommendation *vpaautoscalingv1.RecommendedPodResources
	var groups []vpaautoscalingv1.RecommendedPodResourcesGroup

	scopedGeneration := vpa.ScopedRecommendationGeneration()
	cacheHit := false
	if isScopedDaemonSet {
		// If no scoped input changed since the previous run, reuse final recommendations
		// and avoid rebuilding grouped state for all scope values.
		if cachedRecommendation, cachedGroups, found := r.getScopedRecommendationCache(vpa.ID, scopedGeneration); found {
			recommendation = cachedRecommendation
			groups = cachedGroups
			cacheHit = true
		}
	}
	if !cacheHit {
		resources := r.podResourceRecommender.GetRecommendedPodResources(GetContainerNameToAggregateStateMap(vpa))
		recommendation = logic.MapToListOfRecommendedContainerResources(resources, r.recommendationFormat)
	}

	if isScopedDaemonSet && !cacheHit {
		recommendationByScopeValue := GetContainerNameToAggregateStateMapByScopeValue(vpa)
		scopeValues := make([]string, 0, len(recommendationByScopeValue))
		for scopeValue := range recommendationByScopeValue {
			scopeValues = append(scopeValues, scopeValue)
		}
		slices.Sort(scopeValues)
		// Preallocate output groups to avoid repeated slice growth for large scoped DaemonSets.
		groups = make([]vpaautoscalingv1.RecommendedPodResourcesGroup, 0, len(scopeValues))
		for _, scopeValue := range scopeValues {
			resources := r.podResourceRecommender.GetRecommendedPodResources(recommendationByScopeValue[scopeValue])
			groupRecommendation := logic.MapToListOfRecommendedContainerResources(resources, r.recommendationFormat)
			groups = append(groups, vpaautoscalingv1.RecommendedPodResourcesGroup{
				ScopeValue:               scopeValue,
				ContainerRecommendations: groupRecommendation.ContainerRecommendations,
			})
		}
	}

	if !cacheHit {
		for _, postProcessor := range r.recommendationPostProcessor {
			recommendation = postProcessor.Process(observedVpa, recommendation)
			for i := range groups {
				groupRec := &vpaautoscalingv1.RecommendedPodResources{ContainerRecommendations: groups[i].ContainerRecommendations}
				groupRec = postProcessor.Process(observedVpa, groupRec)
				groups[i].ContainerRecommendations = groupRec.ContainerRecommendations
			}
		}

		if isScopedDaemonSet {
			for i := range groups {
				for j := range groups[i].ContainerRecommendations {
					containerRec := groups[i].ContainerRecommendations[j]
					// Compact group payload: keep only effective target.
					containerRec.LowerBound = nil
					containerRec.UpperBound = nil
					containerRec.UncappedTarget = nil
					groups[i].ContainerRecommendations[j] = containerRec
				}
			}
			r.setScopedRecommendationCache(vpa.ID, scopedGeneration, recommendation, groups)
		}
	}

	status := vpa.AsStatus()
	// Always publish the global recommendation, including for scoped DaemonSets.
	// It is the aggregate across all pods and serves as a safe fallback for
	// clients that do not consume recommendationGroups (for example when the
	// DaemonSetScope feature gate is disabled during a rollback).
	status.Recommendation = recommendation
	status.RecommendationGroups = groups
	vpa.UpdateRecommendation(recommendation)
	if vpa.HasRecommendation() && !had {
		metrics_recommender.ObserveRecommendationLatency(vpa.Created)
	}
	hasMatchingPods := vpa.PodCount > 0
	vpa.UpdateConditions(hasMatchingPods)
	if err := r.clusterState.RecordRecommendation(vpa, time.Now()); err != nil {
		klog.V(0).InfoS("", "err", err)
		if klog.V(4).Enabled() {
			pods := r.clusterState.GetMatchingPods(vpa)
			if len(pods) != vpa.PodCount {
				klog.ErrorS(nil, "ClusterState pod count and matching pods disagree for VPA", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName), "podCount", vpa.PodCount, "matchingPods", pods)
			}
			klog.InfoS("VPA dump", "vpa", vpa, "hasMatchingPods", hasMatchingPods, "podCount", vpa.PodCount, "matchingPods", pods)
		}
	}

	r.observedStatusMu.Lock()
	statusUnchanged := vpa_utils.StatusEqual(&observedVpa.Status, status)
	var oldStatus *vpaautoscalingv1.VerticalPodAutoscalerStatus
	if !statusUnchanged {
		oldStatus = observedVpa.Status.DeepCopy()
	}
	r.observedStatusMu.Unlock()

	if statusUnchanged {
		return
	}

	_, err := vpa_utils.UpdateVpaStatusIfNeeded(
		r.vpaClient.VerticalPodAutoscalers(vpa.ID.Namespace), vpa.ID.VpaName, status, oldStatus)
	if err != nil {
		klog.ErrorS(err, "Cannot update VPA", "vpa", klog.KRef(vpa.ID.Namespace, vpa.ID.VpaName))
		return
	}

	// Keep local cache in sync so repeated processing doesn't recompute/patch identical status.
	r.observedStatusMu.Lock()
	observedVpa.Status = *status
	r.observedStatusMu.Unlock()
}

// UpdateVPAs update VPA CRD objects' status.
func (r *recommender) UpdateVPAs() {
	cnt := metrics_recommender.NewObjectCounter()
	defer cnt.Observe()

	// Create a channel to send VPA updates to workers
	vpaUpdates := make(chan *vpaautoscalingv1.VerticalPodAutoscaler, len(r.clusterState.ObservedVPAs()))

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < r.updateWorkerCount; i++ {
		wg.Go(func() {
			for observedVpa := range vpaUpdates {
				key := model.VpaID{
					Namespace: observedVpa.Namespace,
					VpaName:   observedVpa.Name,
				}

				vpa, found := r.clusterState.VPAs()[key]
				if !found {
					return
				}
				processVPAUpdate(r, vpa, observedVpa)
				cnt.Add(vpa)
			}
		})
	}

	// Send VPA updates to the workers
	for _, observedVpa := range r.clusterState.ObservedVPAs() {
		vpaUpdates <- observedVpa
	}

	// Close the channel to signal workers to stop
	close(vpaUpdates)

	// Wait for all workers to finish
	wg.Wait()
}

func (r *recommender) MaintainCheckpoints(ctx context.Context) {
	if r.useCheckpoints {
		r.checkpointWriter.StoreCheckpoints(ctx, r.updateWorkerCount)

		if time.Since(r.lastCheckpointGC) > r.checkpointsGCInterval {
			r.lastCheckpointGC = time.Now()
			r.clusterStateFeeder.GarbageCollectCheckpoints(ctx)
		}
	}
}

func (r *recommender) RunOnce() {
	timer := metrics_recommender.NewExecutionTimer()
	defer timer.ObserveTotal()

	ctx := context.Background()

	klog.V(3).InfoS("Recommender Run")

	r.clusterStateFeeder.LoadVPAs(ctx)
	timer.ObserveStep("LoadVPAs")

	r.clusterStateFeeder.LoadPods()
	timer.ObserveStep("LoadPods")

	r.clusterStateFeeder.LoadRealTimeMetrics(ctx)
	timer.ObserveStep("LoadMetrics")

	r.clusterStateFeeder.DeleteRemovedPods()
	timer.ObserveStep("DeleteRemovedPods")

	klog.V(3).InfoS("ClusterState is tracking", "pods", len(r.clusterState.Pods()), "vpas", len(r.clusterState.VPAs()))

	r.UpdateVPAs()
	timer.ObserveStep("UpdateVPAs")

	stepCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(r.checkpointsWriteTimeout))
	defer cancelFunc()
	r.MaintainCheckpoints(stepCtx)
	timer.ObserveStep("MaintainCheckpoints")

	r.clusterState.RateLimitedGarbageCollectAggregateCollectionStates(ctx, time.Now(), r.controllerFetcher)
	timer.ObserveStep("GarbageCollect")
	klog.V(3).InfoS("ClusterState is tracking", "aggregateContainerStates", r.clusterState.StateMapSize())
}

// RecommenderFactory makes instances of Recommender.
type RecommenderFactory struct {
	ClusterState model.ClusterState

	ClusterStateFeeder     input.ClusterStateFeeder
	ControllerFetcher      controllerfetcher.ControllerFetcher
	CheckpointWriter       checkpoint.CheckpointWriter
	PodResourceRecommender logic.PodResourceRecommender
	RecommendationFormat   logic.RecommendationFormat
	VpaClient              vpa_api.VerticalPodAutoscalersGetter

	RecommendationPostProcessors []RecommendationPostProcessor

	CheckpointsGCInterval   time.Duration
	CheckpointsWriteTimeout time.Duration
	UseCheckpoints          bool
	UpdateWorkerCount       int
}

// Make creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
func (c RecommenderFactory) Make() Recommender {
	recommender := &recommender{
		clusterState:                  c.ClusterState,
		clusterStateFeeder:            c.ClusterStateFeeder,
		checkpointWriter:              c.CheckpointWriter,
		checkpointsGCInterval:         c.CheckpointsGCInterval,
		checkpointsWriteTimeout:       c.CheckpointsWriteTimeout,
		controllerFetcher:             c.ControllerFetcher,
		useCheckpoints:                c.UseCheckpoints,
		vpaClient:                     c.VpaClient,
		podResourceRecommender:        c.PodResourceRecommender,
		recommendationFormat:          c.RecommendationFormat,
		recommendationPostProcessor:   c.RecommendationPostProcessors,
		lastAggregateContainerStateGC: time.Now(),
		lastCheckpointGC:              time.Now(),
		updateWorkerCount:             c.UpdateWorkerCount,
		scopedRecommendationCache:     make(map[model.VpaID]scopedRecommendationCacheEntry),
	}
	klog.V(3).InfoS("New Recommender created", "recommender", recommender)
	return recommender
}
