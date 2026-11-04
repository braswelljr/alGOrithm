package infrastructure

import (
	"fmt"
	"math"
	"math/rand"
)

// Strategy is how a load balancer picks a server.
type Strategy string

// The strategies Balance knows.
const (
	RoundRobin       Strategy = "round-robin"       // each server in turn
	WeightedRobin    Strategy = "weighted"          // in turn, weight times each
	LeastConnections Strategy = "least-connections" // the server with the fewest open requests, per unit of speed
	PowerOfTwo       Strategy = "power-of-two"      // two random servers, the less loaded one
	Random           Strategy = "random"            // any server
)

// BalanceResult is a Balance run: open requests on each server at each
// tick, how many each served, and how the load spread.
type BalanceResult struct {
	Active  [][]int  `json:"active"` // Active[s][t]
	Served  []int    `json:"served"`
	Peak    int      `json:"peak"`    // the most open requests on one server
	Latency float64  `json:"latency"` // average ticks a request took, queueing included
	Log     []string `json:"log"`
}

// Balance sends one request a tick for requests ticks to servers whose
// speeds are speeds (a server of speed 2 works twice as fast). Each request
// needs a seeded 2 to 12 units of work and shares its server's speed with
// the other open requests there, so a crowded server slows down. Round robin
// ignores speed and overloads slow servers; weighted round robin sends in
// proportion to speed; least connections looks at the current load; power of
// two choices samples two servers and takes the better, nearly as good as
// least connections without looking at every server.
//
// Time: O(T * (servers + open requests)). Space: O(servers * T).
func Balance(strategy Strategy, speeds []int, requests int, seed int64) (BalanceResult, error) {
	switch strategy {
	case RoundRobin, WeightedRobin, LeastConnections, PowerOfTwo, Random:
	default:
		return BalanceResult{}, inputErr("unknown strategy %q", strategy)
	}
	if len(speeds) < 2 || len(speeds) > 10 || requests < 1 || requests > 300 {
		return BalanceResult{}, inputErr("2-10 servers and 1-300 requests")
	}
	for _, s := range speeds {
		if s < 1 || s > 10 {
			return BalanceResult{}, inputErr("speeds 1-10")
		}
	}
	rng := rand.New(rand.NewSource(seed))
	n := len(speeds)
	type req struct {
		left    float64
		arrived int
	}
	open := make([][]*req, n)
	res := BalanceResult{Active: make([][]int, n), Served: make([]int, n)}
	var wheel []int // for weighted round robin
	for s, w := range speeds {
		for i := 0; i < w; i++ {
			wheel = append(wheel, s)
		}
	}
	load := func(s int) float64 { return float64(len(open[s])) / float64(speeds[s]) }
	next, totalLatency, finished := 0, 0, 0
	for tick := 0; finished < requests && tick < 10000; tick++ {
		if tick < requests {
			var s int
			switch strategy {
			case RoundRobin:
				s = next % n
			case WeightedRobin:
				s = wheel[next%len(wheel)]
			case LeastConnections:
				for i := range open {
					if load(i) < load(s) {
						s = i
					}
				}
			case PowerOfTwo:
				a, b := rng.Intn(n), rng.Intn(n)
				s = a
				if load(b) < load(a) {
					s = b
				}
			default:
				s = rng.Intn(n)
			}
			next++
			work := float64(2 + rng.Intn(11))
			open[s] = append(open[s], &req{left: work, arrived: tick})
		}
		for s := range open {
			res.Active[s] = append(res.Active[s], len(open[s]))
			res.Peak = max(res.Peak, len(open[s]))
			if len(open[s]) == 0 {
				continue
			}
			share := float64(speeds[s]) / float64(len(open[s]))
			kept := open[s][:0]
			for _, r := range open[s] {
				r.left -= share
				if r.left <= 1e-9 {
					res.Served[s]++
					finished++
					totalLatency += tick + 1 - r.arrived
				} else {
					kept = append(kept, r)
				}
			}
			open[s] = kept
		}
	}
	res.Latency = float64(totalLatency) / float64(requests)
	for s := range speeds {
		res.Log = append(res.Log, fmt.Sprintf("server %d (speed %d): served %d, most open at once %d", s, speeds[s], res.Served[s], peakOf(res.Active[s])))
	}
	res.Log = append(res.Log, fmt.Sprintf("average latency %.1f ticks", math.Round(res.Latency*10)/10))
	return res, nil
}

func peakOf(v []int) int {
	p := 0
	for _, x := range v {
		p = max(p, x)
	}
	return p
}
