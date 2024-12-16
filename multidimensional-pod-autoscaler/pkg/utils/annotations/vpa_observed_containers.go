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

package annotations

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// MpaObservedContainersLabel is a label used by the vpa observed containers annotation.
	MpaObservedContainersLabel = "vpaObservedContainers"
	listSeparator              = ", "
)

// GetMpaObservedContainersValue creates an annotation value for a given pod.
func GetMpaObservedContainersValue(pod *v1.Pod) string {
	containerNames := make([]string, len(pod.Spec.Containers))
	for i := range pod.Spec.Containers {
		containerNames[i] = pod.Spec.Containers[i].Name
	}
	return strings.Join(containerNames, listSeparator)
}

// ParseMpaObservedContainersValue returns list of containers
// based on a given mpa observed containers annotation value.
func ParseMpaObservedContainersValue(value string) ([]string, error) {
	if value == "" {
		return []string{}, nil
	}
	containerNames := strings.Split(value, listSeparator)
	for i := range containerNames {
		if errs := validation.IsDNS1123Label(containerNames[i]); len(errs) != 0 {
			return nil, fmt.Errorf("incorrect format: %s is not a valid container name: %v", containerNames[i], errs)
		}
	}
	return containerNames, nil
}
