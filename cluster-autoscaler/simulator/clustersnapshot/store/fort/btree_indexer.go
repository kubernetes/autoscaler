package fort

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store/fort/btree"
	"k8s.io/client-go/tools/cache"
)

// CloneableIndexer extends cache.Indexer with a fast Clone operation.
type CloneableIndexer interface {
	cache.Indexer
	Clone() CloneableIndexer
}

// BTreeMap is a generic fast-cloneable map with string keys.
// It uses a B-Tree to provide O(1) structural cloning via Copy-on-Write (COW).
// 
// IMPORTANT: The B-Tree structure itself is COW, meaning that tree.Clone() 
// is very fast and doesn't duplicate the data. However, the VALUES stored inside 
// the tree are NOT automatically deep-copied. 
//
// If you store complex types like slices, maps, or pointers to structs, you 
// MUST perform a manual COW (e.g., shallow-clone the slice) BEFORE updating 
// a value in the tree. This ensures that existing snapshots of the tree 
// remain immutable and consistent.
type BTreeMap[V any] interface {
	Get(key string) (V, bool)
	Set(key string, val V)
	Delete(key string)
	List() []V
	ListKeys() []string
	// Clone performs an O(1) structural clone.
	Clone() BTreeMap[V]
	Len() int
}

type btreeMapItem[V any] struct {
	key string
	val V
}

func btreeMapItemLess[V any](a, b btreeMapItem[V]) bool {
	return a.key < b.key
}

type btreeMap[V any] struct {
	mu   sync.RWMutex
	tree *btree.BTree[btreeMapItem[V]]
}

// NewBTreeMap creates a new B-Tree based cloneable map.
func NewBTreeMap[V any]() BTreeMap[V] {
	return &btreeMap[V]{
		tree: btree.New(32, btreeMapItemLess[V]),
	}
}

func (m *btreeMap[V]) Get(key string) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.tree.Get(btreeMapItem[V]{key: key})
	if !ok {
		var zero V
		return zero, false
	}
	return it.val, true
}

func (m *btreeMap[V]) Set(key string, val V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tree.ReplaceOrInsert(btreeMapItem[V]{key: key, val: val})
}

func (m *btreeMap[V]) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tree.Delete(btreeMapItem[V]{key: key})
}

func (m *btreeMap[V]) List() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]V, 0, m.tree.Len())
	m.tree.Ascend(func(item btreeMapItem[V]) bool {
		res = append(res, item.val)
		return true
	})
	return res
}

func (m *btreeMap[V]) ListKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]string, 0, m.tree.Len())
	m.tree.Ascend(func(item btreeMapItem[V]) bool {
		res = append(res, item.key)
		return true
	})
	return res
}

func (m *btreeMap[V]) Clone() BTreeMap[V] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return &btreeMap[V]{
		// tree.Clone() is O(1) structural copy.
		tree: m.tree.Clone(),
	}
}

func (m *btreeMap[V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tree.Len()
}

// btreeIndexer implements CloneableIndexer (cache.Indexer) using a B-Tree for storage.
type btreeIndexer struct {
	keyFunc cache.KeyFunc
	data    BTreeMap[any]
	lastRV  string
}

// NewBTreeIndexer creates a new fast-cloneable indexer.
func NewBTreeIndexer(keyFunc cache.KeyFunc) CloneableIndexer {
	return &btreeIndexer{
		keyFunc: keyFunc,
		data:    NewBTreeMap[any](),
	}
}

func (i *btreeIndexer) Clone() CloneableIndexer {
	return &btreeIndexer{
		keyFunc: i.keyFunc,
		data:    i.data.Clone(), // O(1) structural clone.
		lastRV:  i.lastRV,
	}
}

func (i *btreeIndexer) Add(obj any) error {
	key, err := i.keyFunc(obj)
	if err != nil {
		return err
	}
	i.data.Set(key, obj)
	return nil
}

func (i *btreeIndexer) Update(obj any) error {
	return i.Add(obj)
}

func (i *btreeIndexer) Delete(obj any) error {
	key, err := i.keyFunc(obj)
	if err != nil {
		return err
	}
	i.data.Delete(key)
	return nil
}

func (i *btreeIndexer) List() []any {
	return i.data.List()
}

func (i *btreeIndexer) ListKeys() []string {
	return i.data.ListKeys()
}

func (i *btreeIndexer) Get(obj any) (item any, exists bool, err error) {
	key, err := i.keyFunc(obj)
	if err != nil {
		return nil, false, err
	}
	return i.GetByKey(key)
}

func (i *btreeIndexer) GetByKey(key string) (item any, exists bool, err error) {
	val, ok := i.data.Get(key)
	return val, ok, nil
}

func (i *btreeIndexer) Replace(objs []any, rv string) error {
	i.data = NewBTreeMap[any]()
	i.lastRV = rv
	for _, obj := range objs {
		if err := i.Add(obj); err != nil {
			return err
		}
	}
	return nil
}

func (i *btreeIndexer) Resync() error {
	return nil
}

func (i *btreeIndexer) LastStoreSyncResourceVersion() string {
	return i.lastRV
}

func (i *btreeIndexer) Bookmark(rv string) {
	i.lastRV = rv
}

// Indexer methods stubs - currently not used by Fort queries.
// If indexing support is needed, these must be implemented using B-Trees as well.

func (i *btreeIndexer) Index(indexName string, obj any) ([]any, error) {
	return nil, fmt.Errorf("Index not implemented in BTreeIndexer")
}

func (i *btreeIndexer) IndexKeys(indexName, indexedValue string) ([]string, error) {
	return nil, fmt.Errorf("IndexKeys not implemented in BTreeIndexer")
}

func (i *btreeIndexer) ListIndexFuncValues(indexName string) []string {
	return nil
}

func (i *btreeIndexer) ByIndex(indexName, indexedValue string) ([]any, error) {
	return nil, fmt.Errorf("ByIndex not implemented in BTreeIndexer")
}

func (i *btreeIndexer) GetIndexerResyncPeriod(indexName string) time.Duration {
	return 0
}

func (i *btreeIndexer) GetIndexers() cache.Indexers {
	return nil
}

func (i *btreeIndexer) AddIndexers(newIndexers cache.Indexers) error {
	if len(newIndexers) > 0 {
		return fmt.Errorf("AddIndexers not implemented in BTreeIndexer")
	}
	return nil
}
