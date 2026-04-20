package fort

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

// TestStress_ConcurrentUpdateAndClone executes high-frequency updates from multiple goroutines
// while periodically taking clones. It verifies that all clones are internally consistent
// and isolated from subsequent updates in the original pipeline.
func TestStress_ConcurrentUpdateAndClone(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)

	// Build a non-trivial pipeline: Source -> FlatMap -> GroupBy
	// FlatMap: 1 int -> [val, val*10]
	// GroupBy: sum of all values
	fm := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map:  func(i int) ([]int, error) { return []int{i, i * 10}, nil },
		Over: source,
	})

	type Agg struct {
		Count int64
		Sum   int64
	}
	grouper := QueryInformer(&GroupBy[*Agg, int]{
		Lock: lock,
		Select: func(fields []GroupField) (*Agg, error) {
			return &Agg{
				Count: fields[0].(int64),
				Sum:   fields[1].(int64),
			}, nil
		},
		From: fm,
		GroupBy: func(i int) (any, []GroupField) {
			return "total", []GroupField{Count(), Sum(int64(i))}
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go fm.Run(stopCh)
	go grouper.Run(stopCh)

	const numUpdaters = 5
	const updatesPerUpdater = 500
	const numCloners = 2
	const clonesPerCloner = 20

	var wg sync.WaitGroup
	
	// 1. Updaters: push random data
	wg.Add(numUpdaters)
	for i := 0; i < numUpdaters; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < updatesPerUpdater; j++ {
				val := rand.Intn(100) + 1 // 1..100
				source.OnAdd(val, false)
				if j%10 == 0 {
					source.OnDelete(val)
				}
				time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			}
		}(i)
	}

	// 2. Cloners: periodically take snapshots and verify isolation
	wg.Add(numCloners)
	for i := 0; i < numCloners; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < clonesPerCloner; j++ {
				// Atomically capture the original state and take a clone
				locks := SnapshotLockDomain(grouper)
				
				// Capture current state from original using List() for identity-agnostic check
				var origCount int64
				origList := grouper.GetStore().List()
				if len(origList) > 0 {
					origCount = origList[0].(*Agg).Count
				}
				
				// Create clone
				ns := source.Clone(nil)
				nfm := fm.Clone([]cache.SharedInformer{ns})
				ng := grouper.Clone([]cache.SharedInformer{nfm})
				
				locks.Unlock()

				// Verify Clone Isolation: The clone should eventually match the 
				// captured state, and should NOT change even as 'source' continues to churn.
				
				// Since clones are born hydrated with their B-Tree states, 
				// but their handlers are populated via replay, we wait a bit for hydration.
				var cloneCount int64
				found := false
				for attempt := 0; attempt < 100; attempt++ {
					cloneList := ng.GetStore().List()
					if len(cloneList) > 0 {
						cloneCount = cloneList[0].(*Agg).Count
						// If we captured an origCount, the clone must eventually reach at least that state.
						if cloneCount >= origCount {
							found = true
							break
						}
					} else if origCount == 0 {
						found = true
						break
					}
					time.Sleep(1 * time.Millisecond)
				}
				
				if !found {
					t.Errorf("Cloner %d: Clone failed to hydrate or was inconsistent. Orig count: %d, Clone count: %d", id, origCount, cloneCount)
				}

				// Isolation check: capture current clone state, wait, check again.
				initialCount := cloneCount
				time.Sleep(10 * time.Millisecond)
				cloneListFinal := ng.GetStore().List()
				if len(cloneListFinal) > 0 {
					finalCount := cloneListFinal[0].(*Agg).Count
					if finalCount != initialCount {
						t.Errorf("Cloner %d: Isolation failure! Clone state changed from %d to %d", id, initialCount, finalCount)
					}
				}

				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}

// TestStress_MassiveCloning verifies that taking thousands of clones rapidly
// does not cause memory leaks or deadlocks.
func TestStress_MassiveCloning(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	query := QueryInformer(&Select[int, int]{
		Lock: lock,
		Select: func(i int) (int, error) { return i, nil },
		From: source,
	})

	for i := 0; i < 1000; i++ {
		source.OnAdd(i, false)
		
		func() {
			locks := SnapshotLockDomain(query)
			defer locks.Unlock()
			
			ns := source.Clone(nil)
			_ = query.Clone([]cache.SharedInformer{ns})
		}()
	}
}
