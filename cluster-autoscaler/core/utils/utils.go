/*
Copyright 2016 The Kubernetes Authors.

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

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// // GetNodeInfoFromTemplate returns NodeInfo object built base on TemplateNodeInfo returned by NodeGroup.TemplateNodeInfo().
// func GetNodeInfoFromTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker, ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
// 	id := nodeGroup.Id()
// 	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
// 	if err != nil {
// 		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
// 	}

// 	labels.UpdateDeprecatedLabels(baseNodeInfo.Node().ObjectMeta.Labels)

// 	pods, err := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
// 	if err != nil {
// 		return nil, errors.ToAutoscalerError(errors.InternalError, err)
// 	}
// 	for _, podInfo := range baseNodeInfo.Pods {
// 		pods = append(pods, podInfo.Pod)
// 	}
// 	fullNodeInfo := schedulerframework.NewNodeInfo(pods...)
// 	fullNodeInfo.SetNode(baseNodeInfo.Node())
// 	sanitizedNodeInfo, typedErr := SanitizeNodeInfo(fullNodeInfo, id, ignoredTaints)
// 	if typedErr != nil {
// 		return nil, typedErr
// 	}
// 	return sanitizedNodeInfo, nil
// }

// isVirtualNode determines if the node is created by virtual kubelet
func isVirtualNode(node *apiv1.Node) bool {
	return node.ObjectMeta.Labels["type"] == "virtual-kubelet"
}

// // FilterOutNodesFromNotAutoscaledGroups return subset of input nodes for which cloud provider does not
// // return autoscaled node group.
// func FilterOutNodesFromNotAutoscaledGroups(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider) ([]*apiv1.Node, errors.AutoscalerError) {
// 	result := make([]*apiv1.Node, 0)

// 	for _, node := range nodes {
// 		// Exclude the virtual node here since it may have lots of resource and exceed the total resource limit
// 		if isVirtualNode(node) {
// 			continue
// 		}
// 		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
// 		if err != nil {
// 			return []*apiv1.Node{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
// 		}
// 		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
// 			result = append(result, node)
// 		}
// 	}
// 	return result, nil
// }

// DeepCopyNodeInfo clones the provided nodeInfo
func DeepCopyNodeInfo(nodeInfo *schedulerframework.NodeInfo) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	newPods := make([]*apiv1.Pod, 0)
	for _, podInfo := range nodeInfo.Pods {
		newPods = append(newPods, podInfo.Pod.DeepCopy())
	}

	// Build a new node info.
	newNodeInfo := schedulerframework.NewNodeInfo(newPods...)
	newNodeInfo.SetNode(nodeInfo.Node().DeepCopy())
	return newNodeInfo, nil
}

// SanitizeNodeInfo modify nodeInfos generated from templates to avoid using duplicated host names
func SanitizeNodeInfo(nodeInfo *schedulerframework.NodeInfo, nodeGroupName string, ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	// Sanitize node name.
	sanitizedNode, err := sanitizeTemplateNode(nodeInfo.Node(), nodeGroupName, ignoredTaints)
	if err != nil {
		return nil, err
	}

	// Update nodename in pods.
	sanitizedPods := make([]*apiv1.Pod, 0)
	for _, podInfo := range nodeInfo.Pods {
		sanitizedPod := podInfo.Pod.DeepCopy()
		sanitizedPod.Spec.NodeName = sanitizedNode.Name
		sanitizedPods = append(sanitizedPods, sanitizedPod)
	}

	// Build a new node info.
	sanitizedNodeInfo := schedulerframework.NewNodeInfo(sanitizedPods...)
	sanitizedNodeInfo.SetNode(sanitizedNode)
	return sanitizedNodeInfo, nil
}

func sanitizeTemplateNode(node *apiv1.Node, nodeGroup string, ignoredTaints taints.TaintKeySet) (*apiv1.Node, errors.AutoscalerError) {
	newNode := node.DeepCopy()
	nodeName := fmt.Sprintf("template-node-for-%s-%d", nodeGroup, rand.Int63())
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		if k != apiv1.LabelHostname {
			newNode.Labels[k] = v
		} else {
			newNode.Labels[k] = nodeName
		}
	}
	newNode.Name = nodeName
	newNode.Spec.Taints = taints.SanitizeTaints(newNode.Spec.Taints, ignoredTaints)
	return newNode, nil
}

func hasHardInterPodAffinity(affinity *apiv1.Affinity) bool {
	if affinity == nil {
		return false
	}
	if affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	if affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	return false
}

// GetNodeCoresAndMemory extracts cpu and memory resources out of Node object
func GetNodeCoresAndMemory(node *apiv1.Node) (int64, int64) {
	cores := getNodeResource(node, apiv1.ResourceCPU)
	memory := getNodeResource(node, apiv1.ResourceMemory)
	return cores, memory
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

// UpdateClusterStateMetrics updates metrics related to cluster state
func UpdateClusterStateMetrics(csr *clusterstate.ClusterStateRegistry) {
	if csr == nil || reflect.ValueOf(csr).IsNil() {
		return
	}
	metrics.UpdateClusterSafeToAutoscale(csr.IsClusterHealthy())
	readiness := csr.GetClusterReadiness()
	metrics.UpdateNodesCount(readiness.Ready, readiness.Unready, readiness.NotStarted, readiness.LongUnregistered, readiness.Unregistered)
}

// GetOldestCreateTime returns oldest creation time out of the pods in the set
func GetOldestCreateTime(pods []*apiv1.Pod) time.Time {
	oldest := time.Now()
	for _, pod := range pods {
		if oldest.After(pod.CreationTimestamp.Time) {
			oldest = pod.CreationTimestamp.Time
		}
	}
	return oldest
}

//// GetOldestCreateTimeWithGpu returns oldest creation time out of pods with GPU in the set
//func GetOldestCreateTimeWithGpu(pods []*apiv1.Pod) (bool, time.Time) {
//	oldest := time.Now()
//	gpuFound := false
//	for _, pod := range pods {
//		if gpu.PodRequestsGpu(pod) {
//			gpuFound = true
//			if oldest.After(pod.CreationTimestamp.Time) {
//				oldest = pod.CreationTimestamp.Time
//			}
//		}
//	}
//	return gpuFound, oldest
//}

// Get min size group
func GetMinSizeNodeGroup(kubeclient kube_client.Interface) int {
	var minSizeNodeGroup int
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get(ctx.Background(), "autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling configmap")
		klog.Fatalf("Failed to get information of autoscaling configmap: %v", err)
	}
	for k, v := range configmaps.Data {
		if k == "min_node_group_size" {
			value, err := strconv.Atoi(v)
			if err != nil {
				klog.Fatalf("Failed to convert string to integer: %v", err)
			}
			minSizeNodeGroup = value
		}
	}
	return minSizeNodeGroup
}

// Get max size group
func GetMaxSizeNodeGroup(kubeclient kube_client.Interface) int {
	var maxSizeNodeGroup int
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get(ctx.Background(), "autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling configmap")
		klog.Fatalf("Failed to get information of autoscaling configmap: %v", err)
	}
	for k, v := range configmaps.Data {
		if k == "max_node_group_size" {
			value, err := strconv.Atoi(v)
			if err != nil {
				klog.Fatalf("Failed to convert string to integer: %v", err)
			}
			maxSizeNodeGroup = value
		}
	}
	return maxSizeNodeGroup
}

// Get access token of FPTCloud
func GetAccessToken(kubeclient kube_client.Interface) string {
	var accessToken string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get(ctx.Background(), "fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "access_token" {
			accessToken = string(v)
		}
	}
	return accessToken
}

// Get vpc_id of customer
func GetVPCId(kubeclient kube_client.Interface) string {
	var vpcID string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get(ctx.Background(), "fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "vpc_id" {
			vpcID = string(v)
		}
	}
	return vpcID
}

// Get cluster_id info of K8S cluster
func GetClusterID(kubeclient kube_client.Interface) string {
	var clusterID string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get(ctx.Background(), "fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "cluster_id" {
			clusterID = string(v)
		}
	}
	return clusterID
}

type Cluster struct {
	Total int `json:"total"`
	Data  []struct {
		ID                  string `json:"id"`
		ClusterSlug         string `json:"cluster_slug"`
		ClusterID           string `json:"cluster_id"`
		VpcID               string `json:"vpc_id"`
		EdgeGatewayID       string `json:"edge_gateway_id"`
		NetworkID           string `json:"network_id"`
		CreatedAt           string `json:"created_at"`
		UpdatedAt           string `json:"updated_at"`
		Name                string `json:"name"`
		Status              string `json:"status"`
		WorkerNodeCount     string `json:"worker_node_count"`
		MasterNodeCount     string `json:"master_node_count"`
		KubernetesVersion   string `json:"kubernetes_version"`
		IsDeleted           string `json:"is_deleted"`
		AwxJobID            string `json:"awx_job_id"`
		AwxParams           string `json:"awx_params"`
		NfsDiskSize         string `json:"nfs_disk_size"`
		NfsStatus           string `json:"nfs_status"`
		IsRunning           string `json:"is_running"`
		ErrorMessage        string `json:"error_message"`
		Templates           string `json:"templates"`
		LoadBalancerSize    string `json:"load_balancer_size"`
		ProcessingMess      string `json:"processing_mess"`
		ClusterType         string `json:"cluster_type"`
		DistributedFirewall string `json:"distributed_firewall"`
	} `json:"data"`
}

// Get ID of cluster
func GetIDCluster(vpcID string, accessToken string, clusterID string) string {
	var id string
	var k8sCluster Cluster
	url := "https://console-api-pilot.fptcloud.com/api/v1/vmware/vpc/" + vpcID + "/kubernetes?page=1&page_size=25"
	token := accessToken
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data_body := []byte(body)
	error := json.Unmarshal(data_body, &k8sCluster)
	if error != nil {
		// if error is not nil
		// print error
		fmt.Println(error)
	}

	//fmt.Println(k8sCluster.Data[0])
	for _, cluster := range k8sCluster.Data {
		//fmt.Println(cluster)
		if cluster.ClusterID == clusterID {
			id = cluster.ID
		}
	}

	defer resp.Body.Close()
	fmt.Println("ID is: ", id)
	return id
}
