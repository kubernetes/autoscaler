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

package eviction

import (
	"fmt"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// PodsEvictionRestriction controls pods evictions. It ensures that we will not evict too
// many pods from one replica set. For replica set will allow to evict one pod or more if
// evictionToleranceFraction is configured.
type PodsEvictionRestriction interface {
	// Evict sends eviction instruction to the api client.
	// Retrurns error if pod cannot be evicted or if client returned error.
	Evict(pod *apiv1.Pod) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *apiv1.Pod) bool
}

type podsEvictionRestrictionImpl struct {
	client          kube_client.Interface
	podsControllers map[string]podReplicaController
	evictionBudget  map[podReplicaController]int
}

// PodsEvictionRestrictionFactory creates PodsEvictionRestriction
type PodsEvictionRestrictionFactory interface {
	// NewPodsEvictionRestriction creates PodsEvictionRestriction for given set of pods.
	NewPodsEvictionRestriction(pods []*apiv1.Pod) PodsEvictionRestriction
}

type podsEvictionRestrictionFactoryImpl struct {
	client                    kube_client.Interface
	minReplicas               int
	evictionToleranceFraction float64
}

type podReplicaController struct {
	Namespace string
	Name      string
	Kind      string
}

// CanEvict checks if pod can be safely evicted
func (e *podsEvictionRestrictionImpl) CanEvict(pod *apiv1.Pod) bool {
	cr, present := e.podsControllers[getPodID(pod)]
	if present {
		return e.evictionBudget[cr] > 0
	}
	return false
}

// Evict sends eviction instruction to api client. Retrurns error if pod cannot be evicted or if client returned error
// Does not check if pod was actually evicted after eviction grace period.
func (e *podsEvictionRestrictionImpl) Evict(podToEvict *apiv1.Pod) error {
	cr, present := e.podsControllers[getPodID(podToEvict)]
	if !present {
		return fmt.Errorf("pod not suitable for eviction %v : not in replicated pods map", podToEvict.Name)
	}
	if e.evictionBudget[cr] < 1 {
		return fmt.Errorf("cannot evict pod %v : eviction budget exceeded", podToEvict.Name)
	}

	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: podToEvict.Namespace,
			Name:      podToEvict.Name,
		},
	}
	err := e.client.CoreV1().Pods(podToEvict.Namespace).Evict(eviction)
	if err != nil {
		glog.Errorf("failed to evict pod %s, error: %v", podToEvict.Name, err)
		return err
	}
	e.evictionBudget[cr] = e.evictionBudget[cr] - 1
	return nil
}

// NewPodsEvictionRestrictionFactory creates PodsEvictionRestrictionFactory
func NewPodsEvictionRestrictionFactory(client kube_client.Interface, minReplicas int, evictionToleranceFraction float64) PodsEvictionRestrictionFactory {
	return &podsEvictionRestrictionFactoryImpl{client: client, minReplicas: minReplicas, evictionToleranceFraction: evictionToleranceFraction}
}

// NewPodsEvictionRestriction creates PodsEvictionRestriction for a given set of pods.
func (f *podsEvictionRestrictionFactoryImpl) NewPodsEvictionRestriction(pods []*apiv1.Pod) PodsEvictionRestriction {
	// We can evict pod only if it is a part of replica set
	// For each replica set we can evict only a fraction of pods.
	// Evictions may be later limited by pod disruption budget if configured.

	livePods := make(map[podReplicaController][]*apiv1.Pod)

	for _, pod := range pods {
		controller, err := getPodReplicaController(pod, &f.client)
		if err != nil {
			glog.Errorf("failed to obtain replication info for pod %s: %v", pod.Name, err)
			continue
		}
		if controller == nil {
			glog.V(2).Infof("pod %s not replicated", pod.Name)
			continue
		}
		livePods[*controller] = append(livePods[*controller], pod)
	}

	podsControllers := make(map[string]podReplicaController)
	controllersEvictionBudget := make(map[podReplicaController]int)
	for controller, replicas := range livePods {
		actual := len(replicas)
		if actual < f.minReplicas {
			glog.V(2).Infof("too few replicas for %v %v/%v. Found %v live pods",
				controller.Kind, controller.Namespace, controller.Name, actual)
			continue
		}

		var configured int
		if controller.Kind == "Job" {
			// Job has no replicas configuration, so we will use actual number of live pods as replicas count.
			configured = actual
		} else {
			var err error
			configured, err = getReplicaCount(controller, f.client)
			if err != nil {
				glog.Errorf("failed to obtain replication info for %v %v/%v. %v",
					controller.Kind, controller.Namespace, controller.Name, err)
				continue
			}
		}

		evictionTolerance := int(float64(configured) * f.evictionToleranceFraction)
		currentlyEvicted := configured - actual
		evictionBudget := evictionTolerance - currentlyEvicted

		if evictionBudget > 0 {
			controllersEvictionBudget[controller] = evictionBudget
		} else {
			controllersEvictionBudget[controller] = 1
			glog.V(2).Infof("configured eviction tolerance for pods from %v %v/%v too low. Setting eviction budget to 1",
				controller.Kind, controller.Namespace, controller.Name)
		}

		for _, pod := range replicas {
			podsControllers[getPodID(pod)] = controller
		}
	}
	return &podsEvictionRestrictionImpl{client: f.client, podsControllers: podsControllers, evictionBudget: controllersEvictionBudget}
}

func getPodReplicaController(pod *apiv1.Pod, client *kube_client.Interface) (*podReplicaController, error) {
	controllerRef, err := controllerRef(pod, client)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain pod controller reference: %v", err)
	}
	if controllerRef == nil {
		return nil, nil
	}
	return &podReplicaController{Namespace: pod.Namespace,
		Name: controllerRef.Name,
		Kind: controllerRef.Kind}, nil
}

func getPodID(pod *apiv1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Namespace + "/" + pod.Name
}

func getReplicaCount(controller podReplicaController, client kube_client.Interface) (int, error) {
	switch controller.Kind {
	case "ReplicationController":
		rc, err := client.CoreV1().ReplicationControllers(controller.Namespace).Get(controller.Name, metav1.GetOptions{})

		if err != nil || rc == nil {
			return 0, fmt.Errorf("replication controller %s/%s is not available, err: %v", controller.Namespace, controller.Name, err)
		}
		if rc.Spec.Replicas == nil || *rc.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replication controller %s/%s has no replicas config", controller.Namespace, controller.Name)
		}
		return int(*rc.Spec.Replicas), nil

	case "ReplicaSet":
		rs, err := client.ExtensionsV1beta1().ReplicaSets(controller.Namespace).Get(controller.Name, metav1.GetOptions{})

		if err != nil || rs == nil {
			return 0, fmt.Errorf("replica set %s/%s is not available, err: %v", controller.Namespace, controller.Name, err)
		}
		if rs.Spec.Replicas == nil || *rs.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replica set %s/%s has no replicas config", controller.Namespace, controller.Name)
		}
		return int(*rs.Spec.Replicas), nil

	case "StatefulSet":
		ss, err := client.AppsV1beta1().StatefulSets(controller.Namespace).Get(controller.Name, metav1.GetOptions{})
		if err != nil || ss == nil {
			return 0, fmt.Errorf("stateful set %s/%s is not available, err: %v", controller.Namespace, controller.Name, err)
		}
		if ss.Spec.Replicas == nil || *ss.Spec.Replicas == 0 {
			return 0, fmt.Errorf("stateful set %s/%s has no replicas config", controller.Namespace, controller.Name)
		}
		return int(*ss.Spec.Replicas), nil
	}

	return 0, nil
}

/*
// controllerRef returns the kind of the controller reference of the pod.
func controllerRef(pod *apiv1.Pod) (*apiv1.SerializedReference, error) {
	controllerRef, found := pod.ObjectMeta.Annotations[apiv1.CreatedByAnnotation]
	if !found {
		return nil, nil
	}
	var sr apiv1.SerializedReference
	if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), []byte(controllerRef), &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}
*/

// controllerRef returns the kind of the controller reference of the pod.
func controllerRef(pod *apiv1.Pod, client *kube_client.Interface) (*metav1.OwnerReference, error) {
	controllerRef := getControllerOf(pod)
	if controllerRef == nil {
		return nil, nil
	}

	// We assume the only reason for an error is because the controller is
	// gone/missing, not for any other cause.
	// TODO(mml): something more sophisticated than this
	// TODO(juntee): determine if it's safe to remove getController(),
	// so that drain can work for controller types that we don't know about
	_, err := getController(pod.Namespace, controllerRef, client)
	if err != nil {
		return nil, err
	}
	return controllerRef, nil
}

// getControllerOf returns a pointer to a copy of the controllerRef if controllee has a controller
func getControllerOf(controllee metav1.Object) *metav1.OwnerReference {
	for _, ref := range controllee.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			return &ref
		}
	}
	return nil
}

func getController(namespace string, controllerRef *metav1.OwnerReference, client *kube_client.Interface) (interface{}, error) {
	switch controllerRef.Kind {
	case "ReplicationController":
		return (*client).Core().ReplicationControllers(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	case "DaemonSet":
		return (*client).Extensions().DaemonSets(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	case "Job":
		return (*client).Batch().Jobs(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	case "ReplicaSet":
		return (*client).Extensions().ReplicaSets(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	case "StatefulSet":
		return (*client).Apps().StatefulSets(namespace).Get(controllerRef.Name, metav1.GetOptions{})
	}
	return nil, fmt.Errorf("Unknown controller kind %q", controllerRef.Kind)
}
