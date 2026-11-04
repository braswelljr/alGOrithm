package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/memory"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type buddyP struct {
	b *memory.Buddy
	n int // names given to unnamed allocations
}

func (p *buddyP) Snapshot() trace.Snapshot { return p.b.Snapshot() }

func (p *buddyP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	switch op.Name {
	case "alloc":
		if len(f) == 1 { // "alloc SIZE": a name is made up
			p.n++
			f = []string{"P" + strconv.Itoa(p.n), f[0]}
		}
		if len(f) != 2 {
			return "", inputError("alloc needs NAME SIZE (or SIZE)")
		}
		size, err := strconv.Atoi(f[1])
		if err != nil {
			return "", inputError(fmt.Sprintf("alloc needs a whole-number size, not %q", f[1]))
		}
		off, err := p.b.Alloc(f[0], size)
		if err != nil {
			return fmt.Sprintf("%s (%d bytes): %v", f[0], size, err), nil
		}
		r.Record(trace.Event{Kind: trace.Push, Values: []int{off}})
		return fmt.Sprintf("%s got %d bytes at %d (internal fragmentation now %d)", f[0], size, off, p.b.Fragmentation()), nil
	case "free":
		if len(f) != 1 {
			return "", inputError("free needs NAME")
		}
		if err := p.b.Free(f[0]); err != nil {
			return err.Error(), nil //nolint:nilerr // freeing a bad block is a step to show, not a failure of the run
		}
		r.Record(trace.Event{Kind: trace.Pop})
		return fmt.Sprintf("freed %s, merging buddies where both are free", f[0]), nil
	}
	return "", unknownOp(op, "alloc", "free")
}
