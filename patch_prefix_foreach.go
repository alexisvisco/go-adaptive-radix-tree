package art

// ForEachPrefixWithSeparator efficiently iterates over all keys with the given prefix.
// maxDepth limits how many separators deep from the prefix to traverse (-1 for unlimited)
// reverse determines whether to traverse in reverse order
func (tr *tree) ForEachPrefixWithSeparator(
	keyPrefix Key,
	callback Callback,
	countSeparator func(Key, Key) int,
	maxDepth int,
	reverse bool,
) {
	if tr.root == nil || len(keyPrefix) == 0 {
		return
	}

	// Navigate to the prefix node first (similar to Search)
	keyOffset := 0
	current := tr.root

	// Keep traversing down the tree until we either:
	// 1. Reach a leaf node
	// 2. Exhaust the prefix
	// 3. Find a mismatch in the prefix
	for current != nil && keyOffset < len(keyPrefix) {
		if current.isLeaf() {
			// Found a leaf - check if it's a match
			leaf := current.Leaf()
			if leaf.PrefixMatch(keyPrefix) {
				// Check if it's within the depth limit
				if maxDepth >= 0 && countSeparator(keyPrefix, leaf.key) > maxDepth {
					return // Exceeds depth limit
				}

				if !callback(current) {
					return // Stop traversal if callback returns false
				}
			}
			return // End traversal, no more nodes to check
		}

		// Check node prefix
		curNode := current.node()
		if curNode.prefixLen > 0 {
			prefixLen := current.match(keyPrefix, keyOffset)
			if prefixLen != minInt(int(curNode.prefixLen), maxPrefixLen) {
				return // Prefix mismatch, no matching keys
			}

			keyOffset += int(curNode.prefixLen)
		}

		// If we've exhausted the keyPrefix, this is where we should start our traversal
		if keyOffset >= len(keyPrefix) {
			break
		}

		// Find the child that matches the next character of the prefix
		next := current.findChildByKey(keyPrefix, keyOffset)
		if *next != nil {
			current = *next
			keyOffset++
		} else {
			return // No matching child, so no keys with this prefix
		}
	}

	// If we get here, we've found the node that corresponds to the prefix
	// Now traverse the subtree rooted at this node
	if current != nil {
		tr.traversePrefixSubtreeV2(current, keyPrefix, callback, countSeparator, maxDepth, reverse)
	}
}

// traversePrefixSubtreeV2 traverses all nodes in the subtree rooted at the given node,
// respecting the depth limit based on the separator.
func (tr *tree) traversePrefixSubtreeV2(
	current *NodeRef,
	keyPrefix Key,
	callback Callback,
	countSeparator func(Key, Key) int,
	maxDepth int,
	reverse bool,
) {
	if current == nil {
		return
	}

	// For leaf nodes, call the callback if within depth limit
	if current.isLeaf() {
		leaf := current.Leaf()

		// Skip if the node exceeds our depth limit
		if maxDepth >= 0 && countSeparator(keyPrefix, leaf.key) > maxDepth {
			return
		}

		if !callback(current) {
			return // Stop traversal if callback returns false
		}
		return
	}

	// Recursively traverse children
	nextFn := newTraverseFunc(current, reverse)
	children := toNode(current).allChildren()

	for {
		idx, hasMore := nextFn()
		if !hasMore {
			break
		}

		if child := children[idx]; child != nil {
			tr.traversePrefixSubtreeV2(child, keyPrefix, callback, countSeparator, maxDepth, reverse)
		}
	}
}
