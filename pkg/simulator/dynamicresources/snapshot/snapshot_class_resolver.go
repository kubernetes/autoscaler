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

package snapshot

import (
	v1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

type snapshotDeviceClassResolver struct {
	// deviceClassMap maps extended resource name to device class
	deviceClassMap map[v1.ResourceName]*resourceapi.DeviceClass
}

var _ schedulerinterface.DeviceClassResolver = &snapshotDeviceClassResolver{}

// newSnapshotDeviceClassResolver implements DeviceClassResolver for a snapshot
func newSnapshotDeviceClassResolver(snapshot *Snapshot) schedulerinterface.DeviceClassResolver {
	deviceClassMap := make(map[v1.ResourceName]*resourceapi.DeviceClass)
	for _, class := range snapshot.listDeviceClasses() {
		if class != nil {
			deviceClassMap[v1.ResourceName(resourceapi.ResourceDeviceClassPrefix+class.Name)] = class
			extendedResourceName := class.Spec.ExtendedResourceName
			if extendedResourceName != nil {
				deviceClassMap[v1.ResourceName(*extendedResourceName)] = class
			}
		}
	}
	return snapshotDeviceClassResolver{deviceClassMap: deviceClassMap}
}

// GetDeviceClass returns the device class name for the given extended resource name
func (s snapshotDeviceClassResolver) GetDeviceClass(resourceName v1.ResourceName) *resourceapi.DeviceClass {
	class, ok := s.deviceClassMap[resourceName]
	if !ok {
		return nil
	}
	return class
}
