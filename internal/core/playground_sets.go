package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type bitsetP struct{ b *sets.Bitset }

func (p *bitsetP) Snapshot() trace.Snapshot { return p.b.Snapshot() }

func (p *bitsetP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "count":
		return fmt.Sprintf("%d members (popcount)", p.b.Count()), nil
	case "members":
		return "members: " + joinInts(p.b.Members()), nil
	case "set", "clear", "test":
		n, err := opInt(op)
		if err != nil {
			return "", err
		}
		if n < 0 || n > 4095 {
			return "", inputError("members are 0 to 4095 here")
		}
		switch op.Name {
		case "set":
			if !p.b.Set(n) {
				return fmt.Sprintf("bit %d was already set", n), nil
			}
			pushed(r, n)
			return fmt.Sprintf("set bit %d (word %d, bit %d)", n, n/64, n%64), nil
		case "clear":
			if !p.b.Clear(n) {
				return fmt.Sprintf("bit %d was not set", n), nil
			}
			popped(r, n)
			return fmt.Sprintf("cleared bit %d", n), nil
		}
		return fmt.Sprintf("bit %d is set: %v", n, p.b.Test(n)), nil
	}
	return "", unknownOp(op, "set", "clear", "test", "count", "members")
}

type multisetP struct{ m *sets.Multiset[int] }

func (p *multisetP) Snapshot() trace.Snapshot { return p.m.Snapshot() }

func (p *multisetP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "remove" && op.Name != "count" {
		return "", unknownOp(op, "add", "remove", "count")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		pushed(r, v)
		return fmt.Sprintf("added %d (now %d of it)", v, p.m.Add(v)), nil
	case "remove":
		if !p.m.Remove(v) {
			return fmt.Sprintf("%d is not in the bag", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed one %d (%d left)", v, p.m.Count(v)), nil
	}
	return fmt.Sprintf("%d occurs %d times", v, p.m.Count(v)), nil
}

type orderedP struct{ s *sets.OrderedSet[int] }

func (p *orderedP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *orderedP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "range" {
		lh, err := opPair(op)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("members in [%d, %d]: %s", lh[0], lh[1], joinInts(p.s.Range(lh[0], lh[1]))), nil
	}
	switch op.Name {
	case "add", "remove", "contains", "rank", "floor", "ceiling":
	default:
		return "", unknownOp(op, "add", "remove", "contains", "rank", "floor", "ceiling", "range")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	near := func(name string, x int, ok bool) string {
		if !ok {
			return fmt.Sprintf("no %s of %d", name, v)
		}
		return fmt.Sprintf("%s of %d is %d", name, v, x)
	}
	switch op.Name {
	case "add":
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already in the set", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d at position %d", v, p.s.Rank(v)), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not in the set", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	case "contains":
		return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
	case "rank":
		return fmt.Sprintf("%d members are smaller than %d", p.s.Rank(v), v), nil
	case "floor":
		x, ok := p.s.Floor(v)
		return near("floor", x, ok), nil
	}
	x, ok := p.s.Ceiling(v)
	return near("ceiling", x, ok), nil
}

type sparseSetP struct{ s *sets.SparseSet }

func (p *sparseSetP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *sparseSetP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "clear" {
		n := p.s.Len()
		p.s.Clear()
		return fmt.Sprintf("cleared %d members in O(1): only the length was reset", n), nil
	}
	if op.Name != "add" && op.Name != "remove" && op.Name != "check" {
		return "", unknownOp(op, "add", "remove", "check", "clear")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		if v < 0 || v >= p.s.Universe() {
			return "", inputError(fmt.Sprintf("%d is outside 0..%d", v, p.s.Universe()-1))
		}
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already a member", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d at dense[%d]", v, p.s.Len()-1), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not a member", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d, moving the last member into its slot", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
}

type roaringP struct{ b *sets.Roaring }

func (p *roaringP) Snapshot() trace.Snapshot { return p.b.Snapshot() }

func (p *roaringP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "fill" {
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if a[0] < 0 || a[1] < a[0] || a[1]-a[0] > 100_000 || a[1] > 1<<32-1 {
			return "", inputError("fill needs LO <= HI, at most 100000 apart, within 0..4294967295")
		}
		for v := a[0]; v <= a[1]; v++ {
			p.b.Add(uint32(v))
		}
		return fmt.Sprintf("added %d..%d: %d members in %d bytes", a[0], a[1], p.b.Len(), p.b.Bytes()), nil
	}
	if op.Name != "add" && op.Name != "remove" && op.Name != "check" {
		return "", unknownOp(op, "add", "remove", "check", "fill")
	}
	v, err := opInt(op)
	if err != nil || v < 0 || v > 1<<32-1 {
		return "", inputError(op.Name + " needs a value in 0..4294967295")
	}
	switch op.Name {
	case "add":
		if !p.b.Add(uint32(v)) {
			return fmt.Sprintf("%d is already a member", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d to the group of high bits %d", v, v>>16), nil
	case "remove":
		if !p.b.Remove(uint32(v)) {
			return fmt.Sprintf("%d is not a member", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.b.Contains(uint32(v))), nil
}
