package classification

import (
	"fmt"
	"math/rand"
)

// Confusion counts predictions against the truth: M[i][j] is how many
// examples of class Classes[i] were predicted as Classes[j]. The diagonal is
// what was right.
type Confusion struct {
	Classes []string
	M       [][]int
}

// ConfusionMatrix builds a Confusion from the true and predicted labels.
//
// Time: O(n + classes^2).
// Space: O(classes^2).
func ConfusionMatrix(actual, predicted []string) (Confusion, error) {
	if len(actual) == 0 || len(actual) != len(predicted) {
		return Confusion{}, fmt.Errorf("%w: %d true labels and %d predictions", ErrInput, len(actual), len(predicted))
	}
	c := Confusion{Classes: Classes(append(append([]string(nil), actual...), predicted...))}
	at := map[string]int{}
	for i, l := range c.Classes {
		at[l] = i
	}
	c.M = make([][]int, len(c.Classes))
	for i := range c.M {
		c.M[i] = make([]int, len(c.Classes))
	}
	for i := range actual {
		c.M[at[actual[i]]][at[predicted[i]]]++
	}
	return c, nil
}

// Score is one class's precision (of those predicted as it, the share that
// were), recall (of those that were, the share found) and F1 (their harmonic
// mean, high only when both are).
type Score struct {
	Class                 string
	Precision, Recall, F1 float64
	Support               int // how many examples truly are this class
}

// Scores returns each class's Score and the accuracy (the diagonal's share).
// A class never predicted has precision 0.
//
// Time: O(classes^2).
// Space: O(classes).
func (c Confusion) Scores() ([]Score, float64) {
	var out []Score
	right, total := 0, 0
	for i, name := range c.Classes {
		tp, predicted, actual := c.M[i][i], 0, 0
		for j := range c.Classes {
			predicted += c.M[j][i]
			actual += c.M[i][j]
			total += c.M[i][j]
		}
		right += tp
		s := Score{Class: name, Support: actual}
		if predicted > 0 {
			s.Precision = float64(tp) / float64(predicted)
		}
		if actual > 0 {
			s.Recall = float64(tp) / float64(actual)
		}
		if s.Precision+s.Recall > 0 {
			s.F1 = 2 * s.Precision * s.Recall / (s.Precision + s.Recall)
		}
		out = append(out, s)
	}
	return out, float64(right) / float64(total)
}

// TrainTestSplit shuffles the indexes 0 to n-1 with seed and holds back
// testShare of them (at least one, at most n-1) for testing, so a model is
// judged on examples it never saw.
//
// Time: O(n).
// Space: O(n).
func TrainTestSplit(n int, testShare float64, seed int64) (train, test []int, err error) {
	if n < 2 || testShare <= 0 || testShare >= 1 {
		return nil, nil, fmt.Errorf("%w: want at least two examples and a test share between 0 and 1", ErrInput)
	}
	p := rand.New(rand.NewSource(seed)).Perm(n)
	k := min(max(int(float64(n)*testShare+0.5), 1), n-1)
	return p[k:], p[:k], nil
}

// KFold shuffles 0 to n-1 with seed and deals them into k folds of nearly
// equal size; cross-validation tests on each fold in turn while training on
// the rest, so every example is tested exactly once.
//
// Time: O(n).
// Space: O(n).
func KFold(n, k int, seed int64) ([][]int, error) {
	if k < 2 || k > n {
		return nil, fmt.Errorf("%w: k must be from 2 to the number of examples (%d)", ErrInput, n)
	}
	p := rand.New(rand.NewSource(seed)).Perm(n)
	folds := make([][]int, k)
	for i, v := range p {
		folds[i%k] = append(folds[i%k], v)
	}
	return folds, nil
}

// CrossValidateKNN runs k-fold cross-validation of KNN, voting among the given
// number of neighbors, and returns each fold's accuracy.
//
// Time: O(n^2 * d).
// Space: O(n).
func CrossValidateKNN(xs [][]float64, labels []string, folds, neighbors int, seed int64) ([]float64, error) {
	if err := check(xs, labels); err != nil {
		return nil, err
	}
	fs, err := KFold(len(xs), folds, seed)
	if err != nil {
		return nil, err
	}
	var acc []float64
	for f, test := range fs {
		var tx [][]float64
		var tl []string
		for g, other := range fs {
			if g == f {
				continue
			}
			for _, i := range other {
				tx, tl = append(tx, xs[i]), append(tl, labels[i])
			}
		}
		right := 0
		for _, i := range test {
			l, _, err := KNN(tx, tl, xs[i], min(neighbors, len(tx)))
			if err != nil {
				return nil, err
			}
			if l == labels[i] {
				right++
			}
		}
		acc = append(acc, float64(right)/float64(len(test)))
	}
	return acc, nil
}
