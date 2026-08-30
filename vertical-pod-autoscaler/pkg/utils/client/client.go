/*
Copyright The Kubernetes Authors.

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

package client

import "k8s.io/apimachinery/pkg/api/meta"

// StripManagedFields is a cache.TransformFunc that removes managed fields
// metadata from objects to reduce memory usage.
var StripManagedFields = func(obj any) (any, error) {
	if accessor, err := meta.Accessor(obj); err == nil {
		if accessor.GetManagedFields() != nil {
			accessor.SetManagedFields(nil)
		}
	}
	return obj, nil
}
