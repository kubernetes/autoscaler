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

package rancher

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

const (
	clusterAPIGroup               = "cluster.x-k8s.io"
	machineDeleteAnnotationKey    = clusterAPIGroup + "/delete-machine"
	machinePhaseProvisioning      = "Provisioning"
	machinePhasePending           = "Pending"
	machinePhaseDeleting          = "Deleting"
	machineDeploymentNameLabelKey = clusterAPIGroup + "/deployment-name"
	machineResourceName           = "machines"
	machineNodeAnnotationKey      = "cluster.x-k8s.io/machine"
)

func getAPIGroupPreferredVersion(client discovery.DiscoveryInterface, apiGroup string) (string, error) {
	groupList, err := client.ServerGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get ServerGroups: %v", err)
	}

	for _, group := range groupList.Groups {
		if group.Name == apiGroup {
			return group.PreferredVersion.Version, nil
		}
	}

	return "", fmt.Errorf("failed to find API group %q", apiGroup)
}

func machineGVR(version string) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    clusterAPIGroup,
		Version:  version,
		Resource: machineResourceName,
	}
}
