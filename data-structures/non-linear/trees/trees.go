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

// Traverse - traverses the tree in order
func (tree *Tree) Traverse() []int {
	return traverse(tree.Root)
}

// traverse -
func traverse(node *Node) []int {
	// if the node is null, return an empty array
	if node == nil {
		return []int{}
	}

	// recursively traverse the left subtree
	left := traverse(node.Left)

	// recursively traverse the right subtree
	right := traverse(node.Right)

	// return the left subtree, the current node, and the right subtree
	return append(append(left, node.Value), right...)
}

// Min - returns the minimum value in the tree
func (tree *Tree) Min() int {
	return min(tree.Root)
}

// min - recursively steps through left subtree and returns the min value
func min(node *Node) int {
	// if the node is null, return 0
	if node == nil {
		return 0
	}

	// if the node has no left child, return the value of the node
	if node.Left == nil {
		return node.Value
	}

	// recursively check the left subtree
	return min(node.Left)
}

// Max - returns the maximum value in the tree
func (tree *Tree) Max() int {
	return max(tree.Root)
}

// max - recursively steps through right subtree and returns the max value
func max(node *Node) int {
	// if the node is null, return 0
	if node == nil {
		return 0
	}

	// if the node has no right child, return the value of the node
	if node.Right == nil {
		return node.Value
	}

	// recursively check the right subtree
	return max(node.Right)
}

// Height - returns the height of the tree
func (tree *Tree) Height() int {
	return height(tree.Root)
}

// height - recursively checks the height of the left and right subtrees
func height(node *Node) int {
	// if the node is null, return 0
	if node == nil {
		return 0
	}

	// recursively check the height of the left subtree
	left := height(node.Left)

	// recursively check the height of the right subtree
	right := height(node.Right)

	// return the greatest of the left and right subtree heights
	if left > right {
		return left + 1
	}
	return right + 1
}

// Graft - Adding a tree or section to a tree
//func (tree *Tree) Graft(node *Node) Tree{
//  return
//}
