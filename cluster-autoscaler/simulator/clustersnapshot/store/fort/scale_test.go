package fort

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

// TestScale_MillionEntries performs high-scale performance benchmarking
// with 1,000,000 objects in a multi-stage query pipeline.
func TestScale_MillionEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-scale test in short mode")
	}

	const numItems = 1_000_000
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	// Pipeline: Source -> FlatMap (1:2 distinct strings) -> GroupBy (i % 1000)
	// Using strings ensures no identity collisions in the FlatMap output indexer.
	fm := QueryInformer(&FlatMap[string, int]{
		Lock: lock,
		Map: func(i int) ([]string, error) {
			return []string{
				fmt.Sprintf("item-%d-a", i),
				fmt.Sprintf("item-%d-b", i),
			}, nil
		},
		Over: source,
	})

	type GroupResult struct {
		Name  string
		Count int64
	}

	grouper := QueryInformer(&GroupBy[*GroupResult, string]{
		Lock: lock,
		Select: func(fields []GroupField) (*GroupResult, error) {
			return &GroupResult{
				Name:  fields[0].(string),
				Count: fields[1].(int64),
			}, nil
		},
		From: fm,
		GroupBy: func(s string) (any, []GroupField) {
			// Extract original index for grouping: "item-123-a" -> 123
			parts := strings.Split(s, "-")
			idx, _ := strconv.Atoi(parts[1])
			groupName := fmt.Sprintf("group-%d", idx%1000)
			return groupName, []GroupField{AnyValue(groupName), Count()}
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go fm.Run(stopCh)
	go grouper.Run(stopCh)

	// 1. Ingestion Phase
	fmt.Printf("Ingesting %d items...\n", numItems)
	start := time.Now()
	for i := 0; i < numItems; i++ {
		source.OnAdd(i, false)
		if i > 0 && i%200_000 == 0 {
			fmt.Printf("  ... %d items ingested (avg %.2fµs/item)\n", i, float64(time.Since(start).Microseconds())/float64(i))
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Ingestion complete: %v (%.2fµs/item)\n", elapsed, float64(elapsed.Microseconds())/numItems)

	// 2. Memory Measurement
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory allocated: %.2f MB\n", float64(m.Alloc)/1024/1024)

	// 3. Cloning Phase (The core performance feature)
	fmt.Println("Performing high-scale clones...")
	start = time.Now()
	numClones := 100
	for i := 0; i < numClones; i++ {
		locks := SnapshotLockDomain(grouper)
		ns := source.Clone(nil)
		nfm := fm.Clone([]cache.SharedInformer{ns})
		_ = grouper.Clone([]cache.SharedInformer{nfm})
		locks.Unlock()
	}
	elapsed = time.Since(start)
	fmt.Printf("Cloning (Source+FlatMap+Grouper) latency at 1M items: %v per snapshot\n", elapsed/time.Duration(numClones))

	// 4. Update Latency Phase
	fmt.Println("Measuring update latency at scale...")
	start = time.Now()
	numUpdates := 1000
	for i := 0; i < numUpdates; i++ {
		// Update item-i to item-(i+numItems)
		source.OnUpdate(i, i+numItems)
	}
	elapsed = time.Since(start)
	fmt.Printf("Update latency at scale: %.2fµs/update (full pipeline propagation)\n", float64(elapsed.Microseconds())/float64(numUpdates))

	// 5. Verification
	fmt.Println("Verifying final aggregation state...")
	var list []any
	found := false
	for i := 0; i < 100; i++ {
		list = grouper.GetStore().List()
		if len(list) == 1000 {
			found = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	if !found {
		t.Errorf("Expected 1000 groups, got %d", len(list))
	} else {
		// Each group should have exactly (numItems * 2 / 1000) = 2000 items
		// We need to find the item in the list because we don't know the exact key DefaultKeyFunc generated for the pointer.
		var foundZero bool
		for _, it := range list {
			res := it.(*GroupResult)
			if res.Name == "group-0" {
				foundZero = true
				if res.Count != 2000 {
					t.Errorf("Expected group-0 to have 2000 items, got %d", res.Count)
				}
				break
			}
		}
		if !foundZero {
			t.Errorf("group-0 not found in results")
		}
	}
}
