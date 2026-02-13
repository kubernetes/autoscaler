/*
Copyright 2019 The Kubernetes Authors.

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

package verdacloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda"
)

const (
	verdacloudProviderIDPrefix = "verdacloud://"
	defaultPodAmountsLimit     = 110
	refreshInterval            = 1 * time.Minute // throttles API calls during rapid loops
)

// VerdacloudManager manages Verdacloud resources for the cluster autoscaler.
type VerdacloudManager struct {
	cfg         *cloudConfig
	sdkProvider *verdacloudSDKProvider
	dcService   *verdacloudWrapper
	asgs        *autoScalingGroups
	lastRefresh time.Time
}

type asgTemplate struct {
	InstanceType *InstanceResource
	Tags         map[string]string
}

// InstanceResource represents the resource configuration of a Verdacloud instance type.
type InstanceResource struct {
	InstanceType string
	Arch         string
	CPU          int64
	Memory       int64
	GPU          int64
}

func createVerdacloudManager(cloudReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*VerdacloudManager, error) {
	cfg := &cloudConfig{}
	if cloudReader != nil {
		decoder := json.NewDecoder(cloudReader)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	if !cfg.isValid() {
		return nil, errors.New("invalid cloud configuration: please verify that image (GPU/CPU), sshKeyIDs, startupScript, and availableLocations are correctly specified in the cloud-config file")
	}

	cfg = verifyCloudConfigAndPatch(cfg)

	// create the sdk provider
	sdkProvider, err := createVerdacloudSDKProvider(cfg)
	if err != nil {
		return nil, err
	}

	// create the verdacloud wrapper using the official SDK client
	dcService := newVerdacloudWrapper(sdkProvider.client)

	manager := &VerdacloudManager{
		cfg:         cfg,
		sdkProvider: sdkProvider,
		dcService:   dcService,
		asgs:        nil,
	}

	manager.asgs, err = newAutoScalingGroups(dcService, discoveryOpts.NodeGroupSpecs, cfg)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// Refresh refreshes the state of the manager from the cloud provider.
func (m *VerdacloudManager) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *VerdacloudManager) forceRefresh() error {
	if err := m.asgs.regenerate(); err != nil {
		return err
	}
	m.lastRefresh = time.Now()
	return nil
}

func (m *VerdacloudManager) allASGActiveInstances(ctx context.Context, asg *Asg) ([]verda.Instance, error) {
	if asg.Name == "" {
		return nil, errors.New("asgName is required")
	}

	instances, err := m.dcService.GetActiveInstancesForAsg(ctx, asg.Name, asg.hostnamePrefix)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// GetAsgSize returns the current size of the ASG.
func (m *VerdacloudManager) GetAsgSize(asg *Asg) (int64, error) {
	if asg == nil {
		return 0, nil
	}

	return int64(asg.curSize), nil
}

// ScaleUpAsg scales up the ASG by the given delta.
func (m *VerdacloudManager) ScaleUpAsg(asg *Asg, delta int) error {
	return m.asgs.scaleUpAsg(asg, delta)
}

// ScaleDownAsg scales down the ASG by the given delta.
func (m *VerdacloudManager) ScaleDownAsg(asg *Asg, delta int) error {
	return m.asgs.scaleDownAsg(asg, delta)
}

func (m *VerdacloudManager) getAsgs() []*Asg {
	return m.asgs.getAsgs()
}

// GetAsgByRef returns the ASG for the given reference.
func (m *VerdacloudManager) GetAsgByRef(ref AsgRef) (*Asg, error) {
	return m.asgs.GetAsgByRef(ref)
}

// GetAsgNodes returns the provider IDs of all nodes in the ASG.
func (m *VerdacloudManager) GetAsgNodes(ctx context.Context, asg *Asg) ([]string, error) {
	instances, err := m.allASGActiveInstances(ctx, asg)
	if err != nil {
		return nil, err
	}
	providerIDs := make([]string, 0, len(instances))
	for _, inst := range instances {
		providerIDs = append(providerIDs, verdacloudProviderIDPrefix+inst.Location+"/"+inst.Hostname)
	}
	return providerIDs, nil
}

// GetAsgForInstance returns the ASG that the given instance belongs to.
func (m *VerdacloudManager) GetAsgForInstance(ref *InstanceRef) (*Asg, error) {
	if ref == nil {
		return nil, errors.New("ref is required")
	}

	return m.asgs.FindASGForInstance(ref)
}

// DeleteInstances deletes the given instances.
func (m *VerdacloudManager) DeleteInstances(instanceRefs []InstanceRef) error {
	if len(instanceRefs) == 0 {
		return nil
	}

	for _, ref := range instanceRefs {
		err := m.asgs.DeleteInstance(ref)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAsg deletes the given ASG.
func (m *VerdacloudManager) DeleteAsg(asg *Asg) error {
	return m.asgs.DeleteAsg(asg.AsgRef)
}

func verifyCloudConfigAndPatch(cfg *cloudConfig) *cloudConfig {

	if cfg.BillingConfig.Contract == "" {
		cfg.BillingConfig.Contract = verda.ContractTypePayAsYouGo
	}
	if cfg.BillingConfig.Price == "" {
		cfg.BillingConfig.Price = verda.PricingTypeFIXED
	}
	return cfg
}

// GetAvailableMachineTypes returns a list of available machine types.
func (m *VerdacloudManager) GetAvailableMachineTypes() ([]string, error) {
	ctx := context.Background()
	instanceTypes, err := m.dcService.ListInstanceTypes(ctx)
	if err != nil {
		return nil, err
	}

	return instanceTypes, nil
}

// GetAvailableGPUTypes returns a map of available GPU types.
func (m *VerdacloudManager) GetAvailableGPUTypes() map[string]struct{} {
	ctx := context.Background()
	instanceTypes, err := m.dcService.ListInstanceTypes(ctx)
	if err != nil {
		return nil
	}

	types := make(map[string]struct{}, len(instanceTypes))
	for _, instanceType := range instanceTypes {
		if strings.HasPrefix(strings.ToUpper(instanceType), "CPU.") {
			continue
		}
		types[instanceType] = struct{}{}
	}

	return types
}

func (m *VerdacloudManager) getInstancesForAsg(ref AsgRef) ([]cloudprovider.Instance, error) {
	asgInstances, err := m.asgs.InstancesForAsg(ref)
	if err != nil {
		return nil, err
	}
	cloudInstances := make([]cloudprovider.Instance, 0, len(asgInstances))
	for _, asgIns := range asgInstances {
		providerID := verdacloudProviderIDPrefix + asgIns.Location + "/" + asgIns.Hostname
		status := strings.ToLower(asgIns.Status)
		switch status {
		case verda.StatusRunning:
			cloudInstances = append(cloudInstances, cloudprovider.Instance{
				Id: providerID,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			})
		case verda.StatusNew, verda.StatusOrdered, verda.StatusProvisioning, verda.StatusValidating:
			cloudInstances = append(cloudInstances, cloudprovider.Instance{
				Id: providerID,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			})
		case verda.StatusOffline, verda.StatusDiscontinued, verda.StatusNotFound, verda.StatusUnknown, verda.StatusDeleting:
			cloudInstances = append(cloudInstances, cloudprovider.Instance{Id: providerID, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}})
		case verda.StatusError, verda.StatusNoCapacity:
			cloudInstances = append(cloudInstances, cloudprovider.Instance{Id: providerID,
				Status: &cloudprovider.InstanceStatus{ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    "verdacloud-instance-deployment-error",
					ErrorMessage: status,
				}}})
		default:
			cloudInstances = append(cloudInstances, cloudprovider.Instance{Id: providerID})
		}
	}
	return cloudInstances, nil
}

func (m *VerdacloudManager) buildNodeFromTemplate(asg *Asg, template *asgTemplate) (*apiv1.Node, error) {
	nodeName := fmt.Sprintf("asg-%s-%d", asg.Name, rand.Int63())

	labels := map[string]string{
		"kubernetes.io/arch":               template.InstanceType.Arch,
		"kubernetes.io/os":                 "linux",
		"node.kubernetes.io/instance-type": template.InstanceType.InstanceType,
		"topology.kubernetes.io/location":  strings.Join(asg.AvailabilityLocations, ","),
		"verda.com/hostname":               asg.Name,
		NodeGroupLabelKey:                  asg.Name,
	}
	if isGPUInstanceType(asg.instanceType) {
		labels[AcceleratorLabel] = asg.instanceType
	}

	capacity := apiv1.ResourceList{
		apiv1.ResourcePods:   *resource.NewQuantity(defaultPodAmountsLimit, resource.DecimalSI),
		apiv1.ResourceCPU:    *resource.NewQuantity(template.InstanceType.CPU, resource.DecimalSI),
		apiv1.ResourceMemory: *resource.NewQuantity(template.InstanceType.Memory, resource.BinarySI),
	}
	if template.InstanceType.GPU > 0 {
		capacity[apiv1.ResourceName(ResourceNvidiaGPU)] = *resource.NewQuantity(template.InstanceType.GPU, resource.DecimalSI)
	}

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName, Labels: labels},
		Status: apiv1.NodeStatus{
			Capacity:    capacity,
			Allocatable: capacity,
			Conditions:  cloudprovider.BuildReadyConditions(),
		},
	}

	nodeCfg := m.cfg.GetNodeConfig(asg.Name)
	for _, l := range nodeCfg.Labels {
		if parts := strings.SplitN(l, "=", 2); len(parts) == 2 {
			node.Labels[parts[0]] = parts[1]
		}
	}
	for k, v := range template.Tags {
		node.Labels[k] = v
	}
	node.Spec.Taints = append([]apiv1.Taint(nil), nodeCfg.Taints...)

	return node, nil
}

func (m *VerdacloudManager) getAsgTemplate(ctx context.Context, asgRef AsgRef) (*asgTemplate, error) {
	asg, err := m.asgs.GetAsgByRef(asgRef)
	if err != nil {
		return nil, err
	}

	instanceDetails, err := m.dcService.GetInstanceTypeDetails(ctx, asg.instanceType)
	if err != nil {
		return nil, fmt.Errorf("get instance type details for %s failed: %w", asg.instanceType, err)
	}

	// Normalize to API's canonical case
	if instanceDetails.InstanceType != asg.instanceType {
		asg.instanceType = instanceDetails.InstanceType
	}

	return &asgTemplate{
		InstanceType: instanceDetails,
		Tags:         make(map[string]string),
	}, nil
}
