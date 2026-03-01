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
	"errors"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	translators "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/translators"
	updater "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"
)

// BufferController performs updates on Buffers and convert them to pods to be injected
type BufferController interface {
	// Run to run the reconciliation loop frequently every x seconds
	Run(stopCh <-chan struct{})
}

type bufferController struct {
	client         *cbclient.CapacityBufferClient
	strategyFilter filters.Filter
	translator     translators.Translator
	quotaAllocator *resourceQuotaAllocator
	updater        updater.StatusUpdater
	queue          workqueue.TypedRateLimitingInterface[string]
}

// NewBufferController creates new bufferController object
func NewBufferController(
	client *cbclient.CapacityBufferClient,
	strategyFilter filters.Filter,
	translator translators.Translator,
	updater updater.StatusUpdater,
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
	}
	bc.configureEventHandlers()
	return bc
}

// NewDefaultBufferController creates bufferController with default configs
func NewDefaultBufferController(
	client *cbclient.CapacityBufferClient,
) BufferController {
	bc := &bufferController{
		client: client,
		// Accepting empty string as it represents nil value for ProvisioningStrategy
		strategyFilter: filters.NewStrategyFilter([]string{capacitybuffer.ActiveProvisioningStrategy, ""}),
		translator: translators.NewCombinedTranslator(
			[]translators.Translator{
				translators.NewPodTemplateBufferTranslator(client),
				translators.NewDefaultScalableObjectsTranslator(client),
				translators.NewResourceLimitsTranslator(client),
			},
		),
		quotaAllocator: newResourceQuotaAllocator(client),
		updater:        *updater.NewStatusUpdater(client),
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](), workqueue.TypedRateLimitingQueueConfig[string]{Name: "CapacityBuffers"},
		),
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
	// TODO: scalable objects
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

// Run to run the controller reconcile loop
func (c *bufferController) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting CapacityBuffer controller workers")

	// Note: We assume the client passed to us has informers that are running and synced.
	// CapacityBufferClient.NewCapacityBufferClientFromClients waits for sync before returning.

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
		runtime.HandleError(fmt.Errorf("error syncing namespace %q", key))
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

	// Filter the desired provisioning strategy
	// Note: We process ALL buffers in the namespace that match the strategy.
	filteredBuffers, _ := c.strategyFilter.Filter(buffers)

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
	updateErrors := c.updater.Update(filteredBuffers)
	for _, err := range updateErrors {
		runtime.HandleError(fmt.Errorf("capacity buffer controller error: %w", err))
	}

	// If there were any errors, return one to trigger requeue
	if len(translationErrors) > 0 || len(allocationErrors) > 0 || len(updateErrors) > 0 {
		return errors.New("encountered errors during reconciliation")
	}

	return nil
}
