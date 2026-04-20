package fort

import (
	"fmt"
	"testing"

	"k8s.io/client-go/tools/cache"
)

type perfItem struct {
	ID    int
	Value int
}

func BenchmarkThroughput(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	depths := []int{1, 3, 5}

	for _, depth := range depths {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("Depth%d/Size%d", depth, size), func(b *testing.B) {
				lock := NewLockGroup()
				source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
				
				var current CloneableSharedInformerQuery = source
				for i := 0; i < depth; i++ {
					// Add a transformation layer
					current = QueryInformer(&Select[perfItem, perfItem]{
						Lock: lock,
						Select: func(in perfItem) (perfItem, error) {
							return perfItem{ID: in.ID, Value: in.Value + 1}, nil
						},
						From: current,
					})
				}

				// Pre-hydrate
				for i := 0; i < size; i++ {
					source.OnAdd(perfItem{ID: i, Value: 0}, true)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					id := i % size
					source.OnUpdate(perfItem{ID: id, Value: i}, perfItem{ID: id, Value: i + 1})
				}
			})
		}
	}
}

func BenchmarkCloningPerformance(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			lock := NewLockGroup()
			source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
			
			// Complex-ish pipeline: source -> select -> grouper
			fm := QueryInformer(&FlatMap[int, perfItem]{
				Lock: lock,
				Map: func(in perfItem) ([]int, error) {
					return []int{in.Value, in.Value * 2}, nil
				},
				Over: source,
			})

			grouper := QueryInformer(&GroupBy[int64, int]{
				Lock: lock,
				Select: func(fields []GroupField) (int64, error) {
					return fields[0].(int64), nil
				},
				From: fm,
				GroupBy: func(i int) (any, []GroupField) {
					return 0, []GroupField{Sum(int64(i))}
				},
			})

			// Hydrate
			for i := 0; i < size; i++ {
				source.OnAdd(perfItem{ID: i, Value: i}, true)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// We must hold the lock to clone
				lock.RLock()
				ns := source.Clone(nil)
				nfm := fm.Clone([]cache.SharedInformer{ns})
				_ = grouper.Clone([]cache.SharedInformer{nfm})
				lock.RUnlock()
			}
		})
	}
}

func BenchmarkJoinPerformance(b *testing.B) {
	sizes := []int{100, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			lock := NewLockGroup()
			s1 := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
			s2 := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)

			joined := QueryInformer(&Join[int, perfItem, perfItem]{
				Lock: lock,
				Select: func(l, r perfItem) (int, error) {
					return l.Value + r.Value, nil
				},
				From: s1,
				Join: s2,
				On: func(l, r perfItem) any {
					if l.ID != 0 {
						return [1]int{l.ID}
					}
					return [1]int{r.ID}
				},
			})

			// To keep results alive
			joined.AddEventHandler(cache.ResourceEventHandlerFuncs{})

			// Hydrate both sides
			for i := 0; i < size; i++ {
				s1.OnAdd(perfItem{ID: i, Value: i}, true)
				s2.OnAdd(perfItem{ID: i, Value: i}, true)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				id := i % size
				// Update one side - triggers join recalculation
				s1.OnUpdate(perfItem{ID: id, Value: i}, perfItem{ID: id, Value: i + 1})
			}
		})
	}
}
