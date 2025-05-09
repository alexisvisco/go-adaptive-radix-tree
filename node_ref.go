package art

import (
	"unsafe"
)

// indexNotFound is a special index value
// that indicates that the index is not found.
const indexNotFound = -1

// nodeNotFound is a special Node pointer
// that indicates that the Node is not found
// for different internal tree operations.
var nodeNotFound *NodeRef //nolint:gochecknoglobals

// NodeRef stores all available tree nodes LeafKind and nodeX types
// as a ref to *unsafe* pointer.
// The kind field is used to determine the type of the Node.
type NodeRef struct {
	ref  unsafe.Pointer
	kind Kind
}

type nodeLeafer interface {
	minimum() *Leaf
	maximum() *Leaf
}

type nodeSizeManager interface {
	hasCapacityForChild() bool
	grow() *NodeRef

	isReadyToShrink() bool
	shrink() *NodeRef
}

type nodeOperations interface {
	addChild(kc keyChar, child *NodeRef)
	deleteChild(kc keyChar) int
}

type nodeChildren interface {
	childAt(idx int) **NodeRef
	allChildren() []*NodeRef
}

type nodeKeyIndexer interface {
	index(kc keyChar) int
}

// noder is an interface that defines methods that
// must be implemented by NodeRef and all Node types.
// extra interfaces are used to group methods by their purpose
// and help with code readability.
type noder interface {
	nodeLeafer
	nodeOperations
	nodeChildren
	nodeKeyIndexer
	nodeSizeManager
}

// toNode converts the NodeRef to specific Node type.
// the idea is to avoid type assertion in the code in multiple places.
func toNode(nr *NodeRef) noder {
	if nr == nil {
		return noopNoder
	}

	switch nr.kind { //nolint:exhaustive
	case Node4Kind:
		return nr.node4()
	case Node16Kind:
		return nr.node16()
	case Node48Kind:
		return nr.node48()
	case Node256Kind:
		return nr.node256()
	default:
		return noopNoder
	}
}

// noop is a no-op noder implementation.
type noop struct{}

func (*noop) minimum() *Leaf             { return nil }
func (*noop) maximum() *Leaf             { return nil }
func (*noop) index(keyChar) int          { return indexNotFound }
func (*noop) childAt(int) **NodeRef      { return &nodeNotFound }
func (*noop) allChildren() []*NodeRef    { return nil }
func (*noop) hasCapacityForChild() bool  { return true }
func (*noop) grow() *NodeRef             { return nil }
func (*noop) isReadyToShrink() bool      { return false }
func (*noop) shrink() *NodeRef           { return nil }
func (*noop) addChild(keyChar, *NodeRef) {}
func (*noop) deleteChild(keyChar) int    { return 0 }

// noopNoder is the default Noder implementation.
var noopNoder noder = &noop{} //nolint:gochecknoglobals

// assert that all Node types implement noder interface.
var _ noder = (*Node4)(nil)
var _ noder = (*Node16)(nil)
var _ noder = (*Node48)(nil)
var _ noder = (*Node256)(nil)

// assert that NodeRef implements public NodeKV interface.
var _ NodeKV = (*NodeRef)(nil)

// Kind returns the Node kind.
func (nr *NodeRef) Kind() Kind {
	return nr.kind
}

// Key returns the Node key for LeafKind nodes.
// for nodeX types, it returns nil.
func (nr *NodeRef) Key() Key {
	if nr.isLeaf() {
		return nr.Leaf().key
	}

	return nil
}

// Value returns the Node value for LeafKind nodes.
// for nodeX types, it returns nil.
func (nr *NodeRef) Value() Value {
	if nr.isLeaf() {
		return nr.Leaf().value
	}

	return nil
}

// isLeaf returns true if the Node is a Leaf Node.
func (nr *NodeRef) isLeaf() bool {
	return nr.kind == LeafKind
}

// setPrefix sets the Node prefix with the new prefix and prefix length.
func (nr *NodeRef) setPrefix(newPrefix []byte, prefixLen int) {
	n := nr.node()

	n.prefixLen = uint16(prefixLen) //#nosec:G115
	for i := 0; i < minInt(prefixLen, maxPrefixLen); i++ {
		n.prefix[i] = newPrefix[i]
	}
}

// minimum returns itself if the Node is a Leaf Node.
// otherwise it returns the minimum Leaf Node under the current Node.
func (nr *NodeRef) minimum() *Leaf {
	if nr.kind == LeafKind {
		return nr.Leaf()
	}

	return toNode(nr).minimum()
}

// maximum returns itself if the Node is a Leaf Node.
// otherwise it returns the maximum Leaf Node under the current Node.
func (nr *NodeRef) maximum() *Leaf {
	if nr.kind == LeafKind {
		return nr.Leaf()
	}

	return toNode(nr).maximum()
}

// findChildByKey returns the child Node reference for the given key.
func (nr *NodeRef) findChildByKey(key Key, keyOffset int) **NodeRef {
	n := toNode(nr)
	idx := n.index(key.charAt(keyOffset))

	return n.childAt(idx)
}

// nodeX/LeafKind casts the NodeRef to the specific nodeX/LeafKind type.
func (nr *NodeRef) node() *Node       { return (*Node)(nr.ref) }    // Node casts NodeRef to Node.
func (nr *NodeRef) node4() *Node4     { return (*Node4)(nr.ref) }   // Node4 casts NodeRef to Node4.
func (nr *NodeRef) node16() *Node16   { return (*Node16)(nr.ref) }  // Node16 casts NodeRef to Node16.
func (nr *NodeRef) node48() *Node48   { return (*Node48)(nr.ref) }  // Node48 casts NodeRef to Node48.
func (nr *NodeRef) node256() *Node256 { return (*Node256)(nr.ref) } // Node256 casts NodeRef to Node256.
func (nr *NodeRef) Leaf() *Leaf       { return (*Leaf)(nr.ref) }    // Leaf casts NodeRef to Leaf.

// addChild adds a new child Node to the current Node.
// If the Node is full, it grows to the next Node type.
func (nr *NodeRef) addChild(kc keyChar, child *NodeRef) {
	n := toNode(nr)

	if n.hasCapacityForChild() {
		n.addChild(kc, child)
	} else {
		bigNode := n.grow()         // grow to the next Node type
		bigNode.addChild(kc, child) // recursively add the child to the new Node
		replaceNode(nr, bigNode)    // replace the current Node with the new Node
	}
}

// deleteChild deletes the child Node from the current Node.
// If the Node can shrink after, it shrinks to the previous Node type.
func (nr *NodeRef) deleteChild(kc keyChar) bool {
	shrank := false
	n := toNode(nr)
	n.deleteChild(kc)

	if n.isReadyToShrink() {
		shrank = true
		smallNode := n.shrink()    // shrink to the previous Node type
		replaceNode(nr, smallNode) // replace the current Node with the shrank Node
	}

	return shrank
}

// match finds the first mismatched index between
// the Node's prefix and the specified key prefix.
// This approach efficiently identifies the mismatch by
// leveraging the Node's existing prefix data.
func (nr *NodeRef) match(key Key, keyOffset int) int /* 1st mismatch index*/ {
	// calc the remaining key length from offset
	keyRemaining := len(key) - keyOffset
	if keyRemaining < 0 {
		return 0
	}

	n := nr.node()

	// the maximum length we can check against the Node's prefix
	maxPrefixLen := minInt(int(n.prefixLen), maxPrefixLen)
	limit := minInt(maxPrefixLen, keyRemaining)

	// compare the key against the Node's prefix
	for i := 0; i < limit; i++ {
		if n.prefix[i] != key[keyOffset+i] {
			return i
		}
	}

	return limit
}

// matchDeep returns the first index where the key mismatches,
// starting with the Node's prefix(see match) and continuing with the minimum Leaf's key.
// It returns the mismatch index or matches up to the key's end.
func (nr *NodeRef) matchDeep(key Key, keyOffset int) int /* mismatch index*/ {
	mismatchIdx := nr.match(key, keyOffset)
	if mismatchIdx < maxPrefixLen {
		return mismatchIdx
	}

	leafKey := nr.minimum().key
	limit := minInt(len(leafKey), len(key)) - keyOffset

	for ; mismatchIdx < limit; mismatchIdx++ {
		if leafKey[keyOffset+mismatchIdx] != key[keyOffset+mismatchIdx] {
			break
		}
	}

	return mismatchIdx
}
