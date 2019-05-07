/*
Copyright 2019 The Kubernetes Authors.

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

package target

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types_v1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	vpa_types_v1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_lister_v1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/core"
)

type beta1TargetSelectorFetcher struct {
	vpaLister vpa_lister_v1beta1.VerticalPodAutoscalerLister
}

// NewBeta1TargetSelectorFetcher returns new instance of VpaTargetSelectorFetcher that uses selector from deprecated v1beta1
func NewBeta1TargetSelectorFetcher(config *rest.Config) VpaTargetSelectorFetcher {
	vpaClient := vpa_clientset.NewForConfigOrDie(config)
	return &beta1TargetSelectorFetcher{
		vpaLister: newAllVpasLister(vpaClient, make(chan struct{})),
	}
}

func (f *beta1TargetSelectorFetcher) Fetch(vpa *vpa_types_v1beta2.VerticalPodAutoscaler) (labels.Selector, error) {
	list, err := f.vpaLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("Failed to list v1beta1 VPAs. Reason: %v", err)
	}

	for _, beta1vpa := range list {
		if beta1vpa.Namespace == vpa.Namespace && beta1vpa.Name == vpa.Name && beta1vpa.UID == vpa.UID {
			if beta1vpa.Spec.Selector == nil {
				return nil, fmt.Errorf("v1beta1 selector not found")
			}
			selector, err := metav1.LabelSelectorAsSelector(beta1vpa.Spec.Selector)
			if err != nil {
				// Intentionally ignore error
				return labels.Nothing(), nil
			}
			klog.Infof("Found deprecated label selector for VPA %s/%s", vpa.Namespace, vpa.Name)
			return selector, nil
		}
	}

	return nil, fmt.Errorf("v1beta1 VPA not found")
}

func newAllVpasLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}) vpa_lister_v1beta1.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.AutoscalingV1beta1().RESTClient(), "verticalpodautoscalers", core.NamespaceAll, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(vpaListWatch,
		&vpa_types_v1beta1.VerticalPodAutoscaler{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := vpa_lister_v1beta1.NewVerticalPodAutoscalerLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync VPA v1beta1 cache during initialization")
	} else {
		klog.Info("Initial VPA v1beta1 synced successfully")
	}
	return vpaLister
}
