/*
Copyright 2020 The Kubernetes Authors.

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

package gce

import (
	"testing"
)

func TestMachineCache(t *testing.T) {
	type CacheQuery struct {
		zone    string
		machine MachineType
	}
	testCases := []struct {
		desc        string
		machines    []CacheQuery
		wantCPU     map[MachineTypeKey]int64
		wantMissing []MachineTypeKey
	}{
		{
			desc: "replacement",
			machines: []CacheQuery{
				{
					zone: "myzone",
					machine: MachineType{
						Name: "e2-standard-2",
						CPU:  1,
					},
				},
				{
					zone: "myzone",
					machine: MachineType{
						Name: "e2-standard-2",
						CPU:  2,
					},
				},
			},
			wantCPU: map[MachineTypeKey]int64{
				{
					MachineTypeName: "e2-standard-2",
					Zone:            "myzone",
				}: 2,
			},
			wantMissing: []MachineTypeKey{
				{
					Zone:            "myzone2",
					MachineTypeName: "e2-standard-4",
				},
			},
		},
	}
	c := NewGceCache()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			for _, m := range tc.machines {
				c.AddMachine(m.machine, m.zone)
			}
			for mt, wantCPU := range tc.wantCPU {
				m, found := c.GetMachine(mt.MachineTypeName, mt.Zone)
				if !found {
					t.Errorf("Expected to find machine in cache type = %q, zone = %q", mt.MachineTypeName, mt.Zone)
				}
				if m.CPU != wantCPU {
					t.Errorf("Wanted CPU %d but got CPU %d for machine type = %q, zone = %q", wantCPU, m.CPU, mt.MachineTypeName, mt.Zone)
				}
			}
			for _, mt := range tc.wantMissing {
				_, found := c.GetMachine(mt.MachineTypeName, mt.Zone)
				if found {
					t.Errorf("Didn't expect to find in cache machine type = %q, zone = %q", mt.MachineTypeName, mt.Zone)
				}
			}
		})
	}
}

func TestListManagedInstancesResultsCache(t *testing.T) {
	checkInCache := func(c *GceCache, migRef GceRef, expectedResults string) {
		result, found := c.GetListManagedInstancesResults(migRef)
		if !found {
			t.Errorf("Results not found for MIG ref: %s", migRef.String())
		}
		if result != expectedResults {
			t.Errorf("Expected results %s for MIG ref: %s, but got: %s", expectedResults, migRef.String(), result)
		}
	}
	migRef := GceRef{
		Project: "project",
		Zone:    "us-test1",
		Name:    "mig",
	}
	c := NewGceCache()
	c.SetListManagedInstancesResults(migRef, "PAGINATED")
	checkInCache(c, migRef, "PAGINATED")
	c.SetListManagedInstancesResults(migRef, "PAGELESS")
	checkInCache(c, migRef, "PAGELESS")
	c.InvalidateAllListManagedInstancesResults()
	if cacheSize := len(c.listManagedInstancesResultsCache); cacheSize > 0 {
		t.Errorf("Expected listManagedInstancesResultsCache to be empty, but it still contains %d entries", cacheSize)
	}
}
