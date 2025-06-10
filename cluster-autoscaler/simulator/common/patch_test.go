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
)

func TestPatchOperations(t *testing.T) {
	p := NewPatch[string, int]()

	// 1. Get and IsDeleted on non-existent key
	if _, found := p.Get("a"); found {
		t.Errorf("Get for 'a' should not find anything")
	}
	if p.IsDeleted("a") {
		t.Errorf("IsDeleted for 'a' should be false")
	}

	// 2. Set 'a' key
	p.Set("a", 1)
	if val, found := p.Get("a"); !found || val != 1 {
		t.Errorf("Get('a') = %v,%v, expected 1,true", val, found)
	}
	if p.IsDeleted("a") {
		t.Errorf("IsDeleted('a') = true, should be false")
	}

	// 3. Overwrite 'a' key
	p.Set("a", 2)
	if val, found := p.Get("a"); !found || val != 2 {
		t.Errorf("Get('a') = %v,%v, expected 2,true", val, found)
	}
	if p.IsDeleted("a") {
		t.Errorf("IsDeleted('a') = true, should be false")
	}

	// 4. Delete 'a' key
	p.Delete("a")
	if val, found := p.Get("a"); found {
		t.Errorf("Get('a') = %v,%v, should not find anything after delete", val, found)
	}
	if !p.IsDeleted("a") {
		t.Errorf("IsDeleted('a') = false, should be true")
	}

	// 5. Set 'a' key again after deletion
	p.Set("a", 3)
	if val, found := p.Get("a"); !found || val != 3 {
		t.Errorf("Get('a') = %v,%v, expected 3,true", val, found)
	}
	if p.IsDeleted("a") {
		t.Errorf("IsDeleted('a') = true, should be false")
	}

	// 6. Delete a non-existent key 'c'
	p.Delete("c")
	if val, found := p.Get("c"); found {
		t.Errorf("Get('c') = %v, %v, should not find anything", val, found)
	}
	if !p.IsDeleted("c") {
		t.Errorf("IsDeleted('c') = false, should be true")
	}
}

func TestCreatePatch(t *testing.T) {
	tests := map[string]struct {
		sourceMap    map[string]int
		addKeys      map[string]int
		deleteKeys   []string
		wantModified map[string]int
		wantDeleted  map[string]bool
	}{
		"SourceMapOnlyNoModifications": {
			sourceMap:    map[string]int{"k1": 1, "k2": 2},
			addKeys:      map[string]int{},
			deleteKeys:   []string{},
			wantModified: map[string]int{"k1": 1, "k2": 2},
			wantDeleted:  map[string]bool{},
		},
		"NilSourceMapAddAndDelete": {
			sourceMap:    nil,
			addKeys:      map[string]int{"a": 1, "b": 2},
			deleteKeys:   []string{"b"},
			wantModified: map[string]int{"a": 1},
			wantDeleted:  map[string]bool{"b": true},
		},
		"EmptySourceMapAddAndDelete": {
			sourceMap:    map[string]int{},
			addKeys:      map[string]int{"x": 10},
			deleteKeys:   []string{"y"},
			wantModified: map[string]int{"x": 10},
			wantDeleted:  map[string]bool{"y": true},
		},
		"NonEmptySourceMapAddOverwriteDelete": {
			sourceMap:    map[string]int{"orig1": 100, "orig2": 200},
			addKeys:      map[string]int{"new1": 300, "orig1": 101},
			deleteKeys:   []string{"orig2", "new1"},
			wantModified: map[string]int{"orig1": 101},
			wantDeleted:  map[string]bool{"orig2": true, "new1": true},
		},
		"DeleteKeyFromSourceMap": {
			sourceMap:    map[string]int{"key_to_delete": 70, "key_to_keep": 80},
			addKeys:      map[string]int{},
			deleteKeys:   []string{"key_to_delete"},
			wantModified: map[string]int{"key_to_keep": 80},
			wantDeleted:  map[string]bool{"key_to_delete": true},
		},
		"AddOnlyNoSourceMap": {
			sourceMap:    nil,
			addKeys:      map[string]int{"add1": 10, "add2": 20},
			deleteKeys:   []string{},
			wantModified: map[string]int{"add1": 10, "add2": 20},
			wantDeleted:  map[string]bool{},
		},
		"DeleteOnlyNoSourceMap": {
			sourceMap:    nil,
			addKeys:      map[string]int{},
			deleteKeys:   []string{"del1", "del2"},
			wantModified: map[string]int{},
			wantDeleted:  map[string]bool{"del1": true, "del2": true},
		},
		"DeleteKeyNotPresentInSourceOrAdded": {
			sourceMap:    map[string]int{"a": 1},
			addKeys:      map[string]int{"b": 2},
			deleteKeys:   []string{"c"},
			wantModified: map[string]int{"a": 1, "b": 2},
			wantDeleted:  map[string]bool{"c": true},
		},
		"AddKeyThenDeleteIt": {
			sourceMap:    map[string]int{"base": 100},
			addKeys:      map[string]int{"temp": 50},
			deleteKeys:   []string{"temp"},
			wantModified: map[string]int{"base": 100},
			wantDeleted:  map[string]bool{"temp": true},
		},
		"AllOperations": {
			sourceMap:    map[string]int{"s1": 1, "s2": 2, "s3": 3},
			addKeys:      map[string]int{"a1": 10, "s2": 22, "a2": 20},
			deleteKeys:   []string{"s1", "a1", "nonexistent"},
			wantModified: map[string]int{"s2": 22, "s3": 3, "a2": 20},
			wantDeleted:  map[string]bool{"s1": true, "a1": true, "nonexistent": true},
		},
		"NoOperationsEmptySource": {
			sourceMap:    map[string]int{},
			addKeys:      map[string]int{},
			deleteKeys:   []string{},
			wantModified: map[string]int{},
			wantDeleted:  map[string]bool{},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			p := NewPatchFromMap(test.sourceMap)

			for k, v := range test.addKeys {
				p.Set(k, v)
			}
			for _, k := range test.deleteKeys {
				p.Delete(k)
			}

			if !maps.Equal(p.modified, test.wantModified) {
				t.Errorf("Modified map mismatch: got %v, want %v", p.modified, test.wantModified)
			}

			if !maps.Equal(p.deleted, test.wantDeleted) {
				t.Errorf("Deleted map mismatch: got %v, want %v", p.deleted, test.wantDeleted)
			}
		})
	}
}

func TestMergePatchesInPlace(t *testing.T) {
	tests := map[string]struct {
		modifiedPatchA     map[string]int
		deletedPatchA      map[string]bool
		modifiedPatchB     map[string]int
		deletedPatchB      map[string]bool
		wantModifiedMerged map[string]int
		wantDeletedMerged  map[string]bool
	}{
		"PatchBOverwritesValueInPatchA": {
			modifiedPatchA:     map[string]int{"a": 1, "b": 2},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{"b": 22, "c": 3},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{"a": 1, "b": 22, "c": 3},
			wantDeletedMerged:  map[string]bool{},
		},
		"PatchBDeletesValuePresentInPatchA": {
			modifiedPatchA:     map[string]int{"a": 1, "b": 2},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{"a": true},
			wantModifiedMerged: map[string]int{"b": 2},
			wantDeletedMerged:  map[string]bool{"a": true},
		},
		"PatchBDeletesValueAlreadyDeletedInPatchA": {
			modifiedPatchA:     map[string]int{},
			deletedPatchA:      map[string]bool{"x": true},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{"x": true, "y": true},
			wantModifiedMerged: map[string]int{},
			wantDeletedMerged:  map[string]bool{"x": true, "y": true},
		},
		"PatchBAddsValuePreviouslyDeletedInPatchA": {
			modifiedPatchA:     map[string]int{},
			deletedPatchA:      map[string]bool{"x": true},
			modifiedPatchB:     map[string]int{"x": 10},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{"x": 10},
			wantDeletedMerged:  map[string]bool{},
		},
		"MergeEmptyPatchBIntoNonEmptyPatchA": {
			modifiedPatchA:     map[string]int{"a": 1},
			deletedPatchA:      map[string]bool{"b": true},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{"a": 1},
			wantDeletedMerged:  map[string]bool{"b": true},
		},
		"MergeNonEmptyPatchBIntoEmptyPatchA": {
			modifiedPatchA:     map[string]int{},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{"a": 1},
			deletedPatchB:      map[string]bool{"b": true},
			wantModifiedMerged: map[string]int{"a": 1},
			wantDeletedMerged:  map[string]bool{"b": true},
		},
		"MergeTwoEmptyPatches": {
			modifiedPatchA:     map[string]int{},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{},
			wantDeletedMerged:  map[string]bool{},
		},
		"NoOverlapBetweenPatchAAndPatchBModifications": {
			modifiedPatchA:     map[string]int{"a1": 1, "a2": 2},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{"b1": 10, "b2": 20},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{"a1": 1, "a2": 2, "b1": 10, "b2": 20},
			wantDeletedMerged:  map[string]bool{},
		},
		"NoOverlapBetweenPatchAAndPatchBDeletions": {
			modifiedPatchA:     map[string]int{},
			deletedPatchA:      map[string]bool{"a1": true, "a2": true},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{"b1": true, "b2": true},
			wantModifiedMerged: map[string]int{},
			wantDeletedMerged:  map[string]bool{"a1": true, "a2": true, "b1": true, "b2": true},
		},
		"PatchBOnlyAddsNewKeysPatchAUnchanged": {
			modifiedPatchA:     map[string]int{"orig": 5},
			modifiedPatchB:     map[string]int{"new1": 100, "new2": 200},
			deletedPatchA:      map[string]bool{},
			deletedPatchB:      map[string]bool{},
			wantModifiedMerged: map[string]int{"orig": 5, "new1": 100, "new2": 200},
			wantDeletedMerged:  map[string]bool{},
		},
		"PatchBOnlyDeletesNewKeysPatchAUnchanged": {
			modifiedPatchA:     map[string]int{"orig": 5},
			deletedPatchA:      map[string]bool{},
			modifiedPatchB:     map[string]int{},
			deletedPatchB:      map[string]bool{"del1": true, "del2": true},
			wantModifiedMerged: map[string]int{"orig": 5},
			wantDeletedMerged:  map[string]bool{"del1": true, "del2": true},
		},
		"AllInOne": {
			modifiedPatchA:     map[string]int{"k1": 1, "k2": 2, "k3": 3},
			deletedPatchA:      map[string]bool{"d1": true, "d2": true},
			modifiedPatchB:     map[string]int{"k2": 22, "k4": 4, "d1": 11},
			deletedPatchB:      map[string]bool{"k3": true, "d2": true, "d3": true},
			wantModifiedMerged: map[string]int{"k1": 1, "k2": 22, "k4": 4, "d1": 11},
			wantDeletedMerged:  map[string]bool{"k3": true, "d2": true, "d3": true},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			patchA := NewPatchFromMap(test.modifiedPatchA)
			for k := range test.deletedPatchA {
				patchA.Delete(k)
			}

			patchB := NewPatchFromMap(test.modifiedPatchB)
			for k := range test.deletedPatchB {
				patchB.Delete(k)
			}

			merged := mergePatchesInPlace(patchA, patchB)

			if merged != patchA {
				t.Errorf("mergePatchesInPlace did not modify patchA inplace, references are different")
			}

			if !maps.Equal(merged.modified, test.wantModifiedMerged) {
				t.Errorf("Modified map mismatch: got %v, want %v", merged.modified, test.wantModifiedMerged)
			}

			if !maps.Equal(merged.deleted, test.wantDeletedMerged) {
				t.Errorf("Deleted map mismatch: got %v, want %v", merged.deleted, test.wantDeletedMerged)
			}
		})
	}
}
