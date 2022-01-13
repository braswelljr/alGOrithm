package trees

type Node struct {
  Value int
  Left  *Node
  Right *Node
}

type Tree struct {
  Root *Node
}

// Init initializes a new tree with a root node
func (tree *Tree) Init() *Tree {
  return tree
}

// New creates a new tree with a root node
func New() *Tree {
  return new(Tree).Init()
}

// Insert inserts a new node into the tree
func (tree *Tree) Insert(value int) {
  // set the root node if it doesn't exist
  tree.Root = insert(tree.Root, value)
}

func insert(node *Node, value int) *Node {
  // if the node doesn't exist, create it
  if node == nil {
    return &Node{Value: value}
  }

  // che
  if value < node.Value {
    node.Left = insert(node.Left, value)
  } else {
    node.Right = insert(node.Right, value)
  }

  return node
}
