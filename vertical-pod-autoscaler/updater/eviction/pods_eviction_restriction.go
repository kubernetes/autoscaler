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
	"k8s.io/apimachinery/pkg/runtime"
	api "k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// PodsEvictionRestriction controls pods evictions. Ensures that we will not evict too many pods from one replica set.
// For replica set will allow to evict one pod or more if evictionToleranceFraction is configured.
type PodsEvictionRestriction interface {
	// Evict sends eviction instruction to api client.
	// Retrurns error if pod cannot be evicted or if client returned error.
	Evict(pod *apiv1.Pod) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *apiv1.Pod) bool
}

type podsEvictionRestrictionImpl struct {
	client         kube_client.Interface
	podsCreators   map[string]podReplicaCreator
	evictionBudget map[podReplicaCreator]int
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

type podReplicaCreator struct {
	Namespace string
	Name      string
	Kind      string
}

// CanEvict checks if pod can be safely evicted
func (e *podsEvictionRestrictionImpl) CanEvict(pod *apiv1.Pod) bool {
	cr, present := e.podsCreators[getPodId(pod)]
	if present {
		return e.evictionBudget[cr] > 0
	}
	return false
}

// Evict sends eviction instruction to api client. Retrurns error if pod cannot be evicted or if client returned error
// Does not check if pod was actually evicted after eviction grace period.
func (e *podsEvictionRestrictionImpl) Evict(podToEvict *apiv1.Pod) error {
	cr, present := e.podsCreators[getPodId(podToEvict)]
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

// NewPodsEvictionRestriction creates PodsEvictionRestriction for given set of pods.
func (f *podsEvictionRestrictionFactoryImpl) NewPodsEvictionRestriction(pods []*apiv1.Pod) PodsEvictionRestriction {
	// We can evict pod only if it is a part of replica set
	// For each replica set we can evict only a fraction of pods.
	// Evictions may be later limited by pod disruption budget if configured.

	livePods := make(map[podReplicaCreator][]*apiv1.Pod)

	for _, pod := range pods {
		creator, err := getPodReplicaCreator(pod)
		if err != nil {
			glog.Errorf("failed to obtain replication info for pod: %v", pod.Name, err)
			continue
		}
		if creator == nil {
			glog.V(2).Infof("pod not replicated %v", pod.Name)
			continue
		}
		livePods[*creator] = append(livePods[*creator], pod)
	}

	podsCreators := make(map[string]podReplicaCreator)
	creatorsEvictionBudget := make(map[podReplicaCreator]int)
	for creator, replicas := range livePods {
		actual := len(replicas)
		if actual < f.minReplicas {
			glog.V(2).Infof("too few replicas for %v %v/%v. Found %v live pods",
				creator.Kind, creator.Namespace, creator.Name, actual)
			continue
		}
		var configured int
		if creator.Kind == "Job" {
			// Job has no replicas configuration, so we will use actual number of live pods as replicas count.
			configured = actual
		} else {
			var err error
			configured, err = getReplicaCount(creator, f.client)
			if err != nil {
				glog.Errorf("failed to obtain replication info for %v %v/%v. %v",
					creator.Kind, creator.Namespace, creator.Name, err)
				continue
			}
		}

		evictionTolerance := int(float64(configured) * f.evictionToleranceFraction)
		currentlyEvicted := configured - actual
		evictionBudget := evictionTolerance - currentlyEvicted

		if evictionBudget > 0 {
			creatorsEvictionBudget[creator] = evictionBudget
		} else {
			creatorsEvictionBudget[creator] = 1
			glog.V(2).Infof("configured eviction tolerance for pods from %v %v/%v too low. Setting eviction budget to 1",
				creator.Kind, creator.Namespace, creator.Name)
		}

		for _, pod := range replicas {
			podsCreators[getPodId(pod)] = creator
		}
	}
	return &podsEvictionRestrictionImpl{client: f.client, podsCreators: podsCreators, evictionBudget: creatorsEvictionBudget}
}

func getPodReplicaCreator(pod *apiv1.Pod) (*podReplicaCreator, error) {
	creator, err := creatorRef(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain pod creator reference: %v", err)
	}
	if creator == nil {
		return nil, nil
	}
	return &podReplicaCreator{Namespace: creator.Reference.Namespace,
		Name: creator.Reference.Name,
		Kind: creator.Reference.Kind}, nil
}

func getPodId(pod *apiv1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Namespace + "/" + pod.Name
}

func getReplicaCount(creator podReplicaCreator, client kube_client.Interface) (int, error) {

	if creator.Kind == "ReplicationController" {
		rc, err := client.CoreV1().ReplicationControllers(creator.Namespace).Get(creator.Name, metav1.GetOptions{})

		if err != nil || rc == nil {
			return 0, fmt.Errorf("replication controller %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if rc.Spec.Replicas == nil || *rc.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replication controller %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*rc.Spec.Replicas), nil

	} else if creator.Kind == "ReplicaSet" {
		rs, err := client.ExtensionsV1beta1().ReplicaSets(creator.Namespace).Get(creator.Name, metav1.GetOptions{})

		if err != nil || rs == nil {
			return 0, fmt.Errorf("replica set %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if rs.Spec.Replicas == nil || *rs.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replica set %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*rs.Spec.Replicas), nil

	} else if creator.Kind == "StatefulSet" {
		ss, err := client.AppsV1beta1().StatefulSets(creator.Namespace).Get(creator.Name, metav1.GetOptions{})
		if err != nil || ss == nil {
			return 0, fmt.Errorf("stateful set %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if ss.Spec.Replicas == nil || *ss.Spec.Replicas == 0 {
			return 0, fmt.Errorf("stateful set %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*ss.Spec.Replicas), nil

	}
	return 0, nil
}

// creatorRef returns the kind of the creator reference of the pod.
func creatorRef(pod *apiv1.Pod) (*apiv1.SerializedReference, error) {
	creatorRef, found := pod.ObjectMeta.Annotations[apiv1.CreatedByAnnotation]
	if !found {
		return nil, nil
	}
	var sr apiv1.SerializedReference
	if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), []byte(creatorRef), &sr); err != nil {
		return nil, err
	}
	return &sr, nil
}
