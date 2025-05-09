package art

// NodeKV with 256 children.
type Node256 struct {
	Node
	children [node256Max + 1]*NodeRef // +1 is for the zero byte child
}

// minimum returns the minimum Leaf Node.
func (n *Node256) minimum() *Leaf {
	return nodeMinimum(n.children[:])
}

// maximum returns the maximum Leaf Node.
func (n *Node256) maximum() *Leaf {
	return nodeMaximum(n.children[:node256Max])
}

// index returns the index of the child with the given key.
func (n *Node256) index(kc keyChar) int {
	if kc.invalid { // handle zero byte in the key
		return node256Max
	}

	return int(kc.ch)
}

// childAt returns the child at the given index.
func (n *Node256) childAt(idx int) **NodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *Node256) allChildren() []*NodeRef {
	return n.children[:]
}

// addChild adds a new child to the Node.
func (n *Node256) addChild(kc keyChar, child *NodeRef) {
	if kc.invalid {
		// handle zero byte in the key
		n.children[node256Max] = child
	} else {
		// insert new child
		n.children[kc.ch] = child
		n.childrenLen++
	}
}

// hasCapacityForChild for Node256 always returns true.
func (n *Node256) hasCapacityForChild() bool {
	return true
}

// grow for Node256 always returns nil,
// because Node256 has the maximum capacity.
func (n *Node256) grow() *NodeRef {
	return nil
}

// isReadyToShrink returns true if the Node can be shrunk.
func (n *Node256) isReadyToShrink() bool {
	return n.childrenLen < node256Min
}

// shrink shrinks the Node to a smaller type.
func (n *Node256) shrink() *NodeRef {
	an48 := factory.newNode48()
	n48 := an48.node48()

	copyNode(&n48.Node, &n.Node)
	n48.children[node48Min] = n.children[node256Max] // copy zero byte child

	for numChildren, i := 0, 0; i < node256Max; i++ {
		if n.children[i] == nil {
			continue // skip if the child is nil
		}
		// copy elements from n256 to n48 to the last position
		n48.insertChildAt(numChildren, byte(i), n.children[i])

		numChildren++
	}

	return an48
}

// deleteChild removes the child with the given key.
func (n *Node256) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node256Max] = nil
	} else if idx := n.index(kc); n.children[idx] != nil {
		// clear the child at the given index
		n.children[idx] = nil
		n.childrenLen--
	}

	return int(n.childrenLen)
}
