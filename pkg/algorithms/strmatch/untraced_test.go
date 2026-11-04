package strmatch_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// TestUntracedMatchersAgreeWithNaive runs KMP, Rabin-Karp, Boyer-Moore-Horspool
// and the Z-algorithm without a recorder on random texts over a small
// alphabet, where matches overlap often: each finds exactly what the naive
// search finds.
func TestUntracedMatchersAgreeWithNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	word := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = "ab"[rng.Intn(2)]
		}
		return string(b)
	}
	matchers := map[string]func(text, pattern string) []int{
		"kmp":        strmatch.KMP,
		"rabin-karp": strmatch.RabinKarp,
		"horspool":   strmatch.BoyerMooreHorspool,
		"z":          strmatch.ZSearch,
	}
	for i := 0; i < 300; i++ {
		text, pattern := word(rng.Intn(40)), word(1+rng.Intn(5))
		want := strmatch.Naive(text, pattern)
		for name, f := range matchers {
			got := f(text, pattern)
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s(%q, %q) = %v, want %v", name, text, pattern, got, want)
			}
		}
	}
	if got := strmatch.KMP("aaaa", "aa"); !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Errorf("KMP(aaaa, aa) = %v, want overlapping matches [0 1 2]", got)
	}
}

// TestTraceSuffixArrayAgrees checks the traced suffix array against the
// untraced one, and both against "banana"'s textbook array and LCPs.
func TestTraceSuffixArrayAgrees(t *testing.T) {
	sa, lcp := strmatch.SuffixArray("banana")
	var log trace.Log
	tsa, tlcp := strmatch.TraceSuffixArray("banana", &log)
	if !reflect.DeepEqual(sa, tsa) || !reflect.DeepEqual(lcp, tlcp) {
		t.Fatalf("traced (%v, %v), untraced (%v, %v)", tsa, tlcp, sa, lcp)
	}
	if !reflect.DeepEqual(sa, []int{5, 3, 1, 0, 4, 2}) {
		t.Errorf("SuffixArray(banana) = %v, want [5 3 1 0 4 2]", sa)
	}
	if log.Total() == 0 {
		t.Error("TraceSuffixArray reported nothing")
	}
}
