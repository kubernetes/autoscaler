package fort

import (
	"context"
	"errors"
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestWatchErrorHandlerPropagation_Chained(t *testing.T) {
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

	var observedErr error
	errHandler := func(r *cache.Reflector, err error) {
		observedErr = err
	}

	grouper.SetWatchErrorHandler(errHandler)

	testErr := errors.New("upstream watch failure")
	s1.TriggerWatchError(testErr)

	if observedErr != testErr {
		t.Errorf("Expected error %v, got %v", testErr, observedErr)
	}
}

func TestWatchErrorHandlerPropagation_Join(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	s2 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	joined := QueryInformer(&Join[int, int, int]{
		Lock: lock,
		Select: func(l, r int) (int, error) { return l + r, nil },
		From:   s1,
		Join:   s2,
		On: func(l, r int) any { return [1]int{0} },
	})

	var observedErr error
	errHandler := func(r *cache.Reflector, err error) {
		observedErr = err
	}

	joined.SetWatchErrorHandler(errHandler)

	testErr1 := errors.New("left source failure")
	s1.TriggerWatchError(testErr1)
	if observedErr != testErr1 {
		t.Errorf("Expected error from left source, got %v", observedErr)
	}

	testErr2 := errors.New("right source failure")
	s2.TriggerWatchError(testErr2)
	if observedErr != testErr2 {
		t.Errorf("Expected error from right source, got %v", observedErr)
	}
}

func TestWatchErrorHandlerWithContextPropagation(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	
	flatMap := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map: func(i int) ([]int, error) { return []int{i}, nil },
		Over: s1,
	})

	var observedErr error
	errHandler := func(ctx context.Context, r *cache.Reflector, err error) {
		observedErr = err
	}

	flatMap.SetWatchErrorHandlerWithContext(errHandler)

	testErr := errors.New("contextual error")
	s1.TriggerWatchError(testErr)

	if observedErr != testErr {
		t.Errorf("Expected contextual error %v, got %v", testErr, observedErr)
	}
}
