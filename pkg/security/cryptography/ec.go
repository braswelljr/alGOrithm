package cryptography

import "fmt"

// Curve is the elliptic curve y^2 = x^3 + Ax + B over the integers mod the
// prime P. Real curves (P-256, Curve25519) use 256-bit primes; a small P lets
// every point be listed. For learning: use crypto/ecdh and crypto/ecdsa in
// real code.
type Curve struct {
	A, B, P int64
}

// ECPoint is a point on a curve, or the point at infinity (the identity) when
// Inf is set.
type ECPoint struct {
	X   int64 `json:"x"`
	Y   int64 `json:"y"`
	Inf bool  `json:"inf,omitempty"`
}

// String writes the point as "(x, y)" or "O" for infinity.
func (p ECPoint) String() string {
	if p.Inf {
		return "O"
	}
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func (c Curve) mod(v int64) int64 { return (v%c.P + c.P) % c.P }

// OnCurve reports whether p satisfies the curve's equation.
func (c Curve) OnCurve(p ECPoint) bool {
	return p.Inf || c.mod(p.Y*p.Y) == c.mod(c.mod(p.X*p.X)*p.X+c.A*p.X+c.B)
}

// Add returns p + q: draw the line through them (the tangent when they are
// equal), find its third point on the curve and reflect it over the x-axis.
// A point plus its reflection is infinity.
//
// Time: O(log P) for the inverse. Space: O(1).
func (c Curve) Add(p, q ECPoint) ECPoint {
	switch {
	case p.Inf:
		return q
	case q.Inf:
		return p
	case p.X == q.X && c.mod(p.Y+q.Y) == 0:
		return ECPoint{Inf: true}
	}
	var slope int64
	if p.X == q.X { // doubling: the tangent's slope (3x^2 + A) / 2y
		slope = c.mod((3*p.X%c.P*p.X + c.A) % c.P * ModPow(2*p.Y, c.P-2, c.P))
	} else {
		slope = c.mod(c.mod(q.Y-p.Y) * ModPow(c.mod(q.X-p.X), c.P-2, c.P))
	}
	x := c.mod(slope*slope - p.X - q.X)
	return ECPoint{X: x, Y: c.mod(slope*c.mod(p.X-x) - p.Y)}
}

// Double returns p + p.
func (c Curve) Double(p ECPoint) ECPoint { return c.Add(p, p) }

// ScalarMult returns k*p by double-and-add, one doubling per bit of k and an
// addition per 1 bit: computing k*p is fast, finding k from k*p (the
// elliptic-curve discrete log) is what is believed hard. steps says what
// happened at each bit, most significant first.
//
// Time: O(log k). Space: O(log k) for the steps.
func (c Curve) ScalarMult(k int64, p ECPoint) (result ECPoint, steps []string) {
	result = ECPoint{Inf: true}
	if k < 0 {
		k, p = -k, ECPoint{X: p.X, Y: c.mod(-p.Y), Inf: p.Inf}
	}
	for bit := 62; bit >= 0; bit-- {
		if k>>bit == 0 {
			continue
		}
		result = c.Double(result)
		line := fmt.Sprintf("bit %d: double to %v", bit, result)
		if k>>bit&1 == 1 {
			result = c.Add(result, p)
			line += fmt.Sprintf(", add P to %v", result)
		}
		steps = append(steps, line)
	}
	return result, steps
}

// Points lists every point on the curve, infinity first, by trying every x
// and y: O(P^2), fine for the small curves it is meant for.
func (c Curve) Points() []ECPoint {
	pts := []ECPoint{{Inf: true}}
	for x := int64(0); x < c.P; x++ {
		for y := int64(0); y < c.P; y++ {
			if p := (ECPoint{X: x, Y: y}); c.OnCurve(p) {
				pts = append(pts, p)
			}
		}
	}
	return pts
}

// Order returns the smallest n > 0 with n*p = infinity.
func (c Curve) Order(p ECPoint) int64 {
	q, n := p, int64(1)
	for !q.Inf {
		q, n = c.Add(q, p), n+1
	}
	return n
}
