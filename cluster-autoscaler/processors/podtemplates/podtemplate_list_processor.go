/*
Copyright 2018 The Kubernetes Authors.

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

// Package podtemplates contains PodTemplate processor used to simulate
// DaemonSet resources from a PodTemplate.
package podtemplates

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
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

// PodTemplateListProcessor processes lists of unschedulable pods.
type PodTemplateListProcessor interface {
	ExtraDaemonsets(context *ca_context.AutoscalingContext) ([]*appsv1.DaemonSet, error)
	CleanUp()
}

// activePodTemplateListProcessor returning pod list generated from PodTemplate instance.
type activePodTemplateListProcessor struct {
	podTemplateLister v1lister.PodTemplateLister
	store             cache.Store
	reflector         *cache.Reflector

	ctx        context.Context
	cancelFunc func()
}

// NewActivePodTemplateListProcessor creates an instance of PodTemplateListProcessor.
func NewActivePodTemplateListProcessor(kubeClient client.Interface) PodTemplateListProcessor {
	internalContext, cancelFunc := context.WithCancel(context.Background())

	return &activePodTemplateListProcessor{
		ctx:               internalContext,
		cancelFunc:        cancelFunc,
		podTemplateLister: newPodTemplateLister(kubeClient, internalContext.Done()),
	}
}

// Process processes lists of extra Daemonset before simulating Daemonset pod node usage.
func (p *activePodTemplateListProcessor) ExtraDaemonsets(
	context *ca_context.AutoscalingContext) ([]*appsv1.DaemonSet, error) {
	// Get all PodTpls since we already filter pods in the listner
	podTpls, err := p.podTemplateLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	result := make([]*appsv1.DaemonSet, 0, len(podTpls))
	for _, podTpl := range podTpls {
		result = append(result, newDaemonSet(podTpl))
	}

	return result, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *activePodTemplateListProcessor) CleanUp() {
	p.cancelFunc()
}

//
// noOpPodTemplateListProcessor is returning pod lists without processing them.
type noOpPodTemplateListProcessor struct{}

// NewDefaultPodTemplateListProcessor creates an instance of PodTemplateListProcessor.
func NewDefaultPodTemplateListProcessor() PodTemplateListProcessor {
	return &noOpPodTemplateListProcessor{}
}

// Process processes lists of extra Daemonset before simulating Daemonset pod node usage.
func (p *noOpPodTemplateListProcessor) ExtraDaemonsets(context *ca_context.AutoscalingContext) ([]*appsv1.DaemonSet, error) {
	return nil, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *noOpPodTemplateListProcessor) CleanUp() {
}

// newDaemonSetLister builds a podTemplate lister.
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

func newDaemonSet(podTemplate *apiv1.PodTemplate) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tpl", podTemplate.Name),
			Namespace: podTemplate.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: podTemplate.Template,
		},
	}
}

func getDaemonsetPodTemplateLabelSet() labels.Set {
	daemonsetPodTemplateLabels := map[string]string{
		PodTemplateDaemonSetLabelKey: PodTemplateDaemonSetLabelValueTrue,
	}
	return labels.Set(daemonsetPodTemplateLabels)
}
