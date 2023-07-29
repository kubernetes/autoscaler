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
	"flag"
	"sort"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned/scheme"
	mpa_api "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/checkpoint"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/metrics/recommender"
	mpa_utils "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
	hpa "k8s.io/kubernetes/pkg/controller/podautoscaler"
	metricsclient "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
)

const (
	// AggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
	AggregateContainerStateGCInterval               = 1 * time.Hour
	scaleCacheEntryLifetime           time.Duration = time.Hour
	scaleCacheEntryFreshnessTime      time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor       float64       = 1.
	defaultResyncPeriod               time.Duration = 10 * time.Minute
)

var (
	checkpointsWriteTimeout = flag.Duration("checkpoints-timeout", time.Minute, `Timeout for writing checkpoints since the start of the recommender's main loop`)
	minCheckpointsPerRun    = flag.Int("min-checkpoints", 10, "Minimum number of checkpoints to write per recommender's main loop")
	memorySaver             = flag.Bool("memory-saver", false, `If true, only track pods which have an associated MPA`)
)

// From HPA
var (
	scaleUpLimitFactor  = 2.0
	scaleUpLimitMinimum = 4.0
)
type timestampedRecommendation struct {
	recommendation int32
	timestamp      time.Time
}
type timestampedScaleEvent struct {
	replicaChange int32  // absolute value, non-negative
	timestamp     time.Time
	outdated      bool
}

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender duties followed by update of recommendations in MPA objects.
	RunOnce(workers int, vpaOrHpa string)
	// GetClusterState returns ClusterState used by Recommender
	GetClusterState() *model.ClusterState
	// GetClusterStateFeeder returns ClusterStateFeeder used by Recommender
	GetClusterStateFeeder() input.ClusterStateFeeder
	// UpdateMPAs computes recommendations and sends MPAs status updates to API Server
	UpdateMPAs(ctx context.Context, vpaOrHpa string)
	// MaintainCheckpoints stores current checkpoints in API Server and garbage collect old ones
	// MaintainCheckpoints writes at least minCheckpoints if there are more checkpoints to write.
	// Checkpoints are written until ctx permits or all checkpoints are written.
	MaintainCheckpoints(ctx context.Context, minCheckpoints int)
}

type recommender struct {
	// Fields inherited from VPA.
	clusterState                  *model.ClusterState
	clusterStateFeeder            input.ClusterStateFeeder
	checkpointWriter              checkpoint.CheckpointWriter
	checkpointsGCInterval         time.Duration
	controllerFetcher             controllerfetcher.ControllerFetcher
	lastCheckpointGC              time.Time
	mpaClient                     mpa_api.MultidimPodAutoscalersGetter
	podResourceRecommender        logic.PodResourceRecommender
	useCheckpoints                bool
	lastAggregateContainerStateGC time.Time

	// Fields for HPA.
	replicaCalc                   *hpa.ReplicaCalculator
	eventRecorder                 record.EventRecorder
	downscaleStabilisationWindow  time.Duration
	// Controllers that need to be synced.
	// Latest unstabilized recommendations for each autoscaler.
	recommendations               map[model.MpaID][]timestampedRecommendation
	recommendationsLock           sync.Mutex
	// Latest autoscaler events.
	scaleUpEvents                 map[model.MpaID][]timestampedScaleEvent
	scaleUpEventsLock             sync.RWMutex
	scaleDownEvents               map[model.MpaID][]timestampedScaleEvent
	scaleDownEventsLock           sync.RWMutex
}

func (r *recommender) GetClusterState() *model.ClusterState {
	return r.clusterState
}

func (r *recommender) GetClusterStateFeeder() input.ClusterStateFeeder {
	return r.clusterStateFeeder
}

// Updates MPA CRD objects' statuses.
// vpaOrHpa can be either 'vpa', 'hpa', or 'both'.
func (r *recommender) UpdateMPAs(ctx context.Context, vpaOrHpa string) {
	cnt := metrics_recommender.NewObjectCounter()
	defer cnt.Observe()

	for _, observedMpa := range r.clusterState.ObservedMpas {
		key := model.MpaID{
			Namespace: observedMpa.Namespace,
			MpaName:   observedMpa.Name,
		}
		klog.V(4).Infof("Recommender is checking MPA %v...", key)

		mpa, found := r.clusterState.Mpas[key]
		if !found {
			klog.V(4).Infof("MPA %v not found in the cluster state map!", key)
			continue
		}

		// Vertical Pod Autoscaling
		if (vpaOrHpa != "hpa") {
			klog.V(4).Infof("Vertical scaling...")
			resources := r.podResourceRecommender.GetRecommendedPodResources(GetContainerNameToAggregateStateMap(mpa))
			had := mpa.HasRecommendation()
			mpa.UpdateRecommendation(getCappedRecommendation(mpa.ID, resources, observedMpa.Spec.ResourcePolicy))
			klog.V(4).Infof("MPA %v recommendation updated: %v (%v)", key, resources, had)
			if mpa.HasRecommendation() && !had {
				metrics_recommender.ObserveRecommendationLatency(mpa.Created)
			}
			hasMatchingPods := mpa.PodCount > 0
			mpa.UpdateConditions(hasMatchingPods)
			if err := r.clusterState.RecordRecommendation(mpa, time.Now()); err != nil {
				klog.Warningf("%v", err)
				if klog.V(4).Enabled() {
					klog.Infof("MPA dump")
					klog.Infof("%+v", mpa)
					klog.Infof("HasMatchingPods: %v", hasMatchingPods)
					klog.Infof("PodCount: %v", mpa.PodCount)
					pods := r.clusterState.GetMatchingPods(mpa)
					klog.Infof("MatchingPods: %+v", pods)
					if len(pods) != mpa.PodCount {
						klog.Errorf("ClusterState pod count and matching pods disagree for mpa %v/%v", mpa.ID.Namespace, mpa.ID.MpaName)
					}
				}
			}
			cnt.Add(mpa)

			_, err := mpa_utils.UpdateMpaStatusIfNeeded(
				r.mpaClient.MultidimPodAutoscalers(mpa.ID.Namespace), mpa.ID.MpaName, mpa.AsStatus(), &observedMpa.Status)
			if err != nil {
				klog.Errorf(
					"Cannot update MPA %v object. Reason: %+v", mpa.ID.MpaName, err)
			}
		}

		// Horizontal Pod Autoscaling
		if (vpaOrHpa != "vpa") {
			observedMpa.Status.Recommendation = mpa.AsStatus().Recommendation
			observedMpa.Status.Conditions = mpa.AsStatus().Conditions
			klog.V(4).Infof("Horizontal scaling...")
			errHPA := r.ReconcileHorizontalAutoscaling(ctx, observedMpa, key)
			if errHPA != nil {
				klog.Errorf("Error updating MPA status: %v", errHPA.Error())
			}
		}
	}
}

// getCappedRecommendation creates a recommendation based on recommended pod
// resources, setting the UncappedTarget to the calculated recommended target
// and if necessary, capping the Target, LowerBound and UpperBound according
// to the ResourcePolicy.
func getCappedRecommendation(mpaID model.MpaID, resources logic.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources {
	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	// Sort the container names from the map. This is because maps are an
	// unordered data structure, and iterating through the map will return
	// a different order on every call.
	containerNames := make([]string, 0, len(resources))
	for containerName := range resources {
		containerNames = append(containerNames, containerName)
	}
	sort.Strings(containerNames)
	// Create the list of recommendations for each container.
	for _, name := range containerNames {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  name,
			Target:         vpa_model.ResourcesAsResourceList(resources[name].Target),
			LowerBound:     vpa_model.ResourcesAsResourceList(resources[name].LowerBound),
			UpperBound:     vpa_model.ResourcesAsResourceList(resources[name].UpperBound),
			UncappedTarget: vpa_model.ResourcesAsResourceList(resources[name].Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: containerResources,
	}
	// Keep the original VPA policy for vertical autoscaling.
	cappedRecommendation, err := vpa_utils.ApplyVPAPolicy(recommendation, policy)
	if err != nil {
		klog.Errorf("Failed to apply policy for MPA %v/%v: %v", mpaID.Namespace, mpaID.MpaName, err)
		return recommendation
	}
	return cappedRecommendation
}

func (r *recommender) MaintainCheckpoints(ctx context.Context, minCheckpointsPerRun int) {
	now := time.Now()
	if r.useCheckpoints {
		if err := r.checkpointWriter.StoreCheckpoints(ctx, now, minCheckpointsPerRun); err != nil {
			klog.Warningf("Failed to store checkpoints. Reason: %+v", err)
		}
		if time.Since(r.lastCheckpointGC) > r.checkpointsGCInterval {
			r.lastCheckpointGC = now
			r.clusterStateFeeder.GarbageCollectCheckpoints()
		}
	}
}

func (r *recommender) RunOnce(workers int, vpaOrHpa string) {
	timer := metrics_recommender.NewExecutionTimer()
	defer timer.ObserveTotal()

	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*checkpointsWriteTimeout))
	defer cancelFunc()

	// From HPA.
	defer utilruntime.HandleCrash()

	klog.V(3).Infof("Recommender Run")
	defer klog.V(3).Infof("Shutting down MPA Recommender")

	r.clusterStateFeeder.LoadMPAs()
	timer.ObserveStep("LoadMPAs")

	r.clusterStateFeeder.LoadPods()
	timer.ObserveStep("LoadPods")

	r.clusterStateFeeder.LoadRealTimeMetrics()
	timer.ObserveStep("LoadMetrics")
	klog.V(3).Infof("ClusterState is tracking %v PodStates and %v MPAs", len(r.clusterState.Pods), len(r.clusterState.Mpas))

	r.UpdateMPAs(ctx, vpaOrHpa)
	timer.ObserveStep("UpdateMPAs")

	r.MaintainCheckpoints(ctx, *minCheckpointsPerRun)
	timer.ObserveStep("MaintainCheckpoints")

	r.clusterState.RateLimitedGarbageCollectAggregateCollectionStates(time.Now(), r.controllerFetcher)
	timer.ObserveStep("GarbageCollect")
	klog.V(3).Infof("ClusterState is tracking %d aggregated container states", r.clusterState.StateMapSize())
}

// RecommenderFactory makes instances of Recommender.
type RecommenderFactory struct {
	ClusterState *model.ClusterState

	ClusterStateFeeder     input.ClusterStateFeeder
	ControllerFetcher      controllerfetcher.ControllerFetcher
	CheckpointWriter       checkpoint.CheckpointWriter
	PodResourceRecommender logic.PodResourceRecommender
	MpaClient              mpa_api.MultidimPodAutoscalersGetter

	CheckpointsGCInterval time.Duration
	UseCheckpoints        bool

	// For HPA.
	EvtNamespacer                 v1core.EventsGetter
	PodInformer                   coreinformers.PodInformer
	MetricsClient                 metricsclient.MetricsClient
	ResyncPeriod                  time.Duration
	DownscaleStabilisationWindow  time.Duration
	Tolerance                     float64
	CpuInitializationPeriod       time.Duration
	DelayOfInitialReadinessStatus time.Duration
}

// Make creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
func (c RecommenderFactory) Make() Recommender {
	// For HPA.
	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.EvtNamespacer.Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "horizontal-pod-autoscaler"})

	recommender := &recommender{
		// From VPA.
		clusterState:                  c.ClusterState,
		clusterStateFeeder:            c.ClusterStateFeeder,
		checkpointWriter:              c.CheckpointWriter,
		checkpointsGCInterval:         c.CheckpointsGCInterval,
		controllerFetcher:             c.ControllerFetcher,
		useCheckpoints:                c.UseCheckpoints,
		mpaClient:                     c.MpaClient,
		podResourceRecommender:        c.PodResourceRecommender,
		lastAggregateContainerStateGC: time.Now(),
		lastCheckpointGC:              time.Now(),

		// From HPA.
		downscaleStabilisationWindow:  c.DownscaleStabilisationWindow,
		// podLister is able to list/get Pods from the shared cache from the informer passed in to
		// NewHorizontalController.
		eventRecorder:                 recorder,
		recommendations:               map[model.MpaID][]timestampedRecommendation{},
		recommendationsLock:           sync.Mutex{},
		scaleUpEvents:                 map[model.MpaID][]timestampedScaleEvent{},
		scaleUpEventsLock:             sync.RWMutex{},
		scaleDownEvents:               map[model.MpaID][]timestampedScaleEvent{},
		scaleDownEventsLock:           sync.RWMutex{},
	}

	replicaCalc := hpa.NewReplicaCalculator(
		c.MetricsClient,
		recommender.clusterStateFeeder.GetPodLister(),
		c.Tolerance,
		c.CpuInitializationPeriod,
		c.DelayOfInitialReadinessStatus,
	)
	recommender.replicaCalc = replicaCalc

	klog.V(3).Infof("New Recommender created!")
	return recommender
}

// NewRecommender creates a new recommender instance.
// Dependencies are created automatically.
// Deprecated; use RecommenderFactory instead.
func NewRecommender(config *rest.Config, checkpointsGCInterval time.Duration, useCheckpoints bool, namespace string, recommenderName string, evtNamespacer v1core.EventsGetter, metricsClient metricsclient.MetricsClient, resyncPeriod time.Duration, downscaleStabilisationWindow time.Duration, tolerance float64, cpuInitializationPeriod time.Duration, delayOfInitialReadinessStatus time.Duration,
) Recommender {
	clusterState := model.NewClusterState(AggregateContainerStateGCInterval)
	kubeClient := kube_client.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(namespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	return RecommenderFactory{
		ClusterState:           clusterState,
		ClusterStateFeeder:     input.NewClusterStateFeeder(config, clusterState, *memorySaver, namespace, "default-metrics-client", recommenderName),
		ControllerFetcher:      controllerFetcher,
		CheckpointWriter:       checkpoint.NewCheckpointWriter(clusterState, vpa_clientset.NewForConfigOrDie(config).AutoscalingV1()),
		MpaClient:              mpa_clientset.NewForConfigOrDie(config).AutoscalingV1alpha1(),
		PodResourceRecommender: logic.CreatePodResourceRecommender(),
		CheckpointsGCInterval:  checkpointsGCInterval,
		UseCheckpoints:         useCheckpoints,

		// For HPA.
		EvtNamespacer:                 evtNamespacer,
		PodInformer:                   factory.Core().V1().Pods(),
		MetricsClient:                 metricsClient,
		ResyncPeriod:                  resyncPeriod,
		DownscaleStabilisationWindow:  downscaleStabilisationWindow,
		Tolerance:                     tolerance,
		CpuInitializationPeriod:       cpuInitializationPeriod,
		DelayOfInitialReadinessStatus: delayOfInitialReadinessStatus,
	}.Make()
}
