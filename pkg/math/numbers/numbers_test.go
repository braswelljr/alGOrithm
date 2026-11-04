package numbers_test

import (
	"errors"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestPrimes(t *testing.T) {
	primes := numbers.Sieve(100)
	if len(primes) != 25 || primes[24] != 97 {
		t.Fatalf("Sieve(100) = %v", primes)
	}
	for n := -5; n <= 100; n++ {
		inSieve := false
		for _, p := range primes {
			inSieve = inSieve || p == n
		}
		if numbers.Prime(n) != inSieve {
			t.Errorf("Prime(%d) disagrees with the sieve", n)
		}
	}
	var log trace.Log
	numbers.TraceSieve(30, &log)
	if log.Count(trace.Mark) != 10 {
		t.Errorf("sieve marked %d primes below 30", log.Count(trace.Mark))
	}
	if numbers.Sieve(1) != nil || !numbers.Even(-4) || numbers.Even(3) {
		t.Error("edges wrong")
	}
}

func TestFactorizeAndPower(t *testing.T) {
	if got := numbers.Factorize(360); !reflect.DeepEqual(got, []int{2, 2, 2, 3, 3, 5}) {
		t.Errorf("Factorize(360) = %v", got)
	}
	if numbers.Factorize(97)[0] != 97 || numbers.Factorize(1) != nil {
		t.Error("Factorize edges wrong")
	}
	if numbers.Power(2, 10, 1000) != 24 || numbers.Power(3, 0, 7) != 1 || numbers.Power(-2, 3, 5) != 2 || numbers.Power(5, 3, 1) != 0 {
		t.Error("Power wrong")
	}
}

func TestPaginate(t *testing.T) {
	p := numbers.Paginate(2, 10, 35)
	if p.Pages != 4 || p.Offset != 10 || p.End != 20 || p.Prev != 1 || p.Next != 3 {
		t.Errorf("page 2 of 35 by 10 = %+v", p)
	}
	if last := numbers.Paginate(9, 10, 35); last.Number != 4 || last.End != 35 || last.Next != 0 {
		t.Errorf("past the end = %+v", last)
	}
	if e := numbers.Paginate(1, 10, 0); e.Pages != 1 || e.End != 0 {
		t.Errorf("empty = %+v", e)
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func TestChineseRemainder(t *testing.T) {
	for _, c := range []struct {
		r, m  []int
		x, md int
		ok    bool
	}{
		{[]int{2, 3, 2}, []int{3, 5, 7}, 23, 105, true},
		{[]int{1, 3}, []int{4, 6}, 9, 12, true}, // not coprime, consistent
		{[]int{1, 2}, []int{4, 6}, 0, 0, false}, // not coprime, inconsistent
		{[]int{-1}, []int{5}, 4, 5, true},
		{[]int{0}, []int{0}, 0, 0, false},
		{nil, nil, 0, 0, false},
		{[]int{1}, []int{2, 3}, 0, 0, false},
	} {
		x, m, err := numbers.ChineseRemainder(c.r, c.m)
		if (err == nil) != c.ok || x != c.x || m != c.md {
			t.Errorf("ChineseRemainder(%v, %v) = %d, %d, %v", c.r, c.m, x, m, err)
		}
		if err != nil && !errors.Is(err, numbers.ErrNumber) {
			t.Errorf("error %v is not ErrNumber", err)
		}
	}
	// brute force over small systems
	for m1 := 1; m1 <= 9; m1++ {
		for m2 := 1; m2 <= 9; m2++ {
			for r1 := 0; r1 < m1; r1++ {
				for r2 := 0; r2 < m2; r2++ {
					want := -1
					for x := 0; x < m1*m2; x++ {
						if x%m1 == r1 && x%m2 == r2 {
							want = x
							break
						}
					}
					x, m, err := numbers.ChineseRemainder([]int{r1, r2}, []int{m1, m2})
					if want < 0 {
						if err == nil {
							t.Fatalf("%d mod %d, %d mod %d: want no solution, got %d", r1, m1, r2, m2, x)
						}
						continue
					}
					if err != nil || x != want || m != m1*m2/gcd(m1, m2) {
						t.Fatalf("%d mod %d, %d mod %d = %d, %d, %v; want %d", r1, m1, r2, m2, x, m, err, want)
					}
				}
			}
		}
	}
}

func TestTotient(t *testing.T) {
	sieve := numbers.TotientSieve(500)
	for n := 1; n <= 500; n++ {
		want := 0
		for k := 1; k <= n; k++ {
			if gcd(n, k) == 1 {
				want++
			}
		}
		if got := numbers.EulerTotient(n); got != want || sieve[n] != want {
			t.Fatalf("phi(%d) = %d, sieve %d, want %d", n, got, sieve[n], want)
		}
	}
	if numbers.EulerTotient(0) != 0 || numbers.TotientSieve(-1) != nil || sieve[0] != 0 {
		t.Error("totient edge cases")
	}
}

func TestSegmentedSieve(t *testing.T) {
	all := numbers.Sieve(20000)
	for _, c := range [][2]int{{0, 1}, {0, 2}, {1, 100}, {90, 97}, {1000, 1100}, {19000, 20000}, {2, 20000}} {
		var want []int
		for _, p := range all {
			if p >= c[0] && p <= c[1] {
				want = append(want, p)
			}
		}
		var log trace.Log
		got := numbers.TraceSegmentedSieve(c[0], c[1], &log)
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(numbers.SegmentedSieve(c[0], c[1]), want) {
			t.Fatalf("SegmentedSieve(%d, %d) = %v, want %v", c[0], c[1], got, want)
		}
	}
	if numbers.SegmentedSieve(10, 5) != nil {
		t.Error("an empty range has primes")
	}
	big := numbers.SegmentedSieve(1_000_000_000_000, 1_000_000_000_100)
	for _, p := range big {
		if ok, _ := numbers.MillerRabin(p); !ok {
			t.Fatalf("%d is not prime", p)
		}
	}
	if len(big) != 4 {
		t.Errorf("primes in 10^12..10^12+100 = %v", big)
	}
}

func TestDivisorsAndPerfect(t *testing.T) {
	for n := 1; n <= 300; n++ {
		var want []int
		for d := 1; d <= n; d++ {
			if n%d == 0 {
				want = append(want, d)
			}
		}
		if got := numbers.Divisors(n); !reflect.DeepEqual(got, want) {
			t.Fatalf("Divisors(%d) = %v", n, got)
		}
	}
	if numbers.Divisors(0) != nil {
		t.Error("Divisors(0)")
	}
	if got := numbers.PerfectNumbers(10000); !reflect.DeepEqual(got, []int{6, 28, 496, 8128}) {
		t.Errorf("PerfectNumbers(10000) = %v", got)
	}
}

func TestFibonacciDoubling(t *testing.T) {
	for n := 0; n <= 92; n++ {
		want, _ := numbers.FibonacciMatrix(n)
		if got, err := numbers.FibonacciDoubling(n); err != nil || got != want {
			t.Fatalf("FibonacciDoubling(%d) = %d, %v; want %d", n, got, err, want)
		}
	}
	for _, n := range []int{-1, 93} {
		if _, err := numbers.FibonacciDoubling(n); !errors.Is(err, numbers.ErrNumber) {
			t.Errorf("FibonacciDoubling(%d) error %v", n, err)
		}
	}
}

func TestExtendedGCDAndInverse(t *testing.T) {
	r := rand.New(rand.NewSource(8))
	for i := 0; i < 1000; i++ {
		a, b := r.Intn(2000)-1000, r.Intn(2000)-1000
		g, x, y := numbers.ExtendedGCD(a, b)
		if a*x+b*y != g || g < 0 || g != numbers.BinaryGCD(a, b) {
			t.Fatalf("ExtendedGCD(%d, %d) = %d %d %d", a, b, g, x, y)
		}
	}
	if inv, ok := numbers.ModInverse(3, 11); !ok || inv != 4 {
		t.Errorf("ModInverse(3, 11) = %d, %v", inv, ok)
	}
	if _, ok := numbers.ModInverse(6, 9); ok {
		t.Error("6 has no inverse mod 9")
	}
	if inv, ok := numbers.ModInverse(-3, 11); !ok || (-3*inv%11+11)%11 != 1 {
		t.Errorf("ModInverse(-3, 11) = %d", inv)
	}
}

func TestBinaryGCD(t *testing.T) {
	for _, c := range [][3]int{{48, 18, 6}, {0, 7, 7}, {7, 0, 7}, {0, 0, 0}, {-12, 8, 4}, {1 << 40, 1 << 20, 1 << 20}} {
		if got := numbers.BinaryGCD(c[0], c[1]); got != c[2] {
			t.Errorf("BinaryGCD(%d, %d) = %d", c[0], c[1], got)
		}
	}
}

func TestMillerRabin(t *testing.T) {
	for n := -5; n < 10_000; n++ {
		if got, _ := numbers.MillerRabin(n); got != numbers.Prime(n) {
			t.Fatalf("MillerRabin(%d) = %v", n, got)
		}
	}
	for _, n := range []int{2147483647, 1000000007, 9223372036854775783} {
		if p, _ := numbers.MillerRabin(n); !p {
			t.Errorf("%d is prime", n)
		}
	}
	for _, n := range []int{3215031751, 3825123056546413051, 561, 9223372036854775807} {
		if p, w := numbers.MillerRabin(n); p || w == 0 {
			t.Errorf("%d is composite (strong pseudoprimes and a Carmichael number), witness %d", n, w)
		}
	}
}

func TestKaratsubaAgainstBig(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	digits := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('0' + r.Intn(10))
		}
		return string(b)
	}
	for i := 0; i < 200; i++ {
		x, y := digits(1+r.Intn(300)), digits(1+r.Intn(300))
		got, err := numbers.Karatsuba(x, y)
		a, _ := new(big.Int).SetString(x, 10)
		b, _ := new(big.Int).SetString(y, 10)
		if err != nil || got != new(big.Int).Mul(a, b).String() {
			t.Fatalf("Karatsuba(%s, %s) wrong", x, y)
		}
	}
	if got, _ := numbers.Karatsuba("0", "123"); got != "0" {
		t.Errorf("0 * 123 = %s", got)
	}
	if _, err := numbers.Karatsuba("12a", "3"); !errors.Is(err, numbers.ErrNumber) {
		t.Error("a letter was accepted")
	}
}

func TestSequences(t *testing.T) {
	if got := numbers.Collatz(6); !reflect.DeepEqual(got, []int{6, 3, 10, 5, 16, 8, 4, 2, 1}) {
		t.Errorf("Collatz(6) = %v", got)
	}
	if len(numbers.Collatz(27)) != 112 || numbers.Collatz(0) != nil {
		t.Error("Collatz(27) has 111 steps")
	}
	p := numbers.PascalTriangle(67)
	if !reflect.DeepEqual(p[4], []int{1, 4, 6, 4, 1}) || p[66][33] != 7219428434016265740 || numbers.PascalTriangle(68) != nil {
		t.Error("PascalTriangle wrong")
	}
	for _, c := range []struct {
		in       string
		from, to int
		want     string
	}{{"255", 10, 16, "ff"}, {"FF", 16, 2, "11111111"}, {"-z", 36, 10, "-35"}, {"0", 10, 2, "0"}} {
		if got, err := numbers.BaseConvert(c.in, c.from, c.to); err != nil || got != c.want {
			t.Errorf("BaseConvert(%q, %d, %d) = %q, %v", c.in, c.from, c.to, got, err)
		}
	}
	if _, err := numbers.BaseConvert("12", 2, 10); !errors.Is(err, numbers.ErrNumber) {
		t.Error("2 is not a binary digit")
	}
	for n := 0; n < 5000; n++ {
		root, steps := numbers.NewtonSqrt(n)
		if root*root > n || (root+1)*(root+1) <= n || len(steps) == 0 {
			t.Fatalf("NewtonSqrt(%d) = %d", n, root)
		}
	}
	if r, _ := numbers.NewtonSqrt(1 << 62); r != 1<<31 {
		t.Errorf("NewtonSqrt(2^62) = %d", r)
	}
	a, b := 0, 1
	for n := 0; n <= 92; n++ {
		if got, err := numbers.FibonacciMatrix(n); err != nil || got != a {
			t.Fatalf("FibonacciMatrix(%d) = %d, want %d", n, got, a)
		}
		a, b = b, a+b
	}
	if _, err := numbers.FibonacciMatrix(93); err == nil {
		t.Error("F(93) overflows")
	}
}
