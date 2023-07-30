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

package mpa

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_lister "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target"
	mpa_api_util "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/klog/v2"
)

// Matcher is capable of returning a single matching MPA object
// for a pod. Will return nil if no matching object is found.
type Matcher interface {
	GetMatchingMPA(pod *core.Pod) *mpa_types.MultidimPodAutoscaler
}

type matcher struct {
	mpaLister       mpa_lister.MultidimPodAutoscalerLister
	selectorFetcher target.MpaTargetSelectorFetcher
}

// NewMatcher returns a new MPA matcher.
func NewMatcher(mpaLister mpa_lister.MultidimPodAutoscalerLister,
	selectorFetcher target.MpaTargetSelectorFetcher) Matcher {
	return &matcher{mpaLister: mpaLister,
		selectorFetcher: selectorFetcher}
}

func (m *matcher) GetMatchingMPA(pod *core.Pod) *mpa_types.MultidimPodAutoscaler {
	configs, err := m.mpaLister.MultidimPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get mpa configs: %v", err)
		return nil
	}
	onConfigs := make([]*mpa_api_util.MpaWithSelector, 0)
	for _, mpaConfig := range configs {
		if mpa_api_util.GetUpdateMode(mpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		selector, err := m.selectorFetcher.Fetch(mpaConfig)
		if err != nil {
			klog.V(3).Infof("skipping MPA object %v because we cannot fetch selector: %s", mpaConfig.Name, err)
			continue
		}
		onConfigs = append(onConfigs, &mpa_api_util.MpaWithSelector{
			Mpa:      mpaConfig,
			Selector: selector,
		})
	}
	klog.V(2).Infof("Let's choose from %d configs for pod %s/%s", len(onConfigs), pod.Namespace, pod.Name)
	result := mpa_api_util.GetControllingMPAForPod(pod, onConfigs)
	if result != nil {
		return result.Mpa
	}
	return nil
}
