package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// strmatchEntries is every entry of the strmatch category.
func strmatchEntries() []Entry {
	return []Entry{
		ahoCorasickEntry(),
		anagramEntry(),
		boyerMooreHorspoolEntry(),
		knuthMorrisPrattEntry(),
		manacherEntry(),
		naiveSearchEntry(),
		rabinKarpEntry(),
		suffixArrayEntry(),
		zAlgorithmEntry(),
	}
}

var strmatchEx = &Input{Text: "abracadabra, abracadabra", Other: "abra"}

func strmatchMk(id, name, file, fn, summary, time, space string, traced func(string, string, trace.Recorder) []int) Entry {
	e := entry(id, name, "algorithms/strmatch/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, time, space, Pattern, true, []string{"pattern"}, strmatchEx
	e.Limits = "text up to 10000 characters"
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if in.Text == "" || in.Other == "" {
			return Output{}, inputError("give a text and a pattern")
		}
		if len(in.Text) > 10_000 {
			return Output{}, inputError("the text is longer than 10000 characters")
		}
		m := traced(in.Text, in.Other, r)
		if m == nil {
			m = []int{}
		}
		return Output{Text: in.Text, Matches: m, Number: ptr(len(m))}, nil
	}
	return e
}

func anagramEntry() Entry {
	ana := entry("anagram", "Anagram", "algorithms/strmatch/anagram.go", "Anagram")
	ana.Summary, ana.Time, ana.Space, ana.Input = "Whether two phrases use exactly the same letters, ignoring case and spaces.", "O(n)", "O(1)", TwoTexts
	ana.Example = &Input{Text: "Dormitory", Other: "dirty room"}
	ana.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" || in.Other == "" {
			return Output{}, inputError("give two words")
		}
		return Output{Bool: ptr(strmatch.Anagram(in.Text, in.Other))}, nil
	}
	return ana
}

func naiveSearchEntry() Entry {
	return strmatchMk("naive-search", "Naive String Search", "naive.go", "Naive", "Try every starting position and compare letter by letter.", "O(n * m)", "O(1)", strmatch.TraceNaive)
}

func knuthMorrisPrattEntry() Entry {
	return strmatchMk("knuth-morris-pratt", "Knuth-Morris-Pratt", "kmp.go", "KMP", "After a mismatch, keep the longest part of the match that still fits, so no text is read twice (1977).", "O(n + m)", "O(m)", strmatch.TraceKMP)
}

func rabinKarpEntry() Entry {
	return strmatchMk("rabin-karp", "Rabin-Karp", "rabinkarp.go", "RabinKarp", "Compare a rolling hash of each window with the pattern's hash, and letters only when they agree (1987).", "O(n + m) average", "O(1)", strmatch.TraceRabinKarp)
}

var patternIndexEx = &Input{Text: "here is a simple example of a pattern in a sample text", Other: "ample"}

func patternIndexMatch(id, name, file, fn, summary, time, space string, traced func(string, string, trace.Recorder) []int) Entry {
	e := entry(id, name, "algorithms/strmatch/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, time, space, Pattern, true, []string{"pattern"}, patternIndexEx
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if in.Text == "" || in.Other == "" || len(in.Text) > 10_000 {
			return Output{}, inputError("give a text of up to 10000 characters and a pattern")
		}
		m := traced(in.Text, in.Other, r)
		if m == nil {
			m = []int{}
		}
		return Output{Text: in.Text, Matches: m, Number: ptr(len(m))}, nil
	}
	e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.%s(%q, %q)", fn, in.Text, in.Other) }
	return e
}

func zAlgorithmEntry() Entry {
	z := patternIndexMatch("z-algorithm", "Z-Algorithm", "zalgorithm.go", "ZSearch", "The Z array of pattern + separator + text: wherever a value reaches the pattern's length, it matches; the Z-box reuses earlier work.", "O(n + m)", "O(n + m)", strmatch.TraceZSearch)
	z.Func = "ZSearch"
	return z
}

func ahoCorasickEntry() Entry {
	aho := entry("aho-corasick", "Aho-Corasick", "algorithms/strmatch/ahocorasick.go", "AhoCorasick")
	aho.Summary, aho.Time, aho.Space, aho.Input, aho.Traced, aho.Params = "Find many patterns in one pass: a trie of the patterns with failure links, so a mismatch falls back without re-reading the text (1975).", "O(n + patterns + matches)", "O(total pattern length)", Pattern, true, []string{"pattern"}
	aho.Example = &Input{Text: "ushers say she sells seashells", Other: "he she his hers sea"}
	aho.Run = func(in Input, r trace.Recorder) (Output, error) {
		pats := splitWords(in.Other)
		if in.Text == "" || len(pats) == 0 || len(in.Text) > 10_000 || len(pats) > 100 {
			return Output{}, inputError("give a text of up to 10000 characters and 1 to 100 patterns")
		}
		ms := strmatch.TraceAhoCorasick(in.Text, pats, r)
		out := Output{Text: in.Text, Matches: []int{}, Number: ptr(len(ms))}
		for _, m := range ms {
			out.Matches = append(out.Matches, m.Start)
			out.Lines = append(out.Lines, fmt.Sprintf("%q at %d", pats[m.Pattern], m.Start))
		}
		out.Steps = []Step{{Op: Op{Name: "trie", Value: strings.Join(pats, " ")}, Result: "the patterns' trie, with failure links", Snapshot: strmatch.AhoCorasickTrie(pats)}}
		return out, nil
	}
	aho.Call = func(in Input) (string, string) {
		q := make([]string, 0)
		for _, p := range splitWords(in.Other) {
			q = append(q, strconv.Quote(p))
		}
		return "", fmt.Sprintf("strmatch.AhoCorasick(%q, []string{%s})", in.Text, strings.Join(q, ", "))
	}
	return aho
}

func manacherEntry() Entry {
	man := entry("manacher", "Manacher's Algorithm", "algorithms/strmatch/manacher.go", "Manacher")
	man.Summary, man.Time, man.Space, man.Input, man.Traced = "The longest palindromic substring in linear time: grow each center's radius from its mirror inside the rightmost palindrome so far (1975).", "O(n)", "O(n)", Text, true
	man.Example = &Input{Text: "forgeeksskeegfor"}
	man.Run = func(in Input, r trace.Recorder) (Output, error) {
		if in.Text == "" || len(in.Text) > 10_000 {
			return Output{}, inputError("give a text of 1 to 10000 characters")
		}
		start, n, radius := strmatch.TraceManacher(in.Text, r)
		labels := make([]string, len(radius))
		for i := range labels {
			labels[i] = "#"
			if i%2 == 1 {
				labels[i] = string(in.Text[i/2])
			}
		}
		return Output{Text: in.Text, Matches: []int{start}, Number: &n, Lines: []string{in.Text[start : start+n]},
			Table: &TableOutput{Rows: 1, Cols: len(radius), Cells: radius, ColLabels: labels}}, nil
	}
	man.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.Manacher(%q)", in.Text) }
	return man
}

func suffixArrayEntry() Entry {
	sa := entry("suffix-array", "Suffix Array", "algorithms/strmatch/suffixarray.go", "SuffixArray")
	sa.Summary, sa.Time, sa.Space, sa.Input = "Every suffix in sorted order by prefix doubling, with the LCP array of neighbors' common prefixes (Kasai).", "O(n log^2 n)", "O(n)", Text
	sa.Example = &Input{Text: "banana"}
	sa.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" || len(in.Text) > 5000 {
			return Output{}, inputError("give a text of 1 to 5000 characters")
		}
		arr, lcp := strmatch.SuffixArray(in.Text)
		t := &TableOutput{Rows: len(arr), Cols: 2, ColLabels: []string{"start", "lcp"}}
		out := Output{Table: t, Values: arr}
		for i, s := range arr {
			t.Cells = append(t.Cells, s, lcp[i])
			t.RowLabels = append(t.RowLabels, strconv.Itoa(i))
			suffix := in.Text[s:]
			if len(suffix) > 60 {
				suffix = suffix[:60] + "..."
			}
			out.Lines = append(out.Lines, fmt.Sprintf("%3d  %s", s, suffix))
		}
		return out, nil
	}
	sa.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.SuffixArray(%q)", in.Text) }
	return sa
}

func boyerMooreHorspoolEntry() Entry {
	return patternIndexMatch("boyer-moore-horspool", "Boyer-Moore-Horspool", "horspool.go", "BoyerMooreHorspool", "Compare each window from its right end and skip ahead by the bad-character shift, often many letters at once (1980).", "O(n / m) best, O(n * m) worst", "O(256)", strmatch.TraceBoyerMooreHorspool)
}
