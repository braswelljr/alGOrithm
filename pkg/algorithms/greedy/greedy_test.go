package greedy_test

import (
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/greedy"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func randomIntervals(r *rand.Rand, n int) []greedy.Interval {
	out := make([]greedy.Interval, n)
	for i := range out {
		s := r.Intn(20)
		out[i] = greedy.Interval{Name: string(rune('a' + i)), Start: s, End: s + 1 + r.Intn(6)}
	}
	return out
}

func overlap(a, b greedy.Interval) bool { return a.Start < b.End && b.Start < a.End }

// bestBrute is the size of the largest non-overlapping subset, by trying all.
func bestBrute(acts []greedy.Interval) int {
	best := 0
	for mask := 0; mask < 1<<len(acts); mask++ {
		var set []greedy.Interval
		ok := true
		for i := range acts {
			if mask&(1<<i) == 0 {
				continue
			}
			for _, s := range set {
				if overlap(s, acts[i]) {
					ok = false
				}
			}
			set = append(set, acts[i])
		}
		if ok && len(set) > best {
			best = len(set)
		}
	}
	return best
}

func TestActivitySelectionIsOptimal(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 200; trial++ {
		acts := randomIntervals(r, 1+r.Intn(10))
		var log trace.Log
		chosen := greedy.TraceActivitySelection(acts, &log)
		for i := 1; i < len(chosen); i++ {
			if overlap(acts[chosen[i-1]], acts[chosen[i]]) {
				t.Fatalf("chosen activities overlap: %v", chosen)
			}
		}
		if want := bestBrute(acts); len(chosen) != want {
			t.Fatalf("chose %d of %v, the best is %d", len(chosen), acts, want)
		}
		if log.Count(trace.Mark) != len(chosen) {
			t.Fatal("one chosen mark per activity taken")
		}
	}
}

func TestIntervalPartitioningUsesTheDepth(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for trial := 0; trial < 200; trial++ {
		iv := randomIntervals(r, 1+r.Intn(12))
		room, n := greedy.IntervalPartitioning(iv)
		depth := 0
		for t0 := 0; t0 < 30; t0++ {
			d := 0
			for _, x := range iv {
				if x.Start <= t0 && t0 < x.End {
					d++
				}
			}
			depth = max(depth, d)
		}
		if n != depth {
			t.Fatalf("%d rooms, the deepest overlap is %d", n, depth)
		}
		for i := range iv {
			for j := i + 1; j < len(iv); j++ {
				if room[i] == room[j] && overlap(iv[i], iv[j]) {
					t.Fatalf("intervals %d and %d share room %d and overlap", i, j, room[i])
				}
			}
		}
	}
}

// optimalCost is the least total weighted code length, by trying every merge
// order (every full binary tree is built by some merge order).
func optimalCost(w []int) int {
	if len(w) < 2 {
		return 0
	}
	best := math.MaxInt
	for i := 0; i < len(w); i++ {
		for j := i + 1; j < len(w); j++ {
			rest := []int{w[i] + w[j]}
			for k := range w {
				if k != i && k != j {
					rest = append(rest, w[k])
				}
			}
			best = min(best, w[i]+w[j]+optimalCost(rest))
		}
	}
	return best
}

func TestHuffman(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for trial := 0; trial < 100; trial++ {
		n := 1 + r.Intn(40)
		b := make([]byte, n)
		for i := range b {
			b[i] = "abcde"[r.Intn(1+r.Intn(5))]
		}
		text := string(b)
		h, err := greedy.HuffmanCoding(text)
		if err != nil {
			t.Fatal(err)
		}
		freq := map[byte]int{}
		for i := 0; i < len(text); i++ {
			freq[text[i]]++
		}
		var w []int
		for _, f := range freq {
			w = append(w, f)
		}
		want := optimalCost(w)
		if len(freq) == 1 {
			want = len(text)
		}
		if h.Bits != want {
			t.Fatalf("%q: %d bits, the optimum is %d", text, h.Bits, want)
		}
		// Prefix-free, and the encoding decodes back.
		for c, code := range h.Codes {
			for d, other := range h.Codes {
				if c != d && strings.HasPrefix(other, code) {
					t.Fatalf("%q is a prefix of %q", code, other)
				}
			}
		}
		var out []byte
		for rest := h.Encoded; rest != ""; {
			found := false
			for c, code := range h.Codes {
				if strings.HasPrefix(rest, code) {
					out, rest, found = append(out, c), rest[len(code):], true
					break
				}
			}
			if !found {
				t.Fatal("the encoding does not decode")
			}
		}
		if string(out) != text {
			t.Fatalf("decoded %q, want %q", out, text)
		}
		if len(h.Merges) != len(freq) || len(h.Tree.Nodes) != 2*len(freq)-1 {
			t.Fatalf("%d merges and %d tree nodes for %d symbols", len(h.Merges), len(h.Tree.Nodes), len(freq))
		}
	}
	if _, err := greedy.HuffmanCoding(""); !errors.Is(err, greedy.ErrInput) {
		t.Error("empty text accepted")
	}
	if greedy.CharLabel(' ') != " " && greedy.CharLabel('\n') != `\n` {
		t.Error("CharLabel wrong")
	}
}

func TestFractionalKnapsack(t *testing.T) {
	items, err := greedy.ParseItems("name weight value\na 10 60\nb 20 100\nc 30 120")
	if err != nil {
		t.Fatal(err)
	}
	v, taken := greedy.FractionalKnapsack(items, 50)
	if v != 240 || !reflect.DeepEqual(taken, []float64{1, 1, 20.0 / 30}) {
		t.Errorf("value %v, taken %v; want 240 (the textbook answer)", v, taken)
	}
	if v, _ := greedy.FractionalKnapsack(items, 0); v != 0 {
		t.Errorf("capacity 0 carries %v", v)
	}
	if _, err := greedy.ParseItems("a 0 5"); !errors.Is(err, greedy.ErrInput) {
		t.Error("a weightless item accepted")
	}
}

func TestJobSequencing(t *testing.T) {
	jobs, err := greedy.ParseJobs("a 2 100\nb 1 19\nc 2 27\nd 1 25\ne 3 15")
	if err != nil {
		t.Fatal(err)
	}
	slots, profit := greedy.JobSequencing(jobs)
	if profit != 142 || !reflect.DeepEqual(slots, []int{2, 0, 4}) {
		t.Errorf("slots %v profit %d; want c a e for 142", slots, profit)
	}
	// Against every subset that can meet its deadlines.
	r := rand.New(rand.NewSource(4))
	for trial := 0; trial < 200; trial++ {
		js := make([]greedy.Job, 1+r.Intn(8))
		for i := range js {
			js[i] = greedy.Job{Name: "j", Deadline: 1 + r.Intn(4), Profit: r.Intn(50)}
		}
		_, got := greedy.JobSequencing(js)
		best := 0
		for mask := 0; mask < 1<<len(js); mask++ {
			var ds []int
			p := 0
			for i := range js {
				if mask&(1<<i) != 0 {
					ds, p = append(ds, js[i].Deadline), p+js[i].Profit
				}
			}
			// feasible iff for every t, at most t jobs have deadline <= t
			ok := true
			for t0 := 1; t0 <= 4; t0++ {
				c := 0
				for _, d := range ds {
					if d <= t0 {
						c++
					}
				}
				ok = ok && c <= t0
			}
			if ok {
				best = max(best, p)
			}
		}
		if got != best {
			t.Fatalf("%v: greedy %d, best %d", js, got, best)
		}
	}
}

// fewestCoins is the optimal coin count by dynamic programming, -1 if none.
func fewestCoins(coins []int, amount int) int {
	dp := make([]int, amount+1)
	for a := 1; a <= amount; a++ {
		dp[a] = -1
		for _, c := range coins {
			if c <= a && dp[a-c] >= 0 && (dp[a] < 0 || dp[a-c]+1 < dp[a]) {
				dp[a] = dp[a-c] + 1
			}
		}
	}
	return dp[amount]
}

func TestCoinChange(t *testing.T) {
	us := []int{1, 5, 10, 25}
	for a := 0; a <= 200; a++ {
		used, ok := greedy.CoinChange(us, a)
		if !ok || len(used) != fewestCoins(us, a) {
			t.Fatalf("US coins, %d: greedy %v is not optimal", a, used)
		}
	}
	used, ok := greedy.CoinChange([]int{1, 3, 4}, 6)
	if !ok || !reflect.DeepEqual(used, []int{4, 1, 1}) || fewestCoins([]int{1, 3, 4}, 6) != 2 {
		t.Errorf("1, 3, 4 for 6: greedy %v; the optimum is 3+3", used)
	}
	if _, ok := greedy.CoinChange([]int{5}, 3); ok {
		t.Error("3 from 5s reported paid")
	}
}

func TestParseErrors(t *testing.T) {
	for _, bad := range []string{"", "a 5 2", "a x 3\nb 1 2", "a 1"} {
		if _, err := greedy.ParseIntervals(bad); !errors.Is(err, greedy.ErrInput) {
			t.Errorf("ParseIntervals(%q) = %v", bad, err)
		}
	}
	if iv, err := greedy.ParseIntervals("name start end\na 1 2"); err != nil || len(iv) != 1 {
		t.Errorf("header line: %v %v", iv, err)
	}
}
