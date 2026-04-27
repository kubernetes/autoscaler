package fort

import (
	"fmt"
	"time"

	"k8s.io/client-go/tools/cache"
)

// grouper implements GroupBy query. It aggregates objects from a source informer
// into groups identified by a comparable key and evaluates aggregate fields.
// It embeds baseInformer to handle event routing and state propagation.
type grouper[Out, In any] struct {
	*baseInformer
	sel     GroupSelectFunc[Out]
	groupBy SingleGroupByFunc[In]
	where   SingleFilterFunc[In]
	source  cache.SharedInformer

	registration cache.ResourceEventHandlerRegistration

	groups BTreeMap[*groupState[Out]]
}

var _ ManualSharedInformer = &grouper[int, int]{}

// groupState maintains the current aggregation state for a single group.
type groupState[Out any] struct {
	count   int
	fields  []any
	lastOut Out
	anyValues map[int]map[any]int
}

func (s *groupState[Out]) clone() *groupState[Out] {
	ns := *s
	ns.fields = append([]any(nil), s.fields...)
	if s.anyValues != nil {
		ns.anyValues = make(map[int]map[any]int, len(s.anyValues))
		for k, v := range s.anyValues {
			inner := make(map[any]int, len(v))
			for ik, iv := range v {
				inner[ik] = iv
			}
			ns.anyValues[k] = inner
		}
	}
	return &ns
}

func newGrouper[Out, In any](lock LockGroup, sel GroupSelectFunc[Out], groupBy SingleGroupByFunc[In], from cache.SharedInformer, where SingleFilterFunc[In]) *grouper[Out, In] {
	return newGrouperWithHandler(sel, groupBy, from, where, NewManualSharedInformerWithOptions(lock, DefaultKeyFunc))
}

func newGrouperWithHandler[Out, In any](sel GroupSelectFunc[Out], groupBy SingleGroupByFunc[In], from cache.SharedInformer, where SingleFilterFunc[In], handler ManualSharedInformer) *grouper[Out, In] {
	g := &grouper[Out, In]{
		baseInformer: newBaseInformer(handler.GetLockGroup(), handler.GetKeyFunc()),
		sel:          sel,
		groupBy:      groupBy,
		where:        where,
		source:       from,
		groups:       NewBTreeMap[*groupState[Out]](),
	}
	g.baseInformer.parent = g

	g.registration, _ = from.AddEventHandler(g)

	// Sync propagation goroutine.
	go func() {
		for {
			if g.registration.HasSynced() {
				g.SetHasSynced()
				return
			}
			select {
			case <-time.After(10 * time.Millisecond):
			case <-g.IsStoppedChan():
				return
			}
		}
	}()

	return g
}

func (g *grouper[Out, In]) GetIndexer() cache.Indexer {
	return &grouperIndexer[Out]{groups: g.groups}
}

func (g *grouper[Out, In]) GetStore() cache.Store {
	return &grouperIndexer[Out]{groups: g.groups}
}

func (g *grouper[Out, In]) OnAddLocked(obj any, isInitial bool) {
	input := obj.(In)
	if g.where != nil && !g.where(input) {
		return
	}

	key, fields := g.groupBy(input)
	keyStr := DefaultKey(key)

	state, ok := g.groups.Get(keyStr)
	var oldOut Out
	var newState *groupState[Out]
	if ok {
		oldOut = state.lastOut
		newState = state.clone()
	} else {
		newState = &groupState[Out]{}
	}

	newFields := g.evaluateFields(fields, newState, input, true)
	newOut, _ := g.sel(newFields)
	newState.lastOut = newOut
	g.groups.Set(keyStr, newState)

	if ok {
		g.dispatchUpdate(oldOut, newOut)
	} else {
		g.dispatchAdd(newOut, isInitial)
	}
}

func (g *grouper[Out, In]) OnUpdateLocked(oldObj, newObj any) {
	oldInput := oldObj.(In)
	newInput := newObj.(In)

	oldPass := g.where == nil || g.where(oldInput)
	newPass := g.where == nil || g.where(newInput)

	if !oldPass && !newPass {
		return
	}
	if !oldPass && newPass {
		g.OnAddLocked(newObj, false)
		return
	}
	if oldPass && !newPass {
		g.OnDeleteLocked(oldObj)
		return
	}

	oldKey, oldFields := g.groupBy(oldInput)
	newKey, newFields := g.groupBy(newInput)

	if oldKey == newKey {
		keyStr := DefaultKey(oldKey)
		state, ok := g.groups.Get(keyStr)
		if !ok {
			g.OnAddLocked(newObj, false)
			return
		}

		oldOut := state.lastOut
		newState := state.clone()

		g.evaluateFields(oldFields, newState, oldInput, false)
		resFields := g.evaluateFields(newFields, newState, newInput, true)
		
		newOut, _ := g.sel(resFields)
		newState.lastOut = newOut
		g.groups.Set(keyStr, newState)
		g.dispatchUpdate(oldOut, newOut)
	} else {
		g.OnDeleteLocked(oldObj)
		g.OnAddLocked(newObj, false)
	}
}

func (g *grouper[Out, In]) OnDeleteLocked(obj any) {
	input := obj.(In)
	if g.where != nil && !g.where(input) {
		return
	}

	key, fields := g.groupBy(input)
	keyStr := DefaultKey(key)

	state, ok := g.groups.Get(keyStr)
	if !ok {
		return
	}

	oldOut := state.lastOut
	newState := state.clone()

	newFields := g.evaluateFields(fields, newState, input, false)

	if newState.count == 0 {
		g.groups.Delete(keyStr)
		g.dispatchDelete(oldOut)
	} else {
		newOut, _ := g.sel(newFields)
		newState.lastOut = newOut
		g.groups.Set(keyStr, newState)
		g.dispatchUpdate(oldOut, newOut)
	}
}

func (g *grouper[Out, In]) dispatchAdd(obj Out, isInitial bool) {
	for _, h := range g.handlers {
		g.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnAdd(obj, isInitial) }, 
			func(l LockedResourceEventHandler) { l.OnAddLocked(obj, isInitial) })
	}
}

func (g *grouper[Out, In]) dispatchUpdate(oldObj, newObj Out) {
	for _, h := range g.handlers {
		g.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnUpdate(oldObj, newObj) }, 
			func(l LockedResourceEventHandler) { l.OnUpdateLocked(oldObj, newObj) })
	}
}

func (g *grouper[Out, In]) dispatchDelete(obj Out) {
	for _, h := range g.handlers {
		g.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnDelete(obj) }, 
			func(l LockedResourceEventHandler) { l.OnDeleteLocked(obj) })
	}
}

func (g *grouper[Out, In]) evaluateFields(fields []GroupField, state *groupState[Out], input In, adding bool) []GroupField {
	if state.fields == nil {
		state.fields = make([]any, len(fields))
	}

	if adding {
		state.count++
	} else {
		state.count--
	}

	res := make([]GroupField, len(fields))
	for i, f := range fields {
		gf := f.(*groupField)
		if gf.count {
			res[i] = int64(state.count)
		} else if gf.sum != nil {
			if state.fields[i] == nil {
				state.fields[i] = int64(0)
			}
			val := *gf.sum
			if adding {
				state.fields[i] = state.fields[i].(int64) + val
			} else {
				state.fields[i] = state.fields[i].(int64) - val
			}
			res[i] = state.fields[i]
		} else if gf.distinct != nil {
			if state.fields[i] == nil {
				state.fields[i] = make(map[any]int)
			}
			m := state.fields[i].(map[any]int)
			newM := make(map[any]int, len(m))
			for k, v := range m {
				newM[k] = v
			}
			
			if adding {
				newM[gf.distinct]++
			} else {
				newM[gf.distinct]--
				if newM[gf.distinct] == 0 {
					delete(newM, gf.distinct)
				}
			}
			state.fields[i] = newM
			
			var distincts []any
			for k := range newM {
				distincts = append(distincts, k)
			}
			res[i] = distincts
		} else if gf.anyValue != nil {
			if state.anyValues == nil {
				state.anyValues = make(map[int]map[any]int)
			}
			m := state.anyValues[i]
			if m == nil {
				m = make(map[any]int)
				state.anyValues[i] = m
			}
			if adding {
				m[gf.anyValue]++
			} else {
				m[gf.anyValue]--
				if m[gf.anyValue] == 0 {
					delete(m, gf.anyValue)
				}
			}
			var picked any
			for k := range m {
				picked = k
				break
			}
			state.fields[i] = picked
			res[i] = picked
		} else if gf.key != nil {
			res[i] = gf.key
		}
	}
	return res
}

func (g *grouper[Out, In]) Clone(newSources []cache.SharedInformer) CloneableSharedInformerQuery {
	ns := newSources[0]
	newLock := ns.(CloneableSharedInformerQuery).GetLockGroup()
	
	ng := &grouper[Out, In]{
		baseInformer: g.baseInformer.clone(),
		sel:          g.sel,
		groupBy:      g.groupBy,
		where:        g.where,
		source:       ns,
		groups:       g.groups.Clone(),
	}
	ng.baseInformer.lock = newLock
	ng.baseInformer.parent = ng

	if ms, ok := ns.(ManualSharedInformer); ok {
		ng.registration, _ = ms.AddEventHandlerNoReplay(ng)
	} else {
		ng.registration, _ = ns.AddEventHandler(ng)
	}

	return ng
}

func (g *grouper[Out, In]) GetController() cache.Controller {
	return nil
}

func (g *grouper[Out, In]) Run(stopCh <-chan struct{}) {
	defer g.SetIsStopped()
	defer func() {
		if g.source != nil && g.registration != nil {
			_ = g.source.RemoveEventHandler(g.registration)
		}
	}()
	<-stopCh
}

func (g *grouper[Out, In]) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	return g.source.SetWatchErrorHandler(handler)
}

func (g *grouper[Out, In]) SetWatchErrorHandlerWithContext(handler cache.WatchErrorHandlerWithContext) error {
	return g.source.SetWatchErrorHandlerWithContext(handler)
}

func (g *grouper[Out, In]) GetSources() []cache.SharedInformer {
	return []cache.SharedInformer{g.source}
}

// grouperIndexer implements cache.Store by wrapping the grouper's internal 
// groups B-Tree. It avoids maintaining a redundant index of Out objects.
type grouperIndexer[Out any] struct {
	groups  BTreeMap[*groupState[Out]]
	keyFunc cache.KeyFunc
}

func (i *grouperIndexer[Out]) Clone() CloneableIndexer {
	return &grouperIndexer[Out]{
		groups:  i.groups.Clone(),
		keyFunc: i.keyFunc,
	}
}

func (i *grouperIndexer[Out]) Add(obj any) error    { return fmt.Errorf("not supported") }
func (i *grouperIndexer[Out]) Update(obj any) error { return fmt.Errorf("not supported") }
func (i *grouperIndexer[Out]) Delete(obj any) error { return fmt.Errorf("not supported") }

func (i *grouperIndexer[Out]) List() []any {
	states := i.groups.List()
	res := make([]any, len(states))
	for idx, s := range states {
		res[idx] = s.lastOut
	}
	return res
}

func (i *grouperIndexer[Out]) ListKeys() []string {
	return i.groups.ListKeys()
}

func (i *grouperIndexer[Out]) Get(obj any) (item any, exists bool, err error) {
	key, err := i.keyFunc(obj)
	if err != nil {
		return nil, false, err
	}
	return i.GetByKey(key)
}

func (i *grouperIndexer[Out]) GetByKey(key string) (item any, exists bool, err error) {
	state, ok := i.groups.Get(key)
	if !ok {
		return nil, false, nil
	}
	return state.lastOut, true, nil
}

func (i *grouperIndexer[Out]) Replace(objs []any, rv string) error {
	return fmt.Errorf("not supported")
}

func (i *grouperIndexer[Out]) Resync() error {
	return nil
}

func (i *grouperIndexer[Out]) Bookmark(rv string) {}

func (i *grouperIndexer[Out]) LastStoreSyncResourceVersion() string {
	return ""
}

func (i *grouperIndexer[Out]) Index(indexName string, obj any) ([]any, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *grouperIndexer[Out]) IndexKeys(indexName, indexedValue string) ([]string, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *grouperIndexer[Out]) ListIndexFuncValues(indexName string) []string {
	return nil
}

func (i *grouperIndexer[Out]) ByIndex(indexName, indexedValue string) ([]any, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *grouperIndexer[Out]) GetIndexers() cache.Indexers {
	return nil
}

func (i *grouperIndexer[Out]) AddIndexers(newIndexers cache.Indexers) error {
	return fmt.Errorf("not supported")
}
