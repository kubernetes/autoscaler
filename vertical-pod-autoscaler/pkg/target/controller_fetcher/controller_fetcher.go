/*
Copyright 2019 The Kubernetes Authors.

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

package controllerfetcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
)

type wellKnownController string

const (
	cronJob               wellKnownController = "CronJob"
	daemonSet             wellKnownController = "DaemonSet"
	deployment            wellKnownController = "Deployment"
	node                  wellKnownController = "Node"
	job                   wellKnownController = "Job"
	replicaSet            wellKnownController = "ReplicaSet"
	replicationController wellKnownController = "ReplicationController"
	statefulSet           wellKnownController = "StatefulSet"
)

const (
	discoveryResetPeriod time.Duration = 5 * time.Minute
)

// ErrNodeInvalidOwner is thrown when a Pod is owned by a Node.
var ErrNodeInvalidOwner = errors.New("node is not a valid owner")

// ControllerKey identifies a controller.
type ControllerKey struct {
	Namespace string
	Kind      string
	Name      string
}

// ControllerKeyWithAPIVersion identifies a controller and API it's defined in.
type ControllerKeyWithAPIVersion struct {
	ControllerKey
	ApiVersion string
	// UID is set for owner references, but may be empty for VPA target references.
	UID types.UID
}

// ControllerFetcher is responsible for finding the topmost well-known or scalable controller
type ControllerFetcher interface {
	// FindTopMostWellKnownOrScalable returns topmost well-known or scalable controller. Error is returned if controller cannot be found.
	FindTopMostWellKnownOrScalable(ctx context.Context, controller *ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error)
}

type controllerFetcher struct {
	kubeClient                   kube_client.Interface
	scaleNamespacer              scale.ScalesGetter
	mapper                       apimeta.RESTMapper
	informersMap                 map[wellKnownController]cache.SharedIndexInformer
	scaleSubresourceCacheStorage controllerCacheStorage
}

func (f *controllerFetcher) periodicallyRefreshCache(ctx context.Context, period time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(period):
			keysToRefresh := f.scaleSubresourceCacheStorage.GetKeysToRefresh()
			klog.V(5).InfoS("Starting to refresh entries in controllerFetchers scaleSubresourceCacheStorage")
			for _, item := range keysToRefresh {
				scale, err := f.scaleNamespacer.Scales(item.namespace).Get(context.TODO(), item.groupResource, item.name, metav1.GetOptions{})
				f.scaleSubresourceCacheStorage.Refresh(item.namespace, item.groupResource, item.name, scale, err)
			}
			klog.V(5).InfoS("Finished refreshing entries in controllerFetchers scaleSubresourceCacheStorage", "refreshed", len(keysToRefresh))
			f.scaleSubresourceCacheStorage.RemoveExpired()
		}
	}
}

func (f *controllerFetcher) Start(ctx context.Context, loopPeriod time.Duration) {
	go f.periodicallyRefreshCache(ctx, loopPeriod)
}

// NewControllerFetcher returns a new instance of controllerFetcher
func NewControllerFetcher(config *rest.Config, kubeClient kube_client.Interface, factory informers.SharedInformerFactory, betweenRefreshes, lifeTime time.Duration, jitterFactor float64, stopCh <-chan struct{}) *controllerFetcher {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		klog.ErrorS(err, "Could not create discoveryClient")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	restClient := kubeClient.CoreV1().RESTClient()
	cachedDiscoveryClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)
	go wait.Until(func() {
		mapper.Reset()
	}, discoveryResetPeriod, stopCh)

	informersMap := map[wellKnownController]cache.SharedIndexInformer{
		daemonSet:             factory.Apps().V1().DaemonSets().Informer(),
		deployment:            factory.Apps().V1().Deployments().Informer(),
		replicaSet:            factory.Apps().V1().ReplicaSets().Informer(),
		statefulSet:           factory.Apps().V1().StatefulSets().Informer(),
		replicationController: factory.Core().V1().ReplicationControllers().Informer(),
		job:                   factory.Batch().V1().Jobs().Informer(),
		cronJob:               factory.Batch().V1().CronJobs().Informer(),
	}

	scaleNamespacer := scale.New(restClient, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	return &controllerFetcher{
		kubeClient:                   kubeClient,
		scaleNamespacer:              scaleNamespacer,
		mapper:                       mapper,
		informersMap:                 informersMap,
		scaleSubresourceCacheStorage: newControllerCacheStorage(betweenRefreshes, lifeTime, jitterFactor),
	}
}

// getOwnerController returns the controlling owner's key, including its expected UID.
// It returns nil when none of the references designates a controller.
func getOwnerController(owners []metav1.OwnerReference, namespace string) *ControllerKeyWithAPIVersion {
	for _, owner := range owners {
		if owner.Controller != nil && *owner.Controller {
			return &ControllerKeyWithAPIVersion{
				ControllerKey: ControllerKey{
					Namespace: namespace,
					Kind:      owner.Kind,
					Name:      owner.Name,
				},
				ApiVersion: owner.APIVersion,
				UID:        owner.UID,
			}
		}
	}
	return nil
}

// getParentOfWellKnownController reads a controller from its informer, falling back
// to one API GET on a cache miss or UID mismatch. Matching UIDs establish identity,
// not metadata freshness. Lookup errors propagate without updating the informer store.
func (f *controllerFetcher) getParentOfWellKnownController(ctx context.Context, informer cache.SharedIndexInformer, controllerKey ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	namespace := controllerKey.Namespace
	name := controllerKey.Name

	obj, exists, err := informer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	var apiObj metav1.Object
	if exists {
		apiObj, err = apimeta.Accessor(obj)
		if err != nil {
			return nil, err
		}
	}
	if !exists || (controllerKey.UID != "" && apiObj.GetUID() != controllerKey.UID) {
		// The informer may be missing the object or still hold its previous
		// incarnation. Read once and validate the API object's UID below.
		apiObj, err = f.getWellKnownController(ctx, controllerKey)
		if err != nil {
			return nil, err
		}
	}
	return getOwnerControllerForObject(apiObj, controllerKey)
}

// getOwnerControllerForObject rejects a different UID when the key identifies a
// specific object, then returns its controlling owner. Target references may omit UID.
func getOwnerControllerForObject(obj metav1.Object, key ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	if key.UID != "" && obj.GetUID() != key.UID {
		return nil, fmt.Errorf("%s %s/%s has UID %q, expected %q", key.Kind, key.Namespace, key.Name, obj.GetUID(), key.UID)
	}
	return getOwnerController(obj.GetOwnerReferences(), key.Namespace), nil
}

// getWellKnownController reads a supported controller through its typed client
// using the caller's context, without API discovery or retries at this layer.
func (f *controllerFetcher) getWellKnownController(ctx context.Context, key ControllerKeyWithAPIVersion) (metav1.Object, error) {
	switch wellKnownController(key.Kind) {
	case daemonSet:
		return f.kubeClient.AppsV1().DaemonSets(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case deployment:
		return f.kubeClient.AppsV1().Deployments(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case statefulSet:
		return f.kubeClient.AppsV1().StatefulSets(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case replicaSet:
		return f.kubeClient.AppsV1().ReplicaSets(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case job:
		return f.kubeClient.BatchV1().Jobs(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case cronJob:
		return f.kubeClient.BatchV1().CronJobs(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	case replicationController:
		return f.kubeClient.CoreV1().ReplicationControllers(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("unknown controller kind %q", key.Kind)
	}
}

// getParentOfController resolves a well-known controller's owner or reads the scale
// subresource for other kinds. An absent scale subresource ends that ownership chain.
func (f *controllerFetcher) getParentOfController(ctx context.Context, controllerKey ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	kind := wellKnownController(controllerKey.Kind)
	informer, exists := f.informersMap[kind]
	if exists {
		return f.getParentOfWellKnownController(ctx, informer, controllerKey)
	}

	groupKind, err := controllerKey.groupKind()
	if err != nil {
		return nil, err
	}

	owner, err := f.getOwnerForScaleResource(ctx, groupKind, controllerKey)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unhandled targetRef %s / %s / %s, last error %w",
			controllerKey.ApiVersion, controllerKey.Kind, controllerKey.Name, err)
	}

	return owner, nil
}

func (c *ControllerKeyWithAPIVersion) groupKind() (schema.GroupKind, error) {
	// TODO: cache response
	groupVersion, err := schema.ParseGroupVersion(c.ApiVersion)
	if err != nil {
		return schema.GroupKind{}, err
	}

	groupKind := schema.GroupKind{
		Group: groupVersion.Group,
		Kind:  c.Kind,
	}

	return groupKind, nil
}

func (f *controllerFetcher) isWellKnown(key *ControllerKeyWithAPIVersion) bool {
	kind := wellKnownController(key.Kind)
	_, exists := f.informersMap[kind]
	return exists
}

func (f *controllerFetcher) getScaleForResource(ctx context.Context, namespace string, groupResource schema.GroupResource, name string) (controller *autoscalingv1.Scale, err error) {
	if ok, scale, err := f.scaleSubresourceCacheStorage.Get(namespace, groupResource, name); ok {
		return scale, err
	}
	scale, err := f.scaleNamespacer.Scales(namespace).Get(ctx, groupResource, name, metav1.GetOptions{})
	f.scaleSubresourceCacheStorage.Insert(namespace, groupResource, name, scale, err)
	return scale, err
}

func (f *controllerFetcher) isWellKnownOrScalable(ctx context.Context, key *ControllerKeyWithAPIVersion) bool {
	if f.isWellKnown(key) {
		return true
	}

	// if not well known check if it supports scaling
	groupKind, err := key.groupKind()
	if err != nil {
		klog.ErrorS(err, "Could not find groupKind", "object", klog.KRef(key.Namespace, key.Name))
		return false
	}

	if wellKnownController(groupKind.Kind) == node {
		return false
	}

	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		klog.ErrorS(err, "Could not find mappings", "groupKind", groupKind, "object", klog.KRef(key.Namespace, key.Name))
		return false
	}

	for _, mapping := range mappings {
		groupResource := mapping.Resource.GroupResource()
		scale, err := f.getScaleForResource(ctx, key.Namespace, groupResource, key.Name)
		if err == nil && scale != nil {
			return true
		}
	}
	return false
}

// getOwnerForScaleResource finds a scale subresource through the REST mappings and
// validates its UID before returning its owner. Nodes cannot be controller owners.
func (f *controllerFetcher) getOwnerForScaleResource(ctx context.Context, groupKind schema.GroupKind, key ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	if wellKnownController(groupKind.Kind) == node {
		// Some pods specify nodes as their owners. This causes performance problems
		// in big clusters when VPA tries to get all nodes. We know nodes aren't
		// valid controllers so we can skip trying to fetch them.
		return nil, ErrNodeInvalidOwner
	}
	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		return nil, err
	}
	var lastError error
	for _, mapping := range mappings {
		groupResource := mapping.Resource.GroupResource()
		scale, err := f.getScaleForResource(ctx, key.Namespace, groupResource, key.Name)
		if err == nil {
			return getOwnerControllerForObject(scale, key)
		}
		lastError = err
	}

	// nothing found, apparently the resource doesn't support scale (or we lack RBAC)
	return nil, lastError
}

func (f *controllerFetcher) FindTopMostWellKnownOrScalable(ctx context.Context, key *ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	if key == nil {
		return nil, nil
	}

	var topMostWellKnownOrScalable *ControllerKeyWithAPIVersion

	wellKnownOrScalable := f.isWellKnownOrScalable(ctx, key)
	if wellKnownOrScalable {
		topMostWellKnownOrScalable = key
	}

	visited := make(map[ControllerKeyWithAPIVersion]bool)
	visited[*key] = true
	for {
		owner, err := f.getParentOfController(ctx, *key)
		if err != nil {
			return nil, err
		}

		if owner == nil {
			return topMostWellKnownOrScalable, nil
		}

		wellKnownOrScalable = f.isWellKnownOrScalable(ctx, owner)
		if wellKnownOrScalable {
			topMostWellKnownOrScalable = owner
		}

		_, alreadyVisited := visited[*owner]
		if alreadyVisited {
			return nil, errors.New("cycle detected in ownership chain")
		}
		visited[*key] = true

		key = owner
	}
}
