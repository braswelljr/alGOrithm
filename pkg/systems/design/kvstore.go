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
	kv := newKVStore(c)
	next := popular(rand.New(rand.NewSource(c.Seed)), c.Keys)
	for op := 0; op < c.Ops; op++ {
		kv.down = c.DownNode >= 0 && op >= c.Ops/3 && op < 2*c.Ops/3
		if c.DownNode >= 0 && op == 2*c.Ops/3 {
			kv.handOff()
		}
		key := next()
		var f *Frame
		if kv.a.recording() {
			f = &Frame{}
		}
		if op%2 == 0 {
			kv.write(op, key, f)
		} else {
			kv.read(op, key, f)
		}
		if f != nil {
			kv.a.send(f, "client", "coord", op, OK)
			kv.a.Frames = append(kv.a.Frames, *f)
		}
	}
	return kv.report(), nil
}

type ringPoint struct {
	at   uint64
	node int
}

// kvStore is the store being simulated: the ring, each node's version of
// each key, the hints held for a down node, and the counts.
type kvStore struct {
	c       KVConfig
	ring    []ringPoint
	a       *Architecture
	version []map[int]int    // node -> key -> version
	hints   map[int][][2]int // intended node -> (key, version)
	latest  map[int]int
	down    bool

	writes, failedWrites, hinted, handed, reads, stale, repairs, failedReads int
}

// newKVStore places eight virtual points per node on the ring and draws the
// architecture.
func newKVStore(c KVConfig) *kvStore {
	kv := &kvStore{c: c, version: make([]map[int]int, c.Nodes), hints: map[int][][2]int{}, latest: map[int]int{}}
	for n := 0; n < c.Nodes; n++ {
		for v := 0; v < 8; v++ {
			kv.ring = append(kv.ring, ringPoint{hash64(fmt.Sprintf("node-%d-%d", n, v)), n})
		}
	}
	sort.Slice(kv.ring, func(i, j int) bool { return kv.ring[i].at < kv.ring[j].at })
	comps := []Component{{ID: "client", Kind: Client, Label: "Client"}, {ID: "coord", Kind: Service, Label: "Coordinator", Notes: fmt.Sprintf("N=%d R=%d W=%d", c.Replicas, c.R, c.W)}}
	links := []string{"client>coord"}
	for n := 0; n < c.Nodes; n++ {
		comps = append(comps, Component{ID: fmt.Sprintf("n%d", n), Kind: Database, Label: fmt.Sprintf("Node %d", n)})
		links = append(links, fmt.Sprintf("coord>n%d", n))
	}
	kv.a = arch(comps, links...)
	for i := range kv.version {
		kv.version[i] = map[int]int{}
	}
	return kv
}

// walk lists the nodes clockwise from the key's place on the ring, each once.
func (kv *kvStore) walk(key int) []int {
	h := hash64(fmt.Sprintf("key-%d", key))
	i := sort.Search(len(kv.ring), func(i int) bool { return kv.ring[i].at >= h })
	var order []int
	seen := map[int]bool{}
	for k := 0; len(order) < kv.c.Nodes; k++ {
		n := kv.ring[(i+k)%len(kv.ring)].node
		if !seen[n] {
			seen[n] = true
			order = append(order, n)
		}
	}
	return order
}

func (kv *kvStore) isUp(n int) bool { return !kv.down || n != kv.c.DownNode }

// handOff gives the returning node the writes held for it.
func (kv *kvStore) handOff() {
	d := kv.c.DownNode
	for _, h := range kv.hints[d] {
		kv.version[d][h[0]] = max(kv.version[d][h[0]], h[1])
		kv.handed++
	}
	kv.hints[d] = nil
}

// write sends a new version to the key's N replicas, a down one's copy going
// to the next healthy node with a hint; it counts if W acknowledge.
func (kv *kvStore) write(op, key int, f *Frame) {
	kv.writes++
	order := kv.walk(key)
	v := kv.latest[key] + 1
	acks := 0
	spare := kv.c.Replicas
	for _, n := range order[:kv.c.Replicas] {
		target := n
		if !kv.isUp(n) {
			for spare < len(order) && !kv.isUp(order[spare]) {
				spare++
			}
			if spare >= len(order) {
				continue
			}
			target = order[spare]
			spare++
			kv.hints[n] = append(kv.hints[n], [2]int{key, v})
			kv.hinted++
		}
		kv.version[target][key] = v
		acks++
		if f != nil {
			status := OK
			if target != n {
				status = Queued
			}
			kv.a.send(f, "coord", fmt.Sprintf("n%d", target), op, status)
		}
	}
	if acks >= kv.c.W {
		kv.latest[key] = v
	} else {
		kv.failedWrites++
	}
	if f != nil {
		f.Note = fmt.Sprintf("write key %d version %d: %d acks of W=%d", key, v, acks, kv.c.W)
	}
}

// read asks the first R healthy replicas, returns the newest version, and
// writes it back to any replica that is behind (read repair).
func (kv *kvStore) read(op, key int, f *Frame) {
	kv.reads++
	pref := kv.walk(key)[:kv.c.Replicas]
	var replies []int
	for _, n := range pref {
		if kv.isUp(n) {
			replies = append(replies, n)
		}
		if len(replies) == kv.c.R {
			break
		}
	}
	if len(replies) < kv.c.R {
		kv.failedReads++
		if f != nil {
			f.Note = fmt.Sprintf("read key %d: only %d replicas up, fewer than R=%d", key, len(replies), kv.c.R)
		}
		return
	}
	newest := 0
	for _, n := range replies {
		newest = max(newest, kv.version[n][key])
	}
	if newest < kv.latest[key] {
		kv.stale++
	}
	for _, n := range pref {
		if kv.isUp(n) && kv.version[n][key] < newest {
			kv.version[n][key] = newest
			kv.repairs++
		}
	}
	if f == nil {
		return
	}
	f.Note = fmt.Sprintf("read key %d from %d replicas: version %d (latest %d)", key, kv.c.R, newest, kv.latest[key])
	for _, n := range replies {
		status := Hit
		if kv.version[n][key] < kv.latest[key] {
			status = Miss
		}
		kv.a.send(f, "coord", fmt.Sprintf("n%d", n), op, status)
	}
}

func (kv *kvStore) report() Report {
	c := kv.c
	last := kv.a.frame("totals: keys stored per node")
	for n := 0; n < c.Nodes; n++ {
		last.load(fmt.Sprintf("n%d", n), len(kv.version[n]), len(kv.hints[n]), -1)
	}
	var r Report
	r.Arch = kv.a
	r.say("ring of %d nodes, each key on %d of them; R + W = %d %s N = %d", c.Nodes, c.Replicas, c.R+c.W, map[bool]string{true: ">", false: "<="}[c.R+c.W > c.Replicas], c.Replicas)
	r.say("%d writes (%d failed to reach W), %d sloppy-quorum writes to a stand-in node, %d handed back", kv.writes, kv.failedWrites, kv.hinted, kv.handed)
	r.say("%d reads: %d stale, %d failed, %d read repairs", kv.reads, kv.stale, kv.failedReads, kv.repairs)
	r.put("hinted writes", float64(kv.hinted))
	r.put("hints handed off", float64(kv.handed))
	r.put("stale reads", float64(kv.stale))
	r.put("read repairs", float64(kv.repairs))
	r.put("failed writes", float64(kv.failedWrites))
	return r
}
