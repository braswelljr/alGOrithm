package text

import "sort"

type trieNode struct {
	next  map[byte]*trieNode
	count int // how many times the word ending here was added
}

// Autocomplete returns up to limit words from words that start with prefix,
// the most frequent in words first, ties in alphabetical order. It builds a
// trie of the words and collects the ones under the prefix's node.
//
// Time: O(total length) to build, O(prefix + subtree + r log r) to answer.
// Space: O(total length).
func Autocomplete(words []string, prefix string, limit int) []string {
	root := &trieNode{next: map[byte]*trieNode{}}
	for _, w := range words {
		n := root
		for i := 0; i < len(w); i++ {
			child, ok := n.next[w[i]]
			if !ok {
				child = &trieNode{next: map[byte]*trieNode{}}
				n.next[w[i]] = child
			}
			n = child
		}
		n.count++
	}
	n := root
	for i := 0; i < len(prefix) && n != nil; i++ {
		n = n.next[prefix[i]]
	}
	if n == nil {
		return nil
	}
	var found []WordCount
	var walk func(n *trieNode, word []byte)
	walk = func(n *trieNode, word []byte) {
		if n.count > 0 {
			found = append(found, WordCount{string(word), n.count})
		}
		for c, child := range n.next {
			walk(child, append(word, c))
		}
	}
	walk(n, []byte(prefix))
	sort.Slice(found, func(i, j int) bool {
		if found[i].Count != found[j].Count {
			return found[i].Count > found[j].Count
		}
		return found[i].Word < found[j].Word
	})
	out := []string{}
	for i := 0; i < len(found) && i < limit; i++ {
		out = append(out, found[i].Word)
	}
	return out
}
