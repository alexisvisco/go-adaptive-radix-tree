package art

// Node4 represents a Node with 4 children.
type Node4 struct {
	Node
	children [node4Max + 1]*NodeRef // pointers to the child nodes, +1 is for the zero byte child
	keys     [node4Max]byte         // keys for the children
	present  [node4Max]byte         // present bits for the keys
}

// minimum returns the minimum Leaf Node.
func (n *Node4) minimum() *Leaf {
	return nodeMinimum(n.children[:])
}

// maximum returns the maximum Leaf Node.
func (n *Node4) maximum() *Leaf {
	return nodeMaximum(n.children[:n.childrenLen])
}

// index returns the index of the given character.
func (n *Node4) index(kc keyChar) int {
	if kc.invalid {
		return node4Max
	}

	return findIndex(n.keys[:n.childrenLen], kc.ch)
}

// childAt returns the child at the given index.
func (n *Node4) childAt(idx int) **NodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *Node4) allChildren() []*NodeRef {
	return n.children[:]
}

// hasCapacityForChild returns true if the Node has room for more children.
func (n *Node4) hasCapacityForChild() bool {
	return n.childrenLen < node4Max
}

// grow converts the Node4 into the Node16.
func (n *Node4) grow() *NodeRef {
	an16 := factory.newNode16()
	n16 := an16.node16()

	copyNode(&n16.Node, &n.Node)
	n16.children[node16Max] = n.children[node4Max] // copy zero byte child

	for i := 0; i < int(n.childrenLen); i++ {
		// skip if the key is not present
		if n.present[i] == 0 {
			continue
		}

		// copy elements from n4 to n16 to the last position
		n16.insertChildAt(i, n.keys[i], n.children[i])
	}

	return an16
}

// isReadyToShrink returns true if the Node is under-utilized and ready to shrink.
func (n *Node4) isReadyToShrink() bool {
	// we have to return the number of children for the current Node(Node4) as
	// `Node.numChildren` plus one if zero Node is not nil.
	// For all higher nodes(16/48/256) we simply copy zero Node to a smaller Node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := n.childrenLen
	if n.children[node4Max] != nil {
		numChildren++
	}

	return numChildren < node4Min
}

// shrink converts the Node4 into the Leaf Node or a Node with fewer children.
func (n *Node4) shrink() *NodeRef {
	// Select the non-nil child Node
	var nonNilChild *NodeRef
	if n.children[0] != nil {
		nonNilChild = n.children[0]
	} else {
		nonNilChild = n.children[node4Max]
	}

	// if the only child is a LeafKind Node, return it
	if nonNilChild.isLeaf() {
		return nonNilChild
	}

	// update the prefix of the child Node
	n.adjustPrefix(nonNilChild.node())

	return nonNilChild
}

// adjustPrefix handles prefix adjustments for a non-LeafKind child.
func (n *Node4) adjustPrefix(childNode *Node) {
	nodePrefLen := int(n.prefixLen)

	// at this point, the Node has only one child
	// copy the key part of the current Node as prefix
	if nodePrefLen < maxPrefixLen {
		n.prefix[nodePrefLen] = n.keys[0]
		nodePrefLen++
	}

	// copy the part of child prefix that fits into the current Node
	if nodePrefLen < maxPrefixLen {
		childPrefLen := minInt(int(childNode.prefixLen), maxPrefixLen-nodePrefLen)
		copy(n.prefix[nodePrefLen:], childNode.prefix[:childPrefLen])
		nodePrefLen += childPrefLen
	}

	// copy the part of the current Node prefix that fits into the child Node
	prefixLen := minInt(nodePrefLen, maxPrefixLen)
	copy(childNode.prefix[:], n.prefix[:prefixLen])
	childNode.prefixLen += n.prefixLen + 1
}

// addChild adds a new child to the Node.
func (n *Node4) addChild(kc keyChar, child *NodeRef) {
	pos := n.findInsertPos(kc)
	n.makeRoom(pos)
	n.insertChildAt(pos, kc.ch, child)
}

// find the insert position for the new child.
func (n *Node4) findInsertPos(kc keyChar) int {
	if kc.invalid {
		return node4Max
	}

	numChildren := int(n.childrenLen)
	for i := 0; i < numChildren; i++ {
		if n.keys[i] > kc.ch {
			return i
		}
	}

	return numChildren
}

// makeRoom creates space for the new child by shifting the elements to the right.
func (n *Node4) makeRoom(pos int) {
	if pos < 0 || pos >= int(n.childrenLen) {
		return
	}

	for i := int(n.childrenLen); i > pos; i-- {
		n.keys[i] = n.keys[i-1]
		n.present[i] = n.present[i-1]
		n.children[i] = n.children[i-1]
	}
}

// insertChildAt inserts the child at the given position.
func (n *Node4) insertChildAt(pos int, ch byte, child *NodeRef) {
	if pos == node4Max {
		n.children[pos] = child
	} else {
		n.keys[pos] = ch
		n.present[pos] = 1
		n.children[pos] = child
		n.childrenLen++
	}
}

// deleteChild deletes the child from the Node.
func (n *Node4) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node4Max] = nil
	} else if idx := n.index(kc); idx >= 0 {
		n.deleteChildAt(idx)
		n.clearLastElement()
	}

	// we have to return the number of children for the current Node(Node4) as
	// `n.numChildren` plus one if null Node is not nil.
	// `Shrink` method can be invoked after this method,
	// `Shrink` can convert this Node into a LeafKind Node type.
	// For all higher nodes(16/48/256) we simply copy null Node to a smaller Node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := int(n.childrenLen)
	if n.children[node4Max] != nil {
		numChildren++
	}

	return numChildren
}

// deleteChildAt deletes the child at the given index
// by shifting the elements to the left to overwrite deleted child.
func (n *Node4) deleteChildAt(idx int) {
	for i := idx; i < int(n.childrenLen) && i+1 < node4Max; i++ {
		n.keys[i] = n.keys[i+1]
		n.present[i] = n.present[i+1]
		n.children[i] = n.children[i+1]
	}

	n.childrenLen--
}

// clearLastElement clears the last element in the Node.
func (n *Node4) clearLastElement() {
	lastIdx := int(n.childrenLen)
	n.keys[lastIdx] = 0
	n.present[lastIdx] = 0
	n.children[lastIdx] = nil
}
