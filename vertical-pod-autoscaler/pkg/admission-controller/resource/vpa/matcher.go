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
	finder := vpa_api_util.NewControllingVPAFinder(m.vpaIndexer, m.selectorFetcher, m.controllerFetcher, admitsPods)
	vpa, _ := finder.GetControllingVPAForPod(ctx, pod)
	return vpa
}

// admitsPods reports whether the admission controller should act on Pods
// controlled by the VPA.
func admitsPods(vpa *vpa_types.VerticalPodAutoscaler) bool {
	return vpa_api_util.GetEffectiveUpdateMode(vpa) != vpa_types.UpdateModeOff || vpa_api_util.HasStartupBoost(vpa)
}
