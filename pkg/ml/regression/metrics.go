package regression

import "fmt"

// MSE is the mean squared error between the targets ys and the predictions
// preds: the average of (y - prediction)^2. Squaring punishes large misses
// more than small ones.
//
// Time: O(n).
// Space: O(1).
func MSE(ys, preds []float64) (float64, error) {
	if len(ys) == 0 || len(ys) != len(preds) {
		return 0, fmt.Errorf("%w: %d targets and %d predictions", ErrInput, len(ys), len(preds))
	}
	s := 0.0
	for i := range ys {
		d := ys[i] - preds[i]
		s += d * d
	}
	return s / float64(len(ys)), nil
}

// R2 is the coefficient of determination: 1 - SSres / SStot, the share of
// the targets' variance the predictions explain. 1 is a perfect fit, 0 is no
// better than always predicting the mean, and it can be negative for a fit
// worse than that. Constant targets give 1 if predicted exactly, else 0.
//
// Time: O(n).
// Space: O(1).
func R2(ys, preds []float64) (float64, error) {
	mse, err := MSE(ys, preds)
	if err != nil {
		return 0, err
	}
	m := mean(ys)
	tot := 0.0
	for _, y := range ys {
		tot += (y - m) * (y - m)
	}
	res := mse * float64(len(ys))
	if tot == 0 {
		if res == 0 {
			return 1, nil
		}
		return 0, nil
	}
	return 1 - res/tot, nil
}
