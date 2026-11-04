// Package numerical holds numerical methods, one file each: root finding
// (bisection, Newton-Raphson, the secant method), integration (the
// trapezoid rule, Simpson's rule), linear systems (Gaussian elimination with
// partial pivoting), matrix products (repeated squaring, Strassen), the fast
// Fourier transform, and Horner's rule.
//
// A polynomial is its coefficients from the constant term up, so
// []float64{-2, 0, 1} is x^2 - 2; Horner evaluates one, and Poly turns one
// into a function for the root finders and integrators:
//
//	f := numerical.Poly([]float64{-2, 0, 1})
//	root, _, _ := numerical.Bisection(f, 0, 2, 1e-9, 100) // 1.41421356...
//
// Iterative methods return every estimate they made, so convergence can be
// watched: bisection gains one bit per step, Newton's method doubles its
// correct digits, and the secant method sits between them.
package numerical

import "errors"

var (
	// ErrBracket is returned by Bisection when f(lo) and f(hi) have the
	// same sign, so no root is bracketed.
	ErrBracket = errors.New("numerical: interval does not bracket a root")
	// ErrNoConvergence is returned when an iteration runs out of steps, or
	// meets a zero derivative or a flat secant.
	ErrNoConvergence = errors.New("numerical: did not converge")
	// ErrSingular is returned by GaussianElimination for a matrix with no
	// unique solution.
	ErrSingular = errors.New("numerical: matrix is singular")
	// ErrShape is returned for matrices or vectors of the wrong sizes.
	ErrShape = errors.New("numerical: wrong shape")
)

// Horner returns the polynomial coeffs (constant term first) at x with
// Horner's rule: a + x(b + x(c + ...)), one multiplication and one addition
// per coefficient and no powers. 0 for no coefficients.
//
// Time: O(n). Space: O(1).
func Horner(coeffs []float64, x float64) float64 {
	y := 0.0
	for i := len(coeffs) - 1; i >= 0; i-- {
		y = y*x + coeffs[i]
	}
	return y
}

// Poly returns the polynomial coeffs (constant term first) as a function,
// evaluated with Horner.
func Poly(coeffs []float64) func(float64) float64 {
	c := append([]float64(nil), coeffs...)
	return func(x float64) float64 { return Horner(c, x) }
}

// Derivative returns the coefficients of the derivative of the polynomial
// coeffs, for NewtonRaphson.
func Derivative(coeffs []float64) []float64 {
	if len(coeffs) < 2 {
		return []float64{0}
	}
	d := make([]float64, len(coeffs)-1)
	for i := range d {
		d[i] = float64(i+1) * coeffs[i+1]
	}
	return d
}
