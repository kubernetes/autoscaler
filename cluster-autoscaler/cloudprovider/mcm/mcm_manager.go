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
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"maps"
	"math"
	"math/big"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/machineutils"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/utils/ptr"

	awsapis "github.com/gardener/machine-controller-manager-provider-aws/pkg/aws/apis"
	azureapis "github.com/gardener/machine-controller-manager-provider-azure/pkg/azure/api"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	machineapi "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/typed/machine/v1alpha1"
	machineinformers "github.com/gardener/machine-controller-manager/pkg/client/informers/externalversions"
	machinelisters "github.com/gardener/machine-controller-manager/pkg/client/listers/machine/v1alpha1"
	machinecodes "github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
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
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/discovery"
	appsinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeletapis "k8s.io/kubelet/pkg/apis"
	//	"github.com/gardener/machine-controller-manager/pkg/util/provider/"
)

const (
	defaultMaxRetryTimeout = 1 * time.Minute
	defaultRetryInterval   = 5 * time.Second
	// defaultResetAnnotationTimeout is the timeout for resetting the priority annotation of a machine
	defaultResetAnnotationTimeout = 10 * time.Second
	// defaultPriorityValue is the default value for the priority annotation used by CA. It is set to 3 because MCM defaults the priority of machine it creates to 3.
	defaultPriorityValue = "3"
	// priorityValueForDeletionCandidateMachines is the priority annotation value set on machines that the CA wants to be deleted. Its value is set to 1.
	priorityValueForDeletionCandidateMachines = "1"
	minResyncPeriodDefault                    = 1 * time.Hour
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
	// newISAvailableReason is the reason in "Progressing" condition when machineDeployment rollout is complete
	newISAvailableReason = "NewMachineSetAvailable"
	// machineDeploymentPausedReason is the reason in "Progressing" condition when machineDeployment is paused
	machineDeploymentPausedReason = "DeploymentPaused"
	// machineDeploymentNameLabel key for Machine Deployment name in machine labels
	machineDeploymentNameLabel = "name"
	// poolNameLabel is the name of the label for gardener worker pool
	poolNameLabel = "worker.gardener.cloud/pool"
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

	// ErrInvalidNodeTemplate is a sentinel error that indicates that the nodeTemplate is invalid.
	ErrInvalidNodeTemplate = errors.New("invalid node template")
	coreResourceNames      = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, "gpu"}
	extraResourceNames     = []v1.ResourceName{gpu.ResourceNvidiaGPU, v1.ResourcePods, v1.ResourceEphemeralStorage}
	knownResourceNames     = slices.Concat(coreResourceNames, extraResourceNames)
)

// McmManager manages the client communication for MachineDeployments.
type McmManager struct {
	namespace               string
	interrupt               chan struct{}
	discoveryOpts           cloudprovider.NodeGroupDiscoveryOptions
	nodeGroups              map[types.NamespacedName]*nodeGroup
	deploymentLister        v1appslister.DeploymentLister
	machineClient           machineapi.MachineV1alpha1Interface
	machineDeploymentLister machinelisters.MachineDeploymentLister
	machineSetLister        machinelisters.MachineSetLister
	machineLister           machinelisters.MachineLister
	machineClassLister      machinelisters.MachineClassLister
	nodeLister              corelisters.NodeLister
	nodeInterface           corev1.NodeInterface
	maxRetryTimeout         time.Duration
	retryInterval           time.Duration
}

type instanceType struct {
	InstanceType      string
	VCPU              resource.Quantity
	Memory            resource.Quantity
	GPU               resource.Quantity
	EphemeralStorage  resource.Quantity
	PodCount          resource.Quantity
	ExtendedResources apiv1.ResourceList
}

type nodeTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Architecture *string
	Labels       map[string]string
	Taints       []apiv1.Taint
}

type machineInfo struct {
	Key                 types.NamespacedName
	NodeName            string
	FailedOrTerminating bool
}

func (m machineInfo) String() string {
	return fmt.Sprintf("(%s|%s)", m.Key, m.NodeName)
}

type scaleDownData struct {
	RevisedToBeDeletedMachineNames sets.Set[string]
	RevisedScaledownAmount         int
	RevisedMachineDeployment       *v1alpha1.MachineDeployment
}

func init() {
	controlBurst = flag.Int("control-apiserver-burst", rest.DefaultBurst, "Throttling burst configuration for the client to control cluster's apiserver.")
	controlQPS = flag.Float64("control-apiserver-qps", float64(rest.DefaultQPS), "Throttling QPS configuration for the client to control cluster's apiserver.")
	targetBurst = flag.Int("target-apiserver-burst", rest.DefaultBurst, "Throttling burst configuration for the client to target cluster's apiserver.")
	targetQPS = flag.Float64("target-apiserver-qps", float64(rest.DefaultQPS), "Throttling QPS configuration for the client to target cluster's apiserver.")
	minResyncPeriod = flag.Duration("min-resync-period", minResyncPeriodDefault, "The minimum resync period configured for the shared informers used by the MCM provider cached listers")
}

func createMCMManagerInternal(discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, retryInterval, maxRetryTimeout time.Duration) (*McmManager, error) {
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

	controlKubeconfig.Burst = *controlBurst
	controlKubeconfig.QPS = float32(*controlQPS)

	controlClientBuilder := ClientBuilder{
		ClientConfig: controlKubeconfig,
	}

	availableResources, err := getAvailableResources(controlClientBuilder)
	if err != nil {
		return nil, err
	}

	controlAppsClient := controlClientBuilder.ClientOrDie("control-apps-client")
	appsInformerFactory := appsinformers.NewSharedInformerFactory(controlAppsClient, *minResyncPeriod)
	deploymentLister := appsInformerFactory.Apps().V1().Deployments().Lister()

	if availableResources[machineGVR] && availableResources[machineSetGVR] && availableResources[machineDeploymentGVR] {
		var (
			machineClassLister machinelisters.MachineClassLister
			syncFuncs          []cache.InformerSynced
		)

		// Initialize control kubeconfig informer factory
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
		targetCoreClientBuilder := ClientBuilder{
			ClientConfig: targetKubeconfig,
		}
		targetCoreClient := targetCoreClientBuilder.ClientOrDie("target-core-shared-informers")
		targetCoreInformerFactory := coreinformers.NewSharedInformerFactory(
			targetCoreClient,
			*minResyncPeriod,
		)

		// Initialize mandatory target cluster node informer
		coreSharedInformers := targetCoreInformerFactory.Core().V1()
		nodeInformer := coreSharedInformers.Nodes().Informer()

		m := &McmManager{
			namespace:               namespace,
			interrupt:               make(chan struct{}),
			nodeGroups:              make(map[types.NamespacedName]*nodeGroup),
			deploymentLister:        deploymentLister,
			machineClient:           controlMachineClient,
			machineClassLister:      machineClassLister,
			machineLister:           machineSharedInformers.Machines().Lister(),
			machineSetLister:        machineSharedInformers.MachineSets().Lister(),
			machineDeploymentLister: machineSharedInformers.MachineDeployments().Lister(),
			nodeLister:              coreSharedInformers.Nodes().Lister(),
			nodeInterface:           targetCoreClient.CoreV1().Nodes(),
			discoveryOpts:           discoveryOpts,
			maxRetryTimeout:         maxRetryTimeout,
			retryInterval:           retryInterval,
		}
		err = m.generateMachineDeploymentMap()
		if err != nil {
			return nil, err
		}
		targetCoreInformerFactory.Start(m.interrupt)
		controlMachineInformerFactory.Start(m.interrupt)
		appsInformerFactory.Start(m.interrupt)

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

func (m *McmManager) generateMachineDeploymentMap() error {
	for _, spec := range m.discoveryOpts.NodeGroupSpecs {
		if err := m.addNodeGroup(spec); err != nil {
			return err
		}
	}
	return nil
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:namespace.machineDeploymentName
func (m *McmManager) addNodeGroup(spec string) error {
	ng, err := buildNodeGroupFromSpec(spec, m)
	if err != nil {
		return err
	}
	key := types.NamespacedName{Namespace: ng.Namespace, Name: ng.Name}
	m.nodeGroups[key] = ng
	return nil
}

// TODO: In general, any controller checking this needs to be dynamic so
// users don't have to restart their controller manager if they change the apiserver.
// Until we get there, the structure here needs to be exposed for the construction of a proper ControllerContext.
func getAvailableResources(clientBuilder ClientBuilder) (map[schema.GroupVersionResource]bool, error) {
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
	return createMCMManagerInternal(discoveryOpts, defaultRetryInterval, defaultMaxRetryTimeout)
}

// getNodeGroup returns the NodeGroup for the given fully-qualified machine name.
func (m *McmManager) getNodeGroup(machineKey types.NamespacedName) (*nodeGroup, error) {
	if machineKey.Name == "" {
		// Considering the possibility when Machine has been deleted but due to cached Node object it appears here.
		return nil, fmt.Errorf("node does not Exists")
	}

	machineObject, err := m.machineLister.Machines(m.namespace).Get(machineKey.Name)
	if err != nil {
		if kube_errors.IsNotFound(err) {
			// Machine has been removed.
			klog.V(4).Infof("Machine was removed before it could be retrieved: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("unable to fetch Machine object for given Machine name %q due to %w", machineKey.Name, err)
	}

	var machineSetName, machineDeploymentName string
	if len(machineObject.OwnerReferences) > 0 {
		machineSetName = machineObject.OwnerReferences[0].Name
	} else {
		return nil, fmt.Errorf("unable to find parent MachineSet for given Machine name %q due to: %w", machineKey.Name, err)
	}

	machineSetObject, err := m.machineSetLister.MachineSets(m.namespace).Get(machineSetName)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch MachineSet object for name %q due to: %w", machineSetName, err)
	}

	if len(machineSetObject.OwnerReferences) > 0 {
		machineDeploymentName = machineSetObject.OwnerReferences[0].Name
	} else {
		return nil, fmt.Errorf("unable to find parent MachineDeployment of given MachineSet name %q due to: %w", machineSetName, err)
	}

	lookupKey := types.NamespacedName{Namespace: m.namespace, Name: machineDeploymentName}
	ng, ok := m.nodeGroups[lookupKey]
	if !ok {
		return nil, fmt.Errorf("could not find NodeGroup for MachineDeployment %q in the managed nodeGroups", machineDeploymentName)
	}
	return ng, nil
}

// Refresh method for the McmManager that will invoke NodeGroup.Refresh for each node gorup and return collected errors.
func (m *McmManager) Refresh() error {
	var collectiveError []error
	for _, ng := range m.nodeGroups {
		collectiveError = append(collectiveError, ng.Refresh())
	}
	return errors.Join(collectiveError...)
}

// Cleanup does nothing at the moment.
// TODO: Enable cleanup method for graceful shutdown.
func (m *McmManager) Cleanup() {
	return
}

// GetMachineDeploymentSize returns the replicas field of the MachineDeployment corresponding to the given node group.
func (m *McmManager) GetMachineDeploymentSize(nodeGroupName string) (int64, error) {
	md, err := m.GetMachineDeploymentObject(nodeGroupName)
	if err != nil {
		return 0, err
	}
	return int64(md.Spec.Replicas), nil
}

// SetMachineDeploymentSize sets the desired size for the backing MachineDeployment of the given nodeGroup.
func (m *McmManager) SetMachineDeploymentSize(ctx context.Context, nodeGroup *nodeGroup, size int64) (bool, error) {
	md, err := m.GetMachineDeploymentObject(nodeGroup.Name)
	if err != nil {
		return true, err
	}
	// don't scale down during rolling update, as that could remove ready node with workload
	if md.Spec.Replicas >= int32(size) && !isRollingUpdateFinished(md) {
		return false, fmt.Errorf("MachineDeployment %s is under rolling update , cannot reduce replica count", md.Name)
	} // #nosec G115 (CWE-190) -- replicas will not overflow the range of int32
	clone := md.DeepCopy()
	clone.Spec.Replicas = int32(size) // #nosec G115 (CWE-190) -- replicas will not overflow the range of int32

	_, err = m.machineClient.MachineDeployments(nodeGroup.Namespace).Update(ctx, clone, metav1.UpdateOptions{})
	return true, err
}

// updateAnnotationOnMachine returns error only when updating the annotations on machine has been failing consequently and deadline is crossed
func (m *McmManager) updateAnnotationOnMachine(ctx context.Context, mcName string, key, val string) (bool, error) {
	machine, err := m.machineLister.Machines(m.namespace).Get(mcName)
	if err != nil {
		if kube_errors.IsNotFound(err) {
			klog.Warningf("Machine %s not found, skipping annotation update", mcName)
			return false, nil
		}
		klog.Errorf("Unable to fetch Machine object %s, Error: %v", mcName, err)
		return true, err
	}
	clone := machine.DeepCopy()
	if clone.Annotations == nil {
		clone.Annotations = make(map[string]string)
	}
	clone.Annotations[key] = val
	_, err = m.machineClient.Machines(machine.Namespace).Update(ctx, clone, metav1.UpdateOptions{})
	if err == nil {
		klog.Infof("Machine %s marked with priority %s successfully", mcName, val)
	}
	return true, err
}

// scaleDownMachineDeployment scales down the MachineDeployment for given name by the length of toDeleteMachineNames after removing machine names that
// are already marked for deletion in the machineutils.TriggerDeletionByMCM of the MachineDeployment.
// It then updates the machineutils.TriggerDeletionByMCM annotation with revised toBeDeletedMachineNames along with the replica count as a atomic operation.
// NOTE: Callers MUST take the NodeGroup scalingMutex before invoking this method.
func (m *McmManager) scaleDownMachineDeployment(ctx context.Context, mdName string, toBeDeletedMachineInfos []machineInfo) (bool, error) {
	md, err := m.GetMachineDeploymentObject(mdName)
	if err != nil {
		return true, err
	}

	numDeletionCandidates := len(toBeDeletedMachineInfos)
	toBeDeletedMachineNames := make([]string, 0, numDeletionCandidates)

	for _, mInfo := range toBeDeletedMachineInfos {
		toBeDeletedMachineNames = append(toBeDeletedMachineNames, mInfo.Key.Name)
	}

	data := computeScaleDownData(md, toBeDeletedMachineNames)
	if data.RevisedScaledownAmount == 0 {
		klog.V(3).Infof("Skipping scaledown since MachineDeployment %q has already marked %v for deletion by MCM, skipping the scale-down", md.Name, toBeDeletedMachineNames)
		return false, nil
	}

	if data.RevisedMachineDeployment == nil {
		klog.V(3).Infof("Skipping scaledown for MachineDeployment %q for toBeDeletedMachineNames: %v", md.Name, toBeDeletedMachineNames)
		return false, nil
	}
	updatedMd, err := m.machineClient.MachineDeployments(data.RevisedMachineDeployment.Namespace).Update(ctx, data.RevisedMachineDeployment, metav1.UpdateOptions{})
	if err != nil {
		return true, err
	}
	klog.V(2).Infof("MachineDeployment %q size decreased from %d to %d, TriggerDeletionByMCM Annotation Value: %q", md.Name, md.Spec.Replicas, updatedMd.Spec.Replicas, updatedMd.Annotations[machineutils.TriggerDeletionByMCM])

	toBeCordonedNodeNames := make([]string, 0, len(data.RevisedToBeDeletedMachineNames))
	for _, mInfo := range toBeDeletedMachineInfos {
		if data.RevisedToBeDeletedMachineNames.Has(mInfo.Key.Name) {
			toBeCordonedNodeNames = append(toBeCordonedNodeNames, mInfo.NodeName)
			klog.V(2).Infof("For MachineDeployment %q, will cordon node: %q corresponding to machine %q", md.Name, mInfo.NodeName, mInfo.Key.Name)
		}
	}
	err = m.cordonNodes(toBeCordonedNodeNames)
	if err != nil {
		// Do not return error as cordoning is best-effort
		klog.Warningf("NodeGroup.deleteMachines() of %q ran into error cordoning nodes: %v", md.Name, err)
	}
	return false, nil
}

func (m *McmManager) retry(fn func(ctx context.Context) (bool, error), resourceType string, operation string, resourceName string) error {
	ctx, cancelFn := context.WithDeadline(context.Background(), time.Now().Add(m.maxRetryTimeout))
	defer cancelFn()
	tick := time.NewTicker(m.retryInterval)
	defer tick.Stop()
	for {
		canRetry, err := fn(ctx)
		if !canRetry {
			return err
		}
		if err != nil {
			klog.Warningf("Unable to perform %s on %s object %s, Error: %v", operation, resourceType, resourceName, err)
			select {
			case <-ctx.Done():
				klog.Errorf("Context has been cancelled, %s of %s object %s will not be retried, Error: %v , timeout occurred", operation, resourceType, resourceName, ctx.Err())
				return err
			case <-tick.C:
				klog.Warningf("Will retry the operation %s on %s object %s", operation, resourceType, resourceName)
				continue
			}
		}
		return nil
	}
}

// GetInstancesForMachineDeployment returns list of cloudprovider.Instance for machines with the given nodeGroupName.
func (m *McmManager) GetInstancesForMachineDeployment(nodeGroupName string) ([]cloudprovider.Instance, error) {
	var (
		list     = []string{nodeGroupName}
		selector = labels.NewSelector()
		req, _   = labels.NewRequirement("name", selection.Equals, list)
	)

	selector = selector.Add(*req)
	machineList, err := m.machineLister.Machines(m.namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch list of Machine objects %v for MachineDeployment %q", err, nodeGroupName)
	}

	nodeList, err := m.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch list of Nodes %v", err)
	}

	instances := make([]cloudprovider.Instance, 0, len(machineList))
	// Bearing O(n2) complexity, assuming we will not have lot of nodes/machines, open for optimisations.
	for _, machine := range machineList {
		instance := findMatchingInstance(nodeList, machine)
		instances = append(instances, instance)
	}
	return instances, nil
}

func findMatchingInstance(nodes []*v1.Node, machine *v1alpha1.Machine) cloudprovider.Instance {
	for _, node := range nodes {
		if machine.Labels["node"] == node.Name {
			return cloudprovider.Instance{Id: node.Spec.ProviderID}
		}
	}
	// No k8s node found , one of the following cases possible
	//  - MCM is unable to fulfill the request to create VM.
	//  - VM is being created
	//	- the VM is up but has not registered yet

	// Report instance with a special placeholder ID so that the autoscaler can track it as an unregistered node.
	// Report InstanceStatus only for `ResourceExhausted` errors
	return cloudprovider.Instance{
		Id:     placeholderInstanceIDForMachineObj(machine.Name),
		Status: generateInstanceStatus(machine),
	}
}

func placeholderInstanceIDForMachineObj(name string) string {
	return fmt.Sprintf("requested://%s", name)
}

// generateInstanceStatus returns cloudprovider.InstanceStatus for the machine obj
func generateInstanceStatus(machine *v1alpha1.Machine) *cloudprovider.InstanceStatus {
	if machine.Status.LastOperation.Type == v1alpha1.MachineOperationCreate {
		if machine.Status.LastOperation.State == v1alpha1.MachineStateFailed && machine.Status.LastOperation.ErrorCode == machinecodes.ResourceExhausted.String() {
			return &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:    machinecodes.ResourceExhausted.String(),
					ErrorMessage: machine.Status.LastOperation.Description,
				},
			}
		}
		return &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}
	}
	return nil
}

// validateNodeTemplate function validates the NodeTemplate object of the MachineClass
func validateNodeTemplate(nodeTemplateAttributes *v1alpha1.NodeTemplate) error {
	var allErrs []error

	for _, attribute := range coreResourceNames {
		if _, ok := nodeTemplateAttributes.Capacity[attribute]; !ok {
			errMessage := fmt.Errorf("the core resource fields %q are mandatory: %w", coreResourceNames, ErrInvalidNodeTemplate)
			klog.Warning(errMessage)
			allErrs = append(allErrs, errMessage)
			break
		}
	}

	if nodeTemplateAttributes.Region == "" || nodeTemplateAttributes.InstanceType == "" || nodeTemplateAttributes.Zone == "" {
		errMessage := fmt.Errorf("InstanceType, Region and Zone attributes are mandatory: %w", ErrInvalidNodeTemplate)
		klog.Warning(errMessage)
		allErrs = append(allErrs, errMessage)
	}

	if allErrs != nil {
		return errors.Join(allErrs...)
	}

	return nil
}

// GetMachineDeploymentAnnotations returns the annotations present on the machine deployment for the provided machine deployment name
func (m *McmManager) GetMachineDeploymentAnnotations(machineDeploymentName string) (map[string]string, error) {
	md, err := m.GetMachineDeploymentObject(machineDeploymentName)
	if err != nil {
		return nil, err
	}
	return md.Annotations, nil
}

// GetMachineDeploymentNodeTemplate returns the NodeTemplate of a node belonging to the same worker pool as the MachineDeployment
// If no node present then it forms the nodeTemplate using the one present in machineClass
func (m *McmManager) GetMachineDeploymentNodeTemplate(nodeGroupName string) (*nodeTemplate, error) {
	md, err := m.GetMachineDeploymentObject(nodeGroupName)
	if err != nil {
		return nil, err
	}

	var (
		workerPool       = getWorkerPoolForMachineDeploy(md)
		list             = []string{workerPool}
		selector         = labels.NewSelector()
		req, _           = labels.NewRequirement(nodegroupset.LabelWorkerPool, selection.Equals, list)
		region           string
		zone             string
		architecture     *string
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
			return nil, fmt.Errorf("unable to fetch %s for %s, Error: %v", kindMachineClass, machineClass.Name, err)
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
				extendedResources := filterExtendedResources(baseNode.Status.Capacity)
				instance = instanceType{
					VCPU:              baseNode.Status.Capacity[apiv1.ResourceCPU],
					Memory:            baseNode.Status.Capacity[apiv1.ResourceMemory],
					GPU:               baseNode.Status.Capacity[gpu.ResourceNvidiaGPU],
					EphemeralStorage:  baseNode.Status.Capacity[apiv1.ResourceEphemeralStorage],
					PodCount:          baseNode.Status.Capacity[apiv1.ResourcePods],
					ExtendedResources: extendedResources,
				}
			} else {
				klog.V(1).Infof("Generating node template only using nodeTemplate from MachineClass %s: template resources-> cpu: %s,memory: %s", machineClass.Name, nodeTemplateAttributes.Capacity.Cpu().String(), nodeTemplateAttributes.Capacity.Memory().String())
				extendedResources := filterExtendedResources(nodeTemplateAttributes.Capacity)
				instance = instanceType{
					VCPU:             nodeTemplateAttributes.Capacity[apiv1.ResourceCPU],
					Memory:           nodeTemplateAttributes.Capacity[apiv1.ResourceMemory],
					GPU:              nodeTemplateAttributes.Capacity["gpu"],
					EphemeralStorage: nodeTemplateAttributes.Capacity[apiv1.ResourceEphemeralStorage],
					// Numbers pods per node will depends on the CNI used and the maxPods kubelet config, default is often 110
					PodCount:          resource.MustParse("110"),
					ExtendedResources: extendedResources,
				}
			}
			instance.InstanceType = nodeTemplateAttributes.InstanceType
			region = nodeTemplateAttributes.Region
			zone = nodeTemplateAttributes.Zone
			architecture = nodeTemplateAttributes.Architecture
			break
		}

		switch mc.Provider {
		case providerAWS:
			var providerSpec *awsapis.AWSProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAWS, nodeGroupName, err)
			}

			awsInstance, exists := AWSInstanceTypes[providerSpec.MachineType]
			if !exists {
				return nil, fmt.Errorf("unable to fetch details for VM type %s", providerSpec.MachineType)
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
			architecture = ptr.To(providerSpec.Tags[apiv1.LabelArchStable])
		case providerAzure:
			var providerSpec *azureapis.AzureProviderSpec
			err = json.Unmarshal(mc.ProviderSpec.Raw, &providerSpec)
			if err != nil {
				return nil, fmt.Errorf("unable to convert from %s to %s for %s, Error: %v", kindMachineClass, providerAzure, nodeGroupName, err)
			}
			azureInstance, exists := AzureInstanceTypes[providerSpec.Properties.HardwareProfile.VMSize]
			if !exists {
				return nil, fmt.Errorf("unable to fetch details for VM type %s", providerSpec.Properties.HardwareProfile.VMSize)
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
			architecture = ptr.To(providerSpec.Tags["kubernetes.io_arch"])
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
		Architecture: architecture,
	}

	return nodeTmpl, nil
}

// GetMachineDeploymentObject returns the MachineDeployment object for the provided machine deployment name
func (m *McmManager) GetMachineDeploymentObject(mdName string) (*v1alpha1.MachineDeployment, error) {
	md, err := m.machineDeploymentLister.MachineDeployments(m.namespace).Get(mdName)
	if err != nil {
		klog.Errorf("unable to fetch MachineDeployment object %q, Error: %v", mdName, err)
		return nil, fmt.Errorf("unable to fetch MachineDeployment object %q, Error: %v", mdName, err)
	}
	return md, nil
}

func isRollingUpdateFinished(md *v1alpha1.MachineDeployment) bool {
	for _, cond := range md.Status.Conditions {
		switch {
		case cond.Type == v1alpha1.MachineDeploymentProgressing && cond.Status == v1alpha1.ConditionTrue && cond.Reason == newISAvailableReason:
			return true
		// NOTE:- This check is for paused machine deployments as is taken from MCM. If the check in MCM changes, this should be updated.
		case cond.Type == v1alpha1.MachineDeploymentProgressing && cond.Status == v1alpha1.ConditionUnknown && cond.Reason == machineDeploymentPausedReason:
			return true
		case cond.Type == v1alpha1.MachineDeploymentProgressing:
			return false
		}
	}
	// no "Progressing" condition means the deployment has not undergone any rolling update yet
	return true
}

// getMachinesForMachineDeployment returns all the machines corresponding to the given machine deployment.
func (m *McmManager) getMachinesForMachineDeployment(mdName string) ([]*v1alpha1.Machine, error) {
	label, _ := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{machineDeploymentNameLabel: mdName}})
	return m.machineLister.Machines(m.namespace).List(label)
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

func filterMachinesMatchingNames(machines []*v1alpha1.Machine, matchingNames sets.Set[string]) (filteredMachines []*v1alpha1.Machine) {
	for _, m := range machines {
		if matchingNames.Has(m.Name) {
			filteredMachines = append(filteredMachines, m)
		}
	}
	return
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
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		fmt.Println("error:", err)
		return &node, err
	}
	nodeName := fmt.Sprintf("%s-%d", name, n.Int64())

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

	// populate extended resources from nodeTemplate
	if len(template.InstanceType.ExtendedResources) > 0 {
		klog.V(2).Infof("Copying extended resources %v to template node.Status.Capacity", template.InstanceType.ExtendedResources)
		maps.Copy(node.Status.Capacity, template.InstanceType.ExtendedResources)
	}

	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = template.Labels
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = template.Taints

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func (m *McmManager) cordonNodes(nodeNames []string) error {
	if len(nodeNames) == 0 {
		return nil
	}
	ctx, cancelFn := context.WithDeadline(context.Background(), time.Now().Add(m.maxRetryTimeout))
	defer cancelFn()
	var errs []error
	for _, nodeName := range nodeNames {
		node, err := m.nodeLister.Get(nodeName)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if node.Spec.Unschedulable {
			klog.V(4).Infof("Node %q is already cordoned", nodeName)
			continue
		}
		if eligibility.HasNoScaleDownAnnotation(node) {
			klog.V(4).Infof("Node %q is marked with ScaleDownDisabledAnnotation %q", nodeName, eligibility.ScaleDownDisabledKey)
			continue
		}
		adjustNode := node.DeepCopy()
		adjustNode.Spec.Unschedulable = true
		_, err = m.nodeInterface.Update(ctx, adjustNode, metav1.UpdateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to cordon Node %q: %w", nodeName, err))
		}
		klog.V(3).Infof("Node %q has been cordoned successfully", nodeName)
	}
	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	return nil
}

// getMachineInfo extracts the machine Key from the given node's providerID if found and checks whether it is failed or terminating and returns the MachineInfo or an error
func (m *McmManager) getMachineInfo(node *apiv1.Node) (*machineInfo, error) {
	machines, err := m.machineLister.Machines(m.namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("cannot list machines in namespace %q due to: %s", m.namespace, err)
	}

	providerID := node.Spec.ProviderID
	var machineName, machineNamespace string
	var isFailedOrTerminating bool
	for _, machine := range machines {
		machineID := strings.Split(machine.Spec.ProviderID, "/")
		nodeID := strings.Split(node.Spec.ProviderID, "/")
		// If registered, the ID will match the cloudprovider instance ID.
		// If unregistered, the ID will match the machine name.
		if machineID[len(machineID)-1] == nodeID[len(nodeID)-1] ||
			nodeID[len(nodeID)-1] == machine.Name {
			machineName = machine.Name
			machineNamespace = machine.Namespace
			isFailedOrTerminating = isMachineFailedOrTerminating(machine)
			break
		}
	}

	if machineName == "" {
		klog.V(3).Infof("No Machine found for node providerID %q", providerID)
		return nil, nil
	}
	return &machineInfo{
		Key: types.NamespacedName{
			Name:      machineName,
			Namespace: machineNamespace,
		},
		NodeName:            node.Name,
		FailedOrTerminating: isFailedOrTerminating,
	}, nil
}

func buildGenericLabels(template *nodeTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract from MCM
	if template.Architecture != nil {
		result[kubeletapis.LabelArch] = *template.Architecture
		result[apiv1.LabelArchStable] = *template.Architecture
	} else {
		result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
		result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	}
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

func buildNodeGroupFromSpec(value string, mcmManager *McmManager) (*nodeGroup, error) {
	spec, err := dynamic.SpecFromString(value, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	s := strings.Split(spec.Name, ".")
	Namespace, Name := s[0], s[1]
	ng := buildNodeGroup(mcmManager, spec.MinSize, spec.MaxSize, Namespace, Name)
	return ng, nil
}

func buildNodeGroup(mcmManager *McmManager, minSize int, maxSize int, namespace string, name string) *nodeGroup {
	return &nodeGroup{
		mcmManager:   mcmManager,
		minSize:      minSize,
		maxSize:      maxSize,
		scalingMutex: sync.Mutex{},
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// isMachineFailedOrTerminating returns true if machine is already being terminated or considered for termination by autoscaler.
// TODO: Move to MCM machineutils.IsMachineFailedOrTerminating after MCM release.
func isMachineFailedOrTerminating(machine *v1alpha1.Machine) bool {
	if !machine.GetDeletionTimestamp().IsZero() || machine.Status.CurrentStatus.Phase == v1alpha1.MachineFailed {
		return true
	}
	return false
}

// filterExtendedResources removes knownResourceNames from allResources and retains only the extendedResources.
func filterExtendedResources(allResources v1.ResourceList) (extendedResources v1.ResourceList) {
	extendedResources = allResources.DeepCopy()
	maps.DeleteFunc(extendedResources, func(name v1.ResourceName, _ resource.Quantity) bool {
		return slices.Contains(knownResourceNames, name)
	})
	return
}

// computeScaleDownData computes fresh scaleDownData for the given input MachineDeployment and the machineNamesForDeletion.
// The output scaleDownData encapsulates the scale-down amount and an updated, non-nil MachineDeployment.
func computeScaleDownData(md *v1alpha1.MachineDeployment, machineNamesForDeletion []string) (data scaleDownData) {
	forDeletionSet := sets.New(machineNamesForDeletion...)
	alreadyMarkedSet := sets.New(getMachineNamesTriggeredForDeletion(md)...)

	uniqueForDeletionSet := forDeletionSet.Difference(alreadyMarkedSet)
	toBeMarkedSet := alreadyMarkedSet.Union(forDeletionSet)

	data.RevisedToBeDeletedMachineNames = uniqueForDeletionSet
	data.RevisedScaledownAmount = uniqueForDeletionSet.Len()
	data.RevisedMachineDeployment = nil

	expectedReplicas := md.Spec.Replicas - int32(data.RevisedScaledownAmount) // #nosec G115 (CWE-190) -- RevisedScaledownAmount will not overflow the range of int32
	if expectedReplicas == md.Spec.Replicas {
		klog.Infof("MachineDeployment %q is already set to %d, no need to scale-down", md.Name, expectedReplicas)
	} else if expectedReplicas < 0 {
		klog.Errorf("Cannot delete machines in MachineDeployment %q, expected decrease in replicas: %d is more than current replicas: %d", md.Name, data.RevisedScaledownAmount, md.Spec.Replicas)
	} else {
		mdCopy := md.DeepCopy()
		if mdCopy.Annotations == nil {
			mdCopy.Annotations = make(map[string]string)
		}
		triggerDeletionAnnotValue := createMachinesTriggeredForDeletionAnnotValue(toBeMarkedSet.UnsortedList())
		if mdCopy.Annotations[machineutils.TriggerDeletionByMCM] != triggerDeletionAnnotValue {
			mdCopy.Annotations[machineutils.TriggerDeletionByMCM] = triggerDeletionAnnotValue
		}
		mdCopy.Spec.Replicas = expectedReplicas
		data.RevisedMachineDeployment = mdCopy
	}
	return
}
