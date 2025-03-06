/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"

	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	ipconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools/consts"
	npconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common/auth"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
)

const (
	maxAddTaintRetries    = 5
	maxGetNodepoolRetries = 3
	clusterId             = "clusterId"
	compartmentId         = "compartmentId"
	nodepoolTags          = "nodepoolTags"
	min                   = "min"
	max                   = "max"
)

var (
	maxRetryDeadline            time.Duration = 5 * time.Second
	conflictRetryInterval       time.Duration = 750 * time.Millisecond
	errInstanceNodePoolNotFound               = errors.New("node pool not found for instance")
)

// NodePoolManager defines the operations required for an *instance-pool based* autoscaler.
type NodePoolManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh() error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// GetNodePools returns list of registered NodePools.
	GetNodePools() []NodePool
	// GetNodePoolNodes returns NodePool nodes.
	GetNodePoolNodes(np NodePool) ([]cloudprovider.Instance, error)
	// GetNodePoolNodes returns NodePool nodes.
	GetExistingNodePoolSizeViaCompute(np NodePool) (int, error)
	// GetNodePoolForInstance returns NodePool to which the given instance belongs.
	GetNodePoolForInstance(instance ocicommon.OciRef) (NodePool, error)
	// GetNodePoolTemplateNode returns a template node for NodePool.
	GetNodePoolTemplateNode(np NodePool) (*apiv1.Node, error)
	// GetNodePoolSize gets NodePool size.
	GetNodePoolSize(np NodePool) (int, error)
	// SetNodePoolSize sets NodePool size.
	SetNodePoolSize(np NodePool, size int) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same NodePool.
	DeleteInstances(np NodePool, instances []ocicommon.OciRef) error
	// Invalidate node pool cache and refresh it
	InvalidateAndRefreshCache() error
	// Taint with ToBeDeletedByClusterAutoscaler to avoid unexpected CA restarts scheduling pods on a node intended to be deleted before restart
	TaintToPreventFurtherSchedulingOnRestart(nodes []*apiv1.Node, client kubernetes.Interface) error
}

type okeClient interface {
	GetNodePool(context.Context, oke.GetNodePoolRequest) (oke.GetNodePoolResponse, error)
	UpdateNodePool(context.Context, oke.UpdateNodePoolRequest) (oke.UpdateNodePoolResponse, error)
	DeleteNode(context.Context, oke.DeleteNodeRequest) (oke.DeleteNodeResponse, error)
	ListNodePools(ctx context.Context, request oke.ListNodePoolsRequest) (oke.ListNodePoolsResponse, error)
}

// CreateNodePoolManager creates an NodePoolManager that can manage autoscaling node pools
func CreateNodePoolManager(cloudConfigPath string, nodeGroupAutoDiscoveryList []string, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, kubeClient kubernetes.Interface) (NodePoolManager, error) {

	var err error
	var configProvider common.ConfigurationProvider

	if os.Getenv(ipconsts.OciUseWorkloadIdentityEnvVar) == "true" {
		klog.Info("using workload identity provider")
		configProvider, err = auth.OkeWorkloadIdentityConfigurationProvider()
		if err != nil {
			return nil, err
		}
	} else if os.Getenv(ipconsts.OciUseInstancePrincipalEnvVar) == "true" || os.Getenv(npconsts.OkeUseInstancePrincipalEnvVar) == "true" {
		klog.Info("using instance principal provider")
		configProvider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
	} else {
		klog.Info("using default configuration provider")
		configProvider = common.DefaultConfigProvider()
	}

	cloudConfig, err := ocicommon.CreateCloudConfig(cloudConfigPath, configProvider, npconsts.OciNodePoolResourceIdent)
	if err != nil {
		return nil, err
	}

	clientConfig := common.CustomClientConfiguration{
		RetryPolicy: ocicommon.NewRetryPolicy(),
	}

	okeClient, err := oke.NewContainerEngineClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oke client")
	}

	okeClient.SetCustomClientConfiguration(clientConfig)

	// undocumented endpoint for testing in dev
	if os.Getenv(npconsts.OkeHostOverrideEnvVar) != "" {
		okeClient.BaseClient.Host = os.Getenv(npconsts.OkeHostOverrideEnvVar)
	}

	// node pools don't need this, but set it anyway
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

	//ociShapeGetter := ocicommon.CreateShapeGetter(computeClient)
	ociShapeGetter := ocicommon.CreateShapeGetter(ocicommon.ShapeClientImpl{ComputeMgmtClient: computeMgmtClient, ComputeClient: computeClient})
	ociTagsGetter := ocicommon.CreateTagsGetter()

	registeredTaintsGetter := CreateRegisteredTaintsGetter()

	manager := &ociManagerImpl{
		cfg:                    cloudConfig,
		okeClient:              &okeClient,
		computeClient:          &computeClient,
		staticNodePools:        map[string]NodePool{},
		ociShapeGetter:         ociShapeGetter,
		ociTagsGetter:          ociTagsGetter,
		registeredTaintsGetter: registeredTaintsGetter,
		nodePoolCache:          newNodePoolCache(&okeClient),
	}

	// auto discover nodepools from compartments with nodeGroupAutoDiscovery parameter
	klog.Infof("checking node groups for autodiscovery ... ")
	for _, arg := range nodeGroupAutoDiscoveryList {
		nodeGroup, err := nodeGroupFromArg(arg)
		if err != nil {
			return nil, fmt.Errorf("unable to construct node group auto discovery from argument: %v", err)
		}
		nodeGroup.manager = manager
		nodeGroup.kubeClient = kubeClient

		manager.nodeGroups = append(manager.nodeGroups, *nodeGroup)
		autoDiscoverNodeGroups(manager, manager.okeClient, *nodeGroup)
	}

	// Contains all the specs from the args that give us the pools.
	for _, arg := range discoveryOpts.NodeGroupSpecs {
		np, err := nodePoolFromArg(arg)
		if err != nil {
			return nil, fmt.Errorf("unable to construct node pool from argument: %v", err)
		}

		np.manager = manager
		np.kubeClient = kubeClient

		manager.staticNodePools[np.Id()] = np
	}

	// wait until we have an initial full cache.
	wait.PollImmediateInfinite(
		10*time.Second,
		func() (bool, error) {
			err := manager.Refresh()
			if err != nil {
				klog.Errorf("unable to fill cache on startup. Retrying: %+v", err)
				return false, nil
			}

			return true, nil
		})

	return manager, nil
}

func autoDiscoverNodeGroups(m *ociManagerImpl, okeClient okeClient, nodeGroup nodeGroupAutoDiscovery) (bool, error) {
	var resp, reqErr = okeClient.ListNodePools(context.Background(), oke.ListNodePoolsRequest{
		ClusterId:     common.String(nodeGroup.clusterId),
		CompartmentId: common.String(nodeGroup.compartmentId),
	})
	if reqErr != nil {
		klog.Errorf("failed to fetch the nodepool list with clusterId: %s, compartmentId: %s. Error: %v", nodeGroup.clusterId, nodeGroup.compartmentId, reqErr)
		return false, reqErr
	}
	for _, nodePoolSummary := range resp.Items {
		if validateNodepoolTags(nodeGroup.tags, nodePoolSummary.FreeformTags, nodePoolSummary.DefinedTags) {
			nodepool := &nodePool{}
			nodepool.id = *nodePoolSummary.Id
			nodepool.minSize = nodeGroup.minSize
			nodepool.maxSize = nodeGroup.maxSize

			nodepool.manager = nodeGroup.manager
			nodepool.kubeClient = nodeGroup.kubeClient

			m.staticNodePools[nodepool.id] = nodepool
			klog.V(5).Infof("auto discovered nodepool in compartment : %s , nodepoolid: %s", nodeGroup.compartmentId, nodepool.id)
		} else {
			klog.Warningf("nodepool ignored as the tags do not satisfy the requirement : %s , %v, %v", *nodePoolSummary.Id, nodePoolSummary.FreeformTags, nodePoolSummary.DefinedTags)
		}
	}
	return true, nil
}

func validateNodepoolTags(nodeGroupTags map[string]string, freeFormTags map[string]string, definedTags map[string]map[string]interface{}) bool {
	if nodeGroupTags != nil {
		for tagKey, tagValue := range nodeGroupTags {
			namespacedTagKey := strings.Split(tagKey, ".")
			if len(namespacedTagKey) == 2 && tagValue != definedTags[namespacedTagKey[0]][namespacedTagKey[1]] {
				return false
			} else if len(namespacedTagKey) != 2 && tagValue != freeFormTags[tagKey] {
				return false
			}
		}
	}
	return true
}

// nodePoolFromArg parses a node group spec represented in the form of `<minSize>:<maxSize>:<ocid>` and produces a node group spec object
func nodePoolFromArg(value string) (*nodePool, error) {
	tokens := strings.SplitN(value, ":", 3)
	if len(tokens) != 3 {
		return nil, fmt.Errorf("wrong nodes configuration: %s", value)
	}

	spec := &nodePool{}
	if size, err := strconv.Atoi(tokens[0]); err == nil {
		spec.minSize = size
	} else {
		return nil, fmt.Errorf("failed to set min size: %s, expected integer", tokens[0])
	}

	if size, err := strconv.Atoi(tokens[1]); err == nil {
		spec.maxSize = size
	} else {
		return nil, fmt.Errorf("failed to set max size: %s, expected integer", tokens[1])
	}

	spec.id = tokens[2]

	klog.Infof("static node spec constructed: %+v", spec)

	return spec, nil
}

// nodeGroupFromArg parses a node group spec represented in the form of
// `clusterId:<clusterId>,compartmentId:<compartmentId>,nodepoolTags:<tagKey1>=<tagValue1>&<tagKey2>=<tagValue2>,min:<min>,max:<max>`
// and produces a node group auto discovery object
func nodeGroupFromArg(value string) (*nodeGroupAutoDiscovery, error) {
	// this regex will find the key-value pairs in any given order if separated with a colon
	regexPattern := `(?:` + compartmentId + `:(?P<` + compartmentId + `>[^,]+)`
	regexPattern = regexPattern + `|` + nodepoolTags + `:(?P<` + nodepoolTags + `>[^,]+)`
	regexPattern = regexPattern + `|` + max + `:(?P<` + max + `>[^,]+)`
	regexPattern = regexPattern + `|` + min + `:(?P<` + min + `>[^,]+)`
	regexPattern = regexPattern + `|` + clusterId + `:(?P<` + clusterId + `>[^,]+)`
	regexPattern = regexPattern + `)(?:,|$)`

	re := regexp.MustCompile(regexPattern)

	parametersMap := make(map[string]string)

	// push key-value pairs into a map
	for _, match := range re.FindAllStringSubmatch(value, -1) {
		for i, name := range re.SubexpNames() {
			if i != 0 && match[i] != "" {
				parametersMap[name] = match[i]
			}
		}
	}

	spec := &nodeGroupAutoDiscovery{}

	if parametersMap[clusterId] != "" {
		spec.clusterId = parametersMap[clusterId]
	} else {
		return nil, fmt.Errorf("failed to set %s, it is missing in node-group-auto-discovery parameter", clusterId)
	}

	if parametersMap[compartmentId] != "" {
		spec.compartmentId = parametersMap[compartmentId]
	} else {
		return nil, fmt.Errorf("failed to set %s, it is missing in node-group-auto-discovery parameter", compartmentId)
	}

	if size, err := strconv.Atoi(parametersMap[min]); err == nil {
		spec.minSize = size
	} else {
		return nil, fmt.Errorf("failed to set %s size: %s, expected integer", min, parametersMap[min])
	}

	if size, err := strconv.Atoi(parametersMap[max]); err == nil {
		spec.maxSize = size
	} else {
		return nil, fmt.Errorf("failed to set %s size: %s, expected integer", max, parametersMap[max])
	}

	if parametersMap[nodepoolTags] != "" {
		nodepoolTags := parametersMap[nodepoolTags]

		spec.tags = make(map[string]string)

		pairs := strings.Split(nodepoolTags, "&")

		for _, pair := range pairs {
			parts := strings.Split(pair, "=")
			if len(parts) == 2 {
				spec.tags[parts[0]] = parts[1]
			} else {
				return nil, fmt.Errorf("nodepoolTags should be given in tagKey=tagValue format, this is not valid: %s", pair)
			}
		}
	} else {
		return nil, fmt.Errorf("failed to set %s, it is missing in node-group-auto-discovery parameter", nodepoolTags)
	}

	klog.Infof("node group auto discovery spec constructed: %+v", spec)

	return spec, nil
}

type ociManagerImpl struct {
	cfg                    *ocicommon.CloudConfig
	okeClient              okeClient
	computeClient          *core.ComputeClient
	ociShapeGetter         ocicommon.ShapeGetter
	ociTagsGetter          ocicommon.TagsGetter
	registeredTaintsGetter RegisteredTaintsGetter
	staticNodePools        map[string]NodePool
	nodeGroups             []nodeGroupAutoDiscovery

	lastRefresh time.Time

	// caches the node pool objects received from OKE.
	// All interactions with OKE's API should go through the cache.
	nodePoolCache *nodePoolCache
}

// Refresh triggers refresh of cached resources.
func (m *ociManagerImpl) Refresh() error {
	if m.lastRefresh.Add(m.cfg.Global.RefreshInterval).After(time.Now()) {
		return nil
	}

	return m.forceRefresh()
}

// InvalidateAndRefreshCache Resets the refresh timer and refreshes
func (m *ociManagerImpl) InvalidateAndRefreshCache() error {
	// set time to 0001-01-01 00:00:00 +0000 UTC
	m.lastRefresh = time.Time{}
	return m.Refresh()
}

// TaintToPreventFurtherSchedulingOnRestart adds a taint to prevent new pods from scheduling onto the node
// this fixes a race condition where a node can be deleted, and if it's not deleted in time, the delete will retry
// and if this second delet fails, it can make the node usable again. This taint prevents this from happening
func (m *ociManagerImpl) TaintToPreventFurtherSchedulingOnRestart(nodes []*apiv1.Node, client kubernetes.Interface) error {
	for _, node := range nodes {
		taintErr := addTaint(node, client, npconsts.ToBeDeletedByClusterAutoscaler, apiv1.TaintEffectNoSchedule)
		if taintErr != nil {
			return taintErr
		}
	}
	return nil
}

func (m *ociManagerImpl) forceRefresh() error {
	// auto discover node groups
	if m.nodeGroups != nil {
		// empty previous nodepool map to do an auto discovery
		m.staticNodePools = make(map[string]NodePool)
		for _, nodeGroup := range m.nodeGroups {
			autoDiscoverNodeGroups(m, m.okeClient, nodeGroup)
		}
	}
	// rebuild nodepool cache
	err := m.nodePoolCache.rebuild(m.staticNodePools, maxGetNodepoolRetries)
	if err != nil {
		return err
	}

	m.lastRefresh = time.Now()
	klog.Infof("Refreshed NodePool list, next refresh after %v", m.lastRefresh.Add(m.cfg.Global.RefreshInterval))
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (m *ociManagerImpl) Cleanup() error {
	return nil
}

// GetNodePools returns list of registered NodePools.
func (m *ociManagerImpl) GetNodePools() []NodePool {
	var nodePools []NodePool
	for _, np := range m.staticNodePools {
		nodePoolInCache := m.nodePoolCache.cache[np.Id()]
		if nodePoolInCache != nil {
			nodePools = append(nodePools, np)
		}
	}
	return nodePools
}

// GetExistingNodePoolSizeViaCompute returns the size of nodepool that are not in a terminal state. This uses compute call to do so
// We do this to avoid any dependency on the internal caching that happens, so that we have the latest node pool state always
func (m *ociManagerImpl) GetExistingNodePoolSizeViaCompute(np NodePool) (int, error) {
	klog.V(4).Infof("getting nodes for node pool: %q", np.Id())
	nodePoolDetails, err := m.nodePoolCache.get(np.Id())
	if err != nil {
		klog.V(4).Error(err, "error fetching detailed nodepool from cache")
		return math.MaxInt32, err
	}
	request := core.ListInstancesRequest{
		CompartmentId: nodePoolDetails.CompartmentId,
		Limit:         common.Int(500),
	}

	displayNamePrefix := getDisplayNamePrefix(*nodePoolDetails.ClusterId, *nodePoolDetails.Id)
	klog.V(5).Infof("Filter used is prefix %q", displayNamePrefix)

	listInstancesFunc := func(request core.ListInstancesRequest) (core.ListInstancesResponse, error) {
		return m.computeClient.ListInstances(context.Background(), request)
	}

	var instances []cloudprovider.Instance

	for r, err := listInstancesFunc(request); ; r, err = listInstancesFunc(request) {
		if err != nil {
			klog.V(5).Error(err, "error while performing listInstancesFunc call")
			return math.MaxInt32, err
		}
		for _, item := range r.Items {
			klog.V(6).Infof("checking instance %q (instance ocid: %q) in state %q", *item.DisplayName, *item.Id, item.LifecycleState)
			if !strings.HasPrefix(*item.DisplayName, displayNamePrefix) {
				continue
			}
			switch item.LifecycleState {
			case core.InstanceLifecycleStateStopped, core.InstanceLifecycleStateTerminated:
				klog.V(4).Infof("skipping instance is in stopped/terminated state: %q", *item.Id)
			case core.InstanceLifecycleStateCreatingImage, core.InstanceLifecycleStateStarting, core.InstanceLifecycleStateProvisioning, core.InstanceLifecycleStateMoving:
				instances = append(instances, cloudprovider.Instance{
					Id: *item.Id,
					Status: &cloudprovider.InstanceStatus{
						State: cloudprovider.InstanceCreating,
					},
				})
			// in case an instance is running, it could either be installing OKE software or become a Ready node.
			// we do not know, but as we only need info if a node is stopped / terminated, we do not care
			case core.InstanceLifecycleStateRunning:
				instances = append(instances, cloudprovider.Instance{
					Id: *item.Id,
					Status: &cloudprovider.InstanceStatus{
						State: cloudprovider.InstanceRunning,
					},
				})
			case core.InstanceLifecycleStateStopping, core.InstanceLifecycleStateTerminating:
				instances = append(instances, cloudprovider.Instance{
					Id: *item.Id,
					Status: &cloudprovider.InstanceStatus{
						State: cloudprovider.InstanceDeleting,
					},
				})
			default:
				klog.Warningf("instance found in unhandled state: (%q = %v)", *item.Id, item.LifecycleState)
			}
		}

		// pagination logic
		if r.OpcNextPage != nil {
			// if there are more items in next page, fetch items from next page
			request.Page = r.OpcNextPage
		} else {
			// no more result, break the loop
			break
		}
	}

	return len(instances), nil
}

func getDisplayNamePrefix(clusterId string, nodePoolId string) string {
	shortNodePoolId := nodePoolId[len(nodePoolId)-11:]
	shortClusterId := clusterId[len(clusterId)-11:]
	return "oke" +
		"-" + shortClusterId +
		"-" + shortNodePoolId
}

// GetNodePoolNodes returns NodePool nodes that are not in a terminal state.
func (m *ociManagerImpl) GetNodePoolNodes(np NodePool) ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("getting nodes for node pool: %q", np.Id())

	nodePool, err := m.nodePoolCache.get(np.Id())
	if err != nil {
		return nil, err
	}

	var instances []cloudprovider.Instance
	for _, node := range nodePool.Nodes {

		if node.NodeError != nil {

			errorClass := cloudprovider.OtherErrorClass
			if *node.NodeError.Code == "LimitExceeded" ||
				(*node.NodeError.Code == "InternalServerError" &&
					strings.Contains(*node.NodeError.Message, "quota")) {
				errorClass = cloudprovider.OutOfResourcesErrorClass
			}

			instances = append(instances, cloudprovider.Instance{
				Id: *node.Id,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
					ErrorInfo: &cloudprovider.InstanceErrorInfo{
						ErrorClass:   errorClass,
						ErrorCode:    *node.NodeError.Code,
						ErrorMessage: *node.NodeError.Message,
					},
				},
			})

			continue
		}

		switch node.LifecycleState {
		case oke.NodeLifecycleStateDeleted:
			klog.V(4).Infof("skipping instance is in deleted state: %q", *node.Id)
		case oke.NodeLifecycleStateDeleting:
			instances = append(instances, cloudprovider.Instance{
				Id: *node.Id,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceDeleting,
				},
			})
		case oke.NodeLifecycleStateCreating, oke.NodeLifecycleStateUpdating:
			instances = append(instances, cloudprovider.Instance{
				Id: *node.Id,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			})
		case oke.NodeLifecycleStateActive:
			instances = append(instances, cloudprovider.Instance{
				Id: *node.Id,
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			})
		default:
			klog.Warningf("instance found in unhandled state: (%q = %v)", *node.Id, node.LifecycleState)
		}
	}

	return instances, nil
}

// GetNodePoolForInstance returns NodePool to which the given instance belongs.
func (m *ociManagerImpl) GetNodePoolForInstance(instance ocicommon.OciRef) (NodePool, error) {
	if instance.NodePoolID == "" {
		klog.V(4).Infof("node pool id missing from reference: %+v", instance)

		// we're looking up an unregistered node, so we can't use node pool id.
		nodePool, err := m.nodePoolCache.getByInstance(instance.InstanceID)
		if err != nil {
			return nil, err
		}

		return m.staticNodePools[*nodePool.Id], nil
	}

	np, found := m.staticNodePools[instance.NodePoolID]
	if !found {
		klog.V(4).Infof("did not find node pool for reference: %+v", instance)
		return nil, errInstanceNodePoolNotFound
	}

	return np, nil
}

// GetNodePoolTemplateNode returns a template node for NodePool.
func (m *ociManagerImpl) GetNodePoolTemplateNode(np NodePool) (*apiv1.Node, error) {

	nodePool, err := m.nodePoolCache.get(np.Id())
	if err != nil {
		return nil, err
	}

	node, err := m.buildNodeFromTemplate(nodePool)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// GetNodePoolSize gets NodePool size.
func (m *ociManagerImpl) GetNodePoolSize(np NodePool) (int, error) {
	return m.nodePoolCache.getSize(np.Id())
}

// SetNodePoolSize sets NodePool size.
func (m *ociManagerImpl) SetNodePoolSize(np NodePool, size int) error {

	err := m.nodePoolCache.setSize(np.Id(), size)
	if err != nil {
		return err
	}

	// We do not wait for the work request to finish or nodes become active on purpose. This allows
	// the autoscaler to make decisions quicker especially since the autoscaler is aware of
	// unregistered nodes in addition to registered nodes.

	return nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same NodePool.
func (m *ociManagerImpl) DeleteInstances(np NodePool, instances []ocicommon.OciRef) error {
	klog.Infof("DeleteInstances called")
	for _, instance := range instances {
		err := m.nodePoolCache.removeInstance(np.Id(), instance.InstanceID, instance.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ociManagerImpl) buildNodeFromTemplate(nodePool *oke.NodePool) (*apiv1.Node, error) {

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", "ok", 555555)

	node.ObjectMeta = metav1.ObjectMeta{
		Name:   nodeName,
		Labels: map[string]string{},
	}

	// Add all the initial node labels from the NodePool configuration to the
	// templated node.
	for _, kv := range nodePool.InitialNodeLabels {
		node.ObjectMeta.Labels[*kv.Key] = *kv.Value
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	freeformTags, err := m.ociTagsGetter.GetNodePoolFreeformTags(nodePool)
	if err != nil {
		return nil, err
	}
	ephemeralStorage, err := getEphemeralResourceRequestsInBytes(freeformTags)
	if err != nil {
		klog.Error(err)
	}
	shape, err := m.ociShapeGetter.GetNodePoolShape(nodePool, ephemeralStorage)
	if err != nil {
		return nil, err
	}

	taints, err := m.registeredTaintsGetter.Get(nodePool)
	if err != nil {
		klog.Warningf("could not extract taints from the nodepool: %s. Continuing on with empty taint list", err)
		taints = []apiv1.Taint{}
	}
	node.Spec = apiv1.NodeSpec{
		Taints: taints,
	}
	if shape.GPU > 0 {
		node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
			Key:    "nvidia.com/gpu",
			Value:  "",
			Effect: "NoSchedule",
		})
	}

	if err != nil {
		return nil, err
	}
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(shape.CPU), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(shape.MemoryInBytes), resource.DecimalSI)
	node.Status.Capacity[ipconsts.ResourceGPU] = *resource.NewQuantity(int64(shape.GPU), resource.DecimalSI)
	if ephemeralStorage != -1 {
		node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(ephemeralStorage, resource.DecimalSI)
	}

	node.Status.Allocatable = node.Status.Capacity

	availabilityDomain, err := getNodePoolAvailabilityDomain(nodePool)
	if err != nil {
		return nil, err
	}

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, ocicommon.BuildGenericLabels(*nodePool.Id, nodeName, shape.Name, availabilityDomain))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

// getNodePoolAvailabilityDomain determines the availability of the node pool.
// This breaks down if the customer specifies more than one placement configuration,
// so best practices should be a node pool per AD if customers care about it during scheduling.
// if there are more than 1AD defined, then we return the first one always.
func getNodePoolAvailabilityDomain(np *oke.NodePool) (string, error) {
	if len(np.NodeConfigDetails.PlacementConfigs) == 0 {
		return "", fmt.Errorf("node pool %q has no placement configurations", *np.Id)
	}

	if len(np.NodeConfigDetails.PlacementConfigs) > 1 {
		klog.Warningf("node pool %q has more than 1 placement config so picking first availability domain", *np.Id)
	}

	// Get the availability domain which is by default in the format of `Uocm:PHX-AD-1`
	// and remove the hash prefix.
	availabilityDomain := strings.Split(*np.NodeConfigDetails.PlacementConfigs[0].AvailabilityDomain, ":")[1]
	return availabilityDomain, nil
}

func addTaint(node *apiv1.Node, client kubernetes.Interface, taintKey string, effect apiv1.TaintEffect) error {
	retryDeadline := time.Now().Add(maxRetryDeadline)
	freshNode := node.DeepCopy()
	var err error
	refresh := false
	for i := 0; i < maxAddTaintRetries; i++ {
		if refresh {
			// Get the newest version of the node.
			freshNode, err = client.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
			if err != nil || freshNode == nil {
				klog.Warningf("Error while adding %v taint on node %v: %v", taintKey, node.Name, err)
			}
		}
		if !addTaintToSpec(freshNode, taintKey, effect) {
			if !refresh {
				// Make sure we have the latest version before skipping update.
				refresh = true
				continue
			}
			return nil
		}
		_, err = client.CoreV1().Nodes().Update(context.TODO(), freshNode, metav1.UpdateOptions{})
		if err != nil && IsConflict(err) && time.Now().Before(retryDeadline) {
			refresh = true
			time.Sleep(conflictRetryInterval)
			continue
		}

		if err != nil {
			klog.Warningf("Error while adding %v taint on node %v: %v", taintKey, node.Name, err)
			return err
		}
		klog.V(1).Infof("Successfully added %v on node %v", taintKey, node.Name)
		return nil
	}
	klog.Errorf("Could not add taint %v on node %v in %d attempts", taintKey, node.Name, maxAddTaintRetries)
	return nil
}

func addTaintToSpec(node *apiv1.Node, taintKey string, effect apiv1.TaintEffect) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == taintKey {
			klog.V(2).Infof("%v already present on node %v, taint: %v", taintKey, node.Name, taint)
			return false
		}
	}
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
		Key:    taintKey,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: effect,
	})
	return true
}

func getEphemeralResourceRequestsInBytes(tags map[string]string) (int64, error) {
	for key, value := range tags {
		if key == npconsts.EphemeralStorageSize {
			klog.V(4).Infof("ephemeral-storage size set with value : %v", value)
			value = strings.ReplaceAll(value, " ", "")
			resourceSize, err := resource.ParseQuantity(value)
			if err != nil {
				return -1, err
			}
			klog.V(4).Infof("ephemeral-storage size = %v (%v)", resourceSize.Value(), resourceSize.Format)
			return resourceSize.Value(), nil
		}
	}
	klog.V(4).Infof("ephemeral-storage size not set as part of the nodepool's freeform tags")
	return -1, nil
}

// IsConflict checks if the error is a conflict
func IsConflict(err error) bool {
	return ReasonForError(err) == metav1.StatusReasonConflict
}

// ReasonForError returns the error's reason
func ReasonForError(err error) metav1.StatusReason {
	if status := APIStatus(nil); errors.As(err, &status) {
		return status.Status().Reason
	}
	return metav1.StatusReasonUnknown
}

// APIStatus allows the conversion of errors into status objects
type APIStatus interface {
	Status() metav1.Status
}
