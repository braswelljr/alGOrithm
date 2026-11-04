package searchtrees_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// exercise runs random inserts and removes against a map and checks valid
// after every step.
func exercise(t *testing.T, name string, s orderedSet, valid func() bool, seed int64) {
	t.Helper()
	const steps, span = 4000, 600
	t.Helper()
	r := rand.New(rand.NewSource(seed))
	model := map[int]bool{}
	for i := 0; i < steps; i++ {
		v := r.Intn(span)
		if r.Intn(3) == 0 {
			if s.Remove(v) != model[v] {
				t.Fatalf("%s: Remove(%d) disagreed at step %d", name, v, i)
			}
			delete(model, v)
		} else {
			if s.Insert(v) == model[v] {
				t.Fatalf("%s: Insert(%d) disagreed at step %d", name, v, i)
			}
			model[v] = true
		}
		if valid != nil && !valid() {
			t.Fatalf("%s: invariants broken after step %d", name, i)
		}
		if probe := r.Intn(span); s.Contains(probe) != model[probe] {
			t.Fatalf("%s: Contains(%d) disagreed", name, probe)
		}
	}
	if got, want := s.Values(), sortedKeys(model); (len(got) > 0 || len(want) > 0) && !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: values %v, want %v", name, got, want)
	}
	if s.Len() != len(model) {
		t.Fatalf("%s: Len %d, want %d", name, s.Len(), len(model))
	}
}

// orderedSet is what every set-like tree here must behave like.
type orderedSet interface {
	Insert(int) bool
	Remove(int) bool
	Contains(int) bool
	Values() []int
	Len() int
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func TestTrees(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	bst := searchtrees.NewBST[int]()
	avl := searchtrees.NewAVL[int]()
	want := map[int]bool{}
	for i := 0; i < 3000; i++ {
		v := r.Intn(500)
		if r.Intn(3) == 0 {
			if bst.Remove(v) != want[v] || avl.Remove(v) != want[v] {
				t.Fatalf("Remove(%d) disagreed", v)
			}
			delete(want, v)
		} else {
			if bst.Insert(v) == want[v] || avl.Insert(v) == want[v] {
				t.Fatalf("Insert(%d) disagreed", v)
			}
			want[v] = true
		}
		if !avl.Valid() {
			t.Fatalf("AVL unbalanced after step %d", i)
		}
	}
	keys := make([]int, 0, len(want))
	for k := range want {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if !reflect.DeepEqual(bst.Values(), keys) || !reflect.DeepEqual(avl.Values(), keys) {
		t.Fatal("trees hold the wrong values")
	}
	sorted := searchtrees.NewAVL[int]()
	chain := searchtrees.NewBST[int]()
	for i := 0; i < 1023; i++ {
		sorted.Insert(i)
		chain.Insert(i)
	}
	if sorted.Height() > 11 || chain.Height() != 1023 {
		t.Errorf("sorted inserts: AVL height %d, BST height %d", sorted.Height(), chain.Height())
	}
	if lo, _ := bst.Min(); lo != keys[0] {
		t.Errorf("Min = %d", lo)
	}
	if hi, _ := bst.Max(); hi != keys[len(keys)-1] {
		t.Errorf("Max = %d", hi)
	}
}

func TestBalancedTrees(t *testing.T) {
	rb := searchtrees.NewRedBlackTree[int]()
	exercise(t, "red-black", rb, rb.Valid, 1)
	for _, deg := range []int{2, 3, 5} {
		b := searchtrees.NewBTree[int](deg)
		exercise(t, fmt.Sprintf("b-tree t=%d", deg), b, b.Valid, int64(deg))
	}
	tr := searchtrees.NewTreap[int](7)
	exercise(t, "treap", tr, tr.Valid, 3)
	sl := searchtrees.NewSkipList[int](5)
	exercise(t, "skip list", sl, nil, 4)

	sorted := searchtrees.NewRedBlackTree[int]()
	bt := searchtrees.NewBTree[int](2)
	for i := 0; i < 1023; i++ {
		sorted.Insert(i)
		bt.Insert(i)
	}
	if h := sorted.Height(); h > 20 {
		t.Errorf("red-black height %d after 1023 sorted inserts, want at most 2 log n", h)
	}
	if h := bt.Height(); h > 10 {
		t.Errorf("B-tree height %d", h)
	}
	for _, s := range []trace.Snapshot{sorted.Snapshot(), searchtrees.NewTreap[int](1).Snapshot(), bt.Snapshot()} {
		for _, n := range s.Nodes {
			if n.Label == "" {
				t.Error("unlabeled node")
			}
		}
	}
	if s := sorted.Snapshot(); s.Nodes[0].Role != "black" {
		t.Error("the root must be black")
	}
	if s := searchtrees.NewBTree[int](2).Snapshot(); len(s.Nodes) != 0 {
		t.Error("empty B-tree should have no nodes")
	}
	if searchtrees.NewRedBlackTree("b", "a").Values()[0] != "a" {
		t.Error("strings")
	}
}

func TestSplayTree(t *testing.T) {
	s := searchtrees.NewSplayTree[int]()
	exercise(t, "splay", s, nil, 7)
	s.Insert(5)
	s.Insert(9)
	s.Contains(5)
	if r, _ := s.Root(); r != 5 {
		t.Errorf("after Contains(5) the root is %d, want 5", r)
	}
}

func TestIntervalTree(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	tree := searchtrees.NewIntervalTree()
	model := map[searchtrees.Interval]bool{}
	for step := 0; step < 3000; step++ {
		lo := r.Intn(200)
		iv := searchtrees.Interval{Lo: lo, Hi: lo + r.Intn(30)}
		if r.Intn(3) == 0 {
			if tree.Remove(iv) != model[iv] {
				t.Fatalf("Remove(%v) disagreed", iv)
			}
			delete(model, iv)
		} else {
			if tree.Insert(iv) == model[iv] {
				t.Fatalf("Insert(%v) disagreed", iv)
			}
			model[iv] = true
		}
		qlo := r.Intn(220)
		q := searchtrees.Interval{Lo: qlo, Hi: qlo + r.Intn(20)}
		want := map[searchtrees.Interval]bool{}
		for iv := range model {
			if iv.Overlaps(q) {
				want[iv] = true
			}
		}
		got := map[searchtrees.Interval]bool{}
		for _, iv := range tree.Overlapping(q) {
			got[iv] = true
		}
		if !reflect.DeepEqual(got, want) && (len(got) > 0 || len(want) > 0) {
			t.Fatalf("step %d: Overlapping(%v) = %v, want %v", step, q, got, want)
		}
	}
	if tree.Len() != len(model) || len(tree.Intervals()) != len(model) {
		t.Fatalf("Len %d, want %d", tree.Len(), len(model))
	}
	if tree.Insert(searchtrees.Interval{Lo: 3, Hi: 1}) {
		t.Error("an interval with Lo > Hi was accepted")
	}
}
