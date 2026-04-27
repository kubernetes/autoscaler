package fort

import (
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

func TestClone_ConsistencyUnderUpdate(t *testing.T) {
	// Root source uses a keyFunc that handles integers.
	keyFunc := func(obj any) (string, error) {
		return string(rune(obj.(int))), nil
	}
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, keyFunc)
	
	// Pipeline: source -> flatMap -> grouper
	fm := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map: func(i int) ([]int, error) {
			res := make([]int, 10)
			for j := 0; j < 10; j++ {
				res[j] = i*1000 + j
			}
			return res, nil
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

	// Start a goroutine that constantly updates the source
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		val := 1
		source.OnAdd(val, true)
		for {
			select {
			case <-stop:
				return
			default:
				newVal := val + 1
				source.OnUpdate(val, newVal)
				val = newVal
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Give the pipeline a moment to warm up before first clone
	time.Sleep(100 * time.Millisecond)

	// Perform multiple clones while updates are happening
	for i := 0; i < 10; i++ {
		// 1. Lock original pipeline
		locks := SnapshotLockDomain(grouper)
		
		// 2. Clone the chain
		// IMPORTANT: For the clone to have state, we must connect it to the snapshot sources.
		ns := source.Clone(nil)
		nfm := fm.Clone([]cache.SharedInformer{ns})
		ncloned := grouper.Clone([]cache.SharedInformer{nfm})
		
		// 3. Unlock original (let updates continue)
		locks.Unlock()

		// 4. Test cloned pipeline's integrity independently
		var results []int64
		var resLock sync.Mutex
		ncloned.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) { 
				resLock.Lock()
				results = append(results, obj.(int64)) 
				resLock.Unlock()
			},
			UpdateFunc: func(old, new any) { 
				resLock.Lock()
				results = append(results, new.(int64)) 
				resLock.Unlock()
			},
		})
		
		// The clone should be immediately populated because AddEventHandler 
		// replays the state of the (snapshot) indexer.
		success := false
		for attempt := 0; attempt < 100; attempt++ {
			resLock.Lock()
			if len(results) > 0 && results[len(results)-1] > 0 {
				success = true
				resLock.Unlock()
				break
			}
			resLock.Unlock()
			time.Sleep(1 * time.Millisecond)
		}

		if !success {
			t.Errorf("Iteration %d: Cloned pipeline failed to populate state. Got %v", i, results)
		}
	}

	close(stop)
	wg.Wait()
}

func TestClone_AtomicSnapshot(t *testing.T) {
	source := NewManualSharedInformer()
	query := QueryInformer(&Select[int, int]{
		Select: func(i int) (int, error) { return i, nil },
		From:   source,
	})

	source.OnAdd(1, true)

	locks := SnapshotLockDomain(query)
	
	// Start an update in background
	updated := make(chan struct{})
	go func() {
		source.OnUpdate(1, 2)
		close(updated)
	}()

	// The update should be blocked by the lock
	select {
	case <-updated:
		t.Errorf("Update should have been blocked by SnapshotLockDomain")
	case <-time.After(50 * time.Millisecond):
		// Success: update is blocked
	}

	locks.Unlock()
	<-updated // Now it should finish
}
