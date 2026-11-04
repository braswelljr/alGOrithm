package strmatch

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Match is one occurrence of one of several patterns.
type Match struct {
	Pattern int `json:"pattern"` // index into the patterns
	Start   int `json:"start"`   // index into the text
}

type acNode struct {
	next   map[byte]int
	fail   int
	output []int // patterns ending here, through the failure links too
	depth  int
	letter byte
	parent int
}

// automaton is the Aho-Corasick machine: a trie of the patterns plus, for each
// node, a failure link to the longest proper suffix that is also in the trie.
type automaton struct {
	nodes    []acNode
	patterns []string
}

func build(patterns []string) *automaton {
	a := &automaton{nodes: []acNode{{next: map[byte]int{}, parent: -1}}, patterns: patterns}
	for p, pat := range patterns {
		cur := 0
		for i := 0; i < len(pat); i++ {
			c := pat[i]
			nxt, ok := a.nodes[cur].next[c]
			if !ok {
				nxt = len(a.nodes)
				a.nodes = append(a.nodes, acNode{next: map[byte]int{}, depth: a.nodes[cur].depth + 1, letter: c, parent: cur})
				a.nodes[cur].next[c] = nxt
			}
			cur = nxt
		}
		if len(pat) > 0 {
			a.nodes[cur].output = append(a.nodes[cur].output, p)
		}
	}
	queue := []int{} // breadth first, so a node's failure is set before its children's
	for _, c := range sortedKeys(a.nodes[0].next) {
		queue = append(queue, a.nodes[0].next[c])
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, c := range sortedKeys(a.nodes[u].next) {
			v := a.nodes[u].next[c]
			f := a.nodes[u].fail
			for f > 0 && a.nodes[f].next[c] == 0 {
				f = a.nodes[f].fail
			}
			if w, ok := a.nodes[f].next[c]; ok && w != v {
				a.nodes[v].fail = w
			}
			a.nodes[v].output = append(a.nodes[v].output, a.nodes[a.nodes[v].fail].output...)
			queue = append(queue, v)
		}
	}
	return a
}

func sortedKeys(m map[byte]int) []byte {
	ks := make([]byte, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
	return ks
}

// AhoCorasick finds every occurrence of every pattern in text in one pass
// (Aho and Corasick, 1975): a trie of the patterns with failure links, so after
// a mismatch the machine falls back to the longest suffix it has seen that
// starts another pattern, never re-reading the text. Matches are ordered by
// where they end, then by pattern; empty patterns match nowhere.
//
// Time: O(n + total pattern length + matches). Space: O(total pattern length).
func AhoCorasick(text string, patterns []string) []Match { return ahoCorasick(text, patterns, nop) }

// TraceAhoCorasick is AhoCorasick, reporting each letter read (Visit,
// Indices[0] in the text, Values[0] the trie node reached) and each match
// (Mark "found" over the matched text, Values[0] the pattern).
func TraceAhoCorasick(text string, patterns []string, r trace.Recorder) []Match {
	return ahoCorasick(text, patterns, rec{r})
}

func ahoCorasick(text string, patterns []string, x rec) []Match {
	a := build(patterns)
	var out []Match
	cur := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		for cur > 0 {
			if _, ok := a.nodes[cur].next[c]; ok {
				break
			}
			cur = a.nodes[cur].fail // fall back to the longest suffix in the trie
		}
		if nxt, ok := a.nodes[cur].next[c]; ok {
			cur = nxt
		}
		x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}, Values: []int{cur}})
		outs := append([]int(nil), a.nodes[cur].output...)
		sort.Ints(outs)
		for _, p := range outs {
			start := i - len(patterns[p]) + 1
			out = append(out, Match{Pattern: p, Start: start})
			x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start, i}, Values: []int{p}, Note: "found"})
		}
	}
	return out
}

// AhoCorasickTrie describes the patterns' trie as a snapshot: the root "*",
// one node per letter, a node where a pattern ends having the role "word",
// and a dashed "fail" edge from each node whose failure link is not the root.
func AhoCorasickTrie(patterns []string) trace.Snapshot {
	a := build(patterns)
	s := trace.Snapshot{Layout: "tree", Nodes: make([]trace.Node, len(a.nodes))}
	for i, n := range a.nodes {
		node := trace.Node{ID: i, Label: string(n.letter), Parent: n.parent}
		switch {
		case i == 0:
			node.Label, node.Role = "*", "root"
		case len(n.output) > 0 && contains(n.output, patterns, n.depth):
			node.Role = "word"
		}
		s.Nodes[i] = node
		if n.parent >= 0 {
			s.Edges = append(s.Edges, trace.Edge{From: n.parent, To: i})
		}
		if i > 0 && n.fail > 0 {
			s.Edges = append(s.Edges, trace.Edge{From: i, To: n.fail, Label: "fail"})
		}
	}
	return s
}

// contains reports whether a pattern of the node's own depth ends at it (not
// only through a failure link).
func contains(outputs []int, patterns []string, depth int) bool {
	for _, p := range outputs {
		if len(patterns[p]) == depth {
			return true
		}
	}
	return false
}
