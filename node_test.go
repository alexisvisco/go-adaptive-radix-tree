package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic properties and behavior of each Node kind.
func TestNodeKindProperties(t *testing.T) {
	t.Parallel()

	// Define a Table of NodeKV Types to Test
	nodeTests := []struct {
		name string
		node *NodeRef
		kind Kind
	}{
		{"Node4Kind Test", factory.newNode4(), Node4Kind},
		{"Node16Kind Test", factory.newNode16(), Node16Kind},
		{"Node48Kind Test", factory.newNode48(), Node48Kind},
		{"Node256Kind Test", factory.newNode256(), Node256Kind},
	}

	// Run NodeKV Kind Tests
	for _, tt := range nodeTests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, tt.node)
			assert.Equal(t, tt.kind, tt.node.kind)
		})
	}

	// Test LeafKind NodeKV
	t.Run("LeafKind NodeKV Test", func(t *testing.T) {
		t.Parallel()

		leaf := factory.newLeaf(Key("key"), "value")
		assert.NotNil(t, leaf)
		assert.Equal(t, LeafKind, leaf.kind)
		assert.Equal(t, "LeafKind", leaf.kind.String())
		assert.Equal(t, Key("key"), leaf.Key())

		val, ok := leaf.Value().(string)
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})
}

func TestUnknownNode(t *testing.T) {
	t.Parallel()

	unknownNode := &NodeRef{kind: Kind(0xFF)}
	assert.Nil(t, unknownNode.maximum())
	assert.Nil(t, unknownNode.minimum())
}

func TestLeafFunctionality(t *testing.T) {
	t.Parallel()

	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, LeafKind, leaf.kind)

	assert.False(t, leaf.Leaf().Match(Key("unknown-key")))

	// Ensure we cannot shrink/grow LeafKind Node
	assert.Nil(t, toNode(leaf).shrink())
	assert.Nil(t, toNode(leaf).grow())
}

// Test matching behavior of LeafKind nodes.
func TestLeafMatchBehavior(t *testing.T) {
	t.Parallel()

	leaf := factory.newLeaf(Key("key"), "value")

	assert.False(t, leaf.Leaf().Match(Key("unknown-key")))
	assert.False(t, leaf.Leaf().Match(nil))
	assert.True(t, leaf.Leaf().Match(Key("key")))

	assert.False(t, leaf.Leaf().PrefixMatch(Key("unknown-key")))
	assert.False(t, leaf.Leaf().PrefixMatch(nil))
	assert.True(t, leaf.Leaf().PrefixMatch(Key("ke")))
}

// Check the setting of prefixes.
func TestNodePrefixSetting(t *testing.T) {
	t.Parallel()

	n4 := factory.newNode4()
	nn := n4.node()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	n4.setPrefix(key, 2)
	assert.Equal(t, 2, int(nn.prefixLen))
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])

	n4.setPrefix(key, maxPrefixLen)
	assert.Equal(t, maxPrefixLen, int(nn.prefixLen))
	assert.Equal(t, []byte{1, 2, 3, 4}, nn.prefix[:4])
}

// Test the matching of nodes with keys.
func TestNodeMatchKeyBehavior(t *testing.T) {
	t.Parallel()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n16 := factory.newNode16()
	n16.setPrefix([]byte{1, 2, 3, 4, 5, 66, 77, 88, 99}, 5)

	assert.Equal(t, 5, n16.match(key, 0))
	assert.Equal(t, 0, n16.match(key, 1))
	assert.Equal(t, 0, n16.match(key, 100))
}

func TestCopyNode(t *testing.T) {
	t.Parallel()

	// Define test data
	src := &Node{
		childrenLen: 3,
		prefixLen:   5,
		prefix:      [maxPrefixLen]byte{'a', 'b', 'c', 'd', 'e'},
	}

	dst := &Node{
		childrenLen: 0,
		prefixLen:   0,
		prefix:      [maxPrefixLen]byte{},
	}

	// Call the function being tested
	copyNode(dst, src)

	// Use assertions to verify the outcomes
	assert.Equal(t, uint16(0), dst.childrenLen, "childrenLen should not be copied")
	assert.Equal(t, src.prefixLen, dst.prefixLen, "prefixLen should be copied correctly")

	maxCopyLen := minInt(int(src.prefixLen), maxPrefixLen)
	for i := 0; i < maxCopyLen; i++ {
		assert.Equal(t, src.prefix[i], dst.prefix[i], "prefix[%d] should be copied correctly", i)
	}
}

// Test adding children to nodes and retrieving them.
func TestNodeAddChildAndFindChild(t *testing.T) {
	t.Parallel()

	nodeKinds := []struct {
		name        string
		node        *NodeRef
		maxChildren int
	}{
		{"Node4Kind", factory.newNode4(), node4Max},
		{"Node16Kind", factory.newNode16(), node16Max},
		{"Node48Kind", factory.newNode48(), node48Max},
		{"Node256Kind", factory.newNode256(), node256Max},
	}

	for _, n := range nodeKinds {
		n := n
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < n.maxChildren; i++ {
				leaf := factory.newLeaf(Key{byte(i)}, i)
				n.node.addChild(keyChar{ch: byte(i)}, leaf)
			}

			for i := 0; i < n.maxChildren; i++ {
				leaf := n.node.findChildByKey(Key{byte(i)}, 0)
				assert.NotNil(t, *leaf, "child should not be nil for key %d", i)
				val, ok := (*leaf).Leaf().value.(int)
				assert.True(t, ok, "value should be of type int")
				assert.Equal(t, i, val, "value should be %d", i)
			}
		})
	}
}

// Test indexing functionality across different nodes.
func TestNodeIndex(t *testing.T) {
	t.Parallel()

	nodes := []*NodeRef{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
		factory.newNode256(),
	}

	for _, n := range nodes {
		maxChildren := 0

		switch n.kind {
		case Node4Kind:
			maxChildren = node4Max
		case Node16Kind:
			maxChildren = node16Max
		case Node48Kind:
			maxChildren = node48Max
		case Node256Kind:
			maxChildren = node256Max
		case LeafKind:
			t.Fatal("LeafKind Node should not be tested here")
		}

		for i := 0; i < maxChildren; i++ {
			leaf := factory.newLeaf(Key{byte(i)}, i)
			n.addChild(keyChar{ch: byte(i)}, leaf)
		}

		for i := 0; i < maxChildren; i++ {
			assert.Equal(t, i, toNode(n).index(keyChar{ch: byte(i)}))
		}
	}
}

// Test minimum and maximum functionality to ensure they return correct LeafKind nodes.
func TestNodesMinimumMaximum(t *testing.T) {
	t.Parallel()

	nodes := []struct {
		node  *NodeRef
		count int
	}{
		{factory.newNode4(), 3},
		{factory.newNode16(), 15},
		{factory.newNode48(), 47},
		{factory.newNode256(), 255},
	}

	for _, n := range nodes {
		n := n
		t.Run(n.node.kind.String(), func(t *testing.T) {
			t.Parallel()

			for j := 1; j <= n.count; j++ {
				kc := keyChar{ch: byte(j)}
				leaf := factory.newLeaf([]byte{byte(j)}, byte(j))
				n.node.addChild(kc, leaf)
			}

			minLeaf := n.node.minimum()
			assert.Equal(t, Key{1}, minLeaf.key)

			maxLeaf := n.node.maximum()
			assert.Equal(t, Key{byte(n.count)}, maxLeaf.key)
		})
	}
}

// Test adding and finding children in a Node4Kind.
func TestNode4AddChildAndFindChild(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	child := factory.newNode4()
	k := Key{1}
	parent.addChild(keyChar{ch: k[0]}, child)

	assert.Equal(t, 1, int(parent.node().childrenLen))
	assert.Equal(t, child, *parent.findChildByKey(k, 0))
}

// Test that Node4Kind maintains sorted order when adding children.
func TestNode4AddChildTwicePreserveSorted(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	child1 := factory.newNode4()
	child2 := factory.newNode4()

	parent.addChild(keyChar{ch: 2}, child1)
	parent.addChild(keyChar{ch: 1}, child2)

	assert.Equal(t, 2, int(parent.node().childrenLen))
	assert.Equal(t, byte(1), parent.node4().keys[0])
	assert.Equal(t, byte(2), parent.node4().keys[1])
}

// Test Node4Kind maintains sorted order with multiple children.
func TestNode4AddChild4PreserveSorted(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	for i := 4; i > 0; i-- {
		parent.addChild(keyChar{ch: byte(i)}, factory.newNode4())
	}

	assert.Equal(t, 4, int(parent.node().childrenLen))
	assert.Equal(t, []byte{1, 2, 3, 4}, parent.node4().keys[:])
}

// Test Node16Kind maintains sorted order with multiple children.
func TestNode16AddChild16PreserveSorted(t *testing.T) {
	t.Parallel()

	parent := factory.newNode16()
	for i := 16; i > 0; i-- {
		parent.addChild(keyChar{ch: byte(i)}, factory.newNode16())
	}

	assert.Equal(t, 16, int(parent.node().childrenLen))

	for i := 0; i < 16; i++ {
		assert.Equal(t, byte(i+1), parent.node16().keys[i])
	}
}

// Test growing a Node.
func TestNodeGrow(t *testing.T) {
	t.Parallel()

	nodeKinds := []struct {
		name     string
		node     *NodeRef
		expected Kind
	}{
		{"Node4Kind", factory.newNode4(), Node16Kind},
		{"Node16Kind", factory.newNode16(), Node48Kind},
		{"Node48Kind", factory.newNode48(), Node256Kind},
	}

	for _, tt := range nodeKinds {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			newNode := toNode(tt.node).grow()
			assert.Equal(t, tt.expected, newNode.kind)
		})
	}
}

// Test shrinking a Node.
func TestNodeShrink(t *testing.T) {
	t.Parallel()

	nodeKinds := []struct {
		name        string
		node        *NodeRef
		expected    Kind
		minChildren int
	}{
		{"Node256Kind", factory.newNode256(), Node48Kind, node256Min},
		{"Node48Kind", factory.newNode48(), Node16Kind, node48Min},
		{"Node16Kind", factory.newNode16(), Node4Kind, node16Min},
		{"Node4Kind", factory.newNode4(), LeafKind, node4Min},
	}

	for _, tt := range nodeKinds {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for j := 0; j < tt.minChildren; j++ {
				if tt.node.kind != Node4Kind {
					tt.node.addChild(keyChar{ch: byte(j)}, factory.newNode4())
				} else {
					tt.node.addChild(keyChar{ch: byte(j)}, factory.newLeaf(Key{byte(j)}, "value"))
				}
			}

			newNode := toNode(tt.node).shrink()
			assert.Equal(t, tt.expected, newNode.kind)
		})
	}
}
