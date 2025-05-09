package art

import (
	"unsafe"
)

// nodeFactory is an interface for creating various types of ART nodes,
// including nodes with different capacities and Leaf nodes.
type nodeFactory interface {
	newNode4() *NodeRef
	newNode16() *NodeRef
	newNode48() *NodeRef
	newNode256() *NodeRef

	newLeaf(key Key, value interface{}) *NodeRef
}

// make sure that objFactory implements all methods of nodeFactory interface.
var _ nodeFactory = &objFactory{}

//nolint:gochecknoglobals
var (
	factory = newObjFactory()
)

// newTree creates a new tree.
func newTree() *tree {
	return &tree{
		version: 0,
		root:    nil,
		size:    0,
	}
}

// objFactory implements nodeFactory interface.
type objFactory struct{}

// newObjFactory creates a new objFactory.
func newObjFactory() nodeFactory {
	return &objFactory{}
}

// Simple obj factory implementation.
func (f *objFactory) newNode4() *NodeRef {
	return &NodeRef{
		kind: Node4Kind,
		ref:  unsafe.Pointer(new(Node4)), //#nosec:G103
	}
}

// newNode16 creates a new Node16 as a NodeRef.
func (f *objFactory) newNode16() *NodeRef {
	return &NodeRef{
		kind: Node16Kind,
		ref:  unsafe.Pointer(new(Node16)), //#nosec:G103
	}
}

// newNode48 creates a new Node48 as a NodeRef.
func (f *objFactory) newNode48() *NodeRef {
	return &NodeRef{
		kind: Node48Kind,
		ref:  unsafe.Pointer(new(Node48)), //#nosec:G103
	}
}

// newNode256 creates a new Node256 as a NodeRef.
func (f *objFactory) newNode256() *NodeRef {
	return &NodeRef{
		kind: Node256Kind,
		ref:  unsafe.Pointer(new(Node256)), //#nosec:G103
	}
}

// newLeaf creates a new Leaf Node as a NodeRef.
// It clones the key to avoid any source key mutation.
func (f *objFactory) newLeaf(key Key, value interface{}) *NodeRef {
	keyClone := make(Key, len(key))
	copy(keyClone, key)

	return &NodeRef{
		kind: LeafKind,
		ref: unsafe.Pointer(&Leaf{ //#nosec:G103
			key:   keyClone,
			value: value,
		}),
	}
}
