package fort

import (
	"context"
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// baseInformer implements common SharedInformer logic (registration, sync propagation, 
// error handling) without providing its own storage.
// It is designed to be embedded in specialized informers (manualInformer, queries).
type baseInformer struct {
	name string

	lock LockGroup

	handlers      map[int]cache.ResourceEventHandler
	nextHandlerId int

	synced        chan struct{}
	isStoppedChan chan struct{}

	hasSynced bool
	isStopped bool

	lastSyncResourceVersion string

	transform cache.TransformFunc

	keyFunc cache.KeyFunc

	watchErrorHandler            cache.WatchErrorHandler
	watchErrorHandlerWithContext cache.WatchErrorHandlerWithContext

	// parent is the outer informer that embeds this baseInformer.
	// It allows baseInformer to delegate store-related calls (like GetStore()) 
	// back to the specialized implementation, ensuring that even if a query 
	// doesn't use standard BTree storage, baseInformer's registration 
	// logic (like atomic replay) still works correctly.
	parent ManualSharedInformer
}

// manualInformer implements ManualSharedInformer using a shared LockGroup 
// and a standard B-Tree indexer for storage. It is the primary "leaf" source 
// for Fort query pipelines.
type manualInformer struct {
	*baseInformer
	indexer CloneableIndexer
}

var _ ManualSharedInformer = &manualInformer{}

func (p *baseInformer) GetLockGroup() LockGroup {
	return p.lock
}

func (p *baseInformer) AddEventHandler(h cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	return p.AddEventHandlerWithOptions(h, cache.HandlerOptions{})
}

func (p *baseInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return p.AddEventHandler(handler)
}

func (p *baseInformer) AddEventHandlerWithOptions(handler cache.ResourceEventHandler, options cache.HandlerOptions) (cache.ResourceEventHandlerRegistration, error) {
	return p.addEventHandler(handler, true)
}

func (p *baseInformer) AddEventHandlerNoReplay(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	return p.addEventHandler(handler, false)
}

type manualInformerRegistration struct {
	informer ManualSharedInformer
	id       int
}

var _ cache.ResourceEventHandlerRegistration = &manualInformerRegistration{}

func (r *manualInformerRegistration) HasSynced() bool {
	return r.informer.HasSynced()
}

// addEventHandler registers a handler, optionally replaying the entire current state.
func (p *baseInformer) addEventHandler(handler cache.ResourceEventHandler, replay bool) (cache.ResourceEventHandlerRegistration, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.isStopped {
		return nil, fmt.Errorf("was not added to shared informer because it has stopped already")
	}

	p.nextHandlerId++
	r := &manualInformerRegistration{
		informer: p.parent,
		id:       p.nextHandlerId,
	}
	p.handlers[p.nextHandlerId] = handler

	if replay {
		// Atomic replay of current state to ensure the new handler is fully hydrated.
		if store := p.parent.GetStore(); store != nil {
			list := store.List()
			for _, obj := range list {
				p.dispatchEvent(handler, 
					func(h cache.ResourceEventHandler) { h.OnAdd(obj, true) }, 
					func(l LockedResourceEventHandler) { l.OnAddLocked(obj, true) })
			}
		}
	}

	return r, nil
}

// dispatchEvent selects the appropriate handler method based on whether the handler
// supports the Locked interface. 
func (p *baseInformer) dispatchEvent(h cache.ResourceEventHandler, std func(cache.ResourceEventHandler), locked func(LockedResourceEventHandler)) {
	defer utilruntime.HandleCrash()
	if m, ok := h.(LockedResourceEventHandler); ok {
		locked(m)
	} else {
		std(h)
	}
}

func (p *baseInformer) RemoveEventHandler(r cache.ResourceEventHandlerRegistration) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	pr := r.(*manualInformerRegistration)
	delete(p.handlers, pr.id)
	return nil
}

func (p *baseInformer) Run(stopCh <-chan struct{}) {
	defer p.SetIsStopped()
	<-stopCh
}

func (p *baseInformer) RunWithContext(ctx context.Context) {
	p.Run(ctx.Done())
}

func (p *baseInformer) LastSyncResourceVersion() string {
	return p.lastSyncResourceVersion
}

func (p *baseInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.watchErrorHandler = handler
	return nil
}

func (p *baseInformer) SetWatchErrorHandlerWithContext(handler cache.WatchErrorHandlerWithContext) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.watchErrorHandlerWithContext = handler
	return nil
}

func (p *baseInformer) TriggerWatchError(err error) {
	p.lock.RLock()
	h := p.watchErrorHandler
	hc := p.watchErrorHandlerWithContext
	p.lock.RUnlock()

	if h != nil {
		h(nil, err)
	}
	if hc != nil {
		hc(context.TODO(), nil, err)
	}
}

func (p *baseInformer) SetTransform(handler cache.TransformFunc) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.transform != nil {
		return fmt.Errorf("Setting transform when it is already set.")
	}
	p.transform = handler
	return nil
}

func (p *baseInformer) HasSynced() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.hasSynced
}

func (p *baseInformer) IsStopped() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.isStopped
}

func (p *baseInformer) IsStoppedChan() <-chan struct{} {
	return p.isStoppedChan
}

func (p *baseInformer) SetHasSynced() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.hasSynced {
		return
	}

	p.hasSynced = true
	close(p.synced)
}

func (p *baseInformer) SetIsStopped() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.isStopped {
		return
	}
	p.isStopped = true
	close(p.isStoppedChan)
}

func (p *baseInformer) GetKeyFunc() cache.KeyFunc {
	return p.keyFunc
}

func (p *baseInformer) GetIndexer() cache.Indexer {
	panic("GetIndexer not implemented")
}

func (p *baseInformer) GetStore() cache.Store {
	panic("GetStore not implemented")
}

func (p *baseInformer) GetController() cache.Controller {
	return nil
}

func (p *baseInformer) OnAdd(obj any, isInInitialList bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.parent.OnAddLocked(obj, isInInitialList)
}

func (p *baseInformer) OnUpdate(oldObj, newObj any) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.parent.OnUpdateLocked(oldObj, newObj)
}

func (p *baseInformer) OnDelete(oldObj any) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.parent.OnDeleteLocked(oldObj)
}

func (p *baseInformer) Clear() {
	panic("Clear not implemented")
}

func (p *manualInformer) GetStore() cache.Store {
	return p.indexer
}

func (p *manualInformer) GetIndexer() cache.Indexer {
	return p.indexer
}

func (p *manualInformer) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()
	list := p.indexer.List()
	for _, obj := range list {
		p.OnDeleteLocked(obj)
	}
}

func (p *manualInformer) OnAddLocked(obj any, isInInitialList bool) {
	transformed := obj
	if p.transform != nil {
		transformed, _ = p.transform(obj)
	}

	p.indexer.Add(transformed)

	for _, h := range p.handlers {
		p.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnAdd(transformed, isInInitialList) }, 
			func(l LockedResourceEventHandler) { l.OnAddLocked(transformed, isInInitialList) })
	}
}

func (p *manualInformer) OnUpdateLocked(oldObj, newObj any) {
	oldTransformed := oldObj
	newTransformed := newObj
	if p.transform != nil {
		oldTransformed, _ = p.transform(oldObj)
		newTransformed, _ = p.transform(newObj)
	}

	oldKey, _ := p.keyFunc(oldTransformed)
	newKey, _ := p.keyFunc(newTransformed)

	if oldKey != newKey {
		p.OnDeleteLocked(oldObj)
		p.OnAddLocked(newObj, false)
		return
	}

	p.indexer.Update(newTransformed)

	for _, h := range p.handlers {
		p.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnUpdate(oldTransformed, newTransformed) }, 
			func(l LockedResourceEventHandler) { l.OnUpdateLocked(oldTransformed, newTransformed) })
	}
}

func (p *manualInformer) OnDeleteLocked(oldObj any) {
	unwrapped := UnwrapDeleted(oldObj)
	transformed := unwrapped
	if p.transform != nil {
		transformed, _ = p.transform(unwrapped)
	}

	p.indexer.Delete(transformed)

	for _, h := range p.handlers {
		p.dispatchEvent(h, 
			func(h cache.ResourceEventHandler) { h.OnDelete(transformed) }, 
			func(l LockedResourceEventHandler) { l.OnDeleteLocked(transformed) })
	}
}

func (p *manualInformer) Clone(_ []cache.SharedInformer) CloneableSharedInformerQuery {
	res := &manualInformer{
		baseInformer: p.baseInformer.clone(),
		indexer:      p.indexer.Clone(),
	}
	res.baseInformer.parent = res
	return res
}

func (p *baseInformer) clone() *baseInformer {
	newB := &baseInformer{
		name:                    p.name,
		handlers:                map[int]cache.ResourceEventHandler{},
		transform:               p.transform,
		isStopped:               p.isStopped,
		hasSynced:               p.hasSynced,
		lastSyncResourceVersion: p.lastSyncResourceVersion,
		keyFunc:                 p.keyFunc,
		lock:                    NewLockGroup(),
		synced:                  make(chan struct{}),
		isStoppedChan:           make(chan struct{}),
	}
	if p.hasSynced {
		close(newB.synced)
	}
	if p.isStopped {
		close(newB.isStoppedChan)
	}
	return newB
}

func (p *baseInformer) SetName(name string) {
	p.name = name
}

func (p *baseInformer) GetSources() []cache.SharedInformer {
	return nil
}

func newBaseInformer(lock LockGroup, keyFunc cache.KeyFunc) *baseInformer {
	return &baseInformer{
		handlers:      map[int]cache.ResourceEventHandler{},
		keyFunc:       keyFunc,
		lock:          lock,
		synced:        make(chan struct{}),
		isStoppedChan: make(chan struct{}),
	}
}

// NewManualSharedInformer creates a ManualSharedInformer with a default lock and keyfunc.
func NewManualSharedInformer() ManualSharedInformer {
	return NewManualSharedInformerWithOptions(NewLockGroup(), DefaultKeyFunc)
}

// NewManualSharedInformerWithOptions creates a ManualSharedInformer with specific options.
func NewManualSharedInformerWithOptions(lock LockGroup, keyFunc cache.KeyFunc) ManualSharedInformer {
	res := &manualInformer{
		baseInformer: newBaseInformer(lock, keyFunc),
		indexer:      NewBTreeIndexer(keyFunc),
	}
	res.baseInformer.parent = res
	return res
}
