package text

import (
	"math"
	"sort"
)

// WordCount is a word and how often it occurs.
type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// WordFrequency counts the words of s (see Words), most frequent first, ties
// in alphabetical order.
//
// Time: O(n + w log w) for w distinct words. Space: O(w).
func WordFrequency(s string) []WordCount {
	counts := map[string]int{}
	for _, w := range Words(s) {
		counts[w]++
	}
	out := make([]WordCount, 0, len(counts))
	for w, c := range counts {
		out = append(out, WordCount{w, c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Word < out[j].Word
	})
	return out
}

// Term is a word and its score in one document.
type Term struct {
	Word  string  `json:"word"`
	Score float64 `json:"score"`
}

// TFIDF scores each word of each document by term frequency times inverse
// document frequency: (count in the document / words in the document) *
// ln(documents / documents containing the word). A word in every document
// scores 0; a word frequent in one document and rare elsewhere scores high.
// Each document's terms come highest first, ties in alphabetical order.
//
// Time: O(total words + t log t) per document. Space: O(distinct words).
func TFIDF(docs []string) [][]Term {
	words := make([][]string, len(docs))
	df := map[string]int{}
	for i, d := range docs {
		words[i] = Words(d)
		seen := map[string]bool{}
		for _, w := range words[i] {
			if !seen[w] {
				seen[w] = true
				df[w]++
			}
		}
	}
	out := make([][]Term, len(docs))
	for i, ws := range words {
		tf := map[string]int{}
		for _, w := range ws {
			tf[w]++
		}
		terms := make([]Term, 0, len(tf))
		for w, c := range tf {
			idf := math.Log(float64(len(docs)) / float64(df[w]))
			terms = append(terms, Term{w, float64(c) / float64(len(ws)) * idf})
		}
		sort.Slice(terms, func(a, b int) bool {
			if terms[a].Score != terms[b].Score {
				return terms[a].Score > terms[b].Score
			}
			return terms[a].Word < terms[b].Word
		})
		out[i] = terms
	}
	return out
}
