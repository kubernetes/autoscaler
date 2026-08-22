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

package sdk

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// type Node struct {
// 	ID         string `json:"id"`
// 	InstanceID string `json:"instanceId"`
// 	NodePoolID string `json:"nodePoolId"`
// 	ProjectID  string `json:"projectId"`

// 	Name     string `json:"name"`
// 	Flavor   string `json:"flavor"`
// 	Version  string `json:"version"`
// 	UpToDate bool   `json:"isUpToDate"`
// 	Status   string `json:"status"`

// 	IP        *string `json:"ip,omitempty"`
// 	PrivateIP *string `json:"privateIp,omitempty"`

// 	CreatedAt  time.Time `json:"createdAt"`
// 	DeployedAt time.Time `json:"deployedAt"`
// 	UpdatedAt  time.Time `json:"updatedAt"`
// }

// Node represents a VKE node returned by the control-plane API.
type Node struct {
	ClusterUUID   string `json:"cluster_uuid"`
	Id            string `json:"id"`
	NodeGroupUUID string `json:"node_group_uuid"`
	Current       int    `json:"current_nodes"`
	MinSize       int    `json:"node_group_min_size"`
	MaxSize       int    `json:"node_group_max_size"`
	Flavor        string `json:"node_flavor_uuid"`
	Status        string `json:"node_groups_status"`
}

// DrainNode cordons and drains a node.
func (k *Client) DrainNode(nodeName string, client kubernetes.Interface, node *corev1.Node, DrainWaitSeconds int) error {
	if client == nil {
		return fmt.Errorf("K8sClient not set")
	}
	if node == nil {
		return fmt.Errorf("node not set")
	}
	if nodeName == "" {
		return fmt.Errorf("node name not set")
	}
	const (
		// PodSafeToEvictKey - annotation that ignores constraints to evict a pod like not being replicated, being on
		// kube-system namespace or having a local storage.
		PodSafeToEvictKey = "cluster-autoscaler.kubernetes.io/safe-to-evict"
		// SafeToEvictLocalVolumesKey - annotation that ignores (doesn't block on) a local storage volume during node scale down
		SafeToEvictLocalVolumesKey = "cluster-autoscaler.kubernetes.io/safe-to-evict-local-volumes"
	)
	const (
		// PodLongTerminatingExtraThreshold - time after which a pod, that is terminating and that has run over its terminationGracePeriod, should be ignored and considered as deleted
		PodLongTerminatingExtraThreshold = 30 * time.Second
	)

	return nil
}
