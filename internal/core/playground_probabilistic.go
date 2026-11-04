package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type bloomP struct{ f *probabilistic.BloomFilter }

func (p *bloomP) Snapshot() trace.Snapshot { return p.f.Snapshot() }

func (p *bloomP) apply(op Op, _ trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "check" {
		return "", unknownOp(op, "add", "check")
	}
	w := strings.TrimSpace(op.Value)
	if w == "" {
		return "", inputError(op.Name + " needs a word")
	}
	bits := fmt.Sprint(p.f.Positions(w))
	if op.Name == "add" {
		p.f.Add(w)
		return fmt.Sprintf("added %q: set bits %s (false-positive rate now %.3f)", w, bits, p.f.FalsePositiveRate()), nil
	}
	if p.f.MightContain(w) {
		return fmt.Sprintf("%q might be in the set: bits %s are all set", w, bits), nil
	}
	return fmt.Sprintf("%q is definitely not in the set: one of bits %s is clear", w, bits), nil
}

type countMinP struct{ s *probabilistic.CountMinSketch }

func (p *countMinP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *countMinP) apply(op Op, _ trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "estimate" {
		return "", unknownOp(op, "add", "estimate")
	}
	w, err := opWord(op)
	if err != nil {
		return "", err
	}
	cols := p.s.Columns(w)
	if op.Name == "add" {
		p.s.Add(w, 1)
		return fmt.Sprintf("added %q: one counter per row, columns %v", w, cols), nil
	}
	return fmt.Sprintf("%q seen at most %d times: the smallest of its counters in columns %v", w, p.s.Estimate(w), cols), nil
}

type hllP struct{ h *probabilistic.HyperLogLog }

func (p *hllP) Snapshot() trace.Snapshot { return p.h.Snapshot() }

func (p *hllP) apply(op Op, _ trace.Recorder) (string, error) {
	switch op.Name {
	case "add":
		w, err := opWord(op)
		if err != nil {
			return "", err
		}
		reg, rank := p.h.Add(w)
		return fmt.Sprintf("added %q: register %d now %d", w, reg, rank), nil
	case "spread":
		n, err := opInt(op)
		if err != nil || n < 1 || n > 1_000_000 {
			return "", inputError("spread needs a count from 1 to 1000000")
		}
		for i := 0; i < n; i++ {
			p.h.Add("spread-" + strconv.Itoa(i))
		}
		return fmt.Sprintf("added %d distinct items", n), nil
	case "count":
		return fmt.Sprintf("about %d distinct items (standard error %.1f%%)", p.h.Count(), 100*p.h.StandardError()), nil
	}
	return "", unknownOp(op, "add", "count", "spread")
}

type cuckooP struct{ f *probabilistic.CuckooFilter }

func (p *cuckooP) Snapshot() trace.Snapshot { return p.f.Snapshot() }

func (p *cuckooP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "check" && op.Name != "remove" {
		return "", unknownOp(op, "add", "check", "remove")
	}
	w, err := opWord(op)
	if err != nil {
		return "", err
	}
	fp, i1, i2 := p.f.Locate(w)
	where := fmt.Sprintf("fingerprint %02x, buckets %d and %d", fp, i1, i2)
	switch op.Name {
	case "add":
		if !p.f.Add(w) {
			return fmt.Sprintf("the filter is full: %q not added (%s)", w, where), nil
		}
		pushed(r, int(fp))
		return fmt.Sprintf("added %q (%s), load %.2f", w, where, p.f.LoadFactor()), nil
	case "remove":
		if !p.f.Remove(w) {
			return fmt.Sprintf("%q was not there (%s)", w, where), nil
		}
		popped(r, int(fp))
		return fmt.Sprintf("removed %q (%s)", w, where), nil
	}
	if p.f.Contains(w) {
		return fmt.Sprintf("%q might be in the set (%s)", w, where), nil
	}
	return fmt.Sprintf("%q is definitely not in the set (%s)", w, where), nil
}
