package greedy

import (
	"container/heap"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxHuffmanText is the longest text Huffman accepts.
const MaxHuffmanText = 100_000

// Huffman is the result of Huffman coding a text.
type Huffman struct {
	// Codes maps each byte of the text to its code, a string of 0s and 1s.
	Codes map[byte]string
	// Encoded is the text written in the codes.
	Encoded string
	// Bits is len(Encoded); OriginalBits is 8 per byte of the text.
	Bits, OriginalBits int
	// Tree is the code tree: a leaf per byte labeled "char:freq", an internal
	// node labeled with the sum, edges labeled 0 (left) and 1 (right).
	Tree trace.Snapshot
	// Merges is the forest after each merge, the first being the leaves
	// alone: how the tree is built.
	Merges []trace.Snapshot
}

// HuffmanCoding builds a Huffman code for text (David Huffman, 1952): start
// with a tree per byte weighted by how often it occurs, and repeatedly merge
// the two lightest trees; a byte's code is its path from the root. Greedy is
// optimal: merging the two rarest symbols first is part of some optimal
// prefix code, so no prefix code encodes the text in fewer bits. A text with
// one distinct byte gets the code "0". Ties go to the lighter tree created
// first, so the code is deterministic.
//
// Time: O(n + k log k) for n bytes with k distinct. Space: O(n + k).
func HuffmanCoding(text string) (Huffman, error) { return huffman(text, nop) }

// TraceHuffmanCoding is HuffmanCoding, reporting each merge (Mark "merge",
// Values the two weights and the sum).
func TraceHuffmanCoding(text string, r trace.Recorder) (Huffman, error) {
	return huffman(text, rec{r})
}

type hnode struct {
	weight, order int
	char          byte
	leaf          bool
	left, right   *hnode
}

type hheap []*hnode

func (h hheap) Len() int { return len(h) }
func (h hheap) Less(i, j int) bool {
	if h[i].weight != h[j].weight {
		return h[i].weight < h[j].weight
	}
	return h[i].order < h[j].order
}
func (h hheap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *hheap) Push(x any)   { *h = append(*h, x.(*hnode)) }
func (h *hheap) Pop() any {
	old := *h
	v := old[len(old)-1]
	*h = old[:len(old)-1]
	return v
}

func huffman(text string, x rec) (Huffman, error) {
	if text == "" || len(text) > MaxHuffmanText {
		return Huffman{}, fmt.Errorf("%w: give 1 to %d characters of text", ErrInput, MaxHuffmanText)
	}
	var freq [256]int
	for i := 0; i < len(text); i++ {
		freq[text[i]]++
	}
	var h hheap
	order := 0
	for c := 0; c < 256; c++ {
		if freq[c] > 0 {
			h = append(h, &hnode{weight: freq[c], order: order, char: byte(c), leaf: true})
			order++
		}
	}
	heap.Init(&h)
	res := Huffman{Codes: map[byte]string{}, OriginalBits: 8 * len(text)}
	res.Merges = append(res.Merges, forest(h))
	for h.Len() > 1 {
		a := heap.Pop(&h).(*hnode)
		b := heap.Pop(&h).(*hnode)
		n := &hnode{weight: a.weight + b.weight, order: order, left: a, right: b}
		order++
		heap.Push(&h, n)
		x.r.Record(trace.Event{Kind: trace.Mark, Values: []int{a.weight, b.weight, n.weight}, Note: "merge"})
		res.Merges = append(res.Merges, forest(h))
	}
	root := h[0]
	var walk func(n *hnode, code string)
	walk = func(n *hnode, code string) {
		if n.leaf {
			if code == "" {
				code = "0" // one distinct byte: still one bit each
			}
			res.Codes[n.char] = code
			return
		}
		walk(n.left, code+"0")
		walk(n.right, code+"1")
	}
	walk(root, "")
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		b.WriteString(res.Codes[text[i]])
	}
	res.Encoded, res.Bits = b.String(), b.Len()
	res.Tree = forest(hheap{root})
	return res, nil
}

// CharLabel writes a byte readably: itself when printable, else Go's escape.
func CharLabel(c byte) string {
	if c > ' ' && c < 0x7f && c != '\'' {
		return string(c)
	}
	q := strconv.QuoteRuneToASCII(rune(c))
	return q[1 : len(q)-1]
}

// forest describes the trees in h, lightest first, as one snapshot.
func forest(h hheap) trace.Snapshot {
	trees := append(hheap(nil), h...)
	sort.Sort(trees)
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var add func(n *hnode, parent int, edge string)
	add = func(n *hnode, parent int, edge string) {
		id := len(s.Nodes)
		label := strconv.Itoa(n.weight)
		if n.leaf {
			label = CharLabel(n.char) + ":" + label
		}
		node := trace.Node{ID: id, Label: label, Parent: parent}
		if parent < 0 {
			node.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: edge})
		}
		s.Nodes = append(s.Nodes, node)
		if !n.leaf {
			add(n.left, id, "0")
			add(n.right, id, "1")
		}
	}
	for _, t := range trees {
		add(t, -1, "")
	}
	return s
}
