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

// HistorySource is an enum type for history source
type HistorySource int

const (
	// Checkpoints is a history source that uses VPA checkpoints custom resources
	Checkpoints HistorySource = iota
	// Prometheus is a history source that uses prometheus
	Prometheus
	// None is a history source that doesn't use any history
	None
	// Undefined is a history source that is not set
	Undefined
)

// ClusterStateFeeder can update state of ClusterState object.
type ClusterStateFeeder interface {
	// Init initializes ClusterStateFeeder
	Init() error

	// LoadVPAs updates clusterState with current state of VPAs.
	LoadVPAs(ctx context.Context)

	// LoadPods updates clusterState with current specification of Pods and their Containers.
	LoadPods()

	// LoadRealTimeMetrics updates clusterState with current usage metrics of containers.
	LoadRealTimeMetrics()

	// GarbageCollectCheckpoints removes historical checkpoints that don't have a matching VPA.
	GarbageCollectCheckpoints()
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
	HistorySource       HistorySource
	PromHistoryConfig   history.PrometheusHistoryProviderConfig
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
		historySource:       m.HistorySource,
		promHistoryConfig:   m.PromHistoryConfig,
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
				klog.Errorf("Cannot initialize watching events. Reason %v", err)
				return
			}
			watchEvictionEvents(watchInterface.ResultChan(), observer)
		}
		for {
			watchEvictionEventsOnce()
			// Wait between attempts, retrying too often breaks API server.
			waitTime := wait.Jitter(evictionWatchRetryWait, evictionWatchJitterFactor)
			klog.V(1).Infof("An attempt to watch eviction events finished. Waiting %v before the next one.", waitTime)
			time.Sleep(waitTime)
		}
	}()
}

func watchEvictionEvents(evictedEventChan <-chan watch.Event, observer oom.Observer) {
	for {
		evictedEvent, ok := <-evictedEventChan
		if !ok {
			klog.V(3).Infof("Eviction event chan closed")
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
	historySource       HistorySource
	promHistoryConfig   history.PrometheusHistoryProviderConfig
	ignoredNamespaces   []string
}

func (feeder *clusterStateFeeder) Init() error {
	switch feeder.historySource {
	case Checkpoints:
		klog.Infof("Using checkpoints as a history provider")
		feeder.initFromCheckpoints()
	case Prometheus:
		klog.Infof("Using prometheus as a history provider")

		provider, promInitErr := history.NewPrometheusHistoryProvider(feeder.promHistoryConfig)
		if promInitErr != nil {
			klog.Errorf("Could not initialize history provider")
			return promInitErr
		}

		historyInitErr := feeder.initFromHistoryProvider(provider)
		if historyInitErr != nil {
			klog.Errorf("Failed to load prometheus history")
			return historyInitErr
		}
	case None:
		klog.Infof("Running without a history provider")
	default:
		klog.Errorf("Wrong history provider option")
		return fmt.Errorf("history provider option is not set. Supported values: prometheus, none, checkpoint")
	}
	return nil
}

// GetHistorySourceFromArg reads the history source from the command line argument and returns the corresponding enum value
func GetHistorySourceFromArg(historySource string) (HistorySource, error) {
	switch historySource {
	case "checkpoint":
		return Checkpoints, nil
	case "prometheus":
		return Prometheus, nil
	case "none":
		return None, nil
	default:
		return Undefined, fmt.Errorf("storage option '%s' is not supported. Supported values: prometheus, none, checkpoint", historySource)
	}
}

func (feeder *clusterStateFeeder) initFromHistoryProvider(historyProvider history.HistoryProvider) error {
	klog.V(3).Info("Initializing VPA from history provider")
	clusterHistory, err := historyProvider.GetClusterHistory()
	if err != nil {
		return err
	}
	if len(clusterHistory) == 0 {
		klog.Warningf("history provider returned no pods")
	}
	for podID, podHistory := range clusterHistory {
		klog.V(4).Infof("Adding pod %v with labels %v", podID, podHistory.LastLabels)
		feeder.clusterState.AddOrUpdatePod(podID, podHistory.LastLabels, apiv1.PodUnknown)
		for containerName, sampleList := range podHistory.Samples {
			containerID := model.ContainerID{
				PodID:         podID,
				ContainerName: containerName,
			}
			if err = feeder.clusterState.AddOrUpdateContainer(containerID, nil); err != nil {
				klog.Warningf("Failed to add container %+v. Reason: %+v", containerID, err)
			}
			klog.V(4).Infof("Adding %d samples for container %v", len(sampleList), containerID)
			for _, sample := range sampleList {
				if err := feeder.clusterState.AddSample(
					&model.ContainerUsageSampleWithKey{
						ContainerUsageSample: sample,
						Container:            containerID,
					}); err != nil {
					klog.Warningf("Error adding metric sample for container %v: %v", containerID, err)
				}
			}
		}
	}
	return nil
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

func (feeder *clusterStateFeeder) initFromCheckpoints() {
	klog.V(3).Info("Initializing VPA from checkpoints")
	feeder.LoadVPAs(context.TODO())

	namespaces := make(map[string]bool)
	for _, v := range feeder.clusterState.Vpas {
		namespaces[v.ID.Namespace] = true
	}

	for namespace := range namespaces {
		klog.V(3).Infof("Fetching checkpoints from namespace %s", namespace)
		checkpointList, err := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Cannot list VPA checkpoints from namespace %s. Reason: %+v", namespace, err)
		}
		for _, checkpoint := range checkpointList.Items {

			klog.V(3).Infof("Loading VPA %s/%s checkpoint for %s", checkpoint.ObjectMeta.Namespace, checkpoint.Spec.VPAObjectName, checkpoint.Spec.ContainerName)
			err = feeder.setVpaCheckpoint(&checkpoint)
			if err != nil {
				klog.Errorf("Error while loading checkpoint. Reason: %+v", err)
			}

		}
	}
}

func (feeder *clusterStateFeeder) GarbageCollectCheckpoints() {
	klog.V(3).Info("Starting garbage collection of checkpoints")
	feeder.LoadVPAs(context.TODO())

	namespaceList, err := feeder.coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Cannot list namespaces. Reason: %+v", err)
		return
	}

	for _, namespaceItem := range namespaceList.Items {
		namespace := namespaceItem.Name
		checkpointList, err := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Cannot list VPA checkpoints from namespace %v. Reason: %+v", namespace, err)
		}
		for _, checkpoint := range checkpointList.Items {
			vpaID := model.VpaID{Namespace: checkpoint.Namespace, VpaName: checkpoint.Spec.VPAObjectName}
			_, exists := feeder.clusterState.Vpas[vpaID]
			if !exists {
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
	klog.V(3).Infof("Start selecting the vpaCRDs.")
	var vpaCRDs []*vpa_types.VerticalPodAutoscaler
	for _, vpaCRD := range allVpaCRDs {
		if feeder.recommenderName == DefaultRecommenderName {
			if !implicitDefaultRecommender(vpaCRD.Spec.Recommenders) && !selectsRecommender(vpaCRD.Spec.Recommenders, &feeder.recommenderName) {
				klog.V(6).Infof("Ignoring vpaCRD %s as current recommender's name %v doesn't appear among its recommenders", klog.KObj(vpaCRD), feeder.recommenderName)
				continue
			}
		} else {
			if implicitDefaultRecommender(vpaCRD.Spec.Recommenders) {
				klog.V(6).Infof("Ignoring vpaCRD %s as %v recommender doesn't process CRDs implicitly destined to %v recommender", klog.KObj(vpaCRD), feeder.recommenderName, DefaultRecommenderName)
				continue
			}
			if !selectsRecommender(vpaCRD.Spec.Recommenders, &feeder.recommenderName) {
				klog.V(6).Infof("Ignoring vpaCRD %s as current recommender's name %v doesn't appear among its recommenders", klog.KObj(vpaCRD), feeder.recommenderName)
				continue
			}
		}

		if slices.Contains(feeder.ignoredNamespaces, vpaCRD.ObjectMeta.Namespace) {
			klog.V(6).Infof("Ignoring vpaCRD %s in namespace %s as namespace is ignored", klog.KObj(vpaCRD), vpaCRD.Namespace)
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
		klog.Errorf("Cannot list VPAs. Reason: %+v", err)
		return
	}

	// Filter out VPAs that specified recommenders with names not equal to "default"
	vpaCRDs := filterVPAs(feeder, allVpaCRDs)

	klog.V(3).Infof("Fetched %d VPAs.", len(vpaCRDs))
	// Add or update existing VPAs in the model.
	vpaKeys := make(map[model.VpaID]bool)
	for _, vpaCRD := range vpaCRDs {
		vpaID := model.VpaID{
			Namespace: vpaCRD.Namespace,
			VpaName:   vpaCRD.Name,
		}

		selector, conditions := feeder.getSelector(ctx, vpaCRD)
		klog.V(4).Infof("Using selector %s for VPA %s", selector.String(), klog.KObj(vpaCRD))

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
			klog.V(3).Infof("Deleting VPA %s", klog.KRef(vpaID.Namespace, vpaID.VpaName))
			if err := feeder.clusterState.DeleteVpa(vpaID); err != nil {
				klog.Errorf("Deleting VPA %s failed: %v", klog.KRef(vpaID.Namespace, vpaID.VpaName), err)
			}
		}
	}
	feeder.clusterState.ObservedVpas = vpaCRDs
}

// LoadPods loads pod into the cluster state.
func (feeder *clusterStateFeeder) LoadPods() {
	podSpecs, err := feeder.specClient.GetPodSpecs()
	if err != nil {
		klog.Errorf("Cannot get SimplePodSpecs. Reason: %+v", err)
	}
	pods := make(map[model.PodID]*spec.BasicPodSpec)
	for _, spec := range podSpecs {
		pods[spec.ID] = spec
	}
	for key := range feeder.clusterState.Pods {
		if _, exists := pods[key]; !exists {
			klog.V(3).Infof("Deleting Pod %s", klog.KRef(key.Namespace, key.PodName))
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
				klog.Warningf("Failed to add container %+v. Reason: %+v", container.ID, err)
			}
		}
	}
}

func (feeder *clusterStateFeeder) LoadRealTimeMetrics() {
	containersMetrics, err := feeder.metricsClient.GetContainersMetrics()
	if err != nil {
		klog.Errorf("Cannot get ContainerMetricsSnapshot from MetricsClient. Reason: %+v", err)
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
				klog.Warningf("Error adding metric sample for container %v: %v", sample.Container, err)
				droppedSampleCount++
			} else {
				sampleCount++
			}
		}
	}
	klog.V(3).Infof("ClusterSpec fed with #%v ContainerUsageSamples for #%v containers. Dropped #%v samples.", sampleCount, len(containersMetrics), droppedSampleCount)
Loop:
	for {
		select {
		case oomInfo := <-feeder.oomChan:
			klog.V(3).Infof("OOM detected %+v", oomInfo)
			if err = feeder.clusterState.RecordOOM(oomInfo.ContainerID, oomInfo.Timestamp, oomInfo.Memory); err != nil {
				klog.Warningf("Failed to record OOM %+v. Reason: %+v", oomInfo, err)
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
		klog.Errorf("Cannot get target selector from VPA's targetRef. Reason: %+v", fetchErr)
		msg = fmt.Sprintf("Cannot read targetRef. Reason: %s", fetchErr.Error())
	}
	return labels.Nothing(), []condition{
		{conditionType: vpa_types.ConfigUnsupported, delete: false, message: msg},
		{conditionType: vpa_types.ConfigDeprecated, delete: true},
	}
}
