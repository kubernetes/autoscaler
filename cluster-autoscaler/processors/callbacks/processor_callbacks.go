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

package callbacks

// ProcessorCallbacks is interface defining extra callback methods which can be called by processors used in extension points.
type ProcessorCallbacks interface {
	// DisableScaleDownForLoop disables scale down for current loop iteration
	DisableScaleDownForLoop()

	// SetExtraValue sets arbitrary value for given key. Value storage will be reset at the beginning of each loop iteration.
	// Arbitrary value storage is used to pass information between processors used in extension points.
	SetExtraValue(key string, value interface{})

	// GetExtraValue gets arbitrary value for given key. If value for given key is not found, found=false will be returned.
	GetExtraValue(key string) (value interface{}, found bool)
}
