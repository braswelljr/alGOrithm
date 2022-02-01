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

// Remove removes a node from the tree
func (tree *Tree) Remove(value int) {
  tree.Root = remove(tree.Root, value)
}

func remove(node *Node, value int) *Node {
  if node == nil {
    return nil
  }

  if value < node.Value {
    node.Left = remove(node.Left, value)
  } else if value > node.Value {
    node.Right = remove(node.Right, value)
  } else {
    // if the node has no children
    if node.Left == nil && node.Right == nil {
      return nil
    }

    // if the node has only a left child
    if node.Right == nil {
      return node.Left
    }

    // if the node has only a right child
    if node.Left == nil {
      return node.Right
    }

    // if the node has both children
    // find the inorder successor
    successor := node.Right
    for successor.Left != nil {
      successor = successor.Left
    }

    // copy the value from the inorder successor
    node.Value = successor.Value

    // recursively remove the inorder successor
    node.Right = remove(node.Right, successor.Value)
  }

  return node
}

// Search searches for a node in the tree
func (tree *Tree) Search(value int) bool {
  return search(tree.Root, value)
}

func search(node *Node, value int) bool {
  if node == nil {
    return false
  }

  // if the value is less than the current node, search the left subtree
  if value < node.Value {
    return search(node.Left, value)
    // if the value is greater than the current node, search the right subtree
  } else if value > node.Value {
    return search(node.Right, value)
  }

  return true
}

// Size - returns the number of nodes in the tree
func (tree *Tree) Size() int {
  return size(tree.Root)
}

func size(node *Node) int {
  // if node has a null value return 0 as the size
  if node == nil {
    return 0
  }

  // recursively check for the size of the left and right subtrees
  return size(node.Left) + size(node.Right) + 1
}

// Clear - removes all nodes from the tree or returns a new tree
func (tree *Tree) Clear() *Tree {
  // clean all the nodes in the tree
  return new(Tree).Init()
}
