package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestClone_FlatMap(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	query := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map:  func(i int) ([]int, error) { return []int{i * 10}, nil },
		Over: source,
	})

	newSource := source.Clone(nil).(ManualSharedInformer)
	cloned := query.Clone([]cache.SharedInformer{newSource})

	var result int
	cloned.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { result = obj.(int) },
	})

	newSource.OnAdd(5, true)
	if result != 50 {
		t.Errorf("Cloned FlatMap failed: expected 50, got %d", result)
	}
}

func TestClone_Join(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	s2 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	onFunc := func(l, r int) any { return [1]int{0} }
	query := QueryInformer(&Join[int, int, int]{
		Lock: lock,
		Select: func(l, r int) (int, error) { return l + r, nil },
		From:   s1,
		Join:   s2,
		On:     onFunc,
	})

	ns1 := s1.Clone(nil).(ManualSharedInformer)
	ns2 := s2.Clone(nil).(ManualSharedInformer)
	
	nj := newJoiner[int, int](ns1.GetLockGroup(), ns1, ns2, onFunc)
	cloned := query.Clone([]cache.SharedInformer{nj})

	var result int
	cloned.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { result = obj.(int) },
	})

	ns1.OnAdd(10, true)
	ns2.OnAdd(20, true)
	if result != 30 {
		t.Errorf("Cloned Join failed: expected 30, got %d", result)
	}
}

func TestClone_GroupBy(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	query := QueryInformer(&GroupBy[int, int]{
		Lock:    lock,
		Select:  func(fields []GroupField) (int, error) { return int(fields[0].(int64)), nil },
		From:    source,
		GroupBy: func(i int) (any, []GroupField) { return [1]int{0}, []GroupField{Count()} },
	})

	newSource := source.Clone(nil).(ManualSharedInformer)
	cloned := query.Clone([]cache.SharedInformer{newSource})

	var result int
	cloned.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { result = obj.(int) },
		UpdateFunc: func(old, new any) { result = new.(int) },
	})

	newSource.OnAdd(1, true)
	newSource.OnAdd(2, false)
	if result != 2 {
		t.Errorf("Cloned GroupBy failed: expected 2, got %d", result)
	}
}

func TestClone_Chained(t *testing.T) {
	lock := NewLockGroup()
	s1 := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	
	q1 := QueryInformer(&Select[int, int]{
		Lock:   lock,
		Select: func(i int) (int, error) { return i + 1, nil },
		From:   s1,
	})
	q2 := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map:  func(i int) ([]int, error) { return []int{i, i * 2}, nil },
		Over: q1,
	})

	ns1 := s1.Clone(nil).(ManualSharedInformer)
	nq1 := q1.Clone([]cache.SharedInformer{ns1})
	nq2 := q2.Clone([]cache.SharedInformer{nq1})

	var results []int
	nq2.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { results = append(results, obj.(int)) },
	})

	ns1.OnAdd(10, true) 
	if len(results) != 2 || results[0] != 11 || results[1] != 22 {
		t.Errorf("Cloned chain failed: %v", results)
	}
}
