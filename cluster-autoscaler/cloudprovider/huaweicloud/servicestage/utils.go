package huaweicloud

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	modelecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

var (
	clusterUpdateLock = sync.Mutex{}
	deleteMux         = sync.Mutex{}
)

const (
	nodePoolIDKey = "kubernetes.io/node-pool.id"
	// DefaultPoolName defines default pool name
	DefaultPoolName = "DefaultPool"

	// ToBeDeletedByCCEAnnotation is a annotation used to mark the node to be deleted by cce.
	ToBeDeletedByCCEAnnotation = "node.kubernetes.io/to-be-deleted-by-cce"
)

// IsValidUUID checks UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	if err == nil {
		return true
	}
	return false
}

func getNodePoolID(node model.Node) string {
	/**
		Cluster Manager will add node-pool.id into annotation
		1. 节点通过node pool controller创建，则返回真实的node pool id
		2. 节点通过v2,v3的接口创建，并且是包周期节点，或者是静态纳管接入节点，则kubernetes.io/node-pool.id值为DefaultPool
		3. 剩余通过v2,v3接口创建的节点，兼容老的autoscaler的规则， kubernetes.io/node-pool.id为AZ#Flavor#OS
		最后如果老的autoscaler都下线时，第3类节点应都放入DefaultPool中,当前行为将非uuid的值都放入到default pool中。
	**/
	id := node.Metadata.Annotations[nodePoolIDKey]
	if !IsValidUUID(id) {
		return DefaultPoolName
	}

	return id
}

func getNodeGroups(nodePool model.NodePool, manager *huaweicloudCloudManager, spec *dynamic.NodeGroupSpec) *NodeGroup {
	maxNodeCount := int(*nodePool.Spec.Autoscaling.MaxNodeCount)
	minNodeCount := int(*nodePool.Spec.Autoscaling.MinNodeCount)

	min := func(x, y int) int {
		if x < y {
			return x
		}
		return y
	}
	max := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}
	if spec != nil {
		maxNodeCount = min(spec.MaxSize, int(*nodePool.Spec.Autoscaling.MaxNodeCount))
		minNodeCount = min(maxNodeCount, max(spec.MinSize, int(*nodePool.Spec.Autoscaling.MinNodeCount)))
	}

	currentSize := int(*nodePool.Status.CurrentNode)
	return &NodeGroup{
		huaweiCloudManager: manager,
		deleteMutex:        &deleteMux,
		clusterUpdateMutex: &clusterUpdateLock,
		nodePoolName:       nodePool.Metadata.Name,
		nodePoolId:         *nodePool.Metadata.Uid,
		clusterName:        manager.clusterName,
		autoscalingEnabled: *nodePool.Spec.Autoscaling.Enable,
		minNodeCount:       minNodeCount,
		maxNodeCount:       maxNodeCount,
		targetSize:         &currentSize,
		conditions:         *nodePool.Status.Conditions,
		nodePoolSpec:       nodePool.Spec,
	}
}

func getConfigNg(do cloudprovider.NodeGroupDiscoveryOptions) map[string]*dynamic.NodeGroupSpec {
	configNg := make(map[string]*dynamic.NodeGroupSpec)
	for _, specStr := range do.NodeGroupSpecs {
		spec, err := dynamic.SpecFromString(specStr, true)
		if err != nil {
			klog.Fatalf("failed to get specify node pools information of config file: %v\n", err)
		}
		configNg[spec.Name] = spec
	}
	return configNg
}

func (hcp *huaweicloudCloudProvider) fixNodePool(nodePool *NodeGroup, initialNodeCount int) {
	if hasScalableResource(nodePool.nodePoolId, nodePool.conditions) {
		return
	}
	if *nodePool.targetSize == initialNodeCount {
		return
	}
	klog.Infof("pool(%s) current node %d not matched with expect node count %d, start to fix",
		nodePool.nodePoolId, *nodePool.targetSize, initialNodeCount)

	// use current node count to fix node expect node count
	nodePool.maxNodeCount = *nodePool.targetSize
	err := hcp.huaweiCloudManager.updateNodeCount(nodePool, *nodePool.targetSize)
	if err != nil {
		klog.Errorf("pool(%s) fix node group size from %d to %d failed: %v", nodePool.nodePoolId, *nodePool.targetSize, err)
		return
	}

	klog.Infof("pool(%s) fix node group size from %d to %d", nodePool.nodePoolId, initialNodeCount, *nodePool.targetSize)
	return
}

func (hcp *huaweicloudCloudProvider) updateNodePoolForNodeUID(nodePoolForNodePoolUID map[string]*NodeGroup) error {
	listNodeReq := &model.ListNodesRequest{
		ClusterId: hcp.huaweiCloudManager.clusterName,
	}
	allNodes, err := hcp.huaweiCloudManager.clusterClient.ListNodes(listNodeReq)
	if err != nil {
		return err
	}

	hcp.nodePoolForNodeUID = make(map[string]*NodeGroup, len(*allNodes.Items))
	for _, node := range *allNodes.Items {
		nodePoolUID := getNodePoolID(node)
		if v, ok := nodePoolForNodePoolUID[nodePoolUID]; ok {
			hcp.nodePoolForNodeUID[*node.Metadata.Uid] = v
		}
	}
	return nil
}

func hasScalableResource(id string, conditions []model.NodePoolCondition) bool {
	for _, condition := range conditions {
		if *condition.Type == "Scalable" {
			klog.V(4).Infof("node pool(%s) condition status: %s, LastTransitTime: %s", id, *condition.Status, condition.LastTransitTime)
			// if condition change to no scalable resource large than 1 min, return false
			// because may be quota insufficient return quick ,if we fix right now, may be creating node is in progress
			// these node may not in status.currentNodes, may lead to delete some creating node
			t, err := time.Parse("2006-01-02T15:04:05Z", *condition.LastTransitTime)
			if err != nil {
				return false
			}
			if *condition.Status == "False" && time.Since(t) >= time.Minute {
				return false
			}
			return true
		}
	}
	return true
}

// HasToBeDeletedByCCEAnnotation check cce annotation has to be deleted
func HasToBeDeletedByCCEAnnotation(node *apiv1.Node) bool {
	if node == nil || node.Annotations == nil {
		return false
	}
	_, exist := node.Annotations[ToBeDeletedByCCEAnnotation]
	return exist
}

func (ng *NodeGroup) getTemplate() (*nodePoolTemplate, error) {
	showNodePoolReq := &model.ShowNodePoolRequest{
		ClusterId:  ng.clusterName,
		NodepoolId: ng.nodePoolId,
	}
	nodePoolInfo, err := ng.huaweiCloudManager.clusterClient.ShowNodePool(showNodePoolReq)
	if err != nil {
		return nil, err
	}

	zone := nodePoolInfo.Spec.NodeTemplate.Az
	if zone == "random" {
		zone = ""
	}
	listFlavorsReq := &modelecs.ListFlavorsRequest{
		AvailabilityZone: &zone,
	}
	resp, err := ng.huaweiCloudManager.ecsClient.ListFlavors(listFlavorsReq)
	if err != nil {
		return nil, err
	}

	for _, flavor := range *(*resp).Flavors {
		if !strings.EqualFold(flavor.Id, nodePoolInfo.Spec.NodeTemplate.Flavor) {
			continue
		}

		gpuType, gpuNum := getGpuInfo(flavor)
		vcpus, _ := strconv.ParseInt(flavor.Vcpus, 10, 64)
		taints := getTaints(nodePoolInfo.Spec.NodeTemplate.Taints)
		return &nodePoolTemplate{
			name:    flavor.Name,
			vcpu:    vcpus,
			ram:     int64(flavor.Ram),
			gpuType: gpuType,
			gpuNum:  gpuNum,
			taints:  taints,
			tags:    nodePoolInfo.Spec.NodeTemplate.K8sTags,
			zone:    nodePoolInfo.Spec.NodeTemplate.Az,
		}, nil
	}
	return nil, nil
}

func getGpuInfo(flavor modelecs.Flavor) (string, int64) {
	if *flavor.OsExtraSpecs.Ecsperformancetype != "gpu" {
		return "", 0
	}
	gpuInfo := strings.Split(*flavor.OsExtraSpecs.PciPassthroughalias, ":")
	if len(gpuInfo) != 2 {
		return "", 0
	}
	gpuNum, err := strconv.ParseInt(gpuInfo[1], 10, 64)
	if err != nil {
		return "", 0
	}
	return gpuInfo[0], gpuNum
}

func getTaints(taints *[]model.Taint) []apiv1.Taint {
	if taints == nil {
		return nil
	}
	result := make([]apiv1.Taint, 0, len(*taints))
	for _, taint := range *taints {
		effect, err := taint.Effect.MarshalJSON()
		if err != nil {
			continue
		}
		effectStr := strings.Trim(string(effect), "\"")
		tmp := apiv1.Taint{
			Key:    *taint.Key,
			Value:  *taint.Value,
			Effect: apiv1.TaintEffect(effectStr),
		}
		result = append(result, tmp)
	}
	return result
}
func (ng *NodeGroup) buildNodeFromTemplate(template *nodePoolTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", ng.nodePoolName, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Spec.Taints = template.taints
	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	var maxPods int64 = 110

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(maxPods, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.vcpu, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.gpuNum, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.ram*1024*1024, resource.DecimalSI)

	node.Status.Allocatable = caculateAllocatableResource(node.Status.Capacity, maxPods)

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *nodePoolTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceTypeStable] = template.name

	result[apiv1.LabelTopologyRegion] = template.region
	result[apiv1.LabelTopologyZone] = template.zone
	result[apiv1.LabelHostname] = nodeName

	// append custom node labels
	for key, value := range template.tags {
		result[key] = value
	}

	return result
}

type allocatableBracket struct {
	threshold            int64
	marginalReservedRate float64
}

const (
	mbPerGB           = 1024
	millicoresPerCore = 1000
)

const (
	// Kubelet "evictionHard: {memory.available}" is subtracted from
	// capacity when calculating allocatable (on top of kube-reserved).
	// We don't have a good place to get it from, but it has been hard-coded
	// to 100Mi since at least k8s 1.4.,here is cce v1.17 default size 504Mi.
	kubeletEvictionHardMemory = 504 * 1024 * 1024
)

func caculateAllocatableResource(capacity apiv1.ResourceList, maxPods int64) apiv1.ResourceList {
	memoryReserved := memoryReservedMB(capacity.Memory().Value()/(1024*1024), maxPods)
	cpuReserved := cpuReservedMillicores(capacity.Cpu().MilliValue())
	memoryReserved = memoryReserved * 1024 * 1024
	memoryReserved += kubeletEvictionHardMemory
	reserved := apiv1.ResourceList{}
	reserved[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpuReserved, resource.DecimalSI)
	reserved[apiv1.ResourceMemory] = *resource.NewQuantity(memoryReserved, resource.BinarySI)
	allocatable := getAllocatable(capacity, reserved)
	allocatable[apiv1.ResourcePods] = capacity[apiv1.ResourcePods]
	return allocatable
}

func getAllocatable(capacity, reserved apiv1.ResourceList) apiv1.ResourceList {
	allocatable := apiv1.ResourceList{}
	for key, value := range capacity {
		quantity := value.DeepCopy()
		if reservedQuantity, found := reserved[key]; found {
			quantity.Sub(reservedQuantity)
		}
		allocatable[key] = quantity
	}
	return allocatable
}

func cpuReservedMillicores(cpuCapacityMillicores int64) int64 {
	return calculateReserved(cpuCapacityMillicores, []allocatableBracket{
		{
			threshold:            0,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            1 * millicoresPerCore,
			marginalReservedRate: 0.01,
		},
		{
			threshold:            2 * millicoresPerCore,
			marginalReservedRate: 0.005,
		},
		{
			threshold:            4 * millicoresPerCore,
			marginalReservedRate: 0.0025,
		},
	})
}

func memoryReservedMB(memoryCapacityMB int64, maxPods int64) int64 {
	return memoryReservedByMemoryTotal(memoryCapacityMB) + memoryReservedByPods(maxPods)
}

func memoryReservedByPods(maxPods int64) int64 {
	if maxPods <= 16 {
		return 700
	}

	if 16 < maxPods && maxPods <= 32 {
		return int64(700 + float64(maxPods-16)*18.75)
	} else if 32 < maxPods && maxPods <= 64 {
		return int64(1024 + float64(maxPods-32)*6.25)
	} else if 64 < maxPods && maxPods <= 128 {
		return int64(1230 + float64(maxPods-64)*7.80)
	} else {
		return int64(1740 + float64(maxPods-128)*11.20)
	}
}

func memoryReservedByMemoryTotal(memoryCapacityMB int64) int64 {
	if memoryCapacityMB <= 8*mbPerGB {
		// do not set any memory reserved for nodes with less than 8 Gb of capacity
		return 0
	}
	return calculateReserved(memoryCapacityMB, []allocatableBracket{
		{
			threshold:            8 * mbPerGB,
			marginalReservedRate: 0.1,
		},
		{
			threshold:            16 * mbPerGB,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            128 * mbPerGB,
			marginalReservedRate: 0.02,
		},
	})
}

func calculateReserved(capacity int64, brackets []allocatableBracket) int64 {
	var reserved float64
	for i, bracket := range brackets {
		c := capacity
		if i < len(brackets)-1 && brackets[i+1].threshold < capacity {
			c = brackets[i+1].threshold
		}
		additionalReserved := float64(c-bracket.threshold) * bracket.marginalReservedRate
		if additionalReserved > 0 {
			reserved += additionalReserved
		}
	}
	return int64(reserved)
}
