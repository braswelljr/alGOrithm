package core

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/math/bitwise"
	"github.com/braswelljr/alGOrithm/pkg/math/geometry"
	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
	"github.com/braswelljr/alGOrithm/pkg/math/numerical"
	"github.com/braswelljr/alGOrithm/pkg/math/randomized"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// numbersEntries is every entry of the numbers category.
func numbersEntries() []Entry {
	return []Entry{
		binaryGcdEntry(),
		chineseRemainderEntry(),
		collatzEntry(),
		divisorsEntry(),
		eulerTotientEntry(),
		evenEntry(),
		extendedGcdEntry(),
		fastPowerEntry(),
		fibonacciFastDoublingEntry(),
		fibonacciMatrixEntry(),
		karatsubaEntry(),
		millerRabinEntry(),
		modInverseEntry(),
		newtonSqrtEntry(),
		pascalTriangleEntry(),
		perfectNumbersEntry(),
		primeEntry(),
		primeFactorsEntry(),
		segmentedSieveEntry(),
		sieveOfEratosthenesEntry(),
	}
}

// geometryEntries is every entry of the geometry category.
func geometryEntries() []Entry {
	return []Entry{
		closestPairEntry(),
		convexHullGrahamEntry(),
		convexHullJarvisEntry(),
		convexHullMonotoneEntry(),
		pointInPolygonEntry(),
		polygonAreaEntry(),
		segmentIntersectionEntry(),
	}
}

// bitwiseEntries is every entry of the bitwise category.
func bitwiseEntries() []Entry {
	return []Entry{
		bitSubsetsEntry(),
		grayCodeEntry(),
		popcountEntry(),
		powerOfTwoEntry(),
		reverseBitsEntry(),
		singleNumberEntry(),
	}
}

// numericalEntries is every entry of the numerical category.
func numericalEntries() []Entry {
	return []Entry{
		bisectionEntry(),
		fftEntry(),
		gaussianEliminationEntry(),
		hornerEntry(),
		matrixPowerEntry(),
		newtonRaphsonEntry(),
		secantMethodEntry(),
		simpsonIntegrationEntry(),
		strassenEntry(),
		trapezoidRuleEntry(),
	}
}

// randomizedEntries is every entry of the randomized category.
func randomizedEntries() []Entry {
	return []Entry{
		kargerMinCutEntry(),
		monteCarloPiEntry(),
		randomizedQuickselectEntry(),
		reservoirSamplingEntry(),
	}
}

func numbersBoolean(id, name, fn, summary, time string, f func(int) bool) Entry {
	e := entry(id, name, "math/numbers/numbers.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{97}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		return Output{Bool: ptr(f(n))}, nil
	}
	return e
}

func sieveOfEratosthenesEntry() Entry {
	sieve := entry("sieve-of-eratosthenes", "Sieve of Eratosthenes", "math/numbers/sieve.go", "Sieve")
	sieve.Summary, sieve.Time, sieve.Space, sieve.Input, sieve.Traced = "Cross out the multiples of each prime from its square; what is never crossed out is prime.", "O(n log log n)", "O(n)", Number, true
	sieve.Example, sieve.Limits = &Input{Values: []int{120}}, fmt.Sprintf("2 <= n <= %d", numbers.MaxSieve)
	sieve.Run = func(in Input, r trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 2 || n > numbers.MaxSieve {
			return Output{}, inputError(fmt.Sprintf("the sieve takes 2 <= n <= %d", numbers.MaxSieve))
		}
		p := numbers.TraceSieve(n, r)
		initial := make([]int, n+1)
		for i := range initial {
			initial[i] = i
		}
		return Output{Initial: initial, Values: p, Number: ptr(len(p))}, nil
	}
	return sieve
}

func primeFactorsEntry() Entry {
	fac := entry("prime-factors", "Prime Factorization", "math/numbers/numbers.go", "Factorize")
	fac.Summary, fac.Time, fac.Space, fac.Input, fac.Example = "The prime factors of n, by trial division.", "O(sqrt n)", "O(log n)", Number, &Input{Values: []int{360}}
	fac.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		return Output{Values: numbers.Factorize(n)}, nil
	}
	return fac
}

func fastPowerEntry() Entry {
	pow := entry("fast-power", "Fast Exponentiation", "math/numbers/numbers.go", "Power")
	pow.Summary, pow.Time, pow.Space, pow.Input, pow.Options = "base^exp mod m by squaring: O(log exp) multiplications instead of exp.", "O(log exp)", "O(1)", TwoNumbers, []Option{{"mod", "modulus", 1_000_000_007, 1, 2_000_000_000}}
	pow.Example = &Input{Values: []int{3, 200}}
	pow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		b, e, err := two(in)
		if err != nil {
			return Output{}, err
		}
		m := in.opt("mod", 1_000_000_007)
		if e < 0 || m < 1 || m > 2_000_000_000 {
			return Output{}, inputError("give exp >= 0 and 1 <= mod <= 2000000000")
		}
		return Output{Number: ptr(numbers.Power(b, e, m))}, nil
	}
	return pow
}

func primeEntry() Entry {
	return numbersBoolean("prime", "Primality", "Prime", "Whether n is prime, by trial division up to sqrt(n).", "O(sqrt n)", numbers.Prime)
}

func evenEntry() Entry {
	return numbersBoolean("even", "Even", "Even", "Whether n is even.", "O(1)", numbers.Even)
}

// toPoints converts geometry points for the visualizer.
func toPoints(pts []geometry.Point) []Point {
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.X, Y: p.Y}
	}
	return out
}

func parsePoints(in Input) ([]geometry.Point, error) {
	pts, err := geometry.ParsePoints(in.Text)
	return pts, wrap(err)
}

// pointsCall is the snippet setup that parses the input's points.
func pointsCall(in Input, call string) (string, string) {
	return fmt.Sprintf("pts, err := geometry.ParsePoints(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), call
}

// segmentRecorder turns a segment trace's segment indexes into the indexes of
// the segments' two points, which is what the visualizer draws.
type segmentRecorder struct{ r trace.Recorder }

func (s segmentRecorder) Record(e trace.Event) {
	idx := make([]int, 0, 2*len(e.Indices))
	for _, i := range e.Indices {
		idx = append(idx, 2*i, 2*i+1)
	}
	e.Indices = idx
	s.r.Record(e)
}

const pointsExample = "0 0\n7 1\n3 3\n9 4\n5 6\n1 8\n8 9\n4 9\n2 5\n6 3\n10 7\n5 0"

var geometryEx = &Input{Text: pointsExample}

func geometryHull(id, name, file, fn, summary, time string, traced func([]geometry.Point, trace.Recorder) []int) Entry {
	e := entry(id, name, "math/geometry/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, "O(n)", Points, true, geometryEx
	e.Limits = fmt.Sprintf("at most %d points", geometry.MaxPoints)
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		h := traced(pts, r)
		hp := make([]geometry.Point, len(h))
		for i, k := range h {
			hp[i] = pts[k]
		}
		area := float64(geometry.PolygonArea(hp)) / 2
		return Output{Points: toPoints(pts), Values: h, Number: ptr(len(h)), Summary: map[string]float64{"hull points": float64(len(h)), "hull area": area}}, nil
	}
	e.Call = func(in Input) (string, string) { return pointsCall(in, "geometry."+fn+"(pts)") }
	return e
}

func closestPairEntry() Entry {
	closest := entry("closest-pair", "Closest Pair of Points", "math/geometry/closest.go", "ClosestPair")
	closest.Summary, closest.Time, closest.Space, closest.Input, closest.Traced, closest.Example = "The two points nearest each other, by divide and conquer: solve each half, then check a narrow strip.", "O(n log^2 n)", "O(n)", Points, true, geometryEx
	closest.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		if len(pts) < 2 {
			return Output{}, inputError("give at least two points")
		}
		i, j, d := geometry.TraceClosestPair(pts, r)
		return Output{Points: toPoints(pts), Values: []int{i, j},
			Lines:   []string{fmt.Sprintf("points %d (%d, %d) and %d (%d, %d)", i, pts[i].X, pts[i].Y, j, pts[j].X, pts[j].Y)},
			Summary: map[string]float64{"distance": math.Sqrt(float64(d)), "squared distance": float64(d)}}, nil
	}
	closest.Call = func(in Input) (string, string) { return pointsCall(in, "geometry.ClosestPair(pts)") }
	return closest
}

func segmentIntersectionEntry() Entry {
	segs := entry("segment-intersection", "Segment Intersection", "math/geometry/segments.go", "SegmentsIntersect")
	segs.Summary, segs.Time, segs.Space, segs.Input, segs.Traced = "Which segments cross or touch, tested pair by pair with orientation (cross product) tests.", "O(n^2)", "O(k)", Points, true
	segs.Example = &Input{Text: "0 0\n8 8\n0 8\n8 0\n2 6\n6 7\n9 1\n9 6\n1 1\n3 0"}
	segs.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		ss, err := geometry.Segments(pts)
		if err != nil {
			return Output{}, wrap(err)
		}
		pairs := geometry.TraceSegmentsIntersect(ss, segmentRecorder{r})
		out := Output{Points: toPoints(pts), Number: ptr(len(pairs))}
		for i := 0; i+1 < len(pts); i += 2 {
			out.Segments = append(out.Segments, [2]int{i, i + 1})
		}
		for _, p := range pairs {
			out.Lines = append(out.Lines, fmt.Sprintf("segment %d and segment %d intersect", p[0], p[1]))
		}
		if len(pairs) == 0 {
			out.Lines = []string{"no two segments intersect"}
		}
		return out, nil
	}
	segs.Call = func(in Input) (string, string) {
		s, _ := pointsCall(in, "")
		return s + "segs, err := geometry.Segments(pts)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "geometry.SegmentsIntersect(segs)"
	}
	return segs
}

func pointInPolygonEntry() Entry {
	pip := entry("point-in-polygon", "Point in Polygon", "math/geometry/polygon.go", "PointInPolygon")
	pip.Summary, pip.Time, pip.Space, pip.Input, pip.Traced = "Whether points lie inside a polygon, by casting a ray to the right and counting the edges it crosses.", "O(n) per point", "O(1)", Points, true
	pip.Example = &Input{Text: "# the polygon, in order\n0 0\n8 0\n8 8\n4 3\n0 8\ntest\n2 2\n4 6\n6 5\n9 4\n4 1"}
	pip.Run = func(in Input, r trace.Recorder) (Output, error) {
		polyText, testText, ok := strings.Cut(in.Text, "test")
		if !ok {
			return Output{}, inputError(`give the polygon's points, a line "test", then the points to test`)
		}
		poly, err := geometry.ParsePoints(polyText)
		if err != nil {
			return Output{}, wrap(err)
		}
		test, err := geometry.ParsePoints(testText)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(poly) < 3 {
			return Output{}, inputError("a polygon needs at least three points")
		}
		outline := make([]int, len(poly))
		for i := range outline {
			outline[i] = i
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: outline, Note: "hull"}) // the polygon's outline
		in2 := geometry.TracePointsInPolygon(poly, test, len(poly), r)
		out := Output{Points: toPoints(append(append([]geometry.Point(nil), poly...), test...))}
		inside := 0
		for i, b := range in2 {
			where := "outside"
			if b {
				where, inside = "inside", inside+1
			}
			out.Lines = append(out.Lines, fmt.Sprintf("(%d, %d) is %s", test[i].X, test[i].Y, where))
		}
		out.Number = &inside
		return out, nil
	}
	return pip
}

func polygonAreaEntry() Entry {
	area := entry("polygon-area", "Polygon Area (shoelace)", "math/geometry/polygon.go", "PolygonArea")
	area.Summary, area.Time, area.Space, area.Input, area.Traced = "The area of a polygon from its vertices in order, exactly, with the shoelace formula.", "O(n)", "O(1)", Points, true
	area.Example = &Input{Text: "0 0\n8 0\n8 8\n4 3\n0 8"}
	area.Run = func(in Input, r trace.Recorder) (Output, error) {
		pts, err := parsePoints(in)
		if err != nil {
			return Output{}, err
		}
		if len(pts) < 3 {
			return Output{}, inputError("a polygon needs at least three points")
		}
		outline := make([]int, len(pts))
		for i := range outline {
			outline[i] = i
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: outline, Note: "hull"})
		twice := geometry.PolygonArea(pts)
		dir := "counterclockwise"
		if twice < 0 {
			dir = "clockwise"
		}
		return Output{Points: toPoints(pts), Lines: []string{"the vertices run " + dir},
			Summary: map[string]float64{"area": math.Abs(float64(twice)) / 2, "twice the signed area": float64(twice)}}, nil
	}
	area.Call = func(in Input) (string, string) { return pointsCall(in, "geometry.PolygonArea(pts)") }
	return area
}

func convexHullGrahamEntry() Entry {
	return geometryHull("convex-hull-graham", "Convex Hull (Graham Scan)", "hull.go", "ConvexHullGraham", "Sort the points by angle around the lowest one, then keep a stack that only turns left (Graham, 1972).", "O(n log n)", geometry.TraceConvexHullGraham)
}

func convexHullJarvisEntry() Entry {
	return geometryHull("convex-hull-jarvis", "Convex Hull (Jarvis March)", "hull.go", "ConvexHullJarvis", "Gift wrapping: from the lowest point, repeatedly take the point with every other to its left (Jarvis, 1973).", "O(n * h) for h hull points", geometry.TraceConvexHullJarvis)
}

func convexHullMonotoneEntry() Entry {
	return geometryHull("convex-hull-monotone", "Convex Hull (Monotone Chain)", "hull.go", "ConvexHullMonotone", "Sort by x, then build the lower and upper hulls with a left-turn stack (Andrew, 1979).", "O(n log n)", geometry.TraceConvexHullMonotone)
}

func arithmeticPair(id, name, file, fn, summary, time string, ex []int, run func(a, b int) (Output, error)) Entry {
	e := entry(id, name, "math/numbers/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", TwoNumbers, &Input{Values: ex}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b, err := two(in)
		if err != nil {
			return Output{}, err
		}
		return run(a, b)
	}
	e.Call = func(in Input) (string, string) {
		a, b := firstOr(in.Values, 0), 0
		if len(in.Values) > 1 {
			b = in.Values[1]
		}
		return "", fmt.Sprintf("numbers.%s(%d, %d)", fn, a, b)
	}
	return e
}

func arithmeticSingle(id, name, file, fn, summary, time string, ex int, run func(n int) (Output, error)) Entry {
	e := entry(id, name, "math/numbers/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{ex}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		return run(n)
	}
	e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("numbers.%s(%d)", fn, firstOr(in.Values, 0)) }
	return e
}

func extendedGcdEntry() Entry {
	egcd := arithmeticPair("extended-gcd", "Extended Euclidean Algorithm", "modular.go", "ExtendedGCD", "gcd(a, b) and the x, y with a*x + b*y = gcd: Euclid, carrying the coefficients along.", "O(log min(a, b))", []int{240, 46},
		func(a, b int) (Output, error) {
			g, x, y := numbers.ExtendedGCD(a, b)
			return Output{Number: &g, Values: []int{g, x, y}, Lines: []string{fmt.Sprintf("%d * %d + %d * %d = %d", a, x, b, y, g)}}, nil
		})
	return egcd
}

func modInverseEntry() Entry {
	inv := arithmeticPair("mod-inverse", "Modular Inverse", "modular.go", "ModInverse", "The x with a*x = 1 (mod m), from the extended Euclidean algorithm; none when a and m share a factor.", "O(log m)", []int{3, 11},
		func(a, m int) (Output, error) {
			x, ok := numbers.ModInverse(a, m)
			out := Output{Bool: &ok}
			if ok {
				out.Number = &x
				out.Lines = []string{fmt.Sprintf("%d * %d = 1 (mod %d)", a, x, m)}
			} else {
				out.Lines = []string{fmt.Sprintf("%d has no inverse modulo %d", a, m)}
			}
			return out, nil
		})
	return inv
}

func binaryGcdEntry() Entry {
	bgcd := arithmeticPair("binary-gcd", "Binary GCD (Stein)", "modular.go", "BinaryGCD", "The greatest common divisor with only shifts and subtraction, taking out shared factors of two first.", "O(log^2 n) bit operations", []int{1071, 462},
		func(a, b int) (Output, error) { return Output{Number: ptr(numbers.BinaryGCD(a, b))}, nil })
	return bgcd
}

func millerRabinEntry() Entry {
	mr := arithmeticSingle("miller-rabin", "Miller-Rabin Primality", "modular.go", "MillerRabin", "Write n-1 = d * 2^s and test bases: a composite fails for one; with twelve bases it is exact for every 64-bit n.", "O(k log^3 n)", 1000000007,
		func(n int) (Output, error) {
			p, w := numbers.MillerRabin(n)
			out := Output{Bool: &p}
			if w > 0 {
				out.Lines = []string{fmt.Sprintf("base %d proves %d composite", w, n)}
			}
			return out, nil
		})
	return mr
}

func karatsubaEntry() Entry {
	kara := entry("karatsuba", "Karatsuba Multiplication", "math/numbers/karatsuba.go", "Karatsuba")
	kara.Summary, kara.Time, kara.Space, kara.Input = "Multiply big numbers with three half-size products instead of four (1960).", "O(n^1.585)", "O(n)", TwoTexts
	kara.Example = &Input{Text: "3141592653589793238462643383279502884197169399375105820974944592", Other: "2718281828459045235360287471352662497757247093699959574966967627"}
	kara.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p, err := numbers.Karatsuba(in.Text, in.Other)
		if err != nil {
			return Output{}, inputError("give two non-negative whole numbers of up to 100000 digits")
		}
		return Output{Text: p, Lines: []string{fmt.Sprintf("%d digits", len(p))}}, nil
	}
	kara.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numbers.Karatsuba(%q, %q)", in.Text, in.Other)
	}
	return kara
}

func collatzEntry() Entry {
	collatz := arithmeticSingle("collatz", "Collatz Sequence", "sequences.go", "Collatz", "Halve even numbers, take 3n+1 of odd ones, until 1: every start tried so far gets there.", "O(steps)", 27,
		func(n int) (Output, error) {
			seq := numbers.Collatz(n)
			if seq == nil {
				return Output{}, inputError("give n >= 1")
			}
			return Output{Values: seq, Initial: seq, Number: ptr(len(seq) - 1)}, nil
		})
	return collatz
}

func pascalTriangleEntry() Entry {
	pascal := arithmeticSingle("pascal-triangle", "Pascal's Triangle", "sequences.go", "PascalTriangle", "Each value is the sum of the two above it, so row n holds the binomial coefficients C(n, k).", "O(rows^2)", 10,
		func(n int) (Output, error) {
			rows := numbers.PascalTriangle(n)
			if rows == nil {
				return Output{}, inputError(fmt.Sprintf("give 1 to %d rows", numbers.MaxPascalRows))
			}
			t := &TableOutput{Rows: n, Cols: n, RowLabels: numberLabels(n, 0), ColLabels: numberLabels(n, 0), Blank: ptr(-1)}
			for _, row := range rows {
				cells := make([]int, n)
				for k := range cells {
					cells[k] = -1
				}
				copy(cells, row)
				t.Cells = append(t.Cells, cells...)
			}
			return Output{Table: t}, nil
		})
	return pascal
}

func newtonSqrtEntry() Entry {
	newton := arithmeticSingle("newton-sqrt", "Integer Square Root (Newton)", "sequences.go", "NewtonSqrt", "x = (x + n/x) / 2 until it stops shrinking: the digits right double each step.", "O(log n)", 1000000,
		func(n int) (Output, error) {
			r, steps := numbers.NewtonSqrt(n)
			if r < 0 {
				return Output{}, inputError("give n >= 0")
			}
			return Output{Number: &r, Values: steps}, nil
		})
	return newton
}

func fibonacciMatrixEntry() Entry {
	fib := arithmeticSingle("fibonacci-matrix", "Fibonacci (matrix power)", "sequences.go", "FibonacciMatrix", "F(n) from the nth power of [[1 1] [1 0]] by repeated squaring: O(log n) instead of n steps.", "O(log n)", 90,
		func(n int) (Output, error) {
			v, err := numbers.FibonacciMatrix(n)
			if err != nil {
				return Output{}, inputError("give 0 <= n <= 92")
			}
			return Output{Number: &v}, nil
		})
	return fib
}

func numberTheorySingle(id, name, file, fn, summary, time, space string, ex int, run func(n int) (Output, error)) Entry {
	e := entry(id, name, "math/numbers/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, space, Number, &Input{Values: []int{ex}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		return run(n)
	}
	e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("numbers.%s(%d)", fn, firstOr(in.Values, 0)) }
	return e
}

func chineseRemainderEntry() Entry {
	crt := entry("chinese-remainder", "Chinese Remainder Theorem", "math/numbers/crt.go", "ChineseRemainder")
	crt.Summary, crt.Time, crt.Space, crt.Input = "The x that leaves given remainders for given moduli, combining one congruence at a time; the moduli need not be coprime.", "O(k log M)", "O(1)", Text
	crt.Example = &Input{Text: "2 3\n3 5\n2 7"}
	crt.Run = func(in Input, _ trace.Recorder) (Output, error) {
		r, m, err := parsePairs(in.Text)
		if err != nil {
			return Output{}, inputError(err.Error() + " (one \"remainder modulus\" pair a line)")
		}
		x, mod, err := numbers.ChineseRemainder(r, m)
		if err != nil {
			return Output{}, inputError("no solution: the congruences disagree, a modulus is below 1, or the product is too large")
		}
		out := Output{Number: ptr(x), Values: []int{x, mod}}
		for i := range r {
			out.Lines = append(out.Lines, fmt.Sprintf("%d mod %d = %d", x, m[i], x%m[i]))
		}
		out.Lines = append(out.Lines, fmt.Sprintf("every solution is %d + %d k", x, mod))
		return out, nil
	}
	crt.Call = func(in Input) (string, string) {
		r, m, _ := parsePairs(in.Text)
		return "", fmt.Sprintf("numbers.ChineseRemainder(%s, %s)", ints(r), ints(m))
	}
	return crt
}

func eulerTotientEntry() Entry {
	phi := numberTheorySingle("euler-totient", "Euler's Totient", "totient.go", "EulerTotient", "How many of 1..n share no factor with n: n times (1 - 1/p) for each prime p dividing it.", "O(sqrt n)", "O(1)", 36,
		func(n int) (Output, error) {
			if n < 1 {
				return Output{}, inputError("give n >= 1")
			}
			p := numbers.EulerTotient(n)
			out := Output{Number: ptr(p), Lines: []string{fmt.Sprintf("phi(%d) = %d", n, p)}}
			if n <= 1000 {
				for k := 1; k <= n; k++ {
					if numbers.BinaryGCD(n, k) == 1 {
						out.Values = append(out.Values, k)
					}
				}
			}
			if n <= 100_000 {
				t := numbers.TotientSieve(min(n, 20))
				out.Lines = append(out.Lines, "phi(1..20) by the sieve: "+joinInts(t[1:]))
			}
			return out, nil
		})
	return phi
}

func segmentedSieveEntry() Entry {
	seg := entry("segmented-sieve", "Segmented Sieve", "math/numbers/segmented.go", "SegmentedSieve")
	seg.Summary, seg.Time, seg.Space, seg.Input, seg.Traced = "The primes in a range [lo, hi], sieving one window of about sqrt(hi) numbers at a time with the primes up to sqrt(hi).", "O((hi - lo) log log hi + sqrt hi)", "O(sqrt hi)", TwoNumbers, true
	seg.Example, seg.Limits = &Input{Values: []int{100, 200}}, "hi - lo below 100000"
	seg.Run = func(in Input, r trace.Recorder) (Output, error) {
		lo, hi, err := two(in)
		if err != nil {
			return Output{}, err
		}
		if hi < lo || hi < 2 || hi-lo >= 100_000 {
			return Output{}, inputError("give lo <= hi, hi >= 2 and hi - lo below 100000")
		}
		p := numbers.TraceSegmentedSieve(lo, hi, r)
		return Output{Values: p, Number: ptr(len(p))}, nil
	}
	seg.Call = func(in Input) (string, string) {
		a, b := firstOr(in.Values, 0), 0
		if len(in.Values) > 1 {
			b = in.Values[1]
		}
		return "", fmt.Sprintf("numbers.SegmentedSieve(%d, %d)", a, b)
	}
	return seg
}

func divisorsEntry() Entry {
	div := numberTheorySingle("divisors", "Divisors", "divisors.go", "Divisors", "Every divisor of n, found in pairs d and n/d while d*d <= n.", "O(sqrt n)", "O(d(n))", 360,
		func(n int) (Output, error) {
			if n < 1 {
				return Output{}, inputError("give n >= 1")
			}
			d := numbers.Divisors(n)
			sum := 0
			for _, v := range d {
				sum += v
			}
			return Output{Values: d, Number: ptr(len(d)), Summary: map[string]float64{"divisors": float64(len(d)), "sum": float64(sum)}}, nil
		})
	return div
}

func perfectNumbersEntry() Entry {
	perfect := numberTheorySingle("perfect-numbers", "Perfect Numbers", "divisors.go", "PerfectNumbers", "The numbers up to a limit equal to the sum of their smaller divisors (6 = 1 + 2 + 3), with a sieve over divisor sums.", "O(n log n)", "O(n)", 10_000,
		func(n int) (Output, error) {
			if n < 1 || n > 1_000_000 {
				return Output{}, inputError("give a limit from 1 to 1000000")
			}
			p := numbers.PerfectNumbers(n)
			out := Output{Values: p, Number: ptr(len(p))}
			for _, v := range p {
				d := numbers.Divisors(v)
				parts := make([]string, len(d)-1)
				for i, x := range d[:len(d)-1] {
					parts[i] = strconv.Itoa(x)
				}
				out.Lines = append(out.Lines, fmt.Sprintf("%d = %s", v, strings.Join(parts, " + ")))
			}
			return out, nil
		})
	perfect.Limits = "a limit of at most 1000000"
	return perfect
}

func fibonacciFastDoublingEntry() Entry {
	fib := numberTheorySingle("fibonacci-fast-doubling", "Fibonacci (Fast Doubling)", "doubling.go", "FibonacciDoubling", "F(n) from F(2k) = F(k)(2F(k+1) - F(k)) and F(2k+1) = F(k)^2 + F(k+1)^2, reading the bits of n from the top.", "O(log n)", "O(1)", 90,
		func(n int) (Output, error) {
			f, err := numbers.FibonacciDoubling(n)
			if err != nil {
				return Output{}, inputError("give n from 0 to 92")
			}
			out := Output{Number: ptr(f)}
			for s := bits.Len(uint(n)) - 1; s >= 0; s-- {
				k := n >> s
				fk, _ := numbers.FibonacciDoubling(k)
				out.Lines = append(out.Lines, fmt.Sprintf("bits %b: k = %d, F(k) = %d", k, k, fk))
			}
			return out, nil
		})
	fib.Limits = "0 <= n <= 92"
	return fib
}

func bitwiseNum(id, name, file, fn, summary, time string, ex int, run func(x uint64, in Input, r trace.Recorder) (Output, error)) Entry {
	e := entry(id, name, "math/bitwise/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{ex}}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 0 {
			return Output{}, inputError("give a number >= 0")
		}
		return run(uint64(n), in, r)
	}
	e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("bitwise.%s(%d)", fn, firstOr(in.Values, 0)) }
	return e
}

func popcountEntry() Entry {
	pop := bitwiseNum("popcount", "Population Count (Kernighan)", "bitwise.go", "PopCount", "Count the set bits by clearing the lowest one with x & (x - 1) until none remain: one step per set bit.", "O(set bits)", 2024,
		func(x uint64, _ Input, r trace.Recorder) (Output, error) {
			w := max(bits.Len64(x), 1)
			out := Output{Lines: []string{binary(x, w)}}
			rec := &trace.Log{Limit: 64}
			n := bitwise.TracePopCount(x, tee{rec, r})
			for _, e := range rec.Events {
				x &^= 1 << e.Indices[0]
				out.Lines = append(out.Lines, fmt.Sprintf("%s  cleared bit %d", binary(x, w), e.Indices[0]))
			}
			out.Number = ptr(n)
			return out, nil
		})
	pop.Traced = true
	return pop
}

func powerOfTwoEntry() Entry {
	pow := bitwiseNum("power-of-two", "Power of Two", "power.go", "IsPowerOfTwo", "x is a power of two when x & (x - 1) is zero; the next power of two comes from copying the top bit into every lower bit.", "O(1)", 40,
		func(x uint64, _ Input, _ trace.Recorder) (Output, error) {
			ok := bitwise.IsPowerOfTwo(x)
			next := bitwise.NextPowerOfTwo(x)
			w := max(bits.Len64(x), 1)
			out := Output{Bool: ptr(ok), Lines: []string{"0 is not a power of two"}}
			if x > 0 {
				out.Lines = []string{binary(x, w) + "  x", binary(x-1, w) + "  x - 1", binary(x&(x-1), w) + "  x & (x - 1)"}
			}
			out.Lines = append(out.Lines, fmt.Sprintf("next power of two: %d (%b)", next, next))
			if next != 0 && next <= math.MaxInt64 {
				out.Number = ptr(int(next))
			}
			return out, nil
		})
	return pow
}

func grayCodeEntry() Entry {
	gray := bitwiseNum("gray-code", "Gray Code", "gray.go", "GraySequence", "Number the values so neighbors differ in one bit: x ^ (x >> 1), decoded by XOR-folding the shifts.", "O(2^bits)", 3,
		func(b uint64, _ Input, _ trace.Recorder) (Output, error) {
			if b > 8 {
				return Output{}, inputError("give a width of 0 to 8 bits")
			}
			seq := bitwise.GraySequence(int(b))
			out := Output{}
			w := max(int(b), 1)
			for i, g := range seq {
				out.Values = append(out.Values, int(g))
				out.Lines = append(out.Lines, fmt.Sprintf("%3d  %s  gray %s  decodes to %d", i, binary(uint64(i), w), binary(g, w), bitwise.GrayDecode(g)))
			}
			return out, nil
		})
	gray.Limits = "a width of 0 to 8 bits"
	return gray
}

func reverseBitsEntry() Entry {
	rev := bitwiseNum("reverse-bits", "Reverse Bits", "tricks.go", "ReverseBits", "Reverse the lowest bits of x, moving one bit at a time from the bottom of x to the bottom of the result.", "O(width)", 11,
		func(x uint64, in Input, _ trace.Recorder) (Output, error) {
			w := in.opt("width", 8)
			if w < 1 || w > 62 {
				return Output{}, inputError("the width is 1 to 62")
			}
			y := bitwise.ReverseBits(x, w)
			return Output{Number: ptr(int(y)), Lines: []string{binary(x&(1<<w-1), w) + "  x", binary(y, w) + "  reversed"}}, nil
		})
	rev.Options = []Option{{"width", "width in bits", 8, 1, 62}}
	rev.Example.Options = map[string]int{"width": 8}
	rev.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("bitwise.ReverseBits(%d, %d)", firstOr(in.Values, 0), in.opt("width", 8))
	}
	return rev
}

func bitSubsetsEntry() Entry {
	subsets := bitwiseNum("bit-subsets", "Submask Enumeration", "tricks.go", "BitSubsets", "Every submask of a mask with sub = (sub - 1) & mask, from the mask itself down to 0; the core of subset dynamic programming.", "O(2^k) for k set bits", 13,
		func(x uint64, _ Input, _ trace.Recorder) (Output, error) {
			if bits.OnesCount64(x) > 10 {
				return Output{}, inputError("give a mask with at most 10 bits set")
			}
			out := Output{}
			w := max(bits.Len64(x), 1)
			for _, s := range bitwise.BitSubsets(x) {
				out.Values = append(out.Values, int(s))
				out.Lines = append(out.Lines, binary(s, w))
			}
			out.Number = ptr(len(out.Values))
			return out, nil
		})
	subsets.Limits = "at most 10 bits set"
	return subsets
}

func singleNumberEntry() Entry {
	single := entry("single-number", "Single Number (XOR)", "math/bitwise/tricks.go", "SingleNumber")
	single.Summary, single.Time, single.Space, single.Input = "The one value without a pair: XOR everything, and every pair cancels itself (v ^ v = 0).", "O(n)", "O(1)", Ints
	single.Example = &Input{Values: []int{4, 1, 2, 1, 2, 7, 7}}
	single.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		out := Output{Number: ptr(bitwise.SingleNumber(in.Values))}
		x := 0
		for i, v := range in.Values {
			x ^= v
			if i < 30 {
				out.Lines = append(out.Lines, fmt.Sprintf("^ %d = %d", v, x))
			}
		}
		return out, nil
	}
	single.Call = func(in Input) (string, string) { return "", "bitwise.SingleNumber(" + ints(in.Values) + ")" }
	return single
}

// tee records to two recorders.
type tee struct{ a, b trace.Recorder }

func (t tee) Record(e trace.Event) { t.a.Record(e); t.b.Record(e) }

// polyText writes coefficients, constant first, as a polynomial.
func polyText(c []int) string {
	var parts []string
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == 0 {
			continue
		}
		term := strconv.Itoa(c[i])
		switch i {
		case 0:
		case 1:
			term += "x"
		default:
			term += "x^" + strconv.Itoa(i)
		}
		parts = append(parts, term)
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.ReplaceAll(strings.Join(parts, " + "), "+ -", "- ")
}

// parseMatrix reads rows of whole numbers, one row a line.
func parseMatrix(text string) ([][]int, error) {
	var m [][]int
	for i, line := range strings.Split(text, "\n") {
		f := splitWords(strings.ReplaceAll(line, "|", " "))
		if len(f) == 0 {
			continue
		}
		row := make([]int, len(f))
		for j, w := range f {
			v, err := strconv.Atoi(w)
			if err != nil {
				return nil, inputError(fmt.Sprintf("line %d: %q is not a whole number", i+1, w))
			}
			row[j] = v
		}
		if len(m) > 0 && len(row) != len(m[0]) {
			return nil, inputError(fmt.Sprintf("line %d: every row needs %d numbers", i+1, len(m[0])))
		}
		m = append(m, row)
	}
	if len(m) == 0 || len(m) > 16 {
		return nil, inputError("give 1 to 16 rows")
	}
	return m, nil
}

// matrixLiteral writes m as a Go [][]int literal.
func matrixLiteral(m [][]int) string {
	rows := make([]string, len(m))
	for i, r := range m {
		rows[i] = "{" + strings.TrimSuffix(strings.TrimPrefix(ints(r), "[]int{"), "}") + "}"
	}
	return "[][]int{" + strings.Join(rows, ", ") + "}"
}

func matrixCall(in Input, call func(m string) string) (string, string) {
	m, err := parseMatrix(in.Text)
	if err != nil {
		return "", ""
	}
	return "", call(matrixLiteral(m))
}

func numericalRoots(id, name, file, fn, summary, time string, opts []Option, solve func(c []float64, in Input) (float64, []float64, error), call func(in Input) string) Entry {
	e := entry(id, name, "math/numerical/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Options = summary, time, "O(steps)", Ints, opts
	ex := map[string]int{}
	for _, o := range opts {
		ex[o.Name] = o.Default
	}
	e.Example = &Input{Values: []int{-2, 0, 1}, Options: ex}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := coefficients(in)
		if err != nil {
			return Output{}, err
		}
		root, tried, err := solve(c, in)
		f := numerical.Poly(c)
		out := Output{Lines: []string{"f(x) = " + polyText(in.Values)}}
		for i, x := range tried {
			out.Lines = append(out.Lines, fmt.Sprintf("step %2d  x = %.12g  f(x) = %.3g", i+1, x, f(x)))
		}
		switch {
		case errors.Is(err, numerical.ErrBracket):
			return Output{}, inputError("f(lo) and f(hi) have the same sign, so no root is bracketed")
		case err != nil:
			out.Lines = append(out.Lines, "did not converge: "+err.Error())
		default:
			out.Lines = append(out.Lines, fmt.Sprintf("root %.12g", root))
		}
		out.Summary = map[string]float64{"root": root, "steps": float64(len(tried)), "f(root)": f(root)}
		return out, nil
	}
	e.Call = func(in Input) (string, string) { return "", call(in) }
	return e
}

var numericalIter = Option{"iterations", "most steps", 60, 1, 1000}

func numericalIntegrate(id, name, fn, summary, time string, rule func(f func(float64) float64, a, b float64, n int) float64) Entry {
	e := entry(id, name, "math/numerical/integrate.go", fn)
	e.Summary, e.Time, e.Space, e.Input = summary, time, "O(1)", Ints
	e.Options = []Option{{"a", "from", 0, -10_000, 10_000}, {"b", "to", 2, -10_000, 10_000}, {"n", "slices", 4, 1, 1_000_000}}
	e.Example = &Input{Values: []int{1, 0, -3, 2}, Options: map[string]int{"a": 0, "b": 2, "n": 4}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := coefficients(in)
		if err != nil {
			return Output{}, err
		}
		a, b, n := float64(in.opt("a", 0)), float64(in.opt("b", 2)), in.opt("n", 4)
		if n < 1 || n > 1_000_000 {
			return Output{}, inputError("give 1 to 1000000 slices")
		}
		got := rule(numerical.Poly(c), a, b, n)
		exact := 0.0
		for i, ci := range c {
			p := float64(i + 1)
			exact += ci * (math.Pow(b, p) - math.Pow(a, p)) / p
		}
		out := Output{Lines: []string{
			fmt.Sprintf("integral of %s from %g to %g", polyText(in.Values), a, b),
			fmt.Sprintf("%s with %d slices: %.12g", name, n, got),
			fmt.Sprintf("exact (from the antiderivative): %.12g", exact),
		}}
		out.Summary = map[string]float64{"estimate": got, "exact": exact, "error": math.Abs(got - exact)}
		return out, nil
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numerical.%s(numerical.Poly(%s), %d, %d, %d)", fn, floats(in.Values), in.opt("a", 0), in.opt("b", 2), in.opt("n", 4))
	}
	return e
}

func bisectionEntry() Entry {
	bis := numericalRoots("bisection", "Bisection Method", "roots.go", "Bisection", "Halve an interval whose ends have opposite signs, keeping the half where the sign changes: slow, one bit a step, but sure.", "O(log(width / tol))",
		[]Option{{"lo", "interval start", 0, -1_000_000, 1_000_000}, {"hi", "interval end", 2, -1_000_000, 1_000_000}, numericalIter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.Bisection(numerical.Poly(c), float64(in.opt("lo", 0)), float64(in.opt("hi", 2)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			return fmt.Sprintf("numerical.Bisection(numerical.Poly(%s), %d, %d, 1e-12, %d)", floats(in.Values), in.opt("lo", 0), in.opt("hi", 2), in.opt("iterations", 60))
		})
	return bis
}

func newtonRaphsonEntry() Entry {
	newton := numericalRoots("newton-raphson", "Newton-Raphson Method", "roots.go", "NewtonRaphson", "Follow the tangent line to where it crosses zero, x - f(x)/f'(x); near a root the correct digits double each step.", "O(steps)",
		[]Option{{"x0", "first guess", 1, -1_000_000, 1_000_000}, numericalIter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.NewtonRaphson(numerical.Poly(c), numerical.Poly(numerical.Derivative(c)), float64(in.opt("x0", 1)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			c := floats(in.Values)
			return fmt.Sprintf("numerical.NewtonRaphson(numerical.Poly(%s), numerical.Poly(numerical.Derivative(%s)), %d, 1e-12, %d)", c, c, in.opt("x0", 1), in.opt("iterations", 60))
		})
	return newton
}

func secantMethodEntry() Entry {
	secant := numericalRoots("secant-method", "Secant Method", "roots.go", "Secant", "Newton's method with the derivative replaced by the slope through the last two guesses; no derivative needed.", "O(steps)",
		[]Option{{"x0", "first guess", 1, -1_000_000, 1_000_000}, {"x1", "second guess", 2, -1_000_000, 1_000_000}, numericalIter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.Secant(numerical.Poly(c), float64(in.opt("x0", 1)), float64(in.opt("x1", 2)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			return fmt.Sprintf("numerical.Secant(numerical.Poly(%s), %d, %d, 1e-12, %d)", floats(in.Values), in.opt("x0", 1), in.opt("x1", 2), in.opt("iterations", 60))
		})
	return secant
}

func gaussianEliminationEntry() Entry {
	gauss := entry("gaussian-elimination", "Gaussian Elimination", "math/numerical/gauss.go", "GaussianElimination")
	gauss.Summary, gauss.Time, gauss.Space, gauss.Input = "Solve A x = b: eliminate below each pivot, swapping in the largest pivot first (partial pivoting), then substitute back.", "O(n^3)", "O(n^2)", Text
	gauss.Example = &Input{Text: "2 1 -1 | 8\n-3 -1 2 | -11\n-2 1 2 | -3"}
	gauss.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		if len(m[0]) != len(m)+1 {
			return Output{}, inputError("give n rows of n + 1 numbers: the coefficients, then the right-hand side")
		}
		a := make([][]float64, len(m))
		b := make([]float64, len(m))
		for i, row := range m {
			a[i], b[i] = toFloats(row[:len(row)-1]), float64(row[len(row)-1])
		}
		x, steps, err := numerical.GaussianEliminationSteps(a, b)
		if err != nil {
			return Output{}, inputError("the matrix is singular: no unique solution")
		}
		out := Output{Summary: map[string]float64{}}
		for _, s := range steps {
			out.Lines = append(out.Lines, s.Note)
			for _, row := range s.Matrix {
				cells := make([]string, len(row))
				for j, v := range row {
					cells[j] = fmt.Sprintf("%8.4g", v)
				}
				out.Lines = append(out.Lines, "  "+strings.Join(cells[:len(row)-1], " ")+" | "+cells[len(row)-1])
			}
		}
		for i, v := range x {
			name := fmt.Sprintf("x%d", i+1)
			out.Summary[name] = v
			out.Lines = append(out.Lines, fmt.Sprintf("%s = %.10g", name, v))
		}
		return out, nil
	}
	gauss.Call = func(in Input) (string, string) {
		m, err := parseMatrix(in.Text)
		if err != nil || len(m[0]) != len(m)+1 {
			return "", ""
		}
		rows := make([]string, len(m))
		var b []int
		for i, row := range m {
			rows[i] = "{" + strings.TrimPrefix(floats(row[:len(row)-1]), "[]float64{")
			b = append(b, row[len(row)-1])
		}
		return "", fmt.Sprintf("numerical.GaussianElimination([][]float64{%s}, %s)", strings.Join(rows, ", "), floats(b))
	}
	return gauss
}

func matrixPowerEntry() Entry {
	matPow := entry("matrix-power", "Matrix Power", "math/numerical/matrix.go", "MatrixPower")
	matPow.Summary, matPow.Time, matPow.Space, matPow.Input = "A square matrix to the kth power by repeated squaring: O(log k) products. [[1 1] [1 0]]^n holds Fibonacci numbers.", "O(n^3 log k)", "O(n^2)", Text
	matPow.Options = []Option{{"k", "power", 10, 0, 1_000_000}}
	matPow.Example = &Input{Text: "1 1\n1 0", Options: map[string]int{"k": 10}}
	matPow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		p, err := numerical.MatrixPower(m, in.opt("k", 10))
		if err != nil {
			return Output{}, inputError("give a square matrix and k >= 0")
		}
		k := in.opt("k", 10)
		return Output{Table: tableOf(p, nil, nil), Lines: []string{
			fmt.Sprintf("k = %d is %b in binary: square once per bit, multiply in once per set bit", k, k),
			"entries wrap around on overflow",
		}}, nil
	}
	matPow.Call = func(in Input) (string, string) {
		return matrixCall(in, func(m string) string { return fmt.Sprintf("numerical.MatrixPower(%s, %d)", m, in.opt("k", 10)) })
	}
	return matPow
}

func strassenEntry() Entry {
	strassen := entry("strassen", "Strassen Multiplication", "math/numerical/matrix.go", "Strassen")
	strassen.Summary, strassen.Time, strassen.Space, strassen.Input = "Multiply square matrices with seven block products instead of eight, recursively: under n^3.", "O(n^2.81)", "O(n^2)", TwoTexts
	strassen.Example = &Input{Text: "1 2 3\n4 5 6\n7 8 9", Other: "9 8 7\n6 5 4\n3 2 1"}
	strassen.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		b, err := parseMatrix(in.Other)
		if err != nil {
			return Output{}, err
		}
		c, err := numerical.Strassen(a, b)
		if err != nil {
			return Output{}, inputError("give two square matrices of the same size")
		}
		size := 1
		for size < len(a) {
			size <<= 1
		}
		return Output{Table: tableOf(c, nil, nil), Lines: []string{fmt.Sprintf("padded to %d x %d; direct multiplication gives the same", size, size)}}, nil
	}
	strassen.Call = func(in Input) (string, string) {
		a, err1 := parseMatrix(in.Text)
		b, err2 := parseMatrix(in.Other)
		if err1 != nil || err2 != nil {
			return "", ""
		}
		return "", fmt.Sprintf("numerical.Strassen(%s, %s)", matrixLiteral(a), matrixLiteral(b))
	}
	return strassen
}

func fftEntry() Entry {
	fft := entry("fft", "Fast Fourier Transform", "math/numerical/fft.go", "FFT")
	fft.Summary, fft.Time, fft.Space, fft.Input = "The frequencies in a signal, from the transforms of its even and odd positions combined with twiddle factors (Cooley-Tukey radix 2).", "O(n log n)", "O(n log n)", Ints
	fft.Example, fft.Limits = &Input{Values: []int{1, 1, 1, 1, 0, 0, 0, 0}}, "at most 1024 values, padded with zeros to a power of two"
	fft.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "transform", 1024); err != nil {
			return Output{}, err
		}
		size := 1
		for size < len(in.Values) {
			size <<= 1
		}
		x := make([]complex128, size)
		for i, v := range in.Values {
			x[i] = complex(float64(v), 0)
		}
		y, _ := numerical.FFT(x)
		out := Output{}
		for k, v := range y {
			if k < 64 {
				out.Lines = append(out.Lines, fmt.Sprintf("X[%d] = %8.4f %+8.4fi   |X| = %.4f", k, real(v), imag(v), math.Hypot(real(v), imag(v))))
			}
		}
		if size > len(in.Values) {
			out.Lines = append(out.Lines, fmt.Sprintf("padded with %d zeros to %d values", size-len(in.Values), size))
		}
		return out, nil
	}
	fft.Call = func(in Input) (string, string) {
		lit := "[]complex128{" + strings.TrimPrefix(ints(in.Values), "[]int{")
		return "x := " + lit + "\n\t", "numerical.FFT(x)"
	}
	return fft
}

func hornerEntry() Entry {
	horner := entry("horner", "Horner's Rule", "math/numerical/numerical.go", "Horner")
	horner.Summary, horner.Time, horner.Space, horner.Input = "Evaluate a polynomial as a + x(b + x(c + ...)): one multiply and one add per coefficient, no powers.", "O(n)", "O(1)", Ints
	horner.Options = []Option{{"x", "x", 3, -1_000_000, 1_000_000}}
	horner.Example = &Input{Values: []int{5, -3, 0, 2}, Options: map[string]int{"x": 3}}
	horner.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := coefficients(in)
		if err != nil {
			return Output{}, err
		}
		x := float64(in.opt("x", 3))
		out := Output{Lines: []string{fmt.Sprintf("p(x) = %s at x = %g", polyText(in.Values), x)}}
		y := 0.0
		for i := len(c) - 1; i >= 0; i-- {
			prev := y
			y = y*x + c[i]
			out.Lines = append(out.Lines, fmt.Sprintf("%g * %g + %g = %g", prev, x, c[i], y))
		}
		out.Summary = map[string]float64{"p(x)": numerical.Horner(c, x)}
		if y == math.Trunc(y) && math.Abs(y) < 1<<53 {
			out.Number = ptr(int(y))
		}
		return out, nil
	}
	horner.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numerical.Horner(%s, %d)", floats(in.Values), in.opt("x", 3))
	}
	return horner
}

func simpsonIntegrationEntry() Entry {
	return numericalIntegrate("simpson-integration", "Simpson's Rule", "Simpson", "Integrate by fitting a parabola through each pair of slices (weights 1, 4, 2, ..., 4, 1); exact for cubics.", "O(n)", numerical.Simpson)
}

func trapezoidRuleEntry() Entry {
	return numericalIntegrate("trapezoid-rule", "Trapezoid Rule", "Trapezoid", "Integrate by joining neighboring points with straight lines and adding the trapezoids' areas; exact for lines.", "O(n)", numerical.Trapezoid)
}

// parseEdges reads an undirected graph, one edge a line ("A B"), as node
// names and index pairs.
func parseEdges(text string) ([]string, [][2]int, error) {
	index := map[string]int{}
	var names []string
	node := func(s string) int {
		if i, ok := index[s]; ok {
			return i
		}
		index[s] = len(names)
		names = append(names, s)
		return len(names) - 1
	}
	var edges [][2]int
	for i, line := range strings.Split(text, "\n") {
		f := splitWords(line)
		if len(f) == 0 {
			continue
		}
		if len(f) < 2 || len(f) > 3 {
			return nil, nil, inputError(fmt.Sprintf("line %d: want an edge \"A B\"", i+1))
		}
		edges = append(edges, [2]int{node(f[0]), node(f[1])})
	}
	if len(names) < 2 || len(names) > 60 || len(edges) > 600 {
		return nil, nil, inputError("give 2 to 60 nodes and at most 600 edges")
	}
	return names, edges, nil
}

func edgesLiteral(edges [][2]int) string {
	parts := make([]string, len(edges))
	for i, e := range edges {
		parts[i] = fmt.Sprintf("{%d, %d}", e[0], e[1])
	}
	return "[][2]int{" + strings.Join(parts, ", ") + "}"
}

func reservoirSamplingEntry() Entry {
	res := entry("reservoir-sampling", "Reservoir Sampling", "math/randomized/randomized.go", "ReservoirSampling")
	res.Summary, res.Time, res.Space, res.Input = "Choose k values uniformly from a stream read once: keep the first k, then let the ith replace a random kept one with probability k/i.", "O(n)", "O(k)", Ints
	res.Options = []Option{{"k", "sample size", 3, 1, 100_000}, seedOption}
	res.Example = &Input{Values: []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, Options: map[string]int{"k": 3, "seed": 1}}
	res.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "sample", 100_000); err != nil {
			return Output{}, err
		}
		k := in.opt("k", 3)
		if k < 1 {
			return Output{}, inputError("k is at least 1")
		}
		s := randomized.ReservoirSampling(in.Values, k, int64(in.opt("seed", 1)))
		return Output{Values: s, Lines: []string{fmt.Sprintf("each value is kept with probability %d/%d", min(k, len(in.Values)), len(in.Values))}}, nil
	}
	res.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.ReservoirSampling(%s, %d, %d)", ints(in.Values), in.opt("k", 3), in.opt("seed", 1))
	}
	return res
}

func monteCarloPiEntry() Entry {
	pi := entry("monte-carlo-pi", "Monte Carlo Pi", "math/randomized/montecarlo.go", "MonteCarloPi")
	pi.Summary, pi.Time, pi.Space, pi.Input, pi.Traced = "Throw random points into a square and count those inside the quarter circle: pi is about 4 * inside / all. The error shrinks like 1/sqrt(n).", "O(n)", "O(1)", Simulation, true
	pi.Options = []Option{{"points", "points", 1000, 1, 1_000_000}, seedOption}
	pi.Example, pi.Limits = &Input{Options: map[string]int{"points": 1000, "seed": 1}}, "at most 1000000 points, the first 2000 drawn"
	pi.Run = func(in Input, r trace.Recorder) (Output, error) {
		n := in.opt("points", 1000)
		if n < 1 || n > 1_000_000 {
			return Output{}, inputError("give 1 to 1000000 points")
		}
		est, samples := randomized.MonteCarloPi(n, int64(in.opt("seed", 1)))
		out := Output{Points: make([]Point, len(samples))}
		inside := 0
		for i, s := range samples {
			out.Points[i] = Point{X: int(s.X * 1000), Y: int(s.Y * 1000)}
			if s.Inside {
				inside++
				r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Note: "inside"})
			}
		}
		out.Summary = map[string]float64{"estimate": est, "error": math.Abs(est - math.Pi), "points": float64(n)}
		out.Lines = []string{fmt.Sprintf("pi is about %.6f after %d points (%d of the %d drawn fell inside)", est, n, inside, len(samples))}
		return out, nil
	}
	pi.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.MonteCarloPi(%d, %d)", in.opt("points", 1000), in.opt("seed", 1))
	}
	return pi
}

func kargerMinCutEntry() Entry {
	karger := entry("karger-min-cut", "Karger's Minimum Cut", "math/randomized/karger.go", "KargerMinCut")
	karger.Summary, karger.Time, karger.Space, karger.Input, karger.Traced = "Merge the ends of random edges until two groups remain; the edges between them are a cut, and the smallest over many trials is probably the minimum.", "O(trials * m)", "O(n + m)", GraphInput, true
	karger.Options = []Option{{"trials", "trials", 50, 1, 10_000}, seedOption}
	karger.Example = &Input{Text: "A B\nA C\nA D\nB C\nB D\nC D\nD E\nC F\nE F\nE G\nE H\nF G\nF H\nG H", Options: map[string]int{"trials": 50, "seed": 1}}
	karger.Run = func(in Input, r trace.Recorder) (Output, error) {
		names, edges, err := parseEdges(in.Text)
		if err != nil {
			return Output{}, err
		}
		trials := in.opt("trials", 50)
		if trials < 1 || trials > 10_000 {
			return Output{}, inputError("give 1 to 10000 trials")
		}
		cut, side, err := randomized.KargerMinCut(len(names), edges, trials, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Number: ptr(cut), Graph: &GraphOutput{Nodes: names}}
		for i, s := range side {
			g := 1
			if s {
				g = 0
			}
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Values: []int{g}, Note: "group"})
		}
		for _, e := range edges {
			out.Graph.Edges = append(out.Graph.Edges, GraphEdge{From: e[0], To: e[1], Weight: 1})
			if side[e[0]] != side[e[1]] {
				out.Lines = append(out.Lines, fmt.Sprintf("cut edge %s - %s", names[e[0]], names[e[1]]))
			}
		}
		out.Lines = append(out.Lines, fmt.Sprintf("smallest cut found in %d trials: %d edges", trials, cut))
		return out, nil
	}
	karger.Call = func(in Input) (string, string) {
		names, edges, err := parseEdges(in.Text)
		if err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("randomized.KargerMinCut(%d, %s, %d, %d)", len(names), edgesLiteral(edges), in.opt("trials", 50), in.opt("seed", 1))
	}
	return karger
}

func randomizedQuickselectEntry() Entry {
	rq := entry("randomized-quickselect", "Randomized Quickselect", "math/randomized/quickselect.go", "RandomizedQuickselect")
	rq.Summary, rq.Time, rq.Space, rq.Input = "The kth smallest value by quickselect around random pivots: always exact (Las Vegas), linear time expected for every input.", "O(n) expected", "O(n)", Ints
	rq.Options = []Option{{"k", "k (0 is the smallest)", 5, 0, 99_999}, seedOption}
	rq.Example = &Input{Values: []int{38, 27, 43, 3, 9, 82, 10, 27, 55, 1, 64, 19}, Options: map[string]int{"k": 5, "seed": 1}}
	rq.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		k := in.opt("k", 5)
		v, ok := randomized.RandomizedQuickselect(in.Values, k, int64(in.opt("seed", 1)))
		if !ok {
			return Output{}, inputError(fmt.Sprintf("k is 0 to %d", len(in.Values)-1))
		}
		return Output{Number: ptr(v)}, nil
	}
	rq.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.RandomizedQuickselect(%s, %d, %d)", ints(in.Values), in.opt("k", 5), in.opt("seed", 1))
	}
	return rq
}
