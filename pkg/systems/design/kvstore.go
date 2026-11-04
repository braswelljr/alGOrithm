package design

import (
	"fmt"
	"math/rand"
	"sort"
)

// KVConfig is a Dynamo-style key-value store's cluster and workload.
type KVConfig struct {
	Nodes    int   // storage nodes on the ring
	Replicas int   // N: copies of every key
	R        int   // replicas a read waits for
	W        int   // replicas a write waits for
	Keys     int   // distinct keys
	Ops      int   // operations, half writes and half reads
	DownNode int   // node that is down for the middle third of the run (-1: none)
	Seed     int64 // which keys and operations
}

// KeyValueStore models a Dynamo-style store (Amazon's Dynamo, Cassandra,
// Riak). Keys are placed on a consistent-hashing ring and copied to the next
// N nodes clockwise. A write waits for W acks and a read for R replies; with
// R + W > N every read overlaps the latest write. When a node is down a
// write goes to the next healthy node with a hint (a sloppy quorum), which
// hands it back when the node returns (hinted handoff). A read that sees
// older versions on some replicas writes the newest back (read repair).
// Anti-entropy with Merkle trees, which repairs keys nobody reads, is the
// security/hashing and cryptography merkle-tree entries. The model uses one
// version counter per key instead of vector clocks.
//
// Time: O(Ops * Nodes). Space: O(Keys * Nodes).
func KeyValueStore(c KVConfig) (Report, error) {
	if err := check(c.Nodes >= 2 && c.Nodes <= 32 && c.Replicas >= 1 && c.Replicas <= c.Nodes && c.R >= 1 && c.R <= c.Replicas &&
		c.W >= 1 && c.W <= c.Replicas && c.Keys >= 1 && c.Ops >= 1 && c.Ops <= 100_000 && c.DownNode >= -1 && c.DownNode < c.Nodes,
		"2 to 32 nodes, 1 <= R, W <= N <= nodes, a down node among them or -1"); err != nil {
		return Report{}, err
	}
	type point struct {
		at   uint64
		node int
	}
	var ring []point
	for n := 0; n < c.Nodes; n++ {
		for v := 0; v < 8; v++ {
			ring = append(ring, point{hash64(fmt.Sprintf("node-%d-%d", n, v)), n})
		}
	}
	sort.Slice(ring, func(i, j int) bool { return ring[i].at < ring[j].at })
	walk := func(key int) []int {
		h := hash64(fmt.Sprintf("key-%d", key))
		i := sort.Search(len(ring), func(i int) bool { return ring[i].at >= h })
		var order []int
		seen := map[int]bool{}
		for k := 0; len(order) < c.Nodes; k++ {
			n := ring[(i+k)%len(ring)].node
			if !seen[n] {
				seen[n] = true
				order = append(order, n)
			}
		}
		return order
	}
	comps := []Component{{ID: "client", Kind: Client, Label: "Client"}, {ID: "coord", Kind: Service, Label: "Coordinator", Notes: fmt.Sprintf("N=%d R=%d W=%d", c.Replicas, c.R, c.W)}}
	links := []string{"client>coord"}
	for n := 0; n < c.Nodes; n++ {
		comps = append(comps, Component{ID: fmt.Sprintf("n%d", n), Kind: Database, Label: fmt.Sprintf("Node %d", n)})
		links = append(links, fmt.Sprintf("coord>n%d", n))
	}
	a := arch(comps, links...)
	version := make([]map[int]int, c.Nodes) // node -> key -> version
	for i := range version {
		version[i] = map[int]int{}
	}
	hints := map[int][][2]int{} // intended node -> (key, version)
	latest := map[int]int{}
	rng := rand.New(rand.NewSource(c.Seed))
	next := popular(rng, c.Keys)
	var writes, failedWrites, hinted, handed, reads, stale, repairs, failedReads int
	for op := 0; op < c.Ops; op++ {
		down := c.DownNode >= 0 && op >= c.Ops/3 && op < 2*c.Ops/3
		if c.DownNode >= 0 && op == 2*c.Ops/3 {
			for _, h := range hints[c.DownNode] {
				version[c.DownNode][h[0]] = max(version[c.DownNode][h[0]], h[1])
				handed++
			}
			hints[c.DownNode] = nil
		}
		isUp := func(n int) bool { return !down || n != c.DownNode }
		key := next()
		order := walk(key)
		pref := order[:c.Replicas]
		var f *Frame
		if a.recording() {
			f = &Frame{}
		}
		if op%2 == 0 {
			writes++
			v := latest[key] + 1
			acks := 0
			spare := c.Replicas
			for _, n := range pref {
				target := n
				if !isUp(n) {
					for spare < len(order) && !isUp(order[spare]) {
						spare++
					}
					if spare >= len(order) {
						continue
					}
					target = order[spare]
					spare++
					hints[n] = append(hints[n], [2]int{key, v})
					hinted++
				}
				version[target][key] = v
				acks++
				if f != nil {
					status := OK
					if target != n {
						status = Queued
					}
					a.send(f, "coord", fmt.Sprintf("n%d", target), op, status)
				}
			}
			if acks >= c.W {
				latest[key] = v
			} else {
				failedWrites++
			}
			if f != nil {
				f.Note = fmt.Sprintf("write key %d version %d: %d acks of W=%d", key, v, acks, c.W)
			}
		} else {
			reads++
			var replies []int
			for _, n := range pref {
				if isUp(n) {
					replies = append(replies, n)
				}
				if len(replies) == c.R {
					break
				}
			}
			if len(replies) < c.R {
				failedReads++
				if f != nil {
					f.Note = fmt.Sprintf("read key %d: only %d replicas up, fewer than R=%d", key, len(replies), c.R)
				}
			} else {
				newest := 0
				for _, n := range replies {
					newest = max(newest, version[n][key])
				}
				if newest < latest[key] {
					stale++
				}
				for _, n := range pref {
					if isUp(n) && version[n][key] < newest {
						version[n][key] = newest
						repairs++
					}
				}
				if f != nil {
					f.Note = fmt.Sprintf("read key %d from %d replicas: version %d (latest %d)", key, c.R, newest, latest[key])
					for _, n := range replies {
						status := Hit
						if version[n][key] < latest[key] {
							status = Miss
						}
						a.send(f, "coord", fmt.Sprintf("n%d", n), op, status)
					}
				}
			}
		}
		if f != nil {
			a.send(f, "client", "coord", op, OK)
			a.Frames = append(a.Frames, *f)
		}
	}
	last := a.frame("totals: keys stored per node")
	for n := 0; n < c.Nodes; n++ {
		last.load(fmt.Sprintf("n%d", n), len(version[n]), len(hints[n]), -1)
	}
	var r Report
	r.Arch = a
	r.say("ring of %d nodes, each key on %d of them; R + W = %d %s N = %d", c.Nodes, c.Replicas, c.R+c.W, map[bool]string{true: ">", false: "<="}[c.R+c.W > c.Replicas], c.Replicas)
	r.say("%d writes (%d failed to reach W), %d sloppy-quorum writes to a stand-in node, %d handed back", writes, failedWrites, hinted, handed)
	r.say("%d reads: %d stale, %d failed, %d read repairs", reads, stale, failedReads, repairs)
	r.put("hinted writes", float64(hinted))
	r.put("hints handed off", float64(handed))
	r.put("stale reads", float64(stale))
	r.put("read repairs", float64(repairs))
	r.put("failed writes", float64(failedWrites))
	return r, nil
}
