package classification

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// GaussianNB is a naive Bayes model for numeric features: for each class, the
// share of examples in it (the prior) and each feature's mean and variance.
type GaussianNB struct {
	Classes []string
	Prior   []float64
	Mean    [][]float64
	Var     [][]float64
}

// FitGaussianNB learns a GaussianNB. "Naive" is the assumption that features
// are independent given the class, so P(x | class) is a product of one bell
// curve per feature; it is rarely true and works surprisingly well anyway.
//
// Time: O(n * d).
// Space: O(classes * d).
func FitGaussianNB(xs [][]float64, labels []string) (*GaussianNB, error) {
	if err := check(xs, labels); err != nil {
		return nil, err
	}
	m := &GaussianNB{Classes: Classes(labels)}
	d := len(xs[0])
	for _, c := range m.Classes {
		var rows [][]float64
		for i, l := range labels {
			if l == c {
				rows = append(rows, xs[i])
			}
		}
		mean, v := make([]float64, d), make([]float64, d)
		for _, r := range rows {
			for j, x := range r {
				mean[j] += x / float64(len(rows))
			}
		}
		for _, r := range rows {
			for j, x := range r {
				v[j] += (x - mean[j]) * (x - mean[j]) / float64(len(rows))
			}
		}
		for j := range v {
			v[j] += 1e-9 // a feature constant within a class must not divide by zero
		}
		m.Prior = append(m.Prior, float64(len(rows))/float64(len(xs)))
		m.Mean, m.Var = append(m.Mean, mean), append(m.Var, v)
	}
	return m, nil
}

// Predict returns the most probable class for row and every class's log
// probability (up to a constant shared by all classes).
func (m *GaussianNB) Predict(row []float64) (string, []float64) {
	scores := make([]float64, len(m.Classes))
	best := 0
	for c := range m.Classes {
		s := math.Log(m.Prior[c])
		for j, x := range row {
			v := m.Var[c][j]
			s += -0.5*math.Log(2*math.Pi*v) - (x-m.Mean[c][j])*(x-m.Mean[c][j])/(2*v)
		}
		scores[c] = s
		if s > scores[best] {
			best = c
		}
	}
	return m.Classes[best], scores
}

// TextNB is a multinomial naive Bayes model over words, the classic spam
// filter: per class, how often each word appears.
type TextNB struct {
	Classes []string
	docs    map[string]int
	words   map[string]map[string]int
	total   map[string]int
	vocab   map[string]bool
	n       int
}

// Words splits text into lowercase words of letters and digits.
func Words(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// FitTextNB learns word counts from docs and their labels.
//
// Time: O(total words).
// Space: O(classes * vocabulary).
func FitTextNB(docs, labels []string) (*TextNB, error) {
	if len(docs) == 0 || len(docs) != len(labels) {
		return nil, fmt.Errorf("%w: %d documents and %d labels", ErrInput, len(docs), len(labels))
	}
	m := &TextNB{Classes: Classes(labels), docs: map[string]int{}, words: map[string]map[string]int{}, total: map[string]int{}, vocab: map[string]bool{}, n: len(docs)}
	for i, d := range docs {
		c := labels[i]
		m.docs[c]++
		if m.words[c] == nil {
			m.words[c] = map[string]int{}
		}
		for _, w := range Words(d) {
			m.words[c][w]++
			m.total[c]++
			m.vocab[w] = true
		}
	}
	return m, nil
}

// Predict returns the most probable class for text and each class's log
// probability. Every word count gets one added (Laplace smoothing), so a
// word never seen in a class lowers its score instead of zeroing it.
func (m *TextNB) Predict(text string) (string, []float64) {
	scores := make([]float64, len(m.Classes))
	best := 0
	v := float64(len(m.vocab))
	for c, name := range m.Classes {
		s := math.Log(float64(m.docs[name]) / float64(m.n))
		for _, w := range Words(text) {
			s += math.Log((float64(m.words[name][w]) + 1) / (float64(m.total[name]) + v))
		}
		scores[c] = s
		if s > scores[best] {
			best = c
		}
	}
	return m.Classes[best], scores
}

// ParseDocs reads one labeled document a line, "label: text".
func ParseDocs(text string) (docs, labels []string, err error) {
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		label, doc, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(label) == "" {
			return nil, nil, fmt.Errorf("%w: line %d: want \"label: text\"", ErrInput, i+1)
		}
		docs, labels = append(docs, strings.TrimSpace(doc)), append(labels, strings.TrimSpace(label))
	}
	if len(docs) == 0 {
		return nil, nil, fmt.Errorf("%w: no documents", ErrInput)
	}
	return docs, labels, nil
}
