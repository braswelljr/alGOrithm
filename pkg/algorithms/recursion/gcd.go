package recursion

// GCD returns the greatest common divisor of a and b with Euclid's algorithm
// (about 300 BC): gcd(a, b) = gcd(b, a mod b), and gcd(a, 0) = a. Signs are
// ignored; GCD(0, 0) is 0.
//
// Time: O(log min(a, b)). Space: O(log min(a, b)) for the recursion.
func GCD(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	if b == 0 {
		return a
	}
	return GCD(b, a%b)
}

// LCM returns the least common multiple of a and b, |a*b| / gcd(a, b), dividing
// first so it overflows only when the result does. LCM with 0 is 0.
//
// Time: O(log min(a, b)). Space: O(log min(a, b)).
func LCM(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	l := a / GCD(a, b) * b
	if l < 0 {
		return -l
	}
	return l
}
