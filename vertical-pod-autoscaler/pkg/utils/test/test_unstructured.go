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

package test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

// UnstructuredBuilder helps building Unstructured for tests.
type UnstructuredBuilder interface {
	WithName(name string) UnstructuredBuilder
	WithNamespace(namespace string) UnstructuredBuilder
	WithApiVersionKind(apiVersion, kind string) UnstructuredBuilder
	WithOwnerReferences(owners []metav1.OwnerReference) UnstructuredBuilder
	AddNestedField(value interface{}, fields ...string) UnstructuredBuilder
	Get() *unstructured.Unstructured
}

// Unstructured returns new UnstructuredBuilder.
func Unstructured() UnstructuredBuilder {
	return &unstructuredBuilder{}
}

type unstructuredBuilder struct {
	name         string
	namespace    string
	apiVersion   string
	kind         string
	owners       []metav1.OwnerReference
	nestedFields []nestedField
}

type nestedField struct {
	value  interface{}
	fields []string
}

func (u *unstructuredBuilder) WithName(name string) UnstructuredBuilder {
	r := *u
	r.name = name
	return &r
}

func (u *unstructuredBuilder) WithNamespace(namespace string) UnstructuredBuilder {
	r := *u
	r.namespace = namespace
	return &r
}

func (u *unstructuredBuilder) WithApiVersionKind(apiVersion, kind string) UnstructuredBuilder {
	r := *u
	r.apiVersion = apiVersion
	r.kind = kind
	return &r
}

func (u *unstructuredBuilder) WithOwnerReferences(owners []metav1.OwnerReference) UnstructuredBuilder {
	r := *u
	r.owners = owners
	return &r
}

func (u *unstructuredBuilder) AddNestedField(value interface{}, fields ...string) UnstructuredBuilder {
	r := *u
	r.nestedFields = append(r.nestedFields, nestedField{
		value:  value,
		fields: fields,
	})
	return &r
}

func (u *unstructuredBuilder) Get() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": u.apiVersion,
			"kind":       u.kind,
			"metadata": map[string]interface{}{
				"namespace": u.namespace,
				"name":      u.name,
			},
		},
	}
	obj.SetOwnerReferences(u.owners)
	for _, nf := range u.nestedFields {
		err := unstructured.SetNestedField(obj.Object, nf.value, nf.fields...)
		if err != nil {
			klog.Fatal(err)
		}
	}
	return obj
}
