package classification

import (
	"errors"
	"math"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

var (
	twoX = [][]float64{{1, 1}, {1, 2}, {2, 1}, {2, 2}, {8, 8}, {8, 9}, {9, 8}, {9, 9}}
	twoY = []string{"a", "a", "a", "a", "b", "b", "b", "b"}
)

func TestKNN(t *testing.T) {
	var log trace.Log
	l, near, err := TraceKNN(twoX, twoY, []float64{2, 3}, 3, &log)
	if err != nil || l != "a" || len(near) != 3 {
		t.Fatalf("%s %v %v", l, near, err)
	}
	if log.Count(trace.Compare) != len(twoX) {
		t.Error("compare events")
	}
	if l, _, _ := KNN(twoX, twoY, []float64{7, 7}, 1); l != "b" {
		t.Errorf("got %s", l)
	}
	if _, _, err := KNN(twoX, twoY, []float64{1}, 1); !errors.Is(err, ErrInput) {
		t.Error("short query accepted")
	}
}

func TestGaussianNB(t *testing.T) {
	m, err := FitGaussianNB(twoX, twoY)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := m.Predict([]float64{1.5, 1.2}); l != "a" {
		t.Errorf("got %s", l)
	}
	if l, _ := m.Predict([]float64{7.5, 9.5}); l != "b" {
		t.Errorf("got %s", l)
	}
}

func TestTextNB(t *testing.T) {
	docs, labels, err := ParseDocs("spam: win money now\nspam: free money offer\nham: lunch with the team\nham: notes from the meeting")
	if err != nil {
		t.Fatal(err)
	}
	m, _ := FitTextNB(docs, labels)
	if l, _ := m.Predict("free money"); l != "spam" {
		t.Errorf("got %s", l)
	}
	if l, _ := m.Predict("team meeting notes"); l != "ham" {
		t.Errorf("got %s", l)
	}
}

func TestPerceptron(t *testing.T) {
	ys := []int{-1, -1, -1, -1, 1, 1, 1, 1}
	p, err := FitPerceptron(twoX, ys, 100)
	if err != nil {
		t.Fatal(err)
	}
	if p.Mistakes[len(p.Mistakes)-1] != 0 {
		t.Error("separable data not learned")
	}
	xor, _ := FitPerceptron([][]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, []int{-1, 1, 1, -1}, 50)
	if xor.Mistakes[len(xor.Mistakes)-1] == 0 {
		t.Error("a perceptron cannot learn XOR")
	}
}

func TestTree(t *testing.T) {
	xs := [][]float64{{1}, {2}, {3}, {4}, {5}, {6}}
	ys := []string{"a", "a", "b", "b", "a", "a"} // needs two splits
	n, err := BuildTree(xs, ys, 3)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range xs {
		if n.Predict(r) != ys[i] {
			t.Errorf("%v: %s", r, n.Predict(r))
		}
	}
	if len(n.Lines()) != 5 {
		t.Errorf("lines %v", n.Lines())
	}
	xor, _ := BuildTree([][]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, []string{"no", "yes", "yes", "no"}, 3)
	if !xor.Leaf() {
		t.Error("no single split helps on XOR, so greedy CART stops at the root")
	}
	stump, _ := BuildTree(twoX, twoY, 1)
	if stump.Leaf() || !stump.Left.Leaf() || stump.Predict([]float64{9, 9}) != "b" {
		t.Error("one split separates the two groups")
	}
}

func TestEvaluate(t *testing.T) {
	c, err := ConfusionMatrix([]string{"a", "a", "b", "b"}, []string{"a", "b", "b", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if c.M[0][0] != 1 || c.M[0][1] != 1 || c.M[1][1] != 2 {
		t.Fatalf("%v", c.M)
	}
	s, acc := c.Scores()
	if acc != 0.75 || s[0].Precision != 1 || s[0].Recall != 0.5 || math.Abs(s[1].Precision-2.0/3) > 1e-12 {
		t.Errorf("%+v %v", s, acc)
	}
	train, test, err := TrainTestSplit(10, 0.3, 1)
	if err != nil || len(train) != 7 || len(test) != 3 {
		t.Errorf("%v %v %v", train, test, err)
	}
	folds, _ := KFold(10, 3, 1)
	seen := map[int]bool{}
	for _, f := range folds {
		for _, i := range f {
			seen[i] = true
		}
	}
	if len(folds) != 3 || len(seen) != 10 {
		t.Errorf("folds %v", folds)
	}
	acc2, err := CrossValidateKNN(twoX, twoY, 4, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acc2 {
		if a != 1 {
			t.Errorf("accuracy %v", acc2)
		}
	}
}

func TestParse(t *testing.T) {
	xs, ls, err := Parse("# x y class\n1 2 cat\n3,4,dog")
	if err != nil || len(xs) != 2 || ls[1] != "dog" {
		t.Fatalf("%v %v %v", xs, ls, err)
	}
	if _, err := ParseRow("1 x"); !errors.Is(err, ErrInput) {
		t.Error("bad row accepted")
	}
}
