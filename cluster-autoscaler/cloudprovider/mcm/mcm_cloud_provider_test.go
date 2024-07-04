// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package mcm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/apps/v1"

	machinecodes "github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	customfake "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mcm/fakeclient"
	"k8s.io/autoscaler/cluster-autoscaler/config"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	mdUpdateErrorMsg = "unable to update machine deployment"
	mcUpdateErrorMsg = "unable to update machine"
)

var (
	nodeGroup1 = "1:3:" + testNamespace + ".machinedeployment-1"
	nodeGroup2 = "0:1:" + testNamespace + ".machinedeployment-1"
	nodeGroup3 = "0:2:" + testNamespace + ".machinedeployment-2"
)

type setup struct {
	nodes                             []*corev1.Node
	machines                          []*v1alpha1.Machine
	machineSets                       []*v1alpha1.MachineSet
	machineDeployments                []*v1alpha1.MachineDeployment
	mcmDeployment                     *v1.Deployment
	machineClasses                    []*v1alpha1.MachineClass
	nodeGroups                        []string
	targetCoreFakeResourceActions     *customfake.ResourceActions
	controlMachineFakeResourceActions *customfake.ResourceActions
}

func setupEnv(setup *setup) ([]runtime.Object, []runtime.Object, []runtime.Object) {
	var controlMachineObjects []runtime.Object
	for _, o := range setup.machines {
		controlMachineObjects = append(controlMachineObjects, o)
	}
	for _, o := range setup.machineSets {
		controlMachineObjects = append(controlMachineObjects, o)
	}
	for _, o := range setup.machineDeployments {
		controlMachineObjects = append(controlMachineObjects, o)
	}
	for _, o := range setup.machineClasses {
		controlMachineObjects = append(controlMachineObjects, o)
	}

	var targetCoreObjects []runtime.Object

	for _, o := range setup.nodes {
		targetCoreObjects = append(targetCoreObjects, o)
	}

	var appsControlObjects []runtime.Object

	if setup.mcmDeployment != nil {
		appsControlObjects = append(appsControlObjects, setup.mcmDeployment)
	}

	return controlMachineObjects, targetCoreObjects, appsControlObjects
}

func TestDeleteNodes(t *testing.T) {
	type action struct {
		node *corev1.Node
	}
	type expect struct {
		machines   []*v1alpha1.Machine
		mdName     string
		mdReplicas int32
		err        error
	}
	type data struct {
		name   string
		setup  setup
		action action
		expect expect
	}
	table := []data{
		{
			"should scale down machine deployment to remove a node",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				mdName:     "machinedeployment-1",
				mdReplicas: 1,
				err:        nil,
			},
		},
		{
			"should scale down machine deployment to remove a placeholder node",
			setup{
				nodes:              nil,
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3"}, []bool{false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2},
			},
			action{node: newNode("node-1", "requested://machine-1", true)},
			expect{
				machines:   newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				mdName:     "machinedeployment-1",
				mdReplicas: 0,
				err:        nil,
			},
		},
		{
			"should not scale down a machine deployment when it is under rolling update",
			setup{
				nodes:       newNodes(2, "fakeID", []bool{true, false}),
				machines:    newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets: newMachineSets(2, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, &v1alpha1.MachineDeploymentStatus{
					Conditions: []v1alpha1.MachineDeploymentCondition{
						{Type: "Progressing"},
					},
				}, nil, nil),
				nodeGroups: []string{nodeGroup1},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   nil,
				mdName:     "machinedeployment-1",
				mdReplicas: 2,
				err:        fmt.Errorf("MachineDeployment machinedeployment-1 is under rolling update , cannot reduce replica count"),
			},
		},
		{
			"should not scale down when machine deployment update call times out",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
				controlMachineFakeResourceActions: &customfake.ResourceActions{
					MachineDeployment: customfake.Actions{
						Update: customfake.CreateFakeResponse(math.MaxInt32, mdUpdateErrorMsg, 0),
					},
				},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				mdName:     "machinedeployment-1",
				mdReplicas: 2,
				err:        fmt.Errorf("unable to scale in machine deployment machinedeployment-1, Error: %v", mdUpdateErrorMsg),
			},
		},
		{
			"should scale down when machine deployment update call fails but passes within the timeout period",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
				controlMachineFakeResourceActions: &customfake.ResourceActions{
					MachineDeployment: customfake.Actions{
						Update: customfake.CreateFakeResponse(2, mdUpdateErrorMsg, 0),
					},
				},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				mdName:     "machinedeployment-1",
				mdReplicas: 1,
				err:        nil,
			},
		},
		{
			"should not scale down a machine deployment when the corresponding machine is already in terminating state",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{true, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3"}, []bool{true}),
				mdName:     "machinedeployment-1",
				mdReplicas: 2,
				err:        nil,
			},
		},
		{
			"should not scale down a machine deployment when the corresponding machine is already in failed state",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", &v1alpha1.MachineStatus{CurrentStatus: v1alpha1.CurrentStatus{Phase: v1alpha1.MachineFailed}}, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			action{node: newNodes(1, "fakeID", []bool{false})[0]},
			expect{
				machines:   newMachines(2, "fakeID", &v1alpha1.MachineStatus{CurrentStatus: v1alpha1.CurrentStatus{Phase: v1alpha1.MachineFailed}}, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				mdName:     "machinedeployment-1",
				mdReplicas: 2,
				err:        nil,
			},
		},
		{
			"should not scale down a machine deployment below the minimum",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{true}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3"}, []bool{false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   nil,
				mdName:     "machinedeployment-1",
				mdReplicas: 1,
				err:        fmt.Errorf("min size reached, nodes will not be deleted"),
			},
		},
		{
			"no scale down of machine deployment if priority of the targeted machine cannot be updated to 1",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-1"),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
				controlMachineFakeResourceActions: &customfake.ResourceActions{
					Machine: customfake.Actions{
						Update: customfake.CreateFakeResponse(math.MaxInt32, mcUpdateErrorMsg, 0),
					},
				},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   nil,
				mdName:     "machinedeployment-1",
				mdReplicas: 2,
				err:        fmt.Errorf("could not prioritize machine machine-1 for deletion, aborting scale in of machine deployment, Error: %s", mcUpdateErrorMsg),
			},
		},
		{
			"should not scale down machine deployment if the node belongs to another machine deployment",
			setup{
				nodes:              newNodes(2, "fakeID", []bool{true, false}),
				machines:           newMachines(2, "fakeID", nil, "machinedeployment-2", "machineset-1", []string{"3", "3"}, []bool{false, false}),
				machineSets:        newMachineSets(1, "machinedeployment-2"),
				machineDeployments: newMachineDeployments(2, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2, nodeGroup3},
			},
			action{node: newNodes(1, "fakeID", []bool{true})[0]},
			expect{
				machines:   nil,
				mdName:     "machinedeployment-2",
				mdReplicas: 2,
				err:        fmt.Errorf("node-1 belongs to a different machinedeployment than machinedeployment-1"),
			},
		},
	}

	for _, entry := range table {
		entry := entry // have a shallow copy of the entry for parallelization of tests
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			stop := make(chan struct{})
			defer close(stop)
			controlMachineObjects, targetCoreObjects, _ := setupEnv(&entry.setup)
			m, trackers, hasSyncedCacheFns := createMcmManager(t, stop, testNamespace, nil, controlMachineObjects, targetCoreObjects, nil)
			defer trackers.Stop()
			waitForCacheSync(t, stop, hasSyncedCacheFns)

			if entry.setup.targetCoreFakeResourceActions != nil {
				trackers.TargetCore.SetFailAtFakeResourceActions(entry.setup.targetCoreFakeResourceActions)
			}
			if entry.setup.controlMachineFakeResourceActions != nil {
				trackers.ControlMachine.SetFailAtFakeResourceActions(entry.setup.controlMachineFakeResourceActions)
			}

			md, err := buildMachineDeploymentFromSpec(entry.setup.nodeGroups[0], m)
			g.Expect(err).To(BeNil())

			err = md.DeleteNodes([]*corev1.Node{entry.action.node})

			if entry.expect.err != nil {
				g.Expect(err).To(Equal(entry.expect.err))
			} else {
				g.Expect(err).To(BeNil())
			}

			machineDeployment, err := m.machineClient.MachineDeployments(m.namespace).Get(context.TODO(), entry.expect.mdName, metav1.GetOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(machineDeployment.Spec.Replicas).To(BeNumerically("==", entry.expect.mdReplicas))

			machines, err := m.machineClient.Machines(m.namespace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
					MatchLabels: map[string]string{"name": md.Name},
				}),
			})

			for _, machine := range machines.Items {
				flag := false
				for _, entryMachineItem := range entry.expect.machines {
					if entryMachineItem.Name == machine.Name {
						g.Expect(machine.Annotations[priorityAnnotationKey]).To(Equal(entryMachineItem.Annotations[priorityAnnotationKey]))
						flag = true
						break
					}
				}
				if !flag {
					g.Expect(machine.Annotations[priorityAnnotationKey]).To(Equal("3"))
				}
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	type expect struct {
		machines []*v1alpha1.Machine
		err      error
	}
	type data struct {
		name   string
		setup  setup
		expect expect
	}
	table := []data{
		{

			"should return an error if MCM has zero available replicas",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{false}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2},
				mcmDeployment:      newMCMDeployment(0),
			},
			expect{
				err: fmt.Errorf("machine-controller-manager is offline. Cluster autoscaler operations would be suspended."),
			},
		},
		{

			"should return an error if MCM deployment is not found",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{false}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2},
			},
			expect{
				err: fmt.Errorf("failed to get machine-controller-manager deployment: deployment.apps \"machine-controller-manager\" not found"),
			},
		},
		{

			"should reset priority of a machine with node without ToBeDeletedTaint to 3",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{false}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2},
				mcmDeployment:      newMCMDeployment(1),
			},
			expect{
				machines: newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"3"}, []bool{false}),
				err:      nil,
			},
		},
		{
			"should not reset priority of a machine to 3 if the node has ToBeDeleted taint",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{true}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				nodeGroups:         []string{nodeGroup2},
				mcmDeployment:      newMCMDeployment(1),
			},
			expect{
				machines: newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				err:      nil,
			},
		},
		{
			"priority reset of machine fails",
			setup{
				nodes:              newNodes(1, "fakeID", []bool{false}),
				machines:           newMachines(1, "fakeID", nil, "machinedeployment-1", "machineset-1", []string{"1"}, []bool{false}),
				machineDeployments: newMachineDeployments(1, 1, nil, nil, nil),
				controlMachineFakeResourceActions: &customfake.ResourceActions{
					Machine: customfake.Actions{
						Update: customfake.CreateFakeResponse(math.MaxInt32, mcUpdateErrorMsg, 0),
					},
				},
				nodeGroups:    []string{nodeGroup2},
				mcmDeployment: newMCMDeployment(1),
			},
			expect{
				machines: []*v1alpha1.Machine{newMachine("machine-1", "fakeID-1", nil, "machinedeployment-1", "machineset-1", "1", false, true)},
				err:      errors.Join(fmt.Errorf("could not reset priority annotation on machine machine-1, Error: %v", mcUpdateErrorMsg)),
			},
		},
	}
	for _, entry := range table {
		entry := entry
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			stop := make(chan struct{})
			defer close(stop)
			controlMachineObjects, targetCoreObjects, appsControlObjects := setupEnv(&entry.setup)
			m, trackers, hasSyncedCacheFns := createMcmManager(t, stop, testNamespace, entry.setup.nodeGroups, controlMachineObjects, targetCoreObjects, appsControlObjects)
			defer trackers.Stop()
			waitForCacheSync(t, stop, hasSyncedCacheFns)

			if entry.setup.targetCoreFakeResourceActions != nil {
				trackers.TargetCore.SetFailAtFakeResourceActions(entry.setup.targetCoreFakeResourceActions)
			}
			if entry.setup.controlMachineFakeResourceActions != nil {
				trackers.ControlMachine.SetFailAtFakeResourceActions(entry.setup.controlMachineFakeResourceActions)
			}
			mcmCloudProvider, err := BuildMcmCloudProvider(m, nil)
			g.Expect(err).To(BeNil())
			err = mcmCloudProvider.Refresh()
			if entry.expect.err != nil {
				g.Expect(err).To(Equal(entry.expect.err))
			} else {
				g.Expect(err).To(BeNil())
			}
			for _, mc := range entry.expect.machines {
				machine, err := m.machineClient.Machines(m.namespace).Get(context.TODO(), mc.Name, metav1.GetOptions{})
				g.Expect(err).To(BeNil())
				g.Expect(mc.Annotations[priorityAnnotationKey]).To(Equal(machine.Annotations[priorityAnnotationKey]))
			}
		})
	}
}

// Different kinds of cases possible and expected cloudprovider.Instance returned for them
// (mobj, mobjPid, nodeobj)   				    -> instance(nodeobj.pid,_)
// (mobj, mobjPid, _)         				    -> instance("requested://<machine-name>",_)
// (mobj, _,_)                 				    -> instance("requested://<machine-name>",_)
// (mobj, _,_) with quota error 				-> instance("requested://<machine-name>",status{'creating',{'outofResourcesClass','ResourceExhausted','<message>'}})
// (mobj, _,_) with invalid credentials error   -> instance("requested://<machine-name>",_)

// Example machine.status.lastOperation for a `ResourceExhausted` error
//
//		lastOperation: {
//			type: Creating
//			state: Failed
//			errorCode: ResourceExhausted
//			description: "Cloud provider message - machine codes error: code = [ResourceExhausted] message = [Create machine "shoot--ddci--cbc-sys-tests03-pool-c32m256-3b-z1-575b9-hlvj6" failed: The following errors occurred: [{QUOTA_EXCEEDED  Quota 'N2_CPUS' exceeded.  Limit: 6000.0 in region europe-west3. [] []}]]."
//		}
//	}
func TestNodes(t *testing.T) {
	const (
		outOfQuotaMachineStatusErrorDescription         = "Cloud provider message - machine codes error: code = [ResourceExhausted] message = [Create machine \"machine-with-vm-create-error-out-of-quota\" failed: The following errors occurred: [{QUOTA_EXCEEDED  Quota 'N2_CPUS' exceeded.  Limit: 6000.0 in region europe-west3. [] []}]]"
		invalidCredentialsMachineStatusErrorDescription = "Cloud provider message - machine codes error: code = [Internal] message = [user is not authorized to perform this action]"
	)
	type expectationPerInstance struct {
		providerID           string
		instanceState        cloudprovider.InstanceState
		instanceErrorClass   cloudprovider.InstanceErrorClass
		instanceErrorCode    string
		instanceErrorMessage string
	}
	type expect struct {
		expectationPerInstanceList []expectationPerInstance
	}
	type data struct {
		name   string
		setup  setup
		expect expect
	}
	table := []data{
		{
			"Correct instances should be returned for machine objects under the machinedeployment",
			setup{
				nodes: []*corev1.Node{newNode("node-1", "fakeID-1", false)},
				machines: func() []*v1alpha1.Machine {
					allMachines := make([]*v1alpha1.Machine, 0, 5)
					allMachines = append(allMachines, newMachine("machine-with-registered-node", "fakeID-1", nil, "machinedeployment-1", "", "", false, true))
					allMachines = append(allMachines, newMachine("machine-with-vm-but-no-node", "fakeID-2", nil, "machinedeployment-1", "", "", false, false))
					allMachines = append(allMachines, newMachine("machine-with-vm-creating", "", nil, "machinedeployment-1", "", "", false, false))
					allMachines = append(allMachines, newMachine("machine-with-vm-create-error-out-of-quota", "", &v1alpha1.MachineStatus{LastOperation: v1alpha1.LastOperation{Type: v1alpha1.MachineOperationCreate, State: v1alpha1.MachineStateFailed, ErrorCode: machinecodes.ResourceExhausted.String(), Description: outOfQuotaMachineStatusErrorDescription}}, "machinedeployment-1", "", "", false, false))
					allMachines = append(allMachines, newMachine("machine-with-vm-create-error-invalid-credentials", "", &v1alpha1.MachineStatus{LastOperation: v1alpha1.LastOperation{Type: v1alpha1.MachineOperationCreate, State: v1alpha1.MachineStateFailed, ErrorCode: machinecodes.Internal.String(), Description: invalidCredentialsMachineStatusErrorDescription}}, "machinedeployment-1", "", "", false, false))
					return allMachines
				}(),
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			expect{
				expectationPerInstanceList: []expectationPerInstance{
					{"fakeID-1", cloudprovider.InstanceState(-1), cloudprovider.InstanceErrorClass(-1), "", ""},
					{placeholderInstanceIDForMachineObj("machine-with-vm-but-no-node"), cloudprovider.InstanceState(-1), cloudprovider.InstanceErrorClass(-1), "", ""},
					{placeholderInstanceIDForMachineObj("machine-with-vm-creating"), cloudprovider.InstanceState(-1), cloudprovider.InstanceErrorClass(-1), "", ""},
					{placeholderInstanceIDForMachineObj("machine-with-vm-create-error-out-of-quota"), cloudprovider.InstanceCreating, cloudprovider.OutOfResourcesErrorClass, machinecodes.ResourceExhausted.String(), outOfQuotaMachineStatusErrorDescription},
					// invalid credentials error is mapped to Internal code as it can't be fixed by trying another zone
					{placeholderInstanceIDForMachineObj("machine-with-vm-create-error-invalid-credentials"), cloudprovider.InstanceState(-1), cloudprovider.InstanceErrorClass(-1), "", ""},
				},
			},
		},
	}

	for _, entry := range table {
		entry := entry // have a shallow copy of the entry for parallelization of tests
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			stop := make(chan struct{})
			defer close(stop)
			controlMachineObjects, targetCoreObjects, _ := setupEnv(&entry.setup)
			m, trackers, hasSyncedCacheFns := createMcmManager(t, stop, testNamespace, nil, controlMachineObjects, targetCoreObjects, nil)
			defer trackers.Stop()
			waitForCacheSync(t, stop, hasSyncedCacheFns)

			if entry.setup.targetCoreFakeResourceActions != nil {
				trackers.TargetCore.SetFailAtFakeResourceActions(entry.setup.targetCoreFakeResourceActions)
			}
			if entry.setup.controlMachineFakeResourceActions != nil {
				trackers.ControlMachine.SetFailAtFakeResourceActions(entry.setup.controlMachineFakeResourceActions)
			}

			md, err := buildMachineDeploymentFromSpec(entry.setup.nodeGroups[0], m)
			g.Expect(err).To(BeNil())

			returnedInstances, err := md.Nodes()
			g.Expect(err).To(BeNil())
			g.Expect(len(returnedInstances)).To(BeNumerically("==", len(entry.expect.expectationPerInstanceList)))

			for _, expectedInstance := range entry.expect.expectationPerInstanceList {
				found := false
				for _, gotInstance := range returnedInstances {
					g.Expect(gotInstance.Id).ToNot(BeEmpty())
					if expectedInstance.providerID == gotInstance.Id {
						if !strings.Contains(gotInstance.Id, "requested://") {
							// must be a machine obj whose node is registered (ready or notReady)
							g.Expect(gotInstance.Status).To(BeNil())
						} else {
							if int(expectedInstance.instanceState) != -1 {
								g.Expect(gotInstance.Status).ToNot(BeNil())
								g.Expect(gotInstance.Status.State).To(Equal(expectedInstance.instanceState))
							}
							if int(expectedInstance.instanceErrorClass) != -1 || expectedInstance.instanceErrorCode != "" || expectedInstance.instanceErrorMessage != "" {
								g.Expect(gotInstance.Status.ErrorInfo).ToNot(BeNil())
								g.Expect(gotInstance.Status.ErrorInfo.ErrorClass).To(Equal(expectedInstance.instanceErrorClass))
								g.Expect(gotInstance.Status.ErrorInfo.ErrorCode).To(Equal(expectedInstance.instanceErrorCode))
								g.Expect(gotInstance.Status.ErrorInfo.ErrorMessage).To(Equal(expectedInstance.instanceErrorMessage))
							}
						}
						found = true
						break
					}
				}
				g.Expect(found).To(BeTrue())
			}
		})
	}
}

func TestGetOptions(t *testing.T) {
	ngAutoScalingOpDefaults := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.5,
		ScaleDownGpuUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:            1 * time.Minute,
		ScaleDownUnreadyTime:             1 * time.Minute,
		MaxNodeProvisionTime:             1 * time.Minute,
		IgnoreDaemonSetsUtilization:      true,
		ZeroOrMaxNodeScaling:             true,
	}

	type expect struct {
		ngOptions *config.NodeGroupAutoscalingOptions
		err       error
	}
	type data struct {
		name   string
		setup  setup
		expect expect
	}
	table := []data{
		{
			"should throw error if machinedeployment cannot be found",
			setup{
				nodeGroups: []string{nodeGroup1},
			},
			expect{
				err: fmt.Errorf("unable to fetch MachineDeployment object machinedeployment-1, Error: machinedeployment.machine.sapcloud.io \"machinedeployment-1\" not found"),
			},
		},
		{
			"should return default nodegroupautoscalingoptions if none are provided",
			setup{
				machineDeployments: newMachineDeployments(1, 2, nil, nil, nil),
				nodeGroups:         []string{nodeGroup1},
			},
			expect{
				ngOptions: &ngAutoScalingOpDefaults,
				err:       nil,
			},
		},
		{
			"should return nodegroupautoscalingoptions with values from mcd if all annotations are present",
			setup{
				machineDeployments: newMachineDeployments(
					1,
					2,
					nil,
					map[string]string{
						ScaleDownUtilizationThresholdAnnotation:    "0.7",
						ScaleDownGpuUtilizationThresholdAnnotation: "0.7",
						ScaleDownUnneededTimeAnnotation:            "5m",
						ScaleDownUnreadyTimeAnnotation:             "5m",
						MaxNodeProvisionTimeAnnotation:             "5m",
					},
					nil,
				),
				nodeGroups: []string{nodeGroup1},
			},
			expect{
				ngOptions: &config.NodeGroupAutoscalingOptions{
					ScaleDownUtilizationThreshold:    0.7,
					ScaleDownGpuUtilizationThreshold: 0.7,
					ScaleDownUnneededTime:            5 * time.Minute,
					ScaleDownUnreadyTime:             5 * time.Minute,
					MaxNodeProvisionTime:             5 * time.Minute,
					IgnoreDaemonSetsUtilization:      ngAutoScalingOpDefaults.IgnoreDaemonSetsUtilization,
					ZeroOrMaxNodeScaling:             ngAutoScalingOpDefaults.ZeroOrMaxNodeScaling,
				},
				err: nil,
			},
		},
		{
			"should return nodegroupautoscalingoptions with annotations values from mcd and remaining defaults",
			setup{
				machineDeployments: newMachineDeployments(
					1,
					2,
					nil,
					map[string]string{
						ScaleDownUtilizationThresholdAnnotation: "0.7",
						ScaleDownUnneededTimeAnnotation:         "5m",
						MaxNodeProvisionTimeAnnotation:          "2m",
					},
					nil,
				),
				nodeGroups: []string{nodeGroup1},
			},
			expect{
				ngOptions: &config.NodeGroupAutoscalingOptions{
					ScaleDownUtilizationThreshold:    0.7,
					ScaleDownGpuUtilizationThreshold: 0.5,
					ScaleDownUnneededTime:            5 * time.Minute,
					ScaleDownUnreadyTime:             1 * time.Minute,
					MaxNodeProvisionTime:             2 * time.Minute,
					IgnoreDaemonSetsUtilization:      ngAutoScalingOpDefaults.IgnoreDaemonSetsUtilization,
					ZeroOrMaxNodeScaling:             ngAutoScalingOpDefaults.ZeroOrMaxNodeScaling,
				},
				err: nil,
			},
		},
	}

	for _, entry := range table {
		entry := entry // have a shallow copy of the entry for parallelization of tests
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			stop := make(chan struct{})
			defer close(stop)
			controlMachineObjects, targetCoreObjects, _ := setupEnv(&entry.setup)
			m, trackers, hasSyncedCacheFns := createMcmManager(t, stop, testNamespace, nil, controlMachineObjects, targetCoreObjects, nil)
			defer trackers.Stop()
			waitForCacheSync(t, stop, hasSyncedCacheFns)

			md, err := buildMachineDeploymentFromSpec(entry.setup.nodeGroups[0], m)
			g.Expect(err).To(BeNil())

			options, err := md.GetOptions(ngAutoScalingOpDefaults)

			if entry.expect.err != nil {
				g.Expect(err).To(Equal(entry.expect.err))
				g.Expect(options).To(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(*options).To(HaveField("ScaleDownUtilizationThreshold", entry.expect.ngOptions.ScaleDownUtilizationThreshold))
				g.Expect(*options).To(HaveField("ScaleDownGpuUtilizationThreshold", entry.expect.ngOptions.ScaleDownGpuUtilizationThreshold))
				g.Expect(*options).To(HaveField("ScaleDownUnneededTime", entry.expect.ngOptions.ScaleDownUnneededTime))
				g.Expect(*options).To(HaveField("ScaleDownUnreadyTime", entry.expect.ngOptions.ScaleDownUnreadyTime))
				g.Expect(*options).To(HaveField("MaxNodeProvisionTime", entry.expect.ngOptions.MaxNodeProvisionTime))
			}
		})
	}
}
