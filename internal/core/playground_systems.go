package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/databases"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type bplusP struct{ t *databases.BPlusTree }

func (p *bplusP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *bplusP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if !p.t.Insert(k, k) {
			return fmt.Sprintf("%d was already there", k), nil
		}
		return fmt.Sprintf("inserted %d; height %d", k, p.t.Height()), nil
	case "delete":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if !p.t.Delete(k) {
			return fmt.Sprintf("%d is not there", k), nil
		}
		return fmt.Sprintf("deleted %d; height %d", k, p.t.Height()), nil
	case "search":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if _, ok := p.t.Search(k, r); ok {
			return fmt.Sprintf("found %d", k), nil
		}
		return fmt.Sprintf("%d not found", k), nil
	case "range":
		lohi, err := opPair(op)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("keys %d to %d: [%s]", lohi[0], lohi[1], joinInts(p.t.Range(lohi[0], lohi[1], r))), nil
	}
	return "", unknownOp(op, "insert", "delete", "search", "range")
}

type lsmP struct{ t *databases.LSMTree }

func (p *lsmP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *lsmP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		f := strings.Fields(op.Value)
		nums := make([]int, len(f))
		for i, s := range f {
			v, err := strconv.Atoi(s)
			if err != nil || len(f) > 2 {
				return "", inputError(fmt.Sprintf("write put KEY or put KEY VALUE, not %q", op.Value))
			}
			nums[i] = v
		}
		if len(nums) == 0 {
			return "", inputError("put needs a key")
		}
		k, v := nums[0], nums[0]
		if len(nums) == 2 {
			v = nums[1]
		}
		p.t.Put(k, v)
		return fmt.Sprintf("put %d = %d; %d runs", k, v, p.t.Runs()), nil
	case "delete":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.t.Delete(k)
		return fmt.Sprintf("wrote a tombstone for %d", k), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		before := p.t.Stats
		v, ok := p.t.Get(k, r)
		searched, skipped := p.t.Stats.RunsSearched-before.RunsSearched, p.t.Stats.FilterSkips-before.FilterSkips
		if !ok {
			return fmt.Sprintf("%d not found (%d runs searched, %d skipped by Bloom filters)", k, searched, skipped), nil
		}
		return fmt.Sprintf("%d = %d (%d runs searched, %d skipped by Bloom filters)", k, v, searched, skipped), nil
	case "flush":
		p.t.Flush()
		return fmt.Sprintf("flushed; %d runs", p.t.Runs()), nil
	case "compact":
		p.t.Compact()
		return fmt.Sprintf("compacted into %d run", p.t.Runs()), nil
	}
	return "", unknownOp(op, "put", "delete", "get", "flush", "compact")
}

type pageP struct{ p *databases.SlottedPage }

func (p *pageP) Snapshot() trace.Snapshot { return p.p.Snapshot() }

func (p *pageP) apply(op Op, _ trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		id, err := p.p.Insert(strings.TrimSpace(op.Value))
		if err != nil {
			return "", wrap(err)
		}
		return fmt.Sprintf("slot %d; %d bytes free", id, p.p.Free()), nil
	case "delete":
		id, err := opInt(op)
		if err != nil {
			return "", err
		}
		if !p.p.Delete(id) {
			return fmt.Sprintf("slot %d is empty", id), nil
		}
		return fmt.Sprintf("freed slot %d; %d bytes in holes", id, p.p.Holes()), nil
	case "get":
		id, err := opInt(op)
		if err != nil {
			return "", err
		}
		if rec, ok := p.p.Get(id); ok {
			return fmt.Sprintf("slot %d holds %q", id, rec), nil
		}
		return fmt.Sprintf("slot %d is empty", id), nil
	case "compact":
		p.p.Compact()
		return fmt.Sprintf("compacted; %d bytes free", p.p.Free()), nil
	}
	return "", unknownOp(op, "insert", "delete", "get", "compact")
}

type poolP struct{ b *databases.BufferPool }

func (p *poolP) Snapshot() trace.Snapshot { return p.b.Snapshot() }

func (p *poolP) apply(op Op, r trace.Recorder) (string, error) {
	page, err := opInt(op)
	if err != nil {
		return "", err
	}
	var pin, dirty bool
	switch op.Name {
	case "read":
	case "write":
		dirty = true
	case "pin":
		pin = true
	case "unpin":
		if !p.b.Unpin(page) {
			return fmt.Sprintf("page %d is not pinned", page), nil
		}
		return fmt.Sprintf("unpinned page %d", page), nil
	default:
		return "", unknownOp(op, "read", "write", "pin", "unpin")
	}
	before := p.b.Stats
	hit, err := p.b.Fetch(page, pin, dirty, r)
	if err != nil {
		return "", wrap(err)
	}
	if hit {
		return fmt.Sprintf("hit: page %d is in memory", page), nil
	}
	res := fmt.Sprintf("miss: read page %d from disk", page)
	if p.b.Stats.Evictions > before.Evictions {
		res += ", evicting a page"
		if p.b.Stats.WriteBacks > before.WriteBacks {
			res += " (dirty, so written back first)"
		}
	}
	return res, nil
}

type extHashP struct{ h *databases.ExtendibleHash }

func (p *extHashP) Snapshot() trace.Snapshot { return p.h.Snapshot() }

func (p *extHashP) apply(op Op, r trace.Recorder) (string, error) {
	k, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		depth := p.h.GlobalDepth()
		added, err := p.h.Insert(k)
		if err != nil {
			return "", wrap(err)
		}
		if !added {
			return fmt.Sprintf("%d was already there", k), nil
		}
		if p.h.GlobalDepth() > depth {
			return fmt.Sprintf("inserted %d; the directory doubled to global depth %d", k, p.h.GlobalDepth()), nil
		}
		return fmt.Sprintf("inserted %d", k), nil
	case "delete":
		if !p.h.Delete(k) {
			return fmt.Sprintf("%d is not there", k), nil
		}
		return fmt.Sprintf("deleted %d", k), nil
	case "search":
		return found(k, p.h.Search(k, r)), nil
	}
	return "", unknownOp(op, "insert", "delete", "search")
}

type linHashP struct{ h *databases.LinearHash }

func (p *linHashP) Snapshot() trace.Snapshot { return p.h.Snapshot() }

func (p *linHashP) apply(op Op, r trace.Recorder) (string, error) {
	k, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		splits := p.h.Splits
		if !p.h.Insert(k) {
			return fmt.Sprintf("%d was already there", k), nil
		}
		if p.h.Splits > splits {
			return fmt.Sprintf("inserted %d; a bucket overflowed, so the next bucket in line split", k), nil
		}
		return fmt.Sprintf("inserted %d", k), nil
	case "delete":
		if !p.h.Delete(k) {
			return fmt.Sprintf("%d is not there", k), nil
		}
		return fmt.Sprintf("deleted %d", k), nil
	case "search":
		return found(k, p.h.Search(k, r)), nil
	}
	return "", unknownOp(op, "insert", "delete", "search")
}
