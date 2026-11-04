package lookup

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type trieNode struct {
	children map[byte]*trieNode
	end      bool // a word ends here
}

// Trie is a prefix tree of words: words share the path of their common
// prefix, one node per letter, so finding a word, or every word with a prefix,
// takes time proportional to its length. The zero Trie is not ready; use
// NewTrie.
type Trie struct {
	root  *trieNode
	words int
}

// NewTrie returns a trie holding words.
func NewTrie(words ...string) *Trie {
	t := &Trie{root: &trieNode{children: map[byte]*trieNode{}}}
	for _, w := range words {
		t.Insert(w)
	}
	return t
}

// Len is the number of words. Time: O(1).
func (t *Trie) Len() int { return t.words }

// Insert adds word and reports whether it was new. Time: O(len(word)).
func (t *Trie) Insert(word string) bool {
	n := t.root
	for i := 0; i < len(word); i++ {
		c := word[i]
		next, ok := n.children[c]
		if !ok {
			next = &trieNode{children: map[byte]*trieNode{}}
			n.children[c] = next
		}
		n = next
	}
	if n.end {
		return false
	}
	n.end = true
	t.words++
	return true
}

func (t *Trie) find(s string) *trieNode {
	n := t.root
	for i := 0; i < len(s) && n != nil; i++ {
		n = n.children[s[i]]
	}
	return n
}

// Contains reports whether word was inserted. Time: O(len(word)).
func (t *Trie) Contains(word string) bool {
	n := t.find(word)
	return n != nil && n.end
}

// WithPrefix returns every word starting with prefix, in order.
// Time: O(len(prefix) + the size of the result).
func (t *Trie) WithPrefix(prefix string) []string {
	var out []string
	var walk func(n *trieNode, word []byte)
	walk = func(n *trieNode, word []byte) {
		if n.end {
			out = append(out, string(word))
		}
		for _, c := range trieKeys(n) {
			walk(n.children[c], append(word, c))
		}
	}
	if n := t.find(prefix); n != nil {
		walk(n, []byte(prefix))
	}
	return out
}

// Remove deletes word and prunes nodes no other word needs. It reports
// whether word was there. Time: O(len(word)).
func (t *Trie) Remove(word string) bool {
	var del func(n *trieNode, i int) (found, prune bool)
	del = func(n *trieNode, i int) (bool, bool) {
		if i == len(word) {
			if !n.end {
				return false, false
			}
			n.end = false
			return true, len(n.children) == 0
		}
		child, ok := n.children[word[i]]
		if !ok {
			return false, false
		}
		found, prune := del(child, i+1)
		if prune {
			delete(n.children, word[i])
		}
		return found, found && !n.end && len(n.children) == 0
	}
	found, _ := del(t.root, 0)
	if found {
		t.words--
	}
	return found
}

func trieKeys(n *trieNode) []byte {
	ks := make([]byte, 0, len(n.children))
	for c := range n.children {
		ks = append(ks, c)
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
	return ks
}

// Snapshot describes the trie: the root, then one trieNode per letter, a trieNode
// where a word ends having the role "word". It implements trace.Snapshotter.
func (t *Trie) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "*", Role: "root", Parent: -1}}}
	var walk func(n *trieNode, id int)
	walk = func(n *trieNode, id int) {
		for _, c := range trieKeys(n) {
			child := n.children[c]
			cid := len(s.Nodes)
			nd := trace.Node{ID: cid, Label: string(c), Parent: id}
			if child.end {
				nd.Role = "word"
			}
			s.Nodes = append(s.Nodes, nd)
			s.Edges = append(s.Edges, trace.Edge{From: id, To: cid})
			walk(child, cid)
		}
	}
	walk(t.root, 0)
	return s
}
