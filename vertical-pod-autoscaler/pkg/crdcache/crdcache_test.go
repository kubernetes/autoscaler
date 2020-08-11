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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	dynamic "k8s.io/client-go/dynamic/fake"
)

var (
	scaleLabelFields   = []string{"status", "scaleLabelSelector"}
	scaleLabelSelector = "." + strings.Join(scaleLabelFields, ".")
	scaleSubResource   = &extensionsv1.CustomResourceSubresources{
		Scale: &extensionsv1.CustomResourceSubresourceScale{
			LabelSelectorPath: &scaleLabelSelector,
		},
	}
	newCrd = func(name string, subResource *extensionsv1.CustomResourceSubresources) *extensionsv1.CustomResourceDefinition {
		return test.CustomResourceDefinition().WithName(name).WithSubresources(subResource).
			WithGroupVersion("test.org", "v1").Get()
	}
	fooCrd         = newCrd("foo", scaleSubResource)
	barCrd         = newCrd("bar", scaleSubResource)
	notScalableCrd = newCrd("not-scalable", nil)

	newUnstructured = func(name, namespace string, crd *extensionsv1.CustomResourceDefinition, scaleLabels string,
		owners []metav1.OwnerReference) *unstructured.Unstructured {
		return test.Unstructured().
			WithName(name).WithNamespace(namespace).
			WithApiVersionKind(test.CrdApiVersionAndKind(crd)).
			AddNestedField(scaleLabels, scaleLabelFields...).
			WithOwnerReferences(owners).Get()
	}

	foo1         = newUnstructured("foo1", "ns1", fooCrd, "name=foo1", nil)
	foo2         = newUnstructured("foo2", "ns2", fooCrd, "name=foo2", nil)
	bar1         = newUnstructured("bar1", "ns2", barCrd, "app=bar", nil)
	notScalable1 = newUnstructured("not-scalable1", "ns1", notScalableCrd, "", nil)
	objNotExist1 = newUnstructured("obj-not-exist1", "ns2", fooCrd, "name=obj-not-exist1", nil)
	objHasOwner  = newUnstructured("has-owner1", "ns1", fooCrd, "",
		[]metav1.OwnerReference{{APIVersion: "v2", Kind: "kind1", Name: "owner"}})

	fooGvr, barGvr = getGroupVersionResource(fooCrd), getGroupVersionResource(barCrd)
	notScalableGvr = getGroupVersionResource(notScalableCrd)
)

func TestFetchSelector(t *testing.T) {
	crdCache := newFakeCrdCache()
	for i, testCase := range []struct {
		gvr       schema.GroupVersionResource
		obj       *unstructured.Unstructured
		expectErr error
	}{
		{gvr: fooGvr, obj: foo1},
		{gvr: fooGvr, obj: foo2},
		{gvr: barGvr, obj: bar1},
		{gvr: notScalableGvr, obj: notScalable1, expectErr: fmt.Errorf("resource %s hasn't scale sub-resource", notScalableGvr.String())},
		{gvr: fooGvr, obj: objNotExist1, expectErr: fmt.Errorf(`foos.test.org "obj-not-exist1" not found`)},
	} {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			selector, err := crdCache.FetchSelector(testCase.gvr, testCase.obj.GetNamespace(), testCase.obj.GetName())
			if testCase.expectErr == nil {
				assert.NoError(t, err)
				expectLabelSelector, found, err := unstructured.NestedString(testCase.obj.Object, scaleLabelFields...)
				assert.True(t, found)
				assert.NoError(t, err)
				assert.Equal(t, expectLabelSelector, selector.String())
			} else {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectErr.Error(), err.Error())
			}
		})
	}
}

func TestIsScalable(t *testing.T) {
	crdCache := newFakeCrdCache()
	for i, testCase := range []struct {
		gvr            schema.GroupVersionResource
		obj            *unstructured.Unstructured
		expectErr      error
		expectScalable bool
	}{
		{gvr: fooGvr, obj: foo1, expectScalable: true},
		{gvr: fooGvr, obj: foo2, expectScalable: true},
		{gvr: barGvr, obj: bar1, expectScalable: true},
		{gvr: notScalableGvr, obj: notScalable1},
		{gvr: fooGvr, obj: objNotExist1, expectErr: fmt.Errorf(`foos.test.org "obj-not-exist1" not found`)},
	} {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			isScalable, err := crdCache.IsScalable(testCase.gvr, testCase.obj.GetNamespace(), testCase.obj.GetName())
			if testCase.expectErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectScalable, isScalable)
			} else {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectErr.Error(), err.Error())
			}
		})
	}
}

func TestGetOwnerReferences(t *testing.T) {
	crdCache := newFakeCrdCache()
	owners, err := crdCache.GetOwnerReferences(fooGvr, objHasOwner.GetNamespace(), objHasOwner.GetName())
	assert.NoError(t, err)
	assert.Equal(t, objHasOwner.GetOwnerReferences(), owners)
}

func newFakeCrdCache() CrdCache {
	dynamicClient := dynamic.NewSimpleDynamicClient(runtime.NewScheme(), foo1, foo2, bar1, notScalable1, objHasOwner)
	extensionClient := extensionclient.NewSimpleClientset(fooCrd, barCrd, notScalableCrd)
	return NewCrdCache(dynamicClient, extensionClient, time.Minute*10)
}

func getGroupVersionResource(crd *extensionsv1.CustomResourceDefinition) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Version:  crd.Spec.Version,
		Resource: crd.Spec.Names.Plural,
	}
}
