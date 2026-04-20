package fort

import (
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

func TestManualInformer_UpdateKeyChangeBug(t *testing.T) {
	lock := NewLockGroup()
	// Use a keyfunc that depends on the value
	keyFunc := func(obj any) (string, error) {
		return obj.(string), nil
	}
	inf := NewManualSharedInformerWithOptions(lock, keyFunc)

	inf.OnAdd("A", true)
	if len(inf.GetStore().List()) != 1 {
		t.Errorf("Expected 1 item")
	}

	// Update "A" to "B". Key changes from "A" to "B".
	inf.OnUpdate("A", "B")

	items := inf.GetStore().List()
	if len(items) != 1 {
		t.Errorf("Expected 1 item after update, but got %d (leak of old key!)", len(items))
		for _, item := range items {
			t.Logf("Item: %v", item)
		}
	}
}

func TestClonePipeline_CyclePanic(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	q1 := &Select[int, int]{Lock: lock, From: source, Select: func(i int) (int, error) { return i, nil }}
	i1 := q1.Build()
	
	// Create a cycle by forcing sources
	i1.(*flatMapper[int, int]).source = i1 

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on recursive clone cycle")
		}
	}()

	memo := make(map[cache.SharedInformer]cache.SharedInformer)
	ClonePipeline(i1, memo)
}

func TestOperator_RegistrationLeak(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)

	// Build a small pipeline
	query := QueryInformer(&Select[int, int]{
		Lock:   lock,
		From:   source,
		Select: func(i int) (int, error) { return i, nil },
	})

	// Get internal informer to check handlers count
	si := source.(*manualInformer)

	si.lock.RLock()
	initialHandlers := len(si.handlers)
	si.lock.RUnlock()
	
	for i := 0; i < 100; i++ {
		// Clone
		cl := query.Clone([]cache.SharedInformer{source})
		
		// Run and stop immediately to unregister
		stopCh := make(chan struct{})
		go cl.Run(stopCh)
		close(stopCh)
		
		// Give a tiny bit of time for the goroutine to cleanup
		time.Sleep(1 * time.Millisecond)
	}

	si.lock.RLock()
	finalHandlers := len(si.handlers)
	si.lock.RUnlock()
	if finalHandlers > initialHandlers {
		t.Errorf("Registration leak detected! Handlers count grew from %d to %d", initialHandlers, finalHandlers)
	}
}


