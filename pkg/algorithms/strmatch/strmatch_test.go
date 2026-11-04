package strmatch_test

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// all is the answer: every start index, by strings.Index from each position.
func all(text, pattern string) []int {
	var out []int
	if pattern == "" {
		return out
	}
	for i := 0; i+len(pattern) <= len(text); i++ {
		if strings.HasPrefix(text[i:], pattern) {
			out = append(out, i)
		}
	}
	return out
}

var matchers = map[string]func(string, string, trace.Recorder) []int{
	"naive": strmatch.TraceNaive,
	"kmp":   strmatch.TraceKMP,
	"rabin": strmatch.TraceRabinKarp,
}

func TestMatchers(t *testing.T) {
	cases := [][2]string{
		{"abracadabra", "abra"}, {"aaaaa", "aa"}, {"hello", "xyz"}, {"", "a"}, {"abc", ""},
		{"abc", "abcd"}, {"ababcabab", "abab"}, {"mississippi", "issi"}, {"x", "x"},
	}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		text := make([]byte, r.Intn(40))
		for j := range text {
			text[j] = "ab"[r.Intn(2)]
		}
		pat := make([]byte, 1+r.Intn(4))
		for j := range pat {
			pat[j] = "ab"[r.Intn(2)]
		}
		cases = append(cases, [2]string{string(text), string(pat)})
	}
	for name, f := range matchers {
		for _, c := range cases {
			var log trace.Log
			got := f(c[0], c[1], &log)
			if want := all(c[0], c[1]); len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
				t.Fatalf("%s(%q, %q) = %v, want %v", name, c[0], c[1], got, want)
			}
		}
	}
	if !reflect.DeepEqual(strmatch.KMPTable("ababaca"), []int{0, 0, 1, 2, 3, 0, 1}) {
		t.Errorf("KMP table = %v", strmatch.KMPTable("ababaca"))
	}
}

func TestAnagram(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"listen", "silent", true}, {"Dormitory", "dirty room", true}, {"abc", "abd", false}, {"aab", "ab", false}, {"", "", true}} {
		if got := strmatch.Anagram(c.a, c.b); got != c.want {
			t.Errorf("Anagram(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func randomText(r *rand.Rand, n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

func TestMoreMatchersAgreeWithNaive(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for i := 0; i < 500; i++ {
		text, pat := randomText(r, r.Intn(50), "abc"), randomText(r, 1+r.Intn(4), "abc")
		want := strmatch.Naive(text, pat)
		var log trace.Log
		for name, got := range map[string][]int{
			"horspool": strmatch.TraceBoyerMooreHorspool(text, pat, &log),
			"z":        strmatch.TraceZSearch(text, pat, &log),
		} {
			if len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
				t.Fatalf("%s(%q, %q) = %v, want %v", name, text, pat, got, want)
			}
		}
	}
}

func TestZArray(t *testing.T) {
	if got := strmatch.ZArray("aabcaabxaaaz"); !reflect.DeepEqual(got, []int{12, 1, 0, 0, 3, 1, 0, 0, 2, 2, 1, 0}) {
		t.Errorf("ZArray = %v", got)
	}
	if len(strmatch.ZArray("")) != 0 {
		t.Error("ZArray of empty")
	}
}

func TestAhoCorasickAgainstNaive(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	for i := 0; i < 300; i++ {
		text := randomText(r, r.Intn(60), "ab")
		var pats []string
		for k := 0; k < 1+r.Intn(5); k++ {
			pats = append(pats, randomText(r, 1+r.Intn(4), "ab"))
		}
		var want []strmatch.Match
		for p, pat := range pats {
			for _, s := range strmatch.Naive(text, pat) {
				want = append(want, strmatch.Match{Pattern: p, Start: s})
			}
		}
		got := strmatch.AhoCorasick(text, pats)
		key := func(m []strmatch.Match) []strmatch.Match {
			m = append([]strmatch.Match(nil), m...)
			sort.Slice(m, func(a, b int) bool {
				if m[a].Pattern != m[b].Pattern {
					return m[a].Pattern < m[b].Pattern
				}
				return m[a].Start < m[b].Start
			})
			return m
		}
		if !reflect.DeepEqual(key(got), key(want)) {
			t.Fatalf("AhoCorasick(%q, %q) = %v, want %v", text, pats, got, want)
		}
	}
	s := strmatch.AhoCorasickTrie([]string{"he", "she", "his", "hers"})
	words := 0
	for _, n := range s.Nodes {
		if n.Role == "word" {
			words++
		}
	}
	if s.Nodes[0].Role != "root" || words != 4 {
		t.Errorf("trie snapshot: %d words", words)
	}
}

func TestManacher(t *testing.T) {
	for s, want := range map[string]string{"babad": "bab", "cbbd": "bb", "a": "a", "forgeeksskeegfor": "geeksskeeg", "abacdfgdcaba": "aba"} {
		start, n, _ := strmatch.Manacher(s)
		if got := s[start : start+n]; got != want {
			t.Errorf("Manacher(%q) = %q, want %q", s, got, want)
		}
	}
	r := rand.New(rand.NewSource(6))
	for i := 0; i < 300; i++ {
		s := randomText(r, 1+r.Intn(30), "ab")
		best := 0
		for a := 0; a < len(s); a++ {
			for b := a; b < len(s); b++ {
				p := s[a : b+1]
				rev := []byte(p)
				for x, y := 0, len(rev)-1; x < y; x, y = x+1, y-1 {
					rev[x], rev[y] = rev[y], rev[x]
				}
				if p == string(rev) {
					best = max(best, len(p))
				}
			}
		}
		if _, n, _ := strmatch.Manacher(s); n != best {
			t.Fatalf("Manacher(%q) length %d, want %d", s, n, best)
		}
	}
}

func TestSuffixArray(t *testing.T) {
	sa, lcp := strmatch.SuffixArray("banana")
	if !reflect.DeepEqual(sa, []int{5, 3, 1, 0, 4, 2}) || !reflect.DeepEqual(lcp, []int{0, 1, 3, 0, 0, 2}) {
		t.Errorf("banana: sa %v lcp %v", sa, lcp)
	}
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 200; i++ {
		s := randomText(r, r.Intn(40), "abc")
		sa, _ := strmatch.SuffixArray(s)
		for k := 1; k < len(sa); k++ {
			if strings.Compare(s[sa[k-1]:], s[sa[k]:]) >= 0 {
				t.Fatalf("%q: suffixes out of order at %d: %v", s, k, sa)
			}
		}
	}
}
