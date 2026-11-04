package trace

import "fmt"

// Row describes values as a row of nodes linked left to right (each edge
// labeled next, or no edges when next is ""), giving the first node the role
// first and the last the role last. Lists, queues, buffers and caches draw
// themselves with it.
func Row[T any](values []T, next string, first, last string) Snapshot {
	s := Snapshot{Layout: "row", Nodes: make([]Node, len(values))}
	for i, v := range values {
		s.Nodes[i] = Node{ID: i, Label: fmt.Sprint(v), Parent: -1}
		if i > 0 && next != "" {
			s.Edges = append(s.Edges, Edge{From: i - 1, To: i, Label: next})
		}
	}
	if n := len(values); n > 0 {
		s.Nodes[0].Role = first
		if n > 1 {
			s.Nodes[n-1].Role = last
		} else if first == "" {
			s.Nodes[0].Role = last
		}
	}
	return s
}
