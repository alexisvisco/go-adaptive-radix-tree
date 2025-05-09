package art

// Node constraints.
const (
	node4Min = 2 // minimum number of children for Node4.
	node4Max = 4 // maximum number of children for Node4.

	node16Min = node4Max + 1 // minimum number of children for Node16.
	node16Max = 16           // maximum number of children for Node16.

	node48Min = node16Max + 1 // minimum number of children for Node48.
	node48Max = 48            // maximum number of children for Node48.

	node256Min = node48Max + 1 // minimum number of children for Node256.
	node256Max = 256           // maximum number of children for Node256.
)

const (
	// maxPrefixLen is maximum prefix length for internal nodes.
	maxPrefixLen = 10
)
