package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/text"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

const maxText = 10_000

// textEntry starts an entry of package text whose input is checked to hold
// 1 to limit characters in Text (and in Other, for two texts and patterns).
func textEntry(id, name, file, fn, summary, timeCost, space string, kind InputKind, ex *Input, limit int) Entry {
	e := entry(id, name, "algorithms/text/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, timeCost, space, kind, ex
	e.Limits = fmt.Sprintf("at most %d characters", limit)
	if kind == Pattern {
		e.Params = []string{"pattern"}
	}
	return e
}

func checkText(in Input, kind InputKind, limit int) error {
	if strings.TrimSpace(in.Text) == "" || len(in.Text) > limit {
		return inputError(fmt.Sprintf("give a text of 1 to %d characters", limit))
	}
	if (kind == TwoTexts || kind == Pattern) && (in.Other == "" || len(in.Other) > limit) {
		return inputError(fmt.Sprintf("give a second text of 1 to %d characters", limit))
	}
	return nil
}

// textLines splits s into lines, without a final empty one.
func textLines(s string) []string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

const pangramText = "The quick brown fox jumps over the lazy dog and keeps on running far away"

func textEntries() []Entry {
	run := func(e *Entry, limit int, f func(in Input, r trace.Recorder) (Output, error)) {
		kind := e.Input
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			if err := checkText(in, kind, limit); err != nil {
				return Output{}, err
			}
			return f(in, r)
		}
	}
	width := Option{"width", "line width", 30, 5, 200}

	lrs := textEntry("longest-repeated-substring", "Longest Repeated Substring", "repeated.go", "LongestRepeated",
		"The longest piece of a text that occurs twice: sort the suffixes, and the answer is the longest common prefix of two neighbors.",
		"O(n log^2 n)", "O(n)", Text, &Input{Text: "banana"}, maxText)
	lrs.Traced = true
	run(&lrs, maxText, func(in Input, r trace.Recorder) (Output, error) {
		sub, at := text.TraceLongestRepeated(in.Text, r)
		if at == nil {
			at = []int{}
		}
		return Output{Text: in.Text, Matches: at, Number: ptr(len(sub)), Lines: []string{strconv.Quote(sub)}}, nil
	})
	lrs.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.LongestRepeated(%q)", in.Text) }

	lcs := textEntry("longest-common-substring", "Longest Common Substring", "common.go", "LongestCommon",
		"The longest run of characters two texts share, from a table of common runs ending at each pair of positions, one row at a time.",
		"O(n * m)", "O(m)", TwoTexts, &Input{Text: "xabcdey", Other: "zzbcdq"}, 2000)
	lcs.Traced = true
	run(&lcs, 2000, func(in Input, r trace.Recorder) (Output, error) {
		a, b, n := text.TraceLongestCommon(in.Text, in.Other, r)
		return Output{Text: in.Text, Matches: []int{a}, Number: &n, Lines: []string{strconv.Quote(in.Text[a : a+n])},
			Summary: map[string]float64{"start in the first": float64(a), "start in the second": float64(b), "length": float64(n)}}, nil
	})
	lcs.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.LongestCommon(%q, %q)", in.Text, in.Other)
	}

	ds := textEntry("distinct-substrings", "Distinct Substrings (suffix automaton)", "automaton.go", "DistinctSubstrings",
		"Count every different substring with a suffix automaton, the smallest machine that recognizes all suffixes, built one character at a time.",
		"O(n)", "O(n)", Text, &Input{Text: "abab"}, maxText)
	run(&ds, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		n, states := text.DistinctSubstrings(in.Text)
		return Output{Number: &n, Summary: map[string]float64{"distinct substrings": float64(n), "automaton states": float64(states)}}, nil
	})
	ds.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.DistinctSubstrings(%q)", in.Text) }

	rot := textEntry("minimal-rotation", "Lexicographically Minimal Rotation (Booth)", "rotation.go", "MinimalRotation",
		"The rotation of a text that sorts first, in linear time, with a failure function over the text written twice (Booth, 1980).",
		"O(n)", "O(n)", Text, &Input{Text: "bbaaccaadd"}, maxText)
	rot.Traced = true
	run(&rot, maxText, func(in Input, r trace.Recorder) (Output, error) {
		k := text.TraceMinimalRotation(in.Text, r)
		return Output{Text: in.Text, Matches: []int{k}, Number: &k, Lines: []string{in.Text[k:] + in.Text[:k]}}, nil
	})
	rot.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.MinimalRotation(%q)", in.Text) }

	lyn := textEntry("lyndon-factorization", "Lyndon Factorization (Duval)", "lyndon.go", "LyndonFactors",
		"Split a text into non-increasing Lyndon words, each smaller than all its rotations, in one pass (Duval, 1983).",
		"O(n)", "O(1)", Text, &Input{Text: "banana"}, maxText)
	lyn.Traced = true
	run(&lyn, maxText, func(in Input, r trace.Recorder) (Output, error) {
		fs := text.TraceLyndonFactors(in.Text, r)
		return Output{Text: in.Text, Lines: fs, Number: ptr(len(fs))}, nil
	})
	lyn.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.LyndonFactors(%q)", in.Text) }

	per := textEntry("smallest-period", "Smallest Period", "period.go", "SmallestPeriod",
		"The shortest block that repeats to give the text, from its longest border (the prefix function of Knuth-Morris-Pratt).",
		"O(n)", "O(n)", Text, &Input{Text: "abcabcabcab"}, maxText)
	per.Traced = true
	run(&per, maxText, func(in Input, r trace.Recorder) (Output, error) {
		p, whole := text.TraceSmallestPeriod(in.Text, r)
		return Output{Text: in.Text, Matches: []int{0}, Number: &p, Bool: &whole, Lines: []string{in.Text[:p]}}, nil
	})
	per.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.SmallestPeriod(%q)", in.Text) }

	fz := textEntry("fuzzy-search", "Fuzzy Search (bitap)", "fuzzy.go", "FuzzySearch",
		"Find a pattern with up to k typos: one machine word per error count tracks which prefixes still match, updated with a shift per character (Wu and Manber, 1992).",
		"O(n * k)", "O(k)", Pattern, &Input{Text: "the color chart has colr and coler typos", Other: "color", Options: map[string]int{"errors": 1}}, maxText)
	fz.Traced = true
	fz.Options = []Option{{"errors", "edits allowed", 1, 0, 8}}
	run(&fz, maxText, func(in Input, r trace.Recorder) (Output, error) {
		ends, err := text.TraceFuzzySearch(in.Text, in.Other, in.opt("errors", 1), r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Text: in.Text, Matches: []int{}, Number: ptr(len(ends))}
		for _, j := range ends {
			start := max(0, j-len(in.Other)+1)
			out.Matches = append(out.Matches, start)
			out.Lines = append(out.Lines, fmt.Sprintf("ends at %d: %q", j, in.Text[start:j+1]))
		}
		return out, nil
	})
	fz.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.FuzzySearch(%q, %q, %d)", in.Text, in.Other, in.opt("errors", 1))
	}

	wf := textEntry("word-frequency", "Word Frequency", "frequency.go", "WordFrequency",
		"Count every word, most frequent first: the first step of search engines, spam filters and word clouds.",
		"O(n + w log w)", "O(w)", Text, &Input{Text: "the cat and the hat and the bat"}, maxText)
	run(&wf, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		counts := text.WordFrequency(in.Text)
		t := &TableOutput{Cols: 1, ColLabels: []string{"count"}}
		for i, c := range counts {
			if i == 50 {
				break
			}
			t.Rows++
			t.Cells = append(t.Cells, c.Count)
			t.RowLabels = append(t.RowLabels, c.Word)
		}
		return Output{Table: t, Number: ptr(len(counts))}, nil
	})
	wf.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.WordFrequency(%q)", in.Text) }

	tf := textEntry("tf-idf", "TF-IDF", "frequency.go", "TFIDF",
		"Score each word of each document (one per line) by how often it appears there times how rare it is elsewhere: the classic relevance measure.",
		"O(total words + t log t)", "O(distinct words)", Text, &Input{Text: "the cat sat on the mat\nthe dog sat on the log\nthe dog barked at the cat"}, maxText)
	run(&tf, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		docs := textLines(in.Text)
		out := Output{Number: ptr(len(docs))}
		for i, terms := range text.TFIDF(docs) {
			parts := []string{}
			for j, t := range terms {
				if j == 5 {
					break
				}
				parts = append(parts, fmt.Sprintf("%s %.3f", t.Word, t.Score))
			}
			out.Lines = append(out.Lines, fmt.Sprintf("document %d: %s", i+1, strings.Join(parts, ", ")))
		}
		return out, nil
	})
	tf.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.TFIDF(%s)", goStrings(textLines(in.Text)))
	}

	gramSize := Option{"n", "n-gram length", 2, 1, 10}
	ng := textEntry("n-grams", "Character N-Grams", "ngram.go", "NGrams",
		"Every run of n characters, overlapping: the building block of fuzzy matching, language detection and plagiarism checks.",
		"O(n * len)", "O(n * len)", Text, &Input{Text: "banana", Options: map[string]int{"n": 2}}, maxText)
	ng.Options = []Option{gramSize}
	run(&ng, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		gs := text.NGrams(in.Text, in.opt("n", 2))
		return Output{Lines: gs, Number: ptr(len(gs))}, nil
	})
	ng.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.NGrams(%q, %d)", in.Text, in.opt("n", 2))
	}

	jc := textEntry("ngram-similarity", "N-Gram Similarity (Jaccard)", "ngram.go", "Jaccard",
		"How alike two texts are, from 0 to 1: the n-grams they share over all the n-grams either has.",
		"O(n * (len a + len b))", "O(n * (len a + len b))", TwoTexts, &Input{Text: "night", Other: "nacht", Options: map[string]int{"n": 2}}, maxText)
	jc.Options = []Option{gramSize}
	run(&jc, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		j, shared := text.Jaccard(in.Text, in.Other, in.opt("n", 2))
		return Output{Lines: shared, Summary: map[string]float64{"similarity": j, "shared": float64(len(shared))}}, nil
	})
	jc.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.Jaccard(%q, %q, %d)", in.Text, in.Other, in.opt("n", 2))
	}

	wrapOut := func(lines []string) Output {
		return Output{Lines: lines, Number: ptr(len(lines))}
	}
	wg := textEntry("word-wrap-greedy", "Word Wrap (greedy)", "wrap.go", "WrapGreedy",
		"Fill each line with as many words as fit, then start the next: what most editors do.",
		"O(n)", "O(n)", Text, &Input{Text: pangramText, Options: map[string]int{"width": 20}}, maxText)
	wg.Options = []Option{width}
	run(&wg, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		return wrapOut(text.WrapGreedy(in.Text, in.opt("width", 30))), nil
	})
	wg.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.WrapGreedy(%q, %d)", in.Text, in.opt("width", 30))
	}

	wb := textEntry("word-wrap-balanced", "Word Wrap (minimum raggedness)", "wrap.go", "WrapBalanced",
		"Break lines to minimize the squared leftover space, ending a line early when it saves a ragged one later, by dynamic programming (the TeX idea).",
		"O(w^2)", "O(w)", Text, &Input{Text: pangramText, Options: map[string]int{"width": 20}}, maxText)
	wb.Options = []Option{width}
	run(&wb, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		lines, cost := text.WrapBalanced(in.Text, in.opt("width", 30))
		out := wrapOut(lines)
		out.Summary = map[string]float64{"raggedness": float64(cost)}
		return out, nil
	})
	wb.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.WrapBalanced(%q, %d)", in.Text, in.opt("width", 30))
	}

	js := textEntry("text-justification", "Text Justification", "wrap.go", "Justify",
		"Wrap words greedily, then spread spaces between them so every line but the last is exactly the width, extra spaces going left first.",
		"O(n)", "O(n)", Text, &Input{Text: "This is an example of text justification.", Options: map[string]int{"width": 16}}, maxText)
	js.Options = []Option{width}
	run(&js, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		return wrapOut(text.Justify(in.Text, in.opt("width", 30))), nil
	})
	js.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.Justify(%q, %d)", in.Text, in.opt("width", 30))
	}

	df := textEntry("myers-diff", "Line Diff (Myers)", "diff.go", "Diff",
		"The fewest line insertions and deletions turning one text into another, exploring diagonals one edit at a time (Myers, 1986): the algorithm behind git diff.",
		"O((n + m) * d)", "O((n + m) * d)", TwoTexts, &Input{Text: "a\nb\nc\na\nb\nb\na", Other: "c\nb\na\nb\na\nc"}, 4000)
	run(&df, 4000, func(in Input, _ trace.Recorder) (Output, error) {
		out := Output{}
		counts := map[byte]int{}
		for _, e := range text.Diff(textLines(in.Text), textLines(in.Other)) {
			counts[e.Op]++
			mark := " "
			if e.Op != '=' {
				mark = string(e.Op)
			}
			out.Lines = append(out.Lines, mark+" "+e.Line)
		}
		out.Summary = map[string]float64{"unchanged": float64(counts['=']), "deleted": float64(counts['-']), "inserted": float64(counts['+'])}
		return out, nil
	})
	df.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.Diff(%s, %s)", goStrings(textLines(in.Text)), goStrings(textLines(in.Other)))
	}

	sx := textEntry("soundex", "Soundex", "soundex.go", "Soundex",
		"A letter and three digits per name, so names that sound alike match: Robert and Rupert are both R163 (Russell and Odell, 1918).",
		"O(n)", "O(1)", Text, &Input{Text: "Robert Rupert Rubin Ashcraft Tymczak Pfister"}, maxText)
	run(&sx, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		out := Output{}
		for _, w := range strings.Fields(in.Text) {
			if code := text.Soundex(w); code != "" {
				out.Lines = append(out.Lines, w+" "+code)
			}
		}
		return out, nil
	})
	sx.Call = func(in Input) (string, string) { return "", fmt.Sprintf("text.Soundex(%q)", firstWord(in.Text)) }

	ac := textEntry("autocomplete", "Autocomplete (trie)", "autocomplete.go", "Autocomplete",
		"Suggest the words that start with what was typed, most used first, by walking a trie of every word down to the prefix.",
		"O(prefix + matches)", "O(total length)", Pattern, &Input{Text: "car cart care car cat dog cart car", Other: "car", Options: map[string]int{"limit": 5}}, maxText)
	ac.Options = []Option{{"limit", "most suggestions", 5, 1, 50}}
	run(&ac, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		got := text.Autocomplete(splitWords(in.Text), strings.TrimSpace(in.Other), in.opt("limit", 5))
		return Output{Lines: got, Number: ptr(len(got))}, nil
	})
	ac.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.Autocomplete(%s, %q, %d)", goStrings(splitWords(in.Text)), strings.TrimSpace(in.Other), in.opt("limit", 5))
	}

	lcp := textEntry("longest-common-prefix", "Longest Common Prefix", "repeated.go", "LongestCommonPrefix",
		"The longest start every word shares, by trimming a candidate until each word begins with it.",
		"O(total length)", "O(1)", Text, &Input{Text: "flower flow flight"}, maxText)
	run(&lcp, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		p := text.LongestCommonPrefix(splitWords(in.Text))
		return Output{Text: p, Number: ptr(len(p))}, nil
	})
	lcp.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.LongestCommonPrefix(%s)", goStrings(splitWords(in.Text)))
	}

	sg := textEntry("spelling-suggestions", "Spelling Suggestions", "suggest.go", "Suggest",
		"The dictionary words within k edits of a misspelled word, closest first: Levenshtein distance that skips hopeless words early.",
		"O(d * n * m)", "O(m)", Pattern, &Input{Text: "apple apply ample maple angle", Other: "appel", Options: map[string]int{"distance": 2}}, maxText)
	sg.Options = []Option{{"distance", "most edits", 2, 0, 5}}
	run(&sg, maxText, func(in Input, _ trace.Recorder) (Output, error) {
		got := text.Suggest(splitWords(in.Text), strings.TrimSpace(in.Other), in.opt("distance", 2))
		out := Output{Number: ptr(len(got))}
		for _, s := range got {
			out.Lines = append(out.Lines, fmt.Sprintf("%s (%d)", s.Word, s.Distance))
		}
		return out, nil
	})
	sg.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("text.Suggest(%s, %q, %d)", goStrings(splitWords(in.Text)), strings.TrimSpace(in.Other), in.opt("distance", 2))
	}

	return []Entry{lrs, lcs, ds, rot, lyn, per, fz, wf, tf, ng, jc, wg, wb, js, df, sx, ac, lcp, sg}
}

func firstWord(s string) string {
	if f := strings.Fields(s); len(f) > 0 {
		return f[0]
	}
	return ""
}
