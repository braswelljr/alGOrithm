package design

import (
	"fmt"
	"math/rand"
)

// LimiterConfig is an API behind several nodes and one global limit.
type LimiterConfig struct {
	Nodes     int   // API nodes behind the load balancer
	Limit     int   // requests allowed per window across all nodes
	Requests  int   // requests in one window, skewed toward a few nodes
	SyncEvery int   // requests between two syncs of the gossiped counters
	Seed      int64 // which node each request lands on
}

// DistributedRateLimiter compares three ways to enforce one rate limit (say
// 1,000 requests a minute per user) across many API nodes. A central counter
// (Redis INCR) is exact but costs a round trip per request and is one more
// thing that can fail. Local limits of Limit / Nodes each need no round trip
// but reject requests on busy nodes while quiet nodes have room. Local
// counters that sync every SyncEvery requests are cheap and nearly fair, but
// let a little over the limit through between syncs. The token or leaky
// bucket each node runs is the networking token-bucket entry.
//
// Time: O(Requests * Nodes). Space: O(Nodes).
func DistributedRateLimiter(c LimiterConfig) (Report, error) {
	if err := check(c.Nodes >= 1 && c.Nodes <= 100 && c.Limit >= 1 && c.Requests >= 1 && c.Requests <= 1_000_000 && c.SyncEvery >= 1,
		"1 to 100 nodes, a positive limit, 1 to 1,000,000 requests, sync every 1 or more"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	pick := popular(rng, c.Nodes)
	nodes := make([]int, c.Requests)
	for i := range nodes {
		nodes[i] = pick()
	}
	ideal := min(c.Requests, c.Limit)
	var r Report
	t := r.table("allowed", "over the limit", "wrongly rejected", "round trips")

	central := min(c.Requests, c.Limit)
	t.add("central counter", central, 0, 0, c.Requests)

	local := make([]int, c.Nodes)
	allowedLocal := 0
	share := c.Limit / c.Nodes
	for _, n := range nodes {
		if local[n] < share {
			local[n]++
			allowedLocal++
		}
	}
	t.add("local limit each", allowedLocal, 0, ideal-allowedLocal, 0)

	seen := 0 // the global count as of the last sync
	pending := make([]int, c.Nodes)
	allowedSync, syncs := 0, 0
	for i, n := range nodes {
		if seen+pending[n] < c.Limit {
			pending[n]++
			allowedSync++
		}
		if (i+1)%c.SyncEvery == 0 {
			for k := range pending {
				seen += pending[k]
				pending[k] = 0
			}
			syncs++
		}
	}
	over := max(0, allowedSync-c.Limit)
	t.add("synced local counters", allowedSync, over, max(0, ideal-allowedSync), syncs*c.Nodes)

	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "Clients"},
		{ID: "lb", Kind: Balancer, Label: "Load balancer"},
		{ID: "api", Kind: Service, Label: fmt.Sprintf("%d API nodes", c.Nodes), Notes: "each keeps a local count"},
		{ID: "redis", Kind: Cache, Label: "Central counter", Notes: "INCR with an expiry"},
	}, "client>lb", "lb>api", "api>redis:count", "api>client:429 when over")
	f := a.frame(fmt.Sprintf("%d requests in one window, limit %d", c.Requests, c.Limit))
	a.send(f, "client", "lb", 1, OK)
	a.send(f, "lb", "api", 1, OK)
	a.send(f, "api", "redis", 1, OK)
	if c.Requests > c.Limit {
		a.send(f, "api", "client", 2, Dropped)
	}
	r.Arch = a
	r.say("limit %d across %d nodes, %d requests skewed toward a few nodes", c.Limit, c.Nodes, c.Requests)
	r.say("central: exact, %d round trips; local %d each: %d wrongly rejected; synced every %d: %d over the limit", c.Requests, share, ideal-allowedLocal, c.SyncEvery, over)
	r.put("central allowed", float64(central))
	r.put("local allowed", float64(allowedLocal))
	r.put("synced allowed", float64(allowedSync))
	r.put("synced over the limit", float64(over))
	return r, nil
}
