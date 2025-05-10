package art

import "errors"

// NodeKV types.
const (
	LeafKind    Kind = 0
	Node4Kind   Kind = 1
	Node16Kind  Kind = 2
	Node48Kind  Kind = 3
	Node256Kind Kind = 4
)

// Traverse Options.
const (
	// Iterate only over LeafKind nodes.
	TraverseLeaf = 1

	// Iterate only over non-LeafKind nodes.
	TraverseNode = 2

	// Iterate over all nodes in the tree.
	TraverseAll = TraverseLeaf | TraverseNode

	// Iterate in reverse order.
	TraverseReverse = 4
)

// These errors can be returned when iteration over the tree.
var (
	ErrConcurrentModification = errors.New("concurrent modification has been detected")
	ErrNoMoreNodes            = errors.New("there are no more nodes in the tree")
)

// Kind is a Node type.
type Kind int

// String returns string representation of the Kind value.
func (k Kind) String() string {
	return []string{"LeafKind", "Node4Kind", "Node16Kind", "Node48Kind", "Node256Kind"}[k]
}

// Key represents the type used for keys in the Adaptive Radix Tree.
// It can consist of any byte sequence, including Unicode characters and null bytes.
type Key []byte

// Value is an interface representing the value type stored in the tree.
// Any type of data can be stored as a Value.
type Value interface{}

// Callback defines the function type used during tree traversal.
// It is invoked for each Node visited in the traversal.
// If the callback function returns false, the iteration is terminated early.
type Callback func(node NodeKV) (cont bool)

// NodeKV represents a Node within the Adaptive Radix Tree.
type NodeKV interface {
	// Kind returns the type of the Node, distinguishing between LeafKind and internal nodes.
	Kind() Kind

	// Key returns the key associated with a LeafKind Node.
	// This method should only be called on LeafKind nodes.
	// Calling this on a non-LeafKind Node will return nil.
	Key() Key

	// Value returns the value stored in a LeafKind Node.
	// This method should only be called on LeafKind nodes.
	// Calling this on a non-LeafKind Node will return nil.
	Value() Value
}

// Iterator provides a mechanism to traverse nodes in key order within the tree.
type Iterator interface {
	// HasNext returns true if there are more nodes to visit during the iteration.
	// Use this method to check for remaining nodes before calling Next.
	HasNext() bool

	// Next returns the next Node in the iteration and advances the iterator's position.
	// If the iteration has no more nodes, it returns ErrNoMoreNodes error.
	// Ensure you call HasNext before invoking Next to avoid errors.
	// If the tree has been structurally modified since the iterator was created,
	// it returns an ErrConcurrentModification error.
	Next() (NodeKV, error)
}

// Tree is an Adaptive Radix Tree interface.
type Tree interface {
	// Insert adds a new key-value pair into the tree.
	// If the key already exists in the tree, it updates its value and returns the old value along with true.
	// If the key is new, it returns nil and false.
	Insert(key Key, value Value) (oldValue Value, updated bool)

	// Delete removes the specified key and its associated value from the tree.
	// If the key is found and deleted, it returns the removed value and true.
	// If the key does not exist, it returns nil and false.
	Delete(key Key) (value Value, deleted bool)

	// Search retrieves the value associated with the specified key in the tree.
	// If the key exists, it returns the value and true.
	// If the key does not exist, it returns nil and false.
	Search(key Key) (value Value, found bool)

	// ForEach iterates over all the nodes in the tree, invoking a provided callback function for each Node.
	// By default, it processes LeafKind nodes in ascending order.
	// The iteration can be customized using options:
	// - Pass TraverseReverse to iterate over nodes in descending order.
	// The iteration stops if the callback function returns false, allowing for early termination.
	ForEach(callback Callback, options ...int)

	// ForEachPrefix iterates over all LeafKind nodes whose keys start with the specified keyPrefix,
	// invoking a provided callback function for each matching Node.
	// By default, the iteration processes nodes in ascending order.
	// Use the TraverseReverse option to iterate over nodes in descending order.
	// Iteration stops if the callback function returns false, allowing for early termination.
	ForEachPrefix(keyPrefix Key, callback Callback, options ...int)

	// Iterator returns an iterator for traversing LeafKind nodes in the tree.
	// By default, the iteration occurs in ascending order.
	// To traverse nodes in reverse (descending) order, pass the TraverseReverse option.
	Iterator(options ...int) Iterator

	// Minimum retrieves the LeafKind Node with the smallest key in the tree.
	// If such a LeafKind is found, it returns its value and true.
	// If the tree is empty, it returns nil and false.
	Minimum() (Value, bool)

	// Maximum retrieves the LeafKind Node with the largest key in the tree.
	// If such a LeafKind is found, it returns its value and true.
	// If the tree is empty, it returns nil and false.
	Maximum() (Value, bool)

	// Size returns the number of key-value pairs stored in the tree.
	Size() int

	ForEachPrefixWithSeparator(
		keyPrefix Key,
		callback Callback,
		countSeparator func(Key, Key) int,
		maxDepth int,
		reverse bool,
	)
}

// New creates a new adaptive radix tree.
func New() Tree {
	return newTree()
}
