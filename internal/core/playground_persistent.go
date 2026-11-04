package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/persistent"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type pStackP struct {
	versionsP
	versions []persistent.Stack[int]
}

func (p *pStackP) Snapshot() trace.Snapshot { return persistent.SnapshotStacks(p.versions...) }

func (p *pStackP) apply(op Op, r trace.Recorder) (string, error) {
	from := p.versions[p.current]
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		old := p.current
		p.versions = append(p.versions, from.Push(v))
		p.current = len(p.versions) - 1
		pushed(r, v)
		return fmt.Sprintf("version %d: pushed %d onto version %d, sharing all of it", p.current, v, old), nil
	case "pop":
		v, rest, ok := from.Pop()
		if !ok {
			return fmt.Sprintf("version %d is empty", p.current), nil
		}
		old := p.current
		p.versions = append(p.versions, rest)
		p.current = len(p.versions) - 1
		popped(r, v)
		return fmt.Sprintf("version %d: popped %d from version %d, which still has it", p.current, v, old), nil
	case "switch":
		return p.switchTo(op, len(p.versions))
	}
	return "", unknownOp(op, "push", "pop", "switch")
}

type pTreeP struct {
	versionsP
	versions []persistent.Tree[int]
}

func (p *pTreeP) Snapshot() trace.Snapshot { return persistent.SnapshotTrees(p.versions...) }

func (p *pTreeP) apply(op Op, r trace.Recorder) (string, error) {
	from := p.versions[p.current]
	switch op.Name {
	case "insert", "remove":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		next := from.Insert(v)
		if op.Name == "remove" {
			next = from.Remove(v)
		}
		if next.Len() == from.Len() {
			return fmt.Sprintf("version %d is unchanged by %s %d", p.current, op.Name, v), nil
		}
		old := p.current
		p.versions = append(p.versions, next)
		p.current = len(p.versions) - 1
		if op.Name == "insert" {
			pushed(r, v)
		} else {
			popped(r, v)
		}
		return fmt.Sprintf("version %d: %s %d on version %d, sharing %d of its %d nodes", p.current, op.Name, v, old, from.SharedNodes(next), from.Len()), nil
	case "switch":
		return p.switchTo(op, len(p.versions))
	}
	return "", unknownOp(op, "insert", "remove", "switch")
}
