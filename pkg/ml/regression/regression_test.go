package regression

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func near(t *testing.T, what string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

func TestLinear(t *testing.T) {
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{1, 3, 5, 7, 9} // y = 2x + 1 exactly
	s, i, err := Linear(x, y)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "slope", s, 2, 1e-12)
	near(t, "intercept", i, 1, 1e-12)
	if _, _, err := Linear([]float64{1, 1}, []float64{1, 2}); !errors.Is(err, ErrInput) {
		t.Error("vertical data accepted")
	}
}

func TestMultipleAndPolynomial(t *testing.T) {
	xs := [][]float64{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 3}}
	var ys []float64
	for _, r := range xs {
		ys = append(ys, 3+2*r[0]-r[1]) // y = 3 + 2a - b
	}
	c, err := Multiple(xs, ys)
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range []float64{3, 2, -1} {
		near(t, "coefficient", c[k], want, 1e-9)
	}
	x := []float64{-2, -1, 0, 1, 2, 3}
	var y []float64
	for _, v := range x {
		y = append(y, 1-v+0.5*v*v)
	}
	p, err := Polynomial(x, y, 2)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "c0", p[0], 1, 1e-9)
	near(t, "c1", p[1], -1, 1e-9)
	near(t, "c2", p[2], 0.5, 1e-9)
	near(t, "eval", Eval(p, 4), 5, 1e-9)
	if _, err := Multiple([][]float64{{1, 2}, {2, 4}, {3, 6}}, []float64{1, 2, 3}); !errors.Is(err, ErrInput) {
		t.Error("dependent features accepted")
	}
}

func TestGradientDescent(t *testing.T) {
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{1, 3, 5, 7, 9}
	f, err := GradientDescent(x, y, 0.05, 2000)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "slope", f.Slope, 2, 1e-6)
	near(t, "intercept", f.Intercept, 1, 1e-6)
	if f.Loss[len(f.Loss)-1] > f.Loss[0] {
		t.Error("the loss grew")
	}
	big, _ := GradientDescent(x, y, 1, 50)
	if big.Loss[len(big.Loss)-1] < big.Loss[0] {
		t.Error("a too-large rate should diverge")
	}
}

func TestLogistic(t *testing.T) {
	xs := [][]float64{{1}, {2}, {3}, {4}, {6}, {7}, {8}, {9}}
	ys := []int{0, 0, 0, 0, 1, 1, 1, 1}
	m, err := FitLogistic(xs, ys, 0.5, 2000)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range xs {
		if m.Predict(r) != ys[i] {
			t.Errorf("x = %v predicted %d", r, m.Predict(r))
		}
	}
	near(t, "boundary", m.Probability([]float64{5}), 0.5, 0.1) // still converging on separable data
	if _, err := FitLogistic(xs, []int{0, 1, 2, 0, 0, 0, 0, 0}, 0.1, 10); !errors.Is(err, ErrInput) {
		t.Error("label 2 accepted")
	}
}

func TestMetrics(t *testing.T) {
	ys := []float64{1, 2, 3, 4}
	mse, _ := MSE(ys, []float64{1, 2, 3, 5})
	near(t, "mse", mse, 0.25, 1e-12)
	r2, _ := R2(ys, ys)
	near(t, "perfect r2", r2, 1, 1e-12)
	r2, _ = R2(ys, []float64{2.5, 2.5, 2.5, 2.5})
	near(t, "mean r2", r2, 0, 1e-12)
}

func TestParse(t *testing.T) {
	xs, ys, err := Parse("# x y\n1 2\n3, 4\n")
	if err != nil || len(xs) != 2 || ys[1] != 4 || xs[1][0] != 3 {
		t.Fatalf("%v %v %v", xs, ys, err)
	}
	if _, _, err := Parse("1"); !errors.Is(err, ErrInput) {
		t.Error("a lone number accepted")
	}
}

func TestParseErrorsAndColumn(t *testing.T) {
	xs, ys, err := Parse("# a, b, y\n1,2,3\n4\t5 6\n\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(xs) != 2 || ys[1] != 6 {
		t.Fatalf("xs %v ys %v", xs, ys)
	}
	if got := Column(xs, 1); len(got) != 2 || got[0] != 2 || got[1] != 5 {
		t.Errorf("column 1 = %v, want [2 5]", got)
	}
	big := strings.Repeat("1 2\n", MaxSamples+1)
	for name, text := range map[string]string{
		"one field":      "1",
		"not a number":   "1 x",
		"infinity":       "1 Inf",
		"not a number 2": "NaN 1",
		"mixed widths":   "1 2\n1 2 3",
		"nothing":        "# only a comment\n",
		"too many":       big,
	} {
		if _, _, err := Parse(text); !errors.Is(err, ErrInput) {
			t.Errorf("%s: err %v", name, err)
		}
	}
}

func TestMetricsEdges(t *testing.T) {
	if r, _ := R2([]float64{2, 2, 2}, []float64{2, 2, 2}); r != 1 {
		t.Errorf("constant targets predicted exactly: R2 %v, want 1", r)
	}
	if r, _ := R2([]float64{2, 2, 2}, []float64{2, 3, 2}); r != 0 {
		t.Errorf("constant targets missed: R2 %v, want 0", r)
	}
	if r, _ := R2([]float64{1, 2, 3}, []float64{3, 2, 1}); r != -3 {
		t.Errorf("a reversed fit: R2 %v, want 1 - 8/2 = -3", r)
	}
	if _, err := R2(nil, nil); !errors.Is(err, ErrInput) {
		t.Errorf("empty R2: %v", err)
	}
	if _, err := MSE([]float64{1}, []float64{1, 2}); !errors.Is(err, ErrInput) {
		t.Errorf("mismatched MSE: %v", err)
	}
}

func TestBinaryAndLogisticErrors(t *testing.T) {
	labels, err := Binary([]float64{0, 1, 1, 0})
	if err != nil || len(labels) != 4 || labels[1] != 1 || labels[3] != 0 {
		t.Errorf("Binary = %v %v", labels, err)
	}
	if _, err := Binary([]float64{0, 0.5}); !errors.Is(err, ErrInput) {
		t.Errorf("a target of 0.5: %v", err)
	}
	xs := [][]float64{{1}, {2}}
	for name, f := range map[string]func() error{
		"no rows":         func() error { _, err := FitLogistic(nil, nil, 0.1, 10); return err },
		"label 2":         func() error { _, err := FitLogistic(xs, []int{0, 2}, 0.1, 10); return err },
		"mixed widths":    func() error { _, err := FitLogistic([][]float64{{1}, {1, 2}}, []int{0, 1}, 0.1, 10); return err },
		"zero rate":       func() error { _, err := FitLogistic(xs, []int{0, 1}, 0, 10); return err },
		"no epochs":       func() error { _, err := FitLogistic(xs, []int{0, 1}, 0.1, 0); return err },
		"gd no pairs":     func() error { _, err := GradientDescent(nil, nil, 0.1, 10); return err },
		"gd zero rate":    func() error { _, err := GradientDescent([]float64{1}, []float64{1}, 0, 10); return err },
		"multiple widths": func() error { _, err := Multiple([][]float64{{1}, {1, 2}}, []float64{1, 2}); return err },
		"multiple short":  func() error { _, err := Multiple([][]float64{{1, 2}}, []float64{1}); return err },
		"multiple counts": func() error { _, err := Multiple([][]float64{{1}}, []float64{1, 2}); return err },
	} {
		if err := f(); !errors.Is(err, ErrInput) {
			t.Errorf("%s: err %v", name, err)
		}
	}
}

func TestGradientDescentDiverges(t *testing.T) {
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{1, 3, 5, 7, 9}
	f, err := GradientDescent(x, y, 5, 10_000)
	if err != nil {
		t.Fatal(err)
	}
	last := f.Loss[len(f.Loss)-1]
	if len(f.Loss) == 10_000 || !(math.IsInf(last, 0) || math.IsNaN(last)) {
		t.Errorf("a rate of 5 should blow up and stop early: %d epochs, last loss %v", len(f.Loss), last)
	}
	for i := 1; i < 5; i++ {
		if f.Loss[i] <= f.Loss[i-1] {
			t.Errorf("diverging loss should grow: %v", f.Loss[:5])
		}
	}
}
