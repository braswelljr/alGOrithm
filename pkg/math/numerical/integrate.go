package numerical

// Trapezoid approximates the integral of f over [a, b] with the trapezoid
// rule on n equal slices: join neighboring points with straight lines and
// add the areas. The error shrinks with the square of the slice width, and
// it is exact for straight lines. n below 1 means 1.
//
// Time: O(n) evaluations. Space: O(1).
func Trapezoid(f func(float64) float64, a, b float64, n int) float64 {
	n = max(n, 1)
	h := (b - a) / float64(n)
	sum := (f(a) + f(b)) / 2
	for i := 1; i < n; i++ {
		sum += f(a + float64(i)*h)
	}
	return sum * h
}

// Simpson approximates the integral of f over [a, b] with Simpson's rule on
// n slices: fit a parabola through each pair of slices, weighting the points
// 1, 4, 2, 4, ..., 4, 1. The error shrinks with the fourth power of the slice
// width, and it is exact for cubics. n is rounded up to an even number of at
// least 2.
//
// Time: O(n) evaluations. Space: O(1).
func Simpson(f func(float64) float64, a, b float64, n int) float64 {
	n = max(n, 2)
	n += n % 2
	h := (b - a) / float64(n)
	sum := f(a) + f(b)
	for i := 1; i < n; i++ {
		w := 2.0
		if i%2 == 1 {
			w = 4
		}
		sum += w * f(a+float64(i)*h)
	}
	return sum * h / 3
}
