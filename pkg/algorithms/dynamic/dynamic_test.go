package dynamic

import (
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestDistance(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{{"kitten", "sitting", 3}, {"", "abc", 3}, {"abc", "", 3}, {"", "", 0}, {"flaw", "lawn", 2}, {"same", "same", 0}} {
		var log trace.Log
		got, table := TraceEditDistance(tc.a, tc.b, &log)
		if got != tc.want || table[len(tc.a)][len(tc.b)] != tc.want {
			t.Errorf("Distance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
		if log.Count(trace.Set) != (len(tc.a)+1)*(len(tc.b)+1) {
			t.Errorf("%q/%q: %d cells set, want every cell once", tc.a, tc.b, log.Count(trace.Set))
		}
		if d, _ := EditDistance(tc.a, tc.b); d != got {
			t.Error("Distance and Trace disagree")
		}
	}
}

// isSubsequence reports whether s appears in w in order.
func isSubsequence(s, w string) bool {
	i := 0
	for j := 0; j < len(w) && i < len(s); j++ {
		if s[i] == w[j] {
			i++
		}
	}
	return i == len(s)
}

func TestLCS(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{{"ABCBDAB", "BDCABA", 4}, {"AGGTAB", "GXTXAYB", 4}, {"", "abc", 0}, {"abc", "def", 0}, {"same", "same", 4}} {
		got, table := LCS(tc.a, tc.b)
		if len(got) != tc.want || table[len(tc.a)][len(tc.b)] != tc.want {
			t.Errorf("LCS(%q, %q) = %q (%d), want length %d", tc.a, tc.b, got, len(got), tc.want)
		}
		if !isSubsequence(got, tc.a) || !isSubsequence(got, tc.b) {
			t.Errorf("LCS(%q, %q) = %q is not in both", tc.a, tc.b, got)
		}
	}
}

func TestMinCoins(t *testing.T) {
	for _, tc := range []struct {
		coins  []int
		amount int
		want   int
	}{{[]int{1, 2, 5}, 11, 3}, {[]int{2}, 3, -1}, {[]int{1}, 0, 0}, {[]int{1, 3, 4}, 6, 2}, {[]int{5, 10}, 1, -1}} {
		got, used, table, err := CoinChange(tc.coins, tc.amount)
		if err != nil || got != tc.want || table[tc.amount] != tc.want {
			t.Errorf("CoinChange(%v, %d) = %d, %v; want %d", tc.coins, tc.amount, got, err, tc.want)
		}
		sum := 0
		for _, c := range used {
			sum += c
		}
		if got >= 0 && (sum != tc.amount || len(used) != got) {
			t.Errorf("CoinChange(%v, %d) used %v", tc.coins, tc.amount, used)
		}
	}
	for _, bad := range [][]int{{0}, {-1}} {
		if _, _, _, err := CoinChange(bad, 3); !errors.Is(err, ErrAmount) {
			t.Errorf("coins %v accepted", bad)
		}
	}
	if _, _, _, err := CoinChange([]int{1}, MaxAmount+1); !errors.Is(err, ErrAmount) {
		t.Error("oversized amount accepted")
	}
}

func TestLIS(t *testing.T) {
	for _, tc := range []struct {
		in   []int
		want int
	}{{[]int{10, 9, 2, 5, 3, 7, 101, 18}, 4}, {nil, 0}, {[]int{5}, 1}, {[]int{3, 3, 3}, 1}, {[]int{5, 4, 3}, 1}, {[]int{-3, -1, 0, 4}, 4}} {
		got, _ := LIS(tc.in)
		if len(got) != tc.want {
			t.Errorf("LIS(%v) = %v, want length %d", tc.in, got, tc.want)
		}
		for i := 1; i < len(got); i++ {
			if got[i] <= got[i-1] {
				t.Errorf("LIS(%v) = %v is not strictly increasing", tc.in, got)
			}
		}
	}
}

func TestKnapsack01AgainstBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 200; trial++ {
		n := r.Intn(9)
		w, v := make([]int, n), make([]int, n)
		for i := range w {
			w[i], v[i] = 1+r.Intn(10), r.Intn(20)
		}
		capacity := r.Intn(30)
		want := 0
		for mask := 0; mask < 1<<n; mask++ {
			tw, tv := 0, 0
			for i := 0; i < n; i++ {
				if mask&(1<<i) != 0 {
					tw, tv = tw+w[i], tv+v[i]
				}
			}
			if tw <= capacity {
				want = max(want, tv)
			}
		}
		var log trace.Log
		got, chosen, table, err := TraceKnapsack01(w, v, capacity, &log)
		if err != nil || got != want || table[n][capacity] != want {
			t.Fatalf("Knapsack01(%v, %v, %d) = %d, %v; want %d", w, v, capacity, got, err, want)
		}
		tw, tv := 0, 0
		for _, i := range chosen {
			tw, tv = tw+w[i], tv+v[i]
		}
		if tw > capacity || tv != want {
			t.Fatalf("chosen %v weighs %d and is worth %d", chosen, tw, tv)
		}
	}
	if _, _, _, err := Knapsack01([]int{1}, nil, 3); !errors.Is(err, ErrInput) {
		t.Error("mismatched lists accepted")
	}
}

func TestMatrixChain(t *testing.T) {
	cost, order, _, err := MatrixChain([]int{40, 20, 30, 10, 30})
	if err != nil || cost != 26000 || order != "((A1(A2A3))A4)" {
		t.Errorf("MatrixChain = %d %q %v", cost, order, err)
	}
	if cost, order, _, _ := MatrixChain([]int{10, 20}); cost != 0 || order != "A1" {
		t.Errorf("one matrix: %d %q", cost, order)
	}
	if _, _, _, err := MatrixChain([]int{5}); !errors.Is(err, ErrInput) {
		t.Error("no matrices accepted")
	}
}

func TestRodCutting(t *testing.T) {
	best, cuts, _, err := RodCutting([]int{1, 5, 8, 9, 10, 17, 17, 20})
	if err != nil || best != 22 {
		t.Fatalf("RodCutting = %d, %v", best, err)
	}
	sum, rev := 0, 0
	prices := []int{1, 5, 8, 9, 10, 17, 17, 20}
	for _, c := range cuts {
		sum, rev = sum+c, rev+prices[c-1]
	}
	if sum != 8 || rev != 22 {
		t.Errorf("cuts %v sum to %d and earn %d", cuts, sum, rev)
	}
}

func TestSubsetSumAgainstBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for trial := 0; trial < 200; trial++ {
		n := r.Intn(9)
		vals := make([]int, n)
		for i := range vals {
			vals[i] = r.Intn(15)
		}
		target := r.Intn(40)
		want := false
		for mask := 0; mask < 1<<n && !want; mask++ {
			s := 0
			for i := 0; i < n; i++ {
				if mask&(1<<i) != 0 {
					s += vals[i]
				}
			}
			want = s == target
		}
		ok, chosen, _, err := SubsetSum(vals, target)
		if err != nil || ok != want {
			t.Fatalf("SubsetSum(%v, %d) = %v, want %v", vals, target, ok, want)
		}
		s := 0
		for _, i := range chosen {
			s += vals[i]
		}
		if ok && s != target {
			t.Fatalf("chosen %v sums to %d", chosen, s)
		}
	}
}

func isPalindrome(s string) bool {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		if s[i] != s[j] {
			return false
		}
	}
	return true
}

func TestLongestPalindromicSubsequence(t *testing.T) {
	for s, want := range map[string]int{"bbbab": 4, "cbbd": 2, "": 0, "a": 1, "character": 5, "agbdba": 5} {
		got, _, err := LongestPalindromicSubsequence(s)
		if err != nil || len(got) != want || !isPalindrome(got) || !isSubsequence(got, s) {
			t.Errorf("LPS(%q) = %q, want length %d", s, got, want)
		}
	}
}

func TestUniquePaths(t *testing.T) {
	if n, _, _ := UniquePaths("...\n...\n..."); n != 6 {
		t.Errorf("3x3 open = %d", n)
	}
	if n, _, _ := UniquePaths("...\n.#.\n..."); n != 2 {
		t.Errorf("3x3 with a center obstacle = %d", n)
	}
	if n, _, _ := UniquePaths("#.\n.."); n != 0 {
		t.Errorf("blocked start = %d", n)
	}
	if _, _, err := UniquePaths(strings.Repeat(strings.Repeat(".", 31)+"\n", 2)); !errors.Is(err, ErrInput) {
		t.Error("31 columns accepted")
	}
}

// eggDropSlow is the textbook O(e * f^2) recurrence, for checking.
func eggDropSlow(e, f int) int {
	dp := make([][]int, e+1)
	for i := range dp {
		dp[i] = make([]int, f+1)
	}
	for j := 0; j <= f; j++ {
		dp[1][j] = j
	}
	for i := 2; i <= e; i++ {
		for j := 1; j <= f; j++ {
			dp[i][j] = j
			for x := 1; x <= j; x++ {
				dp[i][j] = min(dp[i][j], 1+max(dp[i-1][x-1], dp[i][j-x]))
			}
		}
	}
	return dp[e][f]
}

func TestEggDrop(t *testing.T) {
	for e := 1; e <= 4; e++ {
		for f := 0; f <= 40; f++ {
			if got, _, err := EggDrop(e, f); err != nil || got != eggDropSlow(e, f) {
				t.Fatalf("EggDrop(%d, %d) = %d, want %d", e, f, got, eggDropSlow(e, f))
			}
		}
	}
	if got, _, _ := EggDrop(2, 100); got != 14 {
		t.Errorf("EggDrop(2, 100) = %d, want 14", got)
	}
}

func TestLISFastAgreesWithLIS(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for trial := 0; trial < 300; trial++ {
		in := make([]int, r.Intn(40))
		for i := range in {
			in[i] = r.Intn(30) - 10
		}
		slow, _ := LIS(in)
		n, seq, _ := LISFast(in)
		if n != len(slow) || len(seq) != n {
			t.Fatalf("LISFast(%v) = %d, LIS = %d", in, n, len(slow))
		}
		for i := 1; i < len(seq); i++ {
			if seq[i] <= seq[i-1] {
				t.Fatalf("not increasing: %v", seq)
			}
		}
	}
	if n, seq, tails := LISFast([]int{10, 9, 2, 5, 3, 7, 101, 18}); n != 4 || !reflect.DeepEqual(tails, []int{2, 3, 7, 18}) || len(seq) != 4 {
		t.Errorf("LISFast = %d %v %v", n, seq, tails)
	}
}
