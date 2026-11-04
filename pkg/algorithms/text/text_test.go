package text_test

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/text"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func randomText(rng *rand.Rand, n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(b)
}

func TestLongestRepeated(t *testing.T) {
	sub, at := text.LongestRepeated("banana")
	if sub != "ana" || !reflect.DeepEqual(at, []int{1, 3}) {
		t.Errorf("banana: %q %v", sub, at)
	}
	if sub, at := text.LongestRepeated("abc"); sub != "" || at != nil {
		t.Errorf("abc: %q %v", sub, at)
	}
	rng := rand.New(rand.NewSource(1))
	for n := 0; n < 200; n++ {
		s := randomText(rng, rng.Intn(30), "ab")
		want := 0
		for i := range s {
			for j := i + 1; j < len(s); j++ {
				l := 0
				for j+l < len(s) && s[i+l] == s[j+l] {
					l++
				}
				want = max(want, l)
			}
		}
		if sub, _ := text.LongestRepeated(s); len(sub) != want || (want > 0 && strings.Count(s, sub) < 1) {
			t.Fatalf("%q: got %q, want length %d", s, sub, want)
		}
	}
	var log trace.Log
	text.TraceLongestRepeated("banana", &log)
	if log.Count(trace.Mark) != 2 {
		t.Errorf("want 2 found marks, got %d", log.Count(trace.Mark))
	}
}

func TestLongestCommon(t *testing.T) {
	a, b, n := text.LongestCommon("xabcdey", "zzbcdq")
	if n != 3 || "xabcdey"[a:a+n] != "bcd" || "zzbcdq"[b:b+n] != "bcd" {
		t.Errorf("got %d %d %d", a, b, n)
	}
	if _, _, n := text.LongestCommon("abc", "xyz"); n != 0 {
		t.Errorf("no common: %d", n)
	}
}

func TestDistinctSubstrings(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for n := 0; n < 200; n++ {
		s := randomText(rng, rng.Intn(25), "abc")
		set := map[string]bool{}
		for i := range s {
			for j := i + 1; j <= len(s); j++ {
				set[s[i:j]] = true
			}
		}
		got, states := text.DistinctSubstrings(s)
		if got != len(set) || states > max(2*len(s), 1) {
			t.Fatalf("%q: %d (states %d), want %d", s, got, states, len(set))
		}
	}
}

func TestMinimalRotation(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for n := 0; n < 300; n++ {
		s := randomText(rng, 1+rng.Intn(20), "abc")
		best := s
		for k := range s {
			best = min(best, s[k:]+s[:k])
		}
		k := text.MinimalRotation(s)
		if s[k:]+s[:k] != best {
			t.Fatalf("%q: rotation %d is %q, want %q", s, k, s[k:]+s[:k], best)
		}
	}
	if text.MinimalRotation("") != 0 {
		t.Error("empty")
	}
}

func isLyndon(w string) bool {
	for k := 1; k < len(w); k++ {
		if w[k:]+w[:k] <= w {
			return false
		}
	}
	return w != ""
}

func TestLyndonFactors(t *testing.T) {
	if got := text.LyndonFactors("banana"); !reflect.DeepEqual(got, []string{"b", "an", "an", "a"}) {
		t.Errorf("banana: %v", got)
	}
	rng := rand.New(rand.NewSource(4))
	for n := 0; n < 300; n++ {
		s := randomText(rng, rng.Intn(25), "ab")
		fs := text.LyndonFactors(s)
		if strings.Join(fs, "") != s {
			t.Fatalf("%q: factors %v do not rebuild it", s, fs)
		}
		for i, f := range fs {
			if !isLyndon(f) || (i > 0 && fs[i-1] < f) {
				t.Fatalf("%q: bad factorization %v", s, fs)
			}
		}
	}
}

func TestSmallestPeriod(t *testing.T) {
	for _, c := range []struct {
		s     string
		p     int
		whole bool
	}{{"abab", 2, true}, {"abcab", 3, false}, {"aaaa", 1, true}, {"abc", 3, true}, {"", 0, false}} {
		if p, w := text.SmallestPeriod(c.s); p != c.p || w != c.whole {
			t.Errorf("%q: %d %v, want %d %v", c.s, p, w, c.p, c.whole)
		}
	}
}

// sellers is the textbook dynamic program for approximate matching: column j
// is the best edit distance of the pattern to any text substring ending at j.
func sellers(t, p string, k int) []int {
	col := make([]int, len(p)+1)
	for i := range col {
		col[i] = i
	}
	ends := []int{}
	for j := 0; j < len(t); j++ {
		prevDiag := col[0]
		col[0] = 0
		for i := 1; i <= len(p); i++ {
			cost := 1
			if p[i-1] == t[j] {
				cost = 0
			}
			next := min(col[i]+1, col[i-1]+1, prevDiag+cost)
			prevDiag, col[i] = col[i], next
		}
		if col[len(p)] <= k {
			ends = append(ends, j)
		}
	}
	return ends
}

func TestFuzzySearch(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for n := 0; n < 500; n++ {
		s := randomText(rng, rng.Intn(40), "abc")
		p := randomText(rng, 1+rng.Intn(6), "abc")
		k := rng.Intn(len(p) + 1)
		got, err := text.FuzzySearch(s, p, k)
		if err != nil || !reflect.DeepEqual(got, sellers(s, p, k)) {
			t.Fatalf("%q in %q with %d errors: %v (%v), want %v", p, s, k, got, err, sellers(s, p, k))
		}
	}
	if _, err := text.FuzzySearch("x", "", 0); err == nil {
		t.Error("empty pattern accepted")
	}
	if _, err := text.FuzzySearch("x", "ab", 3); err == nil {
		t.Error("too many errors accepted")
	}
}

func TestWordFrequencyAndTFIDF(t *testing.T) {
	got := text.WordFrequency("The cat and the hat. THE end")
	if got[0] != (text.WordCount{Word: "the", Count: 3}) || len(got) != 5 {
		t.Errorf("%v", got)
	}
	tf := text.TFIDF([]string{"the cat sat", "the dog sat", "the dog barked"})
	if tf[0][0].Word != "cat" {
		t.Errorf("cat should lead the first document: %v", tf[0])
	}
	for _, d := range tf {
		for _, term := range d {
			if term.Word == "the" && term.Score != 0 {
				t.Errorf("a word in every document scores 0, not %v", term.Score)
			}
		}
	}
}

func TestNGramsAndJaccard(t *testing.T) {
	if got := text.NGrams("abcd", 2); !reflect.DeepEqual(got, []string{"ab", "bc", "cd"}) {
		t.Errorf("%v", got)
	}
	if text.NGrams("a", 2) != nil {
		t.Error("too short")
	}
	j, shared := text.Jaccard("night", "nacht", 2)
	if j != 1.0/7 || !reflect.DeepEqual(shared, []string{"ht"}) {
		t.Errorf("night/nacht: %v %v", j, shared)
	}
	if j, _ := text.Jaccard("abc", "abc", 2); j != 1 {
		t.Errorf("same: %v", j)
	}
}

const pangram = "The quick brown fox jumps over the lazy dog and keeps on running far away"

func TestWrap(t *testing.T) {
	for _, width := range []int{8, 12, 20, 30} {
		g := text.WrapGreedy(pangram, width)
		b, cost := text.WrapBalanced(pangram, width)
		for _, lines := range [][]string{g, b} {
			if strings.Join(lines, " ") != strings.Join(strings.Fields(pangram), " ") {
				t.Fatalf("width %d: words lost: %v", width, lines)
			}
			for _, l := range lines {
				if len(l) > width {
					t.Fatalf("width %d: %q too long", width, l)
				}
			}
		}
		greedy := 0
		for _, l := range g[:len(g)-1] {
			greedy += (width - len(l)) * (width - len(l))
		}
		if cost > greedy {
			t.Errorf("width %d: balanced cost %d beats greedy %d", width, cost, greedy)
		}
	}
	if got, _ := text.WrapBalanced("aaa bb cc ddddd", 6); !reflect.DeepEqual(got, []string{"aaa", "bb cc", "ddddd"}) {
		t.Errorf("the classic example: %v", got)
	}
}

func TestJustify(t *testing.T) {
	lines := text.Justify("This is an example of text justification.", 16)
	want := []string{"This    is    an", "example  of text", "justification."}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("%q", lines)
	}
}

func lcsLength(a, b []string) int {
	t := make([][]int, len(a)+1)
	for i := range t {
		t[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				t[i][j] = t[i-1][j-1] + 1
			} else {
				t[i][j] = max(t[i-1][j], t[i][j-1])
			}
		}
	}
	return t[len(a)][len(b)]
}

func TestDiff(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	split := func(s string) []string {
		if s == "" {
			return nil
		}
		return strings.Split(s, "")
	}
	for n := 0; n < 300; n++ {
		a, b := split(randomText(rng, rng.Intn(15), "abc")), split(randomText(rng, rng.Intn(15), "abc"))
		var gotA, gotB []string
		edits := 0
		for _, e := range text.Diff(a, b) {
			if e.Op != '+' {
				gotA = append(gotA, e.Line)
			}
			if e.Op != '-' {
				gotB = append(gotB, e.Line)
			}
			if e.Op != '=' {
				edits++
			}
		}
		if !slices.Equal(gotA, a) || !slices.Equal(gotB, b) || edits != len(a)+len(b)-2*lcsLength(a, b) {
			t.Fatalf("%v -> %v: %v", a, b, text.Diff(a, b))
		}
	}
}

func TestSoundex(t *testing.T) {
	for word, want := range map[string]string{
		"Robert": "R163", "Rupert": "R163", "Rubin": "R150", "Ashcraft": "A261",
		"Tymczak": "T522", "Pfister": "P236", "Honeyman": "H555", "Lee": "L000", "": "", "42": "",
	} {
		if got := text.Soundex(word); got != want {
			t.Errorf("%q: %q, want %q", word, got, want)
		}
	}
}

func TestAutocompleteAndPrefix(t *testing.T) {
	words := strings.Fields("car cart care car cat dog cart car")
	if got := text.Autocomplete(words, "car", 3); !reflect.DeepEqual(got, []string{"car", "cart", "care"}) {
		t.Errorf("%v", got)
	}
	if got := text.Autocomplete(words, "x", 3); got != nil {
		t.Errorf("%v", got)
	}
	if got := text.LongestCommonPrefix([]string{"flower", "flow", "flight"}); got != "fl" {
		t.Errorf("%q", got)
	}
	if text.LongestCommonPrefix(nil) != "" {
		t.Error("nil")
	}
}

func TestSuggest(t *testing.T) {
	dict := strings.Fields("apple apply ample maple apple angle")
	got := text.Suggest(dict, "appel", 2)
	want := []text.Suggestion{{Word: "apple", Distance: 2}, {Word: "apply", Distance: 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v, want %v", got, want)
	}
	if got := text.Suggest(dict, "appel", 3); !reflect.DeepEqual(got[2:], []text.Suggestion{{Word: "ample", Distance: 3}, {Word: "maple", Distance: 3}}) {
		t.Errorf("within 3: %v", got)
	}
	if got[0] != (text.Suggestion{Word: "apple", Distance: 2}) {
		t.Errorf("first: %v", got[0])
	}
}

func TestTracesMatchUntraced(t *testing.T) {
	var log trace.Log
	if text.TraceMinimalRotation("bca", &log) != text.MinimalRotation("bca") || log.Total() == 0 {
		t.Error("rotation")
	}
	if p, _ := text.TraceSmallestPeriod("abab", &log); p != 2 {
		t.Error("period")
	}
	if got := text.TraceLyndonFactors("banana", &log); len(got) != 4 {
		t.Error("lyndon")
	}
	if _, _, n := text.TraceLongestCommon("abc", "bc", &log); n != 2 {
		t.Error("common")
	}
	if got, _ := text.TraceFuzzySearch("abcd", "bd", 1, &log); len(got) == 0 {
		t.Error("fuzzy")
	}
}
