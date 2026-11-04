package trees_test

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/trees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func mustParse(t *testing.T, s string) *trees.Tree {
	t.Helper()
	tr, err := trees.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func TestTraversals(t *testing.T) {
	tr := mustParse(t, "[1, 2, 3, 4, 5, null, 6]")
	cases := map[string]struct{ got, want []int }{
		"pre":  {trees.Preorder(tr), []int{1, 2, 4, 5, 3, 6}},
		"in":   {trees.Inorder(tr), []int{4, 2, 5, 1, 3, 6}},
		"post": {trees.Postorder(tr), []int{4, 5, 2, 6, 3, 1}},
	}
	for name, c := range cases {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s = %v, want %v", name, c.got, c.want)
		}
	}
	if got := trees.LevelOrder(tr); !reflect.DeepEqual(got, [][]int{{1}, {2, 3}, {4, 5, 6}}) {
		t.Errorf("level = %v", got)
	}
	var log trace.Log
	trees.TraceInorder(tr, &log)
	if log.Count(trace.Visit) != 6 || log.Events[0].Values[0] != tr.Find(4).ID {
		t.Errorf("trace visits %d, first %v", log.Count(trace.Visit), log.Events[0].Values)
	}
	if ids := trees.Preorder(&trees.Tree{}); ids != nil {
		t.Error("an empty tree has a traversal")
	}
}

func TestMorrisEqualsInorderAndRestoresTheTree(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 50; i++ {
		vals := r.Perm(r.Intn(40))
		tr := trees.FromSorted(nil)
		for _, v := range vals { // a random-shaped BST
			tr = insert(tr, v)
		}
		before := trees.Serialize(tr)
		var log trace.Log
		if got, want := trees.TraceMorris(tr, &log), trees.Inorder(tr); !reflect.DeepEqual(got, want) {
			t.Fatalf("Morris %v, inorder %v", got, want)
		}
		if trees.Serialize(tr) != before {
			t.Fatal("Morris left a thread behind")
		}
		if log.Count(trace.Link) != log.Count(trace.Unlink) {
			t.Fatal("threads made and removed differ")
		}
	}
}

func insert(t *trees.Tree, v int) *trees.Tree {
	link := &t.Root
	for *link != nil {
		if v < (*link).Value {
			link = &(*link).Left
		} else {
			link = &(*link).Right
		}
	}
	*link = &trees.Node{Value: v}
	t.Renumber()
	return t
}

func TestMeasures(t *testing.T) {
	tr := mustParse(t, "1 2 3 4 5")
	if trees.Height(tr) != 3 || trees.Diameter(tr) != 3 || !trees.IsBalanced(tr) {
		t.Errorf("height %d diameter %d balanced %v", trees.Height(tr), trees.Diameter(tr), trees.IsBalanced(tr))
	}
	chain := mustParse(t, "1 null 2 null 3")
	if trees.IsBalanced(chain) || trees.Height(chain) != 3 {
		t.Error("a chain of 3 is not balanced")
	}
	if !trees.IsBST(mustParse(t, "4 2 6 1 3 5 7")) || trees.IsBST(mustParse(t, "5 1 4 null null 3 6")) {
		t.Error("IsBST wrong")
	}
	// The classic trap: every node is ordered with its children, but 6 is in 5's left subtree... of 10.
	if trees.IsBST(mustParse(t, "10 5 15 null 12")) {
		t.Error("IsBST must check whole subtrees, not only children")
	}
	b := trees.FromSorted([]int{1, 2, 3, 4, 5, 6, 7})
	if !trees.IsBST(b) || trees.Height(b) != 3 || !reflect.DeepEqual(trees.Inorder(b), []int{1, 2, 3, 4, 5, 6, 7}) {
		t.Error("FromSorted is not a balanced BST")
	}
}

func TestLCA(t *testing.T) {
	tr := mustParse(t, "3 5 1 6 2 0 8 null null 7 4")
	for _, c := range [][3]int{{5, 1, 3}, {5, 4, 5}, {7, 4, 2}, {6, 8, 3}} {
		if n, ok := trees.LowestCommonAncestor(tr, c[0], c[1]); !ok || n.Value != c[2] {
			t.Errorf("LCA(%d, %d) = %v, want %d", c[0], c[1], n, c[2])
		}
	}
	if _, ok := trees.LowestCommonAncestor(tr, 5, 99); ok {
		t.Error("LCA with a missing value succeeded")
	}
	bst := mustParse(t, "6 2 8 0 4 7 9 null null 3 5")
	if n, _ := trees.LowestCommonAncestorBST(bst, 3, 5); n.Value != 4 {
		t.Errorf("BST LCA(3, 5) = %d", n.Value)
	}
	if n, _ := trees.LowestCommonAncestorBST(bst, 2, 8); n.Value != 6 {
		t.Errorf("BST LCA(2, 8) = %d", n.Value)
	}
}

func TestBuildSerializeRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 40; i++ {
		tr := &trees.Tree{}
		for _, v := range r.Perm(r.Intn(30)) {
			tr = insert(tr, v)
		}
		back := mustParse(t, trees.Serialize(tr))
		if trees.Serialize(back) != trees.Serialize(tr) || !reflect.DeepEqual(trees.Preorder(back), trees.Preorder(tr)) {
			t.Fatalf("round trip changed %q", trees.Serialize(tr))
		}
		rebuilt, err := trees.BuildFromPreIn(trees.Preorder(tr), trees.Inorder(tr))
		if err != nil || trees.Serialize(rebuilt) != trees.Serialize(tr) {
			t.Fatalf("BuildFromPreIn: %v", err)
		}
	}
	if _, err := trees.BuildFromPreIn([]int{1, 2}, []int{1, 1}); !errors.Is(err, trees.ErrInput) {
		t.Error("duplicate values accepted")
	}
	if _, err := trees.BuildFromPreIn([]int{1, 2}, []int{3, 1}); !errors.Is(err, trees.ErrInput) {
		t.Error("mismatched traversals accepted")
	}
	if _, err := trees.Parse("1 x"); !errors.Is(err, trees.ErrInput) {
		t.Error("a word was accepted as a value")
	}
}

func TestPathSumAndInvert(t *testing.T) {
	tr := mustParse(t, "5 4 8 11 null 13 4 7 2 null null 5 1")
	got := trees.PathSum(tr, 22)
	sort.Slice(got, func(i, j int) bool { return got[i][1] < got[j][1] })
	if !reflect.DeepEqual(got, [][]int{{5, 4, 11, 2}, {5, 8, 4, 5}}) {
		t.Errorf("PathSum = %v", got)
	}
	inv := trees.Invert(mustParse(t, "4 2 7 1 3 6 9"))
	if trees.Serialize(inv) != "4 7 2 9 6 3 1" {
		t.Errorf("Invert = %s", trees.Serialize(inv))
	}
	s := inv.Snapshot()
	if len(s.Nodes) != 7 || s.Nodes[0].Role != "root" || len(s.Edges) != 6 {
		t.Errorf("snapshot %+v", s)
	}
}

func TestExpressionTree(t *testing.T) {
	for _, c := range []struct {
		in, prefix, postfix string
		want                int
	}{
		{"1 + 2 * 3", "+ 1 * 2 3", "1 2 3 * +", 7},
		{"(1 + 2) * 3", "* + 1 2 3", "1 2 + 3 *", 9},
		{"10 - 4 - 3", "- - 10 4 3", "10 4 - 3 -", 3},
		{"8 / (3 - 1) * 2", "* / 8 - 3 1 2", "8 3 1 - / 2 *", 8},
		{"42", "42", "42", 42},
	} {
		e, err := trees.ParseExpr(c.in)
		if err != nil {
			t.Fatalf("ParseExpr(%q): %v", c.in, err)
		}
		v, err := e.Eval()
		if err != nil || v != c.want || e.Prefix() != c.prefix || e.Postfix() != c.postfix {
			t.Errorf("%q: %d %q %q (%v)", c.in, v, e.Prefix(), e.Postfix(), err)
		}
		if again, _ := trees.ParseExpr(e.Infix()); again.Postfix() != e.Postfix() {
			t.Errorf("%q: Infix does not parse back", c.in)
		}
	}
	for _, bad := range []string{"", "1 +", "(1", "1)", "1 2", "2(3)", "()", "1 & 2", "* 1"} {
		if _, err := trees.ParseExpr(bad); !errors.Is(err, trees.ErrInput) {
			t.Errorf("ParseExpr(%q) error %v", bad, err)
		}
	}
	e, _ := trees.ParseExpr("1 / (2 - 2)")
	if _, err := e.Eval(); !errors.Is(err, trees.ErrDivide) {
		t.Error("division by zero not reported")
	}
}
