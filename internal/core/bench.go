package core

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/dataset"
)

// Sample is one measured run: the input size, the events the entry reported
// (its operations), and how long an untraced run took.
type Sample struct {
	N     int   `json:"n"`
	Ops   int   `json:"ops"`
	Nanos int64 `json:"nanos"`
}

// Fit is how well one growth class explains the samples: ops is about
// Scale * f(n), off by Error (the root mean square of the relative errors, so
// 0.05 means about 5%).
type Fit struct {
	Class string  `json:"class"`
	Scale float64 `json:"scale"`
	Error float64 `json:"error"`
}

// Measurement is an entry measured at growing sizes, and the growth class
// that fits it best.
type Measurement struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Order   string   `json:"order"`  // how the values were generated: random, sorted, ...
	Stated  string   `json:"stated"` // the entry's stated time
	Samples []Sample `json:"samples"`
	// ByTime is set when the entry reports no events, so the fit is on wall
	// time, which varies from run to run.
	ByTime  bool    `json:"byTime,omitempty"`
	Fits    []Fit   `json:"fits"` // every class, best first
	Best    string  `json:"best"`
	Error   float64 `json:"error"`
	Quality string  `json:"quality"` // good (under 10%), rough (under 25%) or poor
	// Agrees says whether the best fit grows no faster than the worst cost
	// the entry states; it is absent when the statement uses a size the
	// measurement does not vary (O(n * m) with a fixed m).
	Agrees *bool  `json:"agrees,omitempty"`
	Note   string `json:"note,omitempty"`
}

// growth is one class a measurement is fitted to, in order of growth.
type growth struct {
	class string
	rank  float64
	f     func(n float64) float64
}

var growths = []growth{
	{"O(1)", 0, func(float64) float64 { return 1 }},
	{"O(log n)", 1, math.Log2},
	{"O(sqrt n)", 1.5, math.Sqrt},
	{"O(n)", 2, func(n float64) float64 { return n }},
	{"O(n log n)", 3, func(n float64) float64 { return n * math.Log2(n) }},
	{"O(n^2)", 4, func(n float64) float64 { return n * n }},
	{"O(n^3)", 5, func(n float64) float64 { return n * n * n }},
	{"O(2^n)", 6, math.Exp2},
}

// DefaultSizes are the sizes Measure uses when given none: each double the
// last, so the growth from one to the next shows the class.
var DefaultSizes = []int{100, 200, 400, 800}

// Orders are the value orders Measure can generate: every dataset kind
// (random, sorted, reversed, nearly-sorted, few-unique, negative).
func Orders() []string {
	out := make([]string, len(dataset.Kinds))
	for i, k := range dataset.Kinds {
		out[i] = string(k)
	}
	return out
}

// Measure runs the entry id on random values at each of sizes (DefaultSizes
// when empty), seeded so every run is the same, and fits the growth of its
// operation count to O(1), O(log n), O(sqrt n), O(n), O(n log n), O(n^2),
// O(n^3) and O(2^n). It is MeasureOrder with random values.
func Measure(id string, sizes []int, seed int64) (Measurement, error) {
	return MeasureOrder(id, string(dataset.Random), sizes, seed)
}

// MeasureOrder is Measure with values in the given order, one of Orders.
//
// The method: for each class f, the scale c that best explains every point
// is found by least squares on the relative error, minimizing the sum of
// (1 - c*f(n)/ops)^2, which gives c = sum(r) / sum(r^2) with r = f(n)/ops.
// Fitting relative rather than absolute error keeps the largest size from
// drowning out the others. The class with the smallest remaining error wins.
// Operations are the events the entry reports, so the measurement is exact and
// repeatable; time is kept as a second opinion. Sizes run smallest first, and
// the measurement stops early at a size the entry rejects or that takes more
// than a second.
//
// Time: the entry's own cost at each size. Space: the entry's, plus O(len(sizes)).
func MeasureOrder(id, order string, sizes []int, seed int64) (Measurement, error) {
	e, ok := Lookup(id)
	if !ok {
		return Measurement{}, fmt.Errorf("%w: no entry %q (see `algorithm list`)", ErrInput, id)
	}
	if len(sizes) == 0 {
		sizes = DefaultSizes
	}
	if err := checkSizes(sizes); err != nil {
		return Measurement{}, err
	}
	if !containsString(Orders(), order) {
		return Measurement{}, fmt.Errorf("%w: order %q is not one of %s", ErrInput, order, strings.Join(Orders(), ", "))
	}
	if !Measurable(e) {
		return Measurement{}, fmt.Errorf("%w: %s cannot be measured: its input (%s) does not grow with one size", ErrInput, e.ID, e.Input)
	}
	m := Measurement{ID: e.ID, Name: e.Name, Order: order, Stated: e.Time, ByTime: !e.Traced}
	for _, n := range sizes {
		res, err := measureAt(e, n, order, seed)
		if err != nil {
			if len(m.Samples) == 0 {
				return Measurement{}, fmt.Errorf("%s rejected the generated input at size %d, so it cannot be measured this way: %w", e.ID, n, err)
			}
			m.Note = fmt.Sprintf("stopped before size %d: %v", n, err)
			break
		}
		m.Samples = append(m.Samples, res)
		if time.Duration(res.Nanos) > time.Second {
			m.Note = fmt.Sprintf("stopped after size %d: it took over a second", n)
			break
		}
	}
	if len(m.Samples) < 3 {
		return Measurement{}, fmt.Errorf("%w: %s gave %d sizes; three are needed to see a growth (%s)", ErrInput, e.ID, len(m.Samples), m.Note)
	}
	ys := make([]float64, len(m.Samples))
	for i, p := range m.Samples {
		ys[i] = float64(p.Ops)
		if m.ByTime {
			ys[i] = float64(p.Nanos)
		}
	}
	m.Fits = fitGrowth(m.Samples, ys)
	m.Best, m.Error = m.Fits[0].Class, m.Fits[0].Error
	switch {
	case m.Error < 0.10:
		m.Quality = "good"
	case m.Error < 0.25:
		m.Quality = "rough"
	default:
		m.Quality = "poor"
	}
	if stated, ok := StatedRank(e.Time); ok {
		agrees := classRank(m.Best) <= stated
		m.Agrees = &agrees
	}
	return m, nil
}

// Measurable reports whether Measure can grow the entry's input with one
// size: values, texts, grids, graphs, points and numbers. Simulations and
// fixed-shape inputs are not, and neither are data structures: they report
// one event per operation whatever it costs inside, so counting cannot see
// the difference between O(1) and O(log n).
func Measurable(e Entry) bool {
	switch e.Input {
	case Ints, SortedInts, AnyInts, Text, TwoTexts, Pattern, Number, GridInput, GraphInput, Points, Amount, Pages, Tracks:
		return true
	default:
		return false
	}
}

// searchTargets is how many targets a search is measured on at each size.
// One target could sit exactly where a search looks first (two thirds of the
// way in is ternary search's second probe), so the samples sum several,
// spread through the list by the golden ratio, the same at every size.
const searchTargets = 7

// measureAt runs e once at size n, or once per target for a search, and
// sums the operations and the time.
func measureAt(e Entry, n int, order string, seed int64) (Sample, error) {
	in, err := sizedInput(e, n, order, seed)
	if err != nil {
		return Sample{}, err
	}
	runs := 1
	var sorted []int
	if e.Input == SortedInts || e.Input == AnyInts {
		runs = searchTargets
		sorted = in.Values
		if e.Input == SortedInts {
			sorted = append([]int(nil), in.Values...)
			sort.Ints(sorted)
		}
	}
	sample := Sample{N: n}
	for i := 1; i <= runs; i++ {
		if sorted != nil {
			_, frac := math.Modf(float64(i) * 0.6180339887)
			in.Target = sorted[int(frac*float64(n))]
		}
		res, err := execute(e, in, -1)
		if err != nil {
			return Sample{}, err
		}
		for _, c := range res.Counts {
			sample.Ops += c
		}
		sample.Nanos += res.Nanos
	}
	return sample, nil
}

func checkSizes(sizes []int) error {
	if len(sizes) < 3 || len(sizes) > 12 {
		return fmt.Errorf("%w: give 3 to 12 sizes, not %d", ErrInput, len(sizes))
	}
	for i, n := range sizes {
		if n < 1 || n > dataset.MaxValues {
			return fmt.Errorf("%w: size %d is outside 1..%d", ErrInput, n, dataset.MaxValues)
		}
		if i > 0 && n <= sizes[i-1] {
			return fmt.Errorf("%w: sizes must grow: %d comes after %d", ErrInput, n, sizes[i-1])
		}
	}
	return nil
}

// sizedInput is the entry's example input with its growing part replaced by
// one of size n, so options such as frames or a quantum keep their example
// values.
func sizedInput(e Entry, n int, order string, seed int64) (Input, error) {
	in := Input{}
	if e.Example != nil {
		in = *e.Example
		in.Options = map[string]int{}
		for k, v := range e.Example.Options {
			in.Options[k] = v
		}
	}
	r := rand.New(rand.NewSource(seed))
	values := func() ([]int, error) { return dataset.Generate(dataset.Kind(order), n, seed) }
	var err error
	switch e.Input {
	case Ints:
		in.Values, err = values()
	case SortedInts, AnyInts:
		in.Values, err = values()
	case Text:
		in.Text = randomText(r, n, "abcdefghijklmnopqrstuvwxyz")
	case TwoTexts:
		in.Text, in.Other = randomText(r, n, "abcd"), randomText(r, n, "abcd")
	case Pattern:
		in.Text = randomText(r, n, "ab")
	case Number:
		in.Values = []int{n}
	case Amount:
		in.Target = n
	case Pages:
		in.Values = make([]int, n)
		for i := range in.Values {
			in.Values[i] = r.Intn(10)
		}
	case Tracks:
		size := in.opt("size", 200)
		in.Values = make([]int, n)
		for i := range in.Values {
			in.Values[i] = r.Intn(size)
		}
	case GridInput:
		in.Text = openGrid(n)
	case GraphInput:
		in.Text = randomGraph(r, n)
		if e.Example != nil && strings.HasPrefix(e.Example.Text, "directed") {
			in.Text = "directed\n" + in.Text
		}
	case Points:
		var b strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "%d %d\n", r.Intn(10*n), r.Intn(10*n))
		}
		in.Text = b.String()
	default:
		return in, fmt.Errorf("%w: %s cannot be measured", ErrInput, e.ID)
	}
	return in, err
}

func randomText(r *rand.Rand, n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

// openGrid is a wall-free square grid of about n cells, start top left and
// end bottom right.
func openGrid(n int) string {
	side := int(math.Round(math.Sqrt(float64(n))))
	if side < 2 {
		side = 2
	}
	var b strings.Builder
	for row := 0; row < side; row++ {
		for col := 0; col < side; col++ {
			switch {
			case row == 0 && col == 0:
				b.WriteByte('S')
			case row == side-1 && col == side-1:
				b.WriteByte('E')
			default:
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// randomGraph is a connected undirected graph of n nodes: a random spanning
// tree plus about n more edges, weights 1 to 9.
func randomGraph(r *rand.Rand, n int) string {
	var b strings.Builder
	name := func(i int) string { return "N" + strconv.Itoa(i) }
	for i := 1; i < n; i++ {
		fmt.Fprintf(&b, "%s %s %d\n", name(r.Intn(i)), name(i), 1+r.Intn(9))
	}
	for i := 0; i < n; i++ {
		a, c := r.Intn(n), r.Intn(n)
		if a != c {
			fmt.Fprintf(&b, "%s %s %d\n", name(a), name(c), 1+r.Intn(9))
		}
	}
	return b.String()
}

// fitGrowth fits every class to ys measured at the samples' sizes, best first.
func fitGrowth(samples []Sample, ys []float64) []Fit {
	fits := make([]Fit, 0, len(growths))
	for _, g := range growths {
		var sr, srr float64
		finite := true
		for i, p := range samples {
			y := math.Max(ys[i], 1)
			r := g.f(float64(p.N)) / y
			if math.IsInf(r, 0) || math.IsNaN(r) {
				finite = false
				break
			}
			sr += r
			srr += r * r
		}
		if !finite || srr == 0 {
			fits = append(fits, Fit{Class: g.class, Error: math.Inf(1)})
			continue
		}
		c := sr / srr
		var sum float64
		for i, p := range samples {
			d := 1 - c*g.f(float64(p.N))/math.Max(ys[i], 1)
			sum += d * d
		}
		fits = append(fits, Fit{Class: g.class, Scale: c, Error: math.Sqrt(sum / float64(len(samples)))})
	}
	sort.SliceStable(fits, func(i, j int) bool { return fits[i].Error < fits[j].Error })
	for i := range fits {
		if math.IsInf(fits[i].Error, 0) {
			fits[i].Error = -1 // JSON has no infinity: -1 means the class overflowed
		}
	}
	return fits
}

func classRank(class string) float64 {
	for _, g := range growths {
		if g.class == class {
			return g.rank
		}
	}
	return -1
}

// statedRanks maps a stated cost, with spaces removed and V and E read as n,
// to its place among the growth classes; costs between two classes sit
// between their ranks.
var statedRanks = map[string]float64{
	"1": 0, "logn": 1, "log^2n": 1, "log(n)": 1, "sqrtn": 1.5,
	"n": 2, "n+n": 2, "nloglogn": 2.5,
	"nlogn": 3, "(n+n)logn": 3, "n+nlogn": 3,
	"nlog^2n": 3.5, "n^(4/3)": 3.5, "n^1.585": 3.5, "n^1.5": 3.5,
	"n^2": 4, "n*n": 4, "n^2logn": 4.5,
	"n^2.71": 5, "n^2.81": 5, "n^3": 5,
	"2^n": 6, "n*2^n": 6, "n!": 6, "n*n!": 6, "phi^n": 6, "1.618^n": 6,
}

// StatedRank reads the growth classes in a stated cost ("O(n log n) average,
// O(n^2) worst") and returns the rank of the fastest-growing one, or false
// when one of them uses a size other than n (O(n * m)).
func StatedRank(stated string) (float64, bool) {
	best, found := -1.0, false
	rest := stated
	for {
		i := strings.Index(rest, "O(")
		if i < 0 {
			break
		}
		rest = rest[i+2:]
		depth, j := 1, 0
		for ; j < len(rest) && depth > 0; j++ {
			switch rest[j] {
			case '(':
				depth++
			case ')':
				depth--
			}
		}
		expr := strings.NewReplacer(" ", "", "V", "n", "E", "n").Replace(rest[:j-1])
		rank, ok := statedRanks[expr]
		if !ok {
			return 0, false
		}
		best, found = math.Max(best, rank), true
		rest = rest[j:]
	}
	return best, found
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Predict is what a fit says the count is at size n: Scale * f(n), or 0 for
// a class it does not know.
func (f Fit) Predict(n int) float64 {
	for _, g := range growths {
		if g.class == f.Class {
			return f.Scale * g.f(float64(n))
		}
	}
	return 0
}
