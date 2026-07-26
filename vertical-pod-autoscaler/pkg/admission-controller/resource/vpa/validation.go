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

package vpa

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apires "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

// VPAValidationOptions contains the different settings for VPA validation
type VPAValidationOptions struct {
	IsVPACreate          bool
	AllowCPUStartupBoost bool
	AllowDaemonSetScope  bool
	AllowPerVPAConfig    bool
	AllowInPlace         bool
	// ExistingControlledResources contains the controlled resources already
	// present in the old VPA object, which stay allowed on update even if
	// they wouldn't be accepted on create.
	ExistingControlledResources map[corev1.ResourceName]bool
}

func getValidationOptionsForVPA(oldObj *vpa_types.VerticalPodAutoscaler) VPAValidationOptions {
	opts := VPAValidationOptions{
		IsVPACreate:                 oldObj == nil,
		AllowCPUStartupBoost:        allowCPUBoost(oldObj),
		AllowDaemonSetScope:         allowDaemonSetScope(oldObj),
		AllowPerVPAConfig:           allowPerVPAConfig(oldObj),
		AllowInPlace:                allowInPlace(oldObj),
		ExistingControlledResources: existingControlledResources(oldObj),
	}

	return opts
}

func existingControlledResources(oldObj *vpa_types.VerticalPodAutoscaler) map[corev1.ResourceName]bool {
	resources := map[corev1.ResourceName]bool{}
	if oldObj == nil || oldObj.Spec.ResourcePolicy == nil {
		return resources
	}
	for _, policy := range oldObj.Spec.ResourcePolicy.ContainerPolicies {
		if policy.ControlledResources == nil {
			continue
		}
		for _, resource := range *policy.ControlledResources {
			resources[resource] = true
		}
	}
	return resources
}

func allowCPUBoost(oldObj *vpa_types.VerticalPodAutoscaler) bool {
	if features.Enabled(features.CPUStartupBoost) {
		return true
	}

	if oldObj == nil {
		return false
	}

	if oldObj.Spec.StartupBoost != nil && oldObj.Spec.StartupBoost.CPU != nil {
		return true
	}

	if oldObj.Spec.ResourcePolicy != nil && oldObj.Spec.ResourcePolicy.ContainerPolicies != nil {
		for _, policy := range oldObj.Spec.ResourcePolicy.ContainerPolicies {
			if policy.StartupBoost != nil {
				return true
			}
		}
	}

	return false
}

func allowPerVPAConfig(oldObj *vpa_types.VerticalPodAutoscaler) bool {
	if features.Enabled(features.PerVPAConfig) {
		return true
	}

	if oldObj == nil {
		return false
	}

	if oldObj.Spec.UpdatePolicy != nil && oldObj.Spec.UpdatePolicy.EvictAfterOOMSeconds != nil {
		return true
	}
	if oldObj.Spec.ResourcePolicy != nil && oldObj.Spec.ResourcePolicy.ContainerPolicies != nil {
		for _, policy := range oldObj.Spec.ResourcePolicy.ContainerPolicies {
			if policy.OOMBumpUpRatio != nil || policy.OOMMinBumpUp != nil || policy.MemoryAggregationIntervalCount != nil || policy.MemoryAggregationIntervalSeconds != nil {
				return true
			}
		}
	}

	return false
}

func allowInPlace(oldObj *vpa_types.VerticalPodAutoscaler) bool {
	if features.Enabled(features.InPlace) {
		return true
	}

	if oldObj == nil {
		return false
	}

	if oldObj.Spec.UpdatePolicy != nil && oldObj.Spec.UpdatePolicy.UpdateMode != nil && *oldObj.Spec.UpdatePolicy.UpdateMode == vpa_types.UpdateModeInPlace {
		return true
	}

	return false
}

func allowDaemonSetScope(oldObj *vpa_types.VerticalPodAutoscaler) bool {
	if features.Enabled(features.DaemonSetScope) {
		return true
	}

	if oldObj == nil {
		return false
	}

	return oldObj.Spec.Scope != ""
}

func validateVPA(vpa *vpa_types.VerticalPodAutoscaler, opts VPAValidationOptions) ([]string, field.ErrorList) {
	return validateVPASpec(&vpa.Spec, field.NewPath("spec"), opts)
}

func validateVPASpec(spec *vpa_types.VerticalPodAutoscalerSpec, fldPath *field.Path, opts VPAValidationOptions) ([]string, field.ErrorList) {
	allErrs := field.ErrorList{}
	var warnings []string

	// TODO: Add validation for spec.TargetRef
	if spec.TargetRef == nil && opts.IsVPACreate {
		allErrs = append(allErrs, field.Required(fldPath.Child("targetRef"), "If you're using v1beta1 version of the API, please migrate to v1"))
	}

	if spec.UpdatePolicy != nil {
		updatePolicyWarnings, updatePolicyErrs := validateVPASpecUpdatePolicy(spec.UpdatePolicy, fldPath.Child("updatePolicy"), opts)
		warnings = append(warnings, updatePolicyWarnings...)
		allErrs = append(allErrs, updatePolicyErrs...)
	}

	if spec.ResourcePolicy != nil {
		policyWarnings, policyErrs := validateVPASpecResourcePolicy(spec.ResourcePolicy, fldPath.Child("resourcePolicy"), opts)
		warnings = append(warnings, policyWarnings...)
		allErrs = append(allErrs, policyErrs...)
	}

	if spec.StartupBoost != nil {
		allErrs = append(allErrs, validateVPASpecStartupBoost(spec.StartupBoost, fldPath.Child("startupBoost"), opts)...)
	}

	if spec.Scope != "" {
		if !opts.AllowDaemonSetScope {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("scope"), fmt.Sprintf("not supported when feature flag %s is disabled", features.DaemonSetScope)))
		} else if spec.TargetRef != nil && spec.TargetRef.Kind != "DaemonSet" {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("scope"), spec.Scope, "scope is only supported when targetRef.kind is DaemonSet"))
		}
	}

	if len(spec.Recommenders) > 1 {
		allErrs = append(allErrs, field.TooMany(fldPath.Child("recommenders"), len(spec.Recommenders), 1))
	}

	return warnings, allErrs
}

func validateVPASpecUpdatePolicy(updatePolicy *vpa_types.PodUpdatePolicy, fldPath *field.Path, opts VPAValidationOptions) ([]string, field.ErrorList) {
	allErrs := field.ErrorList{}
	var warnings []string

	mode := updatePolicy.UpdateMode
	if mode == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("updateMode"), "updateMode is required if UpdatePolicy is used"))
	} else {
		if _, found := vpa_types.GetUpdateModes()[*mode]; !found {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("updateMode"), *mode, vpa_types.GetUpdateModesList()))
		}
		if *mode == vpa_types.UpdateModeAuto { //nolint:staticcheck
			warnings = append(warnings, fmt.Sprintf("%s: %q mode is deprecated and will be removed in a future API version. Use explicit update modes like: %s. See https://github.com/kubernetes/autoscaler/issues/8424 for more details.", fldPath, *mode, vpa_types.GetUpdateModesList()))
		}

		if *mode == vpa_types.UpdateModeInPlace && !opts.AllowInPlace {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("updateMode"), fmt.Sprintf("in order to use UpdateMode %s, you must enable feature gate %s in the admission-controller args", vpa_types.UpdateModeInPlace, features.InPlace)))
		}
	}
	if minReplicas := updatePolicy.MinReplicas; minReplicas != nil && *minReplicas <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minReplicas"), *minReplicas, "minReplicas has to be positive"))
	}

	if updatePolicy.EvictAfterOOMSeconds != nil {
		if opts.AllowPerVPAConfig {
			if *updatePolicy.EvictAfterOOMSeconds < 1 {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("evictAfterOOMSeconds"), *updatePolicy.EvictAfterOOMSeconds, "must be greater than or equal to 1"))
			}
		} else {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("evictAfterOOMSeconds"), fmt.Sprintf("not supported when feature flag %s is disabled", features.PerVPAConfig)))
		}
	}

	return warnings, allErrs
}

func validateVPASpecResourcePolicy(resourcePolicy *vpa_types.PodResourcePolicy, fldPath *field.Path, opts VPAValidationOptions) ([]string, field.ErrorList) {
	allErrs := field.ErrorList{}
	var warnings []string

	for i, policy := range resourcePolicy.ContainerPolicies {
		policyPath := fldPath.Child("containerPolicies").Index(i)
		if policy.ContainerName == "" {
			allErrs = append(allErrs, field.Required(policyPath.Child("containerName"), ""))
		}

		if policy.Mode != nil {
			if _, found := vpa_types.GetScalingModes()[*policy.Mode]; !found {
				allErrs = append(allErrs, field.NotSupported(policyPath.Child("mode"), *policy.Mode, vpa_types.GetPossibleScalingModes()))
			}
		}

		for resource, minAllowed := range policy.MinAllowed {
			allErrs = append(allErrs, validateResourceResolution(policyPath.Child("minAllowed").Key(string(resource)), resource, minAllowed)...)
			maxAllowed, found := policy.MaxAllowed[resource]
			if found && maxAllowed.Cmp(minAllowed) < 0 {
				allErrs = append(allErrs, field.Invalid(policyPath.Child("maxAllowed").Key(string(resource)), maxAllowed.String(), fmt.Sprintf("max resource for %v is lower than min of \"%v\"", resource, minAllowed.String())))
			}
		}

		for resource, max := range policy.MaxAllowed {
			allErrs = append(allErrs, validateResourceResolution(policyPath.Child("maxAllowed").Key(string(resource)), resource, max)...)
		}

		controlledValues := policy.ControlledValues
		if policy.Mode != nil && controlledValues != nil {
			if *policy.Mode == vpa_types.ContainerScalingModeOff && *controlledValues == vpa_types.ContainerControlledValuesRequestsAndLimits {
				allErrs = append(allErrs, field.Forbidden(policyPath.Child("controlledValues"), "controlledValues shouldn't be specified if container scaling mode is off"))
			}
		}

		if policy.ControlledResources != nil {
			for j, resource := range *policy.ControlledResources {
				if resource == corev1.ResourceCPU || resource == corev1.ResourceMemory {
					continue
				}
				resourcePath := policyPath.Child("controlledResources").Index(j)
				if opts.ExistingControlledResources[resource] {
					warnings = append(warnings, fmt.Sprintf("%s: unsupported value %q is allowed only because it is present in the existing VPA object; supported values: %q, %q", resourcePath, resource, corev1.ResourceCPU, corev1.ResourceMemory))
				} else {
					allErrs = append(allErrs, field.NotSupported(resourcePath, resource, []string{string(corev1.ResourceCPU), string(corev1.ResourceMemory)}))
				}
			}
		}

		if policy.OOMBumpUpRatio != nil {
			if opts.AllowPerVPAConfig {
				ratio := float64(policy.OOMBumpUpRatio.MilliValue()) / 1000.0
				if ratio < 1.0 {
					allErrs = append(allErrs, field.Invalid(policyPath.Child("oomBumpUpRatio"), ratio, "must be greater than or equal to 1.0"))
				}
			} else {
				allErrs = append(allErrs, field.Forbidden(policyPath.Child("oomBumpUpRatio"), fmt.Sprintf("not supported when feature flag %s is disabled", features.PerVPAConfig)))
			}
		}

		if policy.OOMMinBumpUp != nil {
			if opts.AllowPerVPAConfig {
				minBump := policy.OOMMinBumpUp.Value()
				if minBump < 0 {
					allErrs = append(allErrs, field.Invalid(policyPath.Child("oomMinBumpUp"), minBump, "must be greater than or equal to 0"))
				}
			} else {
				allErrs = append(allErrs, field.Forbidden(policyPath.Child("oomMinBumpUp"), fmt.Sprintf("not supported when feature flag %s is disabled", features.PerVPAConfig)))
			}
		}

		if policy.MemoryAggregationIntervalSeconds != nil {
			if opts.AllowPerVPAConfig {
				seconds := *policy.MemoryAggregationIntervalSeconds
				if seconds < 1 {
					allErrs = append(allErrs, field.Invalid(policyPath.Child("memoryAggregationIntervalSeconds"), seconds, "must be greater than or equal to 1"))
				}
			} else {
				allErrs = append(allErrs, field.Forbidden(policyPath.Child("memoryAggregationIntervalSeconds"), fmt.Sprintf("not supported when feature flag %s is disabled", features.PerVPAConfig)))
			}
		}

		if policy.MemoryAggregationIntervalCount != nil {
			if opts.AllowPerVPAConfig {
				count := *policy.MemoryAggregationIntervalCount
				if count < 1 {
					allErrs = append(allErrs, field.Invalid(policyPath.Child("memoryAggregationIntervalCount"), count, "must be greater than or equal to 1"))
				}
			} else {
				allErrs = append(allErrs, field.Forbidden(policyPath.Child("memoryAggregationIntervalCount"), fmt.Sprintf("not supported when feature flag %s is disabled", features.PerVPAConfig)))
			}
		}

		if policy.StartupBoost != nil {
			allErrs = append(allErrs, validateVPASpecStartupBoost(policy.StartupBoost, policyPath.Child("startupBoost"), opts)...)
		}
	}

	return warnings, allErrs
}

func validateVPASpecStartupBoost(startupBoost *vpa_types.StartupBoost, fldPath *field.Path, opts VPAValidationOptions) field.ErrorList {
	allErrs := field.ErrorList{}

	if !opts.AllowCPUStartupBoost {
		return append(allErrs, field.Forbidden(fldPath, fmt.Sprintf("in order to use startupBoost, you must enable feature gate %s in the admission-controller args", features.CPUStartupBoost)))
	}

	cpuBoost := startupBoost.CPU
	if cpuBoost == nil {
		return allErrs
	}
	cpuPath := fldPath.Child("cpu")
	boostType := cpuBoost.Type
	if boostType == "" {
		allErrs = append(allErrs, field.Required(cpuPath.Child("type"), fmt.Sprintf("must be either %s or %s", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType)))
		return allErrs
	}

	switch boostType {
	case vpa_types.FactorStartupBoostType:
		if cpuBoost.Factor == nil {
			allErrs = append(allErrs, field.Required(cpuPath.Child("factor"), "required when type is Factor"))
		} else if *cpuBoost.Factor < 1 {
			allErrs = append(allErrs, field.Invalid(cpuPath.Child("factor"), *cpuBoost.Factor, "must be >= 1 for type Factor"))
		}
	case vpa_types.QuantityStartupBoostType:
		if cpuBoost.Quantity == nil {
			allErrs = append(allErrs, field.Required(cpuPath.Child("quantity"), "required when type is Quantity"))
		} else {
			allErrs = append(allErrs, validateCPUResolution(*cpuBoost.Quantity, cpuPath.Child("quantity"))...)
		}
	default:
		allErrs = append(allErrs, field.NotSupported(cpuPath.Child("type"), boostType, []string{string(vpa_types.FactorStartupBoostType), string(vpa_types.QuantityStartupBoostType)}))
	}
	return allErrs
}

func validateResourceResolution(fldPath *field.Path, name corev1.ResourceName, val apires.Quantity) field.ErrorList {
	switch name {
	case corev1.ResourceCPU:
		return validateCPUResolution(val, fldPath)
	case corev1.ResourceMemory:
		return validateMemoryResolution(val, fldPath)
	}
	return nil
}

func validateCPUResolution(val apires.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if _, precissionPreserved := val.AsScale(apires.Milli); !precissionPreserved {
		allErrs = append(allErrs, field.Invalid(fldPath, val.String(), "must be a whole number of milli CPUs"))
	}
	return allErrs
}

func validateMemoryResolution(val apires.Quantity, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if _, precissionPreserved := val.AsScale(0); !precissionPreserved {
		allErrs = append(allErrs, field.Invalid(fldPath, val.String(), "must be a whole number of bytes"))
	}
	return allErrs
}
