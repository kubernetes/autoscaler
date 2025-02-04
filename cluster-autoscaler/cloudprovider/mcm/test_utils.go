// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package mcm

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/tools/cache"

	machineinternal "github.com/gardener/machine-controller-manager/pkg/apis/machine"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	faketyped "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/typed/machine/v1alpha1/fake"
	machineinformers "github.com/gardener/machine-controller-manager/pkg/client/informers/externalversions"
	mcmcache "github.com/gardener/machine-controller-manager/pkg/util/provider/cache"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	customfake "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mcm/fakeclient"
	appsv1informers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers"
)

var (
	testNamespace  = "test-namespace"
	testTaintValue = fmt.Sprint(time.Now().Unix())
)

func newMachineDeployments(
	machineDeploymentCount int,
	replicas int32,
	statusTemplate *v1alpha1.MachineDeploymentStatus,
	annotations map[string]string,
	labels map[string]string,
) []*v1alpha1.MachineDeployment {
	machineDeployments := make([]*v1alpha1.MachineDeployment, machineDeploymentCount)
	for i := 0; i < machineDeploymentCount; i++ {
		machineDeployment := &v1alpha1.MachineDeployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "machine.sapcloud.io",
				Kind:       "MachineDeployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("machinedeployment-%d", i+1),
				Namespace: testNamespace,
				Labels:    labels,
			},
			Spec: v1alpha1.MachineDeploymentSpec{
				Replicas: replicas,
			},
		}
		if statusTemplate != nil {
			machineDeployment.Status = *statusTemplate.DeepCopy()
		}
		if annotations != nil {
			machineDeployment.Annotations = annotations
		}
		machineDeployments[i] = machineDeployment
	}
	return machineDeployments
}

func newMachineSets(
	machineSetCount int,
	mdName string,
) []*v1alpha1.MachineSet {

	machineSets := make([]*v1alpha1.MachineSet, machineSetCount)
	for i := 0; i < machineSetCount; i++ {
		ms := &v1alpha1.MachineSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "machine.sapcloud.io",
				Kind:       "MachineSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("machineset-%d", i+1),
				Namespace:       testNamespace,
				OwnerReferences: []metav1.OwnerReference{{Name: mdName}},
			},
		}
		machineSets[i] = ms
	}
	return machineSets
}

func newMachine(
	name string,
	providerId string,
	statusTemplate *v1alpha1.MachineStatus,
	mdName, msName string,
	priorityAnnotationValue string,
	setNodeLabel bool,
) *v1alpha1.Machine {
	m := newMachines(1, providerId, statusTemplate, mdName, msName, []string{priorityAnnotationValue})[0]
	m.Name = name
	m.Spec.ProviderID = providerId
	if !setNodeLabel {
		delete(m.Labels, "node")
	}
	return m
}

func generateNames(prefix string, count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("%s-%d", prefix, i+1)
	}
	return names
}

func newMachines(
	machineCount int,
	providerIdGenerateName string,
	statusTemplate *v1alpha1.MachineStatus,
	mdName, msName string,
	priorityAnnotationValues []string,
) []*v1alpha1.Machine {
	machines := make([]*v1alpha1.Machine, machineCount)
	machineNames := generateNames("machine", machineCount)
	nodeNames := generateNames("node", machineCount)
	currentTime := metav1.Now()

	for i := 0; i < machineCount; i++ {
		m := &v1alpha1.Machine{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "machine.sapcloud.io",
				Kind:       "Machine",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineNames[i],
				Namespace: testNamespace,
				OwnerReferences: []metav1.OwnerReference{
					{Name: msName},
				},
				Labels:            map[string]string{machineDeploymentNameLabel: mdName},
				CreationTimestamp: metav1.Now(),
			},
		}

		if providerIdGenerateName != "" {
			m.Spec = v1alpha1.MachineSpec{ProviderID: fmt.Sprintf("%s/i%d", providerIdGenerateName, i+1)}
		}

		m.Labels["node"] = nodeNames[i]
		if statusTemplate != nil {
			m.Status = *newMachineStatus(statusTemplate)
			if m.Status.CurrentStatus.Phase == v1alpha1.MachineTerminating {
				m.DeletionTimestamp = &currentTime
			}
		}
		machines[i] = m
	}
	return machines
}

func newNode(
	nodeName,
	providerId string,
) *corev1.Node {
	node := newNodes(1, providerId)[0]
	clone := node.DeepCopy()
	clone.Name = nodeName
	clone.Spec.ProviderID = providerId
	return clone
}

func newNodes(
	nodeCount int,
	providerIdGenerateName string,
) []*corev1.Node {
	nodes := make([]*corev1.Node, nodeCount)
	nodeNames := generateNames("node", nodeCount)
	for i := 0; i < nodeCount; i++ {
		node := &corev1.Node{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "appsv1",
				Kind:       "Node",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeNames[i],
			},
			Spec: corev1.NodeSpec{
				ProviderID: fmt.Sprintf("%s/i%d", providerIdGenerateName, i+1),
			},
		}

		nodes[i] = node
	}
	return nodes
}

func newMachineStatus(statusTemplate *v1alpha1.MachineStatus) *v1alpha1.MachineStatus {
	if statusTemplate == nil {
		return &v1alpha1.MachineStatus{}
	}

	return statusTemplate.DeepCopy()
}

func newMCMDeployment(availableReplicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller-manager",
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: availableReplicas,
		},
	}
}

func createMcmManager(
	t *testing.T,
	stop <-chan struct{},
	namespace string,
	nodeGroups []string, controlMachineObjects, targetCoreObjects, controlAppsObjects []runtime.Object,
) (*McmManager, *customfake.FakeObjectTrackers, []cache.InformerSynced) {
	g := gomega.NewWithT(t)
	fakeControlMachineClient, controlMachineObjectTracker := customfake.NewMachineClientSet(controlMachineObjects...)
	fakeTypedMachineClient := &faketyped.FakeMachineV1alpha1{
		Fake: &fakeControlMachineClient.Fake,
	}
	fakeTargetCoreClient, targetCoreObjectTracker := customfake.NewCoreClientSet(targetCoreObjects...)
	fakeControlAppsClient, controlAppsObjectTracker := customfake.NewAppsClientSet(controlAppsObjects...)
	fakeObjectTrackers := customfake.NewFakeObjectTrackers(
		controlMachineObjectTracker,
		targetCoreObjectTracker,
		controlAppsObjectTracker,
	)
	fakeObjectTrackers.Start()
	coreTargetInformerFactory := coreinformers.NewFilteredSharedInformerFactory(
		fakeTargetCoreClient,
		100*time.Millisecond,
		namespace,
		nil,
	)
	defer coreTargetInformerFactory.Start(stop)
	coreTargetSharedInformers := coreTargetInformerFactory.Core().V1()
	nodes := coreTargetSharedInformers.Nodes()

	appsControlInformerFactory := appsv1informers.NewFilteredSharedInformerFactory(
		fakeControlAppsClient,
		100*time.Millisecond,
		namespace,
		nil,
	)
	defer appsControlInformerFactory.Start(stop)
	appsControlSharedInformers := appsControlInformerFactory.Apps().V1()

	controlMachineInformerFactory := machineinformers.NewFilteredSharedInformerFactory(
		fakeControlMachineClient,
		100*time.Millisecond,
		namespace,
		nil,
	)
	defer controlMachineInformerFactory.Start(stop)

	machineSharedInformers := controlMachineInformerFactory.Machine().V1alpha1()
	machines := machineSharedInformers.Machines()
	machineSets := machineSharedInformers.MachineSets()
	machineDeployments := machineSharedInformers.MachineDeployments()
	machineClasses := machineSharedInformers.MachineClasses()

	internalExternalScheme := runtime.NewScheme()
	g.Expect(machineinternal.AddToScheme(internalExternalScheme)).To(gomega.Succeed())
	g.Expect(v1alpha1.AddToScheme(internalExternalScheme)).To(gomega.Succeed())

	mcmManager := McmManager{
		namespace: namespace,
		interrupt: make(chan struct{}),
		discoveryOpts: cloudprovider.NodeGroupDiscoveryOptions{
			NodeGroupSpecs: nodeGroups,
		},
		nodeGroups:              make(map[types.NamespacedName]*nodeGroup),
		deploymentLister:        appsControlSharedInformers.Deployments().Lister(),
		machineClient:           fakeTypedMachineClient,
		machineDeploymentLister: machineDeployments.Lister(),
		machineSetLister:        machineSets.Lister(),
		machineLister:           machines.Lister(),
		machineClassLister:      machineClasses.Lister(),
		nodeLister:              nodes.Lister(),
		nodeInterface:           fakeTargetCoreClient.CoreV1().Nodes(),
		maxRetryTimeout:         5 * time.Second,
		retryInterval:           1 * time.Second,
	}
	g.Expect(mcmManager.generateMachineDeploymentMap()).To(gomega.Succeed())
	hasSyncedCachesFns := []cache.InformerSynced{
		nodes.Informer().HasSynced,
		machines.Informer().HasSynced,
		machineSets.Informer().HasSynced,
		machineDeployments.Informer().HasSynced,
		machineClasses.Informer().HasSynced,
		appsControlSharedInformers.Deployments().Informer().HasSynced,
	}

	return &mcmManager, fakeObjectTrackers, hasSyncedCachesFns
}

func waitForCacheSync(t *testing.T, stop <-chan struct{}, hasSyncedCachesFns []cache.InformerSynced) {
	g := gomega.NewWithT(t)
	g.Expect(mcmcache.WaitForCacheSync(stop, hasSyncedCachesFns...)).To(gomega.BeTrue())
}
