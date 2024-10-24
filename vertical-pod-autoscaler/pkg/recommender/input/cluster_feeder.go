/*
Copyright 2018 The Kubernetes Authors.

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

package input

import (
	"context"
	"fmt"
	"slices"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	kube_client "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/oom"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
)

const (
	evictionWatchRetryWait    = 10 * time.Second
	evictionWatchJitterFactor = 0.5
	// DefaultRecommenderName recommender name explicitly (and so implicitly specify that the default recommender should handle them)
	DefaultRecommenderName = "default"
)

// ClusterStateFeeder can update state of ClusterState object.
type ClusterStateFeeder interface {
	// InitFromHistoryProvider loads historical pod spec into clusterState.
	InitFromHistoryProvider(historyProvider history.HistoryProvider)

	// InitFromCheckpoints loads historical checkpoints into clusterState.
	InitFromCheckpoints()

	// LoadVPAs updates clusterState with current state of VPAs.
	LoadVPAs(ctx context.Context)

	// LoadPods updates clusterState with current specification of Pods and their Containers.
	LoadPods()

	// LoadRealTimeMetrics updates clusterState with current usage metrics of containers.
	LoadRealTimeMetrics()

	// GarbageCollectCheckpoints removes historical checkpoints that don't have a matching VPA.
	GarbageCollectCheckpoints()

	// MarkAggregates marks all aggregates in all VPAs as not under VPAs
	MarkAggregates()

	// SweepAggregates garbage collects all aggregates in all VPAs aggregate lists that are no longer under VPAs
	SweepAggregates()
}

// ClusterStateFeederFactory makes instances of ClusterStateFeeder.
type ClusterStateFeederFactory struct {
	ClusterState        *model.ClusterState
	KubeClient          kube_client.Interface
	MetricsClient       metrics.MetricsClient
	VpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	VpaLister           vpa_lister.VerticalPodAutoscalerLister
	PodLister           v1lister.PodLister
	OOMObserver         oom.Observer
	SelectorFetcher     target.VpaTargetSelectorFetcher
	MemorySaveMode      bool
	ControllerFetcher   controllerfetcher.ControllerFetcher
	RecommenderName     string
	IgnoredNamespaces   []string
}

// Make creates new ClusterStateFeeder with internal data providers, based on kube client.
func (m ClusterStateFeederFactory) Make() *clusterStateFeeder {
	return &clusterStateFeeder{
		coreClient:          m.KubeClient.CoreV1(),
		metricsClient:       m.MetricsClient,
		oomChan:             m.OOMObserver.GetObservedOomsChannel(),
		vpaCheckpointClient: m.VpaCheckpointClient,
		vpaLister:           m.VpaLister,
		clusterState:        m.ClusterState,
		specClient:          spec.NewSpecClient(m.PodLister),
		selectorFetcher:     m.SelectorFetcher,
		memorySaveMode:      m.MemorySaveMode,
		controllerFetcher:   m.ControllerFetcher,
		recommenderName:     m.RecommenderName,
		ignoredNamespaces:   m.IgnoredNamespaces,
	}
}

// WatchEvictionEventsWithRetries watches new Events with reason=Evicted and passes them to the observer.
func WatchEvictionEventsWithRetries(kubeClient kube_client.Interface, observer oom.Observer, namespace string) {
	go func() {
		options := metav1.ListOptions{
			FieldSelector: "reason=Evicted",
		}

		watchEvictionEventsOnce := func() {
			watchInterface, err := kubeClient.CoreV1().Events(namespace).Watch(context.TODO(), options)
			if err != nil {
				klog.ErrorS(err, "Cannot initialize watching events")
				return
			}
			watchEvictionEvents(watchInterface.ResultChan(), observer)
		}
		for {
			watchEvictionEventsOnce()
			// Wait between attempts, retrying too often breaks API server.
			waitTime := wait.Jitter(evictionWatchRetryWait, evictionWatchJitterFactor)
			klog.V(1).InfoS("An attempt to watch eviction events finished", "waitTime", waitTime)
			time.Sleep(waitTime)
		}
	}()
}

func watchEvictionEvents(evictedEventChan <-chan watch.Event, observer oom.Observer) {
	for {
		evictedEvent, ok := <-evictedEventChan
		if !ok {
			klog.V(3).InfoS("Eviction event chan closed")
			return
		}
		if evictedEvent.Type == watch.Added {
			evictedEvent, ok := evictedEvent.Object.(*apiv1.Event)
			if !ok {
				continue
			}
			observer.OnEvent(evictedEvent)
		}
	}
}

// Creates clients watching pods: PodLister (listing only not terminated pods).
func newPodClients(kubeClient kube_client.Interface, resourceEventHandler cache.ResourceEventHandler, namespace string) v1lister.PodLister {
	// We are interested in pods which are Running or Unknown (in case the pod is
	// running but there are some transient errors we don't want to delete it from
	// our model).
	// We don't want to watch Pending pods because they didn't generate any usage
	// yet.
	// Succeeded and Failed failed pods don't generate any usage anymore but we
	// don't necessarily want to immediately delete them.
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(apiv1.PodPending))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	indexer, controller := cache.NewIndexerInformer(
		podListWatch,
		&apiv1.Pod{},
		time.Hour,
		resourceEventHandler,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	podLister := v1lister.NewPodLister(indexer)
	stopCh := make(chan struct{})
	go controller.Run(stopCh)
	return podLister
}

// NewPodListerAndOOMObserver creates pair of pod lister and OOM observer.
func NewPodListerAndOOMObserver(kubeClient kube_client.Interface, namespace string) (v1lister.PodLister, oom.Observer) {
	oomObserver := oom.NewObserver()
	podLister := newPodClients(kubeClient, oomObserver, namespace)
	WatchEvictionEventsWithRetries(kubeClient, oomObserver, namespace)
	return podLister, oomObserver
}

type clusterStateFeeder struct {
	coreClient          corev1.CoreV1Interface
	specClient          spec.SpecClient
	metricsClient       metrics.MetricsClient
	oomChan             <-chan oom.OomInfo
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	vpaLister           vpa_lister.VerticalPodAutoscalerLister
	clusterState        *model.ClusterState
	selectorFetcher     target.VpaTargetSelectorFetcher
	memorySaveMode      bool
	controllerFetcher   controllerfetcher.ControllerFetcher
	recommenderName     string
	ignoredNamespaces   []string
}

func (feeder *clusterStateFeeder) InitFromHistoryProvider(historyProvider history.HistoryProvider) {
	klog.V(3).InfoS("Initializing VPA from history provider")
	clusterHistory, err := historyProvider.GetClusterHistory()
	if err != nil {
		klog.ErrorS(err, "Cannot get cluster history")
	}
	for podID, podHistory := range clusterHistory {
		klog.V(4).InfoS("Adding pod with labels", "pod", podID, "labels", podHistory.LastLabels)
		feeder.clusterState.AddOrUpdatePod(podID, podHistory.LastLabels, apiv1.PodUnknown)
		for containerName, sampleList := range podHistory.Samples {
			containerID := model.ContainerID{
				PodID:         podID,
				ContainerName: containerName,
			}
			if err = feeder.clusterState.AddOrUpdateContainer(containerID, nil); err != nil {
				klog.V(0).InfoS("Failed to add container", "container", containerID, "error", err)
			}
			klog.V(4).InfoS("Adding samples for container", "sampleCount", len(sampleList), "container", containerID)
			for _, sample := range sampleList {
				if err := feeder.clusterState.AddSample(
					&model.ContainerUsageSampleWithKey{
						ContainerUsageSample: sample,
						Container:            containerID,
					}); err != nil {
					klog.V(0).InfoS("Failed to add sample", "sample", sample, "error", err)
				}
			}
		}
	}
}

func (feeder *clusterStateFeeder) setVpaCheckpoint(checkpoint *vpa_types.VerticalPodAutoscalerCheckpoint) error {
	vpaID := model.VpaID{Namespace: checkpoint.Namespace, VpaName: checkpoint.Spec.VPAObjectName}
	vpa, exists := feeder.clusterState.Vpas[vpaID]
	if !exists {
		return fmt.Errorf("cannot load checkpoint to missing VPA object %s/%s", vpaID.Namespace, vpaID.VpaName)
	}

	cs := model.NewAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint.Status)
	if err != nil {
		return fmt.Errorf("cannot load checkpoint for VPA %s/%s. Reason: %v", vpaID.Namespace, vpaID.VpaName, err)
	}
	vpa.ContainersInitialAggregateState[checkpoint.Spec.ContainerName] = cs
	return nil
}

func (feeder *clusterStateFeeder) InitFromCheckpoints() {
	klog.V(3).InfoS("Initializing VPA from checkpoints")
	feeder.LoadVPAs(context.TODO())

	namespaces := make(map[string]bool)
	for _, v := range feeder.clusterState.Vpas {
		namespaces[v.ID.Namespace] = true
	}

	for namespace := range namespaces {
		klog.V(3).InfoS("Fetching checkpoints", "namespace", namespace)
		checkpointList, err := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.ErrorS(err, "Cannot list VPA checkpoints", "namespace", namespace)
		}
		for _, checkpoint := range checkpointList.Items {

			klog.V(3).InfoS("Loading checkpoint for VPA", klog.KRef(checkpoint.ObjectMeta.Namespace, checkpoint.Spec.VPAObjectName), "container", checkpoint.Spec.ContainerName)
			err = feeder.setVpaCheckpoint(&checkpoint)
			if err != nil {
				klog.ErrorS(err, "Error while loading checkpoint")
			}

		}
	}
}

func (feeder *clusterStateFeeder) GarbageCollectCheckpoints() {
	klog.V(3).InfoS("Starting garbage collection of checkpoints")
	feeder.LoadVPAs(context.TODO())

	namespaceList, err := feeder.coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.ErrorS(err, "Cannot list namespaces")
		return
	}

	for _, namespaceItem := range namespaceList.Items {
		namespace := namespaceItem.Name
		checkpointList, err := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.ErrorS(err, "Cannot list VPA checkpoints", "namespace", namespace)
		}
		for _, checkpoint := range checkpointList.Items {
			vpaID := model.VpaID{Namespace: checkpoint.Namespace, VpaName: checkpoint.Spec.VPAObjectName}
			vpa, exists := feeder.clusterState.Vpas[vpaID]
			if !exists {
				err = feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).Delete(context.TODO(), checkpoint.Name, metav1.DeleteOptions{})
				if err == nil {
					klog.V(3).InfoS("Orphaned VPA checkpoint cleanup - deleting", klog.KRef(namespace, checkpoint.Name))
				} else {
					klog.ErrorS(err, "Orphaned VPA checkpoint cleanup - error deleting", klog.KRef(namespace, checkpoint.Name))
				}
			}
			// Also clean up a checkpoint if the VPA is still there, but the container is gone. AggregateStateByContainerName
			// merges in the initial aggregates so we can use it to check "both lists" (initial, aggregates) at once
			// TODO(jkyros): could we also just wait until it got "old" enough, e.g. the checkpoint hasn't
			// been updated for an hour, blow it away? Because once we remove it from the aggregate lists, it will stop
			// being maintained.
			_, aggregateExists := vpa.AggregateStateByContainerName()[checkpoint.Spec.ContainerName]
			if !aggregateExists {
				err = feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).Delete(context.TODO(), checkpoint.Name, metav1.DeleteOptions{})
				if err == nil {
					klog.V(3).Infof("Orphaned VPA checkpoint cleanup - deleting %v/%v.", namespace, checkpoint.Name)
				} else {
					klog.Errorf("Cannot delete VPA checkpoint %v/%v. Reason: %+v", namespace, checkpoint.Name, err)
				}
			}
		}
	}
}

func implicitDefaultRecommender(selectors []*vpa_types.VerticalPodAutoscalerRecommenderSelector) bool {
	return len(selectors) == 0
}

func selectsRecommender(selectors []*vpa_types.VerticalPodAutoscalerRecommenderSelector, name *string) bool {
	for _, s := range selectors {
		if s.Name == *name {
			return true
		}
	}
	return false
}

// Filter VPA objects whose specified recommender names are not default
func filterVPAs(feeder *clusterStateFeeder, allVpaCRDs []*vpa_types.VerticalPodAutoscaler) []*vpa_types.VerticalPodAutoscaler {
	klog.V(3).InfoS("Start selecting the vpaCRDs.")
	var vpaCRDs []*vpa_types.VerticalPodAutoscaler
	for _, vpaCRD := range allVpaCRDs {
		if feeder.recommenderName == DefaultRecommenderName {
			if !implicitDefaultRecommender(vpaCRD.Spec.Recommenders) && !selectsRecommender(vpaCRD.Spec.Recommenders, &feeder.recommenderName) {
				klog.V(6).InfoS("Ignoring vpaCRD as current recommender's name doesn't appear among its recommenders", "vpaCRD", klog.KObj(vpaCRD), "recommenderName", feeder.recommenderName)
				continue
			}
		} else {
			if implicitDefaultRecommender(vpaCRD.Spec.Recommenders) {
				klog.V(6).InfoS("Ignoring vpaCRD as recommender doesn't process CRDs implicitly destined to default recommender", "vpaCRD", klog.KObj(vpaCRD), "recommenderName", feeder.recommenderName, "defaultRecommenderName", DefaultRecommenderName)
				continue
			}
			if !selectsRecommender(vpaCRD.Spec.Recommenders, &feeder.recommenderName) {
				klog.V(6).InfoS("Ignoring vpaCRD as current recommender's name doesn't appear among its recommenders", "vpaCRD", klog.KObj(vpaCRD), "recommenderName", feeder.recommenderName)
				continue
			}
		}

		if slices.Contains(feeder.ignoredNamespaces, vpaCRD.ObjectMeta.Namespace) {
			klog.V(6).InfoS("Ignoring vpaCRD as this namespace is ignored", "vpaCRD", klog.KObj(vpaCRD))
			continue
		}

		vpaCRDs = append(vpaCRDs, vpaCRD)
	}
	return vpaCRDs
}

// LoadVPAs fetches VPA objects and loads them into the cluster state.
func (feeder *clusterStateFeeder) LoadVPAs(ctx context.Context) {
	// List VPA API objects.
	allVpaCRDs, err := feeder.vpaLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Cannot list VPAs")
		return
	}

	// Filter out VPAs that specified recommenders with names not equal to "default"
	vpaCRDs := filterVPAs(feeder, allVpaCRDs)

	klog.V(3).InfoS("Fetching VPAs", "count", len(vpaCRDs))
	// Add or update existing VPAs in the model.
	vpaKeys := make(map[model.VpaID]bool)
	for _, vpaCRD := range vpaCRDs {
		vpaID := model.VpaID{
			Namespace: vpaCRD.Namespace,
			VpaName:   vpaCRD.Name,
		}

		selector, conditions := feeder.getSelector(ctx, vpaCRD)
		klog.V(4).InfoS("Using selector", "selector", selector.String(), "vpa", klog.KObj(vpaCRD))

		if feeder.clusterState.AddOrUpdateVpa(vpaCRD, selector) == nil {
			// Successfully added VPA to the model.
			vpaKeys[vpaID] = true

			for _, condition := range conditions {
				if condition.delete {
					delete(feeder.clusterState.Vpas[vpaID].Conditions, condition.conditionType)
				} else {
					feeder.clusterState.Vpas[vpaID].Conditions.Set(condition.conditionType, true, "", condition.message)
				}
			}
		}
	}
	// Delete non-existent VPAs from the model.
	for vpaID := range feeder.clusterState.Vpas {
		if _, exists := vpaKeys[vpaID]; !exists {
			klog.V(3).InfoS("Deleting VPA", "vpa", klog.KRef(vpaID.Namespace, vpaID.VpaName))
			if err := feeder.clusterState.DeleteVpa(vpaID); err != nil {
				klog.ErrorS(err, "Deleting VPA failed", "vpa", klog.KRef(vpaID.Namespace, vpaID.VpaName))
			}
		}
	}
	feeder.clusterState.ObservedVpas = vpaCRDs
}

// MarkAggregates marks all aggregates IsUnderVPA=false, so when we go
// through LoadPods(), the valid ones will get marked back to true, and
// we can garbage collect the false ones from the VPAs' aggregate lists.
func (feeder *clusterStateFeeder) MarkAggregates() {
	for _, vpa := range feeder.clusterState.Vpas {
		for _, container := range vpa.AggregateContainerStates() {
			container.IsUnderVPA = false
		}
		for _, container := range vpa.ContainersInitialAggregateState {
			container.IsUnderVPA = false
		}
	}
}

// SweepAggregates garbage collects all aggregates/initial aggregates from the VPA where the
// aggregate's container no longer exists.
func (feeder *clusterStateFeeder) SweepAggregates() {

	var aggregatesPruned int
	var initialAggregatesPruned int

	// TODO(jkyros): This only removes the container state from the VPA's aggregate states, there
	// is still a reference to them in feeder.clusterState.aggregateStateMap, and those get
	// garbage collected eventually by the rate limited aggregate garbage collector later.
	// Maybe we should clean those up here too since we know which ones are stale?
	for _, vpa := range feeder.clusterState.Vpas {

		for containerKey, container := range vpa.AggregateContainerStates() {
			if !container.IsUnderVPA {
				klog.V(4).Infof("Deleting Aggregate for VPA %s/%s: container %s no longer present", vpa.ID.Namespace, vpa.ID.VpaName, containerKey.ContainerName())
				vpa.DeleteAggregation(containerKey)
				aggregatesPruned = aggregatesPruned + 1

			}
		}
		for containerKey, container := range vpa.ContainersInitialAggregateState {
			if !container.IsUnderVPA {
				klog.V(4).Infof("Deleting Initial Aggregate for VPA %s/%s: container %s no longer present", vpa.ID.Namespace, vpa.ID.VpaName, containerKey)
				delete(vpa.ContainersInitialAggregateState, containerKey)
				initialAggregatesPruned = initialAggregatesPruned + 1

			}
		}
	}
	if initialAggregatesPruned > 0 || aggregatesPruned > 0 {
		klog.Infof("Pruned %d aggregate and %d initial aggregate containers", aggregatesPruned, initialAggregatesPruned)
	}
}

// LoadPods loads pod into the cluster state.
func (feeder *clusterStateFeeder) LoadPods() {
	podSpecs, err := feeder.specClient.GetPodSpecs()
	if err != nil {
		klog.ErrorS(err, "Cannot get SimplePodSpecs")
	}
	pods := make(map[model.PodID]*spec.BasicPodSpec)
	for _, spec := range podSpecs {
		pods[spec.ID] = spec
	}
	for key := range feeder.clusterState.Pods {
		if _, exists := pods[key]; !exists {
			klog.V(3).InfoS("Deleting Pod", "pod", klog.KRef(key.Namespace, key.PodName))
			feeder.clusterState.DeletePod(key)
		}
	}
	for _, pod := range pods {
		if feeder.memorySaveMode && !feeder.matchesVPA(pod) {
			continue
		}
		feeder.clusterState.AddOrUpdatePod(pod.ID, pod.PodLabels, pod.Phase)
		for _, container := range pod.Containers {
			if err = feeder.clusterState.AddOrUpdateContainer(container.ID, container.Request); err != nil {
				klog.V(0).InfoS("Failed to add container", "container", container.ID, "error", err)
			}
		}
	}
}

func (feeder *clusterStateFeeder) LoadRealTimeMetrics() {
	containersMetrics, err := feeder.metricsClient.GetContainersMetrics()
	if err != nil {
		klog.ErrorS(err, "Cannot get ContainerMetricsSnapshot from MetricsClient")
	}

	sampleCount := 0
	droppedSampleCount := 0
	for _, containerMetrics := range containersMetrics {
		for _, sample := range newContainerUsageSamplesWithKey(containerMetrics) {
			if err := feeder.clusterState.AddSample(sample); err != nil {
				// Not all pod states are tracked in memory saver mode
				if _, isKeyError := err.(model.KeyError); isKeyError && feeder.memorySaveMode {
					continue
				}
				klog.V(0).InfoS("Error adding metric sample", "sample", sample, "error", err)
				droppedSampleCount++
			} else {
				sampleCount++
			}
		}
	}
	klog.V(3).InfoS("ClusterSpec fed with ContainerUsageSamples", "sampleCount", sampleCount, "containerCount", len(containersMetrics), "droppedSampleCount", droppedSampleCount)
Loop:
	for {
		select {
		case oomInfo := <-feeder.oomChan:
			klog.V(3).InfoS("OOM detected", "oomInfo", oomInfo)
			if err = feeder.clusterState.RecordOOM(oomInfo.ContainerID, oomInfo.Timestamp, oomInfo.Memory); err != nil {
				klog.V(0).InfoS("Failed to record OOM", "oomInfo", oomInfo, "error", err)
			}
		default:
			break Loop
		}
	}
	metrics_recommender.RecordAggregateContainerStatesCount(feeder.clusterState.StateMapSize())
}

func (feeder *clusterStateFeeder) matchesVPA(pod *spec.BasicPodSpec) bool {
	for vpaKey, vpa := range feeder.clusterState.Vpas {
		podLabels := labels.Set(pod.PodLabels)
		if vpaKey.Namespace == pod.ID.Namespace && vpa.PodSelector.Matches(podLabels) {
			return true
		}
	}
	return false
}

func newContainerUsageSamplesWithKey(metrics *metrics.ContainerMetricsSnapshot) []*model.ContainerUsageSampleWithKey {
	var samples []*model.ContainerUsageSampleWithKey

	for metricName, resourceAmount := range metrics.Usage {
		sample := &model.ContainerUsageSampleWithKey{
			Container: metrics.ID,
			ContainerUsageSample: model.ContainerUsageSample{
				MeasureStart: metrics.SnapshotTime,
				Resource:     metricName,
				Usage:        resourceAmount,
			},
		}
		samples = append(samples, sample)
	}
	return samples
}

type condition struct {
	conditionType vpa_types.VerticalPodAutoscalerConditionType
	delete        bool
	message       string
}

func (feeder *clusterStateFeeder) validateTargetRef(ctx context.Context, vpa *vpa_types.VerticalPodAutoscaler) (bool, condition) {
	//
	if vpa.Spec.TargetRef == nil {
		return false, condition{}
	}
	k := controllerfetcher.ControllerKeyWithAPIVersion{
		ControllerKey: controllerfetcher.ControllerKey{
			Namespace: vpa.Namespace,
			Kind:      vpa.Spec.TargetRef.Kind,
			Name:      vpa.Spec.TargetRef.Name,
		},
		ApiVersion: vpa.Spec.TargetRef.APIVersion,
	}
	top, err := feeder.controllerFetcher.FindTopMostWellKnownOrScalable(ctx, &k)
	if err != nil {
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: fmt.Sprintf("Error checking if target is a topmost well-known or scalable controller: %s", err)}
	}
	if top == nil {
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: fmt.Sprintf("Unknown error during checking if target is a topmost well-known or scalable controller: %s", err)}
	}
	if *top != k {
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: "The targetRef controller has a parent but it should point to a topmost well-known or scalable controller"}
	}
	return true, condition{}
}

func (feeder *clusterStateFeeder) getSelector(ctx context.Context, vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, []condition) {
	selector, fetchErr := feeder.selectorFetcher.Fetch(ctx, vpa)
	if selector != nil {
		validTargetRef, unsupportedCondition := feeder.validateTargetRef(ctx, vpa)
		if !validTargetRef {
			return labels.Nothing(), []condition{
				unsupportedCondition,
				{conditionType: vpa_types.ConfigDeprecated, delete: true},
			}
		}
		return selector, []condition{
			{conditionType: vpa_types.ConfigUnsupported, delete: true},
			{conditionType: vpa_types.ConfigDeprecated, delete: true},
		}
	}
	msg := "Cannot read targetRef"
	if fetchErr != nil {
		klog.ErrorS(fetchErr, "Cannot get target selector from VPA's targetRef")
		msg = fmt.Sprintf("Cannot read targetRef. Reason: %s", fetchErr.Error())
	}
	return labels.Nothing(), []condition{
		{conditionType: vpa_types.ConfigUnsupported, delete: false, message: msg},
		{conditionType: vpa_types.ConfigDeprecated, delete: true},
	}
}
