/*
Copyright 2020 The Kubernetes Authors.

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

package crdcache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	extensionclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extensioninformer "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	extensionlister "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// CrdCache caches CustomResourceDefinition objects via dynamic informers to make searching for VPA controller and target selectors fast.
type CrdCache interface {
	// FetchSelector returns a labelSelector used to gather Pods controlled by the given crd object (groupVersionResource, name and namespace).
	// If the given groupVersionResource is not cached before, it will start caching all objects of it.
	// If error is nil, the returned labelSelector is not nil.
	FetchSelector(gvr schema.GroupVersionResource, namespace, name string) (labels.Selector, error)

	// IsScalable returns true if the given crd object (groupVersionResource, name and namespace) is scalable
	IsScalable(gvr schema.GroupVersionResource, namespace, name string) (bool, error)

	// GetOwnerReferences returns the owner references of the given crd object (groupVersionResource, name and namespace)
	GetOwnerReferences(gvr schema.GroupVersionResource, namespace, name string) ([]metav1.OwnerReference, error)
}

// NewCrdCacheForConfig returns new instance of CrdCache
func NewCrdCacheForConfig(config *rest.Config, resyncTime time.Duration) CrdCache {
	dynamicClient := dynamic.NewForConfigOrDie(config)
	extensionClient := extensionclient.NewForConfigOrDie(config)
	return NewCrdCache(dynamicClient, extensionClient, resyncTime)
}

// NewCrdCache returns new instance of CrdCache
func NewCrdCache(dynamicClient dynamic.Interface, extensionClient extensionclient.Interface, resyncTime time.Duration) CrdCache {
	c := &crdCache{
		dynamicFactory:   dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, resyncTime),
		startedInformers: map[schema.GroupVersionResource]bool{},
	}
	stop := make(chan struct{})
	extensionFactory := extensioninformer.NewSharedInformerFactory(extensionClient, resyncTime)
	extensionInformer := extensionFactory.Apiextensions().V1beta1().CustomResourceDefinitions()
	extensionInformer.Informer() // call Informer to actually create an informer
	extensionFactory.Start(stop)
	extensionFactory.WaitForCacheSync(stop)
	c.extensionLister = extensionInformer.Lister()
	return c
}

type crdCache struct {
	dynamicFactory  dynamicinformer.DynamicSharedInformerFactory
	extensionLister extensionlister.CustomResourceDefinitionLister
	lock            sync.Mutex
	// startedInformers caches started GroupVersionResource informers
	startedInformers map[schema.GroupVersionResource]bool
}

func (c *crdCache) FetchSelector(gvr schema.GroupVersionResource, namespace, name string) (labels.Selector, error) {
	crd, err := c.extensionLister.Get(gvr.GroupResource().String())
	if err != nil {
		return nil, err
	}
	subRes := crd.Spec.Subresources
	if subRes == nil || subRes.Scale == nil || subRes.Scale.LabelSelectorPath == nil ||
		*subRes.Scale.LabelSelectorPath == "" {
		return nil, fmt.Errorf("resource %s hasn't scale sub-resource", gvr.String())
	}
	lister := c.getLister(gvr)
	obj, err := lister.ByNamespace(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("crd object %s %s/%s can't be converted to a unstructured object", gvr.String(), name, namespace)
	}
	selectorStr, err := selectorFromCustomResource(unstructuredObj, *subRes.Scale.LabelSelectorPath)
	if err != nil {
		return nil, fmt.Errorf("can't get label selector from crd object %s %s/%s: %v", gvr.String(), name, namespace, err)
	}
	selector, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector %s from crd object %s %s/%s: %v", selectorStr, gvr.String(), name, namespace, err)
	}
	return selector, nil
}

func (c *crdCache) IsScalable(gvr schema.GroupVersionResource, namespace, name string) (bool, error) {
	crd, err := c.extensionLister.Get(gvr.GroupResource().String())
	if err != nil {
		return false, err
	}
	subRes := crd.Spec.Subresources
	if subRes == nil || subRes.Scale == nil {
		return false, nil
	}
	lister := c.getLister(gvr)
	_, err = lister.ByNamespace(namespace).Get(name)
	return err == nil, err
}

func (c *crdCache) GetOwnerReferences(gvr schema.GroupVersionResource, namespace, name string) ([]metav1.OwnerReference, error) {
	lister := c.getLister(gvr)
	obj, err := lister.ByNamespace(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("crd object %s %s/%s can't be converted to a unstructured object", gvr.String(), name, namespace)
	}
	return unstructuredObj.GetOwnerReferences(), nil
}

// getLister gets the lister for the given GroupVersionResource and makes sure the informer have started and synced
func (c *crdCache) getLister(gvr schema.GroupVersionResource) cache.GenericLister {
	informer := c.dynamicFactory.ForResource(gvr)
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.startedInformers[gvr]; !ok {
		klog.Infof("Watching custom resource definition %s", gvr.String())
		stopCh := make(chan struct{})
		go informer.Informer().Run(stopCh)
		cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced)
		c.startedInformers[gvr] = true
	}
	return informer.Lister()
}

// selectorFromCustomResource gets label selector from the Unstructured object
func selectorFromCustomResource(cr *unstructured.Unstructured, labelSelectorPath string) (string, error) {
	labelSelectorPath = strings.TrimPrefix(labelSelectorPath, ".") // ignore leading period
	labelSelector, _, err := unstructured.NestedString(cr.UnstructuredContent(), strings.Split(labelSelectorPath, ".")...)
	return labelSelector, err
}
