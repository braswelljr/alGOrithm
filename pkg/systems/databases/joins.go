package databases

import (
	"slices"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// JoinStats counts a join's work: comparisons of join keys, passes over the
// right (inner) table, and probes of a hash table or an index.
type JoinStats struct {
	Comparisons int
	InnerScans  int
	Probes      int
}

// joinOn resolves the join columns and the output header: the left table's
// columns, then the right's, a right column whose name the left already has
// prefixed with "r.".
func joinOn(left, right Table, lcol, rcol string) (li, ri int, out Table, err error) {
	if li, err = column(left, lcol); err != nil {
		return
	}
	if ri, err = column(right, rcol); err != nil {
		return
	}
	out.Columns = slices.Clone(left.Columns)
	for _, c := range right.Columns {
		if left.Column(c) >= 0 {
			c = "r." + c
		}
		out.Columns = append(out.Columns, c)
	}
	return
}

func joined(a, b []string) []string { return append(slices.Clone(a), b...) }

// compareRows compares left row i's key with right row j's and reports it to
// r as a Compare of [i, j].
func compareRows(left, right Table, li, ri, i, j int, st *JoinStats, r trace.Recorder) int {
	st.Comparisons++
	r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, j}})
	return compareValues(left.Rows[i][li], right.Rows[j][ri])
}

func match(r trace.Recorder, i, j int) {
	r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i, j}, Note: "match"})
}

// NestedLoopJoin joins left and right on left.lcol = right.rcol the simplest
// way: for every left row, scan every right row. Comparisons of [left,
// right] row indexes and matches go to r. Time: O(n * m) comparisons for n
// and m rows; the right table is scanned n times.
func NestedLoopJoin(left, right Table, lcol, rcol string, r trace.Recorder) (Table, JoinStats, error) {
	li, ri, out, err := joinOn(left, right, lcol, rcol)
	var st JoinStats
	if err != nil {
		return Table{}, st, err
	}
	for i := range left.Rows {
		st.InnerScans++
		for j := range right.Rows {
			if compareRows(left, right, li, ri, i, j, &st, r) == 0 {
				match(r, i, j)
				out.Rows = append(out.Rows, joined(left.Rows[i], right.Rows[j]))
			}
		}
	}
	return out, st, nil
}

// BlockNestedLoopJoin is a nested-loop join that reads the left table block
// rows at a time and scans the right table once per block instead of once
// per row: the same comparisons, but n/block passes over the inner table
// instead of n, which is what costs when the inner table is on disk. Time:
// O(n * m) comparisons, ceil(n / block) scans of the right table.
func BlockNestedLoopJoin(left, right Table, lcol, rcol string, block int, r trace.Recorder) (Table, JoinStats, error) {
	li, ri, out, err := joinOn(left, right, lcol, rcol)
	var st JoinStats
	if err != nil {
		return Table{}, st, err
	}
	block = max(1, block)
	for start := 0; start < len(left.Rows); start += block {
		end := min(start+block, len(left.Rows))
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start, end - 1}, Note: "block"})
		st.InnerScans++
		for j := range right.Rows {
			for i := start; i < end; i++ {
				if compareRows(left, right, li, ri, i, j, &st, r) == 0 {
					match(r, i, j)
					out.Rows = append(out.Rows, joined(left.Rows[i], right.Rows[j]))
				}
			}
		}
	}
	return out, st, nil
}

// IndexNestedLoopJoin builds a sorted index on right.rcol, then for every
// left row binary-searches the index instead of scanning the right table:
// the plan a database picks when the inner column is indexed. Time:
// O(m log m) to build the index (free when it already exists), then
// O(n log m + k) for k matches.
func IndexNestedLoopJoin(left, right Table, lcol, rcol string, r trace.Recorder) (Table, JoinStats, error) {
	li, ri, out, err := joinOn(left, right, lcol, rcol)
	var st JoinStats
	if err != nil {
		return Table{}, st, err
	}
	index := make([]int, len(right.Rows)) // right row numbers in key order
	for j := range index {
		index[j] = j
	}
	slices.SortStableFunc(index, func(a, b int) int { return compareValues(right.Rows[a][ri], right.Rows[b][ri]) })
	for i := range left.Rows {
		st.Probes++
		lo, hi := 0, len(index)
		for lo < hi { // the first index entry not below the key
			mid := (lo + hi) / 2
			if compareRows(left, right, li, ri, i, index[mid], &st, r) > 0 {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		for ; lo < len(index) && compareRows(left, right, li, ri, i, index[lo], &st, r) == 0; lo++ {
			match(r, i, index[lo])
			out.Rows = append(out.Rows, joined(left.Rows[i], right.Rows[index[lo]]))
		}
	}
	return out, st, nil
}

// HashJoin builds a hash table on the right table's join column (each row
// reported to r as a Push of its index), then probes it with each left row:
// only rows in the same bucket are compared. Simplified: the hash table is
// a Go map keyed by the value's text, so 1 and 01 do not match. Time:
// O(n + m) expected, plus O(k) for k matches; space O(m).
func HashJoin(left, right Table, lcol, rcol string, r trace.Recorder) (Table, JoinStats, error) {
	li, ri, out, err := joinOn(left, right, lcol, rcol)
	var st JoinStats
	if err != nil {
		return Table{}, st, err
	}
	buckets := map[string][]int{}
	for j, row := range right.Rows {
		r.Record(trace.Event{Kind: trace.Push, Values: []int{j}})
		buckets[row[ri]] = append(buckets[row[ri]], j)
	}
	for i, row := range left.Rows {
		st.Probes++
		for _, j := range buckets[row[li]] {
			if compareRows(left, right, li, ri, i, j, &st, r) == 0 {
				match(r, i, j)
				out.Rows = append(out.Rows, joined(row, right.Rows[j]))
			}
		}
	}
	return out, st, nil
}

// SortMergeJoin sorts both tables on their join columns, then walks them
// together like the merge of merge sort, joining each run of equal keys on
// the left with the matching run on the right. Comparisons index the sorted
// orders. Time: O(n log n + m log m) to sort (free when the inputs are
// already sorted, say by an index), then O(n + m + k) to merge.
func SortMergeJoin(left, right Table, lcol, rcol string, r trace.Recorder) (Table, JoinStats, error) {
	li, ri, out, err := joinOn(left, right, lcol, rcol)
	var st JoinStats
	if err != nil {
		return Table{}, st, err
	}
	ls, rs := sortedBy(left, li), sortedBy(right, ri)
	i, j := 0, 0
	for i < len(ls.Rows) && j < len(rs.Rows) {
		switch c := compareRows(ls, rs, li, ri, i, j, &st, r); {
		case c < 0:
			i++
		case c > 0:
			j++
		default:
			jEnd := j
			for jEnd < len(rs.Rows) && compareValues(rs.Rows[jEnd][ri], rs.Rows[j][ri]) == 0 {
				jEnd++
			}
			for ; i < len(ls.Rows) && compareValues(ls.Rows[i][li], rs.Rows[j][ri]) == 0; i++ {
				for k := j; k < jEnd; k++ {
					match(r, i, k)
					out.Rows = append(out.Rows, joined(ls.Rows[i], rs.Rows[k]))
				}
			}
			j = jEnd
		}
	}
	return out, st, nil
}

func sortedBy(t Table, col int) Table {
	rows := slices.Clone(t.Rows)
	slices.SortStableFunc(rows, func(a, b []string) int { return compareValues(a[col], b[col]) })
	return Table{Columns: t.Columns, Rows: rows}
}

// JoinColumns reads a join condition: "id=dept_id" names a column of each
// table, "id" the same column in both, and "" the first column name both
// tables have.
func JoinColumns(left, right Table, on string) (lcol, rcol string, ok bool) {
	if a, b, cut := strings.Cut(on, "="); cut {
		return strings.TrimSpace(a), strings.TrimSpace(b), true
	}
	if on = strings.TrimSpace(on); on != "" {
		return on, on, true
	}
	for _, c := range left.Columns {
		if right.Column(c) >= 0 {
			return c, c, true
		}
	}
	return "", "", false
}
