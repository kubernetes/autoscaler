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

package snapshot

import (
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// Snapshot contains a snapshot of all DRA objects taken at a ~single point in time.
type Snapshot struct {
}

// ResourceClaims exposes the Snapshot as schedulerframework.ResourceClaimTracker, in order to interact with
// the scheduler framework.
func (s Snapshot) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return nil
}

// ResourceSlices exposes the Snapshot as schedulerframework.ResourceSliceLister, in order to interact with
// the scheduler framework.
func (s Snapshot) ResourceSlices() schedulerframework.ResourceSliceLister {
	return nil
}

// DeviceClasses exposes the Snapshot as schedulerframework.DeviceClassLister, in order to interact with
// the scheduler framework.
func (s Snapshot) DeviceClasses() schedulerframework.DeviceClassLister {
	return nil
}

// Clone returns a copy of this Snapshot that can be independently modified without affecting this Snapshot.
func (s Snapshot) Clone() Snapshot {
	return Snapshot{}
}
