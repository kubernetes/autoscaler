package fort

import (
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

func TestSyncPropagation_Chained(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	
	// Chain: s1 -> flatMap -> grouper
	flatMap := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map: func(i int) ([]int, error) { return []int{i}, nil },
		Over: s1,
	})

	grouper := QueryInformer(&GroupBy[int, int]{
		Lock: lock,
		Select: func(fields []GroupField) (int, error) { return int(fields[0].(int64)), nil },
		From: flatMap,
		GroupBy: func(i int) (any, []GroupField) { return [1]int{0}, []GroupField{Count()} },
	})

	if grouper.HasSynced() {
		t.Errorf("Grouper should not be synced yet")
	}

	s1.SetHasSynced()

	// Wait a bit for goroutines to propagate sync
	success := false
	for i := 0; i < 10; i++ {
		if grouper.HasSynced() {
			success = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !success {
		t.Errorf("Sync signal did not propagate through the chain to grouper")
	}
}

func TestSyncPropagation_Join(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	s2 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	joined := QueryInformer(&Join[int, int, int]{
		Lock:   lock,
		Select: func(l, r int) (int, error) { return l + r, nil },
		From:   s1,
		Join:   s2,
		On: func(l, r int) any { return [1]int{0} },
	})

	if joined.HasSynced() {
		t.Errorf("Joined should not be synced yet")
	}

	s1.SetHasSynced()
	time.Sleep(10 * time.Millisecond)

	if joined.HasSynced() {
		t.Errorf("Joined should not be synced when only one source is synced")
	}

	s2.SetHasSynced()
	
	success := false
	for i := 0; i < 10; i++ {
		if joined.HasSynced() {
			success = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !success {
		t.Errorf("Joined did not sync after both sources synced")
	}
}
