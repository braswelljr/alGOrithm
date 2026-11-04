package databases

import (
	"fmt"
	"math"
	"strconv"
)

// QueryPlan is what Plan chose for "SELECT * FROM t WHERE cond", and why.
type QueryPlan struct {
	Predicate Predicate
	Rows      int // in the table
	Pages     int // the table's pages: rows / rows per page, rounded up
	// Selectivity is the estimated fraction of rows matching; EstimatedRows
	// that fraction of Rows. ActualRows is what running the plan found.
	Selectivity   float64
	EstimatedRows float64
	ActualRows    int
	// The costs are in page reads: a full scan reads every page; an index
	// scan walks the index (its height) and then reads one page per match,
	// since the table is not stored in index order.
	FullScanCost, IndexScanCost float64
	Indexed                     bool   // whether the condition's column has an index
	Access                      string // "full scan" or "index scan"
	Result                      Table
	Explain                     []string
}

// Plan picks how to answer one condition on t the way a cost-based optimizer
// does, in miniature: estimate how many rows match from statistics, price
// each access path in page reads, and take the cheaper. The statistics are
// the column's distinct count (an equality matches 1/distinct of the rows)
// and its minimum and maximum (a range matches its share of that span,
// assuming values spread evenly). indexed names the column that has a B+
// tree index of the given fanout, or "" for none.
//
// An index wins for a selective condition and loses for a broad one, where
// one random page read per match costs more than reading every page in
// order. Simplified: one table, one condition, no clustered indexes, no
// caching. Time: O(n).
func Plan(t Table, where, indexed string, rowsPerPage, fanout int) (QueryPlan, error) {
	p, err := ParsePredicate(where)
	if err != nil {
		return QueryPlan{}, err
	}
	col, err := column(t, p.Column)
	if err != nil {
		return QueryPlan{}, err
	}
	rowsPerPage, fanout = max(1, rowsPerPage), max(2, fanout)
	n := len(t.Rows)
	q := QueryPlan{Predicate: p, Rows: n, Pages: (n + rowsPerPage - 1) / rowsPerPage, Indexed: indexed != "" && t.Column(indexed) == col}
	q.Selectivity = selectivity(t, col, p)
	q.EstimatedRows = q.Selectivity * float64(n)
	q.FullScanCost = float64(q.Pages)
	height := 1.0
	if n > 1 {
		height = math.Ceil(math.Log(float64(n)) / math.Log(float64(fanout)))
	}
	q.IndexScanCost = height + q.EstimatedRows
	q.Explain = append(q.Explain,
		fmt.Sprintf("%d rows in %d pages of %d rows", n, q.Pages, rowsPerPage),
		fmt.Sprintf("estimated selectivity of %s %s %s: %.3f, so %.1f rows", p.Column, p.Op, p.Value, q.Selectivity, q.EstimatedRows),
		fmt.Sprintf("full scan: %.0f page reads", q.FullScanCost))
	q.Access = "full scan"
	if q.Indexed {
		q.Explain = append(q.Explain, fmt.Sprintf("index scan: %.0f to walk the index + %.1f for the rows = %.1f page reads", height, q.EstimatedRows, q.IndexScanCost))
		if q.IndexScanCost < q.FullScanCost {
			q.Access = "index scan"
		}
	} else {
		q.Explain = append(q.Explain, fmt.Sprintf("no index on %s, so a full scan", p.Column))
	}
	q.Explain = append(q.Explain, "chosen: "+q.Access)
	if q.Result, err = Select(t, p); err != nil {
		return QueryPlan{}, err
	}
	q.ActualRows = len(q.Result.Rows)
	q.Explain = append(q.Explain, fmt.Sprintf("actual rows: %d", q.ActualRows))
	return q, nil
}

// selectivity estimates the fraction of rows of t satisfying p from the
// column's statistics only, as an optimizer would before running anything.
func selectivity(t Table, col int, p Predicate) float64 {
	if len(t.Rows) == 0 {
		return 0
	}
	distinct := map[string]bool{}
	lo, hi, numeric := math.Inf(1), math.Inf(-1), true
	for _, row := range t.Rows {
		distinct[row[col]] = true
		v, err := strconv.ParseFloat(row[col], 64)
		if err != nil {
			numeric = false
			continue
		}
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	eq := 1 / float64(len(distinct))
	switch p.Op {
	case "=":
		return eq
	case "!=":
		return 1 - eq
	}
	v, err := strconv.ParseFloat(p.Value, 64)
	if !numeric || err != nil || hi == lo {
		return 1.0 / 3 // the textbook guess for a range with no statistics
	}
	below := math.Max(0, math.Min(1, (v-lo)/(hi-lo)))
	if p.Op == "<" || p.Op == "<=" {
		return below
	}
	return 1 - below
}
