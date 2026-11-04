package numerical_test

import (
	"errors"
	"math"
	"math/cmplx"
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/math/numerical"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol*math.Max(1, math.Abs(b)) }

func TestHornerAndPoly(t *testing.T) {
	c := []float64{5, -3, 0, 2} // 2x^3 - 3x + 5
	for _, x := range []float64{-2, 0, 1.5, 3} {
		want := 2*x*x*x - 3*x + 5
		if got := numerical.Horner(c, x); !near(got, want, 1e-12) || numerical.Poly(c)(x) != got {
			t.Errorf("Horner at %v = %v, want %v", x, got, want)
		}
	}
	if numerical.Horner(nil, 3) != 0 {
		t.Error("empty polynomial")
	}
	if got := numerical.Derivative(c); !reflect.DeepEqual(got, []float64{-3, 0, 6}) {
		t.Errorf("Derivative = %v", got)
	}
}

func TestRoots(t *testing.T) {
	c := []float64{-2, 0, 1} // x^2 - 2
	f, df := numerical.Poly(c), numerical.Poly(numerical.Derivative(c))
	b, bs, err := numerical.Bisection(f, 0, 2, 1e-10, 200)
	if err != nil || !near(b, math.Sqrt2, 1e-9) {
		t.Errorf("Bisection = %v, %v", b, err)
	}
	n, ns, err := numerical.NewtonRaphson(f, df, 1, 1e-12, 50)
	if err != nil || !near(n, math.Sqrt2, 1e-12) {
		t.Errorf("Newton = %v, %v", n, err)
	}
	s, ss, err := numerical.Secant(f, 1, 2, 1e-12, 50)
	if err != nil || !near(s, math.Sqrt2, 1e-12) {
		t.Errorf("Secant = %v, %v", s, err)
	}
	if len(ns) >= len(ss) || len(ss) >= len(bs) {
		t.Errorf("steps: newton %d, secant %d, bisection %d", len(ns), len(ss), len(bs))
	}
	if _, _, err := numerical.Bisection(f, 2, 3, 1e-9, 100); !errors.Is(err, numerical.ErrBracket) {
		t.Errorf("Bisection without a bracket: %v", err)
	}
	if _, _, err := numerical.Bisection(f, 0, 2, 1e-30, 5); !errors.Is(err, numerical.ErrNoConvergence) {
		t.Errorf("Bisection out of steps: %v", err)
	}
	four := numerical.Poly([]float64{-4, 0, 1})
	if r, _, err := numerical.Bisection(four, 2, 3, 1e-9, 100); err != nil || r != 2 {
		t.Errorf("Bisection at an endpoint: %v", err)
	}
	if _, _, err := numerical.NewtonRaphson(f, df, 0, 1e-9, 50); !errors.Is(err, numerical.ErrNoConvergence) {
		t.Errorf("Newton at a zero derivative: %v", err)
	}
	if _, _, err := numerical.Secant(f, 1, -1, 1e-9, 50); !errors.Is(err, numerical.ErrNoConvergence) {
		t.Errorf("Secant on a flat secant: %v", err)
	}
	cube := numerical.Poly([]float64{-6, 11, -6, 1}) // (x-1)(x-2)(x-3)
	if r, _, _ := numerical.Bisection(cube, 2.5, 10, 1e-10, 200); !near(r, 3, 1e-9) {
		t.Errorf("cubic root = %v", r)
	}
}

func TestIntegration(t *testing.T) {
	cubic := numerical.Poly([]float64{1, 0, -3, 2}) // 2x^3 - 3x^2 + 1 on [0, 2]: 8 - 8 + 2 = 2
	if got := numerical.Simpson(cubic, 0, 2, 2); !near(got, 2, 1e-12) {
		t.Errorf("Simpson of a cubic = %v", got)
	}
	if got := numerical.Simpson(cubic, 0, 2, 3); !near(got, 2, 1e-12) {
		t.Errorf("Simpson with odd n = %v", got)
	}
	line := numerical.Poly([]float64{1, 2})
	if got := numerical.Trapezoid(line, 0, 3, 1); !near(got, 12, 1e-12) {
		t.Errorf("Trapezoid of a line = %v", got)
	}
	errT := math.Abs(numerical.Trapezoid(math.Sin, 0, math.Pi, 64) - 2)
	errS := math.Abs(numerical.Simpson(math.Sin, 0, math.Pi, 64) - 2)
	if errT > 1e-3 || errS > 1e-6 || errS >= errT {
		t.Errorf("sin errors: trapezoid %v, simpson %v", errT, errS)
	}
}

func TestGaussianElimination(t *testing.T) {
	a := [][]float64{{0, 2, 1}, {1, -2, -3}, {-1, 1, 2}} // zero first pivot needs a swap
	b := []float64{-8, 0, 3}
	x, steps, err := numerical.GaussianEliminationSteps(a, b)
	if err != nil || len(steps) == 0 || steps[0].Note != "swap rows 1 and 2" {
		t.Fatalf("steps %v, %v", steps, err)
	}
	for i, row := range a {
		sum := 0.0
		for j := range row {
			sum += row[j] * x[j]
		}
		if !near(sum, b[i], 1e-10) {
			t.Errorf("row %d: %v, want %v", i, sum, b[i])
		}
	}
	r := rand.New(rand.NewSource(1))
	for n := 1; n <= 8; n++ {
		m := make([][]float64, n)
		want := make([]float64, n)
		for i := range m {
			want[i] = float64(r.Intn(21) - 10)
			m[i] = make([]float64, n)
			for j := range m[i] {
				m[i][j] = r.Float64()*20 - 10
			}
		}
		rhs := make([]float64, n)
		for i := range m {
			for j := range m[i] {
				rhs[i] += m[i][j] * want[j]
			}
		}
		got, err := numerical.GaussianElimination(m, rhs)
		if err != nil {
			t.Fatal(err)
		}
		for i := range got {
			if !near(got[i], want[i], 1e-8) {
				t.Fatalf("n=%d: x = %v, want %v", n, got, want)
			}
		}
	}
	if _, err := numerical.GaussianElimination([][]float64{{1, 2}, {2, 4}}, []float64{1, 2}); !errors.Is(err, numerical.ErrSingular) {
		t.Errorf("singular: %v", err)
	}
	if _, err := numerical.GaussianElimination([][]float64{{1, 2}}, []float64{1}); !errors.Is(err, numerical.ErrShape) {
		t.Errorf("shape: %v", err)
	}
}

func TestMatrices(t *testing.T) {
	fib, err := numerical.MatrixPower([][]int{{1, 1}, {1, 0}}, 10)
	if err != nil || fib[0][1] != 55 {
		t.Errorf("Fibonacci matrix = %v, %v", fib, err)
	}
	if id, _ := numerical.MatrixPower([][]int{{2, 3}, {4, 5}}, 0); !reflect.DeepEqual(id, [][]int{{1, 0}, {0, 1}}) {
		t.Errorf("m^0 = %v", id)
	}
	r := rand.New(rand.NewSource(8))
	for n := 1; n <= 13; n++ {
		a, b := make([][]int, n), make([][]int, n)
		for i := range a {
			a[i], b[i] = make([]int, n), make([]int, n)
			for j := range a[i] {
				a[i][j], b[i][j] = r.Intn(19)-9, r.Intn(19)-9
			}
		}
		want, _ := numerical.Multiply(a, b)
		got, err := numerical.Strassen(a, b)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("Strassen n=%d differs", n)
		}
		p3, _ := numerical.MatrixPower(a, 3)
		ab, _ := numerical.Multiply(a, a)
		aaa, _ := numerical.Multiply(ab, a)
		if !reflect.DeepEqual(p3, aaa) {
			t.Fatalf("MatrixPower n=%d differs", n)
		}
	}
	for _, bad := range []error{
		func() error { _, err := numerical.Multiply([][]int{{1, 2}}, [][]int{{1, 2}}); return err }(),
		func() error { _, err := numerical.MatrixPower([][]int{{1, 2}}, 2); return err }(),
		func() error { _, err := numerical.MatrixPower([][]int{{1}}, -1); return err }(),
		func() error { _, err := numerical.Strassen([][]int{{1}}, [][]int{{1, 2}, {3, 4}}); return err }(),
	} {
		if !errors.Is(bad, numerical.ErrShape) {
			t.Errorf("error %v, want ErrShape", bad)
		}
	}
}

func TestFFT(t *testing.T) {
	r := rand.New(rand.NewSource(6))
	for _, n := range []int{1, 2, 4, 8, 64} {
		x := make([]complex128, n)
		for i := range x {
			x[i] = complex(r.Float64(), r.Float64())
		}
		got, err := numerical.FFT(x)
		if err != nil {
			t.Fatal(err)
		}
		for k := range x { // the naive O(n^2) transform
			var want complex128
			for j := range x {
				want += x[j] * cmplx.Rect(1, -2*math.Pi*float64(j*k)/float64(n))
			}
			if cmplx.Abs(got[k]-want) > 1e-9 {
				t.Fatalf("FFT n=%d at %d: %v, want %v", n, k, got[k], want)
			}
		}
		back, _ := numerical.InverseFFT(got)
		for i := range x {
			if cmplx.Abs(back[i]-x[i]) > 1e-12 {
				t.Fatalf("inverse FFT n=%d differs", n)
			}
		}
	}
	if _, err := numerical.FFT(make([]complex128, 3)); !errors.Is(err, numerical.ErrShape) {
		t.Errorf("FFT of 3: %v", err)
	}
	for la := 1; la <= 20; la += 3 {
		for lb := 1; lb <= 20; lb += 4 {
			a, b := make([]float64, la), make([]float64, lb)
			for i := range a {
				a[i] = float64(r.Intn(21) - 10)
			}
			for i := range b {
				b[i] = float64(r.Intn(21) - 10)
			}
			want := make([]float64, la+lb-1)
			for i := range a {
				for j := range b {
					want[i+j] += a[i] * b[j]
				}
			}
			got := numerical.PolyMultiply(a, b)
			for i := range want {
				if math.Abs(got[i]-want[i]) > 1e-9 {
					t.Fatalf("PolyMultiply %v * %v = %v, want %v", a, b, got, want)
				}
			}
		}
	}
	if numerical.PolyMultiply(nil, []float64{1}) != nil {
		t.Error("PolyMultiply of empty")
	}
}
