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

package deallocate

// PolicyNodeGroup is a wrapper for a nodegroup that can have scaling policies
type PolicyNodeGroup interface {
	// ScaleDownPolicy returns whether this nodegroup instances will be deleted or deallocated on scale down
	ScaleDownPolicy() ScaleDownPolicy
}

// ScaleDownPolicy is a string representation of the ScaleDownPolicy
type ScaleDownPolicy string

const (
	// Delete means that on scale down, nodes will be deleted
	Delete ScaleDownPolicy = "Delete"
	// Deallocate means that on scale down, nodes will be deallocated and on scale-up they will
	// be attempted to be started first
	Deallocate ScaleDownPolicy = "Deallocate"
)
