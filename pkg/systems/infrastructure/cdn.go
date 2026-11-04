package infrastructure

import (
	"fmt"
	"math/rand"
)

// RegionStats is one region's edge cache in a CDN run.
type RegionStats struct {
	Region   string `json:"region"`
	Requests int    `json:"requests"`
	Hits     int    `json:"hits"`
	Misses   int    `json:"misses"`
}

// CDNResult is a CDN run.
type CDNResult struct {
	Regions []RegionStats `json:"regions"`
	Origin  int           `json:"origin"` // requests that reached the origin server
	Log     []string      `json:"log"`
}

var regionNames = []string{"us-east", "eu-west", "ap-south", "sa-east", "af-south", "ap-east"}

// CDN sends the given number of requests, one a tick, from the regions,
// for objects whose popularity follows a Zipf-like curve (object k is asked for
// about 1/k as often as object 1), seeded. Each region has its own edge
// cache, which keeps an object for ttl ticks: a hit is served from nearby, a
// miss goes to the origin server. A few popular objects make most requests,
// so even short TTLs give high hit rates; the long tail of rare objects
// still misses.
//
// Time: O(requests). Space: O(regions * objects).
func CDN(regions, objects, requests, ttl int, seed int64) (CDNResult, error) {
	if regions < 1 || regions > len(regionNames) || objects < 1 || objects > 1000 || requests < 1 || requests > 5000 || ttl < 1 || ttl > 5000 {
		return CDNResult{}, inputErr("1-%d regions, 1-1000 objects, 1-5000 requests, ttl 1-5000", len(regionNames))
	}
	rng := rand.New(rand.NewSource(seed))
	weights := make([]float64, objects)
	sum := 0.0
	for k := range weights {
		weights[k] = 1 / float64(k+1)
		sum += weights[k]
	}
	pick := func() int {
		x := rng.Float64() * sum
		for k, w := range weights {
			if x -= w; x <= 0 {
				return k
			}
		}
		return objects - 1
	}
	res := CDNResult{Regions: make([]RegionStats, regions)}
	expires := make([]map[int]int, regions)
	for r := range expires {
		expires[r] = map[int]int{}
		res.Regions[r].Region = regionNames[r]
	}
	for t := 0; t < requests; t++ {
		r, obj := rng.Intn(regions), pick()
		st := &res.Regions[r]
		st.Requests++
		if exp, ok := expires[r][obj]; ok && t < exp {
			st.Hits++
			continue
		}
		st.Misses++
		res.Origin++
		expires[r][obj] = t + ttl
	}
	for _, st := range res.Regions {
		rate := 0.0
		if st.Requests > 0 {
			rate = 100 * float64(st.Hits) / float64(st.Requests)
		}
		res.Log = append(res.Log, fmt.Sprintf("%s: %d requests, %.0f%% served from the edge", st.Region, st.Requests, rate))
	}
	res.Log = append(res.Log, fmt.Sprintf("the origin served %d of %d requests", res.Origin, requests))
	return res, nil
}
