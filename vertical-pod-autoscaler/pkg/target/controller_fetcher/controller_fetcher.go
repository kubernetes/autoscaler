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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingapi "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
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
}

// ControllerFetcher is responsible for finding the topmost well-known or scalable controller
type ControllerFetcher interface {
	// FindTopMostWellKnownOrScalable returns topmost well-known or scalable controller. Error is returned if controller cannot be found.
	FindTopMostWellKnownOrScalable(ctx context.Context, controller *ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error)
}

type controllerFetcher struct {
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
			klog.V(5).Info("Starting to refresh entries in controllerFetchers scaleSubresourceCacheStorage")
			for _, item := range keysToRefresh {
				scale, err := f.scaleNamespacer.Scales(item.namespace).Get(context.TODO(), item.groupResource, item.name, metav1.GetOptions{})
				f.scaleSubresourceCacheStorage.Refresh(item.namespace, item.groupResource, item.name, scale, err)
			}
			klog.V(5).Infof("Finished refreshing %d entries in controllerFetchers scaleSubresourceCacheStorage", len(keysToRefresh))
			f.scaleSubresourceCacheStorage.RemoveExpired()
		}
	}
}

func (f *controllerFetcher) Start(ctx context.Context, loopPeriod time.Duration) {
	go f.periodicallyRefreshCache(ctx, loopPeriod)
}

// NewControllerFetcher returns a new instance of controllerFetcher
func NewControllerFetcher(config *rest.Config, kubeClient kube_client.Interface, factory informers.SharedInformerFactory, betweenRefreshes, lifeTime time.Duration, jitterFactor float64) *controllerFetcher {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		klog.Fatalf("Could not create discoveryClient: %v", err)
	}
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	restClient := kubeClient.CoreV1().RESTClient()
	cachedDiscoveryClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)
	go wait.Until(func() {
		mapper.Reset()
	}, discoveryResetPeriod, make(chan struct{}))

	informersMap := map[wellKnownController]cache.SharedIndexInformer{
		daemonSet:             factory.Apps().V1().DaemonSets().Informer(),
		deployment:            factory.Apps().V1().Deployments().Informer(),
		replicaSet:            factory.Apps().V1().ReplicaSets().Informer(),
		statefulSet:           factory.Apps().V1().StatefulSets().Informer(),
		replicationController: factory.Core().V1().ReplicationControllers().Informer(),
		job:                   factory.Batch().V1().Jobs().Informer(),
		cronJob:               factory.Batch().V1().CronJobs().Informer(),
	}

	for kind, informer := range informersMap {
		stopCh := make(chan struct{})
		go informer.Run(stopCh)
		synced := cache.WaitForCacheSync(stopCh, informer.HasSynced)
		if !synced {
			klog.Warningf("Could not sync cache for %s: %v", kind, err)
		} else {
			klog.Infof("Initial sync of %s completed", kind)
		}
	}

	scaleNamespacer := scale.New(restClient, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	return &controllerFetcher{
		scaleNamespacer:              scaleNamespacer,
		mapper:                       mapper,
		informersMap:                 informersMap,
		scaleSubresourceCacheStorage: newControllerCacheStorage(betweenRefreshes, lifeTime, jitterFactor),
	}
}

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
			}
		}
	}
	return nil
}

func getParentOfWellKnownController(informer cache.SharedIndexInformer, controllerKey ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	namespace := controllerKey.Namespace
	name := controllerKey.Name
	kind := controllerKey.Kind

	obj, exists, err := informer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s %s/%s does not exist", kind, namespace, name)
	}
	switch apiObj := obj.(type) {
	case (*appsv1.DaemonSet):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*appsv1.Deployment):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*appsv1.StatefulSet):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*appsv1.ReplicaSet):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*batchv1.Job):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*batchv1.CronJob):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	case (*corev1.ReplicationController):
		return getOwnerController(apiObj.OwnerReferences, namespace), nil
	}

	return nil, fmt.Errorf("don't know how to read owner controller")
}

func (f *controllerFetcher) getParentOfController(ctx context.Context, controllerKey ControllerKeyWithAPIVersion) (*ControllerKeyWithAPIVersion, error) {
	kind := wellKnownController(controllerKey.Kind)
	informer, exists := f.informersMap[kind]
	if exists {
		return getParentOfWellKnownController(informer, controllerKey)
	}

	groupKind, err := controllerKey.groupKind()
	if err != nil {
		return nil, err
	}

	owner, err := f.getOwnerForScaleResource(ctx, groupKind, controllerKey.Namespace, controllerKey.Name)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Unhandled targetRef %s / %s / %s, last error %v",
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
		Kind:  c.ControllerKey.Kind,
	}

	return groupKind, nil
}

func (f *controllerFetcher) isWellKnown(key *ControllerKeyWithAPIVersion) bool {
	kind := wellKnownController(key.ControllerKey.Kind)
	_, exists := f.informersMap[kind]
	return exists
}

func (f *controllerFetcher) getScaleForResource(ctx context.Context, namespace string, groupResource schema.GroupResource, name string) (controller *autoscalingapi.Scale, err error) {
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

	//if not well known check if it supports scaling
	groupKind, err := key.groupKind()
	if err != nil {
		klog.Errorf("Could not find groupKind for %s/%s: %v", key.Namespace, key.Name, err)
		return false
	}

	if wellKnownController(groupKind.Kind) == node {
		return false
	}

	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		klog.Errorf("Could not find mappings for %s: %v", groupKind, err)
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

func (f *controllerFetcher) getOwnerForScaleResource(ctx context.Context, groupKind schema.GroupKind, namespace, name string) (*ControllerKeyWithAPIVersion, error) {
	if wellKnownController(groupKind.Kind) == node {
		// Some pods specify nodes as their owners. This causes performance problems
		// in big clusters when VPA tries to get all nodes. We know nodes aren't
		// valid controllers so we can skip trying to fetch them.
		return nil, fmt.Errorf("node is not a valid owner")
	}
	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		return nil, err
	}
	var lastError error
	for _, mapping := range mappings {
		groupResource := mapping.Resource.GroupResource()
		scale, err := f.getScaleForResource(ctx, namespace, groupResource, name)
		if err == nil {
			return getOwnerController(scale.OwnerReferences, namespace), nil
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
			return nil, fmt.Errorf("Cycle detected in ownership chain")
		}
		visited[*key] = true

		key = owner
	}
}
