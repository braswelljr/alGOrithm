package databases

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Predicate is a condition on one column: Column Op Value, Op one of = != <
// <= > >=.
type Predicate struct {
	Column, Op, Value string
}

// ParsePredicate reads "age > 30", "city = Accra" or "age>=30".
func ParsePredicate(s string) (Predicate, error) {
	for _, op := range []string{"!=", "<=", ">=", "=", "<", ">"} {
		if a, b, ok := strings.Cut(s, op); ok {
			p := Predicate{Column: strings.TrimSpace(a), Op: op, Value: strings.TrimSpace(b)}
			if p.Column == "" || p.Value == "" {
				break
			}
			return p, nil
		}
	}
	return Predicate{}, fmt.Errorf("%w: cannot read the condition %q (write it like \"age > 30\")", ErrInput, s)
}

// holds reports whether the value v satisfies the predicate.
func (p Predicate) holds(v string) bool {
	c := compareValues(v, p.Value)
	switch p.Op {
	case "=":
		return c == 0
	case "!=":
		return c != 0
	case "<":
		return c < 0
	case "<=":
		return c <= 0
	case ">":
		return c > 0
	default: // ">="
		return c >= 0
	}
}

// Select (sigma) keeps the rows satisfying p. Time: O(n).
func Select(t Table, p Predicate) (Table, error) {
	i, err := column(t, p.Column)
	if err != nil {
		return Table{}, err
	}
	out := Table{Columns: t.Columns}
	for _, row := range t.Rows {
		if p.holds(row[i]) {
			out.Rows = append(out.Rows, row)
		}
	}
	return out, nil
}

// Project (pi) keeps the named columns, in that order, and drops duplicate
// rows, as relational algebra's sets do (SQL keeps them unless asked).
// Time: O(n * c).
func Project(t Table, cols []string) (Table, error) {
	idx := make([]int, len(cols))
	for k, c := range cols {
		i, err := column(t, c)
		if err != nil {
			return Table{}, err
		}
		idx[k] = i
	}
	out := Table{Columns: make([]string, len(cols))}
	for k, i := range idx {
		out.Columns[k] = t.Columns[i]
	}
	seen := map[string]bool{}
	for _, row := range t.Rows {
		nr := make([]string, len(idx))
		for k, i := range idx {
			nr[k] = row[i]
		}
		key := strings.Join(nr, "\x00")
		if !seen[key] {
			seen[key] = true
			out.Rows = append(out.Rows, nr)
		}
	}
	return out, nil
}

// GroupBy (gamma) groups the rows by col and computes one aggregate per
// group: "count", or "sum", "avg", "min" or "max" of the whole-number column
// of. Groups come out in the order they first appear; avg rounds down.
// Time: O(n).
func GroupBy(t Table, col, agg, of string) (Table, error) {
	g, err := column(t, col)
	if err != nil {
		return Table{}, err
	}
	a := -1
	if agg != "count" {
		if a, err = column(t, of); err != nil {
			return Table{}, err
		}
	}
	switch agg {
	case "count", "sum", "avg", "min", "max":
	default:
		return Table{}, fmt.Errorf("%w: unknown aggregate %q (use count, sum, avg, min or max)", ErrInput, agg)
	}
	type acc struct{ n, sum, lo, hi int }
	var order []string
	groups := map[string]*acc{}
	for _, row := range t.Rows {
		k := row[g]
		c, ok := groups[k]
		if !ok {
			c = &acc{}
			groups[k] = c
			order = append(order, k)
		}
		v := 0
		if a >= 0 {
			if v, err = strconv.Atoi(row[a]); err != nil {
				return Table{}, fmt.Errorf("%w: %s of %q needs whole numbers", ErrInput, agg, of)
			}
		}
		if c.n == 0 || v < c.lo {
			c.lo = v
		}
		if c.n == 0 || v > c.hi {
			c.hi = v
		}
		c.n++
		c.sum += v
	}
	name := agg
	if a >= 0 {
		name = agg + "(" + of + ")"
	}
	out := Table{Columns: []string{t.Columns[g], name}}
	for _, k := range order {
		c := groups[k]
		v := c.n
		switch agg {
		case "sum":
			v = c.sum
		case "avg":
			v = c.sum / c.n
		case "min":
			v = c.lo
		case "max":
			v = c.hi
		}
		out.Rows = append(out.Rows, []string{k, strconv.Itoa(v)})
	}
	return out, nil
}

// Stage is one operator of a query and the table it produced.
type Stage struct {
	Op    string
	Table Table
}

// Query runs a pipeline of relational operators over t, separated by "|":
//
//	select COND           keep rows where COND holds ("age > 30")
//	project A,B           keep columns A and B, without duplicate rows
//	group COL AGG [X]     one row per COL value: count, or sum/avg/min/max of X
//	sort COL [desc]       order the rows by COL
//	limit N               keep the first N rows
//
// It returns the table after each stage. Simplified: one table, no joins
// (the join entries do those), one condition per select. Time: O(n log n)
// per sort, O(n) per other stage.
func Query(t Table, query string) ([]Stage, error) {
	var stages []Stage
	for _, part := range strings.Split(query, "|") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, rest, _ := strings.Cut(part, " ")
		rest = strings.TrimSpace(rest)
		var err error
		switch strings.ToLower(name) {
		case "select", "where":
			var p Predicate
			if p, err = ParsePredicate(rest); err == nil {
				t, err = Select(t, p)
			}
		case "project":
			t, err = Project(t, splitList(rest))
		case "group":
			f := strings.Fields(rest)
			switch {
			case len(f) == 2:
				t, err = GroupBy(t, f[0], f[1], "")
			case len(f) == 3:
				t, err = GroupBy(t, f[0], f[1], f[2])
			default:
				err = fmt.Errorf("%w: write group COL count, or group COL sum X", ErrInput)
			}
		case "sort", "order":
			t, err = sortStage(t, rest)
		case "limit":
			var n int
			if n, err = strconv.Atoi(rest); err != nil || n < 0 {
				err = fmt.Errorf("%w: limit needs a whole number, not %q", ErrInput, rest)
			} else {
				t = Table{Columns: t.Columns, Rows: t.Rows[:min(n, len(t.Rows))]}
			}
		default:
			err = fmt.Errorf("%w: unknown operator %q (use select, project, group, sort or limit)", ErrInput, name)
		}
		if err != nil {
			return nil, err
		}
		stages = append(stages, Stage{Op: part, Table: t})
	}
	if len(stages) == 0 {
		return nil, fmt.Errorf("%w: an empty query (write it like \"select age > 30 | project name\")", ErrInput)
	}
	return stages, nil
}

func sortStage(t Table, rest string) (Table, error) {
	f := strings.Fields(rest)
	if len(f) == 0 {
		return Table{}, fmt.Errorf("%w: sort needs a column", ErrInput)
	}
	i, err := column(t, f[0])
	if err != nil {
		return Table{}, err
	}
	desc := len(f) > 1 && strings.EqualFold(f[1], "desc")
	rows := slices.Clone(t.Rows)
	slices.SortStableFunc(rows, func(a, b []string) int {
		if desc {
			return compareValues(b[i], a[i])
		}
		return compareValues(a[i], b[i])
	})
	return Table{Columns: t.Columns, Rows: rows}, nil
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
