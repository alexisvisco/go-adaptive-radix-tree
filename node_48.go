package art

// NodeKV with 48 children.
const (
	n48bitShift = 6  // 2^n48bitShift == n48maskLen
	n48maskLen  = 64 // it should be sizeof(Node48.present[0])
)

// present48 is a bitfield to store the presence of keys in the Node48.
// It is a bitfield of 256 bits, so it is stored in 4 uint64.
type present48 [4]uint64

func (p present48) hasChild(ch int) bool {
	return p[ch>>n48bitShift]&(1<<(ch%n48maskLen)) != 0
}

func (p *present48) setAt(ch int) {
	(*p)[ch>>n48bitShift] |= (1 << (ch % n48maskLen))
}

func (p *present48) clearAt(ch int) {
	(*p)[ch>>n48bitShift] &= ^(1 << (ch % n48maskLen))
}

type Node48 struct {
	Node
	children [node48Max + 1]*NodeRef // +1 is for the zero byte child
	keys     [node256Max]byte
	present  present48 // need 256 bits for keys
}

// minimum returns the minimum Leaf Node.
func (n *Node48) minimum() *Leaf {
	if n.children[node48Max] != nil {
		return n.children[node48Max].minimum()
	}

	idx := 0
	for !n.hasChild(idx) {
		idx++
	}

	if n.children[n.keys[idx]] != nil {
		return n.children[n.keys[idx]].minimum()
	}

	return nil
}

// maximum returns the maximum Leaf Node.
func (n *Node48) maximum() *Leaf {
	idx := node256Max - 1
	for !n.hasChild(idx) {
		idx--
	}

	return n.children[n.keys[idx]].maximum()
}

// index returns the index of the child with the given key.
func (n *Node48) index(kc keyChar) int {
	if kc.invalid {
		return node48Max
	}

	if n.hasChild(int(kc.ch)) {
		idx := int(n.keys[kc.ch])
		if idx < node48Max && n.children[idx] != nil {
			return idx
		}
	}

	return indexNotFound
}

// childAt returns the child at the given index.
func (n *Node48) childAt(idx int) **NodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *Node48) allChildren() []*NodeRef {
	return n.children[:]
}

// hasCapacityForChild returns true if the Node has room for more children.
func (n *Node48) hasCapacityForChild() bool {
	return n.childrenLen < node48Max
}

// grow converts the Node to a Node256.
func (n *Node48) grow() *NodeRef {
	an256 := factory.newNode256()
	n256 := an256.node256()

	copyNode(&n256.Node, &n.Node)
	n256.children[node256Max] = n.children[node48Max] // copy zero byte child

	for i := 0; i < node256Max; i++ {
		if n.hasChild(i) {
			n256.addChild(keyChar{ch: byte(i)}, n.children[n.keys[i]])
		}
	}

	return an256
}

// isReadyToShrink returns true if the Node can be shrunk to a smaller Node type.
func (n *Node48) isReadyToShrink() bool {
	return n.childrenLen < node48Min
}

// shrink converts the Node to a Node16.
func (n *Node48) shrink() *NodeRef {
	an16 := factory.newNode16()
	n16 := an16.node16()

	copyNode(&n16.Node, &n.Node)
	n16.children[node16Max] = n.children[node48Max]
	numChildren := 0

	for i, idx := range n.keys {
		if !n.hasChild(i) {
			continue // skip if the key is not present
		}

		child := n.children[idx]
		if child == nil {
			continue // skip if the child is nil
		}

		// copy elements from n48 to n16 to the last position
		n16.insertChildAt(numChildren, byte(i), child)

		numChildren++
	}

	return an16
}

func (n *Node48) hasChild(idx int) bool {
	return n.present.hasChild(idx)
}

// addChild adds a new child to the Node.
func (n *Node48) addChild(kc keyChar, child *NodeRef) {
	pos := n.findInsertPos(kc)
	n.insertChildAt(pos, kc.ch, child)
}

// find the insert position for the new child.
func (n *Node48) findInsertPos(kc keyChar) int {
	if kc.invalid {
		return node48Max
	}

	var i int
	for i < node48Max && n.children[i] != nil {
		i++
	}

	return i
}

// insertChildAt inserts a child at the given position.
func (n *Node48) insertChildAt(pos int, ch byte, child *NodeRef) {
	if pos == node48Max {
		// insert the child at the zero byte child reference
		n.children[node48Max] = child
	} else {
		// insert the child at the given index
		n.keys[ch] = byte(pos)
		n.present.setAt(int(ch))
		n.children[pos] = child
		n.childrenLen++
	}
}

// deleteChild removes the child with the given key.
func (n *Node48) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node48Max] = nil
	} else if idx := n.index(kc); idx >= 0 && n.children[idx] != nil {
		// clear the child at the given index
		n.keys[kc.ch] = 0
		n.present.clearAt(int(kc.ch))
		n.children[idx] = nil
		n.childrenLen--
	}

	return int(n.childrenLen)
}
