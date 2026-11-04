// Package classification learns to put examples into classes from labeled
// ones (supervised learning), one method per file: k-nearest neighbors,
// Gaussian naive Bayes, naive Bayes on words, the perceptron, decision
// trees, and the tools to judge a classifier honestly (confusion matrix,
// precision, recall, F1, train/test split and k-fold cross-validation).
//
// An example is a row of numeric features with a string label. Nothing here
// changes its input; anything random takes a seed.
//
// TraceKNN reports Compare for each training example measured against the
// query and Mark "group" with the k nearest (Values[0] 0).
package classification

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ErrInput is returned for data that is missing, of mixed width, or settings
// out of range.
var ErrInput = errors.New("classification: invalid input")

// MaxExamples is the most examples Parse accepts.
const MaxExamples = 5_000

// Parse reads one labeled example a line: the features, then the label,
// separated by spaces, tabs or commas ("1.5 2 cat"). Blank lines and lines
// starting with # are skipped.
func Parse(text string) (xs [][]float64, labels []string, err error) {
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) < 2 {
			return nil, nil, fmt.Errorf("%w: line %d: want features then a label", ErrInput, i+1)
		}
		row, perr := numbers(f[:len(f)-1])
		if perr != nil {
			return nil, nil, fmt.Errorf("%w: line %d: %w", ErrInput, i+1, perr)
		}
		if len(xs) > 0 && len(row) != len(xs[0]) {
			return nil, nil, fmt.Errorf("%w: line %d has %d features, the first has %d", ErrInput, i+1, len(row), len(xs[0]))
		}
		xs, labels = append(xs, row), append(labels, f[len(f)-1])
		if len(xs) > MaxExamples {
			return nil, nil, fmt.Errorf("%w: more than %d examples", ErrInput, MaxExamples)
		}
	}
	if len(xs) == 0 {
		return nil, nil, fmt.Errorf("%w: no examples", ErrInput)
	}
	return xs, labels, nil
}

// ParseRow reads one unlabeled row of features ("1.5 2").
func ParseRow(text string) ([]float64, error) {
	row, err := numbers(strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' }))
	if err != nil || len(row) == 0 {
		return nil, fmt.Errorf("%w: want a row of numbers", ErrInput)
	}
	return row, nil
}

func numbers(f []string) ([]float64, error) {
	row := make([]float64, len(f))
	for j, s := range f {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("%q is not a number", s)
		}
		row[j] = v
	}
	return row, nil
}

func check(xs [][]float64, labels []string) error {
	if len(xs) == 0 || len(xs) != len(labels) {
		return fmt.Errorf("%w: %d rows and %d labels", ErrInput, len(xs), len(labels))
	}
	for i, r := range xs {
		if len(r) == 0 || len(r) != len(xs[0]) {
			return fmt.Errorf("%w: row %d has %d features, the first has %d", ErrInput, i, len(r), len(xs[0]))
		}
	}
	return nil
}

// Classes returns the distinct labels, sorted.
func Classes(labels []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range labels {
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return out
}

func distance(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s)
}

// vote returns the most common label in ls, breaking ties by the label that
// comes first in sorted order.
func vote(ls []string) string {
	count := map[string]int{}
	for _, l := range ls {
		count[l]++
	}
	best := ""
	for _, c := range Classes(ls) {
		if count[c] > count[best] || best == "" {
			best = c
		}
	}
	return best
}
