package text

import "strings"

// WrapGreedy breaks s into lines of at most width characters, putting as
// many words on each line as fit; a word longer than width gets a line of
// its own. Whitespace between words becomes one space.
//
// Time: O(n). Space: O(n).
func WrapGreedy(s string, width int) []string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(s) {
		switch {
		case line == "":
			line = w
		case len(line)+1+len(w) <= width:
			line += " " + w
		default:
			lines = append(lines, line)
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// WrapBalanced breaks s into lines of at most width characters minimizing
// raggedness: the sum over every line but the last of the square of its
// unused width. Unlike WrapGreedy it may end a line early to avoid a very
// short one later (the TeX idea, Knuth and Plass 1981, in its simplest form).
// It returns the lines and their cost; a word longer than width gets a line
// of its own at no cost.
//
// Time: O(w^2) for w words. Space: O(w).
func WrapBalanced(s string, width int) ([]string, int) {
	words := strings.Fields(s)
	n := len(words)
	const inf = int(^uint(0) >> 2)
	best, next := make([]int, n+1), make([]int, n+1)
	for i := n - 1; i >= 0; i-- {
		best[i] = inf
		length := -1
		for j := i; j < n; j++ {
			length += 1 + len(words[j])
			if length > width && j > i {
				break
			}
			cost := 0
			if j < n-1 && length <= width {
				cost = (width - length) * (width - length)
			}
			if cost+best[j+1] < best[i] {
				best[i], next[i] = cost+best[j+1], j+1
			}
		}
	}
	var lines []string
	for i := 0; i < n; i = next[i] {
		lines = append(lines, strings.Join(words[i:next[i]], " "))
	}
	return lines, best[0]
}

// Justify breaks s into lines like WrapGreedy and pads every line but the
// last with spaces between words to exactly width characters, the extra
// spaces going to the leftmost gaps. The last line, and a line of one word,
// stays left-aligned without padding.
//
// Time: O(n). Space: O(n).
func Justify(s string, width int) []string {
	lines := WrapGreedy(s, width)
	for i, l := range lines[:max(0, len(lines)-1)] {
		words := strings.Fields(l)
		if len(words) < 2 || len(l) >= width {
			continue
		}
		gaps := len(words) - 1
		spaces := width - len(l) + gaps
		var b strings.Builder
		for g, w := range words {
			b.WriteString(w)
			if g < gaps {
				n := spaces / gaps
				if g < spaces%gaps {
					n++
				}
				b.WriteString(strings.Repeat(" ", n))
			}
		}
		lines[i] = b.String()
	}
	return lines
}
