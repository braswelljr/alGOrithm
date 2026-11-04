package compression

// Run is one run of RunLength: Count copies of Value.
type Run struct {
	Value byte
	Count int
}

// RunLength encodes data as runs of equal bytes: "aaabcc" is a x3, b x1,
// c x2. It shines on long runs (simple images, sparse data) and doubles the
// size of data with none, since each run costs a value and a count.
//
// Time: O(n). Space: O(runs).
func RunLength(data []byte) []Run {
	var runs []Run
	for _, b := range data {
		if n := len(runs); n > 0 && runs[n-1].Value == b {
			runs[n-1].Count++
		} else {
			runs = append(runs, Run{b, 1})
		}
	}
	return runs
}

// RunLengthDecode expands runs back into the bytes, or ErrCorrupt when a
// count is below 1.
//
// Time: O(n). Space: O(n).
func RunLengthDecode(runs []Run) ([]byte, error) {
	var out []byte
	for _, r := range runs {
		if r.Count < 1 {
			return nil, ErrCorrupt
		}
		for k := 0; k < r.Count; k++ {
			out = append(out, r.Value)
		}
	}
	return out, nil
}
