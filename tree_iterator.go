package art

import "errors"

// state represents the iteration state during tree traversal.
type state struct {
	items []*iteratorContext
}

// push adds a new iterator context to the state.
func (s *state) push(ctx *iteratorContext) {
	s.items = append(s.items, ctx)
}

// current returns the current iterator context and a flag indicating if there is any.
func (s *state) current() (*iteratorContext, bool) {
	if len(s.items) == 0 {
		return nil, false
	}

	return s.items[len(s.items)-1], true
}

// discard removes the last iterator context from the state.
func (s *state) discard() {
	if len(s.items) == 0 {
		return
	}

	s.items = s.items[:len(s.items)-1]
}

// iteratorContext represents the context of the tree iterator for one Node.
type iteratorContext struct {
	nextChildFn traverseFunc
	children    []*NodeRef
}

// newIteratorContext creates a new iterator context for the given Node.
func newIteratorContext(nr *NodeRef, reverse bool) *iteratorContext {
	return &iteratorContext{
		nextChildFn: newTraverseFunc(nr, reverse),
		children:    toNode(nr).allChildren(),
	}
}

// next returns the next Node reference and a flag indicating if there are more nodes.
func (ic *iteratorContext) next() (*NodeRef, bool) {
	for {
		idx, ok := ic.nextChildFn()
		if !ok {
			break
		}

		if child := ic.children[idx]; child != nil {
			return child, true
		}
	}

	return nil, false
}

// iterator is a struct for tree traversal iteration.
type iterator struct {
	version  int      // tree version at the time of iterator creation
	tree     *tree    // tree to iterate
	state    *state   // iteration state
	nextNode *NodeRef // next Node to iterate
	reverse  bool     // indicates if the iteration is in reverse order
}

// assert that iterator implements the Iterator interface.
var _ Iterator = (*iterator)(nil)

// newTreeIterator creates a new tree iterator.
func newTreeIterator(tr *tree, opts traverseOpts) Iterator {
	state := &state{}
	state.push(newIteratorContext(tr.root, opts.hasReverse()))

	it := &iterator{
		version:  tr.version,
		tree:     tr,
		nextNode: tr.root,
		state:    state,
		reverse:  opts.hasReverse(),
	}

	if opts&TraverseAll == TraverseAll {
		return it
	}

	bit := &bufferedIterator{
		opts: opts,
		it:   it,
	}

	// peek the first Node or LeafKind
	bit.peek()

	return bit
}

// hasConcurrentModification checks if the tree has been modified concurrently.
func (it *iterator) hasConcurrentModification() bool {
	return it.version != it.tree.version
}

// HasNext returns true if there are more nodes to iterate.
func (it *iterator) HasNext() bool {
	return it.nextNode != nil
}

// Next returns the next Node and an error if any.
// It returns ErrNoMoreNodes if there are no more nodes to iterate.
// It returns ErrConcurrentModification if the tree has been modified concurrently.
func (it *iterator) Next() (NodeKV, error) {
	if !it.HasNext() {
		return nil, ErrNoMoreNodes
	}

	if it.hasConcurrentModification() {
		return nil, ErrConcurrentModification
	}

	current := it.nextNode
	it.next()

	return current, nil
}

// next moves the iterator to the next Node.
func (it *iterator) next() {
	for {
		ctx, ok := it.state.current()
		if !ok {
			it.nextNode = nil // no more nodes to iterate

			return
		}

		nextNode, hasMore := ctx.next()
		if hasMore {
			it.nextNode = nextNode
			it.state.push(newIteratorContext(nextNode, it.reverse))

			return
		}

		it.state.discard() // discard the current context as exhausted
	}
}

// BufferedIterator implements HasNext and Next methods for buffered iteration.
// It allows to iterate over Leaf or non-LeafKind nodes only.
type bufferedIterator struct {
	opts     traverseOpts
	it       Iterator
	nextNode NodeKV
	nextErr  error
}

// HasNext returns true if there are more nodes to iterate.
func (bit *bufferedIterator) HasNext() bool {
	return bit.nextNode != nil
}

// Next returns the next Node or LeafKind Node and an error if any.
// ErrNoMoreNodes is returned if there are no more nodes to iterate.
// ErrConcurrentModification is returned if the tree has been modified concurrently.
func (bit *bufferedIterator) Next() (NodeKV, error) {
	current := bit.nextNode

	if !bit.HasNext() {
		return nil, bit.nextErr
	}

	bit.peek()

	// ErrConcurrentModification should be returned immediately.
	// ErrNoMoreNodes will be return on the next call.
	if errors.Is(bit.nextErr, ErrConcurrentModification) {
		return nil, bit.nextErr
	}

	return current, nil
}

// hasLeafIterator checks if the iterator is for Leaf nodes.
func (bit *bufferedIterator) hasLeafIterator() bool {
	return bit.opts&TraverseLeaf == TraverseLeaf
}

// hasNodeIterator checks if the iterator is for non-LeafKind nodes.
func (bit *bufferedIterator) hasNodeIterator() bool {
	return bit.opts&TraverseNode == TraverseNode
}

// peek looks for the next Node or Leaf Node to iterate.
func (bit *bufferedIterator) peek() {
	for {
		bit.nextNode, bit.nextErr = bit.it.Next()
		if bit.nextErr != nil {
			return
		}

		if bit.matchesFilter() {
			return
		}
	}
}

// matchesFilter checks if the next Node matches the iterator filter.
func (bit *bufferedIterator) matchesFilter() bool {
	// check if the iterator is looking for LeafKind nodes
	if bit.hasLeafIterator() && bit.nextNode.Kind() == LeafKind {
		return true
	}

	// check if the iterator is looking for non-LeafKind nodes
	if bit.hasNodeIterator() && bit.nextNode.Kind() != LeafKind {
		return true
	}

	return false
}
