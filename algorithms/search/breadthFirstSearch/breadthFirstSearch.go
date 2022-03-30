package breadthFirstSearch

// Node is an element in a graph
type Node struct {
  Value      int
  Parent     *Node
  Neighbours []*Node
}

// BreadthFirstSearch performs a breadth first search on a graph or tree
func BreadthFirstSearch(root *Node, target *Node) *Node {
  // set root to nil
  if root == nil {
    return nil
  }

  // create a queue to store nodes
  var visited []*Node
  queue := []*Node{root}

  // loop until queue is empty
  for len(queue) > 0 {
    // dequeue the first node
    node := queue[0]
    // remove the first node from the queue and return it
    queue = queue[1:]

    // check if the node is the target
    if node.Value == target.Value {
      return node // return the node
    }

    // add the node to the visited list
    visited = append(visited, node)

    // add the node's neighbours to the queue
    for _, neighbour := range node.Neighbours {
      // append the neighbour to the queue if it is not already in the queue
      for _, n := range queue {
        if n != neighbour {
          queue = append(queue, neighbour)
        }
      }
    }
  }

  // return nil if the target was not found
  return nil
}
