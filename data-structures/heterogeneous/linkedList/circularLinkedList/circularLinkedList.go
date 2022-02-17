package circularLinkedList

type Node struct {
  value interface{}
  next *Node
}

type List struct {
  head *Node
  tail *Node
  elements []*Node
}
