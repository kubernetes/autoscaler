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
	"k8s.io/apimachinery/pkg/labels"
	vpa_types_v1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/klog"
)

type compositeTargetSelectorFetcher struct {
	primary, backup VpaTargetSelectorFetcher
}

// NewCompositeTargetSelectorFetcher returns a new VpaTargetSelectorFetcher that uses primary
// VpaTargetSelectorFetcher by default and when it fails tries backup VpaTargetSelectorFetcher.
func NewCompositeTargetSelectorFetcher(primary, backup VpaTargetSelectorFetcher) VpaTargetSelectorFetcher {
	return &compositeTargetSelectorFetcher{
		primary: primary,
		backup:  backup,
	}
}

func (f *compositeTargetSelectorFetcher) Fetch(vpa *vpa_types_v1beta2.VerticalPodAutoscaler) (labels.Selector, error) {
	primarySelector, primaryErr := f.primary.Fetch(vpa)
	if primaryErr == nil {
		return primarySelector, primaryErr
	}
	klog.Errorf("Primary VpaTargetSelectorFetcher failed. Err: %v", primaryErr)
	backupSelector, _ := f.backup.Fetch(vpa)
	if backupSelector != nil {
		return backupSelector, nil
	}
	return nil, primaryErr
}
