package databases

import (
	"fmt"
	"strconv"
	"strings"
)

// Table is a relation: named columns and rows of text values. Values compare
// as numbers when both are whole numbers, as text otherwise.
type Table struct {
	Columns []string
	Rows    [][]string
}

// ParseTable reads a table written as comma-separated lines, the first line
// naming the columns ("id,name"). Blank lines are skipped; every row must have
// one value per column.
//
// Time: O(n) in the length of text.
func ParseTable(text string) (Table, error) {
	var t Table
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if t.Columns == nil {
			t.Columns = fields
			continue
		}
		if len(fields) != len(t.Columns) {
			return Table{}, fmt.Errorf("%w: line %d has %d values, want %d (%s)", ErrInput, n+1, len(fields), len(t.Columns), strings.Join(t.Columns, ","))
		}
		t.Rows = append(t.Rows, fields)
	}
	if len(t.Columns) == 0 {
		return Table{}, fmt.Errorf("%w: a table needs a header line naming its columns", ErrInput)
	}
	return t, nil
}

// Column returns the index of the column called name, or -1.
func (t Table) Column(name string) int {
	for i, c := range t.Columns {
		if strings.EqualFold(c, name) {
			return i
		}
	}
	return -1
}

// String writes the table back as comma-separated lines, header first.
func (t Table) String() string {
	lines := make([]string, 0, len(t.Rows)+1)
	lines = append(lines, strings.Join(t.Columns, ","))
	for _, r := range t.Rows {
		lines = append(lines, strings.Join(r, ","))
	}
	return strings.Join(lines, "\n")
}

// Lines is the table as lines of text, header first.
func (t Table) Lines() []string { return strings.Split(t.String(), "\n") }

// compareValues orders two values: numerically when both are whole numbers,
// as text otherwise.
func compareValues(a, b string) int {
	x, errA := strconv.Atoi(a)
	y, errB := strconv.Atoi(b)
	if errA == nil && errB == nil {
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	}
	return strings.Compare(a, b)
}

// column returns the index of name in t, or an error naming the columns.
func column(t Table, name string) (int, error) {
	if i := t.Column(name); i >= 0 {
		return i, nil
	}
	return -1, fmt.Errorf("%w: no column %q (the columns are %s)", ErrInput, name, strings.Join(t.Columns, ", "))
}
