package hashing

import (
	"sort"
	"strconv"
)

// Ring is consistent hashing (Karger et al., 1997): nodes and keys hash onto
// the same circle, and a key belongs to the first node clockwise from it.
// Adding or removing a node moves only the keys between it and its
// neighbor, about 1/n of them, where a plain hash mod n would move almost all.
// Each node is placed Replicas times (virtual nodes) so the load evens out.
// The zero Ring is not ready; use NewRing.
type Ring struct {
	Replicas int
	points   []uint32          // sorted positions on the circle
	owner    map[uint32]string // position -> node
	nodes    map[string]bool
}

// NewRing returns an empty ring placing each node replicas times (at least 1).
func NewRing(replicas int) *Ring {
	return &Ring{Replicas: max(1, replicas), owner: map[uint32]string{}, nodes: map[string]bool{}}
}

func point(s string) uint32 { return Murmur3_32([]byte(s), 0) }

// Add puts node on the ring. Time: O(r log(n*r)) for r replicas.
func (g *Ring) Add(node string) {
	if g.nodes[node] {
		return
	}
	g.nodes[node] = true
	for i := 0; i < g.Replicas; i++ {
		p := point(node + "#" + strconv.Itoa(i))
		if _, taken := g.owner[p]; !taken {
			g.owner[p] = node
			g.points = append(g.points, p)
		}
	}
	sort.Slice(g.points, func(i, j int) bool { return g.points[i] < g.points[j] })
}

// Remove takes node off the ring. Time: O(n*r).
func (g *Ring) Remove(node string) {
	if !g.nodes[node] {
		return
	}
	delete(g.nodes, node)
	kept := g.points[:0]
	for _, p := range g.points {
		if g.owner[p] == node {
			delete(g.owner, p)
		} else {
			kept = append(kept, p)
		}
	}
	g.points = kept
}

// Nodes returns the nodes on the ring, sorted.
func (g *Ring) Nodes() []string {
	out := make([]string, 0, len(g.nodes))
	for n := range g.nodes {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Locate returns the node key belongs to; ok is false on an empty ring.
// Time: O(log(n*r)).
func (g *Ring) Locate(key string) (node string, ok bool) {
	if len(g.points) == 0 {
		return "", false
	}
	p := point(key)
	i := sort.Search(len(g.points), func(i int) bool { return g.points[i] >= p })
	if i == len(g.points) {
		i = 0 // wrap around the circle
	}
	return g.owner[g.points[i]], true
}

// Assign returns each key's node.
func (g *Ring) Assign(keys []string) map[string]string {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k], _ = g.Locate(k)
	}
	return out
}

// Moved counts the keys whose node differs between two assignments.
func Moved(before, after map[string]string) int {
	n := 0
	for k, v := range before {
		if after[k] != v {
			n++
		}
	}
	return n
}
