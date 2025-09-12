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

package client

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	capacitybuffer "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/informers/externalversions"
	bufferslisters "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/listers/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/client-go/discovery"
	kubernetes "k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	batchv1lister "k8s.io/client-go/listers/batch/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	scaleclient "k8s.io/client-go/scale"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingapi "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	klog "k8s.io/klog/v2"
)

// CapacityBufferClient represents client for v1 capacitybuffer CRD.
type CapacityBufferClient struct {
	buffersClient         capacitybuffer.Interface
	kubernetesClient      kubernetes.Interface
	scaleGetter           scaleclient.ScalesGetter
	scaleMapper           meta.RESTMapper
	buffersLister         bufferslisters.CapacityBufferLister
	podTemplateLister     corev1listers.PodTemplateLister
	replicaSetsLister     appsv1listers.ReplicaSetLister
	statefulSetsLister    appsv1listers.StatefulSetLister
	jobsLister            batchv1lister.JobLister
	deploymentLister      appsv1listers.DeploymentLister
	replicationContLister corev1listers.ReplicationControllerLister
}

// NewCapacityBufferClient returns a capacityBufferClient.
func NewCapacityBufferClient(buffersClient capacitybuffer.Interface, kubernetesClient kubernetes.Interface, buffersLister bufferslisters.CapacityBufferLister,
	podTemplateLister corev1listers.PodTemplateLister, replicaSetsLister appsv1listers.ReplicaSetLister, statefulSetsLister appsv1listers.StatefulSetLister,
	jobsLister batchv1lister.JobLister, deploymentLister appsv1listers.DeploymentLister, replicationContLister corev1listers.ReplicationControllerLister,
	scaleGetter scaleclient.ScalesGetter, scaleMapper meta.RESTMapper) (*CapacityBufferClient, error) {
	return &CapacityBufferClient{
		buffersClient:         buffersClient,
		kubernetesClient:      kubernetesClient,
		scaleGetter:           scaleGetter,
		buffersLister:         buffersLister,
		podTemplateLister:     podTemplateLister,
		replicaSetsLister:     replicaSetsLister,
		statefulSetsLister:    statefulSetsLister,
		jobsLister:            jobsLister,
		deploymentLister:      deploymentLister,
		replicationContLister: replicationContLister,
		scaleMapper:           scaleMapper,
	}, nil
}

// NewCapacityBufferClientFromConfig configures and returns a CapacityBufferClient.
func NewCapacityBufferClientFromConfig(kubeConfig *rest.Config) (*CapacityBufferClient, error) {
	buffersClient, err := capacitybuffer.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create capacity buffer client for capacity buffer: %v", err)
	}
	kubernetesClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create kubernetes client for capacity buffer: %v", err)
	}
	scaleGetter, scaleMapper, err := createScaleSubresourceClientGetter(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create scale getter for capacity buffer: %v", err)
	}
	return NewCapacityBufferClientFromClients(buffersClient, kubernetesClient, scaleGetter, scaleMapper)
}

func createScaleSubresourceClientGetter(kubeConfig *rest.Config) (scaleclient.ScalesGetter, meta.RESTMapper, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}
	scaleMapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))
	scaleKindResolver := scaleclient.NewDiscoveryScaleKindResolver(discoveryClient)
	client, err := scaleclient.NewForConfig(kubeConfig, scaleMapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		return nil, nil, err
	}
	return client, scaleMapper, nil
}

// NewCapacityBufferClientFromClients returns a CapacityBufferClient based on the passed clients
func NewCapacityBufferClientFromClients(buffersClient capacitybuffer.Interface, kubernetesClient kubernetes.Interface, scaleGetter scaleclient.ScalesGetter, scaleMapper meta.RESTMapper) (*CapacityBufferClient, error) {
	if buffersClient == nil || kubernetesClient == nil {
		return nil, fmt.Errorf("Couldn't create capacity buffer client")
	}

	stopChannel := make(chan struct{})
	buffersLister, err := newBuffersLister(buffersClient, stopChannel, 5*time.Second)
	if err != nil {
		return nil, err
	}

	factory := informers.NewSharedInformerFactory(kubernetesClient, 1*time.Minute)
	bufferClient := &CapacityBufferClient{
		buffersClient:         buffersClient,
		kubernetesClient:      kubernetesClient,
		scaleGetter:           scaleGetter,
		scaleMapper:           scaleMapper,
		buffersLister:         buffersLister,
		podTemplateLister:     factory.Core().V1().PodTemplates().Lister(),
		replicaSetsLister:     factory.Apps().V1().ReplicaSets().Lister(),
		statefulSetsLister:    factory.Apps().V1().StatefulSets().Lister(),
		jobsLister:            factory.Batch().V1().Jobs().Lister(),
		deploymentLister:      factory.Apps().V1().Deployments().Lister(),
		replicationContLister: factory.Core().V1().ReplicationControllers().Lister(),
	}
	factory.Start(stopChannel)
	informersSynced := factory.WaitForCacheSync(stopChannel)
	for _, synced := range informersSynced {
		if !synced {
			return nil, fmt.Errorf("Can't initiate informer factory syncer")
		}
	}
	return bufferClient, nil
}

// newBuffersLister creates a lister for the buffers in the cluster.
func newBuffersLister(client capacitybuffer.Interface, stopChannel <-chan struct{}, defaultResync time.Duration) (bufferslisters.CapacityBufferLister, error) {
	factory := externalversions.NewSharedInformerFactory(client, defaultResync)
	buffersLister := factory.Autoscaling().V1alpha1().CapacityBuffers().Lister()

	factory.Start(stopChannel)
	informersSynced := factory.WaitForCacheSync(stopChannel)
	for _, synced := range informersSynced {
		if !synced {
			return nil, fmt.Errorf("can't create buffers lister")
		}
	}
	klog.V(2).Info("Successful initial buffers sync")
	return buffersLister, nil
}

// ListCapacityBuffers lists all Capacity buffer.
func (c *CapacityBufferClient) ListCapacityBuffers() ([]*v1.CapacityBuffer, error) {
	if c.buffersLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured to list capacity buffers")
	}

	buffers, err := c.buffersLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("Error fetching capacity buffers: %v", err)
	}
	buffersCopy := []*v1.CapacityBuffer{}
	for _, buffer := range buffers {
		buffersCopy = append(buffersCopy, buffer.DeepCopy())
	}
	return buffersCopy, nil
}

// GetPodTemplate returns pod template with the passed name
func (c *CapacityBufferClient) GetPodTemplate(namespace, name string) (*corev1.PodTemplate, error) {
	if c.podTemplateLister == nil && c.kubernetesClient == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting pod template")
	}
	if c.podTemplateLister != nil {
		template, err := c.podTemplateLister.PodTemplates(namespace).Get(name)
		if err == nil {
			return template.DeepCopy(), nil
		}
	}
	template, err := c.kubernetesClient.CoreV1().PodTemplates(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Capacity buffer client can't get pod template, error %v", err.Error())
	}
	return template.DeepCopy(), nil
}

// UpdateCapacityBuffer fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) UpdateCapacityBuffer(buffer *v1.CapacityBuffer) (*v1.CapacityBuffer, error) {
	if c.buffersClient == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for updating capacity buffer")
	}

	buffer, err := c.buffersClient.AutoscalingV1alpha1().CapacityBuffers(buffer.Namespace).UpdateStatus(context.TODO(), buffer, metav1.UpdateOptions{})
	if err == nil {
		return buffer.DeepCopy(), nil
	}
	return nil, err
}

// CreatePodTemplate creates a pod template
func (c *CapacityBufferClient) CreatePodTemplate(podTemplate *corev1.PodTemplate) (*corev1.PodTemplate, error) {
	if c.kubernetesClient == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for creating pod template")
	}
	template, err := c.kubernetesClient.CoreV1().PodTemplates(podTemplate.Namespace).Create(context.TODO(), podTemplate, metav1.CreateOptions{})
	if err == nil {
		return template.DeepCopy(), nil
	}
	return nil, err
}

// UpdatePodTemplate updates the pod template
func (c *CapacityBufferClient) UpdatePodTemplate(podTemplate *corev1.PodTemplate) (*corev1.PodTemplate, error) {
	if c.kubernetesClient == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for updating pod template")
	}
	template, err := c.kubernetesClient.CoreV1().PodTemplates(podTemplate.Namespace).Update(context.TODO(), podTemplate, metav1.UpdateOptions{})
	if err == nil {
		return template.DeepCopy(), nil
	}
	return nil, err
}

// GetDeployment fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	if c.deploymentLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting deployment")
	}

	obj, err := c.deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployment: %w", err)
	}
	return obj, nil
}

// GetReplicaSet fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) GetReplicaSet(namespace, name string) (*appsv1.ReplicaSet, error) {
	if c.replicaSetsLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting replicaSets")
	}

	obj, err := c.replicaSetsLister.ReplicaSets(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get ReplicaSet: %w", err)
	}
	return obj, nil
}

// GetJob fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) GetJob(namespace, name string) (*batchv1.Job, error) {
	if c.jobsLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting jobs")
	}

	obj, err := c.jobsLister.Jobs(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get Job: %w", err)
	}
	return obj, nil
}

// GetReplicationController fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) GetReplicationController(namespace, name string) (*corev1.ReplicationController, error) {
	if c.replicationContLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting replicationController")
	}
	obj, err := c.replicationContLister.ReplicationControllers(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get ReplicationController: %w", err)
	}
	return obj, nil
}

// GetStatefulSet fetches the cached object using a lister object in the client
func (c *CapacityBufferClient) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	if c.statefulSetsLister == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting StatefulSet")
	}

	obj, err := c.statefulSetsLister.StatefulSets(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get StatefulSet: %w", err)
	}
	return obj, nil
}

// GetScaleObject resolves the api group and kind to group resource and use it to get the scale sub-resource with passed name from the passed namespace
func (c *CapacityBufferClient) GetScaleObject(namespace, group, kind, name string) (*autoscalingapi.Scale, error) {
	if c.scaleMapper == nil || c.scaleGetter == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for scale objects")
	}
	gvk := schema.GroupKind{
		Group: group,
		Kind:  kind,
	}
	mapping, err := c.scaleMapper.RESTMapping(gvk, "")
	if err != nil {
		return nil, err
	}
	obj, err := c.scaleGetter.Scales(namespace).Get(context.TODO(), mapping.Resource.GroupResource(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get scale sub resource: %w", err)
	}

	return obj, nil
}

// GetPodsBySelector resolves the api group and kind to group resource and use it to get the scale sub-resource with passed name from the passed namespace
func (c *CapacityBufferClient) GetPodsBySelector(namespace, selector string) ([]corev1.Pod, error) {
	if c.kubernetesClient == nil {
		return nil, fmt.Errorf("Capacity buffer client is not configured for getting pods objects")
	}
	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector: %v", err)
	}
	podList, err := c.kubernetesClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}
	return podList.Items, nil
}
