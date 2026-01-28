/*
Copyright 2025 The Kubernetes Authors.

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

package fake

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	listersv1 "k8s.io/client-go/listers/apps/v1"
	"sync"
)

// Kubernetes encapsulates a fake Kubernetes client and its corresponding listers.
type Kubernetes struct {
	KubeClient *fake.Clientset // The client for write actions.

	AllNodeLister       *nodeLister
	ReadyNodeLister     *nodeLister
	PodLister           *podLister
	PodDisruptionLister *pdbLister
	DaemonSetLister     *daemonSetLister
}

// NewKubernetes creates a new, fully wired fake Kubernetes simulation.
func NewKubernetes() *Kubernetes {
	KubeClient := fake.NewClientset()

	return &Kubernetes{
		KubeClient:          KubeClient,
		AllNodeLister:       &nodeLister{nodes: make(map[string]*apiv1.Node)},
		ReadyNodeLister:     &nodeLister{nodes: make(map[string]*apiv1.Node)},
		PodLister:           &podLister{pods: make(map[string]*apiv1.Pod)},
		PodDisruptionLister: &pdbLister{pdbs: make([]*policyv1.PodDisruptionBudget, 0), podLister: &podLister{pods: make(map[string]*apiv1.Pod)}},
		DaemonSetLister:     newDaemonSetLister(),
	}
}

// AddNode adds a node to the fake client and the fake listers.
func (k *Kubernetes) AddNode(node *apiv1.Node) {
	k.AllNodeLister.AddNode(node)

	if isNodeReady(node) {
		k.ReadyNodeLister.AddNode(node)
	} else {
		k.ReadyNodeLister.DeleteNode(node.Name)
	}

	k.KubeClient.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
}

// UpdateNode updates a node.
func (k *Kubernetes) UpdateNode(node *apiv1.Node) {
	k.AllNodeLister.AddNode(node)

	if isNodeReady(node) {
		k.ReadyNodeLister.AddNode(node)
	} else {
		k.ReadyNodeLister.DeleteNode(node.Name)
	}

	k.KubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
}

func isNodeReady(node *apiv1.Node) bool {
	if node == nil {
		return false
	}

	for _, condition := range node.Status.Conditions {
		if condition.Type == apiv1.NodeReady {
			return condition.Status == apiv1.ConditionTrue
		}
	}
	return false
}

// AddPod adds a pod to the fake client and the fake pod lister.
func (k *Kubernetes) AddPod(pod *apiv1.Pod) {
	k.PodLister.AddPod(pod)
	k.KubeClient.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
}

// RemovePod deletes a pod from both the lister and the fake client.
func (k *Kubernetes) RemovePod(namespace, name string) {
	k.PodLister.DeletePod(name)
	k.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// RemoveNode deletes a node from listers and the fake client.
func (k *Kubernetes) RemoveNode(name string) {
	k.AllNodeLister.DeleteNode(name)
	k.ReadyNodeLister.DeleteNode(name)
	k.KubeClient.CoreV1().Nodes().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// ListerRegistry returns a real ListerRegistry populated with the fake listers.
func (k *Kubernetes) ListerRegistry() kubernetes.ListerRegistry {
	// For listers we haven't faked yet, we can pass nil.
	return kubernetes.NewListerRegistry(
		k.AllNodeLister,
		k.ReadyNodeLister,
		k.PodLister,
		k.PodDisruptionLister,
		k.DaemonSetLister,
		nil, // ReplicationControllerLister
		nil, // JobLister
		nil, // ReplicaSetLister
		nil, // StatefulSetLister
	)
}

// nodeLister is a fake implementation of the kube_util.NodeLister interface for testing.
type nodeLister struct {
	sync.Mutex
	nodes map[string]*apiv1.Node
}

// List returns the list of nodes stored in the fake.
func (f *nodeLister) List() ([]*apiv1.Node, error) {
	f.Lock()
	defer f.Unlock()
	list := make([]*apiv1.Node, 0, len(f.nodes))
	for _, node := range f.nodes {
		list = append(list, node)
	}
	return list, nil
}

// Get returns a specific node by name.
func (f *nodeLister) Get(name string) (*apiv1.Node, error) {
	f.Lock()
	defer f.Unlock()
	node, found := f.nodes[name]
	if !found {
		return nil, fmt.Errorf("node %s not found", name)
	}
	return node, nil
}

// SetNodes allows a test to update the list of nodes in the fake.
func (f *nodeLister) SetNodes(nodes []*apiv1.Node) {
	f.Lock()
	defer f.Unlock()
	f.nodes = make(map[string]*apiv1.Node, len(nodes))
	for _, node := range nodes {
		f.nodes[node.Name] = node
	}
}

// AddNode adds a node to the map.
func (f *nodeLister) AddNode(node *apiv1.Node) {
	f.Lock()
	defer f.Unlock()
	f.nodes[node.Name] = node
}

// DeleteNode removes a node from the map.
func (f *nodeLister) DeleteNode(name string) {
	f.Lock()
	defer f.Unlock()
	delete(f.nodes, name)
}

// podLister is a fake implementation of the kube_util.PodLister interface for testing.
type podLister struct {
	sync.Mutex
	pods map[string]*apiv1.Pod
}

// List returns the list of pods stored in the fake.
func (f *podLister) List() ([]*apiv1.Pod, error) {
	f.Lock()
	defer f.Unlock()
	list := make([]*apiv1.Pod, 0, len(f.pods))
	for _, pod := range f.pods {
		list = append(list, pod)
	}
	return list, nil
}

// Get returns a specific pod by name.
func (f *podLister) Get(name string) (*apiv1.Pod, error) {
	f.Lock()
	defer f.Unlock()
	pod, found := f.pods[name]
	if !found {
		return nil, fmt.Errorf("pod %s not found", name)
	}
	return pod, nil
}

// SetPods allows a test to update the list of pods in the fake.
func (f *podLister) SetPods(pods []*apiv1.Pod) {
	f.Lock()
	defer f.Unlock()
	f.pods = make(map[string]*apiv1.Pod, len(pods))
	for _, pod := range pods {
		f.pods[pod.Name] = pod
	}
}

// AddPod adds a pod to the map. (Required for AddPod logic)
func (f *podLister) AddPod(pod *apiv1.Pod) {
	f.Lock()
	defer f.Unlock()
	f.pods[pod.Name] = pod
}

// DeletePod removes a pod from the map. (Essential for O(1) deletion)
func (f *podLister) DeletePod(name string) {
	f.Lock()
	defer f.Unlock()
	delete(f.pods, name)
}

// daemonSetLister is a test implementation of StatefulSetLister.
type daemonSetLister struct {
	dSets map[string]*daemonSetNamespaceLister
	listersv1.DaemonSetListerExpansion
}

// newDaemonSetLister returns a new instance of statefulSetLister
func newDaemonSetLister() *daemonSetLister {
	return &daemonSetLister{dSets: make(map[string]*daemonSetNamespaceLister)}
}

// Add adds a new DaemonSet.
func (dl *daemonSetLister) Add(dSet *appsv1.DaemonSet) {
	if dl.dSets[dSet.Namespace] == nil {
		dl.dSets[dSet.Namespace] = &daemonSetNamespaceLister{
			dSets: make(map[string]*appsv1.DaemonSet),
		}
	}
	dl.dSets[dSet.Namespace].dSets[dSet.Name] = dSet
}

// List lists all existing daemon sets.
func (dl *daemonSetLister) List(_ labels.Selector) ([]*appsv1.DaemonSet, error) {
	var result []*appsv1.DaemonSet
	for _, dSet := range dl.dSets {
		el, _ := dSet.List(nil)
		result = append(result, el...)
	}
	return result, nil
}

// DaemonSets returns DaemonSetNamespaceLister for provided namespace.
func (dl *daemonSetLister) DaemonSets(namespace string) listersv1.DaemonSetNamespaceLister {
	if dl.dSets[namespace] == nil {
		dl.dSets[namespace] = &daemonSetNamespaceLister{
			dSets: make(map[string]*appsv1.DaemonSet),
		}
	}
	return dl.dSets[namespace]
}

type daemonSetNamespaceLister struct {
	dSets map[string]*appsv1.DaemonSet
}

func (dNL *daemonSetNamespaceLister) Get(name string) (*appsv1.DaemonSet, error) {
	if _, ok := dNL.dSets[name]; !ok {
		return nil, fmt.Errorf("daemonSet %v does not exist", name)
	}
	return dNL.dSets[name], nil
}

func (dNL *daemonSetNamespaceLister) List(_ labels.Selector) ([]*appsv1.DaemonSet, error) {
	var result []*appsv1.DaemonSet
	for _, ds := range dNL.dSets {
		result = append(result, ds)
	}
	return result, nil
}

type pdbLister struct {
	sync.Mutex
	pdbs      []*policyv1.PodDisruptionBudget
	podLister *podLister
}

// Add adds new pdb.
func (pdl *pdbLister) Add(pdb *policyv1.PodDisruptionBudget) error {
	pdl.pdbs = append(pdl.pdbs, pdb)
	// TODO implement logic.
	return nil
}

func (pdl *pdbLister) PodEvicted(pod *apiv1.Pod) {
	if pod == nil {
		return
	}
	// TODO implement logic.
}

func (pdl *pdbLister) PodScheduled(pod *apiv1.Pod) {

}

// List lists existing pdbs.
func (pdl *pdbLister) List() ([]*policyv1.PodDisruptionBudget, error) {
	return pdl.pdbs, nil
}
