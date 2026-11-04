package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Snippet is the same run written two ways: a Go program that calls the
// library, and the command that runs it. Go is empty for entries whose call
// takes more than the input can say (a structure's operations, a simulation).
type Snippet struct {
	Go  string `json:"go"`
	CLI string `json:"cli"`
}

// MakeSnippet writes e run on in as Go and as a command line.
func MakeSnippet(e Entry, in Input) Snippet {
	return Snippet{Go: goSnippet(e, in), CLI: cliSnippet(e, in)}
}

func quote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\n'\"\\$`!*?[]{}()<>|&;#~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func cliSnippet(e Entry, in Input) string {
	parts := []string{"algorithm", "run", e.ID}
	switch e.Input {
	case Text, Number, TwoNumbers, Ints, SortedInts, AnyInts, Amount, Pages, Tracks:
		for _, v := range in.Values {
			parts = append(parts, strconv.Itoa(v))
		}
		if e.Input == Text {
			parts = append(parts, quote(in.Text))
		}
	case TwoTexts, Pattern:
		parts = append(parts, quote(in.Text), quote(in.Other))
	case GridInput, Processes, Banker, GraphInput, Points:
		parts = append(parts, "--text", quote(in.Text))
	case Partitions:
		parts = append(parts, "--values", quote(joinInts(in.Values)), "--text", quote(in.Text))
	case Ops:
		for _, v := range in.Values {
			parts = append(parts, strconv.Itoa(v))
		}
		var ops []string
		for _, op := range in.Ops {
			ops = append(ops, strings.TrimSpace(op.Name+" "+op.Value))
		}
		if len(ops) > 0 {
			parts = append(parts, "--ops", quote(strings.Join(ops, ", ")))
		}
	case Simulation: // options only, written below
	}
	for _, p := range e.Params {
		switch p {
		case "target":
			parts = append(parts, "--target", strconv.Itoa(in.Target))
		case "amount":
			parts = append(parts, "--amount", strconv.Itoa(in.Target))
		case "shift":
			parts = append(parts, "--shift", strconv.Itoa(in.Shift))
		case "key":
			parts = append(parts, "--key", quote(in.Key))
		case "start":
			if in.Other != "" {
				parts = append(parts, "--other", quote(in.Other))
			}
		case "end":
			if in.Key != "" {
				parts = append(parts, "--key", quote(in.Key))
			}
		}
	}
	names := make([]string, 0, len(in.Options))
	for k := range in.Options {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		parts = append(parts, "--option", k+"="+strconv.Itoa(in.Options[k]))
	}
	return strings.Join(parts, " ")
}

func ints(v []int) string {
	return "[]int{" + strings.Join(strings.Fields(strings.Trim(fmt.Sprint(v), "[]")), ", ") + "}"
}

// calls are the entries whose call the input says completely, by the shape
// of their function.
func goCall(e Entry, in Input) (setup, call string) {
	if e.Call != nil {
		return e.Call(in)
	}
	f := e.Package + "." + e.Func
	switch e.Category {
	case "sorting":
		return "", fmt.Sprintf("%s(%s)", f, ints(in.Values))
	case "searching":
		vals := append([]int(nil), in.Values...)
		if e.Input == SortedInts {
			sort.Ints(vals)
		}
		return "", fmt.Sprintf("%s(%s, %d)", f, ints(vals), in.Target)
	case "strmatch":
		return "", fmt.Sprintf("%s(%q, %q)", f, in.Text, in.Other)
	case "pathfinding":
		return fmt.Sprintf("g, err := pathfinding.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), f + "(g)"
	}
	switch e.ID {
	case "edit-distance", "longest-common-subsequence":
		return "", fmt.Sprintf("%s(%q, %q)", f, in.Text, in.Other)
	case "coin-change":
		return "", fmt.Sprintf("%s(%s, %d)", f, ints(in.Values), in.Target)
	case "longest-increasing-subsequence", "max-subarray", "prefix-sums", "reverse-array":
		return "", fmt.Sprintf("%s(%s)", f, ints(in.Values))
	case "rotate-array":
		return "", fmt.Sprintf("%s(%s, %d)", f, ints(in.Values), in.Target)
	case "sliding-window-max":
		return "", fmt.Sprintf("%s(%s, %d)", f, ints(in.Values), in.opt("window", 3))
	case "factorial", "fibonacci", "fibonacci-memo", "mccarthy-91", "tower-of-hanoi", "prime", "even", "prime-factors", "sieve-of-eratosthenes":
		return "", fmt.Sprintf("%s(%d)", f, firstOr(in.Values, 0))
	case "gcd", "lcm", "ackermann":
		a, b := firstOr(in.Values, 0), 0
		if len(in.Values) > 1 {
			b = in.Values[1]
		}
		return "", fmt.Sprintf("%s(%d, %d)", f, a, b)
	case "palindrome", "rot13":
		return "", fmt.Sprintf("%s(%q)", f, in.Text)
	case "anagram":
		return "", fmt.Sprintf("%s(%q, %q)", f, in.Text, in.Other)
	case "caesar-cipher":
		return "", fmt.Sprintf("%s(%q, %d)", f, in.Text, in.Shift)
	case "vigenere-cipher":
		return "", fmt.Sprintf("%s(%q, %q)", f, in.Text, in.Key)
	case "fifo-page-replacement", "lru-page-replacement", "optimal-page-replacement", "clock-page-replacement":
		return "", fmt.Sprintf("%s(%s, %d)", f, ints(in.Values), in.opt("frames", 3))
	case "fcfs-disk", "sstf-disk":
		return "", fmt.Sprintf("%s(%s, %d, %d)", f, ints(in.Values), in.opt("head", 53), in.opt("size", 200))
	case "scan-disk", "c-scan-disk", "look-disk", "c-look-disk":
		return "", fmt.Sprintf("%s(%s, %d, %d, %v)", f, ints(in.Values), in.opt("head", 53), in.opt("size", 200), in.opt("down", 0) == 1)
	case "fcfs-scheduling", "sjf-scheduling", "srtf-scheduling", "round-robin", "priority-scheduling":
		args := "ps"
		switch e.ID {
		case "round-robin":
			args += ", " + strconv.Itoa(in.opt("quantum", 2))
		case "priority-scheduling":
			args += fmt.Sprintf(", %v", in.opt("preempt", 0) == 1)
		}
		return fmt.Sprintf("ps, err := scheduling.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), f + "(" + args + ")"
	case "race-condition":
		return "", fmt.Sprintf("%s(%d, %d, %v, %d)", f, in.opt("threads", 3), in.opt("increments", 3), in.opt("locked", 0) == 1, in.opt("seed", 1))
	case "dining-philosophers":
		return "", fmt.Sprintf("%s(%d, %d, %v, %d)", f, in.opt("philosophers", 5), in.opt("meals", 2), in.opt("ordered", 0) == 1, in.opt("seed", 1))
	case "producer-consumer":
		return "", fmt.Sprintf("%s(%d, %d, %d, %d, %d)", f, in.opt("producers", 2), in.opt("consumers", 2), in.opt("capacity", 2), in.opt("items", 10), in.opt("seed", 1))
	}
	return "", ""
}

func firstOr(v []int, def int) int {
	if len(v) == 0 {
		return def
	}
	return v[0]
}

func goSnippet(e Entry, in Input) string {
	setup, call := goCall(e, in)
	if call == "" {
		return ""
	}
	return fmt.Sprintf(`package main

import (
	"fmt"

	%q
)

func main() {
	%sfmt.Println(%s)
}
`, e.Import, setup, call)
}
