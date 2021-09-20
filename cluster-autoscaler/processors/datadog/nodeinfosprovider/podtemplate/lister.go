/*
Copyright 2021 The Kubernetes Authors.

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

package podtemplate

import (
	"time"

	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// PodTemplateDaemonSetLabelKey use as label key on PodTemplate corresponding to an extra Daemonset.
	PodTemplateDaemonSetLabelKey = "cluster-autoscaler.kubernetes.io/daemonset-pod"
	// PodTemplateDaemonSetLabelValueTrue use as PodTemplateDaemonSetLabelKey label value.
	PodTemplateDaemonSetLabelValueTrue = "true"
)

// newPodTemplateLister instantiate a new Lister for PodTemplate
func newPodTemplateLister(kubeClient client.Interface, stopchannel <-chan struct{}) v1lister.PodTemplateLister {
	podTemplateWatchOption := func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Everything().String()
		options.LabelSelector = labels.SelectorFromSet(getDaemonsetPodTemplateLabelSet()).String()
	}
	listWatcher := cache.NewFilteredListWatchFromClient(kubeClient.CoreV1().RESTClient(), "podtemplates", apiv1.NamespaceAll, podTemplateWatchOption)
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(listWatcher, &apiv1.PodTemplate{}, time.Hour)
	lister := v1lister.NewPodTemplateLister(store)
	go reflector.Run(stopchannel)
	return lister
}

// getDaemonsetPodTemplateLabelSet returns the labels.Set corresponding to the Daemonset PodTemplate
func getDaemonsetPodTemplateLabelSet() labels.Set {
	daemonsetPodTemplateLabels := map[string]string{
		PodTemplateDaemonSetLabelKey: PodTemplateDaemonSetLabelValueTrue,
	}
	return labels.Set(daemonsetPodTemplateLabels)
}
