package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type gapP struct{ g *buffers.GapBuffer }

func (p *gapP) Snapshot() trace.Snapshot { return p.g.Snapshot() }

func (p *gapP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "type":
		if op.Value == "" {
			return "", inputError("type needs some text")
		}
		p.g.Insert(op.Value)
		pushed(r, len(op.Value))
		return fmt.Sprintf("typed %q: %q, cursor at %d, gap of %d", op.Value, p.g.String(), p.g.Cursor(), p.g.Gap()), nil
	case "move":
		pos, err := opInt(op)
		if err != nil {
			return "", err
		}
		moved := p.g.Move(pos)
		return fmt.Sprintf("cursor at %d, shifting %d bytes across the gap", p.g.Cursor(), moved), nil
	case "backspace", "delete":
		n, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "backspace" {
			n = p.g.Backspace(n)
		} else {
			n = p.g.Delete(n)
		}
		popped(r, n)
		return fmt.Sprintf("deleted %d: %q, gap of %d", n, p.g.String(), p.g.Gap()), nil
	}
	return "", unknownOp(op, "type", "move", "backspace", "delete")
}

type ropeP struct{ r *buffers.Rope }

func (p *ropeP) Snapshot() trace.Snapshot { return p.r.Snapshot() }

func (p *ropeP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		pos, text, ok := strings.Cut(op.Value, " ")
		i, err := strconv.Atoi(pos)
		if !ok || err != nil || text == "" {
			return "", inputError("insert needs a position and some text")
		}
		p.r.Insert(i, text)
		pushed(r, len(text))
		return fmt.Sprintf("inserted %q: %q, depth %d", text, p.r.String(), p.r.Depth()), nil
	case "delete":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		n := p.r.Delete(a[0], a[1])
		popped(r, n)
		return fmt.Sprintf("deleted %d bytes: %q", n, p.r.String()), nil
	case "index":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		b, ok := p.r.Index(i)
		if !ok {
			return fmt.Sprintf("%d is past the end", i), nil
		}
		return fmt.Sprintf("byte %d is %q, found by walking one path", i, b), nil
	case "rebalance":
		p.r.Rebalance()
		return fmt.Sprintf("rebuilt balanced, depth %d", p.r.Depth()), nil
	}
	return "", unknownOp(op, "insert", "delete", "index", "rebalance")
}
