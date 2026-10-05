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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Matcher is capable of returning a single matching VPA object
// for a pod. Will return nil if no matching object is found.
type Matcher interface {
	GetMatchingVPA(ctx context.Context, pod *corev1.Pod) *vpa_types.VerticalPodAutoscaler
}

type matcher struct {
	// vpaIndexer must have the vpa_api_util.TargetRefIndex index registered.
	vpaIndexer        cache.Indexer
	selectorFetcher   target.VpaTargetSelectorFetcher
	controllerFetcher controllerfetcher.ControllerFetcher
}

// NewMatcher returns a new VPA matcher.
func NewMatcher(vpaIndexer cache.Indexer,
	selectorFetcher target.VpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher) Matcher {
	return &matcher{vpaIndexer: vpaIndexer,
		selectorFetcher:   selectorFetcher,
		controllerFetcher: controllerFetcher}
}

func (m *matcher) GetMatchingVPA(ctx context.Context, pod *corev1.Pod) *vpa_types.VerticalPodAutoscaler {
	parentController, err := vpa_api_util.FindParentControllerForPod(ctx, pod, m.controllerFetcher)
	if err != nil {
		klog.ErrorS(err, "Failed to get parent controller for pod", "pod", klog.KObj(pod))
		return nil
	}
	if parentController == nil {
		return nil
	}

	configs, err := m.vpaIndexer.ByIndex(vpa_api_util.TargetRefIndex,
		vpa_api_util.TargetRefIndexKey(parentController.Namespace, parentController.Kind, parentController.Name))
	if err != nil {
		klog.ErrorS(err, "Failed to get vpa configs")
		return nil
	}

	var controllingVpa *vpa_types.VerticalPodAutoscaler
	for _, obj := range configs {
		vpaConfig, ok := obj.(*vpa_types.VerticalPodAutoscaler)
		if !ok {
			klog.ErrorS(nil, "Unexpected object type in VPA cache", "object", obj)
			continue
		}
		if vpa_api_util.GetUpdateMode(vpaConfig) == vpa_types.UpdateModeOff && !vpa_api_util.HasStartupBoost(vpaConfig) {
			continue
		}

		selector, err := m.selectorFetcher.Fetch(ctx, vpaConfig)
		if err != nil {
			klog.V(3).InfoS("Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpaConfig), "error", err)
			continue
		}

		vpaWithSelector := &vpa_api_util.VpaWithSelector{Vpa: vpaConfig, Selector: selector}
		if vpa_api_util.PodMatchesVPA(pod, vpaWithSelector) && vpa_api_util.Stronger(vpaConfig, controllingVpa) {
			controllingVpa = vpaConfig
		}
	}

	return controllingVpa
}
