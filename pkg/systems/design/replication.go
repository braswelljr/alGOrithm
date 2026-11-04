package design

import (
	"fmt"
	"math/rand"
)

// ReplicationConfig is a leader with followers and a user reading back.
type ReplicationConfig struct {
	Followers     int   // read replicas
	MaxLag        int   // most ticks a follower trails the leader
	Writes        int   // updates, each read back right after by the same user
	ReadYourWrite bool  // route a user's reads to the leader for MaxLag ticks after they write
	Seed          int64 // lag and which replica serves a read
}

// LeaderFollower models single-leader replication with asynchronous
// followers (MySQL, PostgreSQL read replicas): writes go to the leader,
// which streams its log to followers that each apply it up to MaxLag ticks
// late; reads are spread over the followers. A user who updates a profile
// and reloads may read a follower that has not caught up and see the old
// value: a read-your-writes violation. Routing a user's reads to the leader
// for a short while after their own write fixes it at the cost of leader
// load. It leaves out failover and synchronous replicas.
//
// Time: O(Writes). Space: O(Followers).
func LeaderFollower(c ReplicationConfig) (Report, error) {
	if err := check(c.Followers >= 1 && c.Followers <= 50 && c.MaxLag >= 0 && c.MaxLag <= 1000 && c.Writes >= 1 && c.Writes <= 100_000, "1 to 50 followers, lag 0 to 1,000, 1 to 100,000 writes"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	comps := []Component{{ID: "user", Kind: Client, Label: "User"}, {ID: "leader", Kind: Database, Label: "Leader", Notes: "takes every write"}}
	links := []string{"user>leader:write", "user>leader:read (own write)"}
	for f := 0; f < c.Followers; f++ {
		comps = append(comps, Component{ID: fmt.Sprintf("f%d", f), Kind: Database, Label: fmt.Sprintf("Follower %d", f)})
		links = append(links, fmt.Sprintf("leader>f%d:replicate", f), fmt.Sprintf("user>f%d:read", f))
	}
	a := arch(comps, links...)
	stale, leaderReads := 0, 0
	for w := 0; w < c.Writes; w++ {
		f := rng.Intn(c.Followers)
		lag := 0
		if c.MaxLag > 0 {
			lag = rng.Intn(c.MaxLag + 1)
		}
		readAfter := rng.Intn(c.MaxLag + 1) // ticks between the write and the read
		toLeader := c.ReadYourWrite && readAfter <= c.MaxLag
		isStale := !toLeader && lag > readAfter
		if toLeader {
			leaderReads++
		}
		if isStale {
			stale++
		}
		if a.recording() {
			fr := a.frame(fmt.Sprintf("write %d, read back after %d ticks from %s: %s", w, readAfter,
				map[bool]string{true: "the leader", false: fmt.Sprintf("follower %d (lag %d)", f, lag)}[toLeader], map[bool]string{true: "old value", false: "new value"}[isStale]))
			a.send(fr, "user", "leader", w, OK)
			a.send(fr, "leader", fmt.Sprintf("f%d", f), w, Queued)
			if toLeader {
				a.Frames[len(a.Frames)-1].Requests = append(a.Frames[len(a.Frames)-1].Requests, Request{Link: 1, ID: w, Status: Hit})
			} else {
				a.send(fr, "user", fmt.Sprintf("f%d", f), w, map[bool]Status{true: Miss, false: Hit}[isStale])
			}
		}
	}
	var r Report
	r.Arch = a
	r.say("%d writes each read back at once: %d saw the old value (%d%%)", c.Writes, stale, percent(stale, c.Writes))
	if c.ReadYourWrite {
		r.say("read-your-writes: %d reads went to the leader, so none saw their own write missing", leaderReads)
	} else {
		r.say("turn on read-your-writes to send a user's reads to the leader just after they write")
	}
	r.put("stale reads", float64(stale))
	r.put("reads on the leader", float64(leaderReads))
	return r, nil
}

// MultiLeaderConfig is two datacenters each with a leader, writing the same keys.
type MultiLeaderConfig struct {
	Keys        int   // keys both sides write
	Writes      int   // writes on each side between two replications
	Rounds      int   // replication rounds
	VectorClock bool  // detect concurrent writes and keep both, instead of last write wins
	Seed        int64 // which keys and timestamps
}

// MultiLeader models multi-leader replication: each datacenter's leader
// takes writes locally (low latency, survives a datacenter outage) and they
// exchange changes every round. When both changed the same key in a round,
// the writes conflict. Last write wins keeps the one with the later
// timestamp and silently drops the other (lost updates); version vectors
// see the writes are concurrent and keep both as siblings for the
// application to merge (a shopping cart's union). Vector clocks themselves
// are the distributed vector-clock entry.
//
// Time: O(Rounds * Writes). Space: O(Keys).
func MultiLeader(c MultiLeaderConfig) (Report, error) {
	if err := check(c.Keys >= 1 && c.Keys <= 10_000 && c.Writes >= 0 && c.Writes <= 10_000 && c.Rounds >= 1 && c.Rounds <= 1000, "1 to 10,000 keys, 0 to 10,000 writes, 1 to 1,000 rounds"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	a := arch([]Component{
		{ID: "us", Kind: Database, Label: "Leader (US)"},
		{ID: "eu", Kind: Database, Label: "Leader (EU)"},
		{ID: "uc", Kind: Client, Label: "US users"},
		{ID: "ec", Kind: Client, Label: "EU users"},
	}, "uc>us:write", "ec>eu:write", "us>eu:replicate", "eu>us:replicate")
	conflicts, lost, siblings := 0, 0, 0
	for round := 0; round < c.Rounds; round++ {
		us, eu := map[int]bool{}, map[int]bool{}
		for w := 0; w < c.Writes; w++ {
			us[rng.Intn(c.Keys)] = true
			eu[rng.Intn(c.Keys)] = true
		}
		n := 0
		for k := range us {
			if eu[k] {
				n++
			}
		}
		conflicts += n
		if c.VectorClock {
			siblings += n
		} else {
			lost += n
		}
		if a.recording() {
			f := a.frame(fmt.Sprintf("round %d: %d keys written on both sides: %s", round, n,
				map[bool]string{true: "kept both as siblings", false: "last write wins dropped one each"}[c.VectorClock]))
			a.send(f, "uc", "us", round, OK)
			a.send(f, "ec", "eu", round, OK)
			st := OK
			if n > 0 {
				st = map[bool]Status{true: Queued, false: Dropped}[c.VectorClock]
			}
			a.send(f, "us", "eu", round, st)
			a.send(f, "eu", "us", round, st)
		}
	}
	var r Report
	r.Arch = a
	r.say("%d conflicting writes over %d rounds", conflicts, c.Rounds)
	if c.VectorClock {
		r.say("version vectors kept %d siblings for the application to merge: nothing lost", siblings)
	} else {
		r.say("last write wins silently dropped %d writes", lost)
	}
	r.put("conflicts", float64(conflicts))
	r.put("lost updates", float64(lost))
	r.put("siblings", float64(siblings))
	return r, nil
}
