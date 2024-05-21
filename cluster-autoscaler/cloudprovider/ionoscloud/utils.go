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

package ionoscloud

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
)

const (
	// ProviderIDPrefix is the prefix of the provider id of a Kubernetes node object.
	ProviderIDPrefix = "ionos://"
	// ErrorCodeUnknownState is set if the IonosCloud Kubernetes instace has an unknown state.
	ErrorCodeUnknownState = "UNKNOWN_STATE"
)

var errMissingNodeID = errors.New("missing node ID")

// convertToInstanceID converts an IonosCloud kubernetes node Id to a cloudprovider.Instance Id.
func convertToInstanceID(nodeID string) string {
	return fmt.Sprintf("%s%s", ProviderIDPrefix, nodeID)
}

// convertToNodeID converts a cloudprovider.Instance Id to an IonosCloud kubernetes node Id.
func convertToNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, ProviderIDPrefix)
}

// convertToInstances converts a list IonosCloud kubernetes nodes to a list of cloudprovider.Instances.
func convertToInstances(nodes []ionos.KubernetesNode) ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, node := range nodes {
		instance, err := convertToInstance(node)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

// to Instance converts an IonosCloud kubernetes node to a cloudprovider.Instance.
func convertToInstance(node ionos.KubernetesNode) (cloudprovider.Instance, error) {
	if node.Id == nil {
		return cloudprovider.Instance{}, errMissingNodeID
	}
	return cloudprovider.Instance{
		Id:     convertToInstanceID(*node.Id),
		Status: convertToInstanceStatus(*node.Metadata.State),
	}, nil
}

// convertToInstanceStatus converts an IonosCloud kubernetes node state to a *cloudprovider.InstanceStatus.
func convertToInstanceStatus(nodeState string) *cloudprovider.InstanceStatus {
	st := &cloudprovider.InstanceStatus{}
	switch nodeState {
	case K8sNodeStateProvisioning, K8sNodeStateProvisioned:
		st.State = cloudprovider.InstanceCreating
	case K8sNodeStateTerminating, K8sNodeStateRebuilding:
		st.State = cloudprovider.InstanceDeleting
	case K8sNodeStateReady:
		st.State = cloudprovider.InstanceRunning
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    ErrorCodeUnknownState,
			ErrorMessage: "Unknown node state: " + nodeState,
		}
	}
	return st
}
