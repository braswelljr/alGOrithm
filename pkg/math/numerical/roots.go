package numerical

import "math"

// Bisection finds a root of f in [lo, hi], where f(lo) and f(hi) differ in
// sign, by halving the interval and keeping the half where the sign changes.
// It stops when the interval is narrower than tol or f is exactly zero, and
// returns the root and each midpoint tried. It is slow (one bit per step) but
// cannot fail once a root is bracketed: ErrBracket when it is not,
// ErrNoConvergence after maxIter steps.
//
// Time: O(log((hi - lo) / tol)) evaluations. Space: O(steps).
func Bisection(f func(float64) float64, lo, hi, tol float64, maxIter int) (float64, []float64, error) {
	flo, fhi := f(lo), f(hi)
	if flo == 0 {
		return lo, nil, nil
	}
	if fhi == 0 {
		return hi, nil, nil
	}
	if math.Signbit(flo) == math.Signbit(fhi) {
		return 0, nil, ErrBracket
	}
	var tried []float64
	for i := 0; i < maxIter; i++ {
		mid := lo + (hi-lo)/2
		tried = append(tried, mid)
		fm := f(mid)
		if fm == 0 || hi-lo < tol {
			return mid, tried, nil
		}
		if math.Signbit(fm) == math.Signbit(flo) {
			lo, flo = mid, fm
		} else {
			hi = mid
		}
	}
	return lo + (hi-lo)/2, tried, ErrNoConvergence
}

// NewtonRaphson finds a root of f from the guess x0 by following the tangent
// line: x <- x - f(x) / df(x), df being the derivative. Near a simple root
// the number of correct digits doubles each step, but a poor guess can
// wander off or cycle. It stops when a step is smaller than tol, and returns
// the root and each estimate. ErrNoConvergence after maxIter steps, on a
// zero derivative, or on a value that is not finite.
//
// Time: O(steps) evaluations. Space: O(steps).
func NewtonRaphson(f, df func(float64) float64, x0, tol float64, maxIter int) (float64, []float64, error) {
	x := x0
	tried := []float64{x}
	for i := 0; i < maxIter; i++ {
		d := df(x)
		if d == 0 {
			return x, tried, ErrNoConvergence
		}
		next := x - f(x)/d
		if math.IsNaN(next) || math.IsInf(next, 0) {
			return x, tried, ErrNoConvergence
		}
		tried = append(tried, next)
		if math.Abs(next-x) < tol {
			return next, tried, nil
		}
		x = next
	}
	return x, tried, ErrNoConvergence
}

// Secant finds a root of f from two guesses, like NewtonRaphson but with the
// derivative replaced by the slope through the last two points, so it needs
// no derivative. It converges with order about 1.618. It stops when a step is
// smaller than tol, and returns the root and each estimate (starting with x0
// and x1). ErrNoConvergence after maxIter steps or on a flat secant.
//
// Time: O(steps) evaluations. Space: O(steps).
func Secant(f func(float64) float64, x0, x1, tol float64, maxIter int) (float64, []float64, error) {
	tried := []float64{x0, x1}
	f0, f1 := f(x0), f(x1)
	for i := 0; i < maxIter; i++ {
		if f1 == f0 {
			return x1, tried, ErrNoConvergence
		}
		next := x1 - f1*(x1-x0)/(f1-f0)
		tried = append(tried, next)
		if math.Abs(next-x1) < tol {
			return next, tried, nil
		}
		x0, f0 = x1, f1
		x1, f1 = next, f(next)
	}
	return x1, tried, ErrNoConvergence
}
