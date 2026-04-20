package fort

import (
	"fmt"
	"time"

	"k8s.io/client-go/tools/cache"
)

// joiner implements a many-to-many join between two source informers.
// It maintains internal B-Trees of join sets to allow fast multi-match lookups.
// It embeds baseInformer to handle downstream event propagation.
type joiner[L, R any] struct {
	*baseInformer
	on      JoinOnFunc[L, R]

	leftSource  cache.SharedInformer
	rightSource cache.SharedInformer

	leftRegistration  cache.ResourceEventHandlerRegistration
	rightRegistration cache.ResourceEventHandlerRegistration

	left  BTreeMap[[]L]
	right BTreeMap[[]R]
}

var _ ManualSharedInformer = &joiner[int, int]{}

func newJoiner[L, R any](lock LockGroup, leftSource, rightSource cache.SharedInformer, on JoinOnFunc[L, R]) *joiner[L, R] {
	return newJoinerWithHandler(leftSource, rightSource, on, NewManualSharedInformerWithOptions(lock, DefaultKeyFunc))
}

func newJoinerWithHandler[L, R any](leftSource, rightSource cache.SharedInformer, on JoinOnFunc[L, R], handler ManualSharedInformer) *joiner[L, R] {
	j := &joiner[L, R]{
		baseInformer: newBaseInformer(handler.GetLockGroup(), handler.GetKeyFunc()),
		on:          on,
		leftSource:  leftSource,
		rightSource: rightSource,
		left:        NewBTreeMap[[]L](),
		right:       NewBTreeMap[[]R](),
	}
	j.baseInformer.parent = j

	j.leftRegistration, _ = leftSource.AddEventHandler(leftHandler[L, R]{j})
	j.rightRegistration, _ = rightSource.AddEventHandler(rightHandler[L, R]{j})

	go func() {
		for {
			if j.leftRegistration.HasSynced() && j.rightRegistration.HasSynced() {
				j.SetHasSynced()
				return
			}
			select {
			case <-time.After(10 * time.Millisecond):
			case <-j.IsStoppedChan():
				return
			}
		}
	}()

	return j
}

func (j *joiner[L, R]) GetIndexer() cache.Indexer {
	return nil
}

func (j *joiner[L, R]) GetStore() cache.Store {
	return nil
}

func (j *joiner[L, R]) OnAddLocked(obj any, init bool) {}
func (j *joiner[L, R]) OnUpdateLocked(old, new any)  {}
func (j *joiner[L, R]) OnDeleteLocked(obj any)       {}

type leftHandler[L, R any] struct{ j *joiner[L, R] }

func (h leftHandler[L, R]) OnAdd(obj any, init bool) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnAddLocked(obj, init)
}
func (h leftHandler[L, R]) OnAddLocked(obj any, init bool) { h.j.onLeftAdd(obj.(L), init) }
func (h leftHandler[L, R]) OnUpdate(old, new any) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnUpdateLocked(old, new)
}
func (h leftHandler[L, R]) OnUpdateLocked(old, new any) { h.j.onLeftUpdate(old.(L), new.(L)) }
func (h leftHandler[L, R]) OnDelete(obj any) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnDeleteLocked(obj)
}
func (h leftHandler[L, R]) OnDeleteLocked(obj any) { h.j.onLeftDelete(obj.(L)) }

type rightHandler[L, R any] struct{ j *joiner[L, R] }

func (h rightHandler[L, R]) OnAdd(obj any, init bool) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnAddLocked(obj, init)
}
func (h rightHandler[L, R]) OnAddLocked(obj any, init bool) { h.j.onRightAdd(obj.(R), init) }
func (h rightHandler[L, R]) OnUpdate(old, new any) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnUpdateLocked(old, new)
}
func (h rightHandler[L, R]) OnUpdateLocked(old, new any) { h.j.onRightUpdate(old.(R), new.(R)) }
func (h rightHandler[L, R]) OnDelete(obj any) {
	h.j.GetLockGroup().Lock()
	defer h.j.GetLockGroup().Unlock()
	h.OnDeleteLocked(obj)
}
func (h rightHandler[L, R]) OnDeleteLocked(obj any) { h.j.onRightDelete(obj.(R)) }

func objectsEqual(a, b any) bool {
	ka := DefaultKey(a)
	kb := DefaultKey(b)
	return ka == kb
}

func (j *joiner[L, R]) onLeftAdd(left L, isInitial bool) {
	key := j.on(left, *new(R))
	keyStr := DefaultKey(key)

	items, _ := j.left.Get(keyStr)
	newItems := append(append([]L(nil), items...), left)
	j.left.Set(keyStr, newItems)

	if rights, ok := j.right.Get(keyStr); ok {
		for _, right := range rights {
			j.dispatchAdd(JoinValue[L, R]{Left: left, Right: right}, isInitial)
		}
	}
}

func (j *joiner[L, R]) onLeftUpdate(oldLeft, newLeft L) {
	oldKey := j.on(oldLeft, *new(R))
	newKey := j.on(newLeft, *new(R))

	if oldKey == newKey {
		keyStr := DefaultKey(oldKey)
		if rights, ok := j.right.Get(keyStr); ok {
			for _, right := range rights {
				j.dispatchUpdate(JoinValue[L, R]{Left: oldLeft, Right: right}, JoinValue[L, R]{Left: newLeft, Right: right})
			}
		}
		slice, _ := j.left.Get(keyStr)
		newSlice := append([]L(nil), slice...)
		for i, v := range newSlice {
			if objectsEqual(v, oldLeft) {
				newSlice[i] = newLeft
				break
			}
		}
		j.left.Set(keyStr, newSlice)
	} else {
		j.onLeftDelete(oldLeft)
		j.onLeftAdd(newLeft, false)
	}
}

func (j *joiner[L, R]) onLeftDelete(left L) {
	key := j.on(left, *new(R))
	keyStr := DefaultKey(key)

	if rights, ok := j.right.Get(keyStr); ok {
		for _, right := range rights {
			j.dispatchDelete(JoinValue[L, R]{Left: left, Right: right})
		}
	}
	slice, _ := j.left.Get(keyStr)
	newSlice := make([]L, 0, len(slice))
	for _, v := range slice {
		if !objectsEqual(v, left) {
			newSlice = append(newSlice, v)
		}
	}
	if len(newSlice) == 0 {
		j.left.Delete(keyStr)
	} else {
		j.left.Set(keyStr, newSlice)
	}
}

func (j *joiner[L, R]) onRightAdd(right R, isInitial bool) {
	key := j.on(*new(L), right)
	keyStr := DefaultKey(key)

	items, _ := j.right.Get(keyStr)
	newItems := append(append([]R(nil), items...), right)
	j.right.Set(keyStr, newItems)

	if lefts, ok := j.left.Get(keyStr); ok {
		for _, left := range lefts {
			j.dispatchAdd(JoinValue[L, R]{Left: left, Right: right}, isInitial)
		}
	}
}

func (j *joiner[L, R]) onRightUpdate(oldRight, newRight R) {
	oldKey := j.on(*new(L), oldRight)
	newKey := j.on(*new(L), newRight)

	if oldKey == newKey {
		keyStr := DefaultKey(oldKey)
		if lefts, ok := j.left.Get(keyStr); ok {
			for _, left := range lefts {
				j.dispatchUpdate(JoinValue[L, R]{Left: left, Right: oldRight}, JoinValue[L, R]{Left: left, Right: newRight})
			}
		}
		slice, _ := j.right.Get(keyStr)
		newSlice := append([]R(nil), slice...)
		for i, v := range newSlice {
			if objectsEqual(v, oldRight) {
				newSlice[i] = newRight
				break
			}
		}
		j.right.Set(keyStr, newSlice)
	} else {
		j.onRightDelete(oldRight)
		j.onRightAdd(newRight, false)
	}
}

func (j *joiner[L, R]) onRightDelete(right R) {
	key := j.on(*new(L), right)
	keyStr := DefaultKey(key)

	if lefts, ok := j.left.Get(keyStr); ok {
		for _, left := range lefts {
			j.dispatchDelete(JoinValue[L, R]{Left: left, Right: right})
		}
	}
	slice, _ := j.right.Get(keyStr)
	newSlice := make([]R, 0, len(slice))
	for _, v := range slice {
		if !objectsEqual(v, right) {
			newSlice = append(newSlice, v)
		}
	}
	if len(newSlice) == 0 {
		j.right.Delete(keyStr)
	} else {
		j.right.Set(keyStr, newSlice)
	}
}

func (j *joiner[L, R]) dispatchAdd(obj JoinValue[L, R], isInitial bool) {
	for _, h := range j.handlers {
		j.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnAdd(obj, isInitial) }, 
			func(l LockedResourceEventHandler) { l.OnAddLocked(obj, isInitial) })
	}
}

func (j *joiner[L, R]) dispatchUpdate(oldObj, newObj JoinValue[L, R]) {
	for _, h := range j.handlers {
		j.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnUpdate(oldObj, newObj) }, 
			func(l LockedResourceEventHandler) { l.OnUpdateLocked(oldObj, newObj) })
	}
}

func (j *joiner[L, R]) dispatchDelete(obj JoinValue[L, R]) {
	for _, h := range j.handlers {
		j.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnDelete(obj) }, 
			func(l LockedResourceEventHandler) { l.OnDeleteLocked(obj) })
	}
}

func (j *joiner[L, R]) Clone(newSources []cache.SharedInformer) CloneableSharedInformerQuery {
	nl := newSources[0]
	nr := newSources[1]
	newLock := nl.(CloneableSharedInformerQuery).GetLockGroup()
	
	nj := &joiner[L, R]{
		baseInformer: j.baseInformer.clone(),
		on:          j.on,
		leftSource:  nl,
		rightSource: nr,
		left:        j.left.Clone(),
		right:       j.right.Clone(),
	}
	nj.baseInformer.lock = newLock
	nj.baseInformer.parent = nj

	if ms, ok := nl.(ManualSharedInformer); ok {
		nj.leftRegistration, _ = ms.AddEventHandlerNoReplay(leftHandler[L, R]{nj})
	} else {
		nj.leftRegistration, _ = nl.AddEventHandler(leftHandler[L, R]{nj})
	}
	if ms, ok := nr.(ManualSharedInformer); ok {
		nj.rightRegistration, _ = ms.AddEventHandlerNoReplay(rightHandler[L, R]{nj})
	} else {
		nj.rightRegistration, _ = nr.AddEventHandler(rightHandler[L, R]{nj})
	}

	return nj
}

func (j *joiner[L, R]) GetController() cache.Controller {
	return nil
}

func (j *joiner[L, R]) Run(stopCh <-chan struct{}) {
	defer j.SetIsStopped()
	defer func() {
		if j.leftSource != nil && j.leftRegistration != nil {
			_ = j.leftSource.RemoveEventHandler(j.leftRegistration)
		}
		if j.rightSource != nil && j.rightRegistration != nil {
			_ = j.rightSource.RemoveEventHandler(j.rightRegistration)
		}
	}()
	<-stopCh
}

func (j *joiner[L, R]) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	_ = j.leftSource.SetWatchErrorHandler(handler)
	_ = j.rightSource.SetWatchErrorHandler(handler)
	return nil
}

func (j *joiner[L, R]) SetWatchErrorHandlerWithContext(handler cache.WatchErrorHandlerWithContext) error {
	_ = j.leftSource.SetWatchErrorHandlerWithContext(handler)
	_ = j.rightSource.SetWatchErrorHandlerWithContext(handler)
	return nil
}

func (j *joiner[L, R]) GetSources() []cache.SharedInformer {
	return []cache.SharedInformer{j.leftSource, j.rightSource}
}

// joinerIndexer implements cache.Store by computing join results on the fly 
// from the left and right B-Trees. This avoids the O(N^2) memory cost of 
// storing all joined pairs explicitly, at the cost of O(N) lookup in GetByKey.
type joinerIndexer[L, R any] struct {
	left    BTreeMap[[]L]
	right   BTreeMap[[]R]
	keyFunc cache.KeyFunc
}

func (i *joinerIndexer[L, R]) Clone() CloneableIndexer {
	return &joinerIndexer[L, R]{
		left:    i.left.Clone(),
		right:   i.right.Clone(),
		keyFunc: i.keyFunc,
	}
}

func (i *joinerIndexer[L, R]) Add(obj any) error    { return fmt.Errorf("not supported") }
func (i *joinerIndexer[L, R]) Update(obj any) error { return fmt.Errorf("not supported") }
func (i *joinerIndexer[L, R]) Delete(obj any) error { return fmt.Errorf("not supported") }

func (i *joinerIndexer[L, R]) List() []any {
	var res []any
	for _, key := range i.left.ListKeys() {
		lefts, _ := i.left.Get(key)
		rights, ok := i.right.Get(key)
		if ok {
			for _, l := range lefts {
				for _, r := range rights {
					res = append(res, JoinValue[L, R]{Left: l, Right: r})
				}
			}
		}
	}
	return res
}

func (i *joinerIndexer[L, R]) ListKeys() []string {
	var res []string
	for _, key := range i.left.ListKeys() {
		lefts, _ := i.left.Get(key)
		rights, ok := i.right.Get(key)
		if ok {
			for _, l := range lefts {
				for _, r := range rights {
					val := JoinValue[L, R]{Left: l, Right: r}
					keyStr, _ := i.keyFunc(val)
					res = append(res, keyStr)
				}
			}
		}
	}
	return res
}

func (i *joinerIndexer[L, R]) Get(obj any) (item any, exists bool, err error) {
	key, err := i.keyFunc(obj)
	if err != nil {
		return nil, false, err
	}
	return i.GetByKey(key)
}

func (i *joinerIndexer[L, R]) GetByKey(key string) (item any, exists bool, err error) {
	for _, obj := range i.List() {
		objKey, _ := i.keyFunc(obj)
		if objKey == key {
			return obj, true, nil
		}
	}
	return nil, false, nil
}

func (i *joinerIndexer[L, R]) Replace(objs []any, rv string) error {
	return fmt.Errorf("not supported")
}

func (i *joinerIndexer[L, R]) Resync() error {
	return nil
}

func (i *joinerIndexer[L, R]) Bookmark(rv string) {}

func (i *joinerIndexer[L, R]) LastStoreSyncResourceVersion() string {
	return ""
}

func (i *joinerIndexer[L, R]) Index(indexName string, obj any) ([]any, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *joinerIndexer[L, R]) IndexKeys(indexName, indexedValue string) ([]string, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *joinerIndexer[L, R]) ListIndexFuncValues(indexName string) []string {
	return nil
}

func (i *joinerIndexer[L, R]) ByIndex(indexName, indexedValue string) ([]any, error) {
	return nil, fmt.Errorf("not supported")
}

func (i *joinerIndexer[L, R]) GetIndexers() cache.Indexers {
	return nil
}

func (i *joinerIndexer[L, R]) AddIndexers(newIndexers cache.Indexers) error {
	return fmt.Errorf("not supported")
}
