/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package nanny

import (
	"fmt"
	"time"

	api "k8s.io/kubernetes/pkg/api"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	cache "k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	runtime "k8s.io/kubernetes/pkg/runtime"
	wait "k8s.io/kubernetes/pkg/util/wait"
	watch "k8s.io/kubernetes/pkg/watch"
)

type kubernetesClient struct {
	namespace  string
	deployment string
	pod        string
	container  string
	clientset  *client.Clientset
	nodeStore  cache.Store
	reflector  *cache.Reflector
}

func (k *kubernetesClient) CountNodes() (uint64, error) {
	err := wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		if k.reflector.LastSyncResourceVersion() == "" {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return 0, err
	}
	return uint64(len(k.nodeStore.List())), nil
}

func (k *kubernetesClient) ContainerResources() (*apiv1.ResourceRequirements, error) {
	pod, err := k.clientset.CoreClient.Pods(k.namespace).Get(k.pod)

	if err != nil {
		return nil, err
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == k.container {
			return &container.Resources, nil
		}
	}
	return nil, fmt.Errorf("Container %s was not found in deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

func (k *kubernetesClient) UpdateDeployment(resources *apiv1.ResourceRequirements) error {
	// First, get the Deployment.
	dep, err := k.clientset.Extensions().Deployments(k.namespace).Get(k.deployment)
	if err != nil {
		return err
	}

	// Modify the Deployment object with our ResourceRequirements.
	for i, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == k.container {
			// Update the deployment.
			dep.Spec.Template.Spec.Containers[i].Resources = *resources
			_, err = k.clientset.ExtensionsClient.Deployments(k.namespace).Update(dep)
			return err
		}
	}

	return fmt.Errorf("Container %s was not found in the deployment %s in namespace %s.", k.container, k.deployment, k.namespace)
}

// NewKubernetesClient gives a KubernetesClient with the given dependencies.
func NewKubernetesClient(namespace, deployment, pod, container string, clientset *client.Clientset) KubernetesClient {
	result := &kubernetesClient{
		namespace:  namespace,
		deployment: deployment,
		pod:        pod,
		container:  container,
		clientset:  clientset,
		nodeStore:  cache.NewStore(cache.MetaNamespaceKeyFunc),
	}
	// Start propagating contents of the nodeStore.
	nodeListWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			return clientset.Core().Nodes().List(options)
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			return clientset.Core().Nodes().Watch(options)
		},
	}
	result.reflector = cache.NewReflector(nodeListWatch, &apiv1.Node{}, result.nodeStore, 0)
	result.reflector.Run()
	return result
}
