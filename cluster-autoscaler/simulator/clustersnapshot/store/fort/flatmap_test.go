package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestFlatMap_UpdateAndDelete(t *testing.T) {
	keyFunc := func(obj any) (string, error) {
		return obj.(string), nil
	}
	lock := NewLockGroup()
	handler := NewManualSharedInformerWithOptions(lock, keyFunc)
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type Item struct {
		ID   int
		Vals []string
	}

	m := &FlatMap[string, Item]{
		Lock: lock,
		Map: func(item Item) ([]string, error) {
			return item.Vals, nil
		},
		Over: source,
	}
	flatMap := newFlatMapperWithHandler[string, Item](m.Map, m.Over, handler)

	results := make(map[string]int)
	flatMap.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			results[obj.(string)]++
		},
		UpdateFunc: func(old, new any) {
			// If it's the same key, we don't need to do much in this test's result map
			// since it's just a count.
		},
		DeleteFunc: func(obj any) {
			results[obj.(string)]--
			if results[obj.(string)] == 0 {
				delete(results, obj.(string))
			}
		},
	})

	i1 := Item{ID: 1, Vals: []string{"a", "b"}}
	source.OnAdd(i1, true)

	if len(results) != 2 || results["a"] != 1 || results["b"] != 1 {
		t.Errorf("Initial flatmap failed: %v", results)
	}

	// Update i1: remove "a", add "c"
	i1Updated := Item{ID: 1, Vals: []string{"b", "c"}}
	source.OnUpdate(i1, i1Updated)

	if len(results) != 2 || results["b"] != 1 || results["c"] != 1 {
		t.Errorf("Update flatmap failed: %v", results)
	}
	if _, ok := results["a"]; ok {
		t.Errorf("Expected 'a' to be deleted")
	}

	// Delete i1
	source.OnDelete(i1Updated)
	if len(results) != 0 {
		t.Errorf("Expected 0 results after delete, got %v", results)
	}
}
