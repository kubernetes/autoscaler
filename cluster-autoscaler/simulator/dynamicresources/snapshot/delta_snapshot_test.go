package snapshot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"
)

func TestMergeDeltaMaps(t *testing.T) {
	testCases := map[string]struct {
		input []map[string]*int
		want  map[string]*int
	}{
		"EmptyInput": {
			input: []map[string]*int{},
			want:  map[string]*int{},
		},
		"SingleMap": {
			input: []map[string]*int{
				{
					"a": ptr.To(1),
					"b": ptr.To(2),
				},
			},
			want: map[string]*int{
				"a": ptr.To(1),
				"b": ptr.To(2),
			},
		},
		"MultipleMapsNoOverlap": {
			input: []map[string]*int{
				{"a": ptr.To(1)},
				{"b": ptr.To(2)},
			},
			want: map[string]*int{
				"a": ptr.To(1),
				"b": ptr.To(2),
			},
		},
		"MultipleMapsWithOverlap": {
			input: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2)},
				{"b": ptr.To(100), "c": ptr.To(3)}, // "b": 100 should be ignored
			},
			want: map[string]*int{
				"a": ptr.To(1),
				"b": ptr.To(2),
				"c": ptr.To(3),
			},
		},
		"MapWithNilValue": {
			input: []map[string]*int{
				{"a": ptr.To(1), "b": nil}, // "b": nil should be removed
				{"c": ptr.To(3)},
			},
			want: map[string]*int{
				"a": ptr.To(1),
				"c": ptr.To(3),
			},
		},
		"LaterMapHasNilValueForExistingKey": {
			input: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2)},
				{"b": nil, "c": ptr.To(3)}, // "b": nil should be ignored, original "b": 2 kept
			},
			want: map[string]*int{
				"a": ptr.To(1),
				"b": ptr.To(2),
				"c": ptr.To(3),
			},
		},
		"NilMapInInput": {
			input: []map[string]*int{
				{"a": ptr.To(1)},
				nil,
				{"c": ptr.To(3)},
			},
			want: map[string]*int{"a": ptr.To(1), "c": ptr.To(3)},
		},
		"AllNilMaps": {
			input: []map[string]*int{nil, nil},
			want:  map[string]*int{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := mergeDeltaMaps(tc.input)
			// Use EquateEmpty to treat nil map and empty map as equal for comparison
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("mergeMaps mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergeNestedDeltaMaps(t *testing.T) {
	testCases := map[string]struct {
		input []map[string]map[string]*int
		want  map[string]map[string]*int
	}{
		"EmptyInput": {
			input: []map[string]map[string]*int{},
			want:  map[string]map[string]*int{},
		},
		"SingleNestedMap": {
			input: []map[string]map[string]*int{
				{
					"outer1": {
						"innerA": ptr.To(1),
						"innerB": ptr.To(2),
					},
				},
			},
			want: map[string]map[string]*int{
				"outer1": {
					"innerA": ptr.To(1),
					"innerB": ptr.To(2),
				},
			},
		},
		"MultipleMapsNoOuterOverlap": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1)}},
				{"outer2": {"innerC": ptr.To(3)}},
			},
			want: map[string]map[string]*int{
				"outer1": {"innerA": ptr.To(1)},
				"outer2": {"innerC": ptr.To(3)},
			},
		},
		"MultipleMapsOuterOverlapNoInnerOverlap": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1)}},
				{"outer1": {"innerB": ptr.To(2)}},
			},
			want: map[string]map[string]*int{
				"outer1": {
					"innerA": ptr.To(1),
					"innerB": ptr.To(2),
				},
			},
		},
		"MultipleMapsOuterAndInnerOverlap": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1), "innerB": ptr.To(2)}},
				{"outer1": {"innerB": ptr.To(100), "innerC": ptr.To(3)}}, // "innerB": 100 ignored
				{"outer2": {"innerD": ptr.To(4)}},
			},
			want: map[string]map[string]*int{
				"outer1": {
					"innerA": ptr.To(1),
					"innerB": ptr.To(2),
					"innerC": ptr.To(3),
				},
				"outer2": {"innerD": ptr.To(4)},
			},
		},
		"NilInnerMapMarksOuterKeyAsIgnored": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1)}},
				{"outer1": nil},                   // This makes outer1 ignored from now on
				{"outer1": {"innerB": ptr.To(2)}}, // This whole map for outer1 is ignored
				{"outer2": {"innerC": ptr.To(3)}},
			},
			want: map[string]map[string]*int{
				"outer1": {"innerA": ptr.To(1)},
				"outer2": {"innerC": ptr.To(3)}, // outer1 is gone
			},
		},
		"NilInnerMapInFirstMap": {
			input: []map[string]map[string]*int{
				{"outer1": nil},
				{"outer1": {"innerA": ptr.To(1)}}, // Ignored
				{"outer2": {"innerC": ptr.To(3)}},
			},
			want: map[string]map[string]*int{
				"outer2": {
					"innerC": ptr.To(3), // outer1 is gone
				},
			},
		},
		"NilValueInInnerMap": {
			// Note: The inner map {"innerA": ptr.To(1), "innerB": nil} becomes {"innerA": ptr.To(1)}
			// after the inner mergeDeltaMaps call within mergeNestedDeltaMaps.
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1), "innerB": nil}}, // innerB: nil removed by inner merge
				{"outer2": {"innerC": ptr.To(3)}},
			},
			want: map[string]map[string]*int{
				"outer1": {"innerA": ptr.To(1)},
				"outer2": {"innerC": ptr.To(3)},
			},
		},
		"NilValueInLaterInnerMapForExistingKey": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1), "innerB": ptr.To(2)}},
				{"outer1": {"innerB": nil, "innerC": ptr.To(3)}}, // innerB: nil ignored, innerB: 2 kept
			},
			want: map[string]map[string]*int{
				"outer1": {
					"innerA": ptr.To(1),
					"innerB": ptr.To(2),
					"innerC": ptr.To(3),
				},
			},
		},
		"NilMapInInputSlice": {
			input: []map[string]map[string]*int{
				{"outer1": {"innerA": ptr.To(1)}},
				nil,
				{"outer2": {"innerC": ptr.To(3)}},
			},
			want: map[string]map[string]*int{
				"outer1": {"innerA": ptr.To(1)},
				"outer2": {"innerC": ptr.To(3)},
			},
		},
		"AllNilMapsInInputSlice": {
			input: []map[string]map[string]*int{
				nil,
				nil,
			},
			want: map[string]map[string]*int{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := mergeNestedDeltaMaps(tc.input)
			// Use EquateEmpty to treat nil map and empty map as equal for comparison
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("mergeNestedMaps mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
