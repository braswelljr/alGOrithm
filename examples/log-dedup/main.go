// Command log-dedup drops repeated log lines with a Bloom filter: a fixed
// few kilobytes answer "seen before?" for millions of lines, at the price of
// rare false positives and never a false negative. Databases (Cassandra,
// RocksDB) use the same filter to skip disk reads for keys that are absent.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
)

func main() {
	lines := []string{
		"ERROR db timeout after 5s",
		"WARN cache miss for user:42",
		"ERROR db timeout after 5s",
		"INFO request served in 12ms",
		"WARN cache miss for user:42",
		"ERROR disk 91% full",
		"ERROR db timeout after 5s",
	}
	seen := probabilistic.NewBloomFilter(1024, 4)
	fmt.Println("unique log lines:")
	dropped := 0
	for _, l := range lines {
		if seen.MightContain(l) {
			dropped++
			continue
		}
		seen.Add(l)
		fmt.Println("  " + l)
	}
	fmt.Printf("dropped %d repeats; false positive rate now %.4f%%\n", dropped, 100*seen.FalsePositiveRate())
}
