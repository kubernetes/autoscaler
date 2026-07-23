/*
Copyright 2025 The Kubernetes Authors.

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

package controller

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/fakepods"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	cbmetrics "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/metrics"
	translators "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators"
	scalableobject "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators/scalable_objects"
	updater "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"
	"k8s.io/utils/clock"
)

const (
	// EventDrivenReconciliationCondition is a condition type that indicates if the buffer is being reconciled via events.
	EventDrivenReconciliationCondition = "EventDrivenReconciliation"
	// DynamicWatchFailedReason is a reason for EventDrivenReconciliationCondition when dynamic watch establishment fails.
	DynamicWatchFailedReason = "DynamicWatchFailed"
)

// BufferController performs updates on Buffers and convert them to pods to be injected
type BufferController interface {
	// Run to run the reconciliation loop frequently every x seconds
	Run(stopCh <-chan struct{})
}

type bufferController struct {
	client                  *cbclient.CapacityBufferClient
	strategyFilter          filters.Filter
	translator              translators.Translator
	quotaAllocator          *resourceQuotaAllocator
	updater                 updater.StatusUpdater
	queue                   workqueue.TypedRateLimitingInterface[string]
	clock                   clock.Clock
	reconciliationTimeCache *cbmetrics.ReconciliationCache

	// Dynamic watching and RBAC updates
	rbacUpdater            CapacityBufferRBACUpdater
	dynamicClient          dynamic.Interface
	dynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	watchedGVKs            sync.Map // map[schema.GroupVersionKind]bool
	stopCh                 <-chan struct{}
}

// NewBufferController creates new bufferController object
func NewBufferController(
	client *cbclient.CapacityBufferClient,
	strategyFilter filters.Filter,
	translator translators.Translator,
	updater updater.StatusUpdater,
	clock clock.Clock,
	reconciliationTimeCache *cbmetrics.ReconciliationCache,
	rbacUpdater CapacityBufferRBACUpdater,
	dynamicClient dynamic.Interface,
) BufferController {
	bc := &bufferController{
		client:         client,
		strategyFilter: strategyFilter,
		translator:     translator,
		quotaAllocator: newResourceQuotaAllocator(client),
		updater:        updater,
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "CapacityBuffers"},
		),
		clock:                   clock,
		reconciliationTimeCache: reconciliationTimeCache,
		rbacUpdater:             rbacUpdater,
		dynamicClient:           dynamicClient,
		dynamicInformerFactory:  dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0),
	}
	bc.configureEventHandlers()
	return bc
}

// InitializeAndRunDefaultBufferController creates the default Capacity buffer controller and processing interval metric collector
// and runs each of them asyncrounsly
func InitializeAndRunDefaultBufferController(
	ctx context.Context,
	client *cbclient.CapacityBufferClient,
	resolver fakepods.Resolver,

) {
	realClock := clock.RealClock{}
	reconciledBuffersCache := cbmetrics.NewReconciliationCache()
	// Accepting empty string as it represents nil value for ProvisioningStrategy
	defaultStrategies := []string{capacitybuffer.ActiveProvisioningStrategy, ""}
	controller := NewDefaultBufferController(client, resolver, defaultStrategies, reconciledBuffersCache, realClock, NewDefaultRBACUpdater(client.GetKubernetesClient()), client.GetDynamicClient())
	go controller.Run(ctx.Done())

	cbmetrics.RegisterReconciliationTimestampCollector(client, defaultStrategies, reconciledBuffersCache, realClock)
}

// NewDefaultBufferController creates bufferController with default configs
func NewDefaultBufferController(
	client *cbclient.CapacityBufferClient,
	resolver fakepods.Resolver,
	strategies []string,
	reconciliationTimeCache *cbmetrics.ReconciliationCache,
	clock clock.Clock,
	rbacUpdater CapacityBufferRBACUpdater,
	dynamicClient dynamic.Interface,
) BufferController {
	bc := &bufferController{
		client:         client,
		strategyFilter: filters.NewStrategyFilter(strategies),
		translator: translators.NewCombinedTranslator(
			[]translators.Translator{
				translators.NewPodTemplateBufferTranslator(client, resolver),
				translators.NewDefaultScalableObjectsTranslator(client, resolver),
			},
		),
		quotaAllocator: newResourceQuotaAllocator(client),
		updater:        *updater.NewStatusUpdater(client),
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "CapacityBuffers"},
		),
		clock:                   clock,
		reconciliationTimeCache: reconciliationTimeCache,
		rbacUpdater:             rbacUpdater,
		dynamicClient:           dynamicClient,
		dynamicInformerFactory:  dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0),
	}
	bc.configureEventHandlers()
	return bc
}

func (c *bufferController) configureEventHandlers() {
	// CapacityBuffer Informer
	c.client.GetBufferInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueNamespace(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldBuf := oldObj.(*v1.CapacityBuffer)
			newBuf := newObj.(*v1.CapacityBuffer)

			// 1. Resync (periodic refresh): reconcile.
			if oldBuf.ResourceVersion == newBuf.ResourceVersion {
				c.enqueueNamespace(newObj)
				return
			}

			// 2. Generation changes (spec changes): reconcile.
			if oldBuf.Generation != newBuf.Generation {
				c.enqueueNamespace(newObj)
				return
			}
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueueNamespace(obj)
		},
	})

	// ResourceQuota Informer
	c.client.GetResourceQuotaInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueNamespace(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldQuota := oldObj.(*corev1.ResourceQuota)
			newQuota := newObj.(*corev1.ResourceQuota)

			// Reconcile only on Status changes (Status.Hard and Status.Used)
			if equality.Semantic.DeepEqual(oldQuota.Status.Hard, newQuota.Status.Hard) &&
				equality.Semantic.DeepEqual(oldQuota.Status.Used, newQuota.Status.Used) {
				return
			}
			c.enqueueNamespace(oldObj)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueueNamespace(obj)
		},
	})

	// PodTemplate Informer
	c.client.GetPodTemplateInformer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingPodTemplate(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMeta, err := meta.Accessor(oldObj)
			if err != nil {
				klog.Errorf("CapacityBuffer controller: failed to get meta for object, err: %v", err)
				return
			}
			newMeta, err := meta.Accessor(newObj)
			if err != nil {
				klog.Errorf("CapacityBuffer controller: failed to get meta for object, err: %v", err)
				return
			}
			if oldMeta.GetGeneration() == newMeta.GetGeneration() {
				return
			}
			c.enqueueBuffersReferencingPodTemplate(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingPodTemplate(obj)
		},
	})

	// Deployment Informer
	c.client.GetDeploymentInformer().AddEventHandler(c.scalableObjectHandlerFuncs(scalableobject.ApiGroupApps, scalableobject.DeploymentKind))

	// ReplicaSet Informer
	c.client.GetReplicaSetInformer().AddEventHandler(c.scalableObjectHandlerFuncs(scalableobject.ApiGroupApps, scalableobject.ReplicaSetKind))

	// StatefulSet Informer
	c.client.GetStatefulSetInformer().AddEventHandler(c.scalableObjectHandlerFuncs(scalableobject.ApiGroupApps, scalableobject.StatefulSetKind))

	// Job Informer
	c.client.GetJobInformer().AddEventHandler(c.scalableObjectHandlerFuncs(scalableobject.ApiGroupBatch, scalableobject.JobKind))

	// ReplicationController Informer
	c.client.GetReplicationControllerInformer().AddEventHandler(c.scalableObjectHandlerFuncs(scalableobject.ApiGroupCore, scalableobject.ReplicationControllerKind))
}

func (c *bufferController) scalableObjectHandlerFuncs(apiGroup, kind string) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingScalableObject(obj, apiGroup, kind)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldMeta, err := meta.Accessor(oldObj)
			if err != nil {
				klog.Errorf("CapacityBuffer controller: failed to get meta for %s/%s object, err: %v", apiGroup, kind, err)
				return
			}
			newMeta, err := meta.Accessor(newObj)
			if err != nil {
				klog.Errorf("CapacityBuffer controller: failed to get meta for %s/%s object, err: %v", apiGroup, kind, err)
				return
			}
			if oldMeta.GetGeneration() == newMeta.GetGeneration() {
				return
			}
			c.enqueueBuffersReferencingScalableObject(newObj, apiGroup, kind)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingScalableObject(obj, apiGroup, kind)
		},
	}
}

func (c *bufferController) enqueueNamespace(obj interface{}) {
	var ns string
	if object, ok := obj.(interface{ GetNamespace() string }); ok {
		ns = object.GetNamespace()
	}
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		if object, ok := tombstone.Obj.(interface{ GetNamespace() string }); ok {
			ns = object.GetNamespace()
		}
	}
	if ns != "" {
		c.queue.Add(ns)
	}
}

func (c *bufferController) enqueueBuffersReferencingPodTemplate(obj interface{}) {
	template, ok := obj.(*corev1.PodTemplate)
	if !ok {
		// handle tombstone
		if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			if cast, ok := tombstone.Obj.(*corev1.PodTemplate); ok {
				template = cast
			}
		}
	}
	if template == nil {
		return
	}

	// Use indexer to find buffers referencing this template
	buffers, err := c.client.GetBufferInformer().GetIndexer().ByIndex(cbclient.PodTemplateRefIndex, template.Name)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error looking up buffers for pod template %s: %w", template.Name, err))
		return
	}

	for _, obj := range buffers {
		buffer := obj.(*v1.CapacityBuffer)
		if buffer.Namespace == template.Namespace {
			c.queue.Add(buffer.Namespace)
			return // we reconcile the whole namespace, so finding one buffer is enough to trigger it.
		}
	}
}

func (c *bufferController) enqueueBuffersReferencingScalableObject(obj interface{}, apiGroup, kind string) {
	object, err := meta.Accessor(obj)
	if err != nil {
		// handle tombstone
		if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			if cast, ok := tombstone.Obj.(metav1.Object); ok {
				object = cast
			}
		}
	}
	if object == nil {
		return
	}

	// Use indexer to find buffers referencing this scalable object
	buffers, err := c.client.GetBufferInformer().GetIndexer().ByIndex(cbclient.ScalableRefIndex, object.GetName())
	if err != nil {
		runtime.HandleError(fmt.Errorf("error looking up buffers for scalable object %s/%s: %w", kind, object.GetName(), err))
		return
	}

	for _, obj := range buffers {
		buffer := obj.(*v1.CapacityBuffer)
		if buffer.Namespace == object.GetNamespace() && buffer.Spec.ScalableRef != nil &&
			buffer.Spec.ScalableRef.Kind == kind && buffer.Spec.ScalableRef.APIGroup == apiGroup {
			c.queue.Add(buffer.Namespace)
			return // we reconcile the whole namespace, so finding one buffer is enough to trigger it.
		}
	}
}

// Run to run the controller reconcile loop
func (c *bufferController) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.stopCh = stopCh

	klog.Info("Starting CapacityBuffer controller workers")

	// Launch a single worker (namespace processing is serial per namespace anyway)
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
	klog.Info("Stopping CapacityBuffer controller")
}

func (c *bufferController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *bufferController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.reconcileNamespace(key)

	if err == nil {
		c.queue.Forget(key)
	} else {
		// Put the item back on the queue to handle it later
		c.queue.AddRateLimited(key)
		runtime.HandleError(fmt.Errorf("capacity buffer controller: error syncing namespace %q, requeueing", key))
	}
	return true
}

// reconcileNamespace reconciles all buffers in a namespace.
//
// We must reconcile all buffers in a namespace because of resource quota allocation.
// If one buffer in a namespace changes, e.g. it requests more resources,
// it may impact other buffers in the namespace.
func (c *bufferController) reconcileNamespace(namespace string) error {
	klog.V(5).Infof("CapacityBuffer controller: reconciling namespace: %s", namespace)
	// List all capacity buffers in the target namespace
	buffers, err := c.client.ListCapacityBuffers(namespace)
	if err != nil {
		return err
	}

	// Ensure dynamic watches for all scalable objects observed in this namespace
	for _, buffer := range buffers {
		if buffer.Spec.ScalableRef != nil {
			c.ensureWatchFromScalableRef(buffer.Spec.ScalableRef)
		}
	}

	// Filter the desired provisioning strategy
	// Note: We process ALL buffers in the namespace that match the strategy.
	filteredBuffers, filteredOutBuffers := c.strategyFilter.Filter(buffers)

	// Update reconciliation time for filtered out buffers
	c.updateReconciliationTimeCache(filteredOutBuffers)

	if len(filteredBuffers) == 0 {
		return nil
	}

	// Sort buffers deterministically by CreationTimestamp, then Name. Stable order
	// is required to prevent flakiness of resource quotas allocation.
	sort.Slice(filteredBuffers, func(i, j int) bool {
		if filteredBuffers[i].CreationTimestamp.Time.Equal(filteredBuffers[j].CreationTimestamp.Time) {
			return filteredBuffers[i].Name < filteredBuffers[j].Name
		}
		return filteredBuffers[i].CreationTimestamp.Before(&filteredBuffers[j].CreationTimestamp)
	})

	// Extract pod specs and number of replicas from filtered buffers
	translationErrors := c.translator.Translate(filteredBuffers)
	for _, err := range translationErrors {
		runtime.HandleError(fmt.Errorf("capacity buffer controller error: %w", err))
	}

	// Allocate resource quotas
	allocationErrors := c.quotaAllocator.Allocate(namespace, filteredBuffers)
	for _, err := range allocationErrors {
		runtime.HandleError(fmt.Errorf("capacity buffer controller error: %w", err))
	}

	// Update buffer status by calling API server
	updatedBuffers, updateErrors := c.updater.Update(filteredBuffers)
	c.updateReconciliationTimeCache(updatedBuffers)
	for _, err := range updateErrors {
		runtime.HandleError(fmt.Errorf("capacity buffer controller error: %w", err))
	}

	// If there were any errors, return one to trigger requeue
	if len(translationErrors) > 0 || len(allocationErrors) > 0 || len(updateErrors) > 0 {
		return errors.New("encountered errors during reconciliation")
	}

	return nil
}

func (c *bufferController) updateReconciliationTimeCache(buffers []*v1.CapacityBuffer) {
	if c.reconciliationTimeCache == nil || len(buffers) == 0 {
		return
	}
	c.reconciliationTimeCache.Update(buffers, c.clock.Now())
}

func (c *bufferController) ensureWatchFromScalableRef(ref *v1.ScalableRef) {
	gk := schema.GroupKind{
		Group: ref.APIGroup,
		Kind:  ref.Kind,
	}

	// Use RESTMapper to find the preferred version and resource mapping
	mapping, err := c.client.GetRESTMapper().RESTMapping(gk)
	if err != nil {
		klog.V(4).Infof("Failed to resolve GVK for %v: %v", gk, err)
		c.markBuffersAsNonEventDriven(gk.WithVersion(""), schema.GroupVersionResource{Group: gk.Group, Resource: strings.ToLower(gk.Kind) + "s"})
		return
	}

	c.ensureWatch(mapping)
}

func (c *bufferController) ensureWatch(mapping *meta.RESTMapping) {
	gvk := mapping.GroupVersionKind
	if _, loaded := c.watchedGVKs.LoadOrStore(gvk, true); loaded {
		return
	}

	klog.V(4).Infof("Establishing dynamic watch for GVK: %v", gvk)
	// Start watch establishment with RBAC guidance and retry mechanism
	go c.establishWatchWithRetry(mapping)
}

func (c *bufferController) establishWatchWithRetry(mapping *meta.RESTMapping) {
	gvk := mapping.GroupVersionKind
	gvr := mapping.Resource

	// Provide RBAC guide once for newly discovered resource type
	if err := c.rbacUpdater.UpdateRBAC(mapping); err != nil {
		klog.Errorf("Failed to provide RBAC guide for GVK %v: %v", gvk, err)
	}

	informer := c.dynamicInformerFactory.ForResource(gvr).Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingDynamicObject(obj, gvk)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.enqueueBuffersReferencingDynamicObject(newObj, gvk)
		},
		DeleteFunc: func(obj interface{}) {
			c.enqueueBuffersReferencingDynamicObject(obj, gvk)
		},
	})

	// Start the informer exactly once
	go informer.Run(c.stopCh)

	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		klog.V(4).Infof("Waiting for dynamic watch cache sync for GVK: %v", gvk)

		// Tolerate failure by checking cache sync with timeout
		syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if !cache.WaitForCacheSync(syncCtx.Done(), informer.HasSynced) {
			klog.V(4).Infof("Cache sync for %v not yet ready (might be RBAC or missing CRD). Retrying...", gvk)
			return false, nil // retry
		}

		klog.V(2).Infof("Successfully established dynamic watch for GVR: %v", gvr)
		return true, nil
	})

	if err != nil {
		klog.Errorf("Exhausted retries for establishing watch on GVK %v: %v", gvk, err)
		c.watchedGVKs.Delete(gvk) // Allow retry in next reconciliation cycle
		c.markBuffersAsNonEventDriven(gvk, gvr)
	}
}

func (c *bufferController) markBuffersAsNonEventDriven(gvk schema.GroupVersionKind, gvr schema.GroupVersionResource) {
	buffers, err := c.client.ListCapacityBuffers("")
	if err != nil {
		klog.Errorf("Failed to list buffers to mark as non-event driven: %v", err)
		return
	}

	var buffersToUpdate []*v1.CapacityBuffer
	for _, buffer := range buffers {
		if buffer.Spec.ScalableRef != nil &&
			buffer.Spec.ScalableRef.Kind == gvk.Kind &&
			buffer.Spec.ScalableRef.APIGroup == gvk.Group {

			msg := fmt.Sprintf("Failed to establish dynamic watch for %v. Reconciliation will be periodic (up to 5m delay). "+
				"Please ensure ClusterAutoscaler has 'get/list/watch' permissions for %s and %s/scale. "+
				"Refer to RBAC GUIDE in controller logs for details.", gvk, gvr.Resource, gvr.Resource)

			newCondition := metav1.Condition{
				Type:               EventDrivenReconciliationCondition,
				Status:             metav1.ConditionFalse,
				Reason:             DynamicWatchFailedReason,
				Message:            msg,
				LastTransitionTime: metav1.NewTime(c.clock.Now()),
			}

			// Add or update the condition
			found := false
			for i, cond := range buffer.Status.Conditions {
				if cond.Type == EventDrivenReconciliationCondition {
					buffer.Status.Conditions[i] = newCondition
					found = true
					break
				}
			}
			if !found {
				buffer.Status.Conditions = append(buffer.Status.Conditions, newCondition)
			}
			buffersToUpdate = append(buffersToUpdate, buffer)
		}
	}

	if len(buffersToUpdate) > 0 {
		_, errs := c.updater.Update(buffersToUpdate)
		for _, err := range errs {
			klog.Errorf("Failed to update buffer status with non-event driven condition: %v", err)
		}
	}
}

func (c *bufferController) enqueueBuffersReferencingDynamicObject(obj interface{}, gvk schema.GroupVersionKind) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		klog.V(4).Infof("CapacityBuffer controller: failed to get meta accessor for dynamic object: %v", err)
		return
	}

	// Use indexer to find buffers referencing this object
	buffers, err := c.client.GetBufferInformer().GetIndexer().ByIndex(cbclient.ScalableRefIndex, metaObj.GetName())
	if err != nil {
		klog.Errorf("CapacityBuffer controller: error looking up buffers for dynamic object %s/%s: %v", gvk.Kind, metaObj.GetName(), err)
		return
	}

	for _, b := range buffers {
		buffer := b.(*v1.CapacityBuffer)
		if buffer.Namespace == metaObj.GetNamespace() && buffer.Spec.ScalableRef != nil &&
			buffer.Spec.ScalableRef.Kind == gvk.Kind && buffer.Spec.ScalableRef.APIGroup == gvk.Group {
			c.enqueueNamespace(buffer)
		}
	}
}
