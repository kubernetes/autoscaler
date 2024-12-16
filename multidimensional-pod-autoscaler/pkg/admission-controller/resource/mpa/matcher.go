/*
Copyright 2024 The Kubernetes Authors.

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

package mpa

import (
	"context"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_lister "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target"
	mpa_api_util "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/klog/v2"
)

// Matcher is capable of returning a single matching MPA object
// for a pod. Will return nil if no matching object is found.
type Matcher interface {
	GetMatchingMPA(ctx context.Context, pod *core.Pod) *mpa_types.MultidimPodAutoscaler
}

type matcher struct {
	mpaLister         mpa_lister.MultidimPodAutoscalerLister
	selectorFetcher   target.MpaTargetSelectorFetcher
	controllerFetcher controllerfetcher.ControllerFetcher
}

// NewMatcher returns a new MPA matcher.
func NewMatcher(mpaLister mpa_lister.MultidimPodAutoscalerLister,
	selectorFetcher target.MpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher) Matcher {
	return &matcher{mpaLister: mpaLister,
		selectorFetcher:   selectorFetcher,
		controllerFetcher: controllerFetcher}
}

func (m *matcher) GetMatchingMPA(ctx context.Context, pod *core.Pod) *mpa_types.MultidimPodAutoscaler {
	parentController, err := mpa_api_util.FindParentControllerForPod(ctx, pod, m.controllerFetcher)
	if err != nil {
		klog.ErrorS(err, "Failed to get parent controller for pod", "pod", klog.KObj(pod))
		return nil
	}
	if parentController == nil {
		return nil
	}

	configs, err := m.mpaLister.MultidimPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get mpa configs: %v", err)
		return nil
	}

	var controllingMpa *mpa_types.MultidimPodAutoscaler
	for _, mpaConfig := range configs {
		if mpa_api_util.GetUpdateMode(mpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		if mpaConfig.Spec.ScaleTargetRef == nil {
			klog.V(5).InfoS("Skipping MPA object because scaleTargetRef is not defined.", "mpa", klog.KObj(mpaConfig))
			continue
		}
		if mpaConfig.Spec.ScaleTargetRef.Kind != parentController.Kind ||
			mpaConfig.Namespace != parentController.Namespace ||
			mpaConfig.Spec.ScaleTargetRef.Name != parentController.Name {
			continue // This pod is not associated to the right controller
		}

		selector, err := m.selectorFetcher.Fetch(ctx, mpaConfig)
		if err != nil {
			klog.V(3).InfoS("Skipping MPA object because we cannot fetch selector", "mpa", klog.KObj(mpaConfig), "error", err)
			continue
		}

		mpaWithSelector := &mpa_api_util.MpaWithSelector{Mpa: mpaConfig, Selector: selector}
		if mpa_api_util.PodMatchesMPA(pod, mpaWithSelector) && mpa_api_util.Stronger(mpaConfig, controllingMpa) {
			controllingMpa = mpaConfig
		}
	}

	return controllingMpa
}
