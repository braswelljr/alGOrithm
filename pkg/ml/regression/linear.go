package regression

import "fmt"

// Linear fits the line y = slope*x + intercept that makes the squared
// errors smallest (ordinary least squares, Legendre and Gauss, about 1805):
// slope = cov(x, y) / var(x), and the line goes through the means.
//
// Time: O(n).
// Space: O(1).
func Linear(x, y []float64) (slope, intercept float64, err error) {
	if len(x) < 2 || len(x) != len(y) {
		return 0, 0, fmt.Errorf("%w: want at least two (x, y) pairs", ErrInput)
	}
	mx, my := mean(x), mean(y)
	var sxy, sxx float64
	for i := range x {
		sxy += (x[i] - mx) * (y[i] - my)
		sxx += (x[i] - mx) * (x[i] - mx)
	}
	if sxx == 0 {
		return 0, 0, fmt.Errorf("%w: every x is the same, so no line fits better than another", ErrInput)
	}
	slope = sxy / sxx
	return slope, my - slope*mx, nil
}

// Multiple fits y = c[0] + c[1]*x1 + ... + c[d]*xd by least squares, solving
// the normal equations (X^T X) c = X^T y with Gaussian elimination, where X
// has a leading column of ones. It returns the d+1 coefficients, intercept
// first.
//
// Time: O(n * d^2 + d^3).
// Space: O(d^2).
func Multiple(xs [][]float64, ys []float64) ([]float64, error) {
	if err := checkXY(xs, ys); err != nil {
		return nil, err
	}
	d := len(xs[0]) + 1
	if len(xs) < d {
		return nil, fmt.Errorf("%w: %d coefficients need at least %d samples", ErrInput, d, d)
	}
	a := make([][]float64, d)
	for i := range a {
		a[i] = make([]float64, d+1)
	}
	for k, row := range xs {
		v := append([]float64{1}, row...)
		for i := 0; i < d; i++ {
			for j := 0; j < d; j++ {
				a[i][j] += v[i] * v[j]
			}
			a[i][d] += v[i] * ys[k] //nolint:gosec // checkXY made len(ys) == len(xs)
		}
	}
	return solve(a)
}

// solve does Gaussian elimination with partial pivoting on the augmented
// matrix a (d rows, d+1 columns), changing a.
func solve(a [][]float64) ([]float64, error) {
	d := len(a)
	for c := 0; c < d; c++ {
		p := c
		for r := c + 1; r < d; r++ {
			if abs(a[r][c]) > abs(a[p][c]) {
				p = r
			}
		}
		if abs(a[p][c]) < 1e-12 {
			return nil, fmt.Errorf("%w: the features are linearly dependent, so the fit is not unique", ErrInput)
		}
		a[c], a[p] = a[p], a[c]
		for r := 0; r < d; r++ {
			if r == c {
				continue
			}
			f := a[r][c] / a[c][c]
			for k := c; k <= d; k++ {
				a[r][k] -= f * a[c][k]
			}
		}
	}
	out := make([]float64, d)
	for i := range out {
		out[i] = a[i][d] / a[i][i]
	}
	return out, nil
}

// Polynomial fits y = c[0] + c[1]*x + ... + c[degree]*x^degree by least
// squares: multiple regression on the features x, x^2, ..., x^degree. A high
// degree fits the samples better and predicts new ones worse (overfitting).
//
// Time: O(n * degree^2 + degree^3).
// Space: O(n * degree).
func Polynomial(x, y []float64, degree int) ([]float64, error) {
	if degree < 1 || degree > 10 {
		return nil, fmt.Errorf("%w: degree must be from 1 to 10", ErrInput)
	}
	xs := make([][]float64, len(x))
	for i, v := range x {
		row := make([]float64, degree)
		p := 1.0
		for j := range row {
			p *= v
			row[j] = p
		}
		xs[i] = row
	}
	return Multiple(xs, y)
}

// Eval returns the value of the polynomial c[0] + c[1]*x + ... at x, by
// Horner's rule.
func Eval(c []float64, x float64) float64 {
	v := 0.0
	for i := len(c) - 1; i >= 0; i-- {
		v = v*x + c[i]
	}
	return v
}

func mean(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
