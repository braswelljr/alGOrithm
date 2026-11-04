package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/ranges"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type segmentP struct {
	rangeP
	t *ranges.SegmentTree
}

func (p *segmentP) Snapshot() trace.Snapshot { return p.tree().Snapshot() }

func (p *segmentP) tree() *ranges.SegmentTree {
	if p.t == nil || p.t.Len() != len(p.values) {
		p.t = ranges.NewSegmentTree(p.values)
	}
	return p.t
}

func (p *segmentP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "update":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if !p.tree().Update(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range", a[0]), nil
		}
		p.values[a[0]] = a[1]
		return fmt.Sprintf("set index %d to %d, updating its O(log n) ancestors", a[0], a[1]), nil
	case "sum", "min":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		f := p.tree().Sum
		if op.Name == "min" {
			f = p.tree().Min
		}
		v, ok := f(a[0], a[1])
		return rangeAnswer(op.Name, a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "update", "sum", "min")
}

type fenwickP struct {
	rangeP
	t *ranges.FenwickTree
}

func (p *fenwickP) Snapshot() trace.Snapshot { return p.tree().Snapshot() }

func (p *fenwickP) tree() *ranges.FenwickTree {
	if p.t == nil || p.t.Len() != len(p.values) {
		p.t = ranges.NewFenwickTree(p.values)
	}
	return p.t
}

func (p *fenwickP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "add":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if !p.tree().Add(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range", a[0]), nil
		}
		p.values[a[0]] += a[1]
		return fmt.Sprintf("added %d at index %d", a[1], a[0]), nil
	case "prefix":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		v, ok := p.tree().PrefixSum(i)
		return rangeAnswer("sum", 0, i, v, ok), nil
	case "range":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := p.tree().RangeSum(a[0], a[1])
		return rangeAnswer("sum", a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "add", "prefix", "range")
}

type sparseP struct{ rangeP }

func (p *sparseP) Snapshot() trace.Snapshot { return ranges.NewSparseTable(p.values).Snapshot() }

func (p *sparseP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "min":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := ranges.NewSparseTable(p.values).Min(a[0], a[1])
		return rangeAnswer("min", a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "min")
}

type disjointP struct {
	rangeP
	op int
}

func (p *disjointP) table() *ranges.DisjointSparseTable {
	return ranges.NewDisjointSparseTable(p.values, disjointOps[p.op].f)
}

func (p *disjointP) Snapshot() trace.Snapshot { return p.table().Snapshot() }

func (p *disjointP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "query":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := p.table().Query(a[0], a[1])
		return rangeAnswer(disjointOps[p.op].name, a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "query")
}
