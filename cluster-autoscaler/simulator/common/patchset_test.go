/*
Copyright 2025 The Kubernetes Authors.

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

package common

import (
	"maps"
	"testing"

	"k8s.io/utils/ptr"
)

func TestPatchSetAsMap(t *testing.T) {
	tests := map[string]struct {
		patchLayers []map[string]*int
		wantMap     map[string]int
	}{
		"EmptyPatchSet": {
			patchLayers: []map[string]*int{},
			wantMap:     map[string]int{},
		},
		"SingleLayerNoDeletions": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": ptr.To(2)}},
			wantMap:     map[string]int{"a": 1, "b": 2},
		},
		"SingleLayerOnlyDeletions": {
			patchLayers: []map[string]*int{{"a": nil, "b": nil}},
			wantMap:     map[string]int{},
		},
		"SingleLayerModificationsAndDeletions": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": nil, "c": ptr.To(3)}},
			wantMap: map[string]int{"a": 1, "c": 3},
		},
		"MultipleLayersNoDeletionsOverwrite": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2)},
				{"b": ptr.To(22), "c": ptr.To(3)}},
			wantMap: map[string]int{"a": 1, "b": 22, "c": 3},
		},
		"MultipleLayersWithDeletions": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": nil, "x": ptr.To(100)},
				{"c": ptr.To(3), "x": ptr.To(101), "a": nil}},
			wantMap: map[string]int{"c": 3, "x": 101},
		},
		"MultipleLayersDeletionInLowerLayer": {
			patchLayers: []map[string]*int{
				{"a": nil, "b": ptr.To(2)},
				{"c": ptr.To(3)}},
			wantMap: map[string]int{"b": 2, "c": 3},
		},
		"MultipleLayersAddAfterDelete": {
			patchLayers: []map[string]*int{
				{"a": nil},
				{"a": ptr.To(11), "b": ptr.To(2)}},
			wantMap: map[string]int{"a": 11, "b": 2},
		},
		"MultipleLayersDeletePreviouslyAdded": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2)},
				{"c": ptr.To(3), "b": nil}},
			wantMap: map[string]int{"a": 1, "c": 3},
		},
		"AllInOne": {
			patchLayers: []map[string]*int{
				{"k1": ptr.To(1), "k2": ptr.To(2)},
				{"k2": ptr.To(22), "k3": ptr.To(3), "k1": nil},
				{"k1": ptr.To(111), "k3": ptr.To(33), "k4": ptr.To(4), "deleted": nil},
			},
			wantMap: map[string]int{"k2": 22, "k3": 33, "k4": 4, "k1": 111},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			patchset := buildTestPatchSet(t, test.patchLayers)
			mergedMap := patchset.AsMap()
			if !maps.Equal(mergedMap, test.wantMap) {
				t.Errorf("AsMap() result mismatch: got %v, want %v", mergedMap, test.wantMap)
			}
		})
	}
}

func TestPatchSetCommit(t *testing.T) {
	tests := map[string]struct {
		patchLayers []map[string]*int
		wantMap     map[string]int // Expected map after each commit
	}{
		"CommitEmptyPatchSet": {
			patchLayers: []map[string]*int{},
			wantMap:     map[string]int{},
		},
		"CommitSingleLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			wantMap:     map[string]int{"a": 1},
		},
		"CommitTwoLayersNoOverlap": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			wantMap:     map[string]int{"a": 1, "b": 2},
		},
		"CommitTwoLayersWithOverwrite": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"a": ptr.To(2)}},
			wantMap:     map[string]int{"a": 2},
		},
		"CommitTwoLayersWithDeleteInTopLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": ptr.To(5)}, {"c": ptr.To(3), "a": nil}},
			wantMap:     map[string]int{"b": 5, "c": 3},
		},
		"CommitTwoLayersWithDeleteInBottomLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": nil}, {"a": ptr.To(11), "c": ptr.To(3)}},
			wantMap:     map[string]int{"a": 11, "c": 3},
		},
		"CommitThreeLayers": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2), "z": ptr.To(100)},
				{"b": ptr.To(22), "c": ptr.To(3), "a": nil},
				{"c": ptr.To(33), "d": ptr.To(4), "a": ptr.To(111), "b": nil, "deleted": nil},
			},
			wantMap: map[string]int{"z": 100, "c": 33, "d": 4, "a": 111},
		},
		"CommitMultipleLayersDeleteAndAddBack": {
			patchLayers: []map[string]*int{
				{"x": ptr.To(10)},
				{"x": nil},
				{"x": ptr.To(30)},
				{"y": ptr.To(40), "x": nil}},
			wantMap: map[string]int{"y": 40},
		},
		"CommitEmptyLayersBetweenDataLayers": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {}, {"b": ptr.To(2)}, {}, {"c": ptr.To(3)}},
			wantMap:     map[string]int{"a": 1, "b": 2, "c": 3},
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			ps := buildTestPatchSet(t, tc.patchLayers)
			initialNumPatches := len(ps.patches)

			if currentMap := ps.AsMap(); !maps.Equal(currentMap, tc.wantMap) {
				t.Errorf("AsMap() before any commits mismatch: got %v, want %v", currentMap, tc.wantMap)
			}

			for i := 0; i < initialNumPatches-1; i++ {
				expectedPatchesAfterCommit := len(ps.patches) - 1
				ps.Commit()
				if len(ps.patches) != expectedPatchesAfterCommit {
					t.Errorf("After commit #%d, expected %d patches, got %d", i+1, expectedPatchesAfterCommit, len(ps.patches))
				}
				if currentMap := ps.AsMap(); !maps.Equal(currentMap, tc.wantMap) {
					t.Errorf("AsMap() after commit #%d mismatch: got %v, want %v", i+1, currentMap, tc.wantMap)
				}
			}

			if initialNumPatches > 0 && len(ps.patches) != 1 {
				t.Errorf("Expected 1 patch after all commits, got %d", len(ps.patches))
			} else if initialNumPatches == 0 && len(ps.patches) != 0 {
				t.Errorf("Expected 0 patches after all commits, got %d", len(ps.patches))
			}
		})
	}
}

func TestPatchSetRevert(t *testing.T) {
	tests := map[string]struct {
		patchLayers         []map[string]*int
		wantInitialMap      map[string]int
		wantMapsAfterRevert []map[string]int
	}{
		"RevertEmptyPatchSet": {
			patchLayers:         []map[string]*int{},
			wantInitialMap:      map[string]int{},
			wantMapsAfterRevert: []map[string]int{{}},
		},
		"RevertSingleLayer": {
			patchLayers:         []map[string]*int{{"a": ptr.To(1)}},
			wantInitialMap:      map[string]int{"a": 1},
			wantMapsAfterRevert: []map[string]int{{"a": 1}},
		},
		"RevertTwoLayersNoOverlap": {
			patchLayers:         []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			wantInitialMap:      map[string]int{"a": 1, "b": 2},
			wantMapsAfterRevert: []map[string]int{{"a": 1}, {"a": 1}},
		},
		"RevertTwoLayersWithOverwrite": {
			patchLayers:         []map[string]*int{{"a": ptr.To(1)}, {"a": ptr.To(2)}},
			wantInitialMap:      map[string]int{"a": 2},
			wantMapsAfterRevert: []map[string]int{{"a": 1}, {"a": 1}},
		},
		"RevertTwoLayersWithDeleteInTopLayer": {
			patchLayers:         []map[string]*int{{"a": ptr.To(1), "b": ptr.To(5)}, {"c": ptr.To(3), "a": nil}},
			wantInitialMap:      map[string]int{"b": 5, "c": 3},
			wantMapsAfterRevert: []map[string]int{{"a": 1, "b": 5}, {"a": 1, "b": 5}},
		},
		"RevertThreeLayers": {
			patchLayers: []map[string]*int{
				{"a": ptr.To(1), "b": ptr.To(2), "z": ptr.To(100)},
				{"b": ptr.To(22), "c": ptr.To(3), "a": nil},
				{"c": ptr.To(33), "d": ptr.To(4), "a": ptr.To(111), "b": nil, "deleted": nil},
			},
			wantInitialMap: map[string]int{"z": 100, "c": 33, "d": 4, "a": 111},
			wantMapsAfterRevert: []map[string]int{
				{"a": 1, "b": 2, "z": 100},
				{"a": 1, "b": 2, "z": 100},
				{"z": 100, "b": 22, "c": 3},
			},
		},
		"RevertMultipleLayersDeleteAndAddBack": {
			patchLayers:    []map[string]*int{{"x": ptr.To(10)}, {"x": nil}, {"x": ptr.To(30)}, {"y": ptr.To(40), "x": nil}},
			wantInitialMap: map[string]int{"y": 40},
			wantMapsAfterRevert: []map[string]int{
				{"x": 10},
				{"x": 10},
				{},
				{"x": 30},
			},
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			ps := buildTestPatchSet(t, tc.patchLayers)
			patchesNumber := len(ps.patches)

			if currentMap := ps.AsMap(); !maps.Equal(currentMap, tc.wantInitialMap) {
				t.Errorf("AsMap() before any reverts mismatch: got %v, want %v", currentMap, tc.wantInitialMap)
			}

			for i := 0; i <= patchesNumber; i++ {
				wantPatchesAfterRevert := len(ps.patches) - 1
				ps.Revert()
				if len(ps.patches) != wantPatchesAfterRevert && len(ps.patches) > 1 {
					t.Errorf("After revert #%d, expected %d patches, got %d", i+1, wantPatchesAfterRevert, len(ps.patches))
				}

				currentMap := ps.AsMap()
				wantMapIndex := max(patchesNumber-i-1, 0)
				wantMapAfterRevert := tc.wantMapsAfterRevert[wantMapIndex]
				if !maps.Equal(currentMap, wantMapAfterRevert) {
					t.Errorf("AsMap() after revert #%d mismatch: got %v, want %v", i+1, currentMap, wantMapAfterRevert)
				}
			}

			if patchesNumber >= 1 && len(ps.patches) != 1 {
				t.Errorf("Expected 1 patch after all reverts, got %d", len(ps.patches))
			} else if patchesNumber == 0 && len(ps.patches) != 0 {
				t.Errorf("Expected 0 patches after all reverts, got %d", len(ps.patches))
			}
		})
	}
}

func TestPatchSetForkRevert(t *testing.T) {
	// 1. Initialize empty patchset
	ps := NewPatchSet(NewPatch[string, int]())
	if initialMap := ps.AsMap(); len(initialMap) != 0 {
		t.Fatalf("Initial AsMap() got %v, want empty", initialMap)
	}

	// 2. Call Fork
	ps.Fork()
	if len(ps.patches) != 2 {
		t.Fatalf("Expected 2 patches after Fork(), got %d", len(ps.patches))
	}

	// 3. Perform some mutation operations on the new layer
	ps.SetCurrent("a", 1)
	ps.SetCurrent("b", 2)
	ps.DeleteCurrent("a")

	// 4. Call Revert
	ps.Revert()
	if len(ps.patches) != 1 {
		t.Fatalf("Expected 1 patch after Revert(), got %d", len(ps.patches))
	}

	// 5. Compare state to the empty map
	if finalMap := ps.AsMap(); len(finalMap) != 0 {
		t.Errorf("AsMap() got %v, want empty", finalMap)
	}
}

func TestPatchSetForkCommit(t *testing.T) {
	// 1. Initialize empty patchset
	ps := NewPatchSet(NewPatch[string, int]())
	if initialMap := ps.AsMap(); len(initialMap) != 0 {
		t.Fatalf("Initial AsMap() got %v, want empty", initialMap)
	}

	// 2. Call Fork two times
	ps.Fork()
	ps.Fork()
	if len(ps.patches) != 3 {
		t.Fatalf("Expected 3 patches after 2xFork(), got %d", len(ps.patches))
	}

	// 3. Perform some mutation operations on the current layer
	ps.SetCurrent("a", 1)
	ps.SetCurrent("b", 2)
	ps.DeleteCurrent("a")

	// 4. Call Commit to persist changes
	ps.Commit()
	if len(ps.patches) != 2 {
		t.Fatalf("Expected 1 patch after Commit(), got %d", len(ps.patches))
	}

	// 5. Call Revert on the empty layer
	ps.Revert()
	if len(ps.patches) != 1 {
		t.Fatalf("Expected 1 patch after Revert(), got %d", len(ps.patches))
	}

	// 6. Compare state to the empty map
	wantMap := map[string]int{"b": 2}
	if finalMap := ps.AsMap(); maps.Equal(wantMap, finalMap) {
		t.Errorf("AsMap() got %v, want %v", finalMap, wantMap)
	}
}

func TestPatchSetFindValue(t *testing.T) {
	tests := map[string]struct {
		patchLayers []map[string]*int
		searchKey   string
		wantValue   int
		wantFound   bool
	}{
		"EmptyPatchSet_KeyNotFound": {
			patchLayers: []map[string]*int{},
			searchKey:   "a",
			wantFound:   false,
		},
		"SingleLayer_KeyFound": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": ptr.To(2)}},
			searchKey:   "a",
			wantValue:   1,
			wantFound:   true,
		},
		"SingleLayer_KeyNotFound": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			searchKey:   "x",
			wantFound:   false,
		},
		"SingleLayer_KeyDeleted": {
			patchLayers: []map[string]*int{{"a": nil, "b": ptr.To(2)}},
			searchKey:   "a",
			wantFound:   false,
		},
		"MultiLayer_KeyInTopLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"a": ptr.To(10), "b": ptr.To(2)}},
			searchKey:   "a",
			wantValue:   10,
			wantFound:   true,
		},
		"MultiLayer_KeyInBottomLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": ptr.To(2)}, {"b": ptr.To(22)}},
			searchKey:   "a",
			wantValue:   1,
			wantFound:   true,
		},
		"MultiLayer_KeyOverwrittenThenDeleted": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"a": ptr.To(10)}, {"a": nil, "b": ptr.To(2)}},
			searchKey:   "a",
			wantFound:   false,
		},
		"MultiLayer_KeyDeletedThenAddedBack": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"a": nil}, {"a": ptr.To(100), "b": ptr.To(2)}},
			searchKey:   "a",
			wantValue:   100,
			wantFound:   true,
		},
		"MultiLayer_KeyNotFoundInAnyLayer": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			searchKey:   "x",
			wantFound:   false,
		},
		"MultiLayer_KeyDeletedInMiddleLayer_NotFound": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"a": nil, "b": ptr.To(2)}, {"c": ptr.To(3)}},
			searchKey:   "a",
			wantFound:   false,
		},
		"MultiLayer_KeyPresentInBase_DeletedInOverlay_ReAddedInTop": {
			patchLayers: []map[string]*int{
				{"k1": ptr.To(1)},
				{"k1": nil},
				{"k1": ptr.To(111)},
			},
			searchKey: "k1",
			wantValue: 111,
			wantFound: true,
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			ps := buildTestPatchSet(t, tc.patchLayers)
			val, found := ps.FindValue(tc.searchKey)

			if found != tc.wantFound || (found && val != tc.wantValue) {
				t.Errorf("FindValue(%q) got val=%v, found=%v; want val=%v, found=%v", tc.searchKey, val, found, tc.wantValue, tc.wantFound)
			}
		})
	}
}

func TestPatchSetOperations(t *testing.T) {
	tests := map[string]struct {
		patchLayers    []map[string]*int
		mutatePatchSet func(ps *PatchSet[string, int])
		searchKey      string
		wantInCurrent  bool
	}{
		"SetCurrent_NewKey_InitiallyEmptyPatchSet": {
			patchLayers: []map[string]*int{},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.SetCurrent("a", 1)
			},
			searchKey:     "a",
			wantInCurrent: true,
		},
		"SetCurrent_NewKey_OnExistingEmptyLayer": {
			patchLayers: []map[string]*int{{}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.SetCurrent("a", 1)
			},
			searchKey:     "a",
			wantInCurrent: true,
		},
		"SetCurrent_OverwriteKey_InCurrentPatch": {
			patchLayers: []map[string]*int{{"a": ptr.To(10)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.SetCurrent("a", 1)
			},
			searchKey:     "a",
			wantInCurrent: true,
		},
		"SetCurrent_OverwriteKey_InLowerPatch_AfterFork": {
			patchLayers: []map[string]*int{{"a": ptr.To(10)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.Fork()
				ps.SetCurrent("a", 1)
			},
			searchKey:     "a",
			wantInCurrent: true,
		},
		"DeleteCurrent_ExistingKey_InCurrentPatch": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.DeleteCurrent("a")
			},
			searchKey:     "a",
			wantInCurrent: false,
		},
		"DeleteCurrent_ExistingKey_InLowerPatch_AfterFork": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.Fork()
				ps.DeleteCurrent("a")
			},
			searchKey:     "a",
			wantInCurrent: false,
		},
		"DeleteCurrent_NonExistentKey_OnExistingEmptyLayer": {
			patchLayers: []map[string]*int{{}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.DeleteCurrent("x")
			},
			searchKey:     "x",
			wantInCurrent: false,
		},
		"DeleteCurrent_NonExistentKey_InitiallyEmptyPatchSet": {
			patchLayers: []map[string]*int{}, // Starts with no layers
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.DeleteCurrent("x")
			},
			searchKey:     "x",
			wantInCurrent: false,
		},
		"InCurrentPatch_KeySetInCurrent": {
			patchLayers:    []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {},
			searchKey:      "a",
			wantInCurrent:  true,
		},
		"InCurrentPatch_KeySetInLower_NotInCurrent_AfterFork": {
			patchLayers:    []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) { ps.Fork() },
			searchKey:      "a",
			wantInCurrent:  false,
		},
		"InCurrentPatch_KeyDeletedInCurrent": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.DeleteCurrent("a")
			},
			searchKey:     "a",
			wantInCurrent: false, // Get in current patch won't find it.
		},
		"InCurrentPatch_NonExistentKey_InitiallyEmptyPatchSet": {
			patchLayers:    []map[string]*int{},
			mutatePatchSet: func(ps *PatchSet[string, int]) {},
			searchKey:      "a",
			wantInCurrent:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ps := buildTestPatchSet(t, tc.patchLayers)
			tc.mutatePatchSet(ps)

			inCurrent := ps.InCurrentPatch(tc.searchKey)
			if inCurrent != tc.wantInCurrent {
				t.Errorf("InCurrentPatch(%q): got %v, want %v", tc.searchKey, inCurrent, tc.wantInCurrent)
			}
		})
	}
}

func TestPatchSetCache(t *testing.T) {
	tests := map[string]struct {
		patchLayers     []map[string]*int
		mutatePatchSet  func(ps *PatchSet[string, int])
		wantCache       map[string]*int
		wantCacheInSync bool
	}{
		"Initial_EmptyPatchSet": {
			patchLayers:     []map[string]*int{},
			mutatePatchSet:  func(ps *PatchSet[string, int]) {},
			wantCache:       map[string]*int{},
			wantCacheInSync: false,
		},
		"Initial_WithData_NoCacheAccess": {
			patchLayers:     []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet:  func(ps *PatchSet[string, int]) {},
			wantCache:       map[string]*int{},
			wantCacheInSync: false,
		},
		"FindValue_PopulatesCacheForKey": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": ptr.To(2)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("a")
			},
			wantCache:       map[string]*int{"a": ptr.To(1)},
			wantCacheInSync: false,
		},
		"FindValue_DeletedKey_PopulatesCacheWithNil": {
			patchLayers: []map[string]*int{{"a": nil, "b": ptr.To(2)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("a")
			},
			wantCache:       map[string]*int{"a": nil},
			wantCacheInSync: false,
		},
		"AsMap_PopulatesAndSyncsCache": {
			patchLayers: []map[string]*int{{"a": ptr.To(1), "b": nil, "c": ptr.To(3)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.AsMap()
			},
			wantCache:       map[string]*int{"a": ptr.To(1), "c": ptr.To(3)}, // Cache does not necessarily track deletions like 'b' key
			wantCacheInSync: true,
		},
		"SetCurrent_UpdatesCache_NewKey": {
			patchLayers: []map[string]*int{{}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.SetCurrent("x", 10)
			},
			wantCache:       map[string]*int{"x": ptr.To(10)},
			wantCacheInSync: false,
		},
		"SetCurrent_UpdatesCache_OverwriteKey": {
			patchLayers: []map[string]*int{{"x": ptr.To(5)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("x")
				ps.SetCurrent("x", 10)
			},
			wantCache:       map[string]*int{"x": ptr.To(10)},
			wantCacheInSync: false,
		},
		"DeleteCurrent_UpdatesCache": {
			patchLayers: []map[string]*int{{"x": ptr.To(5)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("x")
				ps.DeleteCurrent("x")
			},
			wantCache:       map[string]*int{"x": nil},
			wantCacheInSync: false,
		},
		"Revert_ClearsAffectedCacheEntries_And_SetsCacheNotInSync": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2), "a": ptr.To(11)}}, // Layer 0: a=1; Layer 1: b=2, a=11
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("a")
				ps.FindValue("b")
				ps.Revert()
			},
			wantCache:       map[string]*int{},
			wantCacheInSync: false,
		},
		"Revert_OnSyncedCache_SetsCacheNotInSync": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.AsMap()
				ps.Revert()
			},
			wantCache:       map[string]*int{"a": ptr.To(1)},
			wantCacheInSync: false,
		},
		"Commit_DoesNotInvalidateCache_IfValuesConsistent": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("a")
				ps.FindValue("b")
				ps.Commit()
			},
			wantCache:       map[string]*int{"a": ptr.To(1), "b": ptr.To(2)},
			wantCacheInSync: false,
		},
		"Commit_OnSyncedCache_KeepsCacheInSync": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}, {"b": ptr.To(2)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.AsMap()
				ps.Commit()
			},
			wantCache:       map[string]*int{"a": ptr.To(1), "b": ptr.To(2)},
			wantCacheInSync: true,
		},
		"Fork_DoesNotInvalidateCache": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.FindValue("a")
				ps.Fork()
			},
			wantCache:       map[string]*int{"a": ptr.To(1)},
			wantCacheInSync: false,
		},
		"Fork_OnSyncedCache_KeepsCacheInSync": {
			patchLayers: []map[string]*int{{"a": ptr.To(1)}},
			mutatePatchSet: func(ps *PatchSet[string, int]) {
				ps.AsMap()
				ps.Fork()
			},
			wantCache:       map[string]*int{"a": ptr.To(1)},
			wantCacheInSync: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ps := buildTestPatchSet(t, tc.patchLayers)
			tc.mutatePatchSet(ps)

			if !maps.EqualFunc(ps.cache, tc.wantCache, func(a, b *int) bool {
				if a == nil && b == nil {
					return true
				}
				if a == nil || b == nil {
					return false
				}
				return *a == *b
			}) {
				t.Errorf("Cache content mismatch: got %v, want %v", ps.cache, tc.wantCache)
			}

			if ps.cacheInSync != tc.wantCacheInSync {
				t.Errorf("cacheInSync status mismatch: got %v, want %v", ps.cacheInSync, tc.wantCacheInSync)
			}
		})
	}
}

func buildTestPatchSet[K comparable, V any](t *testing.T, patchLayers []map[K]*V) *PatchSet[K, V] {
	t.Helper()

	patchesNumber := len(patchLayers)
	patches := make([]*Patch[K, V], patchesNumber)
	for i := range patchesNumber {
		layerMap := patchLayers[i]
		currentPatch := NewPatch[K, V]()
		for k, vPtr := range layerMap {
			if vPtr != nil {
				currentPatch.Set(k, *vPtr)
			} else {
				currentPatch.Delete(k)
			}
		}
		patches[i] = currentPatch
	}

	return NewPatchSet(patches...)
}
