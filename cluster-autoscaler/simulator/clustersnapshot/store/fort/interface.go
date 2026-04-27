package fort

import (
	"fmt"
	"strconv"
	"sync"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type explicitKeyProvider interface {
	ExplicitKey() string
}

// DefaultKey is a robust key function that handles both K8s objects and primitive types.
// It ensures that even non-meta objects can be indexed without returning errors.
func DefaultKey(obj any) string {
	if obj == nil {
		return "<nil>"
	}
	switch v := obj.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case [2]string:
		return v[0] + "\x00" + v[1]
	case [3]string:
		return v[0] + "\x00" + v[1] + "\x00" + v[2]
	case explicitKeyProvider:
		return v.ExplicitKey()
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Warningf("DefaultKey: unexpected type %T, falling back to string conversion: %v", obj, err)
		return fmt.Sprintf("%v", obj)
	}
	return key
}

// DefaultKeyFunc wraps DefaultKey to satisfy the cache.KeyFunc signature.
func DefaultKeyFunc(obj any) (string, error) {
	return DefaultKey(obj), nil
}

// UnwrapDeleted returns the underlying object if it's wrapped in a DeletedFinalStateUnknown container.
func UnwrapDeleted(obj any) any {
	if d, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return d.Obj
	}
	return obj
}

// LockGroup manages a shared RWMutex for a connected set of query informers (a "Domain").
// In Fort, an entire query Directed Acyclic Graph (DAG)—from leaf sources to final aggregates—
// should share a single LockGroup. This shared lock ensures transactional consistency
// across the domain during complex updates and snapshots.
// When a source informer receives an event, the entire propagation through the DAG
// happens under this lock, ensuring that at any point, a reader sees a consistent
// state across all stages of the pipeline.
type LockGroup interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type lockGroup struct {
	sync.RWMutex
}

// NewLockGroup creates a new shared mutex domain.
func NewLockGroup() LockGroup {
	return &lockGroup{}
}

// CloneableSharedInformerQuery is a shared informer defined by a query that supports
// high-performance cloning. Clones are "born hydrated," meaning they inherit the
// full current state of their parent via O(1) Copy-on-Write (COW) structural
// cloning of the underlying B-Trees. This allows creating consistent snapshots
// of complex query pipelines nearly instantaneously, regardless of dataset size.
type CloneableSharedInformerQuery interface {
	cache.SharedInformer
	// Clone creates a new instance of the query using the provided new sources.
	// The cloned informer starts with the exact same data state as the parent
	// at the moment of cloning.
	//
	// REQUIRES: The caller MUST hold the shared LockGroup of the parent.
	// For most query types, an RLock is sufficient, but ManualSharedInformer
	// and queries that use structural COW cloning (B-Trees) require an
	// exclusive Lock to safely clone their underlying indexers.
	Clone(newSources []cache.SharedInformer) CloneableSharedInformerQuery
	// GetLockGroup returns the shared lock used by this informer domain.
	GetLockGroup() LockGroup
	// SetName sets the name for debug logging.
	SetName(name string)
	// GetSources returns the upstream informers providing data to this query.
	GetSources() []cache.SharedInformer
	// IsStoppedChan returns a channel that is closed when the informer is stopped.
	IsStoppedChan() <-chan struct{}
	// GetKeyFunc returns the key function used by this informer.
	GetKeyFunc() cache.KeyFunc
	// GetIndexer returns the underlying indexer.
	GetIndexer() cache.Indexer
	// GetStore returns the underlying store.
	GetStore() cache.Store
}

// QueryInformer generates a new SharedInformer by running the given query spec.
func QueryInformer(query QuerySpec) CloneableSharedInformerQuery {
	return query.Build()
}

// QuerySpec defines the interface for building a query informer from a declarative specification.
type QuerySpec interface {
	Build() CloneableSharedInformerQuery
}

// Select defines a simple transformation and filtering query.
type Select[Out, In any] struct {
	Lock   LockGroup
	Select SingleSelectFunc[Out, In]
	From   cache.SharedInformer
	Where  SingleFilterFunc[In]
}

type SingleSelectFunc[Out, Left any] func(value Left) (Out, error)
type SingleFilterFunc[In any] func(in In) bool

// JoinValue represents a pair of joined objects.
type JoinValue[Left, Right any] struct {
	Left  Left
	Right Right
}

// Join defines a many-to-many join between two informers.
type Join[Out, Left, Right any] struct {
	Lock   LockGroup
	Select JoinSelectFunc[Out, Left, Right]
	From   cache.SharedInformer
	Join   cache.SharedInformer
	// On defines the join key. If nil, a full Cartesian join is performed.
	On     JoinOnFunc[Left, Right]
	Where  JoinFilterFunc[Left, Right]
}

type JoinSelectFunc[Out, Left, Right any] func(left Left, right Right) (Out, error)
type JoinFilterFunc[Left, Right any] func(left Left, right Right) bool

// JoinOnFunc defines the key used for joining two objects.
type JoinOnFunc[Left, Right any] func(left Left, right Right) any

// GroupBy defines an aggregation query over a single informer.
type GroupBy[Out, In any] struct {
	Lock    LockGroup
	Select  GroupSelectFunc[Out]
	From    cache.SharedInformer
	Where   SingleFilterFunc[In]
	GroupBy SingleGroupByFunc[In]
}

type GroupSelectFunc[Out any] func(fields []GroupField) (Out, error)
type SingleGroupByFunc[In any] func(in In) (any, []GroupField)

// GroupByJoin defines an aggregation query over the results of a join.
type GroupByJoin[Out, Left, Right any] struct {
	Lock    LockGroup
	Select  GroupSelectFunc[Out]
	From    cache.SharedInformer
	Join    cache.SharedInformer
	On      JoinOnFunc[Left, Right]
	Where   JoinFilterFunc[Left, Right]
	GroupBy JoinGroupByFunc[Left, Right]
}

type JoinGroupByFunc[Left, Right any] func(left Left, right Right) (any, []GroupField)

// GroupField represents an individual aggregate field in a GroupBy query.
type GroupField interface{}

type groupField struct {
	key      any
	count    bool
	sum      *int64
	distinct any
	anyValue any
}

// Aggregate builders.

func GroupKey(key any) GroupField {
	return &groupField{key: key}
}

func Count() GroupField {
	return &groupField{count: true}
}

func Sum(val int64) GroupField {
	return &groupField{sum: &val}
}

func Distinct(val any) GroupField {
	return &groupField{distinct: val}
}

func AnyValue(val any) GroupField {
	return &groupField{anyValue: val}
}

// FlatMap defines a one-to-many transformation query.
type FlatMap[Out, In any] struct {
	Lock LockGroup
	Map  FlatMapFunc[Out, In]
	Over cache.SharedInformer
}

type FlatMapFunc[Out, In any] func(obj In) ([]Out, error)

// LockedResourceEventHandler allows processing events without re-acquiring the LockGroup.
// Internal query stages use this to propagate events across the shared-lock domain.
type LockedResourceEventHandler interface {
	OnAddLocked(obj any, isInInitialList bool)
	OnUpdateLocked(oldObj, newObj any)
	OnDeleteLocked(oldObj any)
}

// ManualSharedInformer allows manual triggering of events.
// It is primarily used for testing and for providing hydrated snapshot sources.
type ManualSharedInformer interface {
	CloneableSharedInformerQuery
	cache.ResourceEventHandler
	LockedResourceEventHandler

	SetIsStopped()
	SetHasSynced()
	GetKeyFunc() cache.KeyFunc
	TriggerWatchError(err error)

	// Clear removes all objects from the informer and notifies handlers.
	Clear()

	// AddEventHandlerNoReplay registers a handler without replaying current state.
	// Used during pipeline cloning to prevent redundant O(N) hydration.
	AddEventHandlerNoReplay(h cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error)
}

// SnapshotLockDomain acquires an exclusive Lock on the domain (LockGroup),
// enabling a consistent and safe snapshot. Exclusive lock is REQUIRED because
// B-Tree structural cloning is not thread-safe for concurrent read-only clones.
func SnapshotLockDomain(informers ...CloneableSharedInformerQuery) DomainLock {
	if len(informers) == 0 {
		return &domainLock{}
	}
	lock := informers[0].GetLockGroup()
	lock.Lock()
	return &domainLock{lock: lock, exclusive: true}
}

// DomainLock provides a handle to release a domain-level lock.
type DomainLock interface {
	Unlock()
}

type domainLock struct {
	lock      LockGroup
	exclusive bool
}

func (ls *domainLock) Unlock() {
	if ls.lock != nil {
		if ls.exclusive {
			ls.lock.Unlock()
		} else {
			ls.lock.RUnlock()
		}
	}
}

// ClonePipeline recursively clones a query DAG starting from root, replacing leaf sources.
// It ensures that shared branches in the DAG are only cloned once (memoization).
//
// REQUIRES: The caller MUST hold an exclusive lock on the LockGroup of the root informer
// (and thus the entire domain) before calling this, as it triggers structural B-Tree clones.
func ClonePipeline(root cache.SharedInformer, memo map[cache.SharedInformer]cache.SharedInformer) cache.SharedInformer {
	return clonePipelineRecursive(root, memo, 0)
}

const maxCloneDepth = 100

func clonePipelineRecursive(root cache.SharedInformer, memo map[cache.SharedInformer]cache.SharedInformer, depth int) cache.SharedInformer {
	if depth > maxCloneDepth {
		panic(fmt.Sprintf("Recursive clone depth exceeded %d (possible cycle in query DAG)", maxCloneDepth))
	}

	if repl, ok := memo[root]; ok {
		return repl
	}

	q, ok := root.(CloneableSharedInformerQuery)
	if !ok {
		return root
	}

	sources := q.GetSources()
	if len(sources) == 0 {
		// This is a leaf that was not in the initial memo map.
		return root
	}

	newSources := make([]cache.SharedInformer, len(sources))
	for i, s := range sources {
		newSources[i] = clonePipelineRecursive(s, memo, depth+1)
	}

	res := q.Clone(newSources)
	memo[root] = res
	return res
}
