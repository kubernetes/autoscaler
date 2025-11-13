package snapshot

import (
	v1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	schedulerframework "k8s.io/kube-scheduler/framework"
)

type snapshotDeviceClassResolver struct {
	// resourceName2class maps extended resource name to device class
	resourceName2class map[v1.ResourceName]*resourceapi.DeviceClass
}

var _ schedulerframework.DeviceClassResolver = &snapshotDeviceClassResolver{}

func newSnapshotDeviceClassResolver(snapshot *Snapshot) schedulerframework.DeviceClassResolver {
	resourceName2class := make(map[v1.ResourceName]*resourceapi.DeviceClass)
	for _, class := range snapshot.listDeviceClasses() {
		classResourceName := class.Name
		extendedResourceName := class.Spec.ExtendedResourceName
		if extendedResourceName != nil {
			resourceName2class[v1.ResourceName(*extendedResourceName)] = class
		}

		// Also add the default extended resource name
		defaultResourceName := v1.ResourceName(resourceapi.ResourceDeviceClassPrefix + classResourceName)
		resourceName2class[defaultResourceName] = class
	}
	return snapshotDeviceClassResolver{resourceName2class: resourceName2class}
}

// GetDeviceClass returns the device class for the given extended resource name
func (s snapshotDeviceClassResolver) GetDeviceClass(resourceName v1.ResourceName) *resourceapi.DeviceClass {
	class, found := s.resourceName2class[resourceName]
	if !found {
		return nil
	}
	return class
}
