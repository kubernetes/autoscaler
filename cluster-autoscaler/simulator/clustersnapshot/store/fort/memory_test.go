package fort

import (
	"fmt"
	"runtime"
	"testing"

	"k8s.io/client-go/tools/cache"
)

func getMemory() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc
}

type MemItem struct {
	ID int
	Val int
}

func TestMemoryOverhead_SourceOnly(t *testing.T) {
	const numItems = 100_000
	
	before := getMemory()
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	for i := 0; i < numItems; i++ {
		source.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: Source only (%d items): %.2f MB (%.1f bytes/item)\n", 
		numItems, float64(overhead)/1024/1024, float64(overhead)/float64(numItems))
	
	runtime.KeepAlive(source)
}

func TestMemoryOverhead_Select(t *testing.T) {
	const numItems = 100_000
	
	before := getMemory()
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	sel := QueryInformer(&Select[*MemItem, *MemItem]{
		Lock: lock,
		Select: func(i *MemItem) (*MemItem, error) { 
			return &MemItem{ID: i.ID, Val: i.Val + 1}, nil 
		},
		From: source,
	})
	
	sel.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	for i := 0; i < numItems; i++ {
		source.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: Source + Select (%d items): %.2f MB (%.1f bytes/item)\n", 
		numItems, float64(overhead)/1024/1024, float64(overhead)/float64(numItems))
	
	runtime.KeepAlive(source)
	runtime.KeepAlive(sel)
}

func TestMemoryOverhead_FlatMap(t *testing.T) {
	const numItems = 100_000
	
	before := getMemory()
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	// 1:2 mapping
	fm := QueryInformer(&FlatMap[*MemItem, *MemItem]{
		Lock: lock,
		Map: func(i *MemItem) ([]*MemItem, error) { 
			return []*MemItem{i, {ID: i.ID + 1000000, Val: i.Val}}, nil 
		},
		Over: source,
	})
	
	fm.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	for i := 0; i < numItems; i++ {
		source.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: Source + FlatMap 1:2 (%d items): %.2f MB (%.1f bytes/item)\n", 
		numItems, float64(overhead)/1024/1024, float64(overhead)/float64(numItems))
	
	runtime.KeepAlive(source)
	runtime.KeepAlive(fm)
}

func TestMemoryOverhead_Join(t *testing.T) {
	const numItems = 100_000
	
	before := getMemory()
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	s2 := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	join := QueryInformer(&Join[*MemItem, *MemItem, *MemItem]{
		Lock: lock,
		Select: func(l, r *MemItem) (*MemItem, error) { 
			return &MemItem{ID: l.ID, Val: l.Val + r.Val}, nil 
		},
		From: s1,
		Join: s2,
		On: func(l, r *MemItem) any {
			if l != nil { return l.ID }
			return r.ID
		},
	})
	
	join.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	for i := 0; i < numItems; i++ {
		s1.OnAdd(&MemItem{ID: i, Val: i}, true)
		s2.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: 2 Sources + Join (%d items each): %.2f MB (%.1f bytes/joined-result)\n", 
		numItems, float64(overhead)/1024/1024, float64(overhead)/float64(numItems))
	
	runtime.KeepAlive(s1)
	runtime.KeepAlive(s2)
	runtime.KeepAlive(join)
}

func TestMemoryOverhead_GroupBy(t *testing.T) {
	const numItems = 100_000
	const numGroups = 1000
	
	before := getMemory()
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	grouper := QueryInformer(&GroupBy[int64, *MemItem]{
		Lock: lock,
		Select: func(fields []GroupField) (int64, error) { return fields[0].(int64), nil },
		From: source,
		GroupBy: func(i *MemItem) (any, []GroupField) {
			return i.ID % numGroups, []GroupField{Count()}
		},
	})
	
	grouper.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	for i := 0; i < numItems; i++ {
		source.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: Source + GroupBy (%d items, %d groups): %.2f MB (%.1f bytes/item)\n", 
		numItems, numGroups, float64(overhead)/1024/1024, float64(overhead)/float64(numItems))
	
	runtime.KeepAlive(source)
	runtime.KeepAlive(grouper)
}

func TestMemoryOverhead_Clones(t *testing.T) {
	const numItems = 10_000
	const numClones = 1000
	
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	for i := 0; i < numItems; i++ {
		source.OnAdd(&MemItem{ID: i, Val: i}, true)
	}
	
	before := getMemory()
	clones := make([]cache.SharedInformer, numClones)
	for i := 0; i < numClones; i++ {
		lock.RLock()
		clones[i] = source.Clone(nil)
		lock.RUnlock()
	}
	
	after := getMemory()
	overhead := int64(after) - int64(before)
	if overhead < 0 { overhead = 0 }
	fmt.Printf("Memory: %d Clones of 10k items: %.2f MB (%.1f bytes/clone)\n", 
		numClones, float64(overhead)/1024/1024, float64(overhead)/float64(numClones))
	
	// Prevent GC from collecting clones before measurement
	runtime.KeepAlive(clones)
	runtime.KeepAlive(source)
}
