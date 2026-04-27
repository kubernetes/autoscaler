package fort

import (
	"time"

	"k8s.io/client-go/tools/cache"
)

// flatMapper implements FlatMap query by applying a 1:N mapping function to each 
// object from a source informer. It embeds baseInformer for event management.
// When an input object is updated, flatMapper reconciles the old and new sets 
// of mapped outputs to emit minimal downstream events.
type flatMapper[Out, In any] struct {
	*baseInformer
	mapper       FlatMapFunc[Out, In]
	registration cache.ResourceEventHandlerRegistration
	source       cache.SharedInformer
	indexer      CloneableIndexer
}

var _ ManualSharedInformer = &flatMapper[int, int]{}

func newFlatMapper[Out, In any](lock LockGroup, mapper FlatMapFunc[Out, In], from cache.SharedInformer) *flatMapper[Out, In] {
	return newFlatMapperWithHandler(mapper, from, NewManualSharedInformerWithOptions(lock, DefaultKeyFunc))
}

func newFlatMapperWithHandler[Out, In any](mapper FlatMapFunc[Out, In], from cache.SharedInformer, handler ManualSharedInformer) *flatMapper[Out, In] {
	m := &flatMapper[Out, In]{
		baseInformer: newBaseInformer(handler.GetLockGroup(), handler.GetKeyFunc()),
		mapper:       mapper,
		source:       from,
		indexer:      NewBTreeIndexer(handler.GetKeyFunc()),
	}
	m.baseInformer.parent = m

	m.registration, _ = from.AddEventHandler(m)

	go func() {
		for {
			if m.registration.HasSynced() {
				m.SetHasSynced()
				return
			}
			select {
			case <-time.After(10 * time.Millisecond):
			case <-m.IsStoppedChan():
				return
			}
		}
	}()

	return m
}

func (m *flatMapper[Out, In]) GetStore() cache.Store {
	return m.indexer
}

func (m *flatMapper[Out, In]) GetIndexer() cache.Indexer {
	return m.indexer
}

func (m *flatMapper[O, I]) OnAddLocked(obj any, isInitial bool) {
	input := obj.(I)
	results, _ := m.mapper(input)
	for _, r := range results {
		m.indexer.Add(r)
		m.dispatchAdd(r, isInitial)
	}
}

func (m *flatMapper[O, I]) OnUpdateLocked(oldObj, newObj any) {
	oldInput := oldObj.(I)
	newInput := newObj.(I)
	oldResults, _ := m.mapper(oldInput)
	newResults, _ := m.mapper(newInput)

	keyFunc := m.GetKeyFunc()
	oldKeys := make(map[string]O)
	for _, r := range oldResults {
		key, _ := keyFunc(r)
		oldKeys[key] = r
	}

	newKeys := make(map[string]O)
	for _, r := range newResults {
		key, _ := keyFunc(r)
		newKeys[key] = r
	}

	for key, oldR := range oldKeys {
		if newR, ok := newKeys[key]; ok {
			m.indexer.Update(newR)
			m.dispatchUpdate(oldR, newR)
			delete(newKeys, key)
		} else {
			m.indexer.Delete(oldR)
			m.dispatchDelete(oldR)
		}
	}
	for _, newR := range newKeys {
		m.indexer.Add(newR)
		m.dispatchAdd(newR, false)
	}
}

func (m *flatMapper[O, I]) OnDeleteLocked(oldObj any) {
	input := oldObj.(I)
	results, _ := m.mapper(input)
	for _, r := range results {
		m.indexer.Delete(r)
		m.dispatchDelete(r)
	}
}

func (m *flatMapper[Out, In]) dispatchAdd(obj Out, isInitial bool) {
	for _, h := range m.handlers {
		m.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnAdd(obj, isInitial) }, 
			func(l LockedResourceEventHandler) { l.OnAddLocked(obj, isInitial) })
	}
}

func (m *flatMapper[Out, In]) dispatchUpdate(oldObj, newObj Out) {
	for _, h := range m.handlers {
		m.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnUpdate(oldObj, newObj) }, 
			func(l LockedResourceEventHandler) { l.OnUpdateLocked(oldObj, newObj) })
	}
}

func (m *flatMapper[Out, In]) dispatchDelete(obj Out) {
	for _, h := range m.handlers {
		m.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnDelete(obj) }, 
			func(l LockedResourceEventHandler) { l.OnDeleteLocked(obj) })
	}
}

func (m *flatMapper[Out, In]) Clone(newSources []cache.SharedInformer) CloneableSharedInformerQuery {
	var newSource cache.SharedInformer
	if len(newSources) > 0 {
		newSource = newSources[0]
	} else {
		newSource = m.source
	}

	newLock := newSource.(CloneableSharedInformerQuery).GetLockGroup()

	cloned := &flatMapper[Out, In]{
		baseInformer: m.baseInformer.clone(),
		mapper:       m.mapper,
		source:       newSource,
		indexer:      m.indexer.Clone(),
	}
	cloned.baseInformer.lock = newLock
	cloned.baseInformer.parent = cloned

	if ms, ok := newSource.(ManualSharedInformer); ok {
		cloned.registration, _ = ms.AddEventHandlerNoReplay(cloned)
	} else {
		cloned.registration, _ = newSource.AddEventHandler(cloned)
	}

	return cloned
}

func (m *flatMapper[Out, In]) GetController() cache.Controller {
	return nil
}

func (m *flatMapper[Out, In]) Run(stopCh <-chan struct{}) {
	defer m.SetIsStopped()
	defer func() {
		if m.source != nil && m.registration != nil {
			_ = m.source.RemoveEventHandler(m.registration)
		}
	}()
	<-stopCh
}

func (m *flatMapper[Out, In]) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	return m.source.SetWatchErrorHandler(handler)
}

func (m *flatMapper[Out, In]) SetWatchErrorHandlerWithContext(handler cache.WatchErrorHandlerWithContext) error {
	return m.source.SetWatchErrorHandlerWithContext(handler)
}

func (m *flatMapper[Out, In]) GetSources() []cache.SharedInformer {
	return []cache.SharedInformer{m.source}
}
