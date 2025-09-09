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

// ClusterStateFeeder can update state of clusterState object.
type ClusterStateFeeder interface {
	// InitFromHistoryProvider loads historical pod spec into clusterState.
	InitFromHistoryProvider(historyProvider history.HistoryProvider)

	// InitFromCheckpoints loads historical checkpoints into clusterState.
	InitFromCheckpoints(ctx context.Context)

	// LoadVPAs updates clusterState with current state of VPAs.
	LoadVPAs(ctx context.Context)

	// LoadPods updates clusterState with current specification of Pods and their Containers.
	LoadPods()

	// LoadRealTimeMetrics updates clusterState with current usage metrics of containers.
	LoadRealTimeMetrics(ctx context.Context)

	// GarbageCollectCheckpoints removes historical checkpoints that don't have a matching VPA.
	GarbageCollectCheckpoints(ctx context.Context)
}

// ClusterStateFeederFactory makes instances of ClusterStateFeeder.
type ClusterStateFeederFactory struct {
	ClusterState        model.ClusterState
	KubeClient          kube_client.Interface
	MetricsClient       metrics.MetricsClient
	VpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	VpaCheckpointLister vpa_lister.VerticalPodAutoscalerCheckpointLister
	VpaLister           vpa_lister.VerticalPodAutoscalerLister
	PodLister           v1lister.PodLister
	OOMObserver         oom.Observer
	SelectorFetcher     target.VpaTargetSelectorFetcher
	MemorySaveMode      bool
	ControllerFetcher   controllerfetcher.ControllerFetcher
	RecommenderName     string
	IgnoredNamespaces   []string
	VpaObjectNamespace  string
}

// Make creates new ClusterStateFeeder with internal data providers, based on kube client.
func (m ClusterStateFeederFactory) Make() *clusterStateFeeder {
	return &clusterStateFeeder{
		coreClient:          m.KubeClient.CoreV1(),
		metricsClient:       m.MetricsClient,
		oomChan:             m.OOMObserver.GetObservedOomsChannel(),
		vpaCheckpointClient: m.VpaCheckpointClient,
		vpaCheckpointLister: m.VpaCheckpointLister,
		vpaLister:           m.VpaLister,
		clusterState:        m.ClusterState,
		specClient:          spec.NewSpecClient(m.PodLister),
		selectorFetcher:     m.SelectorFetcher,
		memorySaveMode:      m.MemorySaveMode,
		controllerFetcher:   m.ControllerFetcher,
		recommenderName:     m.RecommenderName,
		ignoredNamespaces:   m.IgnoredNamespaces,
		vpaObjectNamespace:  m.VpaObjectNamespace,
	}
}

// WatchEvictionEventsWithRetries watches new Events with reason=Evicted and passes them to the observer.
func WatchEvictionEventsWithRetries(ctx context.Context, kubeClient kube_client.Interface, observer oom.Observer, namespace string) {
	go func() {
		options := metav1.ListOptions{
			FieldSelector: "reason=Evicted",
		}

		watchEvictionEventsOnce := func() {
			watchInterface, err := kubeClient.CoreV1().Events(namespace).Watch(ctx, options)
			if err != nil {
				klog.ErrorS(err, "Cannot initialize watching events")
				return
			}
			defer watchInterface.Stop()
			watchEvictionEvents(watchInterface.ResultChan(), observer)
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
				watchEvictionEventsOnce()
				// Wait between attempts, retrying too often breaks API server.
				waitTime := wait.Jitter(evictionWatchRetryWait, evictionWatchJitterFactor)
				klog.V(1).InfoS("An attempt to watch eviction events finished", "waitTime", waitTime)
				time.Sleep(waitTime)
			}
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
func newPodClients(kubeClient kube_client.Interface, resourceEventHandler cache.ResourceEventHandler, namespace string, stopCh <-chan struct{}) v1lister.PodLister {
	// We are interested in pods which are Running or Unknown (in case the pod is
	// running but there are some transient errors we don't want to delete it from
	// our model).
	// We don't want to watch Pending pods because they didn't generate any usage
	// yet.
	// Succeeded and Failed failed pods don't generate any usage anymore but we
	// don't necessarily want to immediately delete them.
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(apiv1.PodPending))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	informerOptions := cache.InformerOptions{
		ObjectType:    &apiv1.Pod{},
		ListerWatcher: podListWatch,
		Handler:       resourceEventHandler,
		ResyncPeriod:  time.Hour,
		Indexers:      cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	}

	store, controller := cache.NewInformerWithOptions(informerOptions)
	indexer, ok := store.(cache.Indexer)
	if !ok {
		klog.ErrorS(nil, "Expected Indexer, but got a Store that does not implement Indexer")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	podLister := v1lister.NewPodLister(indexer)
	go controller.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, controller.HasSynced) {
		klog.ErrorS(nil, "Failed to sync Pod cache during initialization")
	}
	return podLister
}

// NewPodListerAndOOMObserver creates pair of pod lister and OOM observer.
func NewPodListerAndOOMObserver(ctx context.Context, kubeClient kube_client.Interface, namespace string, stopCh <-chan struct{}) (v1lister.PodLister, oom.Observer) {
	oomObserver := oom.NewObserver()
	podLister := newPodClients(kubeClient, oomObserver, namespace, stopCh)
	WatchEvictionEventsWithRetries(ctx, kubeClient, oomObserver, namespace)
	return podLister, oomObserver
}

type clusterStateFeeder struct {
	coreClient          corev1.CoreV1Interface
	specClient          spec.SpecClient
	metricsClient       metrics.MetricsClient
	oomChan             <-chan oom.OomInfo
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	vpaCheckpointLister vpa_lister.VerticalPodAutoscalerCheckpointLister
	vpaLister           vpa_lister.VerticalPodAutoscalerLister
	clusterState        model.ClusterState
	selectorFetcher     target.VpaTargetSelectorFetcher
	memorySaveMode      bool
	controllerFetcher   controllerfetcher.ControllerFetcher
	recommenderName     string
	ignoredNamespaces   []string
	vpaObjectNamespace  string
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
	vpa, exists := feeder.clusterState.VPAs()[vpaID]
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

func (feeder *clusterStateFeeder) InitFromCheckpoints(ctx context.Context) {
	klog.V(3).InfoS("Initializing VPA from checkpoints")
	feeder.LoadVPAs(ctx)

	klog.V(3).InfoS("Fetching VPA checkpoints")
	checkpointList, err := feeder.vpaCheckpointLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Cannot list VPA checkpoints")
	}

	namespaces := make(map[string]bool)
	for _, v := range feeder.clusterState.VPAs() {
		namespaces[v.ID.Namespace] = true
	}

	for namespace := range namespaces {
		if feeder.shouldIgnoreNamespace(namespace) {
			klog.V(3).InfoS("Skipping loading VPA Checkpoints from namespace.", "namespace", namespace, "vpaObjectNamespace", feeder.vpaObjectNamespace, "ignoredNamespaces", feeder.ignoredNamespaces)
			continue
		}

		for _, checkpoint := range checkpointList {
			klog.V(3).InfoS("Loading checkpoint for VPA", "checkpoint", klog.KRef(checkpoint.Namespace, checkpoint.Spec.VPAObjectName), "container", checkpoint.Spec.ContainerName)
			err = feeder.setVpaCheckpoint(checkpoint)
			if err != nil {
				klog.ErrorS(err, "Error while loading checkpoint")
			}
		}
	}
}

func (feeder *clusterStateFeeder) GarbageCollectCheckpoints(ctx context.Context) {
	klog.V(3).InfoS("Starting garbage collection of checkpoints")

	allVPAKeys := map[model.VpaID]bool{}

	allVpaResources, err := feeder.vpaLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Cannot list VPAs")
		return
	}
	for _, vpa := range allVpaResources {
		vpaID := model.VpaID{
			Namespace: vpa.Namespace,
			VpaName:   vpa.Name,
		}
		allVPAKeys[vpaID] = true
	}

	namespaceList, err := feeder.coreClient.Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.ErrorS(err, "Cannot list namespaces")
		return
	}

	for _, namespaceItem := range namespaceList.Items {
		namespace := namespaceItem.Name
		// Clean the namespace if any of the following conditions are true:
		// 1. `vpaObjectNamespace` is set and matches the current namespace.
		// 2. `ignoredNamespaces` is set, but the current namespace is not in the list.
		// 3. Neither `vpaObjectNamespace` nor `ignoredNamespaces` is set, so all namespaces are included.
		if feeder.shouldIgnoreNamespace(namespace) {
			klog.V(3).InfoS("Skipping namespace; it does not meet cleanup criteria", "namespace", namespace, "vpaObjectNamespace", feeder.vpaObjectNamespace, "ignoredNamespaces", feeder.ignoredNamespaces)
			continue
		}
		err := feeder.cleanupCheckpointsForNamespace(ctx, namespace, allVPAKeys)
		if err != nil {
			klog.ErrorS(err, "error cleanining checkpoints")
		}
	}
}

func (feeder *clusterStateFeeder) shouldIgnoreNamespace(namespace string) bool {
	// 1. `vpaObjectNamespace` is set but doesn't match the current namespace.
	if feeder.vpaObjectNamespace != "" && namespace != feeder.vpaObjectNamespace {
		return true
	}
	// 2. `ignoredNamespaces` is set, and the current namespace is in the list.
	if len(feeder.ignoredNamespaces) > 0 && slices.Contains(feeder.ignoredNamespaces, namespace) {
		return true
	}
	return false
}

func (feeder *clusterStateFeeder) cleanupCheckpointsForNamespace(ctx context.Context, namespace string, allVPAKeys map[model.VpaID]bool) error {
	var err error
	checkpointList, err := feeder.vpaCheckpointLister.VerticalPodAutoscalerCheckpoints(namespace).List(labels.Everything())

	if err != nil {
		return err
	}
	for _, checkpoint := range checkpointList {
		vpaID := model.VpaID{Namespace: checkpoint.Namespace, VpaName: checkpoint.Spec.VPAObjectName}
		if !allVPAKeys[vpaID] {
			if errFeeder := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).Delete(ctx, checkpoint.Name, metav1.DeleteOptions{}); errFeeder != nil {
				err = fmt.Errorf("failed to delete orphaned checkpoint %s: %w", klog.KRef(namespace, checkpoint.Name), err)
				continue
			}
			klog.V(3).InfoS("Orphaned VPA checkpoint cleanup - deleting", "checkpoint", klog.KRef(namespace, checkpoint.Name))
		}
	}
	return err
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

		if feeder.shouldIgnoreNamespace(vpaCRD.Namespace) {
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
					delete(feeder.clusterState.VPAs()[vpaID].Conditions, condition.conditionType)
				} else {
					feeder.clusterState.VPAs()[vpaID].Conditions.Set(condition.conditionType, true, "", condition.message)
				}
			}
		}
	}
	// Delete non-existent VPAs from the model.
	for vpaID := range feeder.clusterState.VPAs() {
		if _, exists := vpaKeys[vpaID]; !exists {
			klog.V(3).InfoS("Deleting VPA", "vpa", klog.KRef(vpaID.Namespace, vpaID.VpaName))
			if err := feeder.clusterState.DeleteVpa(vpaID); err != nil {
				klog.ErrorS(err, "Deleting VPA failed", "vpa", klog.KRef(vpaID.Namespace, vpaID.VpaName))
			}
		}
	}
	feeder.clusterState.SetObservedVPAs(vpaCRDs)
}

// LoadPods loads pod into the cluster state.
func (feeder *clusterStateFeeder) LoadPods() {
	var podSpecs []*spec.BasicPodSpec
	var err error

	// If memory save mode is enabled and we have VPAs with podLabelSelector,
	// we can optimize by fetching pods using the specific selectors instead of getting all pods
	if feeder.memorySaveMode && feeder.canUseSelectorBasedPodFetching() {
		podSpecs, err = feeder.getPodSpecsWithSelectors()
	} else {
		podSpecs, err = feeder.specClient.GetPodSpecs()
	}

	if err != nil {
		klog.ErrorS(err, "Cannot get SimplePodSpecs")
	}
	pods := make(map[model.PodID]*spec.BasicPodSpec)
	for _, spec := range podSpecs {
		pods[spec.ID] = spec
	}
	for key := range feeder.clusterState.Pods() {
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
		for _, initContainer := range pod.InitContainers {
			podInitContainers := feeder.clusterState.Pods()[pod.ID].InitContainers
			feeder.clusterState.Pods()[pod.ID].InitContainers = append(podInitContainers, initContainer.ID.ContainerName)

		}
	}
}

func (feeder *clusterStateFeeder) LoadRealTimeMetrics(ctx context.Context) {
	containersMetrics, err := feeder.metricsClient.GetContainersMetrics(ctx)
	if err != nil {
		klog.ErrorS(err, "Cannot get ContainerMetricsSnapshot from MetricsClient")
	}

	sampleCount := 0
	droppedSampleCount := 0
	for _, containerMetrics := range containersMetrics {
		// Container metrics are fetched for all pods, however, not all pod states are tracked in memory saver mode.
		if pod, exists := feeder.clusterState.Pods()[containerMetrics.ID.PodID]; exists && pod != nil {
			if slices.Contains(pod.InitContainers, containerMetrics.ID.ContainerName) {
				klog.V(3).InfoS("Skipping metric samples for init container", "pod", klog.KRef(containerMetrics.ID.Namespace, containerMetrics.ID.PodName), "container", containerMetrics.ID.ContainerName)
				droppedSampleCount += len(containerMetrics.Usage)
				continue
			}
		}
		for _, sample := range newContainerUsageSamplesWithKey(containerMetrics) {
			if err := feeder.clusterState.AddSample(sample); err != nil {
				// Not all pod states are tracked in memory saver mode.
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
	for vpaKey, vpa := range feeder.clusterState.VPAs() {
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
	// If there's no targetRef but podLabelSelector is used, that's valid
	if vpa.Spec.TargetRef == nil {
		if vpa.Spec.PodLabelSelector != nil {
			return true, condition{} // podLabelSelector is being used, no targetRef validation needed
		}
		return false, condition{}
	}

	target := fmt.Sprintf("%s.%s/%s", vpa.Spec.TargetRef.APIVersion, vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name)

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
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: fmt.Sprintf("Error checking if target %s is a topmost well-known or scalable controller: %s", target, err)}
	}
	if top == nil {
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: fmt.Sprintf("Unknown error during checking if target %s is a topmost well-known or scalable controller", target)}
	}
	if *top != k {
		return false, condition{conditionType: vpa_types.ConfigUnsupported, delete: false, message: fmt.Sprintf("The target %s has a parent controller but it should point to a topmost well-known or scalable controller", target)}
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

// canUseSelectorBasedPodFetching returns true if we can optimize pod fetching by using
// VPA selectors directly instead of fetching all pods and filtering in memory.
func (feeder *clusterStateFeeder) canUseSelectorBasedPodFetching() bool {
	// Only optimize if we have VPAs that can benefit from selector-based fetching
	for _, vpa := range feeder.clusterState.VPAs() {
		if vpa.PodSelector != nil && vpa.PodSelector != labels.Nothing() {
			return true
		}
	}
	return false
}

// getPodSpecsWithSelectors fetches pods using VPA selectors for optimization.
// This reduces memory usage by only fetching pods that are relevant to the VPAs.
func (feeder *clusterStateFeeder) getPodSpecsWithSelectors() ([]*spec.BasicPodSpec, error) {
	podsMap := make(map[model.PodID]*spec.BasicPodSpec)

	for _, vpa := range feeder.clusterState.VPAs() {
		if vpa.PodSelector == nil || vpa.PodSelector == labels.Nothing() {
			continue
		}

		// Skip VPA if it's in an ignored namespace
		if feeder.shouldIgnoreNamespace(vpa.ID.Namespace) {
			continue
		}

		// Use the spec client with the VPA's selector to get only relevant pods
		podSpecs, err := feeder.specClient.GetPodSpecsWithSelector(vpa.PodSelector)
		if err != nil {
			klog.V(2).InfoS("Error listing pods with selector", "vpa", vpa.ID, "selector", vpa.PodSelector.String(), "error", err)
			continue
		}

		for _, podSpec := range podSpecs {
			if !feeder.shouldIgnoreNamespace(podSpec.ID.Namespace) {
				podsMap[podSpec.ID] = podSpec
			}
		}
	}

	// Convert map to slice
	var podSpecs []*spec.BasicPodSpec
	for _, podSpec := range podsMap {
		podSpecs = append(podSpecs, podSpec)
	}

	return podSpecs, nil
}
