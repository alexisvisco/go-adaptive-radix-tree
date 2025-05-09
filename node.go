package art

// prefix used in the Node to store the key prefix.
// it is used to improve Leaf key comparison performance.
type prefix [maxPrefixLen]byte

// Node is the base struct for all Node types.
// it contains the common fields for all nodeX types.
type Node struct {
	prefix      prefix // prefix of the Node
	prefixLen   uint16 // length of the prefix
	childrenLen uint16 // number of children in the Node4, Node16, Node48, Node256
}

// replaceRef is used to replace Node in-place by updating the reference.
func replaceRef(oldNode **NodeRef, newNode *NodeRef) {
	*oldNode = newNode
}

// replaceNode is used to replace Node in-place by updating the Node.
func replaceNode(oldNode *NodeRef, newNode *NodeRef) {
	*oldNode = *newNode
}
