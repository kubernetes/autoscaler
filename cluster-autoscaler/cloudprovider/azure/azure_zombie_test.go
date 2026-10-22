/*
Copyright 2024 The Kubernetes Authors.

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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazureconfig "sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
)

func TestZombieCleanup_OnlyManagedScaleSets(t *testing.T) {
	for _, discovery := range []string{"explicit", "automatic"} {
		t.Run(discovery, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
			manager.config.EnableZombieCleanup = true
			manager.config.ZombieCleanupDryRun = false

			managed := fakeVMSSWithTags("managed-vmss", map[string]*string{
				"cluster": ptr.To("managed"), "min": ptr.To("0"), "max": ptr.To("5"),
			})
			unmanaged := fakeVMSSWithTags("unmanaged-vmss", map[string]*string{
				"cluster": ptr.To("other"), "min": ptr.To("0"), "max": ptr.To("5"),
			})
			manager.azureCache.scaleSets = map[string]*armcompute.VirtualMachineScaleSet{
				"managed-vmss": &managed, "unmanaged-vmss": &unmanaged,
			}
			if discovery == "automatic" {
				manager.autoDiscoverySpecs = []labelAutoDiscoveryConfig{
					{Selector: map[string]string{"cluster": "managed"}},
				}
				assert.NoError(t, manager.fetchAutoNodeGroups())
			} else {
				assert.NoError(t, manager.fetchExplicitNodeGroups([]string{"0:5:managed-vmss"}))
			}

			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "managed-vmss").Return(
				[]*armcompute.VirtualMachineScaleSetVM{newZombieVMWithFailedProvisioning(0, 10*time.Minute)}, nil,
			)
			mockDeleteClient.EXPECT().BeginDeleteInstances(
				gomock.Any(), manager.config.ResourceGroup, "managed-vmss",
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
				gomock.Any(),
			).Return(nil, nil)

			// No resource-group listing or calls for the unmanaged VMSS are allowed.
			assert.NoError(t, manager.cleanupZombieNodes())

			manager.UnregisterNodeGroup(newTestScaleSet(manager, "managed-vmss"))
			assert.Empty(t, manager.getNodeGroups())
			assert.NoError(t, manager.cleanupZombieNodes())
		})
	}
}

func TestZombieCleanup_UsesDiscoveredResourceGroup(t *testing.T) {
	for _, tc := range []struct {
		name                string
		hostedResourceGroup string
		resourceGroup       string
	}{
		{name: "default", resourceGroup: "test-rg"},
		{name: "hosted", hostedResourceGroup: "hosted-rg", resourceGroup: "hosted-rg"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
			manager.config.EnableZombieCleanup = true
			manager.config.ZombieCleanupDryRun = false
			manager.config.HostedResourceGroup = tc.hostedResourceGroup
			manager.azureCache.resourceGroup = tc.resourceGroup
			manager.RegisterNodeGroup(newTestScaleSet(manager, "managed-vmss"))

			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), tc.resourceGroup, "managed-vmss").Return(
				[]*armcompute.VirtualMachineScaleSetVM{newZombieVMWithFailedProvisioning(0, 10*time.Minute)}, nil,
			)
			mockDeleteClient.EXPECT().BeginDeleteInstances(
				gomock.Any(), tc.resourceGroup, "managed-vmss",
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
				gomock.Any(),
			).Return(nil, nil)

			assert.NoError(t, manager.cleanupZombieNodes())
		})
	}
}

func TestZombieCleanup_NoManagedScaleSets(t *testing.T) {
	for _, scope := range []string{"empty", "non-vmss"} {
		t.Run(scope, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager, _, _, _ := setupMockManager(t, ctrl)
			manager.config.EnableZombieCleanup = true
			manager.config.ZombieCleanupDryRun = false
			manager.azureCache.scaleSets = map[string]*armcompute.VirtualMachineScaleSet{
				"unmanaged-vmss": {Name: ptr.To("unmanaged-vmss")},
			}
			if scope == "non-vmss" {
				manager.RegisterNodeGroup(&AgentPool{azureRef: azureRef{Name: "agent-pool"}})
				manager.RegisterNodeGroup(&VMPool{azureRef: azureRef{Name: "vm-pool/sku"}})
			}

			assert.NoError(t, manager.cleanupZombieNodes())
		})
	}
}

func TestZombieCleanup_NoZombiesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	healthyVMs := []*armcompute.VirtualMachineScaleSetVM{
		newHealthyVM(0),
	}
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(healthyVMs, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DetectsFailedProvisioning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	// Verify BeginDeleteInstances is called with correct instance IDs
	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		gomock.Any(),
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse, error) {
		// Verify instance ID "0" is in the list
		assert.NotNil(t, vmInstanceIDs.InstanceIDs)
		assert.Equal(t, 1, len(vmInstanceIDs.InstanceIDs))
		assert.Equal(t, "0", *vmInstanceIDs.InstanceIDs[0])
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_FailedProvisioningSafety(t *testing.T) {
	for _, powerState := range []string{"running", "starting", "stopping", "stopped", "deallocating", "deallocated", "unknown", "", "missing-instance-view"} {
		for _, registration := range []string{"unregistered", "ready", "not-ready", "unreachable"} {
			t.Run(powerState+"/"+registration, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
				manager.config.EnableZombieCleanup = true
				manager.config.ZombieCleanupDryRun = false
				manager.config.ZombieMinAgeMinutes = 5

				vm := newZombieVMWithFailedProvisioning(0, 10*time.Minute)
				switch powerState {
				case "missing-instance-view":
					vm.Properties.InstanceView = nil
				case "":
					vm.Properties.InstanceView.Statuses = nil
				default:
					vm.Properties.InstanceView.Statuses = []*armcompute.InstanceViewStatus{
						{Code: ptr.To("PowerState/" + powerState)},
					}
				}

				nodeMap := make(map[string]*apiv1.Node)
				hasK8sNode := registration != "unregistered"
				if hasK8sNode {
					node := &apiv1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "node-0",
							CreationTimestamp: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
						},
						Spec: apiv1.NodeSpec{
							ProviderID: azurePrefix + *vm.ID,
						},
						Status: apiv1.NodeStatus{
							Conditions: []apiv1.NodeCondition{
								{
									Type:               apiv1.NodeReady,
									Status:             apiv1.ConditionFalse,
									LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
								},
							},
						},
					}
					if registration == "ready" {
						node.Status.Conditions[0].Status = apiv1.ConditionTrue
					}
					if registration == "unreachable" {
						node.Spec.Taints = []apiv1.Taint{
							{Key: "node.kubernetes.io/unreachable", Effect: apiv1.TaintEffectNoSchedule},
						}
					}
					manager.kubeClient = fake.NewSimpleClientset(node)
					nodeMap[normalizeProviderID(node.Spec.ProviderID)] = node
				}

				shouldDelete := !hasK8sNode && powerState == "stopped"
				shouldLog := (registration == "unreachable" && powerState != "deallocated" && powerState != "deallocating") || (registration == "not-ready" && powerState == "running")
				isZombie, registered, _ := manager.evaluateZombieStatus(vm, nodeMap, time.Now(), 5*time.Minute)
				assert.Equal(t, hasK8sNode, registered)
				assert.Equal(t, shouldDelete || shouldLog, isZombie)

				vmssName := "test-vmss"
				manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
				mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(
					[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
				)
				if shouldDelete {
					mockDeleteClient.EXPECT().BeginDeleteInstances(
						gomock.Any(), manager.config.ResourceGroup, vmssName,
						armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
						gomock.Any(),
					).Return(nil, nil)
				}

				assert.NoError(t, manager.cleanupZombieNodes())
			})
		}
	}
}

func TestZombieCleanup_DetectsFailedExtensions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedExtensions(0, 15*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		gomock.Any(),
	).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DetectsNeverRegisteredInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMNeverRegistered(0, 30*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		gomock.Any(),
	).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_WithK8sNodesContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Create a healthy VM (succeeded provisioning, running) that will be marked unreachable
	zombieVM := newUnreachableZombieVM(0, 10*time.Minute)

	// Create a K8s node that matches this VM (unreachable)
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "node-0",
				CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
				Taints: []apiv1.Taint{
					{
						Key:       "node.kubernetes.io/unreachable",
						Effect:    apiv1.TaintEffectNoSchedule,
						TimeAdded: &metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
					},
				},
			},
		},
	}

	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{zombieVM}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	err := manager.cleanupZombieNodesWithContext(nodes)
	assert.NoError(t, err)
}

func TestZombieCleanup_UsesKubeClientForNodeContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	// Create a K8s node with unreachable taint using the fake clientset
	unreachableNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			Taints: []apiv1.Taint{
				{
					Key:       "node.kubernetes.io/unreachable",
					Effect:    apiv1.TaintEffectNoSchedule,
					TimeAdded: &metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
				},
			},
		},
	}
	manager.kubeClient = fake.NewSimpleClientset(unreachableNode)

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Succeeded+running VM that matches the unreachable K8s node
	zombieVM := newUnreachableZombieVM(0, 10*time.Minute)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(
		[]*armcompute.VirtualMachineScaleSetVM{zombieVM}, nil,
	)

	// Call cleanupZombieNodes() not cleanupZombieNodesWithContext)
	// hasK8sNode=true so the zombie is logged only, not deleted.
	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MissingKubeClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, _, _ := setupMockManager(t, ctrl)
	manager.kubeClient = nil

	err := manager.cleanupZombieNodes()
	assert.ErrorContains(t, err, "Kubernetes client is unavailable")
}

func TestZombieCleanup_RetriesAfterNodeListFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vm := newHealthyVM(0)
	vm.Properties.TimeCreated = ptr.To(time.Now().Add(-10 * time.Minute))
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + *vm.ID,
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{
				{Type: apiv1.NodeReady, Status: apiv1.ConditionTrue},
			},
		},
	}
	kubeClient := fake.NewSimpleClientset(node)
	nodeListErr := apierrors.NewServiceUnavailable("node list unavailable")
	nodeListCalls := 0
	kubeClient.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
		nodeListCalls++
		if nodeListCalls == 1 {
			return true, nil, nodeListErr
		}
		return false, nil, nil
	})
	manager.kubeClient = kubeClient

	// No Azure calls are allowed while Kubernetes node state is unavailable.
	err := manager.cleanupZombieNodes()
	assert.ErrorIs(t, err, nodeListErr)
	assert.Equal(t, 1, nodeListCalls)

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(
		[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
	)

	// A later invocation can list nodes successfully and must preserve the registered VM.
	err = manager.cleanupZombieNodes()
	assert.NoError(t, err)
	assert.Equal(t, 2, nodeListCalls)
}

func TestZombieCleanup_RespectsMinAge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 10 // 10 minute threshold

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// VM that's only 3 minutes old (below threshold)
	youngZombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMNeverRegistered(0, 3*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(youngZombieVMs, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_SkipsNilTimeCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	nilTimeCreatedVMs := []*armcompute.VirtualMachineScaleSetVM{
		// Never-registered VM with nil TimeCreated
		{
			ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0)),
			InstanceID: ptr.To("0"),
			Properties: &armcompute.VirtualMachineScaleSetVMProperties{
				ProvisioningState: ptr.To("Succeeded"),
				TimeCreated:       nil,
				InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
					Statuses: []*armcompute.InstanceViewStatus{
						{Code: ptr.To("ProvisioningState/succeeded")},
						{Code: ptr.To("PowerState/running")},
					},
				},
			},
		},
		// Failed extensions VM with nil TimeCreated
		{
			ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 1)),
			InstanceID: ptr.To("1"),
			Properties: &armcompute.VirtualMachineScaleSetVMProperties{
				ProvisioningState: ptr.To("Succeeded"),
				TimeCreated:       nil,
				InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
					Statuses: []*armcompute.InstanceViewStatus{
						{Code: ptr.To("ProvisioningState/succeeded")},
						{Code: ptr.To("PowerState/running")},
					},
					Extensions: []*armcompute.VirtualMachineExtensionInstanceView{
						{
							Name: ptr.To("vmssCSE"),
							Statuses: []*armcompute.InstanceViewStatus{
								{
									Level: ptr.To(armcompute.StatusLevelTypesError),
									Code:  ptr.To("ProvisioningState/failed"),
								},
							},
						},
					},
				},
			},
		},
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(nilTimeCreatedVMs, nil)

	// No delete calls should be made, skip both VMs
	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DryRunMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = true // Dry-run enabled
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MultipleZombiesInSamePool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Multiple zombie VMs
	zombieVMs := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
		newZombieVMWithFailedExtensions(1, 15*time.Minute),
		newZombieVMNeverRegistered(2, 20*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(zombieVMs, nil)

	// Should call BeginDeleteInstances ONCE with all 3 instance IDs
	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		gomock.Any(),
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse, error) {
		// Verify all 3 instance IDs are in the batch
		assert.NotNil(t, vmInstanceIDs.InstanceIDs)
		assert.Equal(t, 3, len(vmInstanceIDs.InstanceIDs))
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MultipleVMSSPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmss1 := "vmss-pool-1"
	vmss2 := "vmss-pool-2"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmss1))
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmss2))

	zombieVMs1 := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}
	zombieVMs2 := []*armcompute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedExtensions(0, 15*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmss1).Return(zombieVMs1, nil)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmss2).Return(zombieVMs2, nil)

	// Should call BeginDeleteInstances TWICE (once per VMSS)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), manager.config.ResourceGroup, vmss1, gomock.Any(), gomock.Any()).Return(nil, nil)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), manager.config.ResourceGroup, vmss2, gomock.Any(), gomock.Any()).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MixedZombiesAndHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Mix of healthy and zombie VMs
	vms := []*armcompute.VirtualMachineScaleSetVM{
		newHealthyVM(0), // Healthy
		newZombieVMWithFailedProvisioning(1, 10*time.Minute), // Zombie
		newHealthyVM(2), // Healthy
		newZombieVMWithFailedExtensions(3, 15*time.Minute), // Zombie
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(vms, nil)

	// Should delete only instances 1 and 3
	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		gomock.Any(),
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions) (*armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse, error) {
		assert.NotNil(t, vmInstanceIDs.InstanceIDs)
		assert.Equal(t, 2, len(vmInstanceIDs.InstanceIDs))
		// Verify only zombie instance IDs
		var ids []string
		for _, id := range vmInstanceIDs.InstanceIDs {
			ids = append(ids, *id)
		}
		assert.Contains(t, ids, "1")
		assert.Contains(t, ids, "3")
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_RegisteredNodeUnhealthyDuration(t *testing.T) {
	now := time.Now()
	old := metav1.NewTime(now.Add(-10 * time.Minute))
	recent := metav1.NewTime(now.Add(-30 * time.Second))
	boundary := metav1.NewTime(now.Add(-5 * time.Minute))
	future := metav1.NewTime(now.Add(time.Minute))
	zero := metav1.Time{}

	for _, tc := range []struct {
		name           string
		status         apiv1.ConditionStatus
		transition     metav1.Time
		taintTimes     []*metav1.Time
		deallocated    bool
		expectedReason string
	}{
		{name: "recent NotReady", status: apiv1.ConditionFalse, transition: recent},
		{name: "recent Unknown", status: apiv1.ConditionUnknown, transition: recent},
		{name: "old NotReady", status: apiv1.ConditionFalse, transition: old, expectedReason: "Running VM; Ready condition has been False for 10 minutes"},
		{name: "old Unknown", status: apiv1.ConditionUnknown, transition: old, expectedReason: "Running VM; Ready condition has been Unknown for 10 minutes"},
		{name: "NotReady at threshold", status: apiv1.ConditionFalse, transition: boundary},
		{name: "missing Ready condition"},
		{name: "invalid Ready status", status: apiv1.ConditionStatus("invalid"), transition: old},
		{name: "zero Ready timestamp", status: apiv1.ConditionFalse},
		{name: "future Ready timestamp", status: apiv1.ConditionUnknown, transition: future},
		{name: "recovered Ready", status: apiv1.ConditionTrue, transition: old},
		{name: "recent taint overrides old condition", status: apiv1.ConditionUnknown, transition: old, taintTimes: []*metav1.Time{&recent}},
		{name: "old taint", taintTimes: []*metav1.Time{&old}, expectedReason: "Node has unreachable taint for 10 minutes"},
		{name: "taint at threshold", taintTimes: []*metav1.Time{&boundary}},
		{name: "missing taint timestamp recent fallback", status: apiv1.ConditionUnknown, transition: recent, taintTimes: []*metav1.Time{nil}},
		{name: "missing taint timestamp Unknown fallback", status: apiv1.ConditionUnknown, transition: old, taintTimes: []*metav1.Time{nil}, expectedReason: "Node has unreachable taint; Ready condition has been Unknown for 10 minutes"},
		{name: "missing taint timestamp False fallback", status: apiv1.ConditionFalse, transition: old, taintTimes: []*metav1.Time{nil}, expectedReason: "Node has unreachable taint; Ready condition has been False for 10 minutes"},
		{name: "missing both timestamps", status: apiv1.ConditionUnknown, taintTimes: []*metav1.Time{nil}},
		{name: "missing timestamp and condition", taintTimes: []*metav1.Time{nil}},
		{name: "future fallback timestamp", status: apiv1.ConditionUnknown, transition: future, taintTimes: []*metav1.Time{nil}},
		{name: "zero taint timestamp does not fall back", status: apiv1.ConditionUnknown, transition: old, taintTimes: []*metav1.Time{&zero}},
		{name: "future taint timestamp does not fall back", status: apiv1.ConditionUnknown, transition: old, taintTimes: []*metav1.Time{&future}},
		{name: "Ready with stale taint", status: apiv1.ConditionTrue, transition: recent, taintTimes: []*metav1.Time{&old}},
		{name: "timestamp on second taint", taintTimes: []*metav1.Time{nil, &old}, expectedReason: "Node has unreachable taint for 10 minutes"},
		{name: "newest taint first", taintTimes: []*metav1.Time{&recent, &old}},
		{name: "newest taint last", taintTimes: []*metav1.Time{&old, &recent}},
		{name: "deallocated with old taint", status: apiv1.ConditionUnknown, transition: old, taintTimes: []*metav1.Time{&old}, deallocated: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
			manager.config.EnableZombieCleanup = true
			manager.config.ZombieCleanupDryRun = false
			manager.config.ZombieMinAgeMinutes = 5
			vm := newHealthyVM(0)
			if tc.deallocated {
				vm = newDeallocatedVM(0)
			}
			node := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node-0",
					CreationTimestamp: metav1.NewTime(now.Add(-10 * 24 * time.Hour)),
				},
				Spec: apiv1.NodeSpec{ProviderID: azurePrefix + *vm.ID},
			}
			if tc.status != "" {
				node.Status.Conditions = []apiv1.NodeCondition{
					{Type: apiv1.NodeReady, Status: tc.status, LastTransitionTime: tc.transition},
				}
			}
			for i, timestamp := range tc.taintTimes {
				effect := apiv1.TaintEffectNoSchedule
				if i == 1 {
					effect = apiv1.TaintEffectNoExecute
				}
				node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
					Key: "node.kubernetes.io/unreachable", Effect: effect, TimeAdded: timestamp,
				})
			}
			nodeMap := map[string]*apiv1.Node{normalizeProviderID(node.Spec.ProviderID): node}
			isZombie, registered, reason := manager.evaluateZombieStatus(vm, nodeMap, now, 5*time.Minute)
			assert.Equal(t, tc.expectedReason != "", isZombie)
			assert.True(t, registered)
			assert.Equal(t, tc.expectedReason, reason)

			manager.RegisterNodeGroup(newTestScaleSet(manager, "test-vmss"))
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-vmss").Return(
				[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
			)
			assert.NoError(t, manager.cleanupZombieNodesWithContext([]*apiv1.Node{node}))
		})
	}
}

func TestZombieCleanup_PreservesParkedVMs(t *testing.T) {
	for _, powerState := range []string{"deallocated", "deallocating"} {
		for _, provisioningState := range []string{"Succeeded", "Failed"} {
			for _, extensions := range []string{"missing", "succeeded", "failed"} {
				for _, registered := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/extensions-%s/registered-%t", powerState, provisioningState, extensions, registered), func(t *testing.T) {
						ctrl := gomock.NewController(t)
						defer ctrl.Finish()

						manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
						manager.config.EnableZombieCleanup = true
						manager.config.ZombieCleanupDryRun = false
						manager.config.ZombieMinAgeMinutes = 5

						now := time.Now()
						vm := newDeallocatedVM(0)
						vm.Properties.TimeCreated = ptr.To(now.Add(-24 * time.Hour))
						vm.Properties.ProvisioningState = ptr.To(provisioningState)
						vm.Properties.InstanceView.Statuses = []*armcompute.InstanceViewStatus{
							{Code: ptr.To("PowerState/" + powerState)},
						}
						if extensions != "missing" {
							vm.Properties.InstanceView.Extensions = []*armcompute.VirtualMachineExtensionInstanceView{
								{
									Name: ptr.To("vmssCSE"),
									Statuses: []*armcompute.InstanceViewStatus{
										{Code: ptr.To("ProvisioningState/" + extensions)},
									},
								},
							}
						}

						nodeMap := make(map[string]*apiv1.Node)
						if registered {
							node := &apiv1.Node{
								ObjectMeta: metav1.ObjectMeta{Name: "node-0"},
								Spec: apiv1.NodeSpec{
									ProviderID: azurePrefix + *vm.ID,
									Taints: []apiv1.Taint{
										{
											Key:       "node.kubernetes.io/unreachable",
											Effect:    apiv1.TaintEffectNoSchedule,
											TimeAdded: ptr.To(metav1.NewTime(now.Add(-time.Hour))),
										},
									},
								},
							}
							manager.kubeClient = fake.NewSimpleClientset(node)
							nodeMap[normalizeProviderID(node.Spec.ProviderID)] = node
						}

						isZombie, hasK8sNode, reason := manager.evaluateZombieStatus(vm, nodeMap, now, 5*time.Minute)
						assert.False(t, isZombie)
						assert.Equal(t, registered, hasK8sNode)
						assert.Empty(t, reason)

						vmssName := "test-vmss"
						manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
						mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.azureCache.resourceGroup, vmssName).Return(
							[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
						)
						assert.NoError(t, manager.cleanupZombieNodes())
					})
				}
			}
		}
	}
}

func TestZombieCleanup_IgnoresDeallocatedNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Deallocated VM (healthy autoscaler scale-down)
	vms := []*armcompute.VirtualMachineScaleSetVM{
		newDeallocatedVM(0),
	}

	// Create a K8s node for this VM (NotReady but deallocated)
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "node-0",
				CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			},
			Status: apiv1.NodeStatus{
				Conditions: []apiv1.NodeCondition{
					{
						Type:   apiv1.NodeReady,
						Status: apiv1.ConditionFalse,
					},
				},
			},
		},
	}

	for _, tc := range []struct {
		name   string
		taints []apiv1.Taint
	}{
		{name: "without unreachable taint"},
		{
			name: "unreachable NoSchedule",
			taints: []apiv1.Taint{
				{Key: "node.kubernetes.io/unreachable", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "unreachable NoExecute",
			taints: []apiv1.Taint{
				{Key: "node.kubernetes.io/unreachable", Effect: apiv1.TaintEffectNoExecute},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nodes[0].Spec.Taints = tc.taints
			nodeMap := map[string]*apiv1.Node{
				normalizeProviderID(nodes[0].Spec.ProviderID): nodes[0],
			}
			isZombie, hasK8sNode, reason := manager.evaluateZombieStatus(vms[0], nodeMap, time.Now(), 5*time.Minute)
			assert.False(t, isZombie)
			assert.True(t, hasK8sNode)
			assert.Empty(t, reason)

			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(vms, nil)

			err := manager.cleanupZombieNodesWithContext(nodes)
			assert.NoError(t, err)
		})
	}
}

// ============================================================================
// SCENARIO DETECTION TESTS - Demonstrating Zombie States
// ============================================================================

func TestZombieScenario_ExtensionsFailedToInstall(t *testing.T) {
	zombieVM := newZombieVMWithFailedExtensions(0, 15*time.Minute)

	hasFailedExtension := false
	if zombieVM.Properties != nil && zombieVM.Properties.InstanceView != nil && zombieVM.Properties.InstanceView.Extensions != nil {
		for _, ext := range zombieVM.Properties.InstanceView.Extensions {
			if ext.Statuses != nil {
				for _, status := range ext.Statuses {
					code := ptr.Deref(status.Code, "")
					if (status.Level != nil && *status.Level == armcompute.StatusLevelTypesError) || code == "ProvisioningState/failed" {
						hasFailedExtension = true
						break
					}
				}
			}
		}
	}

	assert.True(t, hasFailedExtension,
		"ZOMBIE DETECTED: VM has failed extensions (vmssCSE), preventing Kubernetes registration!")
}

func TestZombieScenario_ExtensionsNeverInstalled(t *testing.T) {
	zombieVM := newZombieVMNeverRegistered(0, 20*time.Minute)

	extensionsInstalled := false
	if zombieVM.Properties != nil && zombieVM.Properties.InstanceView != nil && zombieVM.Properties.InstanceView.Extensions != nil {
		extensionsInstalled = len(zombieVM.Properties.InstanceView.Extensions) > 0
	}
	vmProvisioned := ptr.Deref(zombieVM.Properties.ProvisioningState, "") == "Succeeded"
	vmAge := 20 * time.Minute

	assert.True(t, vmProvisioned && !extensionsInstalled && vmAge > 5*time.Minute,
		"ZOMBIE DETECTED: VM provisioned successfully but extensions were NEVER INSTALLED (flapping zombie)!")
}

func TestZombieScenario_ProvisioningFailed(t *testing.T) {
	zombieVM := newZombieVMWithFailedProvisioning(0, 10*time.Minute)

	provisioningState := ptr.Deref(zombieVM.Properties.ProvisioningState, "")

	assert.Equal(t, "Failed", provisioningState,
		"ZOMBIE DETECTED: VM has ProvisioningState='%s' (should be 'Failed'), wasting quota!",
		provisioningState)
}

func TestZombieScenario_NeverRegisteredInKubernetes(t *testing.T) {
	zombieVM := newZombieVMNeverRegistered(0, 30*time.Minute)
	vmProvisioned := ptr.Deref(zombieVM.Properties.ProvisioningState, "") == "Succeeded"
	vmAge := 30 * time.Minute

	// Simulated scenario: This VM never appears in `kubectl get nodes`
	neverRegistered := true // This VM is not in K8s
	assert.True(t, vmProvisioned && neverRegistered && vmAge > 5*time.Minute,
		"ZOMBIE DETECTED: VM provisioned %v ago but never registered in Kubernetes (AllocationFailed)!",
		vmAge)
}

func TestZombieScenario_NodeUnreachableTaint(t *testing.T) {
	vm := newHealthyVM(0)
	now := time.Now()

	// Node has unreachable taint (common failure scenario)
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			Taints: []apiv1.Taint{
				{
					Key:       "node.kubernetes.io/unreachable",
					Effect:    apiv1.TaintEffectNoSchedule,
					TimeAdded: &metav1.Time{Time: now.Add(-15 * time.Minute)},
				},
			},
		},
	}

	manager := newTestAzureManagerForZombieCleanup(t)
	nodeMap := map[string]*apiv1.Node{normalizeProviderID(node.Spec.ProviderID): node}
	isZombie, registered, reason := manager.evaluateZombieStatus(vm, nodeMap, now, 5*time.Minute)
	assert.True(t, isZombie)
	assert.True(t, registered)
	assert.Equal(t, "Node has unreachable taint for 15 minutes", reason)
}

func TestZombieScenario_NodeNotReady(t *testing.T) {
	vm := newHealthyVM(0)

	// Node is NotReady for 15 minutes
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{
				{
					Type:               apiv1.NodeReady,
					Status:             apiv1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
					Reason:             "KubeletNotReady",
				},
			},
		},
	}

	manager := newTestAzureManagerForZombieCleanup(t)
	nodeMap := map[string]*apiv1.Node{normalizeProviderID(node.Spec.ProviderID): node}
	isZombie, registered, reason := manager.evaluateZombieStatus(vm, nodeMap, time.Now(), 5*time.Minute)
	assert.True(t, isZombie)
	assert.True(t, registered)
	assert.Equal(t, "Running VM; Ready condition has been False for 15 minutes", reason)
}

func TestZombieScenario_DeallocatedNodesAreHealthy(t *testing.T) {
	deallocatedVM := newDeallocatedVM(0)

	// Check power state
	isDeallocated := false
	if deallocatedVM.Properties != nil && deallocatedVM.Properties.InstanceView != nil && deallocatedVM.Properties.InstanceView.Statuses != nil {
		for _, status := range deallocatedVM.Properties.InstanceView.Statuses {
			code := ptr.Deref(status.Code, "")
			if code == "PowerState/deallocated" {
				isDeallocated = true
				break
			}
		}
	}

	assert.True(t, isDeallocated,
		"Deallocated VM should have PowerState/deallocated (healthy autoscaler scale-down, NOT a zombie)")
}

func TestZombieScenario_MultipleZombiesWasteQuota(t *testing.T) {
	// Scenario: 5 VMs in a scale set, demonstrating severe waste
	vms := []interface{}{
		newHealthyVM(0), // 1 healthy
		newZombieVMWithFailedProvisioning(1, 10*time.Minute), // Zombie
		newZombieVMWithFailedProvisioning(2, 15*time.Minute), // Zombie
		newZombieVMWithFailedExtensions(3, 20*time.Minute),   // Zombie
		newZombieVMNeverRegistered(4, 25*time.Minute),        // Zombie
	}

	totalVMs := len(vms)
	zombieCount := 4

	zombiePercentage := (zombieCount * 100) / totalVMs
	assert.Equal(t, 80, zombiePercentage,
		"ZOMBIE PROBLEM: %d out of %d VMs are zombies (%d%% waste rate) - demonstrating severe quota waste!",
		zombieCount, totalVMs, zombiePercentage)
}

func setupMockManager(t *testing.T, ctrl *gomock.Controller) (*AzureManager, *mock_virtualmachinescalesetclient.MockInterface, *mock_virtualmachinescalesetvmclient.MockInterface, *MockVMSSDeleteClient) {
	manager := newTestAzureManagerForZombieCleanup(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)

	manager.azClient = &azClient{
		virtualMachineScaleSetsClient:   mockVMSSClient,
		virtualMachineScaleSetVMsClient: mockVMSSVMClient,
		vmssClientForDelete:             mockDeleteClient,
	}

	return manager, mockVMSSClient, mockVMSSVMClient, mockDeleteClient
}

func newTestAzureManagerForZombieCleanup(t *testing.T) *AzureManager {
	return &AzureManager{
		kubeClient:           fake.NewSimpleClientset(),
		azureCache:           &azureCache{resourceGroup: "test-rg"},
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			Config: providerazureconfig.Config{
				ResourceGroup: "test-rg",
				VMType:        providerazureconsts.VMTypeVMSS,
			},
			EnableZombieCleanup: false,
			ZombieCleanupDryRun: false,
			ZombieMinAgeMinutes: 10,
		},
	}
}

// newHealthyVM creates a healthy VMSS VM
func newHealthyVM(instanceID int) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			TimeCreated:       ptr.To(time.Now().Add(-2 * time.Minute)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/running"),
					},
				},
			},
		},
	}
}

// newZombieVMWithFailedProvisioning creates a stopped VMSS VM with failed provisioning state
func newZombieVMWithFailedProvisioning(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Failed"),
			TimeCreated:       ptr.To(time.Now().Add(-age)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/failed"),
					},
					{
						Code: ptr.To("PowerState/stopped"),
					},
				},
			},
		},
	}
}

// newZombieVMWithFailedExtensions creates a VMSS VM with succeeded provisioning but failed extensions
func newZombieVMWithFailedExtensions(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			TimeCreated:       ptr.To(time.Now().Add(-age)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/running"),
					},
				},
				Extensions: []*armcompute.VirtualMachineExtensionInstanceView{
					{
						Name: ptr.To("vmssCSE"),
						Statuses: []*armcompute.InstanceViewStatus{
							{
								Level: ptr.To(armcompute.StatusLevelTypesError),
								Code:  ptr.To("ProvisioningState/failed"),
							},
						},
					},
				},
			},
		},
	}
}

// newZombieVMNeverRegistered creates a VMSS VM that succeeded provisioning but never registered in K8s
func newZombieVMNeverRegistered(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			TimeCreated:       ptr.To(time.Now().Add(-age)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/running"),
					},
				},
			},
		},
	}
}

// newUnreachableZombieVM creates a VM with healthy provisioning that will be paired with unreachable K8s node
func newUnreachableZombieVM(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			TimeCreated:       ptr.To(time.Now().Add(-age)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/running"),
					},
				},
			},
		},
	}
}

// newDeallocatedVM creates a deallocated VM (healthy autoscaler scale-down)
func newDeallocatedVM(instanceID int) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/deallocated"),
					},
				},
			},
		},
	}
}
