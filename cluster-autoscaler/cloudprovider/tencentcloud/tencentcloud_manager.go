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

package tencentcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	as "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	retryCountStop   = 5
	intervalTimeStop = 5 * time.Second
	tokenExpiredTime = 7200

	serviceName          = "cluster-autoscaler"
	refreshInterval      = 1 * time.Minute
	scaleToZeroSupported = true
)

// network extended resources
const (
	TKERouteENIIP = "tke.cloud.tencent.com/eni-ip"
	TKEDirectENI  = "tke.cloud.tencent.com/direct-eni"
)

// vcuda resources
const (
	VCudaCore   = "tencent.com/vcuda-core"
	VCudaMemory = "tencent.com/vcuda-memory"
)

// GPUMemoryMap is a coefficient to get gpu extended resources
var (
	GPUMemoryMap = map[string]int64{
		"GN10X": 32,
		"GN10S": 16,
		"GN10":  16,
		"GN8":   24,
		"GN7":   16,
		"GN6":   8,
		"GN6S":  8,
		"GN2":   24,
		"GN7vw": 16,
		"GC1":   11,
	}
)

// TencentcloudManager is handles tencentcloud communication and data caching.
type TencentcloudManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh() error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	RegisterAsg(asg Asg)
	// GetAsgs returns list of registered Asgs.
	GetAsgs() []Asg
	// GetAsgNodes returns Asg nodes.
	GetAsgNodes(Asg Asg) ([]cloudprovider.Instance, error)
	// GetAsgForInstance returns Asg to which the given instance belongs.
	GetAsgForInstance(instance TcRef) (Asg, error)
	// GetAsgTemplateNode returns a template node for Asg.
	GetAsgTemplateNode(Asg Asg) (*apiv1.Node, error)
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// GetAsgSize gets Asg size.
	GetAsgSize(Asg Asg) (int64, error)

	// SetAsgSize sets Asg size.
	SetAsgSize(Asg Asg, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same Asg.
	DeleteInstances(instances []TcRef) error
}

type tencentcloudManagerImpl struct {
	mutex                sync.Mutex
	lastRefresh          time.Time
	cloudService         CloudService
	cache                *TencentcloudCache
	regional             bool
	explicitlyConfigured map[TcRef]bool
	interrupt            chan struct{}
}

// CloudConfig represent tencentcloud configuration
type CloudConfig struct {
	Region     string `json:"region"`
	RegionName string `json:"regionName"`
	Zone       string `json:"zone"`
	DryRun     bool   `json:dryRun`
	SecretID   string
	SecretKey  string
	ClusterID  string
	IsTest     bool
}

// LabelAutoScalingGroupID represents the label of AutoScalingGroup
const LabelAutoScalingGroupID = "cloud.tencent.com/auto-scaling-group-id"

var cloudConfig CloudConfig

func readCloudConfig(configReader io.Reader) error {
	if configReader == nil {
		return fmt.Errorf("tencentcloud cloud config is not exists")
	}

	if err := json.NewDecoder(configReader).Decode(&cloudConfig); err != nil {
		return err
	}

	testEnv := os.Getenv("TEST_ENV")
	if testEnv == "true" {
		cloudConfig.IsTest = true
	}

	dryRun := os.Getenv("DRY_RUN")
	if dryRun == "true" {
		cloudConfig.DryRun = true
	}

	secretID := os.Getenv("SECRET_ID")
	secretKey := os.Getenv("SECRET_KEY")
	region := os.Getenv("REGION")
	regionName := os.Getenv("REGION_NAME")
	clusterID := os.Getenv("CLUSTER_ID")
	if secretID == "" {
		return fmt.Errorf("please specify the environment variable: SECRET_ID")
	}
	if secretKey == "" {
		return fmt.Errorf("please specify the environment variable: SECRET_KEY")
	}
	if region == "" {
		return fmt.Errorf("please specify the environment variable: REGION")
	}
	if regionName == "" {
		return fmt.Errorf("please specify the environment variable: REGION_NAME")
	}
	if clusterID == "" {
		return fmt.Errorf("please specify the environment variable: CLUSTER_ID")
	}
	cloudConfig.SecretID = secretID
	cloudConfig.SecretKey = secretKey
	cloudConfig.Region = region
	cloudConfig.RegionName = regionName
	cloudConfig.ClusterID = clusterID

	klog.V(4).Infof("tencentcloud config %+v", cloudConfig)

	return nil
}

// CreateTencentcloudManager constructs tencentcloudManager object.
func CreateTencentcloudManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, regional bool) (TencentcloudManager, error) {
	err := readCloudConfig(configReader)
	if err != nil {
		return nil, err
	}

	credential := common.NewCredential(cloudConfig.SecretID, cloudConfig.SecretKey)
	cvmClient, err := cvm.NewClient(credential, cloudConfig.RegionName, newCVMClientProfile())
	if err != nil {
		return nil, err
	}
	vpcClient, err := vpc.NewClient(credential, cloudConfig.RegionName, newVPCClientProfile())
	if err != nil {
		return nil, err
	}
	asClient, err := as.NewClient(credential, cloudConfig.RegionName, newASClientProfile())
	if err != nil {
		return nil, err
	}
	tkeClient, err := tke.NewClient(credential, cloudConfig.RegionName, newTKEClientProfile())
	if err != nil {
		return nil, err
	}

	var service CloudService
	if cloudConfig.DryRun {
		service = NewCloudMockService(cvmClient, vpcClient, asClient, tkeClient)
	} else {
		service = NewCloudService(cvmClient, vpcClient, asClient, tkeClient)
	}

	manager := &tencentcloudManagerImpl{
		cache:                NewTencentcloudCache(service),
		cloudService:         service,
		regional:             regional,
		interrupt:            make(chan struct{}),
		explicitlyConfigured: make(map[TcRef]bool),
	}

	if err := manager.fetchExplicitAsgs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, fmt.Errorf("failed to fetch ASGs: %v", err)
	}

	go wait.Until(func() {
		if err := manager.cache.RegenerateInstancesCache(); err != nil {
			klog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// Fetch explicitly configured ASGs. These ASGs should never be unregistered
// during refreshes, even if they no longer exist in Tencentcloud.
func (m *tencentcloudManagerImpl) fetchExplicitAsgs(specs []string) error {
	changed := false
	for _, spec := range specs {
		asg, err := m.buildAsgFromFlag(spec)
		if err != nil {
			return err
		}
		if m.registerAsg(asg) {
			changed = true
		}
		m.explicitlyConfigured[asg.TencentcloudRef()] = true
	}

	if changed {
		err := m.cache.RegenerateAutoScalingGroupCache()
		if err != nil {
			return err
		}
		err = m.cache.RegenerateInstancesCache()
		if err != nil {
			return err
		}
		for _, asg := range m.cache.GetAsgs() {
			// Try to build a node from template to validate that this group
			// can be scaled up from 0 nodes.
			// We may never need to do it, so just log error if it fails.
			if _, err := m.GetAsgTemplateNode(asg); err != nil {
				klog.Errorf("Can't build node from template for %s, won't be able to scale from 0: %v", asg.TencentcloudRef().String(), err)
			}
		}
	}

	return nil
}

// GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
func (m *tencentcloudManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

// registerAsg registers asg in TencentcloudManager. Returns true if the node group didn't exist before or its config has changed.
func (m *tencentcloudManagerImpl) registerAsg(asg Asg) bool {
	return m.cache.RegisterAsg(asg)
}

func (m *tencentcloudManagerImpl) buildAsgFromFlag(flag string) (Asg, error) {
	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildAsgFromSpec(s)
}

func (m *tencentcloudManagerImpl) buildAsgFromSpec(s *dynamic.NodeGroupSpec) (Asg, error) {
	return &tcAsg{
		tencentcloudManager: m,
		minSize:             s.MinSize,
		maxSize:             s.MaxSize,
		tencentcloudRef: TcRef{
			ID: s.Name,
		},
	}, nil
}

// Refresh triggers refresh of cached resources.
func (m *tencentcloudManagerImpl) Refresh() error {
	m.cache.InvalidateAllAsgTargetSizes()
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *tencentcloudManagerImpl) forceRefresh() error {
	// TODO refresh

	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed Tencentcloud resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// GetAsgs returns list of registered ASGs.
func (m *tencentcloudManagerImpl) GetAsgs() []Asg {
	return m.cache.GetAsgs()
}

// RegisterAsg registers asg in Tencentcloud Manager.
func (m *tencentcloudManagerImpl) RegisterAsg(asg Asg) {
	m.cache.RegisterAsg(asg)
}

// GetAsgForInstance returns asg of the given Instance
func (m *tencentcloudManagerImpl) GetAsgForInstance(instance TcRef) (Asg, error) {
	return m.cache.FindForInstance(instance)
}

// GetAsgSize gets ASG size.
func (m *tencentcloudManagerImpl) GetAsgSize(asg Asg) (int64, error) {

	targetSize, found := m.cache.GetAsgTargetSize(asg.TencentcloudRef())
	if found {
		return targetSize, nil
	}

	group, err := m.cloudService.GetAutoScalingGroup(asg.TencentcloudRef())
	if err != nil {
		return -1, err
	}
	if group.DesiredCapacity == nil {
		return -1, fmt.Errorf("%s invalid desired capacity", asg.Id())
	}

	m.cache.SetAsgTargetSize(asg.TencentcloudRef(), *group.DesiredCapacity)
	return *group.DesiredCapacity, nil
}

// SetAsgSize sets ASG size.
func (m *tencentcloudManagerImpl) SetAsgSize(asg Asg, size int64) error {
	klog.V(0).Infof("Setting asg %s size to %d", asg.Id(), size)
	err := m.cloudService.ResizeAsg(asg.TencentcloudRef(), uint64(size))
	if err != nil {
		return err
	}
	m.cache.SetAsgTargetSize(asg.TencentcloudRef(), size)
	return nil
}

// Cleanup ...
func (m *tencentcloudManagerImpl) Cleanup() error {
	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *tencentcloudManagerImpl) DeleteInstances(instances []TcRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonAsg, err := m.cache.FindForInstance(instances[0])
	if err != nil {
		return err
	}
	toDeleteInstances := make([]string, 0)
	for _, instance := range instances {
		asg, err := m.cache.FindForInstance(instance)
		if err != nil {
			return err
		}
		if asg != commonAsg {
			return fmt.Errorf("can not delete instances which don't belong to the same ASG")
		}
		toDeleteInstances = append(toDeleteInstances, instance.ID)
	}

	m.cache.InvalidateAsgTargetSize(commonAsg.TencentcloudRef())
	m.cache.cloudService.DeleteInstances(commonAsg, toDeleteInstances)

	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *tencentcloudManagerImpl) GetAsgNodes(asg Asg) ([]cloudprovider.Instance, error) {
	result := make([]cloudprovider.Instance, 0)
	instances, err := m.cloudService.GetAutoScalingInstances(asg.TencentcloudRef())
	if err != nil {
		return result, err
	}
	for _, instance := range instances {
		if instance == nil || instance.LifeCycleState == nil || instance.InstanceId == nil {
			continue
		}
		if *instance.LifeCycleState == "Removing" {
			continue
		}
		instanceRef, err := m.cloudService.GetTencentcloudInstanceRef(instance)
		if err != nil {
			klog.Warning("failed to get tencentcloud instance ref:", err)
		} else if instanceRef.Zone == "" {
			klog.V(4).Infof("Skipping %s, that is scheduling by AS", *instance.InstanceId)
		} else {
			result = append(result, cloudprovider.Instance{Id: instanceRef.ToProviderID()})
		}
	}
	return result, nil
}

// InstanceTemplate represents CVM template
type InstanceTemplate struct {
	InstanceType string
	Region       string
	Zone         string
	Cpu          int64
	Mem          int64
	Gpu          int64
	// gpu虚拟化资源
	VCudaCore int64
	VCudaMem  int64
	// vpc-cni的集群eni-ip资源
	TKERouteENIIP int64
	TKEDirectENI  int64

	Label  map[string]string
	Taints []*tke.Taint
}

// NetworkExtendedResources represents network extended resources
type NetworkExtendedResources struct {
	TKERouteENIIP int64
	TKEDirectENI  int64
}

var networkExtendedResourcesMap = make(map[string]*NetworkExtendedResources)

// GetAsgInstanceTemplate returns instance template for Asg with given ref
func (m *tencentcloudManagerImpl) GetAsgInstanceTemplate(asgRef TcRef) (*InstanceTemplate, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instanceTemplate, found := m.cache.GetAsgInstanceTemplate(asgRef)
	if found {
		return instanceTemplate, nil
	}

	getNetworkExtendedResources := func(instanceType string) (*NetworkExtendedResources, error) {
		if resources, exist := networkExtendedResourcesMap[instanceType]; exist {
			return resources, nil
		}

		pli, err := m.cloudService.DescribeVpcCniPodLimits(instanceType)
		if err != nil {
			return nil, err
		}
		resources := &NetworkExtendedResources{}
		if pli != nil {
			if pli.PodLimits == nil ||
				pli.PodLimits.TKERouteENINonStaticIP == nil ||
				pli.PodLimits.TKEDirectENI == nil {
				return nil, fmt.Errorf("get wrong eni limits(nil)")
			}
			resources.TKEDirectENI = *pli.PodLimits.TKEDirectENI
			resources.TKERouteENIIP = *pli.PodLimits.TKERouteENINonStaticIP
		}

		klog.Infof("%v", resources)
		networkExtendedResourcesMap[instanceType] = resources

		return resources, nil
	}
	instanceInfo, err := m.cloudService.GetInstanceInfoByType(m.cache.GetInstanceType(asgRef))
	if err != nil {
		return nil, err
	}
	npInfo, err := m.cloudService.GetNodePoolInfo(cloudConfig.ClusterID, asgRef.ID)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	for _, label := range npInfo.Labels {
		if label.Name != nil && label.Value != nil {
			labels[*label.Name] = *label.Value
		}
	}
	labels[LabelAutoScalingGroupID] = asgRef.ID

	asg, err := m.cloudService.GetAutoScalingGroup(asgRef)
	if err != nil {
		return nil, err
	}
	if len(asg.SubnetIdSet) < 1 || asg.SubnetIdSet[0] == nil {
		return nil, fmt.Errorf("Failed to get asg zone")
	}
	zone, err := m.cloudService.GetZoneBySubnetID(*asg.SubnetIdSet[0])
	if err != nil {
		return nil, err
	}
	zoneInfo, err := m.cloudService.GetZoneInfo(zone)
	if err != nil {
		return nil, err
	}

	// eni
	networkExtendedResources, err := getNetworkExtendedResources(instanceInfo.InstanceType)
	if err != nil {
		return nil, err
	}

	gpuMult, ok := GPUMemoryMap[instanceInfo.InstanceFamily]
	if !ok {
		gpuMult = 24
	}
	vcudaMem := instanceInfo.GPU * gpuMult * 4

	instanceTemplate = &InstanceTemplate{
		InstanceType:  instanceInfo.InstanceType,
		Region:        cloudConfig.Region,
		Zone:          *zoneInfo.ZoneId,
		Cpu:           instanceInfo.CPU,
		Mem:           instanceInfo.Memory,
		Gpu:           instanceInfo.GPU,
		VCudaCore:     instanceInfo.GPU * 100,
		VCudaMem:      vcudaMem,
		TKEDirectENI:  networkExtendedResources.TKEDirectENI,
		TKERouteENIIP: networkExtendedResources.TKERouteENIIP,
		Label:         labels,
		Taints:        npInfo.Taints,
	}

	m.cache.SetAsgInstanceTemplate(asgRef, instanceTemplate)
	return instanceTemplate, nil
}

func (m *tencentcloudManagerImpl) GetAsgTemplateNode(asg Asg) (*apiv1.Node, error) {

	template, err := m.GetAsgInstanceTemplate(asg.TencentcloudRef())
	if err != nil {
		return nil, err
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", asg.TencentcloudRef().ID, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.Cpu, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.Mem*1024*1024*1024, resource.DecimalSI)
	if template.TKERouteENIIP > 0 {
		node.Status.Capacity[TKERouteENIIP] = *resource.NewQuantity(template.TKERouteENIIP, resource.DecimalSI)
	}
	if template.TKEDirectENI > 0 {
		node.Status.Capacity[TKEDirectENI] = *resource.NewQuantity(template.TKEDirectENI, resource.DecimalSI)
	}
	if template.Gpu > 0 {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.Gpu, resource.DecimalSI)
		node.Status.Capacity[VCudaCore] = *resource.NewQuantity(template.VCudaCore, resource.DecimalSI)
		node.Status.Capacity[VCudaMemory] = *resource.NewQuantity(template.VCudaMem, resource.DecimalSI)

		klog.Infof("Capacity resource set gpu %s(%d)", gpu.ResourceNvidiaGPU, template.Gpu)
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, template.Label)

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromAsg(template.Taints)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	return &node, nil
}

func buildGenericLabels(template *InstanceTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceType] = template.InstanceType

	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelZoneFailureDomain] = template.Zone
	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractTaintsFromAsg(npTaints []*tke.Taint) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, npTaint := range npTaints {
		if npTaint != nil && npTaint.Key != nil && npTaint.Value != nil && npTaint.Effect != nil {
			taints = append(taints, apiv1.Taint{
				Key:    *npTaint.Key,
				Value:  *npTaint.Value,
				Effect: apiv1.TaintEffect(*npTaint.Effect),
			})
		}
	}
	return taints
}
