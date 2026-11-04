// Package dataset turns data from outside into the values the algorithms take,
// and generates datasets to try them on. It parses pasted text, CSV and JSON,
// and reports every value it cannot use with where it was, instead of quietly
// dropping it (docs/rules/data-from-outside-is-unknown-until-checked.md).
package dataset

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// MaxValues is the default limit on how many values one dataset may hold.
const MaxValues = 100_000

// maxErrors is how many bad values an Errors lists before it stops looking.
const maxErrors = 20

// ValueError is one value that could not be used.
type ValueError struct {
	Line   int    `json:"line,omitempty"`   // 1-based line (CSV: row); 0 when not known
	Column int    `json:"column,omitempty"` // 1-based column or item position; 0 when not known
	Value  string `json:"value"`            // the text as it appeared
	Reason string `json:"reason"`           // "not an integer", "out of range", ...
}

func (e ValueError) Error() string {
	where := ""
	switch {
	case e.Line > 0 && e.Column > 0:
		where = fmt.Sprintf("line %d, column %d: ", e.Line, e.Column)
	case e.Line > 0:
		where = fmt.Sprintf("line %d: ", e.Line)
	case e.Column > 0:
		where = fmt.Sprintf("item %d: ", e.Column)
	}
	return fmt.Sprintf("%s%q is %s", where, e.Value, e.Reason)
}

// Errors is every ValueError found in one input, up to a limit.
type Errors []ValueError

func (es Errors) Error() string {
	if len(es) == 1 {
		return es[0].Error()
	}
	parts := make([]string, len(es))
	for i, e := range es {
		parts[i] = e.Error()
	}
	return fmt.Sprintf("%d values could not be read: %s", len(es), strings.Join(parts, "; "))
}

// ErrTooLarge is returned when an input holds more values than its limit.
var ErrTooLarge = errors.New("dataset: too many values")

// ErrEmpty is returned when an input holds no values at all.
var ErrEmpty = errors.New("dataset: no values")

// Options controls how values are read. The zero Options reads integers only,
// up to MaxValues.
type Options struct {
	// Max is the most values accepted; 0 means MaxValues.
	Max int
	// Round accepts decimal numbers and rounds them to the nearest integer.
	// Without it a decimal is an error, never silently truncated.
	Round bool
}

func (o Options) max() int {
	if o.Max > 0 {
		return o.Max
	}
	return MaxValues
}

// errorList collects ValueErrors up to maxErrors.
type errorList struct{ errs Errors }

func (l *errorList) add(e ValueError) bool {
	if len(l.errs) < maxErrors {
		l.errs = append(l.errs, e)
	}
	return len(l.errs) < maxErrors
}

func (l *errorList) err() error {
	if len(l.errs) == 0 {
		return nil
	}
	return l.errs
}

// ParseInt reads one value: an integer, optionally signed, with "_" or ","
// digit separators ("1,000" is one thousand only inside a CSV cell or JSON
// string, since Parse splits on commas), or a decimal when opts.Round is set.
func ParseInt(s string, opts Options) (int, string) {
	s = strings.TrimSpace(s)
	clean := strings.ReplaceAll(strings.ReplaceAll(s, "_", ""), ",", "")
	if v, err := strconv.ParseInt(clean, 10, 0); err == nil {
		return int(v), ""
	} else if errors.Is(err, strconv.ErrRange) {
		return 0, "out of range"
	}
	f, err := strconv.ParseFloat(clean, 64)
	switch {
	case err != nil || math.IsNaN(f) || math.IsInf(f, 0):
		return 0, "not a number"
	case !opts.Round && f != math.Trunc(f):
		return 0, "not an integer"
	case math.Abs(f) > math.MaxInt64/2:
		return 0, "out of range"
	}
	return int(math.Round(f)), ""
}

// Parse reads integers from pasted or typed text, separated by commas,
// semicolons, spaces, tabs or new lines ("5, 3, 9", "5 3 9", one per line).
// Brackets around the list are allowed, so a JSON array of numbers pastes too.
func Parse(text string, opts Options) ([]int, error) {
	var (
		out   []int
		bad   errorList
		limit = opts.max()
		line  = 1
	)
	for _, row := range strings.Split(text, "\n") {
		col := 0
		fields := strings.FieldsFunc(row, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\r' || r == '[' || r == ']'
		})
		for _, f := range fields {
			col++
			v, reason := ParseInt(f, opts)
			if reason != "" {
				if !bad.add(ValueError{Line: line, Column: col, Value: f, Reason: reason}) {
					return nil, bad.err()
				}
				continue
			}
			if len(out) == limit {
				return nil, fmt.Errorf("%w: more than %d", ErrTooLarge, limit)
			}
			out = append(out, v)
		}
		line++
	}
	if err := bad.err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrEmpty
	}
	return out, nil
}

// Table is a CSV file read into rows of cells.
type Table struct {
	Header []string   // the first row, when the file has a header
	Rows   [][]string // the data rows
}

// ReadCSV reads a CSV file. header says whether the first row names the
// columns. Rows may have different lengths. At most opts.Max rows are read.
func ReadCSV(r io.Reader, header bool, opts Options) (*Table, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	cr.ReuseRecord = false
	t := &Table{}
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("dataset: reading CSV: %w", err)
		}
		if header && t.Header == nil {
			t.Header = rec
			continue
		}
		if len(t.Rows) == opts.max() {
			return nil, fmt.Errorf("%w: more than %d rows", ErrTooLarge, opts.max())
		}
		t.Rows = append(t.Rows, rec)
	}
	if len(t.Rows) == 0 {
		return nil, ErrEmpty
	}
	return t, nil
}

// Width is the number of columns of the widest row (or of the header).
func (t *Table) Width() int {
	w := len(t.Header)
	for _, r := range t.Rows {
		if len(r) > w {
			w = len(r)
		}
	}
	return w
}

// Find returns the index of the column whose header is name (ignoring case and
// surrounding space), or -1.
func (t *Table) Find(name string) int {
	for i, h := range t.Header {
		if strings.EqualFold(strings.TrimSpace(h), strings.TrimSpace(name)) {
			return i
		}
	}
	return -1
}

// Column reads column col (0-based) as integers. An empty cell is an error
// unless skipEmpty is set. Errors give the file's line, counting the header.
func (t *Table) Column(col int, skipEmpty bool, opts Options) ([]int, error) {
	if col < 0 || col >= t.Width() {
		return nil, fmt.Errorf("dataset: no column %d (the file has %d)", col+1, t.Width())
	}
	first := 1
	if t.Header != nil {
		first = 2
	}
	var (
		out []int
		bad errorList
	)
	for i, row := range t.Rows {
		cell := ""
		if col < len(row) {
			cell = strings.TrimSpace(row[col])
		}
		if cell == "" && skipEmpty {
			continue
		}
		v, reason := ParseInt(cell, opts)
		if cell == "" {
			reason = "empty"
		}
		if reason != "" {
			if !bad.add(ValueError{Line: first + i, Column: col + 1, Value: cell, Reason: reason}) {
				break
			}
			continue
		}
		out = append(out, v)
	}
	if err := bad.err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrEmpty
	}
	return out, nil
}

// ParseJSON reads integers from JSON: an array of numbers (or numeric
// strings), an array of objects from which field is taken, or an object whose
// field holds such an array. field may be empty for a plain array.
func ParseJSON(data []byte, field string, opts Options) ([]int, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("dataset: reading JSON: %w", err)
	}
	if obj, ok := v.(map[string]any); ok {
		if field == "" {
			return nil, errors.New("dataset: the JSON is an object; name the field that holds the values")
		}
		inner, ok := obj[field]
		if !ok {
			return nil, fmt.Errorf("dataset: the JSON object has no field %q", field)
		}
		v, field = inner, ""
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, errors.New("dataset: the JSON is not an array")
	}
	if len(arr) > opts.max() {
		return nil, fmt.Errorf("%w: more than %d", ErrTooLarge, opts.max())
	}
	var (
		out []int
		bad errorList
	)
	for i, item := range arr {
		if obj, ok := item.(map[string]any); ok {
			if field == "" {
				return nil, errors.New("dataset: the JSON array holds objects; name the field to read")
			}
			item, ok = obj[field]
			if !ok {
				if !bad.add(ValueError{Column: i + 1, Value: "{...}", Reason: fmt.Sprintf("missing the field %q", field)}) {
					break
				}
				continue
			}
		}
		var text string
		switch x := item.(type) {
		case json.Number:
			text = x.String()
		case string:
			text = x
		default:
			text = fmt.Sprint(x)
		}
		n, reason := ParseInt(text, opts)
		if _, isString := item.(string); !isString {
			if _, isNumber := item.(json.Number); !isNumber {
				reason = "not a number"
			}
		}
		if reason != "" {
			if !bad.add(ValueError{Column: i + 1, Value: text, Reason: reason}) {
				break
			}
			continue
		}
		out = append(out, n)
	}
	if err := bad.err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrEmpty
	}
	return out, nil
}
