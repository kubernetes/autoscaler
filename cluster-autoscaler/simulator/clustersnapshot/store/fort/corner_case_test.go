package fort

import (
	"errors"
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestFlatMap_CornerCases(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	
	// Test 1: Returning empty slice should delete existing items
	flatMap := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map: func(i int) ([]int, error) {
			if i == 0 {
				return []int{}, nil
			}
			return []int{i, i * 10}, nil
		},
		Over: source,
	})

	var results []int
	flatMap.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { results = append(results, obj.(int)) },
		DeleteFunc: func(obj any) {
			val := obj.(int)
			for i, v := range results {
				if v == val {
					results = append(results[:i], results[i+1:]...)
					break
				}
			}
		},
	})

	source.OnAdd(1, true) // results: [1, 10]
	if len(results) != 2 {
		t.Errorf("Expected 2 items, got %v", results)
	}

	source.OnUpdate(1, 0) // results: []
	if len(results) != 0 {
		t.Errorf("Expected 0 items after empty update, got %v", results)
	}

	// Test 2: Error in Map function
	source.OnUpdate(0, 5) // results: [5, 50]
	
	errFlatMap := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map: func(i int) ([]int, error) {
			return nil, errors.New("map error")
		},
		Over: source,
	})
	
	errFlatMap.AddEventHandler(cache.ResourceEventHandlerFuncs{})
	source.OnUpdate(5, 6) // Should not crash
}

func TestJoiner_CornerCases(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	s2 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	joined := QueryInformer(&Join[int, int, int]{
		Lock: lock,
		Select: func(l, r int) (int, error) { return l + r, nil },
		From:   s1,
		Join:   s2,
		On:     func(l, r int) any { return [1]int{0} }, // All items join
	})

	var results []int
	joined.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { results = append(results, obj.(int)) },
		DeleteFunc: func(obj any) {
			val := obj.(int)
			for i, v := range results {
				if v == val {
					results = append(results[:i], results[i+1:]...)
					break
				}
			}
		},
	})

	// 1. One side empty
	s1.OnAdd(1, true)
	if len(results) != 0 {
		t.Errorf("Expected 0 results when one side is empty")
	}

	// 2. Multi-match deletion
	s2.OnAdd(10, true) // results: [11]
	s1.OnAdd(2, false) // results: [11, 12]
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %v", results)
	}

	s2.OnDelete(10) // Should delete both 11 and 12
	if len(results) != 0 {
		t.Errorf("Expected 0 results after multi-match delete, got %v", results)
	}
}

func TestGrouper_CornerCases(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	grouper := QueryInformer(&GroupBy[int64, int]{
		Lock:   lock,
		Select: func(fields []GroupField) (int64, error) { return fields[1].(int64), nil },
		From:   source,
		GroupBy: func(i int) (any, []GroupField) {
			return 0, []GroupField{AnyValue(i), Sum(int64(i))}
		},
	})

	var latestSum int64
	var count int
	grouper.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { latestSum = obj.(int64); count++ },
		UpdateFunc: func(old, new any) { latestSum = new.(int64) },
		DeleteFunc: func(obj any) { count-- },
	})

	// 1. Empty source - handled by informer replay (0 items)
	if count != 0 {
		t.Errorf("Expected count 0 for empty source")
	}

	// 2. Sum reaching zero but items still in group
	source.OnAdd(10, true)  // Sum: 10
	source.OnAdd(-10, false) // Sum: 0
	if latestSum != 0 || count != 1 {
		t.Errorf("Expected sum 0 and count 1, got sum %d count %d", latestSum, count)
	}

	// 3. AnyValue consistency
	source.OnDelete(10)
	source.OnDelete(-10)
	if count != 0 {
		t.Errorf("Expected group deleted")
	}
}

func TestGrouper_AnyValueConsistency(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	grouper := QueryInformer(&GroupBy[int, int]{
		Lock:   lock,
		Select: func(fields []GroupField) (int, error) { return fields[0].(int), nil },
		From:   source,
		GroupBy: func(i int) (any, []GroupField) {
			return 0, []GroupField{AnyValue(i)}
		},
	})

	var result int
	grouper.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { result = obj.(int) },
		UpdateFunc: func(old, new any) { result = new.(int) },
	})

	source.OnAdd(100, true)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}

	source.OnAdd(200, false)
	// result should be 100 or 200 (implementation detail, but must be one of them)
	if result != 100 && result != 200 {
		t.Errorf("Expected 100 or 200, got %d", result)
	}

	oldResult := result
	source.OnDelete(oldResult)
	
	// Now result must be the OTHER one
	expected := 100
	if oldResult == 100 {
		expected = 200
	}
	if result != expected {
		t.Errorf("Expected %d after deleting %d, got %d", expected, oldResult, result)
	}
}

func TestGrouper_FilterTransitions(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	grouper := QueryInformer(&GroupBy[int64, int]{
		Lock:   lock,
		Select: func(fields []GroupField) (int64, error) { return fields[0].(int64), nil },
		From:   source,
		Where:  func(i int) bool { return i > 0 }, // Only positive
		GroupBy: func(i int) (any, []GroupField) {
			return 0, []GroupField{Sum(int64(i))}
		},
	})

	var latestSum int64
	var count int
	grouper.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { latestSum = obj.(int64); count++ },
		UpdateFunc: func(old, new any) { latestSum = new.(int64) },
		DeleteFunc: func(obj any) { count-- },
	})

	// 1. Add item that fails filter
	source.OnAdd(-10, true)
	if count != 0 {
		t.Errorf("Expected 0 groups, got %d", count)
	}

	// 2. Update item to pass filter (-10 -> 5)
	source.OnUpdate(-10, 5)
	if count != 1 || latestSum != 5 {
		t.Errorf("Expected 1 group with sum 5, got count %d sum %d", count, latestSum)
	}

	// 3. Update item to fail filter (5 -> -5)
	source.OnUpdate(5, -5)
	if count != 0 {
		t.Errorf("Expected 0 groups after item filtered out, got %d", count)
	}
}

func TestClone_Unsynced(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	query := QueryInformer(&Select[int, int]{
		Lock:   lock,
		Select: func(i int) (int, error) { return i, nil },
		From:   source,
	})

	// Clone BEFORE sync
	ns := source.Clone(nil)
	nq := query.Clone([]cache.SharedInformer{ns})

	if nq.HasSynced() {
		t.Errorf("Cloned query should not be synced if parent wasn't")
	}

	source.SetHasSynced()
	// Clones have their own lifecycle but inherit the fact that they are NOT synced yet.
	// We need to trigger sync on the clones if we want them synced.
	if nq.HasSynced() {
		t.Errorf("Clone should require its own SetHasSynced or sync from source")
	}
}

func TestClone_DistinctCOWIntegrity(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	grouper := QueryInformer(&GroupBy[[]any, int]{
		Lock: lock,
		Select: func(fields []GroupField) ([]any, error) {
			return fields[0].([]any), nil
		},
		From: source,
		GroupBy: func(i int) (any, []GroupField) {
			return 0, []GroupField{Distinct(i)}
		},
	})

	source.OnAdd(1, true)

	// Clone
	ns := source.Clone(nil)
	nq := grouper.Clone([]cache.SharedInformer{ns})

	var cloneResults []any
	nq.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { cloneResults = obj.([]any) },
		UpdateFunc: func(old, new any) { cloneResults = new.([]any) },
	})

	// Initial state of clone should have [1]
	if len(cloneResults) != 1 || cloneResults[0] != 1 {
		t.Errorf("Expected [1] in clone, got %v", cloneResults)
	}

	// Add to ORIGINAL source
	source.OnAdd(2, false)

	// Clone should still have [1] (COW isolation)
	if len(cloneResults) != 1 || cloneResults[0] != 1 {
		t.Errorf("Clone isolation failed! Clone results changed to %v after original update", cloneResults)
	}
}

func TestBTreeMap_EmptyKeys(t *testing.T) {
	m := NewBTreeMap[int]()
	m.Set("", 42)
	val, ok := m.Get("")
	if !ok || val != 42 {
		t.Errorf("BTreeMap failed with empty key")
	}

	m.Delete("")
	if m.Len() != 0 {
		t.Errorf("BTreeMap Delete failed with empty key")
	}
}
