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

package v1alpha1

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// RuntimeValidate will be used to validate any part of the CRD that can not be validated at CRD creation
func (in *NodeOverlay) RuntimeValidate(ctx context.Context) error {
	return multierr.Combine(in.Spec.validateRequirements(ctx), in.Spec.validateCapacity())
}

// This function is used by the NodeOverlay validation webhook to verify the nodeoverlay requirements.
// When this function is called, the nodeoverlay's requirements do not include the requirements from labels.
// NodeOverlay requirements only support well known labels.
func (in *NodeOverlaySpec) validateRequirements(ctx context.Context) (errs error) {
	for _, requirement := range in.Requirements {
		if err := v1.ValidateRequirement(ctx, v1.NodeSelectorRequirementWithMinValues{Key: requirement.Key, Operator: requirement.Operator, Values: requirement.Values}); err != nil {
			errs = multierr.Append(errs, fmt.Errorf("invalid value: %w in requirements, restricted", err))
		}
		if requirement.Operator == corev1.NodeSelectorOpNotIn && len(requirement.Values) == 0 {
			errs = multierr.Append(errs, fmt.Errorf("key %s with operator %s must have a value defined", requirement.Key, requirement.Operator))
		}
	}
	return errs
}

func (in *NodeOverlaySpec) validateCapacity() (errs error) {
	for n := range in.Capacity {
		if v1.WellKnownResources.Has(n) {
			errs = multierr.Append(errs, fmt.Errorf("invalid capacity: %s in resource, restricted", n))
		}
	}
	return errs
}
