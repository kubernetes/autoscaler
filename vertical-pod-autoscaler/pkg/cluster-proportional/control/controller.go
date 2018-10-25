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

package control

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	scalingpolicy "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	informers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/informers/externalversions"
	scalingpolicylister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/debug"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "scaler-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a ScalingPolicy is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a ScalingPolicy fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by ScalingPolicy"
	// MessageResourceSynced is the message used for an Event fired when a ScalingPolicy
	// is synced successfully
	MessageResourceSynced = "ScalingPolicy synced successfully"
)

// Controller is the controller implementation for ScalingPolicy resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClient kubernetes.Interface
	// scalerClient is a clientset for our own API group
	scalerClient clientset.Interface

	//deploymentsLister appslisters.DeploymentLister
	//deploymentsSynced cache.InformerSynced
	scalingPoliciesLister scalingpolicylister.ClusterProportionalScalerLister
	scalingPoliciesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	state *State
}

// NewController returns a new controller for scaling parent
func NewController(
	state *State,
	kubeClient kubernetes.Interface,
	scalerClient clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	scalerInformerFactory informers.SharedInformerFactory) (*Controller, error) {

	// obtain references to shared index informers for the Deployment and ScalingPolicy
	// types.
	//deploymentInformer := kubeInformerFactory.Apps().V1beta2().Deployments()
	scalingPolicyInformer := scalerInformerFactory.Autoscaling().V1beta1().ClusterProportionalScalers()

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	scalingpolicy.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClient:   kubeClient,
		scalerClient: scalerClient,
		//deploymentsLister: deploymentInformer.Lister(),
		//deploymentsSynced: deploymentInformer.Informer().HasSynced,
		scalingPoliciesLister: scalingPolicyInformer.Lister(),
		scalingPoliciesSynced: scalingPolicyInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ScalingPolicies"),
		recorder:              recorder,
		state:                 state,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when ScalingPolicy resources change
	scalingPolicyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueScalingPolicy,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueScalingPolicy(new)
		},
		DeleteFunc: func(old interface{}) {
			controller.enqueueScalingPolicy(old)
		},
	})

	//// Set up an event handler for when Deployment resources change. This
	//// handler will lookup the owner of the given Deployment, and if it is
	//// owned by a ScalingPolicy resource will enqueue that ScalingPolicy resource for
	//// processing. This way, we don't need to implement custom logic for
	//// handling Deployment resources. More info on this pattern:
	//// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	//deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc: controller.handleObject,
	//	UpdateFunc: func(old, new interface{}) {
	//		newDepl := new.(*appsv1beta2.Deployment)
	//		oldDepl := old.(*appsv1beta2.Deployment)
	//		if newDepl.ResourceVersion == oldDepl.ResourceVersion {
	//			// Periodic resync will send update events for all known Deployments.
	//			// Two different versions of the same Deployment will always have different RVs.
	//			return
	//		}
	//		controller.handleObject(new)
	//	},
	//	DeleteFunc: controller.handleObject,
	//})

	return controller, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting scaling controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.scalingPoliciesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process ScalingPolicy resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.state.Run(stopCh)

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// ScalingPolicy resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced ScalingPolicy '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the ScalingPolicy resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the ScalingPolicy resource with this namespace/name
	scalingPolicy, err := c.scalingPoliciesLister.ClusterProportionalScalers(namespace).Get(name)
	if err != nil {
		// The ScalingPolicy resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			glog.Warningf("scaling policy in work queue no longer exists %s/%s", namespace, name)
			c.state.remove(namespace, name)
			//runtime.HandleError(fmt.Errorf("scalingPolicy '%s' in work queue no longer exists", key))
			return nil
		}

		glog.Warningf("error reading scaling policy %s/%s: %q", namespace, name, err)
		return err
	}

	glog.V(8).Infof("syncing scaling policy: %v", debug.Print(scalingPolicy))
	c.state.upsert(scalingPolicy)
	return nil

	//deploymentName := scalingPolicy.Spec.DeploymentName
	//if deploymentName == "" {
	//	// We choose to absorb the error here as the worker would requeue the
	//	// resource otherwise. Instead, the next time the resource is updated
	//	// the resource will be queued again.
	//	runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
	//	return nil
	//}

	//// Get the deployment with the name specified in ScalingPolicy.spec
	//deployment, err := c.deploymentsLister.Deployments(scalingPolicy.Namespace).Get(deploymentName)
	//// If the resource doesn't exist, we'll create it
	//if errors.IsNotFound(err) {
	//	deployment, err = c.kubeclientset.AppsV1beta2().Deployments(scalingPolicy.Namespace).Create(newDeployment(scalingPolicy))
	//}
	//
	//// If an error occurs during Get/Create, we'll requeue the item so we can
	//// attempt processing again later. This could have been caused by a
	//// temporary network failure, or any other transient reason.
	//if err != nil {
	//	return err
	//}

	//// If the Deployment is not controlled by this ScalingPolicy resource, we should log
	//// a warning to the event recorder and ret
	//if !metav1.IsControlledBy(deployment, scalingPolicy) {
	//	msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
	//	c.recorder.Event(scalingPolicy, corev1.EventTypeWarning, ErrResourceExists, msg)
	//	return fmt.Errorf(msg)
	//}
	//
	//// If this number of the replicas on the ScalingPolicy resource is specified, and the
	//// number does not equal the current desired replicas on the Deployment, we
	//// should update the Deployment resource.
	//if scalingPolicy.Spec.Replicas != nil && *scalingPolicy.Spec.Replicas != *deployment.Spec.Replicas {
	//	glog.V(4).Infof("ScalingPolicyr: %d, deplR: %d", *scalingPolicy.Spec.Replicas, *deployment.Spec.Replicas)
	//	deployment, err = c.kubeclientset.AppsV1beta2().Deployments(scalingPolicy.Namespace).Update(newDeployment(scalingPolicy))
	//}
	//
	//// If an error occurs during Update, we'll requeue the item so we can
	//// attempt processing again later. THis could have been caused by a
	//// temporary network failure, or any other transient reason.
	//if err != nil {
	//	return err
	//}
	//
	//// Finally, we update the status block of the ScalingPolicy resource to reflect the
	//// current state of the world
	//err = c.updateScalingPoliciestatus(scalingPolicy, deployment)
	//if err != nil {
	//	return err
	//}
	//
	//c.recorder.Event(scalingPolicy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	//return nil
}

//func (c *Controller) updateScalingPoliciestatus(scalingPolicy *scalerv1alpha1.ScalingPolicy, deployment *appsv1beta2.Deployment) error {
//	// NEVER modify objects from the store. It's a read-only, local cache.
//	// You can use DeepCopy() to make a deep copy of original object and modify this copy
//	// Or create a copy manually for better performance
//	scalingPolicyCopy := scalingPolicy.DeepCopy()
//	scalingPolicyCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
//	// Until #38113 is merged, we must use Update instead of UpdateStatus to
//	// update the Status block of the ScalingPolicy resource. UpdateStatus will not
//	// allow changes to the Spec of the resource, which is ideal for ensuring
//	// nothing other than resource status has been updated.
//	_, err := c.scalerclientset.ScalingpolicyV1alpha1().ScalingPolicies(scalingPolicy.Namespace).Update(scalingPolicyCopy)
//	return err
//}

// enqueueScalingPolicy takes a ScalingPolicy resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than ScalingPolicy.
func (c *Controller) enqueueScalingPolicy(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

//// handleObject will take any resource implementing metav1.Object and attempt
//// to find the ScalingPolicy resource that 'owns' it. It does this by looking at the
//// objects metadata.ownerReferences field for an appropriate OwnerReference.
//// It then enqueues that ScalingPolicy resource to be processed. If the object does not
//// have an appropriate OwnerReference, it will simply be skipped.
//func (c *Controller) handleObject(obj interface{}) {
//	var object metav1.Object
//	var ok bool
//	if object, ok = obj.(metav1.Object); !ok {
//		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
//		if !ok {
//			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
//			return
//		}
//		object, ok = tombstone.Obj.(metav1.Object)
//		if !ok {
//			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
//			return
//		}
//		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
//	}
//	glog.V(4).Infof("Processing object: %s", object.GetName())
//	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
//		// If this object is not owned by a ScalingPolicy, we should not do anything more
//		// with it.
//		if ownerRef.Kind != "ScalingPolicy" {
//			return
//		}
//
//		scalingPolicy, err := c.scalingPoliciesLister.ScalingPolicies(object.GetNamespace()).Get(ownerRef.Name)
//		if err != nil {
//			glog.V(4).Infof("ignoring orphaned object '%s' of scalingPolicy '%s'", object.GetSelfLink(), ownerRef.Name)
//			return
//		}
//
//		c.enqueueScalingPolicy(scalingPolicy)
//		return
//	}
//}
//
//// newDeployment creates a new Deployment for a ScalingPolicy resource. It also sets
//// the appropriate OwnerReferences on the resource so handleObject can discover
//// the ScalingPolicy resource that 'owns' it.
//func newDeployment(scalingPolicy *scalerv1alpha1.ScalingPolicy) *appsv1beta2.Deployment {
//	labels := map[string]string{
//		"app":        "nginx",
//		"controller": scalingPolicy.Name,
//	}
//	return &appsv1beta2.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      scalingPolicy.Spec.DeploymentName,
//			Namespace: scalingPolicy.Namespace,
//			OwnerReferences: []metav1.OwnerReference{
//				*metav1.NewControllerRef(scalingPolicy, schema.GroupVersionKind{
//					Group:   scalerv1alpha1.SchemeGroupVersion.Group,
//					Version: scalerv1alpha1.SchemeGroupVersion.Version,
//					Kind:    "ScalingPolicy",
//				}),
//			},
//		},
//		Spec: appsv1beta2.DeploymentSpec{
//			Replicas: scalingPolicy.Spec.Replicas,
//			Selector: &metav1.LabelSelector{
//				MatchLabels: labels,
//			},
//			Template: corev1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{
//					Labels: labels,
//				},
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{
//						{
//							Name:  "nginx",
//							Image: "nginx:latest",
//						},
//					},
//				},
//			},
//		},
//	}
//}
