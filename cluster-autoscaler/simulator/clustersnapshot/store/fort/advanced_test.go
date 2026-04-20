package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestClone_DiamondDAG(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)

	// Diamond Structure:
	//      Source
	//      /    \
	//   Left    Right (both FlatMap over Source)
	//      \    /
	//       Join  (Joins Left and Right)

	left := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map:  func(i int) ([]int, error) { return []int{i}, nil },
		Over: source,
	})
	right := QueryInformer(&FlatMap[int, int]{
		Lock: lock,
		Map:  func(i int) ([]int, error) { return []int{i}, nil },
		Over: source,
	})

	joined := QueryInformer(&Join[int, int, int]{
		Lock:   lock,
		Select: func(l, r int) (int, error) { return l + r, nil },
		From:   left,
		Join:   right,
		On:     func(l, r int) any { return l },
	})

	// Clone the diamond
	ns := source.Clone(nil).(ManualSharedInformer)
	memo := map[cache.SharedInformer]cache.SharedInformer{
		source: ns,
	}

	nj := ClonePipeline(joined, memo).(CloneableSharedInformerQuery)

	// A Join query is a Select over a joiner.
	// nj (Select) -> joiner -> [nLeft, nRight]
	joinerInf := nj.GetSources()[0].(CloneableSharedInformerQuery)
	sources := joinerInf.GetSources()
	nLeft := sources[0].(CloneableSharedInformerQuery)
	nRight := sources[1].(CloneableSharedInformerQuery)

	if nLeft.GetSources()[0] != ns || nRight.GetSources()[0] != ns {
		t.Errorf("Cloned diamond stages should all point to the same cloned source")
	}
}

func TestGrouper_MultiAggregation(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, DefaultKeyFunc)

	type Aggs struct {
		Count    int64
		Sum      int64
		Any      int
		Distinct []int
	}

	grouper := QueryInformer(&GroupBy[*Aggs, int]{
		Lock: lock,
		Select: func(fields []GroupField) (*Aggs, error) {
			rawDistincts := fields[3].([]any)
			distincts := make([]int, len(rawDistincts))
			for i, d := range rawDistincts {
				distincts[i] = d.(int)
			}
			return &Aggs{
				Count:    fields[0].(int64),
				Sum:      fields[1].(int64),
				Any:      fields[2].(int),
				Distinct: distincts,
			}, nil
		},
		From: source,
		GroupBy: func(i int) (any, []GroupField) {
			return 0, []GroupField{
				Count(),
				Sum(int64(i)),
				AnyValue(i),
				Distinct(i),
			}
		},
	})

	var latest *Aggs
	grouper.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { latest = obj.(*Aggs) },
		UpdateFunc: func(old, new any) { latest = new.(*Aggs) },
	})

	source.OnAdd(10, true)
	source.OnAdd(20, false)
	source.OnAdd(10, false)

	if latest.Count != 3 {
		t.Errorf("Expected count 3, got %d", latest.Count)
	}
	if latest.Sum != 40 {
		t.Errorf("Expected sum 40, got %d", latest.Sum)
	}
	if len(latest.Distinct) != 2 {
		t.Errorf("Expected 2 distinct values, got %v", latest.Distinct)
	}

	source.OnDelete(10)
	if latest.Count != 2 || latest.Sum != 30 {
		t.Errorf("After delete: expected count 2 sum 30, got count %d sum %d", latest.Count, latest.Sum)
	}
}

func TestDefaultKeyFunc_Robustness(t *testing.T) {
	cases := []struct {
		name string
		obj  any
	}{
		{"nil", nil},
		{"string", "hello"},
		{"int", 123},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}},
		{"struct", struct{ A int }{42}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := DefaultKeyFunc(tc.obj)
			if err != nil {
				t.Errorf("DefaultKeyFunc failed for %s: %v", tc.name, err)
			}
			if key == "" {
				t.Errorf("DefaultKeyFunc returned empty key for %s", tc.name)
			}
		})
	}
}

func TestBTreeMap_CloneIsolation(t *testing.T) {
	m := NewBTreeMap[[]int]()
	m.Set("k1", []int{1})

	m2 := m.Clone()
	
	// Modify original slice (this is why we need COW in the handlers)
	v, _ := m.Get("k1")
	v[0] = 99 

	v2, _ := m2.Get("k1")
	if v2[0] != 99 {
		t.Errorf("BTreeMap clone somehow isolated a shallow mutation without COW?")
	}
}

func TestBTreeMap_COWManualIsolation(t *testing.T) {
	m := NewBTreeMap[[]int]()
	m.Set("k1", []int{1})

	m2 := m.Clone()
	
	// Proper COW update
	v, _ := m.Get("k1")
	newV := append([]int(nil), v...)
	newV[0] = 99
	m.Set("k1", newV)

	v1, _ := m.Get("k1")
	v2, _ := m2.Get("k1")

	if v1[0] != 99 {
		t.Errorf("Original not updated")
	}
	if v2[0] != 1 {
		t.Errorf("Clone not isolated")
	}
}
