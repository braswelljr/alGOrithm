package orderedList

import "github.com/braswelljr/alGOrithm/internal/types"

type Node struct {
  Value interface{}
  Next  *Node
}

type List[N types.Number] struct {
  head *Node
  tail *Node
  list []N
}
