// Command cache-sharding spreads cache keys over servers with consistent
// hashing, as memcached clients, Amazon's Dynamo, Cassandra and many load
// balancers do: adding a server moves only about 1/n of the keys instead of
// nearly all of them, as hash(key) mod n would.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/security/hashing"
)

func main() {
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("session:%d", i)
	}
	ring := hashing.NewRing(100)
	for _, s := range []string{"cache-a", "cache-b", "cache-c"} {
		ring.Add(s)
	}
	before := ring.Assign(keys)
	ring.Add("cache-d")
	after := ring.Assign(keys)
	load := map[string]int{}
	for _, s := range after {
		load[s]++
	}
	fmt.Printf("keys moved when cache-d joined: %d of %d (%.0f%%)\n",
		hashing.Moved(before, after), len(keys), 100*float64(hashing.Moved(before, after))/float64(len(keys)))
	for _, s := range ring.Nodes() {
		fmt.Printf("  %s holds %d keys\n", s, load[s])
	}
}
