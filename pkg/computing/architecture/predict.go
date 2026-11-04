package architecture

import (
	"fmt"
	"strconv"
	"strings"
)

// Branch is one branch as it ran: its address and whether it was taken.
type Branch struct {
	PC    int  `json:"pc"`
	Taken bool `json:"taken"`
}

// Prediction is a predictor's guess for one branch, what really happened,
// and the predictor's state after it learned.
type Prediction struct {
	PC        int    `json:"pc"`
	Predicted bool   `json:"predicted"`
	Taken     bool   `json:"taken"`
	State     string `json:"state"`
}

// ParseBranches reads a branch trace: T (taken) or N (not taken), each
// optionally after an address and a colon ("4:T"), separated by spaces or
// commas. Without an address a branch is at address 0.
func ParseBranches(text string) ([]Branch, error) {
	var out []Branch
	for _, f := range strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' || r == '\t' }) {
		pc := 0
		if a, b, ok := strings.Cut(f, ":"); ok {
			n, err := strconv.Atoi(a)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("%w: %q: the address is not a number", ErrInput, f)
			}
			pc, f = n, b
		}
		for _, c := range strings.ToUpper(f) {
			if c != 'T' && c != 'N' {
				return nil, fmt.Errorf("%w: %q: want T or N", ErrInput, f)
			}
			out = append(out, Branch{PC: pc, Taken: c == 'T'})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no branches", ErrInput)
	}
	return out, nil
}

var counterNames = [4]string{"strongly not taken", "weakly not taken", "weakly taken", "strongly taken"}

// OneBit predicts each branch does what it did last time (not taken the
// first time). A loop's branch costs two misses per loop: at the exit, and
// again on re-entry.
//
// Time: O(n) for n branches.
// Space: O(b) for b distinct branch addresses.
func OneBit(trace []Branch) (preds []Prediction, correct int) {
	last := map[int]bool{}
	for _, b := range trace {
		p := last[b.PC]
		last[b.PC] = b.Taken
		state := "last was not taken"
		if b.Taken {
			state = "last was taken"
		}
		preds = append(preds, Prediction{b.PC, p, b.Taken, state})
		if p == b.Taken {
			correct++
		}
	}
	return preds, correct
}

// TwoBit predicts with a 2-bit saturating counter per branch (James Smith,
// 1981), starting weakly not taken: taken moves it up, not taken down, and
// it predicts taken in its upper two states. One surprise no longer flips
// the prediction, so a loop's branch costs one miss per loop, not two.
//
// Time: O(n) for n branches.
// Space: O(b) for b distinct branch addresses.
func TwoBit(trace []Branch) (preds []Prediction, correct int) {
	counters := map[int]int{}
	for _, b := range trace {
		c, ok := counters[b.PC]
		if !ok {
			c = 1
		}
		p := c >= 2
		counters[b.PC] = step(c, b.Taken)
		preds = append(preds, Prediction{b.PC, p, b.Taken, counterNames[counters[b.PC]]})
		if p == b.Taken {
			correct++
		}
	}
	return preds, correct
}

func step(c int, taken bool) int {
	if taken {
		return min(c+1, 3)
	}
	return max(c-1, 0)
}

// Gshare predicts with 2-bit counters indexed by the branch address XOR the
// global history, the outcomes of the last historyBits branches (Scott
// McFarling, 1993). The same branch gets a different counter for each
// recent pattern, so it learns patterns a single counter cannot: a loop of
// four is predicted perfectly once the history covers it.
//
// Time: O(n) for n branches.
// Space: O(2^historyBits) counters.
func Gshare(trace []Branch, historyBits int) (preds []Prediction, correct int, err error) {
	if historyBits < 1 || historyBits > 12 {
		return nil, 0, fmt.Errorf("%w: 1 to 12 history bits", ErrInput)
	}
	mask := 1<<historyBits - 1
	table := make([]int, 1<<historyBits)
	for i := range table {
		table[i] = 1
	}
	history := 0
	for _, b := range trace {
		i := (b.PC ^ history) & mask
		p := table[i] >= 2
		table[i] = step(table[i], b.Taken)
		bit := 0
		if b.Taken {
			bit = 1
		}
		history = (history<<1 | bit) & mask
		preds = append(preds, Prediction{b.PC, p, b.Taken, fmt.Sprintf("counter %d is %s; history %0*b", i, counterNames[table[i]], historyBits, history)})
		if p == b.Taken {
			correct++
		}
	}
	return preds, correct, nil
}
