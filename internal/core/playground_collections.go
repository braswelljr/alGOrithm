package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/collections"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type esMapP struct {
	m *collections.JSMap[string, string]
}

func (p *esMapP) Snapshot() trace.Snapshot { return p.m.Snapshot() }

func (p *esMapP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	need := func(n int) error {
		if len(f) < n {
			return inputError(fmt.Sprintf("%s needs %d values", op.Name, n))
		}
		return nil
	}
	switch op.Name {
	case "set":
		if len(f) == 1 { // a starting value: set it to itself
			f = append(f, f[0])
		}
		if err := need(2); err != nil {
			return "", err
		}
		had := p.m.Has(f[0])
		p.m.Set(f[0], strings.Join(f[1:], " "))
		if had {
			return fmt.Sprintf("updated %s in place", f[0]), nil
		}
		return fmt.Sprintf("set %s at the end", f[0]), nil
	case "get", "has", "delete":
		if err := need(1); err != nil {
			return "", err
		}
		switch op.Name {
		case "get":
			if v, ok := p.m.Get(f[0]); ok {
				return fmt.Sprintf("%s is %s", f[0], v), nil
			}
			return f[0] + " is undefined", nil
		case "has":
			return fmt.Sprintf("has %s: %v", f[0], p.m.Has(f[0])), nil
		}
		return fmt.Sprintf("deleted %s: %v", f[0], p.m.Delete(f[0])), nil
	case "clear":
		p.m.Clear()
		return "cleared", nil
	}
	return "", unknownOp(op, "set", "get", "has", "delete", "clear")
}

type esSetP struct{ s *collections.JSSet[string] }

func (p *esSetP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *esSetP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	other := collections.NewJSSet(f...)
	values := func(s *collections.JSSet[string]) string { return "{" + strings.Join(s.Values(), ", ") + "}" }
	switch op.Name {
	case "add":
		for _, v := range f {
			p.s.Add(v)
		}
		return "added " + strings.Join(f, " ") + ": now " + values(p.s), nil
	case "has":
		return fmt.Sprintf("has %s: %v", op.Value, p.s.Has(op.Value)), nil
	case "delete":
		return fmt.Sprintf("deleted %s: %v", op.Value, p.s.Delete(op.Value)), nil
	case "union":
		return "union " + values(other) + " = " + values(p.s.Union(other)), nil
	case "intersection":
		return "intersection " + values(other) + " = " + values(p.s.Intersection(other)), nil
	case "difference":
		return "difference " + values(other) + " = " + values(p.s.Difference(other)), nil
	case "symmetric-difference":
		return "symmetricDifference " + values(other) + " = " + values(p.s.SymmetricDifference(other)), nil
	case "is-subset-of":
		return fmt.Sprintf("isSubsetOf %s: %v", values(other), p.s.IsSubsetOf(other)), nil
	case "is-disjoint-from":
		return fmt.Sprintf("isDisjointFrom %s: %v", values(other), p.s.IsDisjointFrom(other)), nil
	}
	return "", unknownOp(op, "add", "has", "delete", "union", "intersection", "difference", "symmetric-difference", "is-subset-of", "is-disjoint-from")
}
