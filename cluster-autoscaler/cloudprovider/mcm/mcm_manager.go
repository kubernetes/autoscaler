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
	"flag"
	"fmt"
	"math/rand"
	"net/http"
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
	apiv1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	aws "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws"
	azure "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/discovery"
	coreinformers "k8s.io/client-go/informers"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	maxRetryDeadline        = 1 * time.Minute
	conflictRetryInterval   = 5 * time.Second
	minResyncPeriodDefault  = 1 * time.Hour
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
	// machineGroup is the group version used to identify machine API group objects
	machineGroup = "machine.sapcloud.io"
	// machineGroup is the API version used to identify machine API group objects
	machineVersion = "v1alpha1"
)

var (
	controlBurst    *int
	controlQPS      *float64
	targetBurst     *int
	targetQPS       *float64
	minResyncPeriod *time.Duration

	awsMachineClassGVR   = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "awsmachineclasses"}
	azureMachineClassGVR = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "azuremachineclasses"}
	machineClassGVR      = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machineclasses"}
	machineGVR           = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machines"}
	machineSetGVR        = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machinesets"}
	machineDeploymentGVR = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machinedeployments"}
)

//McmManager manages the client communication for MachineDeployments.
type McmManager struct {
	namespace               string
	interrupt               chan struct{}
	discoveryOpts           cloudprovider.NodeGroupDiscoveryOptions
	machineClient           machineapi.MachineV1alpha1Interface
	machineDeploymentLister machinelisters.MachineDeploymentLister
	machineSetLister        machinelisters.MachineSetLister
	machineLister           machinelisters.MachineLister
	machineClassLister      machinelisters.MachineClassLister
	azureMachineClassLister machinelisters.AzureMachineClassLister
	awsMachineClassLister   machinelisters.AWSMachineClassLister
	nodeLister              corelisters.NodeLister
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

func init() {
	controlBurst = flag.Int("control-apiserver-burst", rest.DefaultBurst, "Throttling burst configuration for the client to control cluster's apiserver.")
	controlQPS = flag.Float64("control-apiserver-qps", float64(rest.DefaultQPS), "Throttling QPS configuration for the client to control cluster's apiserver.")
	targetBurst = flag.Int("target-apiserver-burst", rest.DefaultBurst, "Throttling burst configuration for the client to target cluster's apiserver.")
	targetQPS = flag.Float64("target-apiserver-qps", float64(rest.DefaultQPS), "Throttling QPS configuration for the client to target cluster's apiserver.")
	minResyncPeriod = flag.Duration("min-resync-period", minResyncPeriodDefault, "The minimum resync period configured for the shared informers used by the MCM provider cached listers")
}

func createMCMManagerInternal(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*McmManager, error) {
	var namespace = os.Getenv("CONTROL_NAMESPACE")

	// controlKubeconfig is the cluster where the machine objects are deployed
	controlKubeconfig, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		controlKubeconfigPath := os.Getenv("CONTROL_KUBECONFIG")
		controlKubeconfig, err = clientcmd.BuildConfigFromFlags("", controlKubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	// Check if control APIServer has all requested resources
	controlCoreClientBuilder := CoreControllerClientBuilder{
		ClientConfig: controlKubeconfig,
	}
	availableResources, err := getAvailableResources(controlCoreClientBuilder)
	if err != nil {
		return nil, err
	}

	if availableResources[machineGVR] && availableResources[machineSetGVR] && availableResources[machineDeploymentGVR] {
		var (
			awsMachineClassLister   machinelisters.AWSMachineClassLister
			azureMachineClassLister machinelisters.AzureMachineClassLister
			machineClassLister      machinelisters.MachineClassLister
			syncFuncs               []cache.InformerSynced
		)

		// Initialize control kubeconfig informer factory
		controlKubeconfig.Burst = *controlBurst
		controlKubeconfig.QPS = float32(*controlQPS)
		controlMachineClientBuilder := MachineControllerClientBuilder{
			ClientConfig: controlKubeconfig,
		}
		controlMachineInformerFactory := machineinformers.NewFilteredSharedInformerFactory(
			controlMachineClientBuilder.ClientOrDie("control-machine-shared-informers"),
			*minResyncPeriod,
			namespace,
			nil,
		)

		controlMachineClient := controlMachineClientBuilder.ClientOrDie("control-machine-client").MachineV1alpha1()
		machineSharedInformers := controlMachineInformerFactory.Machine().V1alpha1()

		// Initialize mandatory control cluster informers
		machineInformer := machineSharedInformers.Machines().Informer()
		machineSetInformer := machineSharedInformers.MachineSets().Informer()
		machineDeploymentInformer := machineSharedInformers.MachineDeployments().Informer()

		// Initialize optional control cluster informers
		if availableResources[awsMachineClassGVR] {
			awsMachineClassInformer := machineSharedInformers.AWSMachineClasses().Informer()
			awsMachineClassLister = machineSharedInformers.AWSMachineClasses().Lister()
			syncFuncs = append(syncFuncs, awsMachineClassInformer.HasSynced)
		}
		if availableResources[azureMachineClassGVR] {
			azureMachineClassInformer := machineSharedInformers.AzureMachineClasses().Informer()
			azureMachineClassLister = machineSharedInformers.AzureMachineClasses().Lister()
			syncFuncs = append(syncFuncs, azureMachineClassInformer.HasSynced)
		}
		if availableResources[machineClassGVR] {
			machineClassInformer := machineSharedInformers.MachineClasses().Informer()
			machineClassLister = machineSharedInformers.MachineClasses().Lister()
			syncFuncs = append(syncFuncs, machineClassInformer.HasSynced)
		}

		// targetKubeconfig for the cluster for which nodes will be managed
		targetKubeconfigPath := os.Getenv("TARGET_KUBECONFIG")
		targetKubeconfig, err := clientcmd.BuildConfigFromFlags("", targetKubeconfigPath)
		if err != nil {
			return nil, err
		}

		// Initialize target kubeconfig informer factory
		targetKubeconfig.Burst = *targetBurst
		targetKubeconfig.QPS = float32(*targetQPS)
		targetCoreClientBuilder := CoreControllerClientBuilder{
			ClientConfig: targetKubeconfig,
		}
		targetCoreInformerFactory := coreinformers.NewSharedInformerFactory(
			targetCoreClientBuilder.ClientOrDie("target-core-shared-informers"),
			*minResyncPeriod,
		)

		// Initialize mandatory target cluster node informer
		coreSharedInformers := targetCoreInformerFactory.Core().V1()
		nodeInformer := coreSharedInformers.Nodes().Informer()

		m := &McmManager{
			namespace:               namespace,
			interrupt:               make(chan struct{}),
			machineClient:           controlMachineClient,
			awsMachineClassLister:   awsMachineClassLister,
			azureMachineClassLister: azureMachineClassLister,
			machineClassLister:      machineClassLister,
			machineLister:           machineSharedInformers.Machines().Lister(),
			machineSetLister:        machineSharedInformers.MachineSets().Lister(),
			machineDeploymentLister: machineSharedInformers.MachineDeployments().Lister(),
			nodeLister:              coreSharedInformers.Nodes().Lister(),
			discoveryOpts:           discoveryOpts,
		}

		targetCoreInformerFactory.Start(m.interrupt)
		controlMachineInformerFactory.Start(m.interrupt)

		syncFuncs = append(
			syncFuncs,
			machineInformer.HasSynced,
			machineSetInformer.HasSynced,
			machineDeploymentInformer.HasSynced,
			nodeInformer.HasSynced,
		)

		if !cache.WaitForCacheSync(m.interrupt, syncFuncs...) {
			return nil, fmt.Errorf("Timed out waiting for caches to sync")
		}

		return m, nil
	}

	return nil, fmt.Errorf("Unable to start cloud provider MCM for cluster autoscaler: API GroupVersion %q or %q or %q is not available; \nFound: %#v", machineGVR, machineSetGVR, machineDeploymentGVR, availableResources)
}

// TODO: In general, any controller checking this needs to be dynamic so
//  users don't have to restart their controller manager if they change the apiserver.
// Until we get there, the structure here needs to be exposed for the construction of a proper ControllerContext.
func getAvailableResources(clientBuilder CoreClientBuilder) (map[schema.GroupVersionResource]bool, error) {
	var discoveryClient discovery.DiscoveryInterface

	var healthzContent string
	// If apiserver is not running we should wait for some time and fail only then. This is particularly
	// important when we start apiserver and controller manager at the same time.
	err := wait.PollImmediate(time.Second, 10*time.Second, func() (bool, error) {
		client, err := clientBuilder.Client("controller-discovery")
		if err != nil {
			klog.Errorf("Failed to get api versions from server: %v", err)
			return false, nil
		}

		healthStatus := 0
		resp := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			klog.Errorf("Server isn't healthy yet.  Waiting a little while.")
			return false, nil
		}
		content, _ := resp.Raw()
		healthzContent = string(content)

		discoveryClient = client.Discovery()
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get api versions from server: %v: %v", healthzContent, err)
	}

	resourceMap, err := discoveryClient.ServerResources()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get all supported resources from server: %v", err))
	}
	if len(resourceMap) == 0 {
		return nil, fmt.Errorf("unable to get any supported resources from server")
	}

	allResources := map[schema.GroupVersionResource]bool{}
	for _, apiResourceList := range resourceMap {
		version, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		for _, apiResource := range apiResourceList.APIResources {
			allResources[version.WithResource(apiResource.Name)] = true
		}
	}

	return allResources, nil
}

// CreateMcmManager creates the McmManager
func CreateMcmManager(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*McmManager, error) {
	return createMCMManagerInternal(discoveryOpts)
}

// GetMachineDeploymentForMachine returns the MachineDeployment for the Machine object.
func (m *McmManager) GetMachineDeploymentForMachine(machine *Ref) (*MachineDeployment, error) {
	if machine.Name == "" {
		//Considering the possibility when Machine has been deleted but due to cached Node object it appears here.
		return nil, fmt.Errorf("Node does not Exists")
	}

	machineObject, err := m.machineLister.Machines(m.namespace).Get(machine.Name)
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

	machineSetObject, err := m.machineSetLister.MachineSets(m.namespace).Get(machineSetName)
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

// Refresh does nothing at the moment.
//
func (m *McmManager) Refresh() error {
	return nil
}

// Cleanup does nothing at the moment.
// TODO: Enable cleanup method for graceful shutdown.
func (m *McmManager) Cleanup() {
	return
}

// GetMachineDeploymentSize returns the replicas field of the MachineDeployment
func (m *McmManager) GetMachineDeploymentSize(machinedeployment *MachineDeployment) (int64, error) {
	md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(machinedeployment.Name)
	if err != nil {
		return 0, fmt.Errorf("Unable to fetch MachineDeployment object %s %+v", machinedeployment.Name, err)
	}
	return int64(md.Spec.Replicas), nil
}

// SetMachineDeploymentSize sets the desired size for the Machinedeployment.
func (m *McmManager) SetMachineDeploymentSize(machinedeployment *MachineDeployment, size int64) error {

	retryDeadline := time.Now().Add(maxRetryDeadline)
	for {
		md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(machinedeployment.Name)
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

		_, err = m.machineClient.MachineDeployments(machinedeployment.Namespace).Update(context.TODO(), clone, metav1.UpdateOptions{})
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

// DeleteMachines deletes the Machines and also reduces the desired replicas of the Machinedeplyoment in parallel.
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
			machine, err := m.machineLister.Machines(m.namespace).Get(machine.Name)
			if err != nil && time.Now().Before(retryDeadline) {
				klog.Warningf("Unable to fetch Machine object %s, Error: %s", machine.Name, err)
				time.Sleep(conflictRetryInterval)
				continue
			} else if err != nil {
				// Timeout occurred
				klog.Errorf("Unable to fetch Machine object %s, Error: %s", machine.Name, err)
				return err
			}

			mclone := machine.DeepCopy()

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

			_, err = m.machineClient.Machines(machine.Namespace).Update(context.TODO(), mclone, metav1.UpdateOptions{})
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
		md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(commonMachineDeployment.Name)
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

		_, err = m.machineClient.MachineDeployments(mdclone.Namespace).Update(context.TODO(), mdclone, metav1.UpdateOptions{})
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

// GetMachineDeploymentNodes returns the set of Nodes which belongs to the MachineDeployment.
func (m *McmManager) GetMachineDeploymentNodes(machinedeployment *MachineDeployment) ([]string, error) {
	md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(machinedeployment.Name)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch MachineDeployment object %s, Error: %v", machinedeployment.Name, err)
	}

	machineList, err := m.machineLister.Machines(m.namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch list of Machine objects %v", err)
	}

	nodeList, err := m.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch list of Nodes %v", err)
	}

	var nodes []string
	// Bearing O(n2) complexity, assuming we will not have lot of nodes/machines, open for optimisations.
	for _, machine := range machineList {
		if strings.Contains(machine.Name, md.Name) {
			var found bool
			for _, node := range nodeList {
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

// GetMachineDeploymentNodeTemplate returns the NodeTemplate which belongs to the MachineDeployment.
func (m *McmManager) GetMachineDeploymentNodeTemplate(machinedeployment *MachineDeployment) (*nodeTemplate, error) {

	md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(machinedeployment.Name)
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
		mc, err := m.awsMachineClassLister.AWSMachineClasses(m.namespace).Get(machineClass.Name)
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch AWSMachineClass object %s, Error: %v", machineClass.Name, err)
		}
		awsInstance, exists := aws.InstanceTypes[mc.Spec.MachineType]
		if !exists {
			return nil, fmt.Errorf("Unable to fetch details for VM type %s", mc.Spec.MachineType)
		}
		instance = instanceType{
			InstanceType: awsInstance.InstanceType,
			VCPU:         awsInstance.VCPU,
			MemoryMb:     awsInstance.MemoryMb,
			GPU:          awsInstance.GPU,
		}
		region = mc.Spec.Region
		zone = getZoneValueFromMCLabels(mc.Labels)
	case kindAzureMachineClass:
		mc, err := m.azureMachineClassLister.AzureMachineClasses(m.namespace).Get(machineClass.Name)
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch AzureMachineClass object %s, Error: %v", machineClass.Name, err)
		}
		azureInstance, exists := azure.InstanceTypes[mc.Spec.Properties.HardwareProfile.VMSize]
		if !exists {
			return nil, fmt.Errorf("Unable to fetch details for VM type %s", mc.Spec.Properties.HardwareProfile.VMSize)
		}
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
		mc, err := m.machineClassLister.MachineClasses(m.namespace).Get(machineClass.Name)
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch %s for %s, Error: %v", kindMachineClass, machineClass.Name, err)
		}
		switch mc.Provider {
		case providerAWS:
			var providerSpec *awsapis.AWSProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("Unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAWS, machinedeployment.Name, err)
			}

			awsInstance, exists := aws.InstanceTypes[providerSpec.MachineType]
			if !exists {
				return nil, fmt.Errorf("Unable to fetch details for VM type %s", providerSpec.MachineType)
			}
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
			azureInstance, exists := azure.InstanceTypes[providerSpec.Properties.HardwareProfile.VMSize]
			if !exists {
				return nil, fmt.Errorf("Unable to fetch details for VM type %s", providerSpec.Properties.HardwareProfile.VMSize)
			}
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
	}
	return false
}
