/*
Copyright 2025 The Kubernetes Authors.

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

package utils

// InPlaceDecision is the type of decision that can be made for a pod.
type InPlaceDecision string

// ResizeStatus is the status of resize api call.
type ResizeStatus string

const (
	// InPlaceApproved means we can in-place update the pod.
	InPlaceApproved InPlaceDecision = "InPlaceApproved"
	// InPlaceDeferred means we can't in-place update the pod right now, but we will wait for the next loop to check for in-placeability again
	InPlaceDeferred InPlaceDecision = "InPlaceDeferred"
	// InPlaceEvict means we will attempt to evict the pod.
	InPlaceEvict InPlaceDecision = "InPlaceEvict"
	// InPlaceInfeasible means we can't in-place update the pod right now but we want to retry
	InPlaceInfeasible InPlaceDecision = "InPlaceInfeasible"
	// ResizeStatusDeferred indicates the resize is deferred by kubelet
	ResizeStatusDeferred ResizeStatus = "Deferred"
	// ResizeStatusInfeasible indicates the resize cannot be performed (e.g., insufficient node resources)
	ResizeStatusInfeasible ResizeStatus = "Infeasible"
	// ResizeStatusInProgress indicates the resize is currently being applied
	ResizeStatusInProgress ResizeStatus = "InProgress"
	// ResizeStatusError indicates an error occurred during resize
	ResizeStatusError ResizeStatus = "Error"
	// ResizeStatusUnknown indicates an unknown resize status
	ResizeStatusUnknown ResizeStatus = "Unknown"
	// ResizeStatusNone indicates no resize operation is pending
	ResizeStatusNone ResizeStatus = "None"
)
