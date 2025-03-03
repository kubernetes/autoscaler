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

package target

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

const (
	discoveryResetPeriod time.Duration = 5 * time.Minute
)

// VpaTargetSelectorFetcher gets a labelSelector used to gather Pods controlled by the given VPA.
type VpaTargetSelectorFetcher interface {
	// Fetch returns a labelSelector used to gather Pods controlled by the given VPA.
	// If error is nil, the returned labelSelector is not nil.
	Fetch(ctx context.Context, vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, error)
}

type wellKnownController string

const (
	daemonSet             wellKnownController = "DaemonSet"
	deployment            wellKnownController = "Deployment"
	replicaSet            wellKnownController = "ReplicaSet"
	statefulSet           wellKnownController = "StatefulSet"
	replicationController wellKnownController = "ReplicationController"
	job                   wellKnownController = "Job"
	cronJob               wellKnownController = "CronJob"
)

// NewVpaTargetSelectorFetcher returns new instance of VpaTargetSelectorFetcher
func NewVpaTargetSelectorFetcher(config *rest.Config, kubeClient kube_client.Interface, factory informers.SharedInformerFactory) VpaTargetSelectorFetcher {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		klog.ErrorS(err, "Could not create discoveryClient")
		os.Exit(255)
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
			klog.ErrorS(nil, "Could not sync cache for "+string(kind))
			os.Exit(255)
		} else {
			klog.InfoS("Initial sync completed", "kind", kind)
		}
	}

	scaleNamespacer := scale.New(restClient, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	return &vpaTargetSelectorFetcher{
		scaleNamespacer: scaleNamespacer,
		mapper:          mapper,
		informersMap:    informersMap,
	}
}

// vpaTargetSelectorFetcher implements VpaTargetSelectorFetcher interface
// by querying API server for the controller pointed by VPA's targetRef
type vpaTargetSelectorFetcher struct {
	scaleNamespacer scale.ScalesGetter
	mapper          apimeta.RESTMapper
	informersMap    map[wellKnownController]cache.SharedIndexInformer
}

func (f *vpaTargetSelectorFetcher) Fetch(ctx context.Context, vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	if vpa.Spec.TargetRef == nil {
		return nil, fmt.Errorf("targetRef not defined. If this is a v1beta1 object, switch to v1.")
	}
	kind := wellKnownController(vpa.Spec.TargetRef.Kind)
	informer, exists := f.informersMap[kind]
	if exists {
		return getLabelSelector(informer, vpa.Spec.TargetRef.Kind, vpa.Namespace, vpa.Spec.TargetRef.Name)
	}

	// not on a list of known controllers, use scale sub-resource
	// TODO: cache response
	groupVersion, err := schema.ParseGroupVersion(vpa.Spec.TargetRef.APIVersion)
	if err != nil {
		return nil, err
	}
	groupKind := schema.GroupKind{
		Group: groupVersion.Group,
		Kind:  vpa.Spec.TargetRef.Kind,
	}

	selector, err := f.getLabelSelectorFromResource(ctx, groupKind, vpa.Namespace, vpa.Spec.TargetRef.Name)
	if err != nil {
		return nil, fmt.Errorf("Unhandled targetRef %s / %s / %s, last error %v",
			vpa.Spec.TargetRef.APIVersion, vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name, err)
	}
	return selector, nil
}

func getLabelSelector(informer cache.SharedIndexInformer, kind, namespace, name string) (labels.Selector, error) {
	obj, exists, err := informer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s %s/%s does not exist", kind, namespace, name)
	}
	switch apiObj := obj.(type) {
	case (*appsv1.DaemonSet):
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.Deployment):
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.StatefulSet):
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.ReplicaSet):
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*batchv1.Job):
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*batchv1.CronJob):
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.JobTemplate.Spec.Template.Labels))
	case (*corev1.ReplicationController):
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.Selector))
	}
	return nil, fmt.Errorf("don't know how to read label selector")
}

func (f *vpaTargetSelectorFetcher) getLabelSelectorFromResource(
	ctx context.Context, groupKind schema.GroupKind, namespace, name string,
) (labels.Selector, error) {
	mappings, err := f.mapper.RESTMappings(groupKind)
	if err != nil {
		return nil, err
	}

	var lastError error
	for _, mapping := range mappings {
		groupResource := mapping.Resource.GroupResource()
		scale, err := f.scaleNamespacer.Scales(namespace).Get(ctx, groupResource, name, metav1.GetOptions{})
		if err == nil {
			if scale.Status.Selector == "" {
				return nil, fmt.Errorf("Resource %s/%s has an empty selector for scale sub-resource", namespace, name)
			}
			selector, err := labels.Parse(scale.Status.Selector)
			if err != nil {
				return nil, err
			}
			return selector, nil
		}
		lastError = err
	}

	// nothing found, apparently the resource does support scale (or we lack RBAC)
	return nil, lastError
}
