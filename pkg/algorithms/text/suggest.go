package text

import "sort"

// Suggestion is a dictionary word and its edit distance from the word asked
// about.
type Suggestion struct {
	Word     string `json:"word"`
	Distance int    `json:"distance"`
}

// Suggest returns the words of dictionary within maxDistance edits
// (Levenshtein: insert, delete or substitute one character) of word, closest
// first, ties in alphabetical order, each word once. A word whose length
// differs by more than maxDistance is skipped without computing anything, and
// a row whose smallest cell passes maxDistance stops the computation early.
//
// Time: O(d * n * m) worst case for d dictionary words. Space: O(m).
func Suggest(dictionary []string, word string, maxDistance int) []Suggestion {
	seen := map[string]bool{}
	var out []Suggestion
	for _, w := range dictionary {
		if seen[w] || abs(len(w)-len(word)) > maxDistance {
			continue
		}
		seen[w] = true
		if d, ok := boundedDistance(w, word, maxDistance); ok {
			out = append(out, Suggestion{w, d})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Distance != out[j].Distance {
			return out[i].Distance < out[j].Distance
		}
		return out[i].Word < out[j].Word
	})
	return out
}

// boundedDistance is the edit distance of a and b if it is at most limit.
func boundedDistance(a, b string, limit int) (int, bool) {
	prev, cur := make([]int, len(b)+1), make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		low := cur[0]
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			low = min(low, cur[j])
		}
		if low > limit {
			return 0, false
		}
		prev, cur = cur, prev
	}
	return prev[len(b)], prev[len(b)] <= limit
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
