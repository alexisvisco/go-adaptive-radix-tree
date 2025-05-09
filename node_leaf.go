package art

import "bytes"

// LeafKind Node stores the key-value pair.
type Leaf struct {
	key   Key
	value interface{}
}

// Match returns true if the Leaf Node's key matches the given key.
func (l *Leaf) Match(key Key) bool {
	return len(l.key) == len(key) && bytes.Equal(l.key, key)
}

// PrefixMatch returns true if the LeafKind Node's key has the given key as a prefix.
func (l *Leaf) PrefixMatch(key Key) bool {
	if key == nil || len(l.key) < len(key) {
		return false
	}

	return bytes.Equal(l.key[:len(key)], key)
}
