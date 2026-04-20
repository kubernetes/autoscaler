package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestJoiner_OnUpdate(t *testing.T) {
	lock := NewLockGroup()
	left := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	right := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type L struct{ ID int; Val string }
	type R struct{ ID int; Val string }
	type Joined struct{ LVal, RVal string }

	joinedInformer := QueryInformer(&Join[Joined, L, R]{
		Lock: lock,
		Select: func(l L, r R) (Joined, error) {
			return Joined{LVal: l.Val, RVal: r.Val}, nil
		},
		From: left,
		Join: right,
		On: func(l L, r R) any {
			if l.ID != 0 { return [1]int{l.ID} }
			return [1]int{r.ID}
		},
	})

	var results []Joined
	joinedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			results = append(results, obj.(Joined))
		},
		UpdateFunc: func(old, new any) {
			oldJ := old.(Joined)
			newJ := new.(Joined)
			for i, r := range results {
				if r == oldJ {
					results[i] = newJ
					return
				}
			}
		},
		DeleteFunc: func(obj any) {
			target := obj.(Joined)
			for i, r := range results {
				if r == target {
					results = append(results[:i], results[i+1:]...)
					break
				}
			}
		},
	})

	l1 := L{ID: 1, Val: "L1"}
	r1 := R{ID: 1, Val: "R1"}

	left.OnAdd(l1, true)
	right.OnAdd(r1, true)

	if len(results) != 1 || results[0].LVal != "L1" || results[0].RVal != "R1" {
		t.Errorf("Initial join failed: %v", results)
	}

	// Update left item
	l1Updated := L{ID: 1, Val: "L1-Updated"}
	left.OnUpdate(l1, l1Updated)

	if len(results) != 1 || results[0].LVal != "L1-Updated" {
		t.Errorf("Update join failed: %v", results)
	}

	// Change key of left item (should break join)
	l1NewKey := L{ID: 2, Val: "L1-NewKey"}
	left.OnUpdate(l1Updated, l1NewKey)

	if len(results) != 0 {
		t.Errorf("Join should have been broken after key change, got: %v", results)
	}
}

func TestJoiner_Where(t *testing.T) {
	lock := NewLockGroup()
	left := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	right := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type Item struct{ ID int; Val int }
	
	joinedInformer := QueryInformer(&Join[int, Item, Item]{
		Lock:   lock,
		Select: func(l, r Item) (int, error) { return l.ID + r.ID, nil },
		From:   left,
		Join:   right,
		On: func(l, r Item) any {
			if l.ID != 0 { return [1]int{l.ID} }
			return [1]int{r.ID}
		},
		Where: func(l, r Item) bool {
			return l.Val > 10 && r.Val > 10
		},
	})

	var count int
	joinedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { count++ },
		DeleteFunc: func(obj any) { count-- },
	})

	left.OnAdd(Item{ID: 1, Val: 15}, true)
	right.OnAdd(Item{ID: 1, Val: 5}, true) // Fails Where

	if count != 0 {
		t.Errorf("Expected 0, got %d", count)
	}

	right.OnUpdate(Item{ID: 1, Val: 5}, Item{ID: 1, Val: 20}) // Now matches Where
	if count != 1 {
		t.Errorf("Expected 1 after right update, got %d", count)
	}
}
