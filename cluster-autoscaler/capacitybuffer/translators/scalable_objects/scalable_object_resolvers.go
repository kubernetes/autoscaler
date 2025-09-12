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

package scalableobject

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
)

// Kinds of the supported objects
const (
	DeploymentKind            = "Deployment"
	ReplicaSetKind            = "ReplicaSet"
	StatefulSetKind           = "StatefulSet"
	ReplicationControllerKind = "ReplicationController"
	JobKind                   = "Job"
	ApiGroupApps              = "apps"
	ApiGroupBatch             = "batch"
	ApiGroupCore              = "core"
)

// ScaleObjectPodResolver resolves scale objects into pod specs and number of replicas only if there is at least one exiting pod
type ScaleObjectPodResolver struct {
	client *cbclient.CapacityBufferClient
}

// NewScaleObjectPodResolver returns new ScaleObjectPodResolver
func NewScaleObjectPodResolver(client *cbclient.CapacityBufferClient) *ScaleObjectPodResolver {
	return &ScaleObjectPodResolver{
		client: client,
	}
}

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *ScaleObjectPodResolver) GetTemplateAndReplicas(namespace, group, kind, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetScaleObject(namespace, group, kind, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get scale object: %w", err)
	}
	podsList, err := s.client.GetPodsBySelector(namespace, obj.Status.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get existing pod for scale object: %w", err)
	}
	if len(podsList) == 0 {
		return nil, &obj.Status.Replicas, nil
	}
	pod := getMostRecentPod(podsList)
	return buildPodTemplateFromPod(pod), &obj.Status.Replicas, nil
}

func getMostRecentPod(podList []corev1.Pod) *corev1.Pod {
	sort.Slice(podList, func(i, j int) bool {
		return podList[i].CreationTimestamp.After(podList[j].CreationTimestamp.Time)
	})
	return &podList[0]
}

func buildPodTemplateFromPod(pod *corev1.Pod) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
		},
		Spec: pod.Spec,
	}
}

// GetSupportedScalableObjectResolvers returns the default ScalableObjectResolvers
func GetSupportedScalableObjectResolvers(client *cbclient.CapacityBufferClient) []ScalableObjectTemplateResolver {
	return []ScalableObjectTemplateResolver{
		&deployment{scalableObjectTemplateResolver{client: client, kind: DeploymentKind, apiGroup: ApiGroupApps}},
		&replicaSet{scalableObjectTemplateResolver{client: client, kind: ReplicaSetKind, apiGroup: ApiGroupApps}},
		&statefulSet{scalableObjectTemplateResolver{client: client, kind: StatefulSetKind, apiGroup: ApiGroupApps}},
		&replicationController{scalableObjectTemplateResolver{client: client, kind: ReplicationControllerKind, apiGroup: ApiGroupCore}},
		&job{scalableObjectTemplateResolver{client: client, kind: JobKind, apiGroup: ApiGroupBatch}},
	}
}

// ScalableObjectTemplateResolver is an interface for resolvers that are supported to get pod spec template for a any object
type ScalableObjectTemplateResolver interface {
	GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error)
	GetResolverKey() string
}

type scalableObjectTemplateResolver struct {
	client   *cbclient.CapacityBufferClient
	kind     string
	apiGroup string
}

// GetResolverKey returns a string that distinguishes the resolver by api group and king
func (s *scalableObjectTemplateResolver) GetResolverKey() string {
	return GetResolverKey(s.apiGroup, s.kind)
}

// GetResolverKey returns the key of a resolver given the api group and kind
func GetResolverKey(apiGroup, kind string) string {
	return fmt.Sprintf("%v-%v", apiGroup, kind)
}

type deployment struct{ scalableObjectTemplateResolver }

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *deployment) GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetDeployment(namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj.Spec.Template.DeepCopy(), obj.Spec.Replicas, nil
}

type replicaSet struct{ scalableObjectTemplateResolver }

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *replicaSet) GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetReplicaSet(namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj.Spec.Template.DeepCopy(), obj.Spec.Replicas, nil
}

type statefulSet struct{ scalableObjectTemplateResolver }

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *statefulSet) GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetStatefulSet(namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj.Spec.Template.DeepCopy(), obj.Spec.Replicas, nil
}

type replicationController struct{ scalableObjectTemplateResolver }

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *replicationController) GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetReplicationController(namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj.Spec.Template.DeepCopy(), obj.Spec.Replicas, nil
}

type job struct{ scalableObjectTemplateResolver }

// GetTemplateAndReplicas returns the pod spec template of the passed object name and namespace
func (s *job) GetTemplateAndReplicas(namespace, name string) (*corev1.PodTemplateSpec, *int32, error) {
	obj, err := s.client.GetJob(namespace, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj.Spec.Template.DeepCopy(), obj.Spec.Parallelism, nil
}
