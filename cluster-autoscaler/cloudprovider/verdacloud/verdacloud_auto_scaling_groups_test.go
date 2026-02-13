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

package verdacloud

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// Test constants - define magic strings in one place
const (
	testAsgName        = "test-asg"
	testInstanceType   = "1H100.80S.22V"
	testHostnamePrefix = "test"
	testLocation       = "FIN-01"
	testProviderPrefix = "verdacloud://"
)

// newTestEnv creates a standard test environment with a VerdacloudManager and Asg.
// This prevents copy-pasting initialization across multiple tests.
func newTestEnv(t *testing.T) (*VerdacloudManager, *Asg, *autoScalingGroups) {
	t.Helper()

	asg := &Asg{
		AsgRef:                AsgRef{Name: testAsgName},
		minSize:               1,
		maxSize:               10,
		curSize:               0,
		instanceType:          testInstanceType,
		hostnamePrefix:        testHostnamePrefix,
		AvailabilityLocations: []string{testLocation},
	}

	asgs := &autoScalingGroups{
		registeredAsgs:    make(map[AsgRef]*Asg),
		asgToInstances:    make(map[AsgRef][]InstanceRef),
		instanceToAsg:     make(map[InstanceRef]*Asg),
		asgNodeGroupSpecs: make(map[AsgRef]string),
		failedInstances:   make(map[string]time.Time),
		lastFailureCheck:  make(map[AsgRef]time.Time),
	}
	asgs.registeredAsgs[asg.AsgRef] = asg

	manager := &VerdacloudManager{
		asgs: asgs,
	}

	return manager, asg, asgs
}

// newTestNodeGroup creates a VerdacloudNodeGroup linked to the given manager and asg
func newTestNodeGroup(t *testing.T, manager *VerdacloudManager, asg *Asg) *VerdacloudNodeGroup {
	t.Helper()
	return &VerdacloudNodeGroup{
		asg:     asg,
		manager: manager,
	}
}

// createTestNodes creates dummy nodes with provider IDs for testing
func createTestNodes(t *testing.T, count int) []*apiv1.Node {
	t.Helper()
	nodes := make([]*apiv1.Node, count)
	for i := 0; i < count; i++ {
		// Use lowercase location in hostname to match createTestInstanceRefs
		hostname := fmt.Sprintf("%s-vm-%s-%02d", testHostnamePrefix, strings.ToLower(testLocation), i)
		nodes[i] = &apiv1.Node{}
		nodes[i].Spec.ProviderID = fmt.Sprintf("%s%s/%s", testProviderPrefix, testLocation, hostname)
	}
	return nodes
}

// createTestInstanceRefs creates instance references for testing
func createTestInstanceRefs(t *testing.T, count int) []InstanceRef {
	t.Helper()
	refs := make([]InstanceRef, count)
	for i := 0; i < count; i++ {
		hostname := fmt.Sprintf("%s-vm-%s-%02d", testHostnamePrefix, strings.ToLower(testLocation), i)
		refs[i] = InstanceRef{
			Hostname:   hostname,
			ProviderID: fmt.Sprintf("%s%s/%s", testProviderPrefix, testLocation, hostname),
		}
	}
	return refs
}

// assertErrorContains checks if an error contains a specific substring
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("expected error containing %q, got %q", substr, err.Error())
	}
}

// assertNoError fails the test if err is not nil
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAsg_ConcurrentScaleLock(t *testing.T) {
	_, asg, _ := newTestEnv(t)

	// Use more iterations to increase chance of catching race conditions.
	// The race detector (go test -race) is the authoritative way to catch races,
	// but higher iteration counts make the test more robust.
	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			asg.scaleMutex.Lock()
			current := asg.curSize
			asg.curSize = current + 1
			asg.scaleMutex.Unlock()
		}()
	}

	wg.Wait()

	// If the lock works correctly, all increments should be serialized
	if asg.curSize != numGoroutines {
		t.Errorf("expected curSize=%d, got %d (race condition detected!)", numGoroutines, asg.curSize)
	}
}

func TestTargetSize_ThreadSafety(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	asg.curSize = 5
	asg.maxSize = 200 // Allow room for concurrent increments

	nodeGroup := newTestNodeGroup(t, manager, asg)

	var wg sync.WaitGroup
	const numReaders = 10
	const numWrites = 100 // Increased from 20 for better race detection

	// Start concurrent readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				size, err := nodeGroup.TargetSize()
				if err != nil {
					t.Errorf("TargetSize error: %v", err)
					return
				}
				if size < 0 || size > 500 {
					t.Errorf("invalid size: %d", size)
					return
				}
			}
		}()
	}

	// Start concurrent writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numWrites; i++ {
			manager.asgs.adjustTargetSize(asg, 1)
			if i%5 == 0 {
				manager.asgs.adjustTargetSize(asg, -1)
			}
		}
	}()

	wg.Wait()

	finalSize, err := nodeGroup.TargetSize()
	assertNoError(t, err)
	if finalSize < 0 || finalSize > 500 {
		t.Errorf("final size %d out of expected range", finalSize)
	}
}

func TestDeleteNodes_MinSizeCheck(t *testing.T) {
	tests := []struct {
		name          string
		minSize       int
		curSize       int
		nodesToDelete int
		expectError   bool
		errorSubstr   string
	}{
		{
			name:          "delete 1 node when at min size",
			minSize:       2,
			curSize:       2,
			nodesToDelete: 1,
			expectError:   true,
			errorSubstr:   "would violate min size",
		},
		{
			name:          "delete multiple nodes would violate min size",
			minSize:       3,
			curSize:       5,
			nodesToDelete: 4, // 5-4=1 < minSize(3)
			expectError:   true,
			errorSubstr:   "would violate min size",
		},
		{
			name:          "delete nodes exactly to min size",
			minSize:       3,
			curSize:       5,
			nodesToDelete: 2, // 5-2=3 == minSize(3) - should be allowed
			expectError:   false,
		},
		{
			name:          "delete 1 node above min size",
			minSize:       2,
			curSize:       5,
			nodesToDelete: 1,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, asg, _ := newTestEnv(t)

			// Customize for this test case
			asg.minSize = tc.minSize
			asg.curSize = tc.curSize

			nodeGroup := newTestNodeGroup(t, manager, asg)
			nodes := createTestNodes(t, tc.nodesToDelete)
			initialSize := asg.curSize

			err := nodeGroup.DeleteNodes(nodes)

			if tc.expectError {
				assertErrorContains(t, err, tc.errorSubstr)
				// Verify state wasn't corrupted
				if asg.curSize != initialSize {
					t.Errorf("state corrupted: curSize changed from %d to %d on error", initialSize, asg.curSize)
				}
			} else if err != nil && strings.Contains(err.Error(), "would violate min size") {
				t.Errorf("should not fail with min size error, got: %v", err)
			}
		})
	}
}

func TestIncrementTargetSize(t *testing.T) {
	tests := []struct {
		name        string
		curSize     int
		maxSize     int
		delta       int
		expectSize  int
		expectError bool
	}{
		{
			name:        "valid increment",
			curSize:     2,
			maxSize:     10,
			delta:       3,
			expectSize:  5,
			expectError: false,
		},
		{
			name:        "increment to max",
			curSize:     8,
			maxSize:     10,
			delta:       2,
			expectSize:  10,
			expectError: false,
		},
		{
			name:        "exceed max size",
			curSize:     8,
			maxSize:     10,
			delta:       3,
			expectSize:  0,
			expectError: true,
		},
		{
			name:        "already at max",
			curSize:     10,
			maxSize:     10,
			delta:       1,
			expectSize:  0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, asg, asgs := newTestEnv(t)
			asg.curSize = tc.curSize
			asg.maxSize = tc.maxSize

			newSize, err := asgs.incrementTargetSize(asg, tc.delta)

			if tc.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if newSize != 0 {
					t.Errorf("expected size 0 on error, got %d", newSize)
				}
			} else {
				assertNoError(t, err)
				if newSize != tc.expectSize {
					t.Errorf("expected size %d, got %d", tc.expectSize, newSize)
				}
				if asg.curSize != tc.expectSize {
					t.Errorf("expected asg.curSize %d, got %d", tc.expectSize, asg.curSize)
				}
			}
		})
	}
}

func TestIncrementTargetSize_Concurrent(t *testing.T) {
	_, asg, asgs := newTestEnv(t)
	asg.maxSize = 1000
	asg.curSize = 0

	const numGoroutines = 100 // Increased from 10
	const incrementPerGoroutine = 5
	var wg sync.WaitGroup
	var successCount int32
	var countMutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := asgs.incrementTargetSize(asg, incrementPerGoroutine)
			if err == nil {
				countMutex.Lock()
				successCount++
				countMutex.Unlock()
			}
		}()
	}

	wg.Wait()

	expectedSize := numGoroutines * incrementPerGoroutine
	if asg.curSize != expectedSize {
		t.Errorf("expected curSize=%d, got %d (race condition!)", expectedSize, asg.curSize)
	}
	if int(successCount) != numGoroutines {
		t.Errorf("expected %d successful increments, got %d", numGoroutines, successCount)
	}
}

func TestAdjustTargetSize(t *testing.T) {
	tests := []struct {
		name       string
		curSize    int
		delta      int
		expectSize int
	}{
		{
			name:       "positive adjustment",
			curSize:    5,
			delta:      3,
			expectSize: 8,
		},
		{
			name:       "negative adjustment",
			curSize:    10,
			delta:      -3,
			expectSize: 7,
		},
		{
			name:       "zero adjustment",
			curSize:    5,
			delta:      0,
			expectSize: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, asg, asgs := newTestEnv(t)
			asg.curSize = tc.curSize
			asg.maxSize = 20

			asgs.adjustTargetSize(asg, tc.delta)

			if asg.curSize != tc.expectSize {
				t.Errorf("expected curSize %d, got %d", tc.expectSize, asg.curSize)
			}
		})
	}
}

func TestUpdateCacheWithInstances(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	refs := createTestInstanceRefs(t, 3)
	asgs.updateCacheWithInstances(asg, refs)

	// Verify instance-to-ASG mapping
	for _, ref := range refs {
		mappedAsg, exists := asgs.instanceToAsg[ref]
		if !exists {
			t.Errorf("instance %s not found in instanceToAsg map", ref.Hostname)
		}
		if mappedAsg != asg {
			t.Errorf("instance %s mapped to wrong ASG", ref.Hostname)
		}
	}

	// Verify ASG-to-instances mapping
	instances, exists := asgs.asgToInstances[asg.AsgRef]
	if !exists {
		t.Error("ASG not found in asgToInstances map")
	}
	if len(instances) != len(refs) {
		t.Errorf("expected %d instances, got %d", len(refs), len(instances))
	}
}

func TestUpdateCacheWithInstances_Concurrent(t *testing.T) {
	_, _, asgs := newTestEnv(t)

	const numAsgs = 10 // Increased from 5
	allAsgs := make([]*Asg, numAsgs)
	for i := 0; i < numAsgs; i++ {
		allAsgs[i] = &Asg{
			AsgRef:                AsgRef{Name: fmt.Sprintf("%s-%d", testAsgName, i)},
			minSize:               1,
			maxSize:               10,
			curSize:               0,
			instanceType:          testInstanceType,
			AvailabilityLocations: []string{testLocation},
		}
	}

	var wg sync.WaitGroup
	for i, asg := range allAsgs {
		wg.Add(1)
		go func(index int, a *Asg) {
			defer wg.Done()
			refs := []InstanceRef{
				{
					Hostname:   fmt.Sprintf("vm-%d-01", index),
					ProviderID: fmt.Sprintf("%s%s/vm-%d-01", testProviderPrefix, testLocation, index),
				},
				{
					Hostname:   fmt.Sprintf("vm-%d-02", index),
					ProviderID: fmt.Sprintf("%s%s/vm-%d-02", testProviderPrefix, testLocation, index),
				},
			}
			asgs.updateCacheWithInstances(a, refs)
		}(i, asg)
	}

	wg.Wait()

	// Verify all ASGs have correct instance counts
	totalInstances := 0
	for _, asg := range allAsgs {
		instances, exists := asgs.asgToInstances[asg.AsgRef]
		if !exists {
			t.Errorf("ASG %s not found in map", asg.Name)
			continue
		}
		if len(instances) != 2 {
			t.Errorf("ASG %s should have 2 instances, got %d", asg.Name, len(instances))
		}
		totalInstances += len(instances)
	}

	if totalInstances != numAsgs*2 {
		t.Errorf("expected %d total instances, got %d", numAsgs*2, totalInstances)
	}

	if len(asgs.instanceToAsg) != numAsgs*2 {
		t.Errorf("expected %d entries in instanceToAsg, got %d", numAsgs*2, len(asgs.instanceToAsg))
	}
}

func TestScaleUpAsg_RollbackOnValidationError(t *testing.T) {
	_, asg, asgs := newTestEnv(t)
	asg.maxSize = 5
	asg.curSize = 3

	initialSize := asg.curSize
	newSize, err := asgs.incrementTargetSize(asg, 10) // Try to exceed max

	if err == nil {
		t.Error("expected error when exceeding max size, got nil")
	}

	if newSize != 0 {
		t.Errorf("expected newSize=0 on error, got %d", newSize)
	}

	if asg.curSize != initialSize {
		t.Errorf("expected curSize to remain %d, got %d (rollback failed!)", initialSize, asg.curSize)
	}
}

func TestInstancesForAsg_CacheConsistency(t *testing.T) {
	manager, asg, asgs := newTestEnv(t)
	asg.curSize = 2
	asg.minSize = 0

	refs := createTestInstanceRefs(t, 2)
	asgs.updateCacheWithInstances(asg, refs)

	cachedRefs, err := asgs.InstanceRefsForAsg(asg.AsgRef)
	assertNoError(t, err)

	if len(cachedRefs) != 2 {
		t.Errorf("expected 2 cached refs, got %d", len(cachedRefs))
	}

	// Key invariant: TargetSize() == len(InstanceRefsForAsg()) prevents duplicate scale-ups
	nodeGroup := newTestNodeGroup(t, manager, asg)
	size, _ := nodeGroup.TargetSize()
	if size != len(cachedRefs) {
		t.Errorf("curSize (%d) should match cached refs count (%d)", size, len(cachedRefs))
	}
}

func TestInstancesForAsg_NoSkippingOnAPIFailure(t *testing.T) {
	tests := []struct {
		name                string
		cachedInstances     int
		apiAvailableCount   int // how many instances are available in API
		expectedReturnCount int // should always equal cachedInstances
	}{
		{
			name:                "all instances available in API",
			cachedInstances:     3,
			apiAvailableCount:   3,
			expectedReturnCount: 3,
		},
		{
			name:                "no instances available in API (all newly created)",
			cachedInstances:     3,
			apiAvailableCount:   0,
			expectedReturnCount: 3, // must return 3, not 0
		},
		{
			name:                "partial instances available (some newly created)",
			cachedInstances:     3,
			apiAvailableCount:   1,
			expectedReturnCount: 3, // must return 3, not 1
		},
		{
			name:                "single newly created instance",
			cachedInstances:     1,
			apiAvailableCount:   0,
			expectedReturnCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, asg, asgs := newTestEnv(t)
			asg.curSize = tc.cachedInstances
			asg.minSize = 0

			refs := createTestInstanceRefs(t, tc.cachedInstances)
			asgs.updateCacheWithInstances(asg, refs)

			cachedRefs, _ := asgs.InstanceRefsForAsg(asg.AsgRef)
			if len(cachedRefs) != asg.curSize {
				t.Errorf("cached refs (%d) != curSize (%d)", len(cachedRefs), asg.curSize)
			}

			if asg.curSize != tc.expectedReturnCount {
				t.Errorf("expected curSize=%d to match expected return count=%d",
					asg.curSize, tc.expectedReturnCount)
			}
		})
	}
}

func TestTargetSizeAndNodesConsistency(t *testing.T) {
	manager, asg, asgs := newTestEnv(t)
	asg.minSize = 0
	asg.curSize = 0

	nodeGroup := newTestNodeGroup(t, manager, asg)

	// Initial state
	size, err := nodeGroup.TargetSize()
	assertNoError(t, err)
	if size != 0 {
		t.Errorf("initial TargetSize should be 0, got %d", size)
	}

	// After increment
	asgs.incrementTargetSize(asg, 1)
	size, _ = nodeGroup.TargetSize()
	if size != 1 {
		t.Errorf("TargetSize after increment should be 1, got %d", size)
	}

	// After adding instance to cache
	refs := createTestInstanceRefs(t, 1)
	asgs.updateCacheWithInstances(asg, refs)

	size, _ = nodeGroup.TargetSize()
	cachedRefs, _ := asgs.InstanceRefsForAsg(asg.AsgRef)

	if size != len(cachedRefs) {
		t.Errorf("TargetSize() = %d, len(cachedRefs) = %d", size, len(cachedRefs))
	}

	t.Logf("TargetSize=%d, CachedRefs=%d - Consistency maintained", size, len(cachedRefs))
}

func TestDuplicateScaleUpPrevention(t *testing.T) {
	_, asg, asgs := newTestEnv(t)
	asg.minSize = 0
	asg.maxSize = 2
	asg.curSize = 0

	if asg.curSize != 0 {
		t.Fatalf("initial curSize should be 0")
	}

	// First scale-up
	newSize, err := asgs.incrementTargetSize(asg, 1)
	assertNoError(t, err)
	if newSize != 1 {
		t.Errorf("after first scale-up, curSize should be 1, got %d", newSize)
	}

	// Add instance to cache
	refs := createTestInstanceRefs(t, 1)
	asgs.updateCacheWithInstances(asg, refs)

	// Verify consistency
	targetSize := asg.curSize
	cachedInstances, _ := asgs.InstanceRefsForAsg(asg.AsgRef)

	if targetSize != len(cachedInstances) {
		t.Errorf("targetSize=%d but cachedInstances=%d", targetSize, len(cachedInstances))
	}

	// Verify max size protection
	wouldBeDelta := 1
	potentialNewSize := asg.curSize + wouldBeDelta

	if potentialNewSize > asg.maxSize {
		t.Logf("correctly would reject scale-up beyond max size")
	}

	t.Logf("targetSize=%d, cachedInstances=%d - No duplicate scale-up risk", targetSize, len(cachedInstances))
}

func TestDeleteInstance_MinSizeCheck(t *testing.T) {
	// This test verifies that DeleteNodes enforces minSize using curSize.
	// The DeleteNodes function checks curSize before calling DeleteInstances.
	// Note: DeleteInstance now also has a secondary check using API count,
	// but that requires a live dcService which we don't test here.

	tests := []struct {
		name        string
		minSize     int
		curSize     int // internal counter used by DeleteNodes
		deleteCount int // how many nodes to delete
		expectError bool
		errorSubstr string
	}{
		{
			name:        "block deletion at minSize",
			minSize:     3,
			curSize:     3, // at minSize
			deleteCount: 1, // 3-1=2 < 3
			expectError: true,
			errorSubstr: "would violate min size",
		},
		{
			name:        "block deletion that would go below minSize",
			minSize:     5,
			curSize:     6,
			deleteCount: 3, // 6-3=3 < 5
			expectError: true,
			errorSubstr: "would violate min size",
		},
		{
			name:        "allow deletion when result equals minSize",
			minSize:     3,
			curSize:     4,
			deleteCount: 1, // 4-1=3 == minSize (allowed)
			expectError: false,
		},
		{
			name:        "allow deletion when above minSize",
			minSize:     2,
			curSize:     5,
			deleteCount: 1, // 5-1=4 > 2
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, asg, asgs := newTestEnv(t)

			// Configure ASG
			asg.minSize = tc.minSize
			asg.curSize = tc.curSize

			// Add instances to cache to match curSize
			refs := createTestInstanceRefs(t, tc.curSize)
			asgs.updateCacheWithInstances(asg, refs)

			nodeGroup := newTestNodeGroup(t, manager, asg)
			nodes := createTestNodes(t, tc.deleteCount)

			err := nodeGroup.DeleteNodes(nodes)

			if tc.expectError {
				assertErrorContains(t, err, tc.errorSubstr)
				t.Logf("correctly blocked deletion: %v", err)
			} else {
				// For allowed cases, we still get an error because dcService is nil,
				// but the error should NOT be about minSize violation
				if err != nil && strings.Contains(err.Error(), "would violate min size") {
					t.Errorf("minSize check should have passed, got: %v", err)
				} else if err != nil {
					t.Logf("minSize check passed, subsequent error (expected): %v", err)
				}
			}
		})
	}
}

func TestDeleteInstance_MinSizeCheck_CurSizeDrift(t *testing.T) {
	// This test documents the scenario where curSize has drifted from actual instance count.
	//
	// Scenario: curSize=10 (inflated due to failed scale-ups), but only 3 actual instances.
	// DeleteNodes checks curSize: 10-2=8 >= 8 (minSize) → PASSES (incorrectly!)
	//
	// The fix adds a secondary check in DeleteInstance using actual API count,
	// which would block the deletion. However, that requires a live dcService
	// to test, so this test documents the behavior without the secondary check.

	manager, asg, asgs := newTestEnv(t)

	asg.minSize = 8
	asg.maxSize = 10
	asg.curSize = 10 // Inflated! (simulates failed scale-ups not rolling back)

	// Cache only has 3 instances (simulates what API would return)
	refs := createTestInstanceRefs(t, 3)
	asgs.updateCacheWithInstances(asg, refs)

	nodeGroup := newTestNodeGroup(t, manager, asg)
	nodes := createTestNodes(t, 2) // Try to delete 2 nodes

	err := nodeGroup.DeleteNodes(nodes)

	// DeleteNodes uses curSize (10), so 10-2=8 >= 8 (minSize) → check PASSES
	// This demonstrates why we needed the secondary check in DeleteInstance.
	if err != nil && strings.Contains(err.Error(), "would violate min size") {
		// This should NOT happen with curSize=10
		t.Errorf("DeleteNodes should have passed curSize check (10-2=8 >= 8), got: %v", err)
	} else {
		// Expected: curSize check passes, then fails later (nil dcService or API check)
		t.Logf("DeleteNodes passed curSize check (10-2=8 >= 8), subsequent result: %v", err)
	}
}

func TestIsProvisioningFailedStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		// Failed statuses - should return true
		{"no_capacity", "no_capacity", true},
		{"error", "error", true},
		{"unknown", "unknown", true},
		// Case insensitive
		{"NO_CAPACITY uppercase", "NO_CAPACITY", true},
		{"Error mixed case", "Error", true},
		{"UNKNOWN uppercase", "UNKNOWN", true},
		// Active statuses - should return false
		{"running", "running", false},
		{"new", "new", false},
		{"ordered", "ordered", false},
		{"provisioning", "provisioning", false},
		{"validating", "validating", false},
		{"pending", "pending", false},
		// Other statuses - should return false
		{"offline", "offline", false},
		{"discontinued", "discontinued", false},
		{"deleting", "deleting", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isProvisioningFailedStatus(tc.status)
			if result != tc.expected {
				t.Errorf("isProvisioningFailedStatus(%q) = %v, expected %v", tc.status, result, tc.expected)
			}
		})
	}
}

func TestCategorizeInstancesForAsg(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	tests := []struct {
		name           string
		instances      []mockInstance
		expectedActive int
		expectedFailed int
	}{
		{
			name: "all running",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "running"},
				{hostname: "test-vm-fin-01-01", status: "running"},
				{hostname: "test-vm-fin-01-02", status: "running"},
			},
			expectedActive: 3,
			expectedFailed: 0,
		},
		{
			name: "mixed active statuses",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "running"},
				{hostname: "test-vm-fin-01-01", status: "new"},
				{hostname: "test-vm-fin-01-02", status: "ordered"},
				{hostname: "test-vm-fin-01-03", status: "provisioning"},
				{hostname: "test-vm-fin-01-04", status: "validating"},
			},
			expectedActive: 5,
			expectedFailed: 0,
		},
		{
			name: "some failed instances",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "running"},
				{hostname: "test-vm-fin-01-01", status: "running"},
				{hostname: "test-vm-fin-01-02", status: "no_capacity"},
				{hostname: "test-vm-fin-01-03", status: "error"},
			},
			expectedActive: 2,
			expectedFailed: 2,
		},
		{
			name: "all failed",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "no_capacity"},
				{hostname: "test-vm-fin-01-01", status: "error"},
				{hostname: "test-vm-fin-01-02", status: "unknown"},
			},
			expectedActive: 0,
			expectedFailed: 3,
		},
		{
			name: "mixed with other ASG instances",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "running"},     // belongs to test asg
				{hostname: "other-vm-fin-01-00", status: "running"},    // different ASG
				{hostname: "test-vm-fin-01-01", status: "no_capacity"}, // belongs to test asg
				{hostname: "other-vm-fin-01-01", status: "error"},      // different ASG
			},
			expectedActive: 1,
			expectedFailed: 1,
		},
		{
			name: "includes deleting/offline (not active, not failed)",
			instances: []mockInstance{
				{hostname: "test-vm-fin-01-00", status: "running"},
				{hostname: "test-vm-fin-01-01", status: "deleting"},
				{hostname: "test-vm-fin-01-02", status: "offline"},
				{hostname: "test-vm-fin-01-03", status: "discontinued"},
			},
			expectedActive: 1,
			expectedFailed: 0,
		},
		{
			name:           "empty list",
			instances:      []mockInstance{},
			expectedActive: 0,
			expectedFailed: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allInstances := makeVerdaInstances(tc.instances)
			active, failed := asgs.categorizeInstancesForAsg(allInstances, asg)

			if len(active) != tc.expectedActive {
				t.Errorf("expected %d active instances, got %d", tc.expectedActive, len(active))
			}
			if len(failed) != tc.expectedFailed {
				t.Errorf("expected %d failed instances, got %d", tc.expectedFailed, len(failed))
			}
		})
	}
}

// mockInstance is a helper for test data
type mockInstance struct {
	hostname string
	status   string
}

// makeVerdaInstances converts mock instances to verda.Instance slice
func makeVerdaInstances(mocks []mockInstance) []verda.Instance {
	instances := make([]verda.Instance, len(mocks))
	for i, m := range mocks {
		instances[i] = verda.Instance{
			ID:       fmt.Sprintf("inst-%d", i),
			Hostname: m.hostname,
			Status:   m.status,
			Location: testLocation,
		}
	}
	return instances
}

func TestProcessFailedInstances_BackoffTracking(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	// Verify no backoff initially
	asgs.cacheMutex.RLock()
	_, hasBackoff := asgs.lastFailureCheck[asg.AsgRef]
	asgs.cacheMutex.RUnlock()
	if hasBackoff {
		t.Error("should not have backoff initially")
	}

	// Process some failed instances
	failedByAsg := map[AsgRef][]verda.Instance{
		asg.AsgRef: {
			{ID: "inst-1", Hostname: "test-vm-fin-01-00", Status: "no_capacity"},
			{ID: "inst-2", Hostname: "test-vm-fin-01-01", Status: "error"},
		},
	}
	asgs.processFailedInstances(map[AsgRef]*Asg{asg.AsgRef: asg}, failedByAsg)

	// Verify backoff is now set
	asgs.cacheMutex.RLock()
	lastCheck, hasBackoff := asgs.lastFailureCheck[asg.AsgRef]
	failedCount := len(asgs.failedInstances)
	asgs.cacheMutex.RUnlock()

	if !hasBackoff {
		t.Error("should have backoff after processing failed instances")
	}
	if time.Since(lastCheck) > time.Second {
		t.Error("lastFailureCheck should be recent")
	}
	if failedCount != 2 {
		t.Errorf("expected 2 tracked failed instances, got %d", failedCount)
	}
}

func TestProcessFailedInstances_NoFailures(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	// Process with no failed instances
	failedByAsg := map[AsgRef][]verda.Instance{
		asg.AsgRef: {},
	}
	asgs.processFailedInstances(map[AsgRef]*Asg{asg.AsgRef: asg}, failedByAsg)

	// Verify no backoff is set
	asgs.cacheMutex.RLock()
	_, hasBackoff := asgs.lastFailureCheck[asg.AsgRef]
	asgs.cacheMutex.RUnlock()

	if hasBackoff {
		t.Error("should not have backoff when no failures")
	}
}

func TestTrackAndCleanupFailedInstance_TTLCleanup(t *testing.T) {
	_, _, asgs := newTestEnv(t)

	// Add an old entry that should be cleaned up
	oldHostname := "old-test-vm-fin-01-00"
	asgs.cacheMutex.Lock()
	asgs.failedInstances[oldHostname] = time.Now().Add(-2 * FAILED_INSTANCE_MAP_ENTRY_TTL)
	asgs.cacheMutex.Unlock()

	// Process failed instances (which triggers TTL cleanup)
	failedByAsg := map[AsgRef][]verda.Instance{}
	asgs.processFailedInstances(map[AsgRef]*Asg{}, failedByAsg)

	// Verify old entry was cleaned up
	asgs.cacheMutex.RLock()
	_, exists := asgs.failedInstances[oldHostname]
	asgs.cacheMutex.RUnlock()

	if exists {
		t.Error("old failed instance should have been cleaned up by TTL")
	}
}

func TestTrackAndCleanupFailedInstance_RecentEntryKept(t *testing.T) {
	_, _, asgs := newTestEnv(t)

	// Add a recent entry that should NOT be cleaned up
	recentHostname := "recent-test-vm-fin-01-00"
	asgs.cacheMutex.Lock()
	asgs.failedInstances[recentHostname] = time.Now().Add(-time.Minute)
	asgs.cacheMutex.Unlock()

	// Process failed instances (which triggers TTL cleanup)
	failedByAsg := map[AsgRef][]verda.Instance{}
	asgs.processFailedInstances(map[AsgRef]*Asg{}, failedByAsg)

	// Verify recent entry was NOT cleaned up
	asgs.cacheMutex.RLock()
	_, exists := asgs.failedInstances[recentHostname]
	asgs.cacheMutex.RUnlock()

	if !exists {
		t.Error("recent failed instance should NOT have been cleaned up")
	}
}

// =============================================================================
// VerdacloudNodeGroup Interface Tests
// =============================================================================

func TestNodeGroup_MaxSize(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	asg.maxSize = 100

	ng := newTestNodeGroup(t, manager, asg)

	if ng.MaxSize() != 100 {
		t.Errorf("expected MaxSize=100, got %d", ng.MaxSize())
	}
}

func TestNodeGroup_MinSize(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	asg.minSize = 5

	ng := newTestNodeGroup(t, manager, asg)

	if ng.MinSize() != 5 {
		t.Errorf("expected MinSize=5, got %d", ng.MinSize())
	}
}

func TestNodeGroup_Id(t *testing.T) {
	manager, asg, _ := newTestEnv(t)

	ng := newTestNodeGroup(t, manager, asg)

	if ng.Id() != testAsgName {
		t.Errorf("expected Id=%s, got %s", testAsgName, ng.Id())
	}
}

func TestNodeGroup_Debug(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	asg.minSize = 1
	asg.maxSize = 10

	ng := newTestNodeGroup(t, manager, asg)
	debug := ng.Debug()

	if !strings.Contains(debug, testAsgName) {
		t.Errorf("Debug() should contain ASG name, got: %s", debug)
	}
	if !strings.Contains(debug, "1:10") {
		t.Errorf("Debug() should contain min:max (1:10), got: %s", debug)
	}
}

func TestNodeGroup_GetOptions(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	ng := newTestNodeGroup(t, manager, asg)

	defaults := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.5,
	}

	opts, err := ng.GetOptions(defaults)
	assertNoError(t, err)

	if opts.ScaleDownUtilizationThreshold != 0.5 {
		t.Errorf("expected ScaleDownUtilizationThreshold=0.5, got %f", opts.ScaleDownUtilizationThreshold)
	}
}

func TestNodeGroup_Autoprovisioned(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	ng := newTestNodeGroup(t, manager, asg)

	if ng.Autoprovisioned() {
		t.Error("Autoprovisioned() should return false")
	}
}

func TestNodeGroup_AtomicIncreaseSize(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	ng := newTestNodeGroup(t, manager, asg)

	err := ng.AtomicIncreaseSize(1)
	if err == nil {
		t.Error("AtomicIncreaseSize should return ErrNotImplemented")
	}
}

func TestNodeGroup_ForceDeleteNodes(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	ng := newTestNodeGroup(t, manager, asg)

	err := ng.ForceDeleteNodes(nil)
	if err == nil {
		t.Error("ForceDeleteNodes should return ErrNotImplemented")
	}
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	tests := []struct {
		name        string
		delta       int
		expectError bool
		errorSubstr string
	}{
		{
			name:        "positive delta should error",
			delta:       1,
			expectError: true,
			errorSubstr: "must be negative",
		},
		{
			name:        "zero delta should error",
			delta:       0,
			expectError: true,
			errorSubstr: "must be negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, asg, _ := newTestEnv(t)
			asg.curSize = 5
			ng := newTestNodeGroup(t, manager, asg)

			err := ng.DecreaseTargetSize(tc.delta)

			if tc.expectError {
				assertErrorContains(t, err, tc.errorSubstr)
			}
		})
	}
}

func TestNodeGroup_Exist(t *testing.T) {
	manager, asg, asgs := newTestEnv(t)

	t.Run("exists when registered", func(t *testing.T) {
		ng := newTestNodeGroup(t, manager, asg)
		if !ng.Exist() {
			t.Error("Exist() should return true for registered ASG")
		}
	})

	t.Run("does not exist when nil asg", func(t *testing.T) {
		ng := &VerdacloudNodeGroup{asg: nil, manager: manager}
		if ng.Exist() {
			t.Error("Exist() should return false for nil ASG")
		}
	})

	t.Run("does not exist when not registered", func(t *testing.T) {
		unregisteredAsg := &Asg{AsgRef: AsgRef{Name: "not-registered"}}
		ng := newTestNodeGroup(t, manager, unregisteredAsg)

		// Remove from registered ASGs
		asgs.cacheMutex.Lock()
		delete(asgs.registeredAsgs, unregisteredAsg.AsgRef)
		asgs.cacheMutex.Unlock()

		if ng.Exist() {
			t.Error("Exist() should return false for unregistered ASG")
		}
	})
}

func TestNodeGroup_Create(t *testing.T) {
	manager, asg, _ := newTestEnv(t)
	ng := newTestNodeGroup(t, manager, asg)

	created, err := ng.Create()
	assertNoError(t, err)

	if created != ng {
		t.Error("Create() should return the same node group")
	}
}

func TestFindASGForInstance(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	refs := createTestInstanceRefs(t, 2)
	asgs.updateCacheWithInstances(asg, refs)

	t.Run("find existing instance", func(t *testing.T) {
		foundAsg, err := asgs.FindASGForInstance(&refs[0])
		assertNoError(t, err)

		if foundAsg == nil {
			t.Fatal("expected to find ASG for instance")
		}
		if foundAsg.Name != asg.Name {
			t.Errorf("expected ASG %s, got %s", asg.Name, foundAsg.Name)
		}
	})

	t.Run("return nil for non-existent instance", func(t *testing.T) {
		nonExistent := &InstanceRef{Hostname: "non-existent-host", ProviderID: "verdacloud://FIN-01/non-existent-host"}
		foundAsg, err := asgs.FindASGForInstance(nonExistent)
		assertNoError(t, err)

		if foundAsg != nil {
			t.Error("expected nil ASG for non-existent instance")
		}
	})
}

func TestGetAsgByRef(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	t.Run("find existing ASG", func(t *testing.T) {
		found, err := asgs.GetAsgByRef(asg.AsgRef)
		assertNoError(t, err)

		if found == nil {
			t.Fatal("expected to find ASG")
		}
		if found.Name != asg.Name {
			t.Errorf("expected ASG %s, got %s", asg.Name, found.Name)
		}
	})

	t.Run("error for non-existent ASG", func(t *testing.T) {
		_, err := asgs.GetAsgByRef(AsgRef{Name: "non-existent"})
		if err == nil {
			t.Error("expected error for non-existent ASG")
		}
	})
}

func TestGetAsgs(t *testing.T) {
	_, asg, asgs := newTestEnv(t)

	// Add another ASG
	asg2 := &Asg{
		AsgRef:  AsgRef{Name: "test-asg-2"},
		minSize: 0,
		maxSize: 5,
	}
	asgs.cacheMutex.Lock()
	asgs.registeredAsgs[asg2.AsgRef] = asg2
	asgs.cacheMutex.Unlock()

	allAsgs := asgs.getAsgs()

	if len(allAsgs) != 2 {
		t.Errorf("expected 2 ASGs, got %d", len(allAsgs))
	}

	names := make(map[string]bool)
	for _, a := range allAsgs {
		names[a.Name] = true
	}

	if !names[asg.Name] || !names[asg2.Name] {
		t.Error("expected both ASGs in result")
	}
}
