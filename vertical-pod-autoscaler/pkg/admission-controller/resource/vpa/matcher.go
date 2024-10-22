/*
Copyright 2020 The Kubernetes Authors.

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

package vpa

import (
	"context"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Matcher is capable of returning a single matching VPA object
// for a pod. Will return nil if no matching object is found.
type Matcher interface {
	GetMatchingVPA(ctx context.Context, pod *core.Pod) *vpa_types.VerticalPodAutoscaler
}

type matcher struct {
	vpaLister         vpa_lister.VerticalPodAutoscalerLister
	selectorFetcher   target.VpaTargetSelectorFetcher
	controllerFetcher controllerfetcher.ControllerFetcher
}

// NewMatcher returns a new VPA matcher.
func NewMatcher(vpaLister vpa_lister.VerticalPodAutoscalerLister,
	selectorFetcher target.VpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher) Matcher {
	return &matcher{vpaLister: vpaLister,
		selectorFetcher:   selectorFetcher,
		controllerFetcher: controllerFetcher}
}

func (m *matcher) GetMatchingVPA(ctx context.Context, pod *core.Pod) *vpa_types.VerticalPodAutoscaler {
	configs, err := m.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Failed to get vpa configs")
		return nil
	}
	onConfigs := make([]*vpa_api_util.VpaWithSelector, 0)
	for _, vpaConfig := range configs {
		if vpa_api_util.GetUpdateMode(vpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		selector, err := m.selectorFetcher.Fetch(ctx, vpaConfig)
		if err != nil {
			klog.V(3).InfoS("Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpaConfig), "error", err)
			continue
		}
		onConfigs = append(onConfigs, &vpa_api_util.VpaWithSelector{
			Vpa:      vpaConfig,
			Selector: selector,
		})
	}
	klog.V(2).InfoS("Let's choose from", "configs", len(onConfigs), "pod", klog.KObj(pod))
	result := vpa_api_util.GetControllingVPAForPod(ctx, pod, onConfigs, m.controllerFetcher)
	if result != nil {
		return result.Vpa
	}
	return nil
}
