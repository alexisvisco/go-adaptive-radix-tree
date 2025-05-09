package art

// deleteRecursively removes a Node associated with the key from the tree.
func (tr *tree) deleteRecursively(nrp **NodeRef, key Key, keyOffset int) (Value, treeOpResult) {
	if tr == nil || *nrp == nil || len(key) == 0 {
		return nil, treeOpNoChange
	}

	nr := *nrp
	if nr.isLeaf() {
		return tr.handleLeafDeletion(nrp, key)
	}

	return tr.handleInternalNodeDeletion(nr, key, keyOffset)
}

// handleLeafDeletion removes a Leaf Node associated with the key from the tree.
func (tr *tree) handleLeafDeletion(nrp **NodeRef, key Key) (Value, treeOpResult) {
	if leaf := (*nrp).Leaf(); leaf.Match(key) {
		replaceRef(nrp, nil)

		return leaf.value, treeOpDeleted
	}

	return nil, treeOpNoChange
}

// handleInternalNodeDeletion removes a Node associated with the key from the Node.
func (tr *tree) handleInternalNodeDeletion(nr *NodeRef, key Key, keyOffset int) (Value, treeOpResult) {
	n := nr.node()

	if n.prefixLen > 0 {
		if mismatchIdx := nr.match(key, keyOffset); mismatchIdx != minInt(int(n.prefixLen), maxPrefixLen) {
			return nil, treeOpNoChange
		}

		keyOffset += int(n.prefixLen)
	}

	next := nr.findChildByKey(key, keyOffset)
	if *next == nil {
		return nil, treeOpNoChange
	}

	if (*next).isLeaf() {
		return tr.handleDeletionInChild(nr, *next, key, keyOffset)
	}

	return tr.deleteRecursively(next, key, keyOffset+1)
}

// handleDeletionInChild removes a Leaf Node from the child Node.
func (tr *tree) handleDeletionInChild(curNR, nextNR *NodeRef, key Key, keyOffset int) (Value, treeOpResult) {
	leaf := (*nextNR).Leaf()
	if !leaf.Match(key) {
		return nil, treeOpNoChange
	}

	curNR.deleteChild(key.charAt(keyOffset))

	return leaf.value, treeOpDeleted
}
