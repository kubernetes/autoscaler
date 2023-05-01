/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package instancepools

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools/consts"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"github.com/pkg/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common/auth"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/workrequests"
)

var (
	internalPollInterval            = 15 * time.Second
	errInstanceInstancePoolNotFound = errors.New("instance-pool not found for instance")
)

// InstancePoolManager defines the operations required for an *instance-pool based* autoscaler.
type InstancePoolManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh() error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// GetInstancePools returns list of registered InstancePools.
	GetInstancePools() []*InstancePoolNodeGroup
	// GetInstancePoolNodes returns InstancePool nodes.
	GetInstancePoolNodes(ip InstancePoolNodeGroup) ([]cloudprovider.Instance, error)
	// GetInstancePoolForInstance returns InstancePool to which the given instance belongs.
	GetInstancePoolForInstance(instance ocicommon.OciRef) (*InstancePoolNodeGroup, error)
	// GetInstancePoolTemplateNode returns a template node for InstancePool.
	GetInstancePoolTemplateNode(ip InstancePoolNodeGroup) (*apiv1.Node, error)
	// GetInstancePoolSize gets the InstancePool size.
	GetInstancePoolSize(ip InstancePoolNodeGroup) (int, error)
	// SetInstancePoolSize sets the InstancePool size.
	SetInstancePoolSize(ip InstancePoolNodeGroup, size int) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same InstancePool.
	DeleteInstances(ip InstancePoolNodeGroup, instances []ocicommon.OciRef) error
}

// InstancePoolManagerImpl is the implementation of an instance-pool based autoscaler on OCI.
type InstancePoolManagerImpl struct {
	cfg                 *ocicommon.CloudConfig
	ShapeGetter         ocicommon.ShapeGetter
	staticInstancePools map[string]*InstancePoolNodeGroup
	lastRefresh         time.Time
	// caches the instance pool and instance summary objects received from OCI.
	// All interactions with OCI's API should go through the poolCache.
	instancePoolCache *instancePoolCache
	kubeClient        kubernetes.Interface
}

// CreateInstancePoolManager constructs the InstancePoolManager object.
func CreateInstancePoolManager(cloudConfigPath string, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, kubeClient kubernetes.Interface) (InstancePoolManager, error) {

	var err error
	var configProvider common.ConfigurationProvider

	clientConfig := common.CustomClientConfiguration{
		RetryPolicy: ocicommon.NewRetryPolicy(),
	}

	// Preference to Workload Identity if set to true
	if os.Getenv(consts.OciUseWorkloadIdentityEnvVar) == "true" {
		klog.V(4).Info("using workload identity...")
		configProvider, err = auth.OkeWorkloadIdentityConfigurationProvider()
		if err != nil {
			return nil, err
		}
		// try instance principal is set to true
	} else if os.Getenv(consts.OciUseInstancePrincipalEnvVar) == "true" {
		klog.V(4).Info("using instance principals...")
		configProvider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
		// default to default provider
	} else {
		klog.Info("using default configuration provider")
		configProvider = common.DefaultConfigProvider()
	}
	providerRegion, _ := configProvider.Region()
	klog.Infof("OCI provider region: %s ", providerRegion)

	cloudConfig, err := ocicommon.CreateCloudConfig(cloudConfigPath, configProvider, consts.OciInstancePoolResourceIdent)
	if err != nil {
		return nil, err
	}

	computeMgmtClient, err := core.NewComputeManagementClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create compute management client")
	}
	computeMgmtClient.SetCustomClientConfiguration(clientConfig)

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create compute client")
	}
	computeClient.SetCustomClientConfiguration(clientConfig)

	networkClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create virtual network client")
	}
	networkClient.SetCustomClientConfiguration(clientConfig)

	workRequestClient, err := workrequests.NewWorkRequestClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create work request client")
	}
	workRequestClient.SetCustomClientConfiguration(clientConfig)

	ipManager := &InstancePoolManagerImpl{
		cfg:                 cloudConfig,
		staticInstancePools: map[string]*InstancePoolNodeGroup{},
		ShapeGetter:         ocicommon.CreateShapeGetter(ocicommon.ShapeClientImpl{ComputeMgmtClient: computeMgmtClient, ComputeClient: computeClient}),
		instancePoolCache:   newInstancePoolCache(&computeMgmtClient, &computeClient, &networkClient, &workRequestClient),
		kubeClient:          kubeClient,
	}

	// Contains all the specs from the args that give us the pools.
	for _, arg := range discoveryOpts.NodeGroupSpecs {
		ip, err := instancePoolFromArg(arg)
		if err != nil {
			return nil, fmt.Errorf("unable to construct instance pool from argument: %v", err)
		}

		ip.manager = ipManager
		ip.kubeClient = kubeClient

		ipManager.staticInstancePools[ip.Id()] = ip
	}

	// wait until we have an initial full poolCache.
	err = wait.PollImmediateInfinite(
		10*time.Second,
		func() (bool, error) {
			err := ipManager.Refresh()
			if err != nil {
				klog.Errorf("unable to fill cache on startup. Retrying: %+v", err)
				return false, nil
			}

			return true, nil
		})
	if err != nil {
		return nil, err
	}

	return ipManager, nil
}

// instancePoolFromArg parses a instancepool spec represented in the form of `<minSize>:<maxSize>:<ocid>` and produces an instance pool wrapper spec object
func instancePoolFromArg(value string) (*InstancePoolNodeGroup, error) {

	if !strings.Contains(value, consts.OciInstancePoolResourceIdent) {
		return nil, fmt.Errorf("instance pool manager does not work with resources of type: %s", value)
	}

	tokens := strings.SplitN(value, ":", 3)
	if len(tokens) != 3 || !strings.HasPrefix(tokens[2], "ocid") {
		return nil, fmt.Errorf("incorrect instance configuration: %s", value)
	}

	spec := &InstancePoolNodeGroup{}
	if size, err := strconv.Atoi(tokens[0]); err == nil {
		spec.minSize = size
	} else {
		return nil, fmt.Errorf("failed to set pool min size: %s %v", tokens[0], err)
	}

	if size, err := strconv.Atoi(tokens[1]); err == nil {
		spec.maxSize = size
	} else {
		return nil, fmt.Errorf("failed to set pool max size: %s %v", tokens[1], err)
	}

	spec.id = tokens[2]

	klog.Infof("static instance pool wrapper spec constructed: %+v", spec)

	return spec, nil
}

// Refresh triggers refresh of cached resources.
func (m *InstancePoolManagerImpl) Refresh() error {
	if m.lastRefresh.Add(m.cfg.Global.RefreshInterval).After(time.Now()) {
		return nil
	}

	return m.forceRefresh()
}

func (m *InstancePoolManagerImpl) forceRefresh() error {
	if m.cfg == nil {
		return errors.New("instance pool manager does have a required config")
	}
	m.ShapeGetter.Refresh()
	err := m.instancePoolCache.rebuild(m.staticInstancePools, *m.cfg)
	if err != nil {
		return err
	}

	m.lastRefresh = time.Now()
	klog.Infof("Refreshed instance-pool list, next refresh after %v", m.lastRefresh.Add(m.cfg.Global.RefreshInterval))
	return nil
}

func (m *InstancePoolManagerImpl) forceRefreshInstancePool(instancePoolID string) error {

	if m.cfg == nil {
		return errors.New("instance pool manager does have a required config")
	}

	if instancePoolCache, found := m.staticInstancePools[instancePoolID]; found {
		return m.instancePoolCache.rebuild(map[string]*InstancePoolNodeGroup{instancePoolID: instancePoolCache}, *m.cfg)
	}
	return errors.New("instance pool not found")
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (m *InstancePoolManagerImpl) Cleanup() error {
	return nil
}

// GetInstancePools returns list of registered InstancePools.
func (m *InstancePoolManagerImpl) GetInstancePools() []*InstancePoolNodeGroup {
	var instancePools []*InstancePoolNodeGroup
	for _, np := range m.staticInstancePools {
		instancePools = append(instancePools, np)
	}
	return instancePools
}

// GetInstancePoolNodes returns InstancePool nodes that are not in a terminal state.
func (m *InstancePoolManagerImpl) GetInstancePoolNodes(ip InstancePoolNodeGroup) ([]cloudprovider.Instance, error) {

	klog.V(4).Infof("getting (cached) instances for node pool: %q", ip.Id())

	instanceSummaries, err := m.instancePoolCache.getInstanceSummaries(ip.Id())
	if err != nil {
		return nil, err
	}

	var providerInstances []cloudprovider.Instance
	for _, instance := range *instanceSummaries {
		status := &cloudprovider.InstanceStatus{}
		switch *instance.State {
		case string(core.InstanceLifecycleStateStopped), string(core.InstanceLifecycleStateTerminated):
			klog.V(4).Infof("skipping instance is in stopped/terminated state: %q", *instance.Id)
		case string(core.InstanceLifecycleStateRunning):
			status.State = cloudprovider.InstanceRunning
		case string(core.InstanceLifecycleStateCreatingImage):
			status.State = cloudprovider.InstanceCreating
		case string(core.InstanceLifecycleStateStarting):
			status.State = cloudprovider.InstanceCreating
		case string(core.InstanceLifecycleStateMoving):
			status.State = cloudprovider.InstanceCreating
		case string(core.InstanceLifecycleStateProvisioning):
			status.State = cloudprovider.InstanceCreating
		case string(core.InstanceLifecycleStateTerminating):
			status.State = cloudprovider.InstanceDeleting
		case string(core.InstanceLifecycleStateStopping):
			status.State = cloudprovider.InstanceDeleting
		case consts.InstanceStateUnfulfilled:
			status.State = cloudprovider.InstanceCreating
			status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OutOfResourcesErrorClass,
				ErrorCode:    consts.InstanceStateUnfulfilled,
				ErrorMessage: "OCI cannot provision additional instances for this instance pool. Review quota and/or capacity.",
			}
		}

		// Instance not in a terminal or unknown state, ok to add.
		if status.State != 0 {
			providerInstances = append(providerInstances, cloudprovider.Instance{
				Id:     *instance.Id,
				Status: status,
			})
		}
	}

	return providerInstances, nil
}

// GetInstancePoolForInstance returns InstancePool to which the given instance belongs. If
// PoolID is not set on the specified OciRef, we will look for a match.
func (m *InstancePoolManagerImpl) GetInstancePoolForInstance(instanceDetails ocicommon.OciRef) (*InstancePoolNodeGroup, error) {
	if m.cfg.Global.UseNonMemberAnnotation && instanceDetails.InstancePoolID == consts.OciInstancePoolIDNonPoolMember {
		// Instance is not part of a configured pool. Return early and avoid additional API calls.
		klog.V(4).Infof(instanceDetails.Name + " is not a member of any of the specified instance pool(s) and already annotated as " +
			consts.OciInstancePoolIDNonPoolMember)
		return nil, errInstanceInstancePoolNotFound
	}

	if instanceDetails.CompartmentID == "" {
		// cfg.Global.CompartmentID would be set to tenancy OCID at runtime if compartment was not set.
		instanceDetails.CompartmentID = m.cfg.Global.CompartmentID
	}

	if ip, ok := m.staticInstancePools[instanceDetails.InstancePoolID]; ok {
		return ip, nil
	}
	// This instance is not in the cache.
	// Try to resolve the pool ID and other details, though it may not be a member of an instance-pool we manage.
	foundInstanceDetails, err := m.instancePoolCache.findInstanceByDetails(instanceDetails)
	if err != nil || foundInstanceDetails == nil || foundInstanceDetails.InstancePoolID == "" {
		if m.cfg.Global.UseNonMemberAnnotation && err == errInstanceInstancePoolNotFound {
			_ = ocicommon.AnnotateNode(m.kubeClient, instanceDetails.Name, consts.OciInstancePoolIDAnnotation, consts.OciInstancePoolIDNonPoolMember)
		}
		return nil, err
	}

	// Optionally annotate & label the node so that it does not need to be searched for in subsequent iterations.
	_ = ocicommon.AnnotateNode(m.kubeClient, foundInstanceDetails.Name, consts.OciInstanceIDAnnotation, foundInstanceDetails.InstanceID)
	_ = ocicommon.AnnotateNode(m.kubeClient, foundInstanceDetails.Name, consts.OciInstancePoolIDAnnotation, foundInstanceDetails.InstancePoolID)
	_ = ocicommon.AnnotateNode(m.kubeClient, foundInstanceDetails.Name, consts.OciAnnotationCompartmentID, foundInstanceDetails.CompartmentID)
	_ = ocicommon.LabelNode(m.kubeClient, foundInstanceDetails.Name, apiv1.LabelTopologyZone, foundInstanceDetails.AvailabilityDomain)
	_ = ocicommon.LabelNode(m.kubeClient, foundInstanceDetails.Name, apiv1.LabelFailureDomainBetaZone, foundInstanceDetails.AvailabilityDomain)
	_ = ocicommon.LabelNode(m.kubeClient, foundInstanceDetails.Name, apiv1.LabelInstanceType, foundInstanceDetails.Shape)
	_ = ocicommon.LabelNode(m.kubeClient, foundInstanceDetails.Name, apiv1.LabelInstanceTypeStable, foundInstanceDetails.Shape)
	_ = ocicommon.SetNodeProviderID(m.kubeClient, foundInstanceDetails.Name, foundInstanceDetails.InstanceID)

	return m.staticInstancePools[foundInstanceDetails.InstancePoolID], nil
}

// GetInstancePoolTemplateNode returns a template node for the InstancePool.
func (m *InstancePoolManagerImpl) GetInstancePoolTemplateNode(ip InstancePoolNodeGroup) (*apiv1.Node, error) {

	instancePool, err := m.instancePoolCache.getInstancePool(ip.Id())
	if err != nil {
		return nil, err
	}

	node, err := m.buildNodeFromTemplate(instancePool)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// GetInstancePoolSize gets the instance-pool size.
func (m *InstancePoolManagerImpl) GetInstancePoolSize(ip InstancePoolNodeGroup) (int, error) {
	return m.instancePoolCache.getSize(ip.Id())
}

// SetInstancePoolSize sets instance-pool size.
func (m *InstancePoolManagerImpl) SetInstancePoolSize(np InstancePoolNodeGroup, size int) error {
	klog.Infof("SetInstancePoolSize (%d) called on instance pool %s", size, np.Id())

	setSizeErr := m.instancePoolCache.setSize(np.Id(), size)
	klog.V(5).Infof("SetInstancePoolSize was called: refreshing instance pool cache")
	// refresh instance pool cache after update (regardless if there was an error or not)
	_ = m.forceRefreshInstancePool(np.Id())
	if setSizeErr != nil {
		return setSizeErr
	}

	// Interface says this function should wait until node group size is updated.

	// We do not wait for the work request to finish or nodes become active on purpose. This allows
	// the autoscaler to make decisions quicker especially since the autoscaler is aware of
	// unregistered nodes in addition to registered nodes.

	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same instance-pool.
func (m *InstancePoolManagerImpl) DeleteInstances(instancePool InstancePoolNodeGroup, instances []ocicommon.OciRef) error {
	klog.Infof("DeleteInstances called on instance pool %s", instancePool.Id())

	for _, instance := range instances {
		// removeInstance auto decrements instance pool size.
		detached := m.instancePoolCache.removeInstance(instancePool, instance.InstanceID)
		if !detached {
			return fmt.Errorf("could not delete instance %s from instance pool %s", instance.InstanceID, instancePool.Id())
		}
	}

	return nil
}

func (m *InstancePoolManagerImpl) buildNodeFromTemplate(instancePool *core.InstancePool) (*apiv1.Node, error) {

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", "inst", 555555)

	ocidParts := strings.Split(*instancePool.Id, ".")
	instanceIDPlaceholder := ocidParts[0] + "." + "instance" + "." + ocidParts[2] + "." + ocidParts[3] + ".tbd"

	annotations := make(map[string]string)
	annotations[consts.OciAnnotationCompartmentID] = *instancePool.CompartmentId
	annotations[consts.OciInstancePoolIDAnnotation] = *instancePool.Id
	annotations[consts.OciInstanceIDAnnotation] = instanceIDPlaceholder

	node.ObjectMeta = metav1.ObjectMeta{
		Name:        nodeName,
		Labels:      map[string]string{},
		Annotations: annotations,
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}
	shape, err := m.ShapeGetter.GetInstancePoolShape(instancePool)

	if err != nil {
		return nil, err
	}

	if shape.GPU > 0 {
		node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
			Key:    "nvidia.com/gpu",
			Value:  "",
			Effect: "NoSchedule",
		})
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(shape.CPU), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(shape.MemoryInBytes), resource.DecimalSI)
	node.Status.Capacity[consts.ResourceGPU] = *resource.NewQuantity(int64(shape.GPU), resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	availabilityDomain, err := getInstancePoolAvailabilityDomain(instancePool)
	if err != nil {
		return nil, err
	}

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, ocicommon.BuildGenericLabels(*instancePool.Id, nodeName, shape.Name, availabilityDomain))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

// getInstancePoolAvailabilityDomain determines the availability of the instance pool.
// This breaks down if the customer specifies more than one placement configuration,
// so best practices should be a node pool per AD if customers care about it during scheduling.
// if there are more than 1AD defined, then we return the first one always.
func getInstancePoolAvailabilityDomain(ip *core.InstancePool) (string, error) {
	if len(ip.PlacementConfigurations) == 0 {
		// At least one placement configuration is required for an instance pool, so we should not get here.
		return "", fmt.Errorf("instance-pool %q does not have the required placement configurations", *ip.Id)
	}

	if len(ip.PlacementConfigurations) > 1 {
		klog.Warningf("instance-pool %q has more than 1 placement config so picking first availability domain", *ip.Id)
	}

	// Get the availability domain which is by default in the format of `Uocm:PHX-AD-1`
	// and remove the hash prefix.
	availabilityDomain := strings.Split(*ip.PlacementConfigurations[0].AvailabilityDomain, ":")[1]
	return availabilityDomain, nil
}
