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
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/oom"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

const (
	evictionWatchRetryWait                     = 10 * time.Second
	evictionWatchJitterFactor                  = 0.5
	scaleCacheLoopPeriod         time.Duration = 7 * time.Second
	scaleCacheEntryLifetime      time.Duration = time.Hour
	scaleCacheEntryFreshnessTime time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor  float64       = 1.
	defaultResyncPeriod          time.Duration = 10 * time.Minute
)

// ClusterStateFeeder can update state of ClusterState object.
type ClusterStateFeeder interface {
	// InitFromHistoryProvider loads historical pod spec into clusterState.
	InitFromHistoryProvider(historyProvider history.HistoryProvider)

	// InitFromCheckpoints loads historical checkpoints into clusterState.
	InitFromCheckpoints()

	// LoadVPAs updates clusterState with current state of VPAs.
	LoadVPAs()

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
	}
}

// NewClusterStateFeeder creates new ClusterStateFeeder with internal data providers, based on kube client config.
// Deprecated; Use ClusterStateFeederFactory instead.
func NewClusterStateFeeder(config *rest.Config, clusterState *model.ClusterState, memorySave bool, namespace string) ClusterStateFeeder {
	kubeClient := kube_client.NewForConfigOrDie(config)
	podLister, oomObserver := NewPodListerAndOOMObserver(kubeClient, namespace)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(namespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	controllerFetcher.Start(context.TODO(), scaleCacheLoopPeriod)
	return ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       newMetricsClient(config, namespace),
		VpaCheckpointClient: vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		VpaLister:           vpa_api_util.NewVpasLister(vpa_clientset.NewForConfigOrDie(config), make(chan struct{}), namespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(config, kubeClient, factory),
		MemorySaveMode:      memorySave,
		ControllerFetcher:   controllerFetcher,
	}.Make()
}

func newMetricsClient(config *rest.Config, namespace string) metrics.MetricsClient {
	metricsGetter := resourceclient.NewForConfigOrDie(config)
	return metrics.NewMetricsClient(metricsGetter, namespace)
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
}

func (feeder *clusterStateFeeder) InitFromHistoryProvider(historyProvider history.HistoryProvider) {
	klog.V(3).Info("Initializing VPA from history provider")
	clusterHistory, err := historyProvider.GetClusterHistory()
	if err != nil {
		klog.Errorf("Cannot get cluster history: %v", err)
	}
	for podID, podHistory := range clusterHistory {
		klog.V(4).Infof("Adding pod %v with labels %v", podID, podHistory.LastLabels)
		feeder.clusterState.AddOrUpdatePod(podID, podHistory.LastLabels, apiv1.PodUnknown)
		for containerName, sampleList := range podHistory.Samples {
			containerID := model.ContainerID{
				PodID:         podID,
				ContainerName: containerName,
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
}

func (feeder *clusterStateFeeder) setVpaCheckpoint(checkpoint *vpa_types.VerticalPodAutoscalerCheckpoint) error {
	vpaID := model.VpaID{Namespace: checkpoint.Namespace, VpaName: checkpoint.Spec.VPAObjectName}
	vpa, exists := feeder.clusterState.Vpas[vpaID]
	if !exists {
		return fmt.Errorf("cannot load checkpoint to missing VPA object %+v", vpaID)
	}

	cs := model.NewAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint.Status)
	if err != nil {
		return fmt.Errorf("cannot load checkpoint for VPA %+v. Reason: %v", vpa.ID, err)
	}
	vpa.ContainersInitialAggregateState[checkpoint.Spec.ContainerName] = cs
	return nil
}

func (feeder *clusterStateFeeder) InitFromCheckpoints() {
	klog.V(3).Info("Initializing VPA from checkpoints")
	feeder.LoadVPAs()

	namespaces := make(map[string]bool)
	for _, v := range feeder.clusterState.Vpas {
		namespaces[v.ID.Namespace] = true
	}

	for namespace := range namespaces {
		klog.V(3).Infof("Fetching checkpoints from namespace %s", namespace)
		checkpointList, err := feeder.vpaCheckpointClient.VerticalPodAutoscalerCheckpoints(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Cannot list VPA checkpoints from namespace %v. Reason: %+v", namespace, err)
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
	feeder.LoadVPAs()

	namspaceList, err := feeder.coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Cannot list namespaces. Reason: %+v", err)
		return
	}

	for _, namespaceItem := range namspaceList.Items {
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

// Fetch VPA objects and load them into the cluster state.
func (feeder *clusterStateFeeder) LoadVPAs() {
	// List VPA API objects.
	vpaCRDs, err := feeder.vpaLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Cannot list VPAs. Reason: %+v", err)
		return
	}
	klog.V(3).Infof("Fetched %d VPAs.", len(vpaCRDs))
	// Add or update existing VPAs in the model.
	vpaKeys := make(map[model.VpaID]bool)
	for _, vpaCRD := range vpaCRDs {
		vpaID := model.VpaID{
			Namespace: vpaCRD.Namespace,
			VpaName:   vpaCRD.Name,
		}

		selector, conditions := feeder.getSelector(vpaCRD)
		klog.Infof("Using selector %s for VPA %s/%s", selector.String(), vpaCRD.Namespace, vpaCRD.Name)

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
			klog.V(3).Infof("Deleting VPA %v", vpaID)
			feeder.clusterState.DeleteVpa(vpaID)
		}
	}
	feeder.clusterState.ObservedVpas = vpaCRDs
}

// Load pod into the cluster state.
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
			klog.V(3).Infof("Deleting Pod %v", key)
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

func (feeder *clusterStateFeeder) validateTargetRef(vpa *vpa_types.VerticalPodAutoscaler) (bool, condition) {
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
	top, err := feeder.controllerFetcher.FindTopMostWellKnownOrScalable(&k)
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

func (feeder *clusterStateFeeder) getSelector(vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, []condition) {
	selector, fetchErr := feeder.selectorFetcher.Fetch(vpa)
	if selector != nil {
		validTargetRef, unsupportedCondition := feeder.validateTargetRef(vpa)
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
