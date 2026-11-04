// Package persistent holds immutable data structures: an update returns a new
// version and leaves the old one exactly as it was, and the versions share
// every part the update did not touch (structural sharing), so keeping a
// hundred versions costs far less than a hundred copies. Undo histories,
// functional languages, React state and Git's object store all work this
// way.
//
// Stack shares its tail between versions; Tree copies only the path from the
// root to the change. Values of either are safe to share between goroutines,
// since nothing is ever written after it is made.
package persistent

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// visitFunc is called for each node of one version, with a key identifying
// the node, its label, and the key of its parent (nil for the version's first
// node). It reports whether the node is new; a node already drawn, and so
// everything below it, is shared and need not be walked again.
type visitFunc func(key any, label string, parent any, side string) bool

// versions draws several versions of a linked structure at once, each node
// once however many versions reach it, so the sharing shows.
func versions(names []string, walk func(i int, visit visitFunc)) trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	ids := map[any]int{}
	for i, name := range names {
		head := len(s.Nodes)
		s.Nodes = append(s.Nodes, trace.Node{ID: head, Label: name, Role: "root", Parent: -1})
		walk(i, func(key any, label string, parent any, side string) bool {
			from := head
			if parent != nil {
				from = ids[parent]
			}
			id, seen := ids[key]
			if !seen {
				id = len(s.Nodes)
				ids[key] = id
				s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: label, Parent: from})
			}
			s.Edges = append(s.Edges, trace.Edge{From: from, To: id, Label: side})
			return !seen
		})
	}
	return s
}

func versionName(i int) string { return fmt.Sprintf("v%d", i) }
