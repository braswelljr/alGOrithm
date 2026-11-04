package databases

import (
	"fmt"
	"strconv"
	"strings"
)

// Action is what one operation of a transaction does.
type Action uint8

// The actions of a schedule.
const (
	Read Action = iota
	Write
	Commit
	Abort
	Begin
)

var actionNames = [...]string{"R", "W", "C", "A", "B"}

// String is the action's one-letter name: R, W, C, A or B.
func (a Action) String() string {
	if int(a) < len(actionNames) {
		return actionNames[a]
	}
	return "?"
}

// TxOp is one operation of a schedule: transaction Txn reads or writes Item
// (a write may carry a Value), or begins, commits or aborts.
type TxOp struct {
	Txn    string
	Action Action
	Item   string
	Value  int
}

// String writes the operation in textbook notation: "R1(x)", "W2(y)", "C1".
func (o TxOp) String() string {
	n := strings.TrimPrefix(o.Txn, "T")
	if o.Action == Read || o.Action == Write {
		return fmt.Sprintf("%s%s(%s)", o.Action, n, o.Item)
	}
	return o.Action.String() + n
}

// ParseSchedule reads a schedule in either of two notations, mixed freely:
// the textbook one, "r1(x) w2(x) c1 a2", or one operation per comma-separated
// part, "T1 read x, T2 write x 5, T1 commit". Actions are read (r), write (w),
// commit (c), abort (a) and begin (b); a write may give the value written.
//
// Time: O(n) in the length of s.
func ParseSchedule(s string) ([]TxOp, error) {
	var ops []TxOp
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		if isTxn(f[0]) {
			op, err := parseLong(f)
			if err != nil {
				return nil, err
			}
			ops = append(ops, op)
			continue
		}
		for _, tok := range f {
			op, err := parseShort(tok)
			if err != nil {
				return nil, err
			}
			ops = append(ops, op)
		}
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("%w: an empty schedule (write it like \"r1(x) w2(x) c1 c2\")", ErrInput)
	}
	return ops, nil
}

func isTxn(s string) bool {
	if len(s) < 2 || (s[0] != 'T' && s[0] != 't') {
		return false
	}
	_, err := strconv.Atoi(s[1:])
	return err == nil
}

func action(s string) (Action, bool) {
	switch strings.ToLower(s) {
	case "r", "read":
		return Read, true
	case "w", "write":
		return Write, true
	case "c", "commit":
		return Commit, true
	case "a", "abort", "rollback":
		return Abort, true
	case "b", "begin", "start":
		return Begin, true
	}
	return 0, false
}

func parseLong(f []string) (TxOp, error) {
	if len(f) < 2 {
		return TxOp{}, fmt.Errorf("%w: %q needs an action (read, write, commit, abort, begin)", ErrInput, strings.Join(f, " "))
	}
	a, ok := action(f[1])
	if !ok {
		return TxOp{}, fmt.Errorf("%w: unknown action %q in %q", ErrInput, f[1], strings.Join(f, " "))
	}
	op := TxOp{Txn: "T" + f[0][1:], Action: a}
	if a == Read || a == Write {
		if len(f) < 3 {
			return TxOp{}, fmt.Errorf("%w: %q needs an item", ErrInput, strings.Join(f, " "))
		}
		op.Item = f[2]
		if a == Write && len(f) > 3 {
			v, err := strconv.Atoi(f[3])
			if err != nil {
				return TxOp{}, fmt.Errorf("%w: the value in %q is not a whole number", ErrInput, strings.Join(f, " "))
			}
			op.Value = v
		}
	}
	return op, nil
}

// parseShort reads "r1(x)", "w2(y)", "c1", "a3" or "b1".
func parseShort(tok string) (TxOp, error) {
	bad := fmt.Errorf("%w: cannot read %q (write r1(x), w2(x), c1, a1 or b1)", ErrInput, tok)
	if len(tok) < 2 {
		return TxOp{}, bad
	}
	a, ok := action(tok[:1])
	if !ok {
		return TxOp{}, bad
	}
	rest := tok[1:]
	num, item := rest, ""
	if i := strings.IndexByte(rest, '('); i >= 0 {
		if !strings.HasSuffix(rest, ")") {
			return TxOp{}, bad
		}
		num, item = rest[:i], rest[i+1:len(rest)-1]
	}
	if _, err := strconv.Atoi(num); err != nil {
		return TxOp{}, bad
	}
	if (a == Read || a == Write) != (item != "") {
		return TxOp{}, bad
	}
	return TxOp{Txn: "T" + num, Action: a, Item: item}, nil
}

// txnOrder lists the transactions of ops in the order they first appear.
func txnOrder(ops []TxOp) []string {
	seen := map[string]bool{}
	var out []string
	for _, o := range ops {
		if !seen[o.Txn] {
			seen[o.Txn] = true
			out = append(out, o.Txn)
		}
	}
	return out
}
