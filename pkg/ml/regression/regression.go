// Package regression fits models that predict a number from examples
// (supervised learning), one method per file: simple and multiple linear
// regression in closed form, linear regression by gradient descent,
// polynomial regression, logistic regression for yes-or-no answers, and the
// scores that say how good a fit is (mean squared error and R^2).
//
// A sample is a row of features xs[i] and a target ys[i]. Nothing here
// changes its input. Every method is deterministic.
package regression

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ErrInput is returned for data that is missing, of mixed width, too small
// for the model, or settings out of range.
var ErrInput = errors.New("regression: invalid input")

// MaxSamples is the most samples Parse accepts.
const MaxSamples = 5_000

// Parse reads one sample a line: the features, then the target, separated by
// spaces, tabs or commas ("x y" for one feature). Blank lines and lines
// starting with # are skipped.
func Parse(text string) (xs [][]float64, ys []float64, err error) {
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) < 2 {
			return nil, nil, fmt.Errorf("%w: line %d: want features then a target", ErrInput, i+1)
		}
		row := make([]float64, len(f))
		for j, s := range f {
			v, perr := strconv.ParseFloat(s, 64)
			if perr != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, nil, fmt.Errorf("%w: line %d: %q is not a number", ErrInput, i+1, s)
			}
			row[j] = v
		}
		if len(xs) > 0 && len(row)-1 != len(xs[0]) {
			return nil, nil, fmt.Errorf("%w: line %d has %d features, the first has %d", ErrInput, i+1, len(row)-1, len(xs[0]))
		}
		xs, ys = append(xs, row[:len(row)-1]), append(ys, row[len(row)-1])
		if len(xs) > MaxSamples {
			return nil, nil, fmt.Errorf("%w: more than %d samples", ErrInput, MaxSamples)
		}
	}
	if len(xs) == 0 {
		return nil, nil, fmt.Errorf("%w: no samples", ErrInput)
	}
	return xs, ys, nil
}

// Column returns feature j of every row, for the one-feature methods.
func Column(xs [][]float64, j int) []float64 {
	out := make([]float64, len(xs))
	for i, row := range xs {
		out[i] = row[j]
	}
	return out
}

func checkXY(xs [][]float64, ys []float64) error {
	if len(xs) == 0 || len(xs) != len(ys) {
		return fmt.Errorf("%w: %d rows and %d targets", ErrInput, len(xs), len(ys))
	}
	for i, row := range xs {
		if len(row) != len(xs[0]) {
			return fmt.Errorf("%w: row %d has %d features, the first has %d", ErrInput, i, len(row), len(xs[0]))
		}
	}
	return nil
}
