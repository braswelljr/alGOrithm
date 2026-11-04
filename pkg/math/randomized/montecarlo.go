package randomized

import "math/rand"

// MaxPoints is the most points MonteCarloPi throws.
const MaxPoints = 10_000_000

// Sample is one point of MonteCarloPi in the unit square, and whether it fell
// inside the quarter circle of radius 1.
type Sample struct {
	X, Y   float64
	Inside bool
}

// MonteCarloPi estimates pi by throwing n random points into the unit square
// and counting those inside the quarter circle x^2 + y^2 <= 1, whose area is
// pi/4: pi is about 4 * inside / n. A Monte Carlo method: the error shrinks
// only like 1/sqrt(n), so each extra digit costs 100 times the points. It
// returns the estimate and the first 2000 samples (to draw). 0 and nil when n
// is below 1 or above MaxPoints.
//
// Time: O(n). Space: O(1) beyond the samples kept.
func MonteCarloPi(n int, seed int64) (float64, []Sample) {
	if n < 1 || n > MaxPoints {
		return 0, nil
	}
	rng := rand.New(rand.NewSource(seed))
	inside := 0
	var kept []Sample
	for i := 0; i < n; i++ {
		x, y := rng.Float64(), rng.Float64()
		in := x*x+y*y <= 1
		if in {
			inside++
		}
		if i < 2000 {
			kept = append(kept, Sample{x, y, in})
		}
	}
	return 4 * float64(inside) / float64(n), kept
}
