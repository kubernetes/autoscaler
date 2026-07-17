/*
Copyright 2024 The Kubernetes Authors.

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

package provider

import (
	"k8s.io/apimachinery/pkg/labels"
)

// allObjectsLister is used by Provider to list DRA objects instead of directly using the API lister
// interfaces, so that test lister implementations don't have to implement the full API lister interfaces.
// Provider only needs to be able to list all DRA objects in the cluster, there shouldn't ever be a need to
// list a subset. The API lister interfaces, on the other hand, require implementing subset selection methods.
type allObjectsLister[O any] interface {
	ListAll() ([]O, error)
}

// apiLister is satisfied by any API object lister. Only defined as a type constraint for allObjectsApiLister.
type apiLister[O any] interface {
	List(selector labels.Selector) ([]O, error)
}

// allObjectsApiLister implements allObjectsLister by wrapping an API object lister.
type allObjectsApiLister[L apiLister[O], O any] struct {
	apiLister L
}

// ListAll lists all objects.
func (l *allObjectsApiLister[L, O]) ListAll() ([]O, error) {
	return l.apiLister.List(labels.Everything())
}
