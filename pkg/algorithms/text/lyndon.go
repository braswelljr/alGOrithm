package text

import "github.com/braswelljr/alGOrithm/pkg/trace"

// LyndonFactors splits s into its Lyndon factorization by Duval's algorithm
// (1983): the unique way to write s as w1 w2 ... wk with w1 >= w2 >= ... >= wk,
// each word a Lyndon word (strictly smaller than every one of its proper
// rotations). The factors concatenate back to s; "" gives none.
//
// Time: O(n). Space: O(1) beyond the factors. s is not changed.
func LyndonFactors(s string) []string { return lyndonFactors(s, nop) }

// TraceLyndonFactors is LyndonFactors, reporting each comparison and each
// factor as Mark "found".
func TraceLyndonFactors(s string, r trace.Recorder) []string { return lyndonFactors(s, rec{r}) }

func lyndonFactors(s string, x rec) []string {
	var out []string
	n := len(s)
	for i := 0; i < n; {
		j, k := i+1, i
		for j < n {
			x.compare(k, j)
			if s[k] > s[j] {
				break
			}
			if s[k] < s[j] {
				k = i
			} else {
				k++
			}
			j++
		}
		for i <= k {
			x.found(i, i+j-k)
			out = append(out, s[i:i+j-k])
			i += j - k
		}
	}
	return out
}
