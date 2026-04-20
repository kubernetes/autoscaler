package fort

import (
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	fcache "k8s.io/client-go/tools/cache/testing"
)

func TestIntegration_RealInformerToQuery(t *testing.T) {
	lw := fcache.NewFakeControllerSource()
	inf := cache.NewSharedInformer(lw, &corev1.Pod{}, 0)
	
	lock := NewLockGroup()
	query := QueryInformer(&Select[string, *corev1.Pod]{
		Lock: lock,
		From: inf,
		Select: func(p *corev1.Pod) (string, error) { return p.Name, nil },
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go inf.Run(stopCh)
	go query.Run(stopCh)

	// Wait for sync
	cache.WaitForCacheSync(stopCh, inf.HasSynced)

	// Live update
	lw.Add(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-live"}})

	// Verify propagation
	found := false
	for i := 0; i < 50; i++ {
		list := query.GetStore().List()
		if len(list) == 1 && list[0].(string) == "pod-live" {
			found = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !found {
		t.Errorf("Query failed to receive live update from real informer")
	}
}

func TestIntegration_QueryToRealInformer(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)
	
	// Create a Fort query
	query := QueryInformer(&Select[int, int]{
		Lock: lock,
		From: source,
		Select: func(i int) (int, error) { return i * 10, nil },
	})

	// Verify compatibility with standard ResourceEventHandler
	var received int
	var mu sync.Mutex
	query.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			mu.Lock()
			defer mu.Unlock()
			received = obj.(int)
		},
	})

	source.OnAdd(5, true)
	
	if received != 50 {
		t.Errorf("Standard event handler failed to receive event from query informer")
	}
}

func TestIntegration_RealInformerHydration(t *testing.T) {
	lw := fcache.NewFakeControllerSource()
	inf := cache.NewSharedInformer(lw, &corev1.Pod{}, 0)
	
	// Add data BEFORE starting
	lw.Add(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}})
	
	lock := NewLockGroup()
	query := QueryInformer(&Select[string, *corev1.Pod]{
		Lock: lock,
		From: inf,
		Select: func(p *corev1.Pod) (string, error) { return p.Name, nil },
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go inf.Run(stopCh)
	go query.Run(stopCh)

	// Verify query is born hydrated
	found := false
	for i := 0; i < 50; i++ {
		if query.HasSynced() && len(query.GetStore().List()) == 1 {
			found = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !found {
		t.Errorf("Query failed to hydrate from real informer")
	}
}
