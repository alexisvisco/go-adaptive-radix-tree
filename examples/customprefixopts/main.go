package main

import (
	"bytes"
	"fmt"

	art "github.com/alexisvisco/go-adaptive-radix-tree/v2"
)

func main() {

	tree := art.New()

	terms := []string{
		"/etc",
		"/etc/hosts",
		"/var",
		"/var/log/",
		"/var/log/syslog",
		"/home",
		"/home/user",
	}

	for _, term := range terms {
		tree.Insert(art.Key(term), 0)
	}

	ForEachPrefixWithDepth(tree, art.Key("/"), func(node art.NodeKV) (cont bool) {
		if node.Kind() == art.LeafKind {
			fmt.Println("node:", string(node.Key()))
		}
		return true
	}, art.TraverseLeaf, 0)

}

func ForEachPrefixWithDepth(tree art.Tree, key art.Key, callback art.Callback, opts int, maxDepth int) {
	opts &= art.TraverseLeaf | art.TraverseReverse // keep only LeafKind and reverse options

	tree.ForEach(func(n art.NodeKV) bool {
		current, ok := n.(*art.NodeRef)
		if !ok {
			return false
		}

		if leaf := current.Leaf(); leaf.PrefixMatch(key) {

			depths := numberOfSlashesFromPrefix(key, current.Key())
			if depths > maxDepth {
				// Skip this node if it exceeds the depth limit, but continue searching
				return true
			}

			return callback(current)
		}

		return true
	}, opts)

}

func numberOfSlashesFromPrefix(key art.Key, current art.Key) int {
	if len(key) == 1 && key[0] == '/' && len(current) > 1 {
		// For root path, we count all slashes in the rest of the path
		count := 0
		for i := 1; i < len(current); i++ {
			if current[i] == '/' {
				count++
			}
		}
		return count
	}

	// If they're exactly the same path, return 0
	if bytes.Equal(key, current) {
		return 0
	}

	// The case with trailing slash in the key
	if len(key) > 0 && key[len(key)-1] == '/' {
		// Get the remainder after the key
		remainder := current[len(key):]

		// Count slashes in the remainder
		count := 0
		for i := 0; i < len(remainder); i++ {
			if remainder[i] == '/' {
				count++
			}
		}
		return count
	}

	// Get the path after the prefix
	remainder := current[len(key):]

	// Count slashes in the remainder
	count := 0
	for i := 0; i < len(remainder); i++ {
		if remainder[i] == '/' {
			count++
		}
	}

	return count
}
