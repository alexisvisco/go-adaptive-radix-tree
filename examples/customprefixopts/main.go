package main

import (
	"bytes"
	"fmt"
	"path/filepath"

	art "github.com/alexisvisco/go-adaptive-radix-tree/v2"
)

func main() {

	tree := art.New()

	terms := []string{
		"/",
		"/etc",
		"/etc/hosts",
		"/var",
		"/var/log",
		"/var/b/",
		"/var/log/syslog",
		"/home",
		"/home/user",
	}

	// check number of slashes after /var
	for _, term := range terms {
		count := numberOfSlashesFromPrefix([]byte("/var"), []byte(term))
		fmt.Println("number of slashes after /var in", term, ":", count)
	}

	for _, term := range terms {
		tree.Insert(art.Key(term), 0)
	}

	tree.ForEachPrefixWithSeparator(art.Key("/"), func(node art.NodeKV) (cont bool) {
		if node.Kind() == art.LeafKind {
			println("node:", string(node.Key()))
		}
		return true
	}, numberOfSlashesFromPrefix, 2, false)

}

// numberOfSlashesFromPrefix counts how many slashes are in 'current' after the 'key' prefix.
func numberOfSlashesFromPrefix(prefix, current art.Key) int {
	// Remove trailing slash from prefix if present (unless it's root)
	cleanPrefix := prefix
	if len(prefix) > 1 && prefix[len(prefix)-1] == filepath.Separator {
		cleanPrefix = prefix[:len(prefix)-1]
	}

	cleanCurrent := current
	if len(current) > 1 && current[len(current)-1] == filepath.Separator {
		cleanCurrent = current[:len(current)-1]
	}

	if bytes.Equal(cleanPrefix, cleanCurrent) {
		return 0
	}

	if !bytes.HasPrefix(cleanCurrent, cleanPrefix) {
		return -1
	}

	trimmed := bytes.TrimPrefix(cleanCurrent, cleanPrefix)
	paths := bytes.Split(trimmed, []byte{filepath.Separator})
	count := len(paths)
	for _, path := range paths {
		if bytes.Equal(path, []byte{}) {
			count--
		}
	}

	return count
}
