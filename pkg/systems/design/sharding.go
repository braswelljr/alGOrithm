package design

import (
	"fmt"
	"math/rand"
	"sort"
)

// ShardingConfig is data to split across shards.
type ShardingConfig struct {
	Keys       int   // rows, keyed by an increasing id (a sign-up time, an order number)
	Shards     int   // database shards
	Writes     int   // new rows written after the initial load (always the newest ids)
	HotPercent int   // percent of reads that go to the newest tenth of the ids
	Reads      int   // reads
	Seed       int64 // which rows are read
}

// Sharding compares three ways to split a table across shards. Range
// sharding (ids 0 to 999 on shard 0, ...) keeps ranges together for scans
// but sends every new row, and the hot recent ones, to the last shard: a
// hotspot. Hash sharding (hash(id) mod shards) spreads load evenly but
// scatters a range scan over every shard. A directory (a lookup table of id
// ranges to shards) can move any range anywhere, at the cost of one more
// lookup and a service to keep. It reports each strategy's busiest and
// quietest shard. Moving data when shards are added is the resharding entry.
//
// Time: O(Keys + Writes + Reads). Space: O(Shards).
func Sharding(c ShardingConfig) (Report, error) {
	if err := check(c.Keys >= 1 && c.Keys <= 1_000_000 && c.Shards >= 2 && c.Shards <= 256 && c.Writes >= 0 && c.HotPercent >= 0 && c.HotPercent <= 100 && c.Reads >= 0,
		"1 to 1,000,000 keys, 2 to 256 shards, a hot percent of 0 to 100"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	total := c.Keys + c.Writes
	var touched []int
	for i := 0; i < c.Writes; i++ {
		touched = append(touched, c.Keys+i)
	}
	for i := 0; i < c.Reads; i++ {
		if rng.Intn(100) < c.HotPercent {
			touched = append(touched, total-1-rng.Intn(max(1, total/10)))
		} else {
			touched = append(touched, rng.Intn(total))
		}
	}
	per := (c.Keys + c.Shards - 1) / c.Shards
	strategies := []struct {
		name  string
		shard func(id int) int
	}{
		{"range", func(id int) int { return min(id/per, c.Shards-1) }},
		{"hash", func(id int) int { return int(hash64(fmt.Sprint(id)) % uint64(c.Shards)) }},
		{"directory (ranges rebalanced by load)", nil},
	}
	var r Report
	t := r.table("busiest shard (%)", "quietest shard (%)", "busiest / average x100")
	for _, s := range strategies {
		load := make([]int, c.Shards)
		if s.shard == nil {
			// a directory splits the touched ids into ranges of equal load
			sorted := append([]int(nil), touched...)
			sort.Ints(sorted)
			for i := range sorted {
				load[min(i*c.Shards/max(1, len(sorted)), c.Shards-1)]++
			}
		} else {
			for _, id := range touched {
				load[s.shard(id)]++
			}
		}
		hi, lo := 0, len(touched)
		for _, n := range load {
			hi, lo = max(hi, n), min(lo, n)
		}
		avg := max(1, len(touched)/c.Shards)
		t.add(s.name, percent(hi, len(touched)), percent(lo, len(touched)), hi*100/avg)
		r.put(s.name+" busiest (%)", float64(percent(hi, len(touched))))
	}
	r.say("%d writes of new ids and %d reads (%d%% to the newest tenth) over %d shards", c.Writes, c.Reads, c.HotPercent, c.Shards)
	r.say("range sharding piles the new and hot ids on the last shard; hashing spreads them; a directory balances by moving ranges")
	return r, nil
}

// ReshardingConfig is a cluster that grows.
type ReshardingConfig struct {
	Keys int // keys stored
	From int // shards before
	To   int // shards after
}

// Resharding counts the keys that move when a cluster grows from From to To
// shards. With hash mod n almost every key moves (all but about From/To of
// them), so growing means copying nearly everything. With consistent hashing
// (100 virtual nodes per shard) only the keys the new shards take over move:
// close to the least possible, (To-From)/To, and closer with more virtual
// nodes per shard. A directory moves exactly the ranges chosen. The ring
// itself is the security/hashing consistent-hashing entry; rendezvous
// hashing is in systems/distributed.
//
// Time: O(Keys log(To * 100)). Space: O(To * 100).
func Resharding(c ReshardingConfig) (Report, error) {
	if err := check(c.Keys >= 1 && c.Keys <= 200_000 && c.From >= 1 && c.To > c.From && c.To <= 256, "1 to 200,000 keys, growing to at most 256 shards"); err != nil {
		return Report{}, err
	}
	ring := func(n int) func(uint64) int {
		type point struct {
			at    uint64
			shard int
		}
		var pts []point
		for s := 0; s < n; s++ {
			for v := 0; v < 100; v++ {
				pts = append(pts, point{hash64(fmt.Sprintf("shard-%d-%d", s, v)), s})
			}
		}
		sort.Slice(pts, func(i, j int) bool { return pts[i].at < pts[j].at })
		return func(h uint64) int {
			i := sort.Search(len(pts), func(i int) bool { return pts[i].at >= h })
			return pts[i%len(pts)].shard
		}
	}
	before, after := ring(c.From), ring(c.To)
	modMoved, ringMoved := 0, 0
	for k := 0; k < c.Keys; k++ {
		h := hash64(fmt.Sprintf("key-%d", k))
		if h%uint64(c.From) != h%uint64(c.To) {
			modMoved++
		}
		if before(h) != after(h) {
			ringMoved++
		}
	}
	ideal := c.Keys * (c.To - c.From) / c.To
	var r Report
	t := r.table("keys moved", "moved (%)")
	t.add("hash mod n", modMoved, percent(modMoved, c.Keys))
	t.add("consistent hashing", ringMoved, percent(ringMoved, c.Keys))
	t.add("the least possible", ideal, percent(ideal, c.Keys))
	r.say("growing %d to %d shards: mod n moves %d%% of keys, consistent hashing %d%%, the least possible is %d%%",
		c.From, c.To, percent(modMoved, c.Keys), percent(ringMoved, c.Keys), percent(ideal, c.Keys))
	r.put("mod n moved (%)", float64(percent(modMoved, c.Keys)))
	r.put("consistent hashing moved (%)", float64(percent(ringMoved, c.Keys)))
	return r, nil
}
