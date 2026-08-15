/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"

	balancerapi "k8s.io/autoscaler/balancer/pkg/apis/balancer.x-k8s.io/v1alpha1"
	balancerclientset "k8s.io/autoscaler/balancer/pkg/client/clientset/versioned"
	balancerscheme "k8s.io/autoscaler/balancer/pkg/client/clientset/versioned/scheme"
	balancerinformers "k8s.io/autoscaler/balancer/pkg/client/informers/externalversions/balancer.x-k8s.io/v1alpha1"
	balancerlisters "k8s.io/autoscaler/balancer/pkg/client/listers/balancer.x-k8s.io/v1alpha1"
)

const controllerAgentName = "balancer-controller"

// Controller is the controller implementation for Balancer resources
type Controller struct {
	// For balancer object access.
	balancerClientSet balancerclientset.Interface
	balancerLister    balancerlisters.BalancerLister
	balancerSynced    cache.InformerSynced

	core CoreInterface

	// workqueue is a rate limited work queue.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new Balancer controller
func NewController(
	balancerClientSet balancerclientset.Interface,
	balancerInformer balancerinformers.BalancerInformer,
	eventinterface typedcorev1.EventInterface,

	core CoreInterface,
	resync time.Duration,
) *Controller {

	// Create event recorder.
	// Add balancer-controller types to the default Kubernetes Scheme so Events can be
	// logged for balancer-controller types.
	utilruntime.Must(balancerscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: eventinterface})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		balancerClientSet: balancerClientSet,
		balancerLister:    balancerInformer.Lister(),
		balancerSynced:    balancerInformer.Informer().HasSynced,
		recorder:          recorder,
		core:              core,

		// Workqueue will process the items every resync period.
		workqueue: workqueue.NewNamedRateLimitingQueue(NewFixedItemIntervalRateLimiter(resync), "Balancer"),
	}

	klog.Info("Setting up event handlers for Balancer")
	// Set up an event handler for when Balancer resources change
	balancerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueBalancer,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueBalancer(new)
		},
		DeleteFunc: controller.deleteBalancer,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shut down the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.V(1).Info("Starting Balancer controller")
	if ok := cache.WaitForCacheSync(stopCh, c.balancerSynced, c.core.IsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	// Launch workers to process Balancers
	for i := 0; i < threadiness; i++ {
		go wait.Until(func() {
			for c.processNextWorkItem() {
			}
		}, time.Second, stopCh)
	}

	klog.V(1).Info("Balancer controller is running")
	<-stopCh
	klog.V(1).Info("Shutting down Balancer controller")
	return nil
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	objKey, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We expect strings to come off the workqueue. These are of the
	// form namespace/name.
	key, ok := objKey.(string)
	if !ok {
		c.dropKey(objKey)
		klog.Errorf("expected string in workqueue but got %#v", objKey)
		return true
	}

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("Invalid resource key: %s", key)
		return false
	}
	klog.V(3).Infof("Processing balancer %s", key)

	again := c.syncHandler(namespace, name)
	if again {
		c.workqueue.AddRateLimited(key)
		c.workqueue.Done(key)
	} else {
		c.dropKey(key)
	}
	return true
}

// syncHandler Processes balancer with the given key and updates the Status.
// Return true if the key should be re-processed. False if the key is to
// be dropped.
func (c *Controller) syncHandler(namespace, name string) bool {

	// Get the Balancer resource with this namespace/name
	balancer, err := c.balancerLister.Balancers(namespace).Get(name)
	if err != nil {
		// The Balancer resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Warningf("Balancer %s/%s not found, dropping from queue", namespace, name)
			return false
		}
		klog.Warningf("Balancer %s/%s not obtained: %s", namespace, name, err.Error())
		// Maybe there is a chance...
		return true
	}

	// Make a deep copy to avoid modifying concurrently used informer object.
	balancer = balancer.DeepCopy()
	originalStatus := balancer.Status.DeepCopy()

	statusInfo, processError := c.core.ProcessBalancer(balancer, time.Now())

	if statusInfo != nil {
		balancer.Status.Replicas = statusInfo.replicasObserved
	}
	selector := balancer.Spec.Selector
	balancer.Status.Selector = metav1.FormatLabelSelector(&selector)
	setConditionsBasedOnError(balancer, processError, time.Now())
	if processError != nil {
		klog.Warningf("Failed to process balancer %s/%s: %s", namespace, name, processError.Error())
		c.recorder.Event(balancer, corev1.EventTypeWarning, "UnableToBalance", processError.Error())
	}

	if err = c.updateStatusIfNeeded(originalStatus, balancer); err != nil {
		c.recorder.Event(balancer, corev1.EventTypeWarning, "StatusNotUpdated", err.Error())
		klog.Warningf("Failed to update status of balancer %s/%s: %s", namespace, name, err.Error())
	}
	return true
}

// enqueueBalancer takes a Balancer resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Balancer.
func (c *Controller) enqueueBalancer(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) deleteBalancer(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.dropKey(key)
}

func (c *Controller) dropKey(key interface{}) {
	c.workqueue.Forget(key)
	c.workqueue.Done(key)
}

// updateStatusIfNeeded calls updateStatus only if the status of the new Balancer is not the same as the old status
func (c *Controller) updateStatusIfNeeded(oldStatus *balancerapi.BalancerStatus, new *balancerapi.Balancer) error {
	// skip a write if we wouldn't need to update
	if apiequality.Semantic.DeepEqual(oldStatus, &new.Status) {
		return nil
	}
	_, err := c.balancerClientSet.BalancerV1alpha1().Balancers(new.Namespace).UpdateStatus(context.TODO(), new, metav1.UpdateOptions{})
	return err
}
