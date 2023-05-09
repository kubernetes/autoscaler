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
	v1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/discovery"
	coreinformers "k8s.io/client-go/informers"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeletapis "k8s.io/kubelet/pkg/apis"
)

const (
	maxRetryTimeout        = 1 * time.Minute
	conflictRetryInterval  = 5 * time.Second
	minResyncPeriodDefault = 1 * time.Hour
	// machinePriorityAnnotation is the annotation to set machine priority while deletion
	machinePriorityAnnotation = "machinepriority.machine.sapcloud.io"
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
	// machineDeploymentProgressing tells that deployment is progressing. Progress for a MachineDeployment is considered when a new machine set is created or adopted, and when new machines scale up or old machines scale down.
	// Progress is not estimated for paused MachineDeployments. It is also updated if progressDeadlineSeconds is not specified(treated as infinite deadline), in which case it would never be updated to "false".
	machineDeploymentProgressing v1alpha1.MachineDeploymentConditionType = "Progressing"
	// newISAvailableReason is the reason in "Progressing" condition when machineDeployment rollout is complete
	newISAvailableReason = "NewMachineSetAvailable"
	// conditionTrue means the given condition status is true
	conditionTrue v1alpha1.ConditionStatus = "True"
)

var (
	controlBurst    *int
	controlQPS      *float64
	targetBurst     *int
	targetQPS       *float64
	minResyncPeriod *time.Duration

	machineClassGVR      = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machineclasses"}
	machineGVR           = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machines"}
	machineSetGVR        = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machinesets"}
	machineDeploymentGVR = schema.GroupVersionResource{Group: machineGroup, Version: machineVersion, Resource: "machinedeployments"}
)

// McmManager manages the client communication for MachineDeployments.
type McmManager struct {
	namespace               string
	interrupt               chan struct{}
	discoveryOpts           cloudprovider.NodeGroupDiscoveryOptions
	machineClient           machineapi.MachineV1alpha1Interface
	machineDeploymentLister machinelisters.MachineDeploymentLister
	machineSetLister        machinelisters.MachineSetLister
	machineLister           machinelisters.MachineLister
	machineClassLister      machinelisters.MachineClassLister
	nodeLister              corelisters.NodeLister
}

type instanceType struct {
	InstanceType     string
	VCPU             resource.Quantity
	Memory           resource.Quantity
	GPU              resource.Quantity
	EphemeralStorage resource.Quantity
	PodCount         resource.Quantity
}

type nodeTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Labels       map[string]string
	Taints       []apiv1.Taint
}

type machineNameNodeNamePair struct {
	machineName string
	nodeName    string
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
			machineClassLister machinelisters.MachineClassLister
			syncFuncs          []cache.InformerSynced
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
// users don't have to restart their controller manager if they change the apiserver.
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

	_, resourceMap, err := discoveryClient.ServerGroupsAndResources()
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
		// Considering the possibility when Machine has been deleted but due to cached Node object it appears here.
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
		return nil, fmt.Errorf("unable to find parent MachineDeployment of given MachineSet object %s %+v", machineSetName, err)
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
			return nil, fmt.Errorf("Error occurred while parsing the spec")
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
	md, err := m.getMachineDeploymentUntilDeadline(machinedeployment.Name, conflictRetryInterval, time.Now().Add(maxRetryTimeout))
	if err != nil {
		klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s", machinedeployment.Name, err)
		return err
	}

	// don't scale down during rolling update, as that could remove ready node with workload
	if md.Spec.Replicas >= int32(size) && !isRollingUpdateFinished(md) {
		return fmt.Errorf("MachineDeployment %s is under rolling update , cannot reduce replica count", md.Name)
	}

	retryDeadline := time.Now().Add(maxRetryTimeout)
	for {
		// fetching fresh copy of machineDeployment
		md, err := m.getMachineDeploymentUntilDeadline(machinedeployment.Name, conflictRetryInterval, retryDeadline)
		if err != nil {
			klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s", machinedeployment.Name, err)
			return err
		}

		clone := md.DeepCopy()
		clone.Spec.Replicas = int32(size)

		_, err = m.machineClient.MachineDeployments(machinedeployment.Namespace).Update(context.TODO(), clone, metav1.UpdateOptions{})
		if err != nil && time.Now().Before(retryDeadline) {
			klog.Warningf("Unable to update MachineDeployment object %s, Error: %+v , will retry in %s", machinedeployment.Name, err, conflictRetryInterval)
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
		mdclone                             *v1alpha1.MachineDeployment
		terminatingMachines                 []*v1alpha1.Machine
		expectedToTerminateMachineNodePairs []machineNameNodeNamePair
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
			return fmt.Errorf("cannot delete machines which don't belong to the same MachineDeployment")
		}
	}

	md, err := m.getMachineDeploymentUntilDeadline(commonMachineDeployment.Name, conflictRetryInterval, time.Now().Add(maxRetryTimeout))
	if err != nil {
		klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s", commonMachineDeployment.Name, err)
		return err
	}
	if !isRollingUpdateFinished(md) {
		return fmt.Errorf("MachineDeployment %s is under rolling update , cannot reduce replica count", commonMachineDeployment.Name)
	}

	for _, machine := range machines {

		// Trying to update the priority of machine till retryDeadline
		retryDeadline := time.Now().Add(maxRetryTimeout)
		for {
			machine, err := m.machineLister.Machines(m.namespace).Get(machine.Name)
			if err != nil && time.Now().Before(retryDeadline) {
				klog.Warningf("Unable to fetch Machine object %s, Error: %s , will retry in %s", machine.Name, err, conflictRetryInterval)
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
			} else {
				expectedToTerminateMachineNodePairs = append(expectedToTerminateMachineNodePairs, machineNameNodeNamePair{mclone.Name, mclone.Labels["node"]})
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
				klog.Warningf("Unable to update Machine object %s, Error: %s , will retry in %s", machine.Name, err, conflictRetryInterval)
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

		klog.Infof("Machine %s of machineDeployment %s marked with priority 1 successfully", machine.Name, md.Name)
	}

	// Trying to update the machineDeployment till the retryDeadline
	retryDeadline := time.Now().Add(maxRetryTimeout)
	for {
		// fetch fresh copy of machineDeployment
		md, err = m.getMachineDeploymentUntilDeadline(commonMachineDeployment.Name, conflictRetryInterval, retryDeadline)
		if err != nil {
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
			klog.Warningf("Unable to update MachineDeployment object %s, Error: %s , will retry in %s", commonMachineDeployment.Name, err, conflictRetryInterval)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to update MachineDeployment object %s, Error: %s , timeout occurred", commonMachineDeployment.Name, err)
			return err
		}

		// Break out of loop when update succeeds
		break
	}

	klog.V(2).Infof("MachineDeployment %s size decreased to %d , should remove following {machine, corresponding node} pairs %s ", commonMachineDeployment.Name, mdclone.Spec.Replicas, expectedToTerminateMachineNodePairs)

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
				// No node found - either the machine has not registered yet or AWS is unable to fulfill the request.
				// Report a special ID so that the autoscaler can track it as an unregistered node.
				nodes = append(nodes, fmt.Sprintf("requested://%s", machine.Name))
			}
		}
	}
	return nodes, nil
}

// validateNodeTemplate function validates the NodeTemplate object of the MachineClass
func validateNodeTemplate(nodeTemplateAttributes *v1alpha1.NodeTemplate) error {
	var allErrs []error

	capacityAttributes := []v1.ResourceName{"cpu", "gpu", "memory"}

	for _, attribute := range capacityAttributes {
		if _, ok := nodeTemplateAttributes.Capacity[attribute]; !ok {
			errMessage := fmt.Errorf("CPU, GPU and memory fields are mandatory")
			klog.Warning(errMessage)
			allErrs = append(allErrs, errMessage)
			break
		}
	}

	if nodeTemplateAttributes.Region == "" || nodeTemplateAttributes.InstanceType == "" || nodeTemplateAttributes.Zone == "" {
		errMessage := fmt.Errorf("InstanceType, Region and Zone attributes are mandatory")
		klog.Warning(errMessage)
		allErrs = append(allErrs, errMessage)
	}

	if allErrs != nil {
		return fmt.Errorf("%s", allErrs)
	}

	return nil
}

// GetMachineDeploymentNodeTemplate returns the NodeTemplate of a node belonging to the same worker pool as the machinedeployment
// If no node present then it forms the nodeTemplate using the one present in machineClass
func (m *McmManager) GetMachineDeploymentNodeTemplate(machinedeployment *MachineDeployment) (*nodeTemplate, error) {

	md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(machinedeployment.Name)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch MachineDeployment object %s, Error: %v", machinedeployment.Name, err)
	}

	var (
		workerPool       = getWorkerPoolForMachineDeploy(md)
		list             = []string{workerPool}
		selector         = labels.NewSelector()
		req, _           = labels.NewRequirement(nodegroupset.LabelWorkerPool, selection.Equals, list)
		region           string
		zone             string
		instance         instanceType
		machineClass     = md.Spec.Template.Spec.Class
		nodeTemplateSpec = md.Spec.Template.Spec.NodeTemplateSpec
	)

	selector = selector.Add(*req)
	nodes, err := m.nodeLister.List(selector)
	if err != nil {
		return nil, fmt.Errorf("error fetching node object for worker pool %s, Error: %v", workerPool, err)
	}

	switch machineClass.Kind {
	case kindMachineClass:
		mc, err := m.machineClassLister.MachineClasses(m.namespace).Get(machineClass.Name)
		if err != nil {
			return nil, fmt.Errorf("Unable to fetch %s for %s, Error: %v", kindMachineClass, machineClass.Name, err)
		}

		if nodeTemplateAttributes := mc.NodeTemplate; nodeTemplateAttributes != nil {

			err := validateNodeTemplate(nodeTemplateAttributes)
			if err != nil {
				return nil, fmt.Errorf("nodeTemplate validation error in MachineClass %s : %s", mc.Name, err)
			}

			filteredNodes := filterOutNodes(nodes, nodeTemplateAttributes.InstanceType)

			if len(filteredNodes) > 0 {
				klog.V(1).Infof("Nodes already existing in the worker pool %s", workerPool)
				baseNode := filteredNodes[0]
				klog.V(1).Infof("Worker pool node used to form template is %s and its capacity is cpu: %s, memory:%s", baseNode.Name, baseNode.Status.Capacity.Cpu().String(), baseNode.Status.Capacity.Memory().String())
				instance = instanceType{
					VCPU:             baseNode.Status.Capacity[apiv1.ResourceCPU],
					Memory:           baseNode.Status.Capacity[apiv1.ResourceMemory],
					GPU:              baseNode.Status.Capacity[gpu.ResourceNvidiaGPU],
					EphemeralStorage: baseNode.Status.Capacity[apiv1.ResourceEphemeralStorage],
					PodCount:         baseNode.Status.Capacity[apiv1.ResourcePods],
				}
			} else {
				klog.V(1).Infof("Generating node template only using nodeTemplate from MachineClass %s: template resources-> cpu: %s,memory: %s", machineClass.Name, nodeTemplateAttributes.Capacity.Cpu().String(), nodeTemplateAttributes.Capacity.Memory().String())
				instance = instanceType{
					VCPU:   nodeTemplateAttributes.Capacity[apiv1.ResourceCPU],
					Memory: nodeTemplateAttributes.Capacity[apiv1.ResourceMemory],
					GPU:    nodeTemplateAttributes.Capacity["gpu"],
					// Numbers pods per node will depends on the CNI used and the maxPods kubelet config, default is often 110
					PodCount: resource.MustParse("110"),
				}
			}
			instance.InstanceType = nodeTemplateAttributes.InstanceType
			region = nodeTemplateAttributes.Region
			zone = nodeTemplateAttributes.Zone
			break
		}

		switch mc.Provider {
		case providerAWS:
			var providerSpec *awsapis.AWSProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("Unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAWS, machinedeployment.Name, err)
			}

			awsInstance, exists := AWSInstanceTypes[providerSpec.MachineType]
			if !exists {
				return nil, fmt.Errorf("Unable to fetch details for VM type %s", providerSpec.MachineType)
			}
			instance = instanceType{
				InstanceType: awsInstance.InstanceType,
				VCPU:         awsInstance.VCPU,
				Memory:       awsInstance.Memory,
				GPU:          awsInstance.GPU,
				// Numbers pods per node will depends on the CNI used and the maxPods kubelet config, default is often 110
				PodCount: resource.MustParse("110"),
			}
			region = providerSpec.Region
			zone = getZoneValueFromMCLabels(mc.Labels)
		case providerAzure:
			var providerSpec *azureapis.AzureProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("Unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAzure, machinedeployment.Name, err)
			}
			azureInstance, exists := AzureInstanceTypes[providerSpec.Properties.HardwareProfile.VMSize]
			if !exists {
				return nil, fmt.Errorf("Unable to fetch details for VM type %s", providerSpec.Properties.HardwareProfile.VMSize)
			}
			instance = instanceType{
				InstanceType: azureInstance.InstanceType,
				VCPU:         azureInstance.VCPU,
				Memory:       azureInstance.Memory,
				GPU:          azureInstance.GPU,
				// Numbers pods per node will depends on the CNI used and the maxPods kubelet config, default is often 110
				PodCount: resource.MustParse("110"),
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

func isRollingUpdateFinished(md *v1alpha1.MachineDeployment) bool {
	for _, cond := range md.Status.Conditions {
		switch {
		case cond.Type == machineDeploymentProgressing && cond.Status == conditionTrue && cond.Reason == newISAvailableReason:
			return true
		case cond.Type == machineDeploymentProgressing:
			return false
		}
	}
	// no "Progressing" condition means the deployment has not undergone any rolling update yet
	return true
}

// getMachineDeploymentUntilDeadline returns error only when fetching the machineDeployment has been failing consequently and deadline is crossed
func (m *McmManager) getMachineDeploymentUntilDeadline(mdName string, retryInterval time.Duration, deadline time.Time) (*v1alpha1.MachineDeployment, error) {
	for {
		md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(mdName)
		if err != nil && time.Now().Before(deadline) {
			klog.Warningf("Unable to fetch MachineDeployment object %s, Error: %s, will retry in %s", mdName, err, retryInterval)
			time.Sleep(conflictRetryInterval)
			continue
		} else if err != nil {
			// Timeout occurred
			klog.Errorf("Unable to fetch MachineDeployment object %s, Error: %s, timeout occurred", mdName, err)
			return nil, err
		}
		return md, nil
	}
}

func filterOutNodes(nodes []*v1.Node, instanceType string) []*v1.Node {
	var filteredNodes []*v1.Node
	for _, node := range nodes {
		if node.Status.Capacity != nil && getInstanceTypeForNode(node) == instanceType {
			filteredNodes = append(filteredNodes, node)
		}
	}

	return filteredNodes
}

func getInstanceTypeForNode(node *v1.Node) string {
	var instanceTypeLabelValue string
	if node.Labels != nil {
		if val, ok := node.Labels[apiv1.LabelInstanceTypeStable]; ok {
			instanceTypeLabelValue = val
		} else if val, ok := node.Labels[apiv1.LabelInstanceType]; ok {
			instanceTypeLabelValue = val
		}
	}

	return instanceTypeLabelValue
}

func getWorkerPoolForMachineDeploy(md *v1alpha1.MachineDeployment) string {
	if md.Spec.Template.Spec.NodeTemplateSpec.Labels != nil {
		if value, exists := md.Spec.Template.Spec.NodeTemplateSpec.Labels[nodegroupset.LabelWorkerPool]; exists {
			return value
		}
	}

	return ""
}

func getZoneValueFromMCLabels(labels map[string]string) string {
	var zone string

	if labels != nil {
		if value, exists := labels[apiv1.LabelZoneFailureDomainStable]; exists {
			// Prefer zone value from the new label
			zone = value
		} else if value, exists := labels[apiv1.LabelZoneFailureDomain]; exists {
			// Fallback to zone value from deprecated label if new label value doesn't exist
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

	node.Status.Capacity[apiv1.ResourcePods] = template.InstanceType.PodCount
	node.Status.Capacity[apiv1.ResourceCPU] = template.InstanceType.VCPU
	if template.InstanceType.GPU.Cmp(resource.MustParse("0")) != 0 {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = template.InstanceType.GPU
	}
	node.Status.Capacity[apiv1.ResourceMemory] = template.InstanceType.Memory
	node.Status.Capacity[apiv1.ResourceEphemeralStorage] = template.InstanceType.EphemeralStorage
	// added most common hugepages sizes. This will help to consider the template node while finding similar node groups
	node.Status.Capacity["hugepages-1Gi"] = *resource.NewQuantity(0, resource.DecimalSI)
	node.Status.Capacity["hugepages-2Mi"] = *resource.NewQuantity(0, resource.DecimalSI)

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
