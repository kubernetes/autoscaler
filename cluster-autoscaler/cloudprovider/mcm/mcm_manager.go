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

This file was copied and modified from the kubernetes/autoscaler project
https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-release-1.1/cluster-autoscaler/cloudprovider/aws/aws_manager.go

Modifications Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.
*/

package mcm

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	awsapis "github.com/gardener/machine-controller-manager-provider-aws/pkg/aws/apis"
	azureapis "github.com/gardener/machine-controller-manager-provider-azure/pkg/azure/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	machineapi "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/typed/machine/v1alpha1"
	machineinformers "github.com/gardener/machine-controller-manager/pkg/client/informers/externalversions"
	machinelisters "github.com/gardener/machine-controller-manager/pkg/client/listers/machine/v1alpha1"

	//corecontroller "github.com/gardener/machine-controller-manager/pkg/controller"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	aws "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	azure "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	maxRetryDeadline        = 1 * time.Minute
	conflictRetryInterval   = 5 * time.Second
	// machinePriorityAnnotation is the annotation to set machine priority while deletion
	machinePriorityAnnotation = "machinepriority.machine.sapcloud.io"
	// kindAWSMachineClass is the kind for machine class used by In-tree AWS provider
	kindAWSMachineClass = "AWSMachineClass"
	// kindAzureMachineClass is the kind for machine class used by In-tree Azure provider
	kindAzureMachineClass = "AzureMachineClass"
	// kindMachineClass is the kind for generic machine class used by the OOT providers
	kindMachineClass = "MachineClass"
	// providerAWS is the provider type for AWS machine class objects
	providerAWS = "AWS"
	// providerAzure is the provider type for Azure machine class object
	providerAzure = "Azure"
)

//McmManager manages the client communication for MachineDeployments.
type McmManager struct {
	namespace               string
	interrupt               chan struct{}
	discoveryOpts           cloudprovider.NodeGroupDiscoveryOptions
	machineclient           machineapi.MachineV1alpha1Interface
	coreclient              kubernetes.Interface
	machineDeploymentLister machinelisters.MachineDeploymentLister
	machinelisters          machinelisters.MachineLister
}

type instanceType struct {
	InstanceType string
	VCPU         int64
	MemoryMb     int64
	GPU          int64
}

type nodeTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Labels       map[string]string
	Taints       []apiv1.Taint
}

func createMCMManagerInternal(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*McmManager, error) {

	// controlKubeconfig for the cluster for which machine-controller-manager will create machines.
	controlKubeconfig, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		controlKubeconfigPath := os.Getenv("CONTROL_KUBECONFIG")
		controlKubeconfig, err = clientcmd.BuildConfigFromFlags("", controlKubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	controlClientBuilder := SimpleClientBuilder{
		ClientConfig: controlKubeconfig,
	}

	machineClient := controlClientBuilder.ClientOrDie("machine-controller-manager-client").MachineV1alpha1()
	namespace := os.Getenv("CONTROL_NAMESPACE")
	machineInformerFactory := machineinformers.NewFilteredSharedInformerFactory(
		controlClientBuilder.ClientOrDie("machine-shared-informers"),
		12*time.Hour,
		namespace,
		nil,
	)
	machineSharedInformers := machineInformerFactory.Machine().V1alpha1()

	targetCoreKubeconfigPath := os.Getenv("TARGET_KUBECONFIG")
	targetCoreKubeconfig, err := clientcmd.BuildConfigFromFlags("", targetCoreKubeconfigPath)
	if err != nil {
		return nil, err
	}

	targetCoreClient, err := kubernetes.NewForConfig(targetCoreKubeconfig)
	if err != nil {
		return nil, err
	}

	manager := &McmManager{
		namespace:               namespace,
		interrupt:               make(chan struct{}),
		machineclient:           machineClient,
		coreclient:              targetCoreClient,
		machineDeploymentLister: machineSharedInformers.MachineDeployments().Lister(),
		machinelisters:          machineSharedInformers.Machines().Lister(),
		discoveryOpts:           discoveryOpts,
	}

	return manager, nil
}

//CreateMcmManager creates the McmManager
func CreateMcmManager(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*McmManager, error) {
	return createMCMManagerInternal(discoveryOpts)
}

//GetMachineDeploymentForMachine returns the MachineDeployment for the Machine object.
func (m *McmManager) GetMachineDeploymentForMachine(machine *Ref) (*MachineDeployment, error) {
	if machine.Name == "" {
		//Considering the possibility when Machine has been deleted but due to cached Node object it appears here.
		return nil, fmt.Errorf("Node does not Exists")
	}
	machineObject, err := m.machineclient.Machines(m.namespace).Get(context.TODO(), machine.Name, metav1.GetOptions{})
	if err != nil {
		if kube_errors.IsNotFound(err) {
			// Machine has been removed.
			klog.V(4).Infof("Machine was removed before it could be retrieved: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("Unable to fetch Machine object %s %+v", machine.Name, err)
	}

	var machineSetName, machineDeploymentName string
	if len(machineObject.OwnerReferences) > 0 {
		machineSetName = machineObject.OwnerReferences[0].Name
	} else {
		return nil, fmt.Errorf("Unable to find parent MachineSet of given Machine object %s %+v", machine.Name, err)
	}

	machineSetObject, err := m.machineclient.MachineSets(m.namespace).Get(context.TODO(), machineSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch MachineSet object %s %+v", machineSetName, err)
	}

	if len(machineSetObject.OwnerReferences) > 0 {
		machineDeploymentName = machineSetObject.OwnerReferences[0].Name
	} else {
		return nil, fmt.Errorf("Unable to find parent MachineDeployment of given MachineSet object %s %+v", machineSetName, err)
	}

	mcmRef := Ref{
		Name:      machineDeploymentName,
		Namespace: m.namespace,
	}

	discoveryOpts := m.discoveryOpts
	specs := discoveryOpts.NodeGroupSpecs
	var min, max int
	for _, spec := range specs {
		s, err := dynamic.SpecFromString(spec, true)
		if err != nil {
			return nil, fmt.Errorf("Error occured while parsing the spec")
		}

		str := strings.Split(s.Name, ".")
		_, Name := str[0], str[1]

		if Name == machineDeploymentName {
			min = s.MinSize
			max = s.MaxSize
			break
		}
	}

	return &MachineDeployment{
		mcmRef,
		m,
		min,
		max,
	}, nil
}

//Cleanup does nothing at the moment.
//TODO: Enable cleanup method for graceful shutdown.
func (m *McmManager) Cleanup() {
	return
}

//GetMachineDeploymentSize returns the replicas field of the MachineDeployment
func (m *McmManager) GetMachineDeploymentSize(machinedeployment *MachineDeployment) (int64, error) {
	md, err := m.machineclient.MachineDeployments(machinedeployment.Namespace).Get(context.TODO(), machinedeployment.Name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("Unable to fetch MachineDeployment object %s %+v", machinedeployment.Name, err)
	}
	return int64(md.Spec.Replicas), nil
}

//SetMachineDeploymentSize sets the desired size for the Machinedeployment.
func (m *McmManager) SetMachineDeploymentSize(machinedeployment *MachineDeployment, size int64) error {

	retryDeadline := time.Now().Add(maxRetryDeadline)
	for {
		md, err := m.machineclient.MachineDeployments(machinedeployment.Namespace).Get(context.TODO(), machinedeployment.Name, metav1.GetOptions{})
		if err != nil && time.Now().Before(retryDeadline) {
			klog.Warningf("Unable to fetch MachineDeployment object %s, Error: %+v", machinedeployment.Name, err)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s", machinedeployment.Name, err)
			return err
		}

		clone := md.DeepCopy()
		clone.Spec.Replicas = int32(size)

		_, err = m.machineclient.MachineDeployments(machinedeployment.Namespace).Update(context.TODO(), clone, metav1.UpdateOptions{})
		if err != nil && time.Now().Before(retryDeadline) {
			klog.Warningf("Unable to update MachineDeployment object %s, Error: %+v", machinedeployment.Name, err)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to update MachineDeployment object %s, Error: %s", machinedeployment.Name, err)
			return err
		}

		// Break out of loop when update succeeds
		break
	}
	return nil
}

//DeleteMachines deletes the Machines and also reduces the desired replicas of the Machinedeplyoment in parallel.
func (m *McmManager) DeleteMachines(machines []*Ref) error {

	var (
		mdclone             *v1alpha1.MachineDeployment
		terminatingMachines []*v1alpha1.Machine
	)

	if len(machines) == 0 {
		return nil
	}
	commonMachineDeployment, err := m.GetMachineDeploymentForMachine(machines[0])
	if err != nil {
		return err
	}

	for _, machine := range machines {
		machinedeployment, err := m.GetMachineDeploymentForMachine(machine)
		if err != nil {
			return err
		}
		if machinedeployment.Name != commonMachineDeployment.Name {
			return fmt.Errorf("Cannot delete machines which don't belong to the same MachineDeployment")
		}
	}

	for _, machine := range machines {

		retryDeadline := time.Now().Add(maxRetryDeadline)
		for {
			mach, err := m.machineclient.Machines(machine.Namespace).Get(context.TODO(), machine.Name, metav1.GetOptions{})
			if err != nil && time.Now().Before(retryDeadline) {
				klog.Warningf("Unable to fetch Machine object %s, Error: %s", machine.Name, err)
				time.Sleep(conflictRetryInterval)
				continue
			} else if err != nil {
				// Timeout occurred
				klog.Errorf("Unable to fetch Machine object %s, Error: %s", machine.Name, err)
				return err
			}

			mclone := mach.DeepCopy()

			if isMachineTerminating(mclone) {
				terminatingMachines = append(terminatingMachines, mclone)
			}
			if mclone.Annotations != nil {
				if mclone.Annotations[machinePriorityAnnotation] == "1" {
					klog.Infof("Machine %q priority is already set to 1, hence skipping the update", machine.Name)
					break
				}
				mclone.Annotations[machinePriorityAnnotation] = "1"
			} else {
				mclone.Annotations = make(map[string]string)
				mclone.Annotations[machinePriorityAnnotation] = "1"
			}

			_, err = m.machineclient.Machines(machine.Namespace).Update(context.TODO(), mclone, metav1.UpdateOptions{})
			if err != nil && time.Now().Before(retryDeadline) {
				klog.Warningf("Unable to update Machine object %s, Error: %s", machine.Name, err)
				time.Sleep(conflictRetryInterval)
				continue
			} else if err != nil {
				// Timeout occurred
				klog.Errorf("Unable to update Machine object %s, Error: %s", machine.Name, err)
				return err
			}

			// Break out of loop when update succeeds
			break
		}
	}

	retryDeadline := time.Now().Add(maxRetryDeadline)
	for {
		md, err := m.machineclient.MachineDeployments(commonMachineDeployment.Namespace).Get(context.TODO(), commonMachineDeployment.Name, metav1.GetOptions{})
		if err != nil && time.Now().Before(retryDeadline) {
			klog.Warningf("Unable to fetch MachineDeployment object %s, Error: %s", commonMachineDeployment.Name, err)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s", commonMachineDeployment.Name, err)
			return err
		}

		mdclone = md.DeepCopy()
		if (int(mdclone.Spec.Replicas) - len(machines)) < 0 {
			return fmt.Errorf("Unable to delete machine in MachineDeployment object %s , machine replicas are < 0 ", commonMachineDeployment.Name)
		}
		expectedReplicas := mdclone.Spec.Replicas - int32(len(machines)) + int32(len(terminatingMachines))
		if expectedReplicas == mdclone.Spec.Replicas {
			klog.Infof("MachineDeployment %q is already set to %d, skipping the update", mdclone.Name, expectedReplicas)
			break
		}

		mdclone.Spec.Replicas = expectedReplicas

		_, err = m.machineclient.MachineDeployments(mdclone.Namespace).Update(context.TODO(), mdclone, metav1.UpdateOptions{})
		if err != nil && time.Now().Before(retryDeadline) {
			klog.Warningf("Unable to update MachineDeployment object %s, Error: %s", commonMachineDeployment.Name, err)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to update MachineDeployment object %s, Error: %s", commonMachineDeployment.Name, err)
			return err
		}

		// Break out of loop when update succeeds
		break
	}

	klog.V(2).Infof("MachineDeployment %s size decreased to %d", commonMachineDeployment.Name, mdclone.Spec.Replicas)

	return nil
}

//GetMachineDeploymentNodes returns the set of Nodes which belongs to the MachineDeployment.
func (m *McmManager) GetMachineDeploymentNodes(machinedeployment *MachineDeployment) ([]string, error) {
	md, err := m.machineclient.MachineDeployments(m.namespace).Get(context.TODO(), machinedeployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch MachineDeployment object %s, Error: %v", machinedeployment.Name, err)
	}
	//machinelist, err := m.machinelisters.Machines(m.namespace).List(labels.Everything())
	machinelist, err := m.machineclient.Machines(m.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch list of Machine objects %v", err)
	}

	nodelist, err := m.coreclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch list of Nodes %v", err)
	}

	var nodes []string
	// Bearing O(n2) complexity, assuming we will not have lot of nodes/machines, open for optimisations.
	for _, machine := range machinelist.Items {
		if strings.Contains(machine.Name, md.Name) {
			var found bool
			for _, node := range nodelist.Items {
				if machine.Labels["node"] == node.Name {
					nodes = append(nodes, node.Spec.ProviderID)
					found = true
					break
				}
			}
			if !found {
				// No node found - either the machine has not registered yet or AWS is unable to fufill the request.
				// Report a special ID so that the autoscaler can track it as an unregistered node.
				nodes = append(nodes, fmt.Sprintf("requested://%s", machine.Name))
			}
		}
	}
	return nodes, nil
}

//GetMachineDeploymentNodeTemplate returns the NodeTemplate which belongs to the MachineDeployment.
func (m *McmManager) GetMachineDeploymentNodeTemplate(machinedeployment *MachineDeployment) (*nodeTemplate, error) {

	md, err := m.machineclient.MachineDeployments(m.namespace).Get(context.TODO(), machinedeployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch MachineDeployment object %s, Error: %v", machinedeployment.Name, err)
	}

	var (
		region   string
		zone     string
		instance instanceType

		machineClass     = md.Spec.Template.Spec.Class
		nodeTemplateSpec = md.Spec.Template.Spec.NodeTemplateSpec
	)

	switch machineClass.Kind {
	case kindAWSMachineClass:
		mc, err := m.machineclient.AWSMachineClasses(m.namespace).Get(context.TODO(), machineClass.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch AWSMachineClass object %s, Error: %v", machinedeployment.Name, err)
		}
		awsInstance := aws.InstanceTypes[mc.Spec.MachineType]
		instance = instanceType{
			InstanceType: awsInstance.InstanceType,
			VCPU:         awsInstance.VCPU,
			MemoryMb:     awsInstance.MemoryMb,
			GPU:          awsInstance.GPU,
		}
		region = mc.Spec.Region
		zone = getZoneValueFromMCLabels(mc.Labels)
	case kindAzureMachineClass:
		mc, err := m.machineclient.AzureMachineClasses(m.namespace).Get(context.TODO(), machineClass.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch AzureMachineClass object %s, Error: %v", machinedeployment.Name, err)
		}
		azureInstance := azure.InstanceTypes[mc.Spec.Properties.HardwareProfile.VMSize]
		instance = instanceType{
			InstanceType: azureInstance.InstanceType,
			VCPU:         azureInstance.VCPU,
			MemoryMb:     azureInstance.MemoryMb,
			GPU:          azureInstance.GPU,
		}
		region = mc.Spec.Location
		if mc.Spec.Properties.Zone != nil {
			zone = mc.Spec.Location + "-" + strconv.Itoa(*mc.Spec.Properties.Zone)
		}
	case kindMachineClass:
		mc, err := m.machineclient.MachineClasses(m.namespace).Get(context.TODO(), machineClass.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch %s for %s, Error: %v", kindMachineClass, machinedeployment.Name, err)
		}
		switch mc.Provider {
		case providerAWS:
			var providerSpec *awsapis.AWSProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("Unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAWS, machinedeployment.Name, err)
			}

			awsInstance := aws.InstanceTypes[providerSpec.MachineType]
			instance = instanceType{
				InstanceType: awsInstance.InstanceType,
				VCPU:         awsInstance.VCPU,
				MemoryMb:     awsInstance.MemoryMb,
				GPU:          awsInstance.GPU,
			}
			region = providerSpec.Region
			zone = getZoneValueFromMCLabels(mc.Labels)
		case providerAzure:
			var providerSpec *azureapis.AzureProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("Unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAzure, machinedeployment.Name, err)
			}

			azureInstance := aws.InstanceTypes[providerSpec.Properties.HardwareProfile.VMSize]
			instance = instanceType{
				InstanceType: azureInstance.InstanceType,
				VCPU:         azureInstance.VCPU,
				MemoryMb:     azureInstance.MemoryMb,
				GPU:          azureInstance.GPU,
			}
			region = providerSpec.Location
			if providerSpec.Properties.Zone != nil {
				zone = providerSpec.Location + "-" + strconv.Itoa(*providerSpec.Properties.Zone)
			}
		default:
			return nil, cloudprovider.ErrNotImplemented
		}
	default:
		return nil, cloudprovider.ErrNotImplemented
	}

	labels := make(map[string]string)
	taints := make([]apiv1.Taint, 0)

	if nodeTemplateSpec.ObjectMeta.Labels != nil {
		labels = nodeTemplateSpec.ObjectMeta.Labels
	}
	if nodeTemplateSpec.Spec.Taints != nil {
		taints = nodeTemplateSpec.Spec.Taints
	}

	nodeTmpl := &nodeTemplate{
		InstanceType: &instance,
		Region:       region,
		Zone:         zone, // will be implemented in MCM
		Labels:       labels,
		Taints:       taints,
	}

	return nodeTmpl, nil
}

func getZoneValueFromMCLabels(labels map[string]string) string {
	var zone string

	if labels != nil {
		if value, exists := labels[apiv1.LabelZoneFailureDomainStable]; exists {
			// Prefer zone value from the new label
			zone = value
		} else if value, exists := labels[apiv1.LabelZoneFailureDomain]; exists {
			// Fallback to zone value from deprecated label if new lable value doesn't exist
			zone = value
		}
	}

	return zone
}

func (m *McmManager) buildNodeFromTemplate(name string, template *nodeTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// Numbers pods per node will depends on the CNI used and the maxPods kubelet config, default is often 100
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(100, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.VCPU, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.InstanceType.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = template.Labels
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = template.Taints

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *nodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract from MCM
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch

	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceType] = template.InstanceType.InstanceType
	result[apiv1.LabelInstanceTypeStable] = template.InstanceType.InstanceType

	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelZoneRegionStable] = template.Region

	result[apiv1.LabelZoneFailureDomain] = template.Zone
	result[apiv1.LabelZoneFailureDomainStable] = template.Zone

	result[apiv1.LabelHostname] = nodeName
	return result
}

// isMachineTerminating returns true if machine is already being terminated or considered for termination by autoscaler.
func isMachineTerminating(machine *v1alpha1.Machine) bool {
	if !machine.GetDeletionTimestamp().IsZero() {
		klog.Infof("Machine %q is already being terminated, and hence skipping the deletion", machine.Name)
		return true
	} else if machine.Annotations != nil && machine.Annotations[machinePriorityAnnotation] == "1" {
		klog.Infof("Machine %q 's priority is set to 1, we assume it to be triggered for deletion by autoscaler earlier, hence skipping deletion", machine.Name)
		return true
	}
	return false
}
