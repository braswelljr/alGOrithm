package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/dynamic"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/greedy"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// greedyEntries is every entry of the greedy category.
func greedyEntries() []Entry {
	return []Entry{
		activitySelectionEntry(),
		fractionalKnapsackEntry(),
		greedyCoinChangeEntry(),
		huffmanCodingEntry(),
		intervalPartitioningEntry(),
		jobSequencingEntry(),
	}
}

func activitySelectionEntry() Entry {
	act := entry("activity-selection", "Activity Selection", "algorithms/greedy/activity.go", "ActivitySelection")
	act.Summary, act.Time, act.Space, act.Input, act.Traced = "The most activities that do not overlap: always take the one that ends first.", "O(n log n)", "O(n)", Text, true
	act.Example = &Input{Text: intervalsExample}
	act.Run = func(in Input, r trace.Recorder) (Output, error) {
		acts, err := greedy.ParseIntervals(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		chosen := greedy.TraceActivitySelection(acts, r)
		taken := map[int]bool{}
		var names []string
		for _, i := range chosen {
			taken[i] = true
			names = append(names, acts[i].Name)
		}
		var spans []Span
		for i, a := range acts {
			state := "idle"
			if taken[i] {
				state = "running"
			}
			spans = append(spans, Span{Lane: a.Name, Label: a.Name, Start: a.Start, End: a.End, State: state})
		}
		return Output{Timeline: spans, Number: ptr(len(chosen)), Lines: []string{"chosen: " + strings.Join(names, ", ")}}, nil
	}
	act.Call = func(in Input) (string, string) {
		return parseCall("acts", "greedy.ParseIntervals", in.Text), "greedy.ActivitySelection(acts)"
	}
	return act
}

func intervalPartitioningEntry() Entry {
	part := entry("interval-partitioning", "Interval Partitioning", "algorithms/greedy/partition.go", "IntervalPartitioning")
	part.Summary, part.Time, part.Space, part.Input, part.Traced = "Give every interval a room so overlapping ones never share, with the fewest rooms.", "O(n log n)", "O(n)", Text, true
	part.Example = &Input{Text: intervalsExample}
	part.Run = func(in Input, r trace.Recorder) (Output, error) {
		iv, err := greedy.ParseIntervals(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		room, n := greedy.TraceIntervalPartitioning(iv, r)
		var spans []Span
		for i, x := range iv {
			spans = append(spans, Span{Lane: "room " + strconv.Itoa(room[i]+1), Label: x.Name, Start: x.Start, End: x.End, State: "running"})
		}
		sort.SliceStable(spans, func(a, b int) bool { return spans[a].Lane < spans[b].Lane })
		return Output{Timeline: spans, Number: &n, Summary: map[string]float64{"rooms": float64(n)}}, nil
	}
	part.Call = func(in Input) (string, string) {
		return parseCall("iv", "greedy.ParseIntervals", in.Text), "greedy.IntervalPartitioning(iv)"
	}
	return part
}

func huffmanCodingEntry() Entry {
	huff := entry("huffman-coding", "Huffman Coding", "algorithms/greedy/huffman.go", "HuffmanCoding")
	huff.Summary, huff.Time, huff.Space, huff.Input, huff.Traced = "The shortest prefix code for a text: merge the two rarest symbols, again and again (Huffman, 1952).", "O(n + k log k)", "O(n + k)", Text, true
	huff.Example = &Input{Text: "abracadabra alakazam"}
	huff.Limits = fmt.Sprintf("text up to %d characters", greedy.MaxHuffmanText)
	huff.Run = func(in Input, r trace.Recorder) (Output, error) {
		h, err := greedy.TraceHuffmanCoding(in.Text, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Text: h.Encoded, Number: &h.Bits, Summary: map[string]float64{
			"bits": float64(h.Bits), "bits uncoded (8 per byte)": float64(h.OriginalBits),
			"compression ratio": float64(h.Bits) / float64(h.OriginalBits),
		}}
		chars := make([]int, 0, len(h.Codes))
		for c := range h.Codes {
			chars = append(chars, int(c))
		}
		sort.Ints(chars)
		for _, c := range chars {
			out.Lines = append(out.Lines, fmt.Sprintf("%q = %s", greedy.CharLabel(byte(c)), h.Codes[byte(c)]))
		}
		for i, f := range h.Merges {
			op, res := Op{Name: "merge", Value: strconv.Itoa(i)}, plural(len(rootsOf(f)), "tree")+" left"
			if i == 0 {
				op, res = Op{Name: "start"}, fmt.Sprintf("one tree per symbol (%d)", len(rootsOf(f)))
			}
			out.Steps = append(out.Steps, Step{Op: op, Result: res, Snapshot: f})
		}
		return out, nil
	}
	huff.Call = func(in Input) (string, string) { return "", fmt.Sprintf("greedy.HuffmanCoding(%q)", in.Text) }
	return huff
}

func fractionalKnapsackEntry() Entry {
	knap := entry("fractional-knapsack", "Fractional Knapsack", "algorithms/greedy/knapsack.go", "FractionalKnapsack")
	knap.Summary, knap.Time, knap.Space, knap.Input, knap.Traced = "The most value in a knapsack when items can be split: best value per weight first.", "O(n log n)", "O(n)", Text, true
	knap.Options = []Option{{"capacity", "capacity", 50, 0, 1_000_000}}
	knap.Example = &Input{Text: itemsExample, Options: map[string]int{"capacity": 50}}
	knap.Run = func(in Input, r trace.Recorder) (Output, error) {
		items, err := greedy.ParseItems(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		v, taken := greedy.TraceFractionalKnapsack(items, in.opt("capacity", 50), r)
		out := Output{Summary: map[string]float64{"value carried": v}}
		for i, f := range taken {
			if f > 0 {
				out.Lines = append(out.Lines, fmt.Sprintf("%s: %.0f%% (weight %.4g, value %.4g)", items[i].Name, f*100, f*float64(items[i].Weight), f*float64(items[i].Value)))
			}
		}
		return out, nil
	}
	knap.Call = func(in Input) (string, string) {
		return parseCall("items", "greedy.ParseItems", in.Text), fmt.Sprintf("greedy.FractionalKnapsack(items, %d)", in.opt("capacity", 50))
	}
	return knap
}

func jobSequencingEntry() Entry {
	jobs := entry("job-sequencing", "Job Sequencing with Deadlines", "algorithms/greedy/jobs.go", "JobSequencing")
	jobs.Summary, jobs.Time, jobs.Space, jobs.Input, jobs.Traced = "The most profit from unit-time jobs with deadlines: richest first, each in the latest free slot.", "O(n log n + n * d)", "O(d)", Text, true
	jobs.Example = &Input{Text: jobsExample}
	jobs.Run = func(in Input, r trace.Recorder) (Output, error) {
		js, err := greedy.ParseJobs(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		slots, profit := greedy.TraceJobSequencing(js, r)
		var spans []Span
		var order []string
		for s, j := range slots {
			if j < 0 {
				spans = append(spans, Span{Lane: "schedule", Label: "idle", Start: s, End: s + 1, State: "idle"})
				continue
			}
			spans = append(spans, Span{Lane: "schedule", Label: js[j].Name, Start: s, End: s + 1, State: "running"})
			order = append(order, js[j].Name)
		}
		return Output{Timeline: spans, Number: &profit, Lines: []string{"order: " + strings.Join(order, ", ")}}, nil
	}
	jobs.Call = func(in Input) (string, string) {
		return parseCall("jobs", "greedy.ParseJobs", in.Text), "greedy.JobSequencing(jobs)"
	}
	return jobs
}

func greedyCoinChangeEntry() Entry {
	coin := entry("greedy-coin-change", "Greedy Coin Change", "algorithms/greedy/coins.go", "CoinChange")
	coin.Summary, coin.Time, coin.Space, coin.Input, coin.Traced, coin.Params = "Pay with the largest coin that fits, again and again; optimal for some coin systems and not for others.", "O(k log k + amount / smallest)", "O(k)", Amount, true, []string{"amount"}
	coin.Example = &Input{Values: []int{1, 3, 4}, Target: 6}
	coin.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "pay with (the coins)", 100); err != nil {
			return Output{}, err
		}
		if in.Target < 0 || in.Target > dynamic.MaxAmount {
			return Output{}, inputError(fmt.Sprintf("the amount is 0 to %d", dynamic.MaxAmount))
		}
		used, ok := greedy.TraceCoinChange(in.Values, in.Target, r)
		best, bestCoins, _, err := dynamic.CoinChange(in.Values, in.Target)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Values: used, Bool: &ok, Summary: map[string]float64{"greedy coins": float64(len(used)), "fewest possible": float64(best)}}
		switch {
		case !ok && best < 0:
			out.Lines = []string{fmt.Sprintf("%d cannot be paid with these coins", in.Target)}
		case !ok:
			out.Lines = []string{fmt.Sprintf("greedy gets stuck, but %v pays it", bestCoins)}
		case len(used) == best:
			out.Lines = []string{fmt.Sprintf("greedy %v is optimal here", used)}
		default:
			out.Lines = []string{fmt.Sprintf("greedy uses %d coins %v; the optimum %v uses %d", len(used), used, bestCoins, best)}
		}
		return out, nil
	}
	coin.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("greedy.CoinChange(%s, %d)", ints(in.Values), in.Target)
	}
	return coin
}

// plural writes "1 tree" or "3 trees".
func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

// rootsOf returns the root nodes of a forest snapshot.
func rootsOf(s trace.Snapshot) []trace.Node {
	var out []trace.Node
	for _, n := range s.Nodes {
		if n.Parent < 0 {
			out = append(out, n)
		}
	}
	return out
}
