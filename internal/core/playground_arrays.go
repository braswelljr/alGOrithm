package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// jsArrayP is the js-array playground: the array, and the roles to draw
// after the last operation.
type jsArrayP struct {
	a     *arrays.Array[int]
	roles map[int]string
}

func (p *jsArrayP) Snapshot() trace.Snapshot { return arrayRow(p.a.Values(), p.roles) }

func (p *jsArrayP) apply(op Op, r trace.Recorder) (string, error) {
	p.a.SetRecorder(r)
	defer p.a.SetRecorder(nil)
	p.roles = nil
	n := p.a.Len()
	args, err := fieldInts(op.Value, op.Name)
	need := func(k int) error {
		if err != nil {
			return err
		}
		if len(args) < k {
			return inputError(fmt.Sprintf("%s needs %d number(s)", op.Name, k))
		}
		return nil
	}
	list := func(s []int) string { return "[" + strings.ReplaceAll(joinInts(s), " ", ", ") + "]" }
	switch op.Name {
	case "push":
		return p.opPush(args, n, need)
	case "pop", "shift":
		return p.opPop(op)
	case "unshift":
		return p.opUnshift(args, need)
	case "splice":
		return p.opSplice(args, n, need, list)
	case "slice":
		return p.opSlice(args, n, need, list)
	case "concat":
		return p.opConcat(args, list, err)
	case "at":
		return p.opAt(args, n, need)
	case "with":
		return p.opWith(args, need, list)
	case "fill":
		return p.opFill(args, n, need)
	case "copy-within":
		return p.opCopyWithin(args, n, need)
	case "reverse":
		return p.opReverse()
	case "sort":
		return p.opSort()
	case "sort-default":
		return p.opSortDefault()
	case "index-of", "last-index-of", "includes":
		return p.opIndexOf(op, args, need)
	case "join":
		return p.opJoin(op)
	}
	return "", unknownOp(op, "push", "pop", "shift", "unshift", "splice", "slice", "concat", "at", "with", "fill", "copy-within", "reverse", "sort", "sort-default", "index-of", "last-index-of", "includes", "join")
}

func (p *jsArrayP) opPush(args []int, n int, need func(int) error) (string, error) {
	if err := need(1); err != nil {
		return "", err
	}
	l := p.a.Push(args...)
	p.roles = spanRoles(n, l, "new")
	return fmt.Sprintf("push returned the new length %d", l), nil
}

func (p *jsArrayP) opPop(op Op) (string, error) {
	pop := p.a.Pop
	if op.Name == "shift" {
		pop = p.a.Shift
	}
	v, ok := pop()
	if !ok {
		return op.Name + " on an empty array returns undefined", nil
	}
	return fmt.Sprintf("%s returned %d", op.Name, v), nil
}

func (p *jsArrayP) opUnshift(args []int, need func(int) error) (string, error) {
	if err := need(1); err != nil {
		return "", err
	}
	l := p.a.Unshift(args...)
	p.roles = spanRoles(0, len(args), "new")
	return fmt.Sprintf("unshift returned the new length %d (every element moved)", l), nil
}

func (p *jsArrayP) opSplice(args []int, n int, need func(int) error, list func([]int) string) (string, error) {
	if err := need(2); err != nil {
		return "", err
	}
	start := args[0]
	removed := p.a.Splice(start, args[1], args[2:]...)
	s := start
	if s < 0 {
		s = max(n+s, 0)
	}
	s = min(s, n)
	p.roles = spanRoles(s, s+len(args)-2, "new")
	return fmt.Sprintf("splice removed %s", list(removed)), nil
}

func (p *jsArrayP) opSlice(args []int, n int, need func(int) error, list func([]int) string) (string, error) {
	if err := need(2); err != nil {
		return "", err
	}
	got := p.a.Slice(args[0], args[1]).Values()
	s, e := args[0], args[1]
	if s < 0 {
		s = max(n+s, 0)
	}
	if e < 0 {
		e = max(n+e, 0)
	}
	p.roles = spanRoles(min(s, n), min(e, n), "range")
	return fmt.Sprintf("slice returned the copy %s (the array is unchanged)", list(got)), nil
}

func (p *jsArrayP) opConcat(args []int, list func([]int) string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("concat returned the new array %s (this one is unchanged)", list(p.a.Concat(args).Values())), nil
}

func (p *jsArrayP) opAt(args []int, n int, need func(int) error) (string, error) {
	if err := need(1); err != nil {
		return "", err
	}
	v, ok := p.a.At(args[0])
	if !ok {
		return fmt.Sprintf("at(%d) returns undefined", args[0]), nil
	}
	i := args[0]
	if i < 0 {
		i += n
	}
	p.roles = map[int]string{i: "range"}
	return fmt.Sprintf("at(%d) is %d", args[0], v), nil
}

func (p *jsArrayP) opWith(args []int, need func(int) error, list func([]int) string) (string, error) {
	if err := need(2); err != nil {
		return "", err
	}
	w, ok := p.a.With(args[0], args[1])
	if !ok {
		return fmt.Sprintf("with(%d, ...) throws a RangeError", args[0]), nil
	}
	return fmt.Sprintf("with returned the copy %s (this one is unchanged)", list(w.Values())), nil
}

func (p *jsArrayP) opFill(args []int, n int, need func(int) error) (string, error) {
	if err := need(1); err != nil {
		return "", err
	}
	s, e := 0, n
	if len(args) > 1 {
		s = args[1]
	}
	if len(args) > 2 {
		e = args[2]
	}
	p.a.Fill(args[0], s, e)
	if s < 0 {
		s = max(n+s, 0)
	}
	if e < 0 {
		e = max(n+e, 0)
	}
	p.roles = spanRoles(min(s, n), min(e, n), "new")
	return fmt.Sprintf("filled with %d", args[0]), nil
}

func (p *jsArrayP) opCopyWithin(args []int, n int, need func(int) error) (string, error) {
	if err := need(2); err != nil {
		return "", err
	}
	e := n
	if len(args) > 2 {
		e = args[2]
	}
	p.a.CopyWithin(args[0], args[1], e)
	return "copied within, the length unchanged", nil
}

func (p *jsArrayP) opReverse() (string, error) {
	p.a.Reverse()
	return "reversed in place", nil
}

func (p *jsArrayP) opSort() (string, error) {
	p.a.Sort(func(x, y int) bool { return x < y })
	return "sorted numerically, with sort((a, b) => a - b)", nil
}

func (p *jsArrayP) opSortDefault() (string, error) {
	p.a.SortDefault()
	return "sorted the JavaScript default way: by string form", nil
}

func (p *jsArrayP) opIndexOf(op Op, args []int, need func(int) error) (string, error) {
	if err := need(1); err != nil {
		return "", err
	}
	switch op.Name {
	case "index-of":
		return fmt.Sprintf("indexOf(%d) = %d", args[0], arrays.IndexOf(p.a, args[0])), nil
	case "last-index-of":
		return fmt.Sprintf("lastIndexOf(%d) = %d", args[0], arrays.LastIndexOf(p.a, args[0])), nil
	}
	return fmt.Sprintf("includes(%d) = %v", args[0], arrays.Includes(p.a, args[0])), nil
}

func (p *jsArrayP) opJoin(op Op) (string, error) {
	sep := op.Value
	if sep == "" {
		sep = ","
	}
	return fmt.Sprintf("join(%q) = %q", sep, p.a.Join(sep)), nil
}

type dynamicP struct{ d *arrays.Dynamic }

func (p *dynamicP) Snapshot() trace.Snapshot { return p.d.Snapshot() }

func (p *dynamicP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "append":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		before, copies := p.d.Cap(), p.d.Copies
		p.d.Append(v)
		pushed(r, v)
		if p.d.Cap() != before {
			return fmt.Sprintf("appended %d: full, so grew %d to %d slots and copied %d values", v, before, p.d.Cap(), p.d.Copies-copies), nil
		}
		return fmt.Sprintf("appended %d into a spare slot", v), nil
	case "pop":
		v, ok := p.d.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	case "insert", "set":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if op.Name == "set" {
			if !p.d.Set(a[0], a[1]) {
				return fmt.Sprintf("index %d is out of range", a[0]), nil
			}
			return fmt.Sprintf("set index %d to %d in O(1)", a[0], a[1]), nil
		}
		if !p.d.Insert(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range 0..%d", a[0], p.d.Len()), nil
		}
		return fmt.Sprintf("inserted %d at %d, shifting %d values right", a[1], a[0], p.d.Len()-1-a[0]), nil
	case "delete", "get":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "get" {
			if v, ok := p.d.Get(i); ok {
				return fmt.Sprintf("index %d is %d, read in O(1)", i, v), nil
			}
			return fmt.Sprintf("index %d is out of range", i), nil
		}
		v, ok := p.d.Delete(i)
		if !ok {
			return fmt.Sprintf("index %d is out of range", i), nil
		}
		return fmt.Sprintf("deleted %d at %d, shifting %d values left", v, i, p.d.Len()-i), nil
	}
	return "", unknownOp(op, "append", "pop", "insert", "delete", "get", "set")
}
