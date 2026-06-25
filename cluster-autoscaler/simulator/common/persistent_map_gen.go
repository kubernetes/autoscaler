package common

import (
	"math/bits"
	"unsafe"
)

//go:noescape
//go:linkname nilinterhash runtime.nilinterhash
func nilinterhash(p unsafe.Pointer, h uintptr) uintptr

//go:noescape
//go:linkname strhash runtime.strhash
func strhash(p unsafe.Pointer, h uintptr) uintptr

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, size uintptr) uintptr

func interfaceHash(x interface{}) uint32 {
	return uint32(nilinterhash(unsafe.Pointer(&x), 0))
}

func stringHash(x string) uint32 {
	return uint32(strhash(unsafe.Pointer(&x), 0))
}

func uint8Hash(x uint8) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 1))
}

func int8Hash(x int8) uint32 {
	return uint8Hash(uint8(x))
}

func uint16Hash(x uint16) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 2))
}

func int16Hash(x int16) uint32 {
	return uint16Hash(uint16(x))
}

func uint32Hash(x uint32) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 4))
}

func int32Hash(x int32) uint32 {
	return uint32Hash(uint32(x))
}

func uint64Hash(x uint64) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 8))
}

func int64Hash(x int64) uint32 {
	return uint64Hash(uint64(x))
}

func intHash(x int) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, uintptr(unsafe.Sizeof(x))))
}

func uintHash(x uint) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, uintptr(unsafe.Sizeof(x))))
}

func boolHash(x bool) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 1))
}

func runeHash(x rune) uint32 {
	return int32Hash(int32(x))
}

func float64Hash(x float64) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 8))
}

func float32Hash(x float32) uint32 {
	return uint32(memhash(unsafe.Pointer(&x), 0, 4))
}

func hashKey[K comparable](key K) uint32 {
	switch k := any(key).(type) {
	case string:
		return stringHash(k)
	case int:
		return intHash(k)
	case int64:
		return int64Hash(k)
	case int32:
		return int32Hash(k)
	case uint64:
		return uint64Hash(k)
	case float64:
		return float64Hash(k)
	default:
		return interfaceHash(key)
	}
}

// editSession represents a unique transaction session for transient map mutations.
type editSession struct {
	dummy bool
}

// box is a helper used to return state from recursive calls
type box struct {
	val any
}

// node represents a generic node in the Hash Array Mapped Trie
type node[K comparable, V any] interface {
	assoc(session *editSession, shift uint, hash uint32, key K, val V, addedLeaf *box) node[K, V]
	without(session *editSession, shift uint, hash uint32, key K, removedLeaf *box) node[K, V]
	find(shift uint, hash uint32, key K) (V, bool)
	rangeElements(f func(K, V) bool) bool
}

// bitmapNodeEntry represents a single slot in a BitmapIndexedNode dense array.
type bitmapNodeEntry[K comparable, V any] struct {
	isNode bool
	hash   uint32
	key    K
	value  V
	child  node[K, V]
}

// bitmapIndexedNode is a compact node using a 32-bit bitmap to track occupied slots.
type bitmapIndexedNode[K comparable, V any] struct {
	session *editSession
	bitmap  uint32
	array   []bitmapNodeEntry[K, V]
}

func (n *bitmapIndexedNode[K, V]) index(bit uint32) int {
	return bits.OnesCount32(n.bitmap & (bit - 1))
}

func (n *bitmapIndexedNode[K, V]) ensureEditable(session *editSession) *bitmapIndexedNode[K, V] {
	if session != nil && n.session == session {
		return n
	}
	newArray := make([]bitmapNodeEntry[K, V], len(n.array))
	copy(newArray, n.array)
	return &bitmapIndexedNode[K, V]{session: session, bitmap: n.bitmap, array: newArray}
}

func (n *bitmapIndexedNode[K, V]) find(shift uint, hash uint32, key K) (V, bool) {
	bit := uint32(1) << ((hash >> shift) & 0x1f)
	if (n.bitmap & bit) == 0 {
		var zero V
		return zero, false
	}
	idx := n.index(bit)
	entry := n.array[idx]
	if entry.isNode {
		return entry.child.find(shift+5, hash, key)
	}
	if entry.key == key {
		return entry.value, true
	}
	var zero V
	return zero, false
}

func (n *bitmapIndexedNode[K, V]) assoc(session *editSession, shift uint, hash uint32, key K, val V, addedLeaf *box) node[K, V] {
	bit := uint32(1) << ((hash >> shift) & 0x1f)
	idx := n.index(bit)

	if (n.bitmap & bit) != 0 {
		entry := n.array[idx]
		if entry.isNode {
			child := entry.child.assoc(session, shift+5, hash, key, val, addedLeaf)
			if child == entry.child {
				return n
			}
			editable := n.ensureEditable(session)
			editable.array[idx].child = child
			return editable
		}
		if entry.key == key {
			if safeEqual(entry.value, val) {
				return n
			}
			editable := n.ensureEditable(session)
			editable.array[idx].value = val
			return editable
		}
		addedLeaf.val = addedLeaf
		subNode := createNode(session, shift+5, entry.key, entry.value, entry.hash, key, val, hash)
		editable := n.ensureEditable(session)
		editable.array[idx] = bitmapNodeEntry[K, V]{
			isNode: true,
			hash:   entry.hash,
			child:  subNode,
		}
		return editable
	}

	addedLeaf.val = addedLeaf
	card := bits.OnesCount32(n.bitmap)
	if card >= 16 {
		var nodes [32]node[K, V]
		j := 0
		for i := uint(0); i < 32; i++ {
			if (n.bitmap & (1 << i)) != 0 {
				entry := n.array[j]
				if entry.isNode {
					nodes[i] = entry.child
				} else {
					nodes[i] = &bitmapIndexedNode[K, V]{
						session: session,
						bitmap:  1 << ((entry.hash >> (shift + 5)) & 0x1f),
						array:   []bitmapNodeEntry[K, V]{entry},
					}
				}
				j++
			}
		}
		idx := (hash >> shift) & 0x1f
		nodes[idx] = &bitmapIndexedNode[K, V]{
			session: session,
			bitmap:  1 << ((hash >> (shift + 5)) & 0x1f),
			array: []bitmapNodeEntry[K, V]{{
				isNode: false,
				hash:   hash,
				key:    key,
				value:  val,
			}},
		}
		return &arrayNode[K, V]{session: session, count: card + 1, array: nodes}
	}

	newArray := make([]bitmapNodeEntry[K, V], card+1)
	copy(newArray[:idx], n.array[:idx])
	newArray[idx] = bitmapNodeEntry[K, V]{
		isNode: false,
		hash:   hash,
		key:    key,
		value:  val,
	}
	copy(newArray[idx+1:], n.array[idx:])
	if session != nil && n.session == session {
		n.bitmap |= bit
		n.array = newArray
		return n
	}
	return &bitmapIndexedNode[K, V]{session: session, bitmap: n.bitmap | bit, array: newArray}
}

func (n *bitmapIndexedNode[K, V]) without(session *editSession, shift uint, hash uint32, key K, removedLeaf *box) node[K, V] {
	bit := uint32(1) << ((hash >> shift) & 0x1f)
	if (n.bitmap & bit) == 0 {
		return n
	}
	idx := n.index(bit)
	entry := n.array[idx]
	if entry.isNode {
		child := entry.child.without(session, shift+5, hash, key, removedLeaf)
		if child == entry.child {
			return n
		}
		if child == nil {
			if n.bitmap == bit {
				return nil
			}
			newArray := make([]bitmapNodeEntry[K, V], len(n.array)-1)
			copy(newArray[:idx], n.array[:idx])
			copy(newArray[idx:], n.array[idx+1:])
			if session != nil && n.session == session {
				n.bitmap ^= bit
				n.array = newArray
				return n
			}
			return &bitmapIndexedNode[K, V]{session: session, bitmap: n.bitmap ^ bit, array: newArray}
		}
		editable := n.ensureEditable(session)
		editable.array[idx].child = child
		return editable
	}

	if entry.key == key {
		removedLeaf.val = removedLeaf
		if n.bitmap == bit {
			return nil
		}
		newArray := make([]bitmapNodeEntry[K, V], len(n.array)-1)
		copy(newArray[:idx], n.array[:idx])
		copy(newArray[idx:], n.array[idx+1:])
		if session != nil && n.session == session {
			n.bitmap ^= bit
			n.array = newArray
			return n
		}
		return &bitmapIndexedNode[K, V]{session: session, bitmap: n.bitmap ^ bit, array: newArray}
	}
	return n
}

func (n *bitmapIndexedNode[K, V]) rangeElements(f func(K, V) bool) bool {
	for _, entry := range n.array {
		if entry.isNode {
			if !entry.child.rangeElements(f) {
				return false
			}
		} else {
			if !f(entry.key, entry.value) {
				return false
			}
		}
	}
	return true
}

// arrayNode represents a flat, 32-way branched node for fast lookups in highly populated nodes.
type arrayNode[K comparable, V any] struct {
	session *editSession
	count   int
	array   [32]node[K, V]
}

func (n *arrayNode[K, V]) ensureEditable(session *editSession) *arrayNode[K, V] {
	if session != nil && n.session == session {
		return n
	}
	return &arrayNode[K, V]{session: session, count: n.count, array: n.array}
}

func (n *arrayNode[K, V]) find(shift uint, hash uint32, key K) (V, bool) {
	idx := (hash >> shift) & 0x1f
	child := n.array[idx]
	if child == nil {
		var zero V
		return zero, false
	}
	return child.find(shift+5, hash, key)
}

func (n *arrayNode[K, V]) assoc(session *editSession, shift uint, hash uint32, key K, val V, addedLeaf *box) node[K, V] {
	idx := (hash >> shift) & 0x1f
	child := n.array[idx]
	if child == nil {
		newChild := &bitmapIndexedNode[K, V]{
			session: session,
			bitmap:  1 << ((hash >> (shift + 5)) & 0x1f),
			array: []bitmapNodeEntry[K, V]{{
				isNode: false,
				hash:   hash,
				key:    key,
				value:  val,
			}},
		}
		addedLeaf.val = addedLeaf
		editable := n.ensureEditable(session)
		editable.array[idx] = newChild
		editable.count++
		return editable
	}

	newChild := child.assoc(session, shift+5, hash, key, val, addedLeaf)
	if newChild == child {
		return n
	}
	editable := n.ensureEditable(session)
	editable.array[idx] = newChild
	return editable
}

func (n *arrayNode[K, V]) without(session *editSession, shift uint, hash uint32, key K, removedLeaf *box) node[K, V] {
	idx := (hash >> shift) & 0x1f
	child := n.array[idx]
	if child == nil {
		return n
	}
	newChild := child.without(session, shift+5, hash, key, removedLeaf)
	if newChild == child {
		return n
	}
	if newChild == nil {
		if n.count <= 8 {
			var newArray []bitmapNodeEntry[K, V]
			var bitmap uint32
			for i := uint(0); i < 32; i++ {
				if uint32(i) != idx && n.array[i] != nil {
					bitmap |= 1 << i
					c := n.array[i]
					if bNode, ok := c.(*bitmapIndexedNode[K, V]); ok && bits.OnesCount32(bNode.bitmap) == 1 && !bNode.array[0].isNode {
						newArray = append(newArray, bNode.array[0])
					} else {
						newArray = append(newArray, bitmapNodeEntry[K, V]{
							isNode: true,
							child:  c,
						})
					}
				}
			}
			return &bitmapIndexedNode[K, V]{session: session, bitmap: bitmap, array: newArray}
		}
		editable := n.ensureEditable(session)
		editable.array[idx] = nil
		editable.count--
		return editable
	}
	editable := n.ensureEditable(session)
	editable.array[idx] = newChild
	return editable
}

func (n *arrayNode[K, V]) rangeElements(f func(K, V) bool) bool {
	for _, child := range n.array {
		if child != nil {
			if !child.rangeElements(f) {
				return false
			}
		}
	}
	return true
}

type collisionEntry[K comparable, V any] struct {
	key   K
	value V
}

// hashCollisionNode handles keys that have different values but produce the exact same 32-bit hash.
type hashCollisionNode[K comparable, V any] struct {
	session *editSession
	hash    uint32
	array   []collisionEntry[K, V]
}

func (n *hashCollisionNode[K, V]) ensureEditable(session *editSession) *hashCollisionNode[K, V] {
	if session != nil && n.session == session {
		return n
	}
	newArray := make([]collisionEntry[K, V], len(n.array))
	copy(newArray, n.array)
	return &hashCollisionNode[K, V]{session: session, hash: n.hash, array: newArray}
}

func (n *hashCollisionNode[K, V]) find(shift uint, hash uint32, key K) (V, bool) {
	for _, entry := range n.array {
		if entry.key == key {
			return entry.value, true
		}
	}
	var zero V
	return zero, false
}

func (n *hashCollisionNode[K, V]) assoc(session *editSession, shift uint, hash uint32, key K, val V, addedLeaf *box) node[K, V] {
	if hash == n.hash {
		for idx, entry := range n.array {
			if entry.key == key {
				if safeEqual(entry.value, val) {
					return n
				}
				editable := n.ensureEditable(session)
				editable.array[idx].value = val
				return editable
			}
		}

		addedLeaf.val = addedLeaf
		newArray := make([]collisionEntry[K, V], len(n.array)+1)
		copy(newArray, n.array)
		newArray[len(n.array)] = collisionEntry[K, V]{key: key, value: val}
		if session != nil && n.session == session {
			n.array = newArray
			return n
		}
		return &hashCollisionNode[K, V]{session: session, hash: n.hash, array: newArray}
	}

	return (&bitmapIndexedNode[K, V]{
		session: session,
		bitmap:  1 << ((n.hash >> shift) & 0x1f),
		array: []bitmapNodeEntry[K, V]{{
			isNode: true,
			hash:   n.hash,
			child:  n,
		}},
	}).assoc(session, shift, hash, key, val, addedLeaf)
}

func (n *hashCollisionNode[K, V]) without(session *editSession, shift uint, hash uint32, key K, removedLeaf *box) node[K, V] {
	if hash != n.hash {
		return n
	}
	for idx, entry := range n.array {
		if entry.key == key {
			removedLeaf.val = removedLeaf
			if len(n.array) == 1 {
				return nil
			}
			if len(n.array) == 2 {
				otherIdx := 1 - idx
				other := n.array[otherIdx]
				return &bitmapIndexedNode[K, V]{
					session: session,
					bitmap:  1 << ((n.hash >> shift) & 0x1f),
					array: []bitmapNodeEntry[K, V]{{
						isNode: false,
						hash:   n.hash,
						key:    other.key,
						value:  other.value,
					}},
				}
			}
			newArray := make([]collisionEntry[K, V], len(n.array)-1)
			copy(newArray[:idx], n.array[:idx])
			copy(newArray[idx:], n.array[idx+1:])
			if session != nil && n.session == session {
				n.array = newArray
				return n
			}
			return &hashCollisionNode[K, V]{session: session, hash: n.hash, array: newArray}
		}
	}
	return n
}

func (n *hashCollisionNode[K, V]) rangeElements(f func(K, V) bool) bool {
	for _, entry := range n.array {
		if !f(entry.key, entry.value) {
			return false
		}
	}
	return true
}

func createNode[K comparable, V any](session *editSession, shift uint, key1 K, val1 V, hash1 uint32, key2 K, val2 V, hash2 uint32) node[K, V] {
	if hash1 == hash2 {
		return &hashCollisionNode[K, V]{
			session: session,
			hash:    hash1,
			array: []collisionEntry[K, V]{
				{key: key1, value: val1},
				{key: key2, value: val2},
			},
		}
	}

	box := &box{}
	return (&bitmapIndexedNode[K, V]{session: session, bitmap: 0, array: nil}).
		assoc(session, shift, hash1, key1, val1, box).
		assoc(session, shift, hash2, key2, val2, box)
}

// persistentMap is a true Hash Array Mapped Trie (HAMT) persistent key-value map.
type persistentMap[K comparable, V any] struct {
	root node[K, V]
	size int
}

func (m *persistentMap[K, V]) Len() int {
	return m.size
}

func (m *persistentMap[K, V]) Load(key K) (value V, ok bool) {
	if m.root == nil {
		var zero V
		return zero, false
	}
	return m.root.find(0, hashKey(key), key)
}

func (m *persistentMap[K, V]) Store(key K, value V) *persistentMap[K, V] {
	hash := hashKey(key)
	addedLeaf := &box{}
	var newRoot node[K, V]
	if m.root == nil {
		newRoot = &bitmapIndexedNode[K, V]{
			bitmap: 1 << ((hash >> 0) & 0x1f),
			array: []bitmapNodeEntry[K, V]{{
				isNode: false,
				hash:   hash,
				key:    key,
				value:  value,
			}},
		}
		addedLeaf.val = addedLeaf
	} else {
		newRoot = m.root.assoc(nil, 0, hash, key, value, addedLeaf)
	}

	if newRoot == m.root {
		return m
	}

	newSize := m.size
	if addedLeaf.val != nil {
		newSize++
	}

	return &persistentMap[K, V]{root: newRoot, size: newSize}
}

func (m *persistentMap[K, V]) Delete(key K) *persistentMap[K, V] {
	if m.root == nil {
		return m
	}
	hash := hashKey(key)
	removedLeaf := &box{}
	newRoot := m.root.without(nil, 0, hash, key, removedLeaf)
	if newRoot == m.root {
		return m
	}
	newSize := m.size
	if removedLeaf.val != nil {
		newSize--
	}
	return &persistentMap[K, V]{root: newRoot, size: newSize}
}

func (m *persistentMap[K, V]) Range(f func(K, V) bool) {
	if m.root != nil {
		m.root.rangeElements(f)
	}
}

func (m *persistentMap[K, V]) ToNativeMap() map[K]V {
	result := make(map[K]V)
	m.Range(func(key K, value V) bool {
		result[key] = value
		return true
	})
	return result
}

func (m *persistentMap[K, V]) AsTransient() *transientMap[K, V] {
	return &transientMap[K, V]{
		session: &editSession{},
		root:    m.root,
		size:    m.size,
	}
}

type persistentMapItem[K comparable, V any] struct {
	Key   K
	Value V
}

// NewpersistentMap returns a new persistentMap containing all items.
func NewpersistentMap[K comparable, V any](items ...persistentMapItem[K, V]) *persistentMap[K, V] {
	if len(items) == 0 {
		return &persistentMap[K, V]{}
	}
	session := &editSession{}
	transient := &transientMap[K, V]{session: session}
	for _, item := range items {
		transient.Store(item.Key, item.Value)
	}
	return transient.Persistent()
}

// NewpersistentMapFromNativeMap returns a new persistentMap containing all items in m.
func NewpersistentMapFromNativeMap[K comparable, V any](m map[K]V) *persistentMap[K, V] {
	if len(m) == 0 {
		return &persistentMap[K, V]{}
	}
	session := &editSession{}
	transient := &transientMap[K, V]{session: session}
	for key, value := range m {
		transient.Store(key, value)
	}
	return transient.Persistent()
}

// transientMap is a mutable, single-threaded view of a HAMT that optimizes batch modifications.
type transientMap[K comparable, V any] struct {
	session *editSession
	root    node[K, V]
	size    int
}

func (t *transientMap[K, V]) Len() int {
	return t.size
}

func (t *transientMap[K, V]) Load(key K) (V, bool) {
	if t.root == nil {
		var zero V
		return zero, false
	}
	return t.root.find(0, hashKey(key), key)
}

func (t *transientMap[K, V]) Store(key K, value V) {
	if t.session == nil {
		panic("Transient Store called after Persistent transition!")
	}
	hash := hashKey(key)
	addedLeaf := &box{}
	if t.root == nil {
		t.root = &bitmapIndexedNode[K, V]{
			session: t.session,
			bitmap:  1 << ((hash >> 0) & 0x1f),
			array: []bitmapNodeEntry[K, V]{{
				isNode: false,
				hash:   hash,
				key:    key,
				value:  value,
			}},
		}
		addedLeaf.val = addedLeaf
	} else {
		t.root = t.root.assoc(t.session, 0, hash, key, value, addedLeaf)
	}

	if addedLeaf.val != nil {
		t.size++
	}
}

func (t *transientMap[K, V]) Delete(key K) {
	if t.session == nil {
		panic("Transient Delete called after Persistent transition!")
	}
	if t.root == nil {
		return
	}
	hash := hashKey(key)
	removedLeaf := &box{}
	t.root = t.root.without(t.session, 0, hash, key, removedLeaf)
	if removedLeaf.val != nil {
		t.size--
	}
}

func (t *transientMap[K, V]) Range(f func(K, V) bool) {
	if t.root != nil {
		t.root.rangeElements(f)
	}
}

func (t *transientMap[K, V]) ToNativeMap() map[K]V {
	result := make(map[K]V)
	t.Range(func(key K, value V) bool {
		result[key] = value
		return true
	})
	return result
}

func (t *transientMap[K, V]) Persistent() *persistentMap[K, V] {
	if t.session == nil {
		panic("Transient Persistent called multiple times!")
	}
	p := &persistentMap[K, V]{root: t.root, size: t.size}
	t.root = nil
	t.session = nil // Permanently freeze all nodes owned by this session
	return p
}
