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
	"fmt"

	resourceapi "k8s.io/api/resource/v1"
)

type snapshotClassLister struct {
	snapshot *Snapshot
}

func (s snapshotClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return s.snapshot.listDeviceClasses(), nil
}

func (s snapshotClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	class, found := s.snapshot.getDeviceClass(className)
	if !found {
		return nil, fmt.Errorf("DeviceClass %q not found", className)
	}
	return class, nil
}
