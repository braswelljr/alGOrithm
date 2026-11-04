package dataset

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		text string
		opts Options
		want []int
		err  string // substring of the error; "" for none
	}{
		{"commas", "5, 3, 9", Options{}, []int{5, 3, 9}, ""},
		{"spaces and lines", "5 3\n9\t-1", Options{}, []int{5, 3, 9, -1}, ""},
		{"json array", "[1,2,3]", Options{}, []int{1, 2, 3}, ""},
		{"underscores", "1_000 2", Options{}, []int{1000, 2}, ""},
		{"decimal refused", "1.5 2", Options{}, nil, `line 1, column 1: "1.5" is not an integer`},
		{"decimal rounded", "1.5 2.4", Options{Round: true}, []int{2, 2}, ""},
		{"word", "4\n7 x", Options{}, nil, `line 2, column 2: "x" is not a number`},
		{"too large", "1 2 3", Options{Max: 2}, nil, "too many values"},
		{"empty", "  \n ", Options{}, nil, "no values"},
		{"overflow", "99999999999999999999", Options{}, nil, "out of range"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.text, tc.opts)
			if tc.err == "" {
				if err != nil || !reflect.DeepEqual(got, tc.want) {
					t.Errorf("Parse(%q) = %v, %v; want %v", tc.text, got, err, tc.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("Parse(%q) error = %v, want one containing %q", tc.text, err, tc.err)
			}
		})
	}
}

func TestParseReportsEveryBadValue(t *testing.T) {
	_, err := Parse("a 1 b\nc", Options{})
	var es Errors
	if !errors.As(err, &es) || len(es) != 3 {
		t.Fatalf("Parse error = %v, want 3 ValueErrors", err)
	}
	if es[2].Line != 2 || es[2].Column != 1 || es[2].Value != "c" {
		t.Errorf("third error = %+v, want line 2 column 1 value c", es[2])
	}
}

const csvFile = `name,score,age
ada,90,36
alan,"1,200",41
grace,,85
`

func TestCSV(t *testing.T) {
	tab, err := ReadCSV(strings.NewReader(csvFile), true, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if tab.Find(" Score ") != 1 || tab.Find("missing") != -1 || tab.Width() != 3 {
		t.Errorf("Find/Width wrong: %d %d %d", tab.Find("Score"), tab.Find("missing"), tab.Width())
	}
	if got, err := tab.Column(2, false, Options{}); err != nil || !reflect.DeepEqual(got, []int{36, 41, 85}) {
		t.Errorf("age = %v, %v", got, err)
	}
	if got, err := tab.Column(1, true, Options{}); err != nil || !reflect.DeepEqual(got, []int{90, 1200}) {
		t.Errorf("score skipping empty = %v, %v", got, err)
	}
	if _, err := tab.Column(1, false, Options{}); err == nil || !strings.Contains(err.Error(), `line 4, column 2: "" is empty`) {
		t.Errorf("score with empty cell: error = %v", err)
	}
	if _, err := tab.Column(0, false, Options{}); err == nil || !strings.Contains(err.Error(), `line 2, column 1: "ada" is not a number`) {
		t.Errorf("name column: error = %v", err)
	}
	if _, err := tab.Column(5, false, Options{}); err == nil {
		t.Error("column 6 of a 3-column file succeeded")
	}
	if _, err := ReadCSV(strings.NewReader("a,b\n"), true, Options{}); !errors.Is(err, ErrEmpty) {
		t.Errorf("header only: error = %v, want ErrEmpty", err)
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name, data, field string
		want              []int
		err               string
	}{
		{"numbers", `[3, 1, 2]`, "", []int{3, 1, 2}, ""},
		{"numeric strings", `["3", "-1"]`, "", []int{3, -1}, ""},
		{"objects", `[{"v": 5}, {"v": 7}]`, "v", []int{5, 7}, ""},
		{"wrapped", `{"data": [{"v": 5}]}`, "data", nil, "name the field"},
		{"object field", `{"values": [9, 8]}`, "values", []int{9, 8}, ""},
		{"missing field", `[{"v": 5}, {"w": 7}]`, "v", nil, `item 2: "{...}" is missing the field "v"`},
		{"bool", `[1, true]`, "", nil, `item 2: "true" is not a number`},
		{"not an array", `42`, "", nil, "not an array"},
		{"broken", `[1, 2`, "", nil, "reading JSON"},
		{"decimal", `[1.25]`, "", nil, "not an integer"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseJSON([]byte(tc.data), tc.field, Options{})
			if tc.err == "" {
				if err != nil || !reflect.DeepEqual(got, tc.want) {
					t.Errorf("ParseJSON(%s) = %v, %v; want %v", tc.data, got, err, tc.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("ParseJSON(%s) error = %v, want one containing %q", tc.data, err, tc.err)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	for _, k := range Kinds {
		a, err := Generate(k, 50, 7)
		if err != nil || len(a) != 50 {
			t.Fatalf("Generate(%s) = %d values, %v", k, len(a), err)
		}
		b, _ := Generate(k, 50, 7)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("Generate(%s) is not reproducible from its seed", k)
		}
	}
	s, _ := Generate(Sorted, 100, 1)
	for i := 1; i < len(s); i++ {
		if s[i-1] > s[i] {
			t.Fatalf("sorted dataset out of order at %d", i)
		}
	}
	r, _ := Generate(Reversed, 100, 1)
	for i := 1; i < len(r); i++ {
		if r[i-1] < r[i] {
			t.Fatalf("reversed dataset out of order at %d", i)
		}
	}
	few, _ := Generate(FewUnique, 200, 1)
	distinct := map[int]bool{}
	for _, v := range few {
		distinct[v] = true
	}
	if len(distinct) > 5 {
		t.Errorf("few-unique has %d distinct values", len(distinct))
	}
	if _, err := Generate("zigzag", 5, 1); err == nil {
		t.Error("unknown kind accepted")
	}
	if _, err := Generate(Random, MaxValues+1, 1); err == nil {
		t.Error("oversized dataset accepted")
	}
	if got, err := Generate(Random, 0, 1); err != nil || len(got) != 0 {
		t.Errorf("size 0 = %v, %v", got, err)
	}
}
