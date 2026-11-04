package design

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// The layout of a Snowflake ID, as Twitter introduced it: one unused sign
// bit, 41 bits of milliseconds since an epoch, 10 bits of worker, 12 bits of
// sequence within the millisecond.
const (
	workerBits   = 10
	sequenceBits = 12
	maxWorker    = 1<<workerBits - 1
	maxSequence  = 1<<sequenceBits - 1
)

// ErrClockBackwards is returned by Snowflake.Next when the clock is behind
// the last ID's millisecond: handing out an ID then could repeat one.
var ErrClockBackwards = errors.New("design: clock moved backwards")

// Snowflake generates Snowflake IDs for one worker: 64-bit integers that sort
// by creation time, unique across workers without any coordination, up to
// 4,096 a millisecond per worker.
type Snowflake struct {
	worker   int64
	lastMs   int64
	sequence int64
}

// NewSnowflake returns a generator for worker (0 to 1023).
func NewSnowflake(worker int) (*Snowflake, error) {
	if err := check(worker >= 0 && worker <= maxWorker, "worker %d is outside 0 to %d", worker, maxWorker); err != nil {
		return nil, err
	}
	return &Snowflake{worker: int64(worker), lastMs: -1}, nil
}

// Next returns the next ID at the clock reading nowMs (milliseconds since the
// epoch), or ok false when this millisecond's 4,096 IDs are used up (the
// caller waits for the next millisecond), or ErrClockBackwards when the clock
// went back.
//
// Time: O(1). Space: O(1).
func (s *Snowflake) Next(nowMs int64) (id int64, ok bool, err error) {
	switch {
	case nowMs < s.lastMs:
		return 0, false, ErrClockBackwards
	case nowMs == s.lastMs:
		if s.sequence == maxSequence {
			return 0, false, nil
		}
		s.sequence++
	default:
		s.lastMs, s.sequence = nowMs, 0
	}
	return nowMs<<(workerBits+sequenceBits) | s.worker<<sequenceBits | s.sequence, true, nil
}

// DecodeSnowflake splits an ID into its millisecond, worker and sequence.
//
// Time: O(1). Space: O(1).
func DecodeSnowflake(id int64) (ms int64, worker, sequence int) {
	return id >> (workerBits + sequenceBits), int(id >> sequenceBits & maxWorker), int(id & maxSequence)
}

// SnowflakeConfig is a run of Snowflake generators.
type SnowflakeConfig struct {
	Workers      int // generators, each with its own worker number
	PerMs        int // IDs each worker is asked for every millisecond
	Millis       int // milliseconds to run
	SkewAt       int // millisecond at which worker 0's clock jumps back (0: never)
	SkewMs       int // how far it jumps back
	FirstMs      int // the clock's first reading
	SampleWorker int // the worker whose first IDs are decoded in the report
}

// SnowflakeRun runs Workers generators for Millis milliseconds, each asked
// for PerMs IDs a millisecond, and reports the IDs generated, the requests
// that waited for the next millisecond (more than 4,096 asked in one), and
// what happened when worker 0's clock jumped back: it refuses to hand out IDs
// until the clock passes its last one again, which is how Snowflake stays
// unique. The model leaves out real time; each millisecond is one step.
//
// Time: O(Workers * PerMs * Millis). Space: O(Workers * PerMs * Millis) for the check.
func SnowflakeRun(c SnowflakeConfig) (Report, error) {
	if err := check(c.Workers >= 1 && c.Workers <= maxWorker+1 && c.PerMs >= 1 && c.Millis >= 1 && c.SkewAt >= 0 && c.SkewMs >= 0 &&
		c.FirstMs >= 0 && c.SampleWorker >= 0 && c.SampleWorker < c.Workers, "workers 1 to 1024, positive counts, and a sample worker among them"); err != nil {
		return Report{}, err
	}
	gens := make([]*Snowflake, c.Workers)
	for w := range gens {
		gens[w], _ = NewSnowflake(w)
	}
	var r Report
	seen := map[int64]bool{}
	ids, waits, refused := 0, 0, 0
	t := r.table("worker", "ms", "sequence")
	for m := 0; m < c.Millis; m++ {
		for w, g := range gens {
			clock := int64(c.FirstMs + m)
			if w == 0 && c.SkewAt > 0 && m >= c.SkewAt && m < c.SkewAt+c.SkewMs {
				clock -= int64(c.SkewMs)
			}
			for k := 0; k < c.PerMs; k++ {
				id, ok, err := g.Next(clock)
				switch {
				case err != nil:
					refused++
					continue
				case !ok:
					waits++
					continue
				}
				if seen[id] {
					return Report{}, fmt.Errorf("design: duplicate id %d", id)
				}
				seen[id] = true
				ids++
				if w == c.SampleWorker && len(t.Rows) < 5 {
					ms, wk, seq := DecodeSnowflake(id)
					t.add(fmt.Sprintf("id %d", id), wk, int(ms), seq)
				}
			}
		}
	}
	r.say("41 bits of milliseconds | 10 bits of worker | 12 bits of sequence: sortable by time, unique with no coordination")
	r.say("%d IDs from %d workers over %d ms, all unique", ids, c.Workers, c.Millis)
	if waits > 0 {
		r.say("%d requests waited for the next millisecond: a worker makes at most 4,096 IDs a millisecond", waits)
	}
	if refused > 0 {
		r.say("worker 0's clock went back %d ms; it refused %d requests until it passed its last ID again", c.SkewMs, refused)
	}
	r.put("ids", float64(ids))
	r.put("waited", float64(waits))
	r.put("refused (clock went back)", float64(refused))
	return r, nil
}

// IDOrderingConfig is an index to fill with IDs of each scheme.
type IDOrderingConfig struct {
	Inserts  int   // IDs inserted
	PageSize int   // index entries per page
	PerMs    int   // IDs created each millisecond
	Seed     int64 // the random bits
}

// IDOrdering inserts the same number of random UUIDs (version 4), time-
// ordered UUIDs (version 7) and ULIDs into a sorted index, as a database's
// primary key, and counts the inserts that land in the index's last page (an
// append, cheap and cache friendly) against those that land in an older page
// (a random write that drags a cold page into memory and splits pages). The
// model compares only the 64 most significant bits of each ID and leaves out
// page splitting itself.
//
// Time: O(Inserts^2) for the sorted inserts. Space: O(Inserts).
func IDOrdering(c IDOrderingConfig) (Report, error) {
	if err := check(c.Inserts >= 1 && c.Inserts <= 20_000 && c.PageSize >= 1 && c.PerMs >= 1, "inserts 1 to 20,000, page size and ids a millisecond at least 1"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	schemes := []struct {
		name string
		next func(i int) uint64
	}{
		{"UUID v4 (random)", func(int) uint64 { return rng.Uint64() }},
		{"UUID v7 (48-bit ms + random)", func(i int) uint64 { return uint64(i/c.PerMs)<<16 | uint64(rng.Intn(1<<16)) }},
		{"ULID (48-bit ms + 80 random bits)", func(i int) uint64 { return uint64(i/c.PerMs)<<16 | uint64(rng.Intn(1<<16)) }},
	}
	var r Report
	t := r.table("appends to the last page", "writes to older pages", "random writes (%)")
	for _, s := range schemes {
		var index []uint64
		appends, random := 0, 0
		for i := 0; i < c.Inserts; i++ {
			id := s.next(i)
			at := sort.Search(len(index), func(j int) bool { return index[j] > id })
			if at/c.PageSize >= len(index)/c.PageSize {
				appends++
			} else {
				random++
			}
			index = append(index, 0)
			copy(index[at+1:], index[at:])
			index[at] = id
		}
		t.add(s.name, appends, random, percent(random, c.Inserts))
		r.put(s.name+" random writes (%)", float64(percent(random, c.Inserts)))
	}
	r.say("random IDs scatter inserts across the whole index; time-ordered IDs append at the end, like an auto-increment key, but need no central counter")
	return r, nil
}

// TicketConfig is a pair (or more) of ticket servers handing out IDs.
type TicketConfig struct {
	Servers int // ticket servers, each counting with its own offset
	IDs     int // IDs requested, round robin across the servers that are up
	DownAt  int // request number at which server 1 goes down (0: never)
}

// TicketServer hands out IDs as Flickr's ticket servers do: each of n
// database servers counts by n from its own offset (server 1: 1, 3, 5...;
// server 2: 2, 4, 6...), so they never collide. When one is down, the others
// carry on and its numbers are skipped: IDs stay unique but are no longer
// dense or strictly ordered across servers.
//
// Time: O(IDs). Space: O(IDs).
func TicketServer(c TicketConfig) (Report, error) {
	if err := check(c.Servers >= 1 && c.Servers <= 16 && c.IDs >= 1 && c.IDs <= 10_000 && c.DownAt >= 0, "servers 1 to 16, ids 1 to 10,000"); err != nil {
		return Report{}, err
	}
	next := make([]int, c.Servers)
	for i := range next {
		next[i] = i + 1
	}
	var r Report
	var got []int
	for k := 0; k < c.IDs; k++ {
		s := k % c.Servers
		if c.DownAt > 0 && k >= c.DownAt && s == 0 && c.Servers > 1 {
			s = 1
		}
		got = append(got, next[s])
		next[s] += c.Servers
	}
	if len(got) <= 20 {
		r.say("ids handed out: %v", got)
	} else {
		r.say("first ids handed out: %v", got[:20])
	}
	maxID := 0
	for _, id := range got {
		maxID = max(maxID, id)
	}
	r.say("each server counts by %d from its own offset, so no two servers ever give the same id", c.Servers)
	r.put("ids", float64(len(got)))
	r.put("highest id", float64(maxID))
	r.put("unused numbers below the highest", float64(maxID-len(got)))
	return r, nil
}
