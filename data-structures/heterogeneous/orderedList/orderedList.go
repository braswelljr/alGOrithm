package orderedList

import "github.com/braswelljr/alGOrithm/internal/types"

// Node - A node in a list
// @param value - The value of the node
// @param next - The next node in the list
type Node struct {
  value interface{}
  Next  *Node
}

// List - A list of nodes
// @param head - The first node in the list
// @param tail - The last node in the list
// @param list - The list of nodes
type List[N types.Number] struct {
  head *Node
  tail *Node
  list []N
}
