package regression

import (
	"fmt"
	"math"
)

// Fit is a line learned by gradient descent, with the mean squared error
// after every epoch (one pass over the data), so the learning can be seen.
type Fit struct {
	Slope, Intercept float64
	Loss             []float64
}

// GradientDescent learns the line y = slope*x + intercept by starting at 0
// and repeatedly stepping both against the gradient of the mean squared
// error, scaled by rate. It reaches the same line as Linear when rate is
// small enough; too large a rate makes the loss grow instead, which the
// returned Loss shows (it stops early once the loss is not finite).
//
// Time: O(n * epochs).
// Space: O(epochs) for the loss curve.
func GradientDescent(x, y []float64, rate float64, epochs int) (Fit, error) {
	if len(x) == 0 || len(x) != len(y) {
		return Fit{}, fmt.Errorf("%w: want (x, y) pairs", ErrInput)
	}
	if rate <= 0 || epochs < 1 || epochs > 1_000_000 {
		return Fit{}, fmt.Errorf("%w: rate must be positive and epochs from 1 to 1,000,000", ErrInput)
	}
	var f Fit
	n := float64(len(x))
	for e := 0; e < epochs; e++ {
		var gs, gi float64
		for i := range x {
			err := f.Slope*x[i] + f.Intercept - y[i]
			gs += 2 * err * x[i] / n
			gi += 2 * err / n
		}
		f.Slope -= rate * gs
		f.Intercept -= rate * gi
		loss := 0.0
		for i := range x {
			d := f.Slope*x[i] + f.Intercept - y[i]
			loss += d * d / n
		}
		f.Loss = append(f.Loss, loss)
		if math.IsInf(loss, 0) || math.IsNaN(loss) {
			break
		}
	}
	return f, nil
}
