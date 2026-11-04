package searchtrees

import (
	"fmt"
	"strconv"

	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type avlNode[T cmp.Ordered] struct {
	value       T
	left, right *avlNode[T]
	height      int
}

// AVL is an AVL tree of distinct values. The zero AVL is empty and ready.
type AVL[T cmp.Ordered] struct {
	root *avlNode[T]
	size int
	// Rotations counts the rotations done so far, for teaching.
	Rotations int
	rec       trace.Recorder
}

// NewAVL returns a tree holding values.
func NewAVL[T cmp.Ordered](values ...T) *AVL[T] {
	t := &AVL[T]{}
	for _, v := range values {
		t.Insert(v)
	}
	return t
}

// SetRecorder makes the tree report each rotation (Mark "rotate", Note says
// which) to r; nil stops it.
func (t *AVL[T]) SetRecorder(r trace.Recorder) { t.rec = r }

func (t *AVL[T]) record(note string) {
	t.Rotations++
	if t.rec != nil {
		t.rec.Record(trace.Event{Kind: trace.Mark, Note: note})
	}
}

// Len is the number of values. Time: O(1).
func (t *AVL[T]) Len() int { return t.size }

// Height is the number of levels; 0 for an empty tree. Time: O(1).
func (t *AVL[T]) Height() int { return avlHeight(t.root) }

// Insert adds v and reports whether it was new. Time: O(log n).
func (t *AVL[T]) Insert(v T) bool {
	var added bool
	t.root, added = t.insert(t.root, v)
	if added {
		t.size++
	}
	return added
}

// Remove deletes v and reports whether it was there. Time: O(log n).
func (t *AVL[T]) Remove(v T) bool {
	var removed bool
	t.root, removed = t.remove(t.root, v)
	if removed {
		t.size--
	}
	return removed
}

// Contains reports whether v is in the tree. Time: O(log n).
func (t *AVL[T]) Contains(v T) bool {
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

// Values returns the values in ascending order. Time: O(n).
func (t *AVL[T]) Values() []T {
	out := make([]T, 0, t.size)
	var walk func(*avlNode[T])
	walk = func(n *avlNode[T]) {
		if n != nil {
			walk(n.left)
			out = append(out, n.value)
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

func avlHeight[T cmp.Ordered](n *avlNode[T]) int {
	if n == nil {
		return 0
	}
	return n.height
}

func avlFix[T cmp.Ordered](n *avlNode[T]) {
	n.height = 1 + max(avlHeight(n.left), avlHeight(n.right))
}

func avlBalance[T cmp.Ordered](n *avlNode[T]) int { return avlHeight(n.left) - avlHeight(n.right) }

func (t *AVL[T]) rotateRight(n *avlNode[T]) *avlNode[T] {
	l := n.left
	n.left, l.right = l.right, n
	avlFix(n)
	avlFix(l)
	return l
}

func (t *AVL[T]) rotateLeft(n *avlNode[T]) *avlNode[T] {
	r := n.right
	n.right, r.left = r.left, n
	avlFix(n)
	avlFix(r)
	return r
}

// rebalance restores the AVL property at n with one or two rotations.
func (t *AVL[T]) rebalance(n *avlNode[T]) *avlNode[T] {
	avlFix(n)
	switch b := avlBalance(n); {
	case b > 1 && avlBalance(n.left) >= 0:
		t.record("rotate right")
		return t.rotateRight(n)
	case b > 1:
		t.record("rotate left-right")
		n.left = t.rotateLeft(n.left)
		return t.rotateRight(n)
	case b < -1 && avlBalance(n.right) <= 0:
		t.record("rotate left")
		return t.rotateLeft(n)
	case b < -1:
		t.record("rotate right-left")
		n.right = t.rotateRight(n.right)
		return t.rotateLeft(n)
	}
	return n
}

func (t *AVL[T]) insert(n *avlNode[T], v T) (*avlNode[T], bool) {
	if n == nil {
		return &avlNode[T]{value: v, height: 1}, true
	}
	var added bool
	switch {
	case v < n.value:
		n.left, added = t.insert(n.left, v)
	case v > n.value:
		n.right, added = t.insert(n.right, v)
	default:
		return n, false
	}
	return t.rebalance(n), added
}

func (t *AVL[T]) remove(n *avlNode[T], v T) (*avlNode[T], bool) {
	if n == nil {
		return nil, false
	}
	var removed bool
	switch {
	case v < n.value:
		n.left, removed = t.remove(n.left, v)
	case v > n.value:
		n.right, removed = t.remove(n.right, v)
	default:
		if n.left == nil {
			return n.right, true
		}
		if n.right == nil {
			return n.left, true
		}
		succ := n.right
		for succ.left != nil {
			succ = succ.left
		}
		n.value = succ.value
		n.right, _ = t.remove(n.right, succ.value)
		removed = true
	}
	return t.rebalance(n), removed
}

// Valid reports whether the tree is ordered and balanced, with correct
// heights. Time: O(n).
func (t *AVL[T]) Valid() bool {
	var check func(n *avlNode[T], lo, hi *T) bool
	check = func(n *avlNode[T], lo, hi *T) bool {
		if n == nil {
			return true
		}
		if (lo != nil && n.value <= *lo) || (hi != nil && n.value >= *hi) {
			return false
		}
		if n.height != 1+max(avlHeight(n.left), avlHeight(n.right)) || avlBalance(n) > 1 || avlBalance(n) < -1 {
			return false
		}
		return check(n.left, lo, &n.value) && check(n.right, &n.value, hi)
	}
	return check(t.root, nil, nil)
}

// Snapshot describes the tree in pre-order, each node labeled with its value
// and its edges "left" or "right". It implements trace.Snapshotter.
func (t *AVL[T]) Snapshot() trace.Snapshot {
	return binarySnapshot(t.root, func(n *avlNode[T]) (*avlNode[T], *avlNode[T]) { return n.left, n.right },
		func(n *avlNode[T], isRoot bool) (string, string) {
			return fmt.Sprint(n.value), rootRole(isRoot, "h"+strconv.Itoa(n.height))
		})
}
