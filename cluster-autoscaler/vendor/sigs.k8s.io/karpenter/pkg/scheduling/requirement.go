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
	"math"
	"math/rand"
	"strconv"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

//go:generate go tool -modfile=../../go.tools.mod controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="."

// Requirement is an efficient represenatation of corev1.NodeSelectorRequirement
// +k8s:deepcopy-gen=true
type Requirement struct {
	Key        string
	complement bool
	values     sets.Set[string]
	gte        *int // inclusive lower bound (Gt is converted to Gte)
	lte        *int // inclusive upper bound (Lt is converted to Lte)
	MinValues  *int
}

// NewRequirementWithFlexibility constructs new requirement from the combination of key, values, minValues and the operator that
// connects the keys and values. GT and LT operators are canonicalized to GTE and LTE respectively.
// nolint:gocyclo
func NewRequirementWithFlexibility(key string, operator corev1.NodeSelectorOperator, minValues *int, values ...string) *Requirement {
	if normalized, ok := v1.NormalizedLabels[key]; ok {
		key = normalized
	}

	// This is a super-common case, so optimize for it an inline everything.
	if operator == corev1.NodeSelectorOpIn {
		s := make(sets.Set[string], len(values))
		for _, value := range values {
			s[value] = sets.Empty{}
		}
		return &Requirement{
			Key:        key,
			values:     s,
			complement: false,
			MinValues:  minValues,
		}
	}

	r := &Requirement{
		Key:        key,
		values:     sets.New[string](),
		complement: true,
		MinValues:  minValues,
	}
	if operator == corev1.NodeSelectorOpIn || operator == corev1.NodeSelectorOpDoesNotExist {
		r.complement = false
	}
	if operator == corev1.NodeSelectorOpIn || operator == corev1.NodeSelectorOpNotIn {
		r.values.Insert(values...)
	}
	if operator == corev1.NodeSelectorOpGt {
		value, _ := strconv.Atoi(values[0]) // prevalidated
		if value == math.MaxInt {
			// Gt MaxInt matches nothing
			return NewRequirement(key, corev1.NodeSelectorOpDoesNotExist)
		}
		value++ // canonicalize GT N to GTE N+1
		r.gte = &value
	}
	if operator == corev1.NodeSelectorOpLt {
		value, _ := strconv.Atoi(values[0]) // prevalidated
		value--                             // canonicalize LT N to LTE N-1
		r.lte = &value
	}
	if operator == v1.NodeSelectorOpGte {
		value, _ := strconv.Atoi(values[0]) // prevalidated
		r.gte = &value
	}
	if operator == v1.NodeSelectorOpLte {
		value, _ := strconv.Atoi(values[0]) // prevalidated
		r.lte = &value
	}
	return r
}

func NewRequirement(key string, operator corev1.NodeSelectorOperator, values ...string) *Requirement {
	return NewRequirementWithFlexibility(key, operator, nil, values...)
}

// BoundedNodeSelectorRequirements handles the case where both gte and lte exist.
// Unlike other operators which intersect into a single values set, bounds are stored
// separately and must be serialized as two distinct NodeSelectorRequirements.
// This method is separate from NodeSelectorRequirement() to avoid slice allocations
// in the common case where only one bound exists.
func (r *Requirement) BoundedNodeSelectorRequirements() []v1.NodeSelectorRequirementWithMinValues {
	return []v1.NodeSelectorRequirementWithMinValues{
		{Key: r.Key, Operator: v1.NodeSelectorOpGte, Values: []string{strconv.FormatInt(int64(*r.gte), 10)}, MinValues: r.MinValues},
		{Key: r.Key, Operator: v1.NodeSelectorOpLte, Values: []string{strconv.FormatInt(int64(*r.lte), 10)}, MinValues: r.MinValues},
	}
}

func (r *Requirement) NodeSelectorRequirement() v1.NodeSelectorRequirementWithMinValues {
	switch {
	case r.gte != nil:
		return v1.NodeSelectorRequirementWithMinValues{
			Key:       r.Key,
			Operator:  v1.NodeSelectorOpGte,
			Values:    []string{strconv.FormatInt(int64(lo.FromPtr(r.gte)), 10)},
			MinValues: r.MinValues,
		}
	case r.lte != nil:
		return v1.NodeSelectorRequirementWithMinValues{
			Key:       r.Key,
			Operator:  v1.NodeSelectorOpLte,
			Values:    []string{strconv.FormatInt(int64(lo.FromPtr(r.lte)), 10)},
			MinValues: r.MinValues,
		}
	case r.complement:
		switch {
		case len(r.values) > 0:
			return v1.NodeSelectorRequirementWithMinValues{
				Key:       r.Key,
				Operator:  corev1.NodeSelectorOpNotIn,
				Values:    sets.List(r.values),
				MinValues: r.MinValues,
			}
		default:
			return v1.NodeSelectorRequirementWithMinValues{
				Key:       r.Key,
				Operator:  corev1.NodeSelectorOpExists,
				MinValues: r.MinValues,
			}
		}
	default:
		switch {
		case len(r.values) > 0:
			return v1.NodeSelectorRequirementWithMinValues{
				Key:       r.Key,
				Operator:  corev1.NodeSelectorOpIn,
				Values:    sets.List(r.values),
				MinValues: r.MinValues,
			}
		default:
			return v1.NodeSelectorRequirementWithMinValues{
				Key:       r.Key,
				Operator:  corev1.NodeSelectorOpDoesNotExist,
				MinValues: r.MinValues,
			}
		}
	}
}

// Intersection constraints the Requirement from the incoming requirements
// nolint:gocyclo
func (r *Requirement) Intersection(requirement *Requirement) *Requirement {
	// Complement
	complement := r.complement && requirement.complement

	// Boundaries
	gte := maxIntPtr(r.gte, requirement.gte)
	lte := minIntPtr(r.lte, requirement.lte)
	minValues := maxIntPtr(r.MinValues, requirement.MinValues)
	if gte != nil && lte != nil && *gte > *lte {
		return NewRequirementWithFlexibility(r.Key, corev1.NodeSelectorOpDoesNotExist, minValues)
	}

	// Values
	var values sets.Set[string]
	if r.complement && requirement.complement {
		values = r.values.Union(requirement.values)
	} else if r.complement && !requirement.complement {
		values = requirement.values.Difference(r.values)
	} else if !r.complement && requirement.complement {
		values = r.values.Difference(requirement.values)
	} else {
		values = r.values.Intersection(requirement.values)
	}
	for v := range values {
		if !withinBounds(v, gte, lte) {
			values.Delete(v)
		}
	}
	// Remove boundaries for concrete sets
	if !complement {
		gte, lte = nil, nil
	}
	return &Requirement{Key: r.Key, values: values, complement: complement, gte: gte, lte: lte, MinValues: minValues}
}

// nolint:gocyclo
// HasIntersection is a more efficient implementation of Intersection
// It validates whether there is an intersection between the two requirements without actually creating the sets
// This prevents the garbage collector from having to spend cycles cleaning up all of these created set objects
func (r *Requirement) HasIntersection(requirement *Requirement) bool {
	gte := maxIntPtr(r.gte, requirement.gte)
	lte := minIntPtr(r.lte, requirement.lte)
	if gte != nil && lte != nil && *gte > *lte {
		return false
	}
	// Both requirements have a complement
	if r.complement && requirement.complement {
		return true
	}
	// Only one requirement has a complement
	if r.complement && !requirement.complement {
		for v := range requirement.values {
			if !r.values.Has(v) && withinBounds(v, gte, lte) {
				return true
			}
		}
		return false
	}
	if !r.complement && requirement.complement {
		for v := range r.values {
			if !requirement.values.Has(v) && withinBounds(v, gte, lte) {
				return true
			}
		}
		return false
	}
	// Both requirements are non-complement requirements
	for v := range r.values {
		if requirement.values.Has(v) && withinBounds(v, gte, lte) {
			return true
		}
	}
	return false
}

func (r *Requirement) Any() string {
	switch r.Operator() {
	case corev1.NodeSelectorOpIn:
		return r.values.UnsortedList()[0]
	case corev1.NodeSelectorOpNotIn, corev1.NodeSelectorOpExists:
		min := 0
		max := math.MaxInt64
		if r.gte != nil {
			min = *r.gte
		}
		if r.lte != nil {
			max = *r.lte + 1
		}
		return fmt.Sprint(rand.Intn(max-min) + min) //nolint:gosec
	}
	return ""
}

// Has returns true if the requirement allows the value
func (r *Requirement) Has(value string) bool {
	if r.complement {
		return !r.values.Has(value) && withinBounds(value, r.gte, r.lte)
	}
	return r.values.Has(value) && withinBounds(value, r.gte, r.lte)
}

func (r *Requirement) Values() []string {
	return r.values.UnsortedList()
}

func (r *Requirement) Insert(items ...string) {
	r.values.Insert(items...)
}

func (r *Requirement) Operator() corev1.NodeSelectorOperator {
	if r.complement {
		if r.Len() < math.MaxInt64 {
			return corev1.NodeSelectorOpNotIn
		}
		return corev1.NodeSelectorOpExists // corev1.NodeSelectorOpGt and corev1.NodeSelectorOpLt are treated as "Exists" with bounds
	}
	if r.Len() > 0 {
		return corev1.NodeSelectorOpIn
	}
	return corev1.NodeSelectorOpDoesNotExist
}

func (r *Requirement) Len() int {
	if r.complement {
		return math.MaxInt64 - r.values.Len()
	}
	return r.values.Len()
}

func (r *Requirement) String() string {
	var s string
	switch r.Operator() {
	case corev1.NodeSelectorOpExists, corev1.NodeSelectorOpDoesNotExist:
		s = fmt.Sprintf("%s %s", r.Key, r.Operator())
	default:
		values := sets.List(r.values)
		if length := len(values); length > 5 {
			values = append(values[:5], fmt.Sprintf("and %d others", length-5))
		}
		s = fmt.Sprintf("%s %s %s", r.Key, r.Operator(), values)
	}
	if r.gte != nil {
		s += fmt.Sprintf(" >=%d", *r.gte)
	}
	if r.lte != nil {
		s += fmt.Sprintf(" <=%d", *r.lte)
	}
	if r.MinValues != nil {
		s += fmt.Sprintf(" minValues %d", *r.MinValues)
	}
	return s
}

func withinBounds(valueAsString string, gte, lte *int) bool {
	if gte == nil && lte == nil {
		return true
	}
	// If bounds are set, non integer values are invalid
	val, err := strconv.Atoi(valueAsString)
	if err != nil {
		return false
	}
	if gte != nil && val < *gte {
		return false
	}
	if lte != nil && val > *lte {
		return false
	}
	return true
}

func minIntPtr(a, b *int) *int {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if *a < *b {
		return a
	}
	return b
}

func maxIntPtr(a, b *int) *int {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if *a > *b {
		return a
	}
	return b
}
