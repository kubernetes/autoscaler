package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestGrouper_UpdateAndDelete(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type Data struct {
		ID       int
		Category string
		Value    int64
	}
	type Aggregated struct {
		Category string
		Sum      int64
	}

	grouped := QueryInformer(&GroupBy[Aggregated, Data]{
		Lock: lock,
		Select: func(fields []GroupField) (Aggregated, error) {
			return Aggregated{
				Category: fields[0].(string),
				Sum:      fields[1].(int64),
			}, nil
		},
		From: source,
		GroupBy: func(d Data) (any, []GroupField) {
			return [1]string{d.Category},
				[]GroupField{
					AnyValue(d.Category),
					Sum(d.Value),
				}
		},
	})

	results := make(map[string]int64)
	grouped.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			res := obj.(Aggregated)
			results[res.Category] = res.Sum
		},
		UpdateFunc: func(oldObj, newObj any) {
			res := newObj.(Aggregated)
			results[res.Category] = res.Sum
		},
		DeleteFunc: func(obj any) {
			res := obj.(Aggregated)
			delete(results, res.Category)
		},
	})

	d1 := Data{ID: 1, Category: "A", Value: 10}
	d2 := Data{ID: 2, Category: "A", Value: 20}
	
	source.OnAdd(d1, true)
	if results["A"] != 10 {
		t.Errorf("Expected 10, got %d", results["A"])
	}

	source.OnAdd(d2, false)
	if results["A"] != 30 {
		t.Errorf("Expected 30, got %d", results["A"])
	}

	// Update d1
	d1Updated := Data{ID: 1, Category: "A", Value: 15}
	source.OnUpdate(d1, d1Updated)
	if results["A"] != 35 {
		t.Errorf("Expected 35 after update, got %d", results["A"])
	}

	// Change category of d2
	d2Updated := Data{ID: 2, Category: "B", Value: 20}
	source.OnUpdate(d2, d2Updated)
	if results["A"] != 15 {
		t.Errorf("Expected 15 for A, got %d", results["A"])
	}
	if results["B"] != 20 {
		t.Errorf("Expected 20 for B, got %d", results["B"])
	}

	// Delete d1
	source.OnDelete(d1Updated)
	if _, ok := results["A"]; ok {
		t.Errorf("Category A should have been deleted")
	}
}

func TestGrouper_Where(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type Data struct {
		Val int
	}

	grouped := QueryInformer(&GroupBy[int, Data]{
		Lock: lock,
		Select: func(fields []GroupField) (int, error) {
			return int(fields[0].(int64)), nil
		},
		From: source,
		Where: func(d Data) bool {
			return d.Val > 10
		},
		GroupBy: func(d Data) (any, []GroupField) {
			return [1]string{"constant"}, []GroupField{Count()}
		},
	})

	var count int
	grouped.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { count = obj.(int) },
		UpdateFunc: func(old, new any) { count = new.(int) },
		DeleteFunc: func(obj any) { count = 0 },
	})

	source.OnAdd(Data{Val: 5}, true)
	if count != 0 {
		t.Errorf("Expected 0, got %d", count)
	}

	source.OnAdd(Data{Val: 15}, false)
	if count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}
}
