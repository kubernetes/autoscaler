/*
Copyright The Kubernetes Authors.

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

package api

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
)

// VPAFilter reports whether a VPA is eligible to control Pods. VPAs for which
// it returns false are ignored by ControllingVPAFinder.
type VPAFilter func(*vpa_types.VerticalPodAutoscaler) bool

// ControllingVPAFinder finds the VPA controlling a given Pod. Only VPAs whose
// targetRef points to the Pod's parent controller are considered, so the cost
// of a lookup is independent of the total number of VPAs in the cluster.
//
// The selector of each candidate VPA is resolved through selectorFetcher on
// every lookup. Fetching a selector may involve an API call, so callers doing
// many lookups should pass a fetcher serving selectors fetched up front.
type ControllingVPAFinder struct {
	// vpaIndexer must have the TargetRefIndex index registered.
	vpaIndexer        cache.Indexer
	selectorFetcher   target.VpaTargetSelectorFetcher
	controllerFetcher controllerfetcher.ControllerFetcher
	// filter is optional; a nil filter accepts every VPA.
	filter VPAFilter
}

// NewControllingVPAFinder returns a ControllingVPAFinder looking up VPAs in
// vpaIndexer, which must have the TargetRefIndex index registered. Only VPAs
// accepted by filter are considered; a nil filter accepts every VPA.
func NewControllingVPAFinder(vpaIndexer cache.Indexer, selectorFetcher target.VpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher, filter VPAFilter) *ControllingVPAFinder {
	return &ControllingVPAFinder{
		vpaIndexer:        vpaIndexer,
		selectorFetcher:   selectorFetcher,
		controllerFetcher: controllerFetcher,
		filter:            filter,
	}
}

// GetControllingVPAForPod returns the strongest (see Stronger) VPA that targets
// the Pod's parent controller, is accepted by the filter and whose selector
// matches the Pod's labels, along with that selector. Returns a nil VPA if
// there is no such VPA.
func (f *ControllingVPAFinder) GetControllingVPAForPod(ctx context.Context, pod *corev1.Pod) (*vpa_types.VerticalPodAutoscaler, labels.Selector) {
	parentController, err := FindParentControllerForPod(ctx, pod, f.controllerFetcher)
	if err != nil {
		klog.ErrorS(err, "Failed to get parent controller for pod", "pod", klog.KObj(pod))
		return nil, nil
	}
	if parentController == nil {
		return nil, nil
	}

	candidates, err := f.vpaIndexer.ByIndex(TargetRefIndex,
		TargetRefIndexKey(parentController.Namespace, parentController.Kind, parentController.Name))
	if err != nil {
		klog.ErrorS(err, "Failed to get VPAs by targetRef", "pod", klog.KObj(pod))
		return nil, nil
	}

	podLabels := labels.Set(pod.GetLabels())
	var controlling *vpa_types.VerticalPodAutoscaler
	var controllingSelector labels.Selector
	for _, obj := range candidates {
		vpa, ok := obj.(*vpa_types.VerticalPodAutoscaler)
		if !ok {
			klog.ErrorS(nil, "Unexpected object type in VPA cache", "object", obj)
			continue
		}
		if f.filter != nil && !f.filter(vpa) {
			continue
		}
		// Check strength before fetching the selector, as fetching may be expensive.
		if !Stronger(vpa, controlling) {
			continue
		}
		selector, err := f.selectorFetcher.Fetch(ctx, vpa)
		if err != nil {
			klog.V(3).InfoS("Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpa), "error", err)
			continue
		}
		if PodLabelsMatchVPA(pod.Namespace, podLabels, vpa.Namespace, selector) {
			controlling, controllingSelector = vpa, selector
		}
	}
	return controlling, controllingSelector
}
