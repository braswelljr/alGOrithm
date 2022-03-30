package aStarSearch

type Node struct {
  parent *Node
  g, f, h int
}

type Heuristic func(node *Node, goal *Node) int

// AStarSearch A* search algorithm
func AStarSearch(start *Node, goal *Node, heuristic Heuristic) []*Node {
  open := make([]*Node, 0)
  closed := make([]*Node, 0)
  start.g = 0
  start.h = heuristic(start, goal)
  start.f = start.g + start.h
  open = append(open, start)
  for len(open) > 0 {
    current := open[0]
    for i := 1; i < len(open); i++ {
      if open[i].f < current.f {
        current = open[i]
      }
    }
    open = append(open[:0], open[1:]...)
    closed = append(closed, current)
    if current == goal {
      return reconstructPath(current)
    }
    for _, neighbor := range current.neighbors {
      if !contains(closed, neighbor) {
        if !contains(open, neighbor) {
          neighbor.g = current.g + 1
          neighbor.h = heuristic(neighbor, goal)
          neighbor.f = neighbor.g + neighbor.h
          open = append(open, neighbor)
        } else {
          if current.g+1 < neighbor.g {
            neighbor.g = current.g + 1
            neighbor.f = neighbor.g + neighbor.h
          }
        }
      }
    }
  }
  return nil
}

func reconstructPath(current *Node) []*Node {
  path := make([]*Node, 0)
  for current != nil {
    path = append(path, current)
    current = current.parent
  }
  return path
}
