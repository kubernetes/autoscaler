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

func TestStaleInstanceCleanup_OnlyManagedScaleSets(t *testing.T) {
	for _, discovery := range []string{"explicit", "automatic"} {
		t.Run(discovery, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
			manager.config.EnableStaleInstanceCleanup = true
			manager.config.StaleInstanceCleanupDryRun = false

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
				[]*armcompute.VirtualMachineScaleSetVM{newStaleVMWithFailedProvisioning(0, 10*time.Minute)}, nil,
			)
			mockDeleteClient.EXPECT().BeginDeleteInstances(
				gomock.Any(), manager.config.ResourceGroup, "managed-vmss",
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
				gomock.Any(),
			).Return(nil, nil)

			// No resource-group listing or calls for the unmanaged VMSS are allowed.
			assert.NoError(t, manager.cleanupStaleInstances())

			manager.UnregisterNodeGroup(newTestScaleSet(manager, "managed-vmss"))
			assert.Empty(t, manager.getNodeGroups())
			assert.NoError(t, manager.cleanupStaleInstances())
		})
	}
}

func TestStaleInstanceCleanup_UsesDiscoveredResourceGroup(t *testing.T) {
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
			manager.config.EnableStaleInstanceCleanup = true
			manager.config.StaleInstanceCleanupDryRun = false
			manager.config.HostedResourceGroup = tc.hostedResourceGroup
			manager.azureCache.resourceGroup = tc.resourceGroup
			manager.RegisterNodeGroup(newTestScaleSet(manager, "managed-vmss"))

			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), tc.resourceGroup, "managed-vmss").Return(
				[]*armcompute.VirtualMachineScaleSetVM{newStaleVMWithFailedProvisioning(0, 10*time.Minute)}, nil,
			)
			mockDeleteClient.EXPECT().BeginDeleteInstances(
				gomock.Any(), tc.resourceGroup, "managed-vmss",
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
				gomock.Any(),
			).Return(nil, nil)

			assert.NoError(t, manager.cleanupStaleInstances())
		})
	}
}

func TestStaleInstanceCleanup_NoManagedScaleSets(t *testing.T) {
	for _, scope := range []string{"empty", "non-vmss"} {
		t.Run(scope, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager, _, _, _ := setupMockManager(t, ctrl)
			manager.config.EnableStaleInstanceCleanup = true
			manager.config.StaleInstanceCleanupDryRun = false
			manager.azureCache.scaleSets = map[string]*armcompute.VirtualMachineScaleSet{
				"unmanaged-vmss": {Name: ptr.To("unmanaged-vmss")},
			}
			if scope == "non-vmss" {
				manager.RegisterNodeGroup(&AgentPool{azureRef: azureRef{Name: "agent-pool"}})
				manager.RegisterNodeGroup(&VMPool{azureRef: azureRef{Name: "vm-pool/sku"}})
			}

			assert.NoError(t, manager.cleanupStaleInstances())
		})
	}
}

func TestStaleInstanceCleanup_NoStaleInstancesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	healthyVMs := []*armcompute.VirtualMachineScaleSetVM{
		newHealthyVM(0),
	}
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(healthyVMs, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_DetectsFailedProvisioning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

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

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_FailedProvisioningSafety(t *testing.T) {
	for _, powerState := range []string{"running", "starting", "stopping", "stopped", "deallocating", "deallocated", "unknown", "", "missing-instance-view"} {
		for _, registration := range []string{"unregistered", "ready", "not-ready", "unreachable"} {
			t.Run(powerState+"/"+registration, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
				manager.config.EnableStaleInstanceCleanup = true
				manager.config.StaleInstanceCleanupDryRun = false
				manager.config.StaleInstanceMinAgeMinutes = 5

				vm := newStaleVMWithFailedProvisioning(0, 10*time.Minute)
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
				isStaleInstance, registered, _ := manager.evaluateStaleInstanceStatus(vm, nodeMap, time.Now(), 5*time.Minute)
				assert.Equal(t, hasK8sNode, registered)
				assert.Equal(t, shouldDelete || shouldLog, isStaleInstance)

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

				assert.NoError(t, manager.cleanupStaleInstances())
			})
		}
	}
}

func TestStaleInstanceCleanup_DetectsFailedExtensions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedExtensions(0, 15*time.Minute),
	}
	seedStaleInstanceObservations(manager, vmssName, staleInstanceVMs)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		gomock.Any(),
	).Return(nil, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_DetectsNeverRegisteredInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMNeverRegistered(0, 30*time.Minute),
	}
	seedStaleInstanceObservations(manager, vmssName, staleInstanceVMs)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

	mockDeleteClient.EXPECT().BeginDeleteInstances(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		gomock.Any(),
	).Return(nil, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_WithK8sNodesContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Create a healthy VM (succeeded provisioning, running) that will be marked unreachable
	staleInstanceVM := newUnreachableVM(0, 10*time.Minute)

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

	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{staleInstanceVM}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

	err := manager.cleanupStaleInstancesWithContext(nodes, time.Now())
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_UsesKubeClientForNodeContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

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
	staleInstanceVM := newUnreachableVM(0, 10*time.Minute)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(
		[]*armcompute.VirtualMachineScaleSetVM{staleInstanceVM}, nil,
	)

	// Call cleanupStaleInstances() not cleanupStaleInstancesWithContext)
	// hasK8sNode=true so the unhealthy instance is logged only, not deleted.
	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_MissingKubeClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, _, _ := setupMockManager(t, ctrl)
	manager.kubeClient = nil

	err := manager.cleanupStaleInstances()
	assert.ErrorContains(t, err, "Kubernetes client is unavailable")
}

func TestStaleInstanceCleanup_RetriesAfterNodeListFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

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
	err := manager.cleanupStaleInstances()
	assert.ErrorIs(t, err, nodeListErr)
	assert.Equal(t, 1, nodeListCalls)

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(
		[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
	)

	// A later invocation can list nodes successfully and must preserve the registered VM.
	err = manager.cleanupStaleInstances()
	assert.NoError(t, err)
	assert.Equal(t, 2, nodeListCalls)
}

func TestStaleInstanceCleanup_RespectsMinAge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 10 // 10 minute threshold

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// VM that's only 3 minutes old (below threshold)
	youngStaleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMNeverRegistered(0, 3*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(youngStaleInstanceVMs, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_SkipsNilTimeCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

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
	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_DryRunMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = true // Dry-run enabled
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_MultipleStaleInstancesInSamePool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Multiple stale VMs
	staleInstanceVMs := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedProvisioning(0, 10*time.Minute),
		newStaleVMWithFailedExtensions(1, 15*time.Minute),
		newStaleVMNeverRegistered(2, 20*time.Minute),
	}
	seedStaleInstanceObservations(manager, vmssName, staleInstanceVMs)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(staleInstanceVMs, nil)

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

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_MultipleVMSSPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmss1 := "vmss-pool-1"
	vmss2 := "vmss-pool-2"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmss1))
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmss2))

	staleInstanceVMs1 := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedProvisioning(0, 10*time.Minute),
	}
	staleInstanceVMs2 := []*armcompute.VirtualMachineScaleSetVM{
		newStaleVMWithFailedExtensions(0, 15*time.Minute),
	}
	seedStaleInstanceObservations(manager, vmss2, staleInstanceVMs2)

	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmss1).Return(staleInstanceVMs1, nil)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmss2).Return(staleInstanceVMs2, nil)

	// Should call BeginDeleteInstances TWICE (once per VMSS)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), manager.config.ResourceGroup, vmss1, gomock.Any(), gomock.Any()).Return(nil, nil)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), manager.config.ResourceGroup, vmss2, gomock.Any(), gomock.Any()).Return(nil, nil)

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_MixedStaleInstancesAndHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, mockDeleteClient := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

	vmssName := "test-vmss"
	manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

	// Mix of healthy and stale VMs
	vms := []*armcompute.VirtualMachineScaleSetVM{
		newHealthyVM(0), // Healthy
		newStaleVMWithFailedProvisioning(1, 10*time.Minute), // Stale instance
		newHealthyVM(2), // Healthy
		newStaleVMWithFailedExtensions(3, 15*time.Minute), // Stale instance
	}
	seedStaleInstanceObservations(manager, vmssName, vms)

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
		// Verify only stale instance IDs
		var ids []string
		for _, id := range vmInstanceIDs.InstanceIDs {
			ids = append(ids, *id)
		}
		assert.Contains(t, ids, "1")
		assert.Contains(t, ids, "3")
		return nil, nil
	})

	err := manager.cleanupStaleInstances()
	assert.NoError(t, err)
}

func TestStaleInstanceCleanup_RegisteredNodeUnhealthyDuration(t *testing.T) {
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
			manager.config.EnableStaleInstanceCleanup = true
			manager.config.StaleInstanceCleanupDryRun = false
			manager.config.StaleInstanceMinAgeMinutes = 5
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
			isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nodeMap, now, 5*time.Minute)
			assert.Equal(t, tc.expectedReason != "", isStaleInstance)
			assert.True(t, registered)
			assert.Equal(t, tc.expectedReason, reason)

			manager.RegisterNodeGroup(newTestScaleSet(manager, "test-vmss"))
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-vmss").Return(
				[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
			)
			assert.NoError(t, manager.cleanupStaleInstancesWithContext([]*apiv1.Node{node}, now))
		})
	}
}

func TestStaleInstanceCleanup_RequiresKnownPowerState(t *testing.T) {
	for _, powerState := range []string{"missing-instance-view", "missing-statuses", "nil-status", "missing-code", "no-power-state", "", "unknown", "unrecognized", "starting", "stopping", "running", "stopped"} {
		for _, extensions := range []string{"missing", "succeeded", "failed"} {
			t.Run(powerState+"/extensions-"+extensions, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				manager, _, vmClient, deleteClient := setupMockManager(t, ctrl)
				manager.config.EnableStaleInstanceCleanup = true
				manager.config.StaleInstanceCleanupDryRun = false
				manager.config.StaleInstanceMinAgeMinutes = 5
				vmssName := "test-vmss"
				manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))

				now := time.Now()
				vm := newStaleVMNeverRegistered(0, 24*time.Hour)
				view := vm.Properties.InstanceView
				if extensions != "missing" {
					view.Extensions = []*armcompute.VirtualMachineExtensionInstanceView{
						{
							Name: ptr.To("vmssCSE"),
							Statuses: []*armcompute.InstanceViewStatus{
								{Code: ptr.To("ProvisioningState/" + extensions)},
							},
						},
					}
				}
				switch powerState {
				case "missing-instance-view":
					vm.Properties.InstanceView = nil
				case "missing-statuses":
					view.Statuses = nil
				case "nil-status":
					view.Statuses = []*armcompute.InstanceViewStatus{nil}
				case "missing-code":
					view.Statuses = []*armcompute.InstanceViewStatus{{}}
				case "no-power-state":
					view.Statuses = []*armcompute.InstanceViewStatus{{Code: ptr.To("ProvisioningState/succeeded")}}
				default:
					view.Statuses = []*armcompute.InstanceViewStatus{{Code: ptr.To("PowerState/" + powerState)}}
				}

				shouldDelete := powerState == "running" || powerState == "stopped"
				isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, now, 5*time.Minute)
				assert.Equal(t, shouldDelete, isStaleInstance)
				assert.False(t, registered)
				if shouldDelete {
					assert.NotEmpty(t, reason)
				} else {
					assert.Empty(t, reason)
				}

				vmClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.azureCache.resourceGroup, vmssName).Return(
					[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
				).Times(2)
				assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now))
				if shouldDelete {
					deleteClient.EXPECT().BeginDeleteInstances(
						gomock.Any(), manager.azureCache.resourceGroup, vmssName,
						armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
						gomock.Any(),
					).Return(nil, nil)
				}
				assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now.Add(6*time.Minute)))
			})
		}
	}
}

func TestStaleInstanceCleanup_RegistrationGracePeriod(t *testing.T) {
	for _, interruption := range []string{
		"registered", "starting", "stopping", "deallocated", "deallocating", "unknown",
		"provisioning", "missing-timestamp", "missing-identity", "missing-vm",
		"vm-list-error", "node-list-error", "missing-client", "deregistered", "recreated",
	} {
		t.Run(interruption, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			manager, _, vmClient, deleteClient := setupMockManager(t, ctrl)
			manager.config.EnableStaleInstanceCleanup = true
			manager.config.StaleInstanceCleanupDryRun = false
			manager.config.StaleInstanceMinAgeMinutes = 5
			group := newTestScaleSet(manager, "test-vmss")
			manager.RegisterNodeGroup(group)
			now := time.Now()
			vm := newStaleVMNeverRegistered(0, 24*time.Hour)
			vms := []*armcompute.VirtualMachineScaleSetVM{vm}
			var listErr error
			vmClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.azureCache.resourceGroup, group.Name).DoAndReturn(
				func(context.Context, string, string) ([]*armcompute.VirtualMachineScaleSetVM, error) {
					return vms, listErr
				},
			).AnyTimes()

			assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now))
			assert.Len(t, manager.staleInstanceObservations, 1)
			assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now.Add(4*time.Minute)))

			var nodes []*apiv1.Node
			originalProperties := vm.Properties
			originalID := vm.ID
			interrupted := newStaleVMNeverRegistered(0, 24*time.Hour)
			interrupted.Properties.TimeCreated = originalProperties.TimeCreated
			vm.Properties = interrupted.Properties
			switch interruption {
			case "registered":
				nodes = []*apiv1.Node{{Spec: apiv1.NodeSpec{ProviderID: azurePrefix + *vm.ID}}}
			case "starting", "stopping", "deallocated", "deallocating", "unknown":
				vm.Properties.InstanceView.Statuses = []*armcompute.InstanceViewStatus{
					{Code: ptr.To("PowerState/" + interruption)},
				}
			case "provisioning":
				vm.Properties.ProvisioningState = ptr.To("Creating")
			case "missing-timestamp":
				vm.Properties.TimeCreated = nil
			case "missing-identity":
				vm.ID = nil
			case "missing-vm":
				vms = nil
			case "vm-list-error":
				listErr = fmt.Errorf("VM list unavailable")
			case "node-list-error":
				client := fake.NewSimpleClientset()
				client.PrependReactor("list", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, apierrors.NewServiceUnavailable("node list unavailable")
				})
				manager.kubeClient = client
			case "missing-client":
				manager.kubeClient = nil
			case "deregistered":
				manager.UnregisterNodeGroup(group)
			case "recreated":
				vm.Properties.TimeCreated = ptr.To(now.Add(-2 * time.Hour))
			}

			if interruption == "node-list-error" || interruption == "missing-client" {
				assert.Error(t, manager.cleanupStaleInstances())
			} else {
				assert.NoError(t, manager.cleanupStaleInstancesWithContext(nodes, now.Add(4*time.Minute)))
			}
			if interruption == "recreated" {
				assert.Len(t, manager.staleInstanceObservations, 1)
				for key, firstSeen := range manager.staleInstanceObservations {
					assert.Equal(t, vm.Properties.TimeCreated.UTC(), key.createdAt)
					assert.Equal(t, now.Add(4*time.Minute), firstSeen, "recreated VM must not inherit the old timer")
				}
			} else {
				assert.Empty(t, manager.staleInstanceObservations, "interrupted observations must not retain the old timer")
			}

			vm.Properties = originalProperties
			vm.ID = originalID
			vms = []*armcompute.VirtualMachineScaleSetVM{vm}
			listErr = nil
			manager.kubeClient = fake.NewSimpleClientset()
			manager.RegisterNodeGroup(group)
			assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now.Add(6*time.Minute)))
			assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now.Add(11*time.Minute)))

			deleteClient.EXPECT().BeginDeleteInstances(
				gomock.Any(), manager.azureCache.resourceGroup, group.Name,
				armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{ptr.To("0")}},
				gomock.Any(),
			).Return(nil, nil)
			assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, now.Add(11*time.Minute+time.Second)))
		})
	}
}

func TestStaleInstanceCleanup_GracePeriodDryRunAndRestart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	now := time.Now()
	vm := newStaleVMNeverRegistered(0, 24*time.Hour)
	for restart := 0; restart < 2; restart++ {
		manager, _, vmClient, _ := setupMockManager(t, ctrl)
		manager.config.StaleInstanceCleanupDryRun = true
		manager.config.StaleInstanceMinAgeMinutes = 5
		manager.RegisterNodeGroup(newTestScaleSet(manager, "test-vmss"))
		vmClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.azureCache.resourceGroup, "test-vmss").Return(
			[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
		).Times(2)

		start := now.Add(time.Duration(restart) * time.Hour)
		assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, start))
		assert.Len(t, manager.staleInstanceObservations, 1)
		for _, firstSeen := range manager.staleInstanceObservations {
			assert.Equal(t, start, firstSeen, "a new manager must grant a fresh grace period")
		}
		assert.NoError(t, manager.cleanupStaleInstancesWithContext(nil, start.Add(6*time.Minute)))
	}
}

func TestStaleInstanceCleanup_PreservesParkedVMs(t *testing.T) {
	for _, powerState := range []string{"deallocated", "deallocating"} {
		for _, provisioningState := range []string{"Succeeded", "Failed"} {
			for _, extensions := range []string{"missing", "succeeded", "failed"} {
				for _, registered := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/extensions-%s/registered-%t", powerState, provisioningState, extensions, registered), func(t *testing.T) {
						ctrl := gomock.NewController(t)
						defer ctrl.Finish()

						manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
						manager.config.EnableStaleInstanceCleanup = true
						manager.config.StaleInstanceCleanupDryRun = false
						manager.config.StaleInstanceMinAgeMinutes = 5

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

						isStaleInstance, hasK8sNode, reason := manager.evaluateStaleInstanceStatus(vm, nodeMap, now, 5*time.Minute)
						assert.False(t, isStaleInstance)
						assert.Equal(t, registered, hasK8sNode)
						assert.Empty(t, reason)

						vmssName := "test-vmss"
						manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
						mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.azureCache.resourceGroup, vmssName).Return(
							[]*armcompute.VirtualMachineScaleSetVM{vm}, nil,
						)
						assert.NoError(t, manager.cleanupStaleInstances())
					})
				}
			}
		}
	}
}

func TestStaleInstanceCleanup_IgnoresDeallocatedNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, _, mockVMSSVMClient, _ := setupMockManager(t, ctrl)
	manager.config.EnableStaleInstanceCleanup = true
	manager.config.StaleInstanceCleanupDryRun = false
	manager.config.StaleInstanceMinAgeMinutes = 5

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
			isStaleInstance, hasK8sNode, reason := manager.evaluateStaleInstanceStatus(vms[0], nodeMap, time.Now(), 5*time.Minute)
			assert.False(t, isStaleInstance)
			assert.True(t, hasK8sNode)
			assert.Empty(t, reason)

			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(vms, nil)

			err := manager.cleanupStaleInstancesWithContext(nodes, time.Now())
			assert.NoError(t, err)
		})
	}
}

// ============================================================================
// SCENARIO DETECTION TESTS - Demonstrating Stale Instance States
// ============================================================================

func TestStaleInstanceScenario_ExtensionsFailedToInstall(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vm := newStaleVMWithFailedExtensions(0, 15*time.Minute)
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, time.Now(), 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.False(t, registered)
	assert.Contains(t, reason, "Extensions FAILED")
}

func TestStaleInstanceScenario_ExtensionsNeverInstalled(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vm := newStaleVMNeverRegistered(0, 20*time.Minute)
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, time.Now(), 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.False(t, registered)
	assert.Contains(t, reason, "Extensions NOT INSTALLED")
}

func TestStaleInstanceScenario_ProvisioningFailed(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vm := newStaleVMWithFailedProvisioning(0, 10*time.Minute)
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, time.Now(), 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.False(t, registered)
	assert.Equal(t, "Provisioning FAILED and VM is stopped", reason)
}

func TestStaleInstanceScenario_NeverRegisteredInKubernetes(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vm := newStaleVMNeverRegistered(0, 30*time.Minute)
	vm.Properties.InstanceView.Extensions = []*armcompute.VirtualMachineExtensionInstanceView{
		{
			Name: ptr.To("vmssCSE"),
			Statuses: []*armcompute.InstanceViewStatus{
				{Code: ptr.To("ProvisioningState/succeeded")},
			},
		},
	}
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, time.Now(), 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.False(t, registered)
	assert.Contains(t, reason, "Never registered in K8s")
}

func TestStaleInstanceScenario_NodeUnreachableTaint(t *testing.T) {
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

	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	nodeMap := map[string]*apiv1.Node{normalizeProviderID(node.Spec.ProviderID): node}
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nodeMap, now, 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.True(t, registered)
	assert.Equal(t, "Node has unreachable taint for 15 minutes", reason)
}

func TestStaleInstanceScenario_NodeNotReady(t *testing.T) {
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

	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	nodeMap := map[string]*apiv1.Node{normalizeProviderID(node.Spec.ProviderID): node}
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nodeMap, time.Now(), 5*time.Minute)
	assert.True(t, isStaleInstance)
	assert.True(t, registered)
	assert.Equal(t, "Running VM; Ready condition has been False for 15 minutes", reason)
}

func TestStaleInstanceScenario_DeallocatedNodesAreHealthy(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vm := newDeallocatedVM(0)
	vm.Properties.TimeCreated = ptr.To(time.Now().Add(-24 * time.Hour))
	isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, time.Now(), 5*time.Minute)
	assert.False(t, isStaleInstance)
	assert.False(t, registered)
	assert.Empty(t, reason)
}

func TestStaleInstanceScenario_MultipleStaleInstancesWasteQuota(t *testing.T) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)
	vms := []*armcompute.VirtualMachineScaleSetVM{
		newHealthyVM(0), // 1 healthy
		newStaleVMWithFailedProvisioning(1, 10*time.Minute), // Stale instance
		newStaleVMWithFailedProvisioning(2, 15*time.Minute), // Stale instance
		newStaleVMWithFailedExtensions(3, 20*time.Minute),   // Stale instance
		newStaleVMNeverRegistered(4, 25*time.Minute),        // Stale instance
	}

	now := time.Now()
	var staleInstanceIDs []string
	for _, vm := range vms {
		isStaleInstance, registered, reason := manager.evaluateStaleInstanceStatus(vm, nil, now, 5*time.Minute)
		assert.False(t, registered)
		if isStaleInstance {
			assert.NotEmpty(t, reason)
			staleInstanceIDs = append(staleInstanceIDs, *vm.InstanceID)
		} else {
			assert.Empty(t, reason)
		}
	}
	assert.ElementsMatch(t, []string{"1", "2", "3", "4"}, staleInstanceIDs)
}

// seedStaleInstanceObservations models a completed grace period for tests of deletion batching.
func seedStaleInstanceObservations(manager *AzureManager, group string, vms []*armcompute.VirtualMachineScaleSetVM) {
	if manager.staleInstanceObservations == nil {
		manager.staleInstanceObservations = make(map[staleInstanceObservationKey]time.Time)
	}
	for _, vm := range vms {
		if vm.Properties.TimeCreated == nil {
			continue
		}
		manager.staleInstanceObservations[staleInstanceObservationKey{
			nodeGroupID: group,
			providerID:  normalizeProviderID(*vm.ID),
			createdAt:   vm.Properties.TimeCreated.UTC(),
		}] = time.Now().Add(-time.Duration(manager.config.StaleInstanceMinAgeMinutes+1) * time.Minute)
	}
}

func setupMockManager(t *testing.T, ctrl *gomock.Controller) (*AzureManager, *mock_virtualmachinescalesetclient.MockInterface, *mock_virtualmachinescalesetvmclient.MockInterface, *MockVMSSDeleteClient) {
	manager := newTestAzureManagerForStaleInstanceCleanup(t)

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

func newTestAzureManagerForStaleInstanceCleanup(t *testing.T) *AzureManager {
	return &AzureManager{
		kubeClient:           fake.NewSimpleClientset(),
		azureCache:           &azureCache{resourceGroup: "test-rg"},
		explicitlyConfigured: make(map[string]bool),
		config: &Config{
			Config: providerazureconfig.Config{
				ResourceGroup: "test-rg",
				VMType:        providerazureconsts.VMTypeVMSS,
			},
			EnableStaleInstanceCleanup: false,
			StaleInstanceCleanupDryRun: false,
			StaleInstanceMinAgeMinutes: 10,
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

// newStaleVMWithFailedProvisioning creates a stopped VMSS VM with failed provisioning state
func newStaleVMWithFailedProvisioning(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
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

// newStaleVMWithFailedExtensions creates a VMSS VM with succeeded provisioning but failed extensions
func newStaleVMWithFailedExtensions(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
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

// newStaleVMNeverRegistered creates a VMSS VM that succeeded provisioning but never registered in K8s
func newStaleVMNeverRegistered(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
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

// newUnreachableVM creates a VM with healthy provisioning that will be paired with unreachable K8s node
func newUnreachableVM(instanceID int, age time.Duration) *armcompute.VirtualMachineScaleSetVM {
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
