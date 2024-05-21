/*
Copyright 2021 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/client"
	as "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/as/v20180419"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
)

const (
	refreshInterval      = 1 * time.Minute
	scaleToZeroSupported = true
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
	Region string `json:"region"`
}

// LabelAutoScalingGroupID represents the label of AutoScalingGroup
const LabelAutoScalingGroupID = "cloud.tencent.com/auto-scaling-group-id"

var cloudConfig CloudConfig

func readCloudConfig() error {
	cloudConfig.Region = os.Getenv("REGION")
	if cloudConfig.Region == "" {
		return errors.New("invalid REGION")
	}

	klog.V(4).Infof("tencentcloud config %+v", cloudConfig)

	return nil
}

// CreateTencentcloudManager constructs tencentcloudManager object.
func CreateTencentcloudManager(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, regional bool) (TencentcloudManager, error) {
	err := readCloudConfig()
	if err != nil {
		return nil, err
	}

	credential := common.NewCredential(
		os.Getenv("SECRET_ID"),
		os.Getenv("SECRET_KEY"),
	)
	cvmClient := client.NewClient(credential, cloudConfig.Region, newCVMClientProfile())
	vpcClient := client.NewClient(credential, cloudConfig.Region, newVPCClientProfile())
	asClient := client.NewClient(credential, cloudConfig.Region, newASClientProfile())

	service := NewCloudService(cvmClient, vpcClient, asClient)

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

	return m.cache.cloudService.DeleteInstances(commonAsg, toDeleteInstances)
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

	Tags []*as.Tag
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

	instanceInfo, err := m.cloudService.GetInstanceInfoByType(m.cache.GetInstanceType(asgRef))
	if err != nil {
		klog.Warningf("Failed to query instance type `%s` for %s: %v", m.cache.GetInstanceType(asgRef), asgRef.ID, err)
		return nil, cloudprovider.ErrNotImplemented
	}

	labels := make(map[string]string)
	labels[LabelAutoScalingGroupID] = asgRef.ID

	asg, err := m.cloudService.GetAutoScalingGroup(asgRef)
	if err != nil {
		return nil, err
	}
	if len(asg.SubnetIdSet) < 1 || asg.SubnetIdSet[0] == nil {
		return nil, fmt.Errorf("failed to get asg zone")
	}
	zone, err := m.cloudService.GetZoneBySubnetID(*asg.SubnetIdSet[0])
	if err != nil {
		return nil, err
	}
	zoneInfo, err := m.cloudService.GetZoneInfo(zone)
	if err != nil {
		return nil, err
	}

	instanceTemplate = &InstanceTemplate{
		InstanceType: instanceInfo.InstanceType,
		Region:       cloudConfig.Region,
		Zone:         *zoneInfo.ZoneId,
		Cpu:          instanceInfo.CPU,
		Mem:          instanceInfo.Memory,
		Gpu:          instanceInfo.GPU,
		Tags:         asg.Tags,
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

	resourcesFromTags := extractAllocatableResourcesFromAsg(template.Tags)
	klog.V(5).Infof("Extracted resources from ASG tags %v", resourcesFromTags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))

	node.Spec.Taints = extractTaintsFromAsg(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	return &node, nil
}

func buildGenericLabels(template *InstanceTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceType] = template.InstanceType
	result[apiv1.LabelInstanceTypeStable] = template.InstanceType

	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelZoneRegionStable] = template.Region
	result[apiv1.LabelZoneFailureDomain] = template.Zone
	result[apiv1.LabelZoneFailureDomainStable] = template.Zone
	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractLabelsFromAsg(tags []*as.Tag) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/label/")
		// Extract TKE labels from ASG
		if len(splits) <= 1 {
			splits = strings.Split(k, "tencentcloud:")
		}
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				result[label] = v
			}
		}
	}

	return result
}

func extractAllocatableResourcesFromAsg(tags []*as.Tag) map[string]*resource.Quantity {
	result := make(map[string]*resource.Quantity)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/resources/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				quantity, err := resource.ParseQuantity(v)
				if err != nil {
					continue
				}
				result[label] = &quantity
			}
		}
	}

	return result
}

func extractAllocatableResourcesFromTags(tags map[string]string) map[string]*resource.Quantity {
	result := make(map[string]*resource.Quantity)

	for k, v := range tags {
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/resources/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				quantity, err := resource.ParseQuantity(v)
				if err != nil {
					klog.Warningf("Failed to parse resource quanitity '%s' for resource '%s'", v, label)
					continue
				}
				result[label] = &quantity
			}
		}
	}

	return result
}

func extractTaintsFromAsg(tags []*as.Tag) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		// The tag value must be in the format <tag>:NoSchedule
		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")
		if r.MatchString(v) {
			splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
			if len(splits) > 1 {
				values := strings.SplitN(v, ":", 2)
				if len(values) > 1 {
					taints = append(taints, apiv1.Taint{
						Key:    splits[1],
						Value:  values[0],
						Effect: apiv1.TaintEffect(values[1]),
					})
				}
			}
		}
	}
	return taints
}
