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

package logic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func countStatusPatches(t *testing.T, vpas []conflictVpa) int {
	t.Helper()
	client, _ := runReconcile(vpas, false)
	n := 0
	for _, a := range client.Actions() {
		if a.GetVerb() == "patch" && a.GetSubresource() == "status" {
			n++
		}
	}
	return n
}

func TestReconcileTargetConflictsPatches(t *testing.T) {
	// The message updateTargetConflictCondition builds for vpa-2 when vpa-1 controls the target.
	const upToDate = `Conflict: multiple active VPAs target the same resource; VPA "vpa-1" controls it and this VPA is not applied`

	tests := []struct {
		name        string
		vpas        []conflictVpa
		wantPatches int
	}{
		{
			name: "unchanged conflict sends no patch",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour},
				{name: "vpa-2", age: time.Hour, priorMessage: upToDate},
			},
			wantPatches: 0,
		},
		{
			name: "changed message still patches",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour},
				{name: "vpa-2", age: time.Hour, priorMessage: `Conflict: VPA "vpa-9" controls it`},
			},
			wantPatches: 1,
		},
		{
			name: "new conflict patches once",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour},
				{name: "vpa-2", age: time.Hour},
			},
			wantPatches: 1,
		},
		{
			name:        "resolved conflict patches once",
			vpas:        []conflictVpa{{name: "vpa-1", priorMessage: "old"}},
			wantPatches: 1,
		},
		{
			name:        "no conflict and no prior condition sends no patch",
			vpas:        []conflictVpa{{name: "vpa-1"}},
			wantPatches: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantPatches, countStatusPatches(t, tc.vpas))
		})
	}
}
