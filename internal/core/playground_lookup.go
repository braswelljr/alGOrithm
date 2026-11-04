package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

type lfuP struct {
	c *lookup.LFUCache[int, int]
}

func (p *lfuP) Snapshot() trace.Snapshot { return p.c.Snapshot() }

func (p *lfuP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		if f := strings.Fields(op.Value); len(f) == 1 {
			op.Value += " " + f[0]
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		pushed(r, kv[0])
		if old, ok := p.c.Put(kv[0], kv[1]); ok {
			popped(r, old)
			return fmt.Sprintf("put %d: %d, evicted %d (least frequently used)", kv[0], kv[1], old), nil
		}
		return fmt.Sprintf("put %d: %d", kv[0], kv[1]), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k); ok {
			f, _ := p.c.Frequency(k)
			return fmt.Sprintf("%d is %d (used %d times)", k, v, f), nil
		}
		return fmt.Sprintf("%d is not cached", k), nil
	}
	return "", unknownOp(op, "put", "get")
}

type setP struct{ s *lookup.Set[int] }

func (p *setP) Snapshot() trace.Snapshot { return p.s.Snapshot() }

func (p *setP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "remove" && op.Name != "contains" {
		return "", unknownOp(op, "add", "remove", "contains")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already in the set", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d", v), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not in the set", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
}

type trieP struct{ t *lookup.Trie }

func (p *trieP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *trieP) apply(op Op, _ trace.Recorder) (string, error) {
	switch op.Name {
	case "insert", "search", "prefix", "remove":
	default:
		return "", unknownOp(op, "insert", "search", "prefix", "remove")
	}
	w := strings.TrimSpace(op.Value)
	if w == "" && op.Name != "prefix" {
		return "", inputError(op.Name + " needs a word")
	}
	if len(w) > 64 {
		return "", inputError("words are at most 64 letters")
	}
	switch op.Name {
	case "insert":
		if p.t.Insert(w) {
			return "inserted " + strconv.Quote(w), nil
		}
		return strconv.Quote(w) + " is already there", nil
	case "search":
		return fmt.Sprintf("contains %q: %v", w, p.t.Contains(w)), nil
	case "prefix":
		words := p.t.WithPrefix(w)
		if len(words) == 0 {
			return fmt.Sprintf("no words start with %q", w), nil
		}
		return fmt.Sprintf("words starting with %q: %s", w, strings.Join(words, ", ")), nil
	}
	if p.t.Remove(w) {
		return "removed " + strconv.Quote(w), nil
	}
	return strconv.Quote(w) + " is not there", nil
}

type hashP struct {
	t *lookup.HashTable[int, int]
}

func (p *hashP) Snapshot() trace.Snapshot { return p.t.Snapshot() }

func (p *hashP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		f := strings.Fields(op.Value)
		if len(f) == 1 {
			op.Value += " " + f[0] // a lone KEY is stored as its own value
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		before := p.t.Buckets()
		p.t.Put(kv[0], kv[1])
		pushed(r, kv[0])
		msg := fmt.Sprintf("put %d: %d into bucket %d", kv[0], kv[1], p.t.Bucket(kv[0]))
		if p.t.Buckets() != before {
			msg += fmt.Sprintf(" (too full, so resized to %d buckets and rehashed)", p.t.Buckets())
		}
		return msg, nil
	case "get", "remove":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "get" {
			if v, ok := p.t.Get(k); ok {
				return fmt.Sprintf("%d is %d (bucket %d)", k, v, p.t.Bucket(k)), nil
			}
			return fmt.Sprintf("%d is not there", k), nil
		}
		if p.t.Remove(k) {
			popped(r, k)
			return fmt.Sprintf("removed %d", k), nil
		}
		return fmt.Sprintf("%d is not there", k), nil
	}
	return "", unknownOp(op, "put", "get", "remove")
}

type ufP struct{ u *lookup.UnionFind[int] }

func (p *ufP) Snapshot() trace.Snapshot { return p.u.Snapshot() }

func (p *ufP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "add", "find":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "add" {
			if p.u.Add(v) {
				pushed(r, v)
				return fmt.Sprintf("added %d (%d groups)", v, p.u.Groups()), nil
			}
			return fmt.Sprintf("%d is already there", v), nil
		}
		if root, ok := p.u.Find(v); ok {
			return fmt.Sprintf("%d is in the group of %d", v, root), nil
		}
		return fmt.Sprintf("%d is not there", v), nil
	case "union", "connected":
		ab, err := opPair(op)
		if err != nil {
			return "", err
		}
		if op.Name == "connected" {
			return fmt.Sprintf("%d and %d connected: %v", ab[0], ab[1], p.u.Connected(ab[0], ab[1])), nil
		}
		if p.u.Union(ab[0], ab[1]) {
			r.Record(trace.Event{Kind: trace.Link, Values: ab})
			return fmt.Sprintf("merged the groups of %d and %d (%d groups)", ab[0], ab[1], p.u.Groups()), nil
		}
		return fmt.Sprintf("%d and %d were already together", ab[0], ab[1]), nil
	}
	return "", unknownOp(op, "add", "union", "find", "connected")
}

type lruP struct{ c *lookup.LRU[int, int] }

func (p *lruP) Snapshot() trace.Snapshot { return p.c.Snapshot() }

func (p *lruP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		if len(strings.Fields(op.Value)) == 1 {
			op.Value += " " + op.Value
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		pushed(r, kv[0])
		if old, ok := p.c.Put(kv[0], kv[1]); ok {
			popped(r, old)
			return fmt.Sprintf("put %d: %d, evicted %d (least recently used)", kv[0], kv[1], old), nil
		}
		return fmt.Sprintf("put %d: %d", kv[0], kv[1]), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k); ok {
			return fmt.Sprintf("%d is %d (now the newest)", k, v), nil
		}
		return fmt.Sprintf("%d is not cached", k), nil
	}
	return "", unknownOp(op, "put", "get")
}

type ttlP struct {
	c   *lookup.TTLCache[int, int]
	now int64
}

func (p *ttlP) Snapshot() trace.Snapshot { return p.c.Snapshot() }

func (p *ttlP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		evicted := p.c.Put(a[0], a[1], p.now)
		pushed(r, a[0])
		if len(evicted) > 0 {
			return fmt.Sprintf("at time %d put %d=%d, evicting %v", p.now, a[0], a[1], evicted), nil
		}
		return fmt.Sprintf("at time %d put %d=%d, expiring at %d", p.now, a[0], a[1], p.now+p.c.TTL()), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k, p.now); ok {
			return fmt.Sprintf("at time %d get %d: %d", p.now, k, v), nil
		}
		return fmt.Sprintf("at time %d get %d: missing or expired", p.now, k), nil
	case "tick":
		n, err := opInt(op)
		if err != nil || n < 0 {
			return "", inputError("tick needs a number of steps, 0 or more")
		}
		p.now += int64(n)
		return fmt.Sprintf("the clock is now %d", p.now), nil
	}
	return "", unknownOp(op, "put", "get", "tick")
}
