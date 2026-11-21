/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v8"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient/mock_virtualmachineclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/clusterstate"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaleup/orchestrator"
	coretest "sigs.k8s.io/cluster-autoscaler/pkg/core/test"
	"sigs.k8s.io/cluster-autoscaler/pkg/estimator"
	processorstest "sigs.k8s.io/cluster-autoscaler/pkg/processors/test"
	"sigs.k8s.io/cluster-autoscaler/pkg/resourcequotas"
	kubeutil "sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/taints"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testAzureClientMocks struct {
	vmssClient   *mock_virtualmachinescalesetclient.MockInterface
	vmssVMClient *mock_virtualmachinescalesetvmclient.MockInterface
	vmClient     *mock_virtualmachineclient.MockInterface
	deleteClient *MockVMSSDeleteClient
}

// testAzureProviderFixture is the Azure side of the test graph: provider and
// manager backed by overridable Azure SDK client mocks, without CAS core.
// Use this to test Azure-specific discovery, caching, and resizing without
// running core scaling decisions.
type testAzureProviderFixture struct {
	manager  *AzureManager
	provider *AzureCloudProvider
	mocks    *testAzureClientMocks
}

func newTestAzureManager(t *testing.T) *AzureManager {
	t.Helper()
	return newTestProviderWithMocks(t).manager
}

func newTestProvider(t *testing.T) *AzureCloudProvider {
	t.Helper()
	return newTestProviderWithMocks(t).provider
}

// newTestProviderWithMocks builds the Azure provider and manager with default,
// overridable Azure SDK mocks. Use it for provider-only tests that need to set
// API expectations or inspect the provider's cache.
func newTestProviderWithMocks(t *testing.T) *testAzureProviderFixture {
	t.Helper()
	// Don't need to return this or register cleanup; gomock.NewController registers cleanup with t automatically.
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)
	expectedVMSSVMs := newTestVMSSVMList(3)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), "rg").Return(expectedScaleSets, nil).AnyTimes()
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), "rg", "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	expectedVMs := newTestVMList(3)
	mockVMClient.EXPECT().List(gomock.Any(), "rg").Return(expectedVMs, nil).AnyTimes()
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	manager := &AzureManager{
		env:                  azure.PublicCloud,
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			Config: providerazureconfig.Config{
				ResourceGroup: "rg",
				VMType:        providerazureconsts.VMTypeVMSS,
				Location:      "eastus",
			},
			MaxDeploymentsCount: 2,
			Deployment:          "deployment",
			EnableForceDelete:   true,
		},
		azClient: &azClient{
			virtualMachineScaleSetsClient:   mockVMSSClient,
			virtualMachineScaleSetVMsClient: mockVMSSVMClient,
			virtualMachinesClient:           mockVMClient,
			vmssClientForDelete:             mockDeleteClient,
			deploymentClient: &DeploymentClientMock{
				FakeStore: map[string]armresources.DeploymentExtended{
					"deployment": {
						Name: ptr.To("deployment"),
						Properties: &armresources.DeploymentPropertiesExtended{
							TemplateLink: &armresources.TemplateLink{
								URI: ptr.To("http://example.com/template"),
							},
						},
					},
				},
				TemplateStore: map[string]map[string]interface{}{
					"deployment": {
						resourcesFieldName: []interface{}{
							map[string]interface{}{typeFieldName: nsgResourceType},
							map[string]interface{}{typeFieldName: rtResourceType},
						},
					},
				},
			},
		},
	}

	cache, error := newAzureCache(manager.azClient, refreshInterval, *manager.config)
	assert.NoError(t, error)

	manager.azureCache = cache
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	return &testAzureProviderFixture{
		manager: manager,
		provider: &AzureCloudProvider{
			azureManager:    manager,
			resourceLimiter: resourceLimiter,
		},
		mocks: &testAzureClientMocks{
			vmssClient:   mockVMSSClient,
			vmssVMClient: mockVMSSVMClient,
			vmClient:     mockVMClient,
			deleteClient: mockDeleteClient,
		},
	}
}

// testAutoscalerCoreFixture is the provider-independent planning side of the
// test graph: real autoscaling context, cluster state, and scale-up orchestration
// backed by fake Kubernetes clients.
type testAutoscalerCoreFixture struct {
	autoscalingCtx *ca_context.AutoscalingContext
	registry       *clusterstate.ClusterStateRegistry
	scaleUp        *orchestrator.ScaleUpOrchestrator
}

func newTestAutoscalerCore(t *testing.T, provider cloudprovider.CloudProvider) *testAutoscalerCoreFixture {
	t.Helper()

	options := config.AutoscalingOptions{
		MaxCoresTotal:                  1000,
		MaxMemoryTotal:                 1 << 40,
		MaxNodeGroupBinpackingDuration: time.Second,
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			MaxNodeProvisionTime: 15 * time.Minute,
		},
	}
	processors, templates := processorstest.NewTestProcessors(options)
	listers := kubeutil.NewListerRegistry(nil, nil, kubeutil.NewTestPodLister(nil), nil, nil, nil, nil, nil, nil)
	autoscalingCtx, err := coretest.NewScaleTestAutoscalingContext(options, fake.NewClientset(),
		listers, provider, nil, nil, templates)
	require.NoError(t, err)
	require.NoError(t, autoscalingCtx.ClusterSnapshot.SetClusterState(t.Context(), nil, nil, nil, nil))

	registry := clusterstate.NewNotifiedClusterStateRegistry(provider, autoscalingCtx.LogRecorder,
		coretest.NewBackoff(), processors.NodeGroupConfigProcessor, templates,
		clusterstate.WithScaleStateNotifier(processors.ScaleStateNotifier))
	trackerFactory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
		CustomResourcesProcessor: processors.CustomResourcesProcessor,
	})
	estimatorBuilder, err := estimator.NewEstimatorBuilder(estimator.BinpackingEstimatorName,
		estimator.NewThresholdBasedEstimationLimiter(nil), estimator.NewDecreasingPodOrderer(), nil, false)
	require.NoError(t, err)
	scaleUp := orchestrator.New()
	scaleUp.Initialize(&autoscalingCtx, processors, registry, estimatorBuilder, taints.TaintConfig{}, trackerFactory)

	return &testAutoscalerCoreFixture{
		autoscalingCtx: &autoscalingCtx,
		registry:       registry,
		scaleUp:        scaleUp,
	}
}

// testAzureAutoscalerFixture connects the Azure provider fixture to the core
// fixture for in-process scaling tests. Use this to test things end to end without a live Azure or Kubernetes API.
type testAzureAutoscalerFixture struct {
	azure *testAzureProviderFixture
	core  *testAutoscalerCoreFixture
}

// newTestAzureAutoscaler connects the Azure provider to real CAS planning and
// scale-up components. Use it for end-to-end, in-process tests of scale-up,
// resize failures, and backoff without live Azure or Kubernetes APIs.
func newTestAzureAutoscaler(t *testing.T) *testAzureAutoscalerFixture {
	t.Helper()
	fixture := newTestProviderWithMocks(t)
	core := newTestAutoscalerCore(t, fixture.provider)
	// propagate core limits to the Azure provider.
	fixture.provider.resourceLimiter = ca_context.NewResourceLimiterFromAutoscalingOptions(core.autoscalingCtx.AutoscalingOptions)
	return &testAzureAutoscalerFixture{
		azure: fixture,
		core:  core,
	}
}

func TestBuildAzureCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	m := newTestAzureManager(t)
	_, err := BuildAzureCloudProvider(m, resourceLimiter)
	assert.NoError(t, err)
}

func TestName(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, provider.Name(), "azure")
}

func TestNodeGroups(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 0)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	assert.True(t, registered)
	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager),
	)
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 2)
}

func TestHasInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := newTestProvider(t)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	provider.azureManager.azClient.agentPoolClient = mockAgentpoolclient
	provider.azureManager.azureCache.clusterName = "test-cluster"
	provider.azureManager.azureCache.clusterResourceGroup = "test-rg"
	provider.azureManager.azureCache.enableVMsAgentPool = true // enable VMs agent pool to support mixed node group types

	// Simulate node groups and instances
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)
	expectedVMsPoolVMs := newTestVMsPoolVMList(3)
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMsPoolVMs, nil).AnyTimes()
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	vmssType := armcontainerservice.AgentPoolTypeVirtualMachineScaleSets
	vmssPool := armcontainerservice.AgentPool{
		Name: ptr.To("test-asg"),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmssType,
		},
	}

	vmsPool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&vmssPool, &vmsPool)
	mockAgentpoolclient.EXPECT().NewListPager(
		provider.azureManager.azureCache.clusterResourceGroup,
		provider.azureManager.azureCache.clusterName, nil).
		Return(fakeAPListPager).AnyTimes()

	// Register node groups
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 0)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	provider.azureManager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)

	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager),
	)
	provider.azureManager.explicitlyConfigured[vmsNodeGroupName] = true
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 2)

	// Refresh cache
	provider.azureManager.forceRefresh()

	// Test HasInstance for a node from the VMSS pool
	node := newApiNode(armcompute.OrchestrationModeUniform, 0)
	hasInstance, err := provider.azureManager.azureCache.HasInstance(node.Spec.ProviderID)
	assert.True(t, hasInstance)
	assert.NoError(t, err)

	// Test HasInstance for a node from the VMs pool
	vmsPoolNode := newVMsNode(0)
	hasInstance, err = provider.azureManager.azureCache.HasInstance(vmsPoolNode.Spec.ProviderID)
	assert.True(t, hasInstance)
	assert.NoError(t, err)
}

func TestHasInstanceReportsDeletingNodesAsGone(t *testing.T) {
	testCases := []struct {
		name                        string
		strictCacheUpdates          bool
		expectDeletingBeforeRefresh bool
	}{
		{
			name:                        "non-strict cache updates",
			strictCacheUpdates:          false,
			expectDeletingBeforeRefresh: true,
		},
		{
			name:                        "strict cache updates",
			strictCacheUpdates:          true,
			expectDeletingBeforeRefresh: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := newTestAzureManager(t)
			manager.config.StrictCacheUpdates = tc.strictCacheUpdates

			expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)
			expectedVMSSVMs := newTestVMSSVMList(3)

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			manager.azClient.vmssClientForDelete = mockDeleteClient

			registered := manager.RegisterNodeGroup(newTestScaleSet(manager, "test-asg"))
			manager.explicitlyConfigured["test-asg"] = true
			assert.True(t, registered)
			assert.NoError(t, manager.forceRefresh())

			deletedNode := newApiNode(armcompute.OrchestrationModeUniform, 0)
			survivingNode := newApiNode(armcompute.OrchestrationModeUniform, 1)

			// Before the scale-down both instances are reported as present.
			hasInstance, err := manager.azureCache.HasInstance(deletedNode.Spec.ProviderID)
			assert.True(t, hasInstance)
			assert.NoError(t, err)

			scaleSet, ok := manager.azureCache.registeredNodeGroups[0].(*ScaleSet)
			assert.True(t, ok)
			assert.NoError(t, scaleSet.DeleteNodes(context.Background(), []*apiv1.Node{deletedNode}))

			// In non-strict mode, deleting state is proactive. In strict mode, it is
			// only reflected after refresh from Azure.
			hasInstance, err = manager.azureCache.HasInstance(deletedNode.Spec.ProviderID)
			if tc.expectDeletingBeforeRefresh {
				assert.False(t, hasInstance)
				assert.NoError(t, err)
			} else {
				assert.True(t, hasInstance)
				assert.NoError(t, err)
			}

			// Simulate Azure still listing the VM in the Deleting provisioning state
			// during the deletion window, then refresh the cache.
			expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
			assert.NoError(t, manager.forceRefresh())

			// The node being deleted must now be reported as gone (false, nil),
			// even though it is still tracked by the node group.
			hasInstance, err = manager.azureCache.HasInstance(deletedNode.Spec.ProviderID)
			assert.False(t, hasInstance)
			assert.NoError(t, err)
			assert.NotNil(t, manager.azureCache.getInstanceFromCache(deletedNode.Spec.ProviderID))

			// Nodes that are not being deleted continue to be reported as present.
			hasInstance, err = manager.azureCache.HasInstance(survivingNode.Spec.ProviderID)
			assert.True(t, hasInstance)
			assert.NoError(t, err)
		})
	}
}

func TestUnownedInstancesFallbackToDeletionTaint(t *testing.T) {
	// VMSS Instances that belong to a VMSS on the cluster but do not belong to a registered ASG
	// should return err unimplemented for HasInstance
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	provider := newTestProvider(t)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	// 	// Simulate VMSS instances
	unregisteredVMSSInstance := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unregistered-vmss-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "azure:///subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/unregistered-vmss-instance-id/virtualMachines/0",
		},
	}
	// Mock responses to simulate that the instance belongs to a VMSS not in any registered ASG
	expectedVMSSVMs := newTestVMSSVMList(1)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "unregistered-vmss-instance-id").Return(expectedVMSSVMs, nil).AnyTimes()

	// Call HasInstance and check the result
	hasInstance, err := provider.azureManager.azureCache.HasInstance(unregisteredVMSSInstance.Spec.ProviderID)
	assert.False(t, hasInstance)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestHasInstanceProviderIDErrorValidation(t *testing.T) {
	provider := newTestProvider(t)
	// Test case: Node with an empty ProviderID
	nodeWithoutValidProviderID := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	_, err := provider.HasInstance(context.Background(), nodeWithoutValidProviderID)
	assert.Equal(t, "ProviderID for node: test-node is empty, skipped", err.Error())

	// Test cases: Nodes with invalid ProviderID prefixes
	invalidProviderIDs := []string{
		"aazure://",
		"kubemark://",
		"kwok://",
		"incorrect!",
	}

	for _, providerID := range invalidProviderIDs {
		invalidProviderIDNode := &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Spec: apiv1.NodeSpec{
				ProviderID: providerID,
			},
		}
		_, err := provider.HasInstance(context.Background(), invalidProviderIDNode)
		assert.Equal(t, "invalid azure ProviderID prefix for node: test-node, skipped", err.Error())
	}
}

func TestMixedNodeGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := newTestProvider(t)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	provider.azureManager.azureCache.clusterName = "test-cluster"
	provider.azureManager.azureCache.clusterResourceGroup = "test-rg"
	provider.azureManager.azureCache.enableVMsAgentPool = true // enable VMs agent pool to support mixed node group types
	provider.azureManager.azClient.agentPoolClient = mockAgentpoolclient

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)
	expectedVMsPoolVMs := newTestVMsPoolVMList(3)
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMsPoolVMs, nil).AnyTimes()
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()

	vmssType := armcontainerservice.AgentPoolTypeVirtualMachineScaleSets
	vmssPool := armcontainerservice.AgentPool{
		Name: ptr.To("test-asg"),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmssType,
		},
	}

	vmsPool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&vmssPool, &vmsPool)
	mockAgentpoolclient.EXPECT().NewListPager(provider.azureManager.azureCache.clusterResourceGroup, provider.azureManager.azureCache.clusterName, nil).
		Return(fakeAPListPager).AnyTimes()

	assert.Equal(t, len(provider.NodeGroups(context.Background())), 0)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"),
	)
	provider.azureManager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)

	registered = provider.azureManager.RegisterNodeGroup(
		newTestVMsPool(provider.azureManager),
	)
	provider.azureManager.explicitlyConfigured[vmsNodeGroupName] = true
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 2)

	// refresh cache
	provider.azureManager.forceRefresh()

	// node from vmss pool
	node := newApiNode(armcompute.OrchestrationModeUniform, 0)
	group, err := provider.NodeGroupForNode(context.Background(), node)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(context.Background()), 1)
	assert.Equal(t, group.MaxSize(context.Background()), 5)

	// node from vms pool
	vmsPoolNode := newVMsNode(0)
	group, err = provider.NodeGroupForNode(context.Background(), vmsPoolNode)
	assert.NoError(t, err)
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, group.Id(), vmsNodeGroupName)
	assert.Equal(t, group.MinSize(context.Background()), 3)
	assert.Equal(t, group.MaxSize(context.Background()), 10)
}

func TestNodeGroupForNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orchestrationModes := []armcompute.OrchestrationMode{armcompute.OrchestrationModeUniform, armcompute.OrchestrationModeFlexible}

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		t.Run(fmt.Sprintf("OrchestrationMode_%v", orchMode), func(t *testing.T) {
			expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
			provider := newTestProvider(t)
			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil)
			provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
			provider.azureManager.azClient.virtualMachinesClient = mockVMClient
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()

			if orchMode == armcompute.OrchestrationModeUniform {
				mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
				mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
				provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
			} else {
				provider.azureManager.config.EnableVmssFlexNodes = true
				mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
				mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
			}

			registered := provider.azureManager.RegisterNodeGroup(
				newTestScaleSet(provider.azureManager, "test-asg"))
			provider.azureManager.explicitlyConfigured["test-asg"] = true
			assert.True(t, registered)
			assert.Equal(t, len(provider.NodeGroups(context.Background())), 1)

			node := newApiNode(orchMode, 0)
			// refresh cache
			provider.azureManager.forceRefresh()
			group, err := provider.NodeGroupForNode(context.Background(), node)
			assert.NoError(t, err)
			assert.NotNil(t, group, "Group should not be nil")
			assert.Equal(t, group.Id(), "test-asg")
			assert.Equal(t, group.MinSize(context.Background()), 1)
			assert.Equal(t, group.MaxSize(context.Background()), 5)

			hasInstance, err := provider.HasInstance(context.Background(), node)
			assert.True(t, hasInstance)
			assert.NoError(t, err)

			// test node in cluster that is not in a group managed by cluster autoscaler
			nodeNotInGroup := &apiv1.Node{
				Spec: apiv1.NodeSpec{
					ProviderID: "azure:///subscriptions/subscription/resourceGroups/test-resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/test/virtualMachines/test-instance-id-not-in-group",
				},
			}
			group, err = provider.NodeGroupForNode(context.Background(), nodeNotInGroup)
			assert.NoError(t, err)
			assert.Nil(t, group)

			hasInstance, err = provider.HasInstance(context.Background(), nodeNotInGroup)
			assert.False(t, hasInstance)
			assert.Error(t, err)
			assert.Equal(t, err, cloudprovider.ErrNotImplemented)
		})
	}
}

func TestNodeGroupForNodeWithNoProviderId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups(context.Background())), 1)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	group, err := provider.NodeGroupForNode(context.Background(), node)

	assert.NoError(t, err)
	assert.Equal(t, group, nil)
}
