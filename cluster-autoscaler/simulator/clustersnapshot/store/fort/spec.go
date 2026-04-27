package fort

import "k8s.io/client-go/tools/cache"

func getLock(lock LockGroup, source cache.SharedInformer) LockGroup {
	if lock != nil {
		return lock
	}
	if q, ok := source.(CloneableSharedInformerQuery); ok {
		return q.GetLockGroup()
	}
	return NewLockGroup()
}

func getSourceKeyFunc(source cache.SharedInformer) cache.KeyFunc {
	if q, ok := source.(CloneableSharedInformerQuery); ok {
		return q.GetKeyFunc()
	}
	return DefaultKeyFunc
}

func (q *Select[Out, In]) Build() CloneableSharedInformerQuery {
	lock := getLock(q.Lock, q.From)
	inf := newFlatMapper[Out, In](lock, func(value In) ([]Out, error) {
		if q.Where == nil || q.Where(value) {
			out, err := q.Select(value)
			if err != nil {
				return nil, err
			}
			return []Out{out}, nil
		}
		return nil, nil
	}, q.From)
	inf.SetName("select-query")
	return inf
}

func (q *Join[Out, Left, Right]) Build() CloneableSharedInformerQuery {
	lock := getLock(q.Lock, q.From)
	on := q.On
	if on == nil {
		on = func(l Left, r Right) any { return 0 }
	}

	kl := getSourceKeyFunc(q.From)
	kr := getSourceKeyFunc(q.Join)

	joinKeyFunc := func(obj any) (string, error) {
		jv := obj.(JoinValue[Left, Right])
		l, _ := kl(jv.Left)
		r, _ := kr(jv.Right)
		return l + " // " + r, nil
	}

	sq := &Select[Out, JoinValue[Left, Right]]{
		Lock: lock,
		Select: func(joined JoinValue[Left, Right]) (Out, error) {
			return q.Select(joined.Left, joined.Right)
		},
		From: newJoinerWithHandler(q.From, q.Join, on, NewManualSharedInformerWithOptions(lock, joinKeyFunc)),
		Where: func(joined JoinValue[Left, Right]) bool {
			if q.Where == nil {
				return true
			}
			return q.Where(joined.Left, joined.Right)
		},
	}
	inf := sq.Build()
	inf.SetName("join-query")
	return inf
}

func (q *GroupBy[Out, In]) Build() CloneableSharedInformerQuery {
	lock := getLock(q.Lock, q.From)
	inf := newGrouper[Out, In](lock, q.Select, q.GroupBy, q.From, q.Where)
	inf.SetName("groupBy-query")
	return inf
}

func (q *GroupByJoin[Out, Left, Right]) Build() CloneableSharedInformerQuery {
	lock := getLock(q.Lock, q.From)
	on := q.On
	if on == nil {
		on = func(l Left, r Right) any { return 0 }
	}

	kl := getSourceKeyFunc(q.From)
	kr := getSourceKeyFunc(q.Join)

	joinKeyFunc := func(obj any) (string, error) {
		jv := obj.(JoinValue[Left, Right])
		l, _ := kl(jv.Left)
		r, _ := kr(jv.Right)
		return l + " // " + r, nil
	}

	g := &GroupBy[Out, JoinValue[Left, Right]]{
		Lock:   lock,
		Select: q.Select,
		From:   newJoinerWithHandler(q.From, q.Join, on, NewManualSharedInformerWithOptions(lock, joinKeyFunc)),
		Where: func(joined JoinValue[Left, Right]) bool {
			if q.Where == nil {
				return true
			}
			return q.Where(joined.Left, joined.Right)
		},
		GroupBy: func(joined JoinValue[Left, Right]) (any, []GroupField) {
			return q.GroupBy(joined.Left, joined.Right)
		},
	}
	inf := g.Build()
	inf.SetName("groupByJoin-query")
	return inf
}

func (q *FlatMap[Out, In]) Build() CloneableSharedInformerQuery {
	lock := getLock(q.Lock, q.Over)
	inf := newFlatMapper[Out, In](lock, q.Map, q.Over)
	inf.SetName("flatMap-query")
	return inf
}
