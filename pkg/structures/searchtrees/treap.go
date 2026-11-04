package searchtrees

import (
	"cmp"
	"fmt"
	"math/rand"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type treapNode[T cmp.Ordered] struct {
	value       T
	priority    int
	left, right *treapNode[T]
}

// Treap is a binary search tree by value and a max-heap by a random priority
// given to each node (Seidel and Aragon, 1989): rotations keep the heap
// order, and random priorities make the tree balanced in expectation, with no
// bookkeeping. Values are distinct. The zero Treap is not ready; use
// NewTreap.
type Treap[T cmp.Ordered] struct {
	root *treapNode[T]
	size int
	rng  *rand.Rand
}

// NewTreap returns an empty treap whose priorities come from seed.
func NewTreap[T cmp.Ordered](seed int64) *Treap[T] {
	return &Treap[T]{rng: rand.New(rand.NewSource(seed))}
}

// Len is the number of values. Time: O(1).
func (t *Treap[T]) Len() int { return t.size }

func treapRight[T cmp.Ordered](n *treapNode[T]) *treapNode[T] {
	l := n.left
	n.left, l.right = l.right, n
	return l
}

func treapLeft[T cmp.Ordered](n *treapNode[T]) *treapNode[T] {
	r := n.right
	n.right, r.left = r.left, n
	return r
}

// Contains reports whether v is in the treap. Time: O(log n) expected.
func (t *Treap[T]) Contains(v T) bool {
	for n := t.root; n != nil; {
		switch {
		case v < n.value:
			n = n.left
		case v > n.value:
			n = n.right
		default:
			return true
		}
	}
	return false
}

// Insert adds v and reports whether it was new: it goes in as a leaf, then
// rotates up while its priority beats its parent's. Time: O(log n) expected.
func (t *Treap[T]) Insert(v T) bool {
	if t.Contains(v) {
		return false
	}
	p := t.rng.Intn(1000)
	var ins func(n *treapNode[T]) *treapNode[T]
	ins = func(n *treapNode[T]) *treapNode[T] {
		if n == nil {
			return &treapNode[T]{value: v, priority: p}
		}
		if v < n.value {
			n.left = ins(n.left)
			if n.left.priority > n.priority {
				n = treapRight(n)
			}
		} else {
			n.right = ins(n.right)
			if n.right.priority > n.priority {
				n = treapLeft(n)
			}
		}
		return n
	}
	t.root = ins(t.root)
	t.size++
	return true
}

// Remove deletes v and reports whether it was there: it rotates the node down
// toward its higher-priority child until it is a leaf, then cuts it.
// Time: O(log n) expected.
func (t *Treap[T]) Remove(v T) bool {
	if !t.Contains(v) {
		return false
	}
	var del func(n *treapNode[T]) *treapNode[T]
	del = func(n *treapNode[T]) *treapNode[T] {
		switch {
		case v < n.value:
			n.left = del(n.left)
		case v > n.value:
			n.right = del(n.right)
		case n.left == nil:
			return n.right
		case n.right == nil:
			return n.left
		case n.left.priority > n.right.priority:
			n = treapRight(n)
			n.right = del(n.right)
		default:
			n = treapLeft(n)
			n.left = del(n.left)
		}
		return n
	}
	t.root = del(t.root)
	t.size--
	return true
}

// Values returns the values in ascending order. Time: O(n).
func (t *Treap[T]) Values() []T {
	out := make([]T, 0, t.size)
	var walk func(*treapNode[T])
	walk = func(n *treapNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Valid reports whether the treap is ordered by value and heap-ordered by
// priority. Time: O(n).
func (t *Treap[T]) Valid() bool {
	var check func(n *treapNode[T], lo, hi *T) bool
	check = func(n *treapNode[T], lo, hi *T) bool {
		if n == nil {
			return true
		}
		if (lo != nil && n.value <= *lo) || (hi != nil && n.value >= *hi) {
			return false
		}
		for _, c := range []*treapNode[T]{n.left, n.right} {
			if c != nil && c.priority > n.priority {
				return false
			}
		}
		return check(n.left, lo, &n.value) && check(n.right, &n.value, hi)
	}
	return check(t.root, nil, nil)
}

// Snapshot describes the treap in pre-order, each node's Role its priority
// ("p=812").
func (t *Treap[T]) Snapshot() trace.Snapshot {
	return binarySnapshot(t.root, func(n *treapNode[T]) (*treapNode[T], *treapNode[T]) { return n.left, n.right },
		func(n *treapNode[T], _ bool) (string, string) {
			return fmt.Sprint(n.value), fmt.Sprintf("p=%d", n.priority)
		})
}
