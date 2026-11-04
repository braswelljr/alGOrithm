package regression

import (
	"errors"
	"math"
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
