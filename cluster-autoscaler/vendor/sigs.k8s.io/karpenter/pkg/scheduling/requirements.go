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

package scheduling

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/awslabs/operatorpkg/option"
	"github.com/samber/lo"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// Requirements are an efficient set representation under the hood. Since its underlying
// types are slices and maps, this type should not be used as a pointer.
type Requirements map[string]*Requirement

func NewRequirements(requirements ...*Requirement) Requirements {
	r := Requirements{}
	for _, requirement := range requirements {
		r.Add(requirement)
	}
	return r
}

// NewRequirements constructs requirements from NodeSelectorRequirementWithMinValues
func NewNodeSelectorRequirementsWithMinValues(requirements ...v1.NodeSelectorRequirementWithMinValues) Requirements {
	r := NewRequirements()
	for _, requirement := range requirements {
		r.Add(NewRequirementWithFlexibility(requirement.Key, requirement.Operator, requirement.MinValues, requirement.Values...))
	}
	return r
}

// NewRequirements constructs requirements from NodeSelectorRequirements
func NewNodeSelectorRequirements(requirements ...corev1.NodeSelectorRequirement) Requirements {
	r := NewRequirements()
	for _, requirement := range requirements {
		r.Add(NewRequirementWithFlexibility(requirement.Key, requirement.Operator, nil, requirement.Values...))
	}
	return r
}

// NewLabelRequirements constructs requirements from labels
func NewLabelRequirements(labels map[string]string) Requirements {
	requirements := NewRequirements()
	for key, value := range labels {
		requirements.Add(NewRequirement(key, corev1.NodeSelectorOpIn, value))
	}
	return requirements
}

// NewPodRequirements constructs requirements from a pod and treats any preferred requirements as required.
func NewPodRequirements(pod *corev1.Pod) Requirements {
	return newPodRequirements(pod, podRequirementTypeAll)
}

// NewStrictPodRequirements constructs requirements from a pod and only includes true requirements (not preferences).
func NewStrictPodRequirements(pod *corev1.Pod) Requirements {
	return newPodRequirements(pod, podRequirementTypeRequiredOnly)
}

type podRequirementType byte

const (
	podRequirementTypeAll = iota
	podRequirementTypeRequiredOnly
)

func newPodRequirements(pod *corev1.Pod, typ podRequirementType) Requirements {
	requirements := NewLabelRequirements(pod.Spec.NodeSelector)
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil {
		return requirements
	}
	if typ == podRequirementTypeAll {
		// The legal operators for pod affinity and anti-affinity are In, NotIn, Exists, DoesNotExist.
		// Select heaviest preference and treat as a requirement. An outer loop will iteratively unconstrain them if unsatisfiable.
		if preferred := pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution; len(preferred) > 0 {
			sort.Slice(preferred, func(i int, j int) bool { return preferred[i].Weight > preferred[j].Weight })
			requirements.Add(NewNodeSelectorRequirements(preferred[0].Preference.MatchExpressions...).Values()...)
		}
	}

	// Select first requirement. An outer loop will iteratively remove OR requirements if unsatisfiable
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil &&
		len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) > 0 {
		requirements.Add(NewNodeSelectorRequirements(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions...).Values()...)
	}
	return requirements
}

// HasPreferredNodeAffinity returns true if the pod has a preferred node affinity term
func HasPreferredNodeAffinity(p *corev1.Pod) bool {
	if p == nil {
		return false
	}
	return p.Spec.Affinity != nil && p.Spec.Affinity.NodeAffinity != nil && len(p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0
}

func (r Requirements) NodeSelectorRequirements() []v1.NodeSelectorRequirementWithMinValues {
	result := make([]v1.NodeSelectorRequirementWithMinValues, 0, len(r))
	for _, req := range r {
		if req.gte != nil && req.lte != nil {
			result = append(result, req.BoundedNodeSelectorRequirements()...)
		} else {
			result = append(result, req.NodeSelectorRequirement())
		}
	}
	return result
}

// Add requirements to provided requirements. Mutates existing requirements
func (r Requirements) Add(requirements ...*Requirement) {
	for _, requirement := range requirements {
		if existing, ok := r[requirement.Key]; ok {
			requirement = requirement.Intersection(existing)
		}
		r[requirement.Key] = requirement
	}
}

// Keys returns unique set of the label keys from the requirements
func (r Requirements) Keys() sets.Set[string] {
	keys := sets.New[string]()
	for key := range r {
		keys.Insert(key)
	}
	return keys
}

func (r Requirements) Values() []*Requirement {
	return lo.Values(r)
}

func (r Requirements) Has(key string) bool {
	_, ok := r[key]
	return ok
}

func (r Requirements) Get(key string) *Requirement {
	if _, ok := r[key]; !ok {
		// If not defined, allow any values with the exists operator
		return NewRequirement(key, corev1.NodeSelectorOpExists)
	}
	return r[key]
}

type CompatibilityOptions struct {
	AllowUndefined sets.Set[string]
}

var AllowUndefinedWellKnownLabels = func(options *CompatibilityOptions) {
	options.AllowUndefined = v1.WellKnownLabels
}

func (r Requirements) IsCompatible(requirements Requirements, options ...option.Function[CompatibilityOptions]) bool {
	return r.Compatible(requirements, options...) == nil
}

// Compatible ensures the provided requirements can loosely be met.
func (r Requirements) Compatible(requirements Requirements, options ...option.Function[CompatibilityOptions]) error {
	opts := option.Resolve(options...)

	// Custom Labels must intersect, but if not defined are denied.
	for key := range requirements {
		if opts.AllowUndefined.Has(key) {
			continue
		}
		if operator := requirements.Get(key).Operator(); r.Has(key) || operator == corev1.NodeSelectorOpNotIn || operator == corev1.NodeSelectorOpDoesNotExist {
			continue
		}
		// break early so we only report the first error
		return fmt.Errorf("label %q does not have known values%s", key, labelHint(r, key, opts.AllowUndefined))
	}
	// Well Known Labels must intersect, but if not defined, are allowed.
	return r.Intersects(requirements)
}

func getSuffix(key string) string {
	before, after, found := strings.Cut(key, "/")
	return lo.Ternary(found, after, before)
}

func labelHint(r Requirements, key string, allowedUndefined sets.Set[string]) string {
	for wellKnown := range allowedUndefined {
		if strings.Contains(wellKnown, key) {
			return fmt.Sprintf(" (typo of %q?)", wellKnown)
		}
		if strings.HasSuffix(wellKnown, getSuffix(key)) {
			return fmt.Sprintf(" (typo of %q?)", wellKnown)
		}
	}
	for existing := range r {
		if strings.Contains(existing, key) {
			return fmt.Sprintf(" (typo of %q?)", existing)
		}
		if strings.HasSuffix(existing, getSuffix(key)) {
			return fmt.Sprintf(" (typo of %q?)", existing)
		}
	}
	return ""
}

// badKeyError allows lazily generating the error string in the case of a bad key error. When requirements fail
// to match, we are most often interested in the failure and not why it fails.
type badKeyError struct {
	key      string
	incoming *Requirement
	existing *Requirement
}

func (b badKeyError) Error() string {
	return fmt.Sprintf("key %s, %s not in %s", b.key, b.incoming, b.existing)
}

// intersectKeys is much faster and allocates less han getting the two key sets separately and intersecting them
func (r Requirements) intersectKeys(rhs Requirements) sets.Set[string] {
	smallest := r
	largest := rhs
	if len(smallest) > len(largest) {
		smallest, largest = largest, smallest
	}
	keys := sets.Set[string]{}

	for key := range smallest {
		if _, ok := largest[key]; ok {
			keys.Insert(key)
		}
	}
	return keys
}

// Intersects returns errors if the requirements don't have overlapping values, undefined keys are allowed
func (r Requirements) Intersects(requirements Requirements) (errs error) {
	for key := range r.intersectKeys(requirements) {
		existing := r.Get(key)
		incoming := requirements.Get(key)
		if !existing.HasIntersection(incoming) {
			// where the incoming requirement has operator { NotIn, DoesNotExist }
			if operator := incoming.Operator(); operator == corev1.NodeSelectorOpNotIn || operator == corev1.NodeSelectorOpDoesNotExist {
				// and the existing requirement has operator { NotIn, DoesNotExist }
				if operator := existing.Operator(); operator == corev1.NodeSelectorOpNotIn || operator == corev1.NodeSelectorOpDoesNotExist {
					continue
				}
			}
			errs = multierr.Append(errs, badKeyError{
				key:      key,
				incoming: incoming,
				existing: existing,
			})
		}
	}
	return errs
}

func (r Requirements) HasMinValues() bool {
	for _, req := range r {
		if req.MinValues != nil {
			return true
		}
	}
	return false
}

func (r Requirements) String() string {
	requirements := lo.Reject(r.Values(), func(requirement *Requirement, _ int) bool {
		return v1.RestrictedLabels.Has(requirement.Key)
	})
	stringRequirements := lo.Map(requirements, func(requirement *Requirement, _ int) string { return requirement.String() })
	slices.Sort(stringRequirements)
	return strings.Join(stringRequirements, ", ")
}
