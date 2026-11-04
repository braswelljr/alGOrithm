// Command unique-visitors estimates how many different visitors a site had
// with HyperLogLog, in a few kilobytes instead of a set of every visitor id.
// Redis's PFCOUNT, BigQuery's APPROX_COUNT_DISTINCT and analytics pipelines
// count distinct users this way.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
)

func main() {
	hll := probabilistic.NewHyperLogLog(12) // 4096 registers
	exact := map[string]bool{}
	// 50,000 page views from 12,000 visitors, many coming back.
	for i := 0; i < 50_000; i++ {
		visitor := fmt.Sprintf("visitor-%d", (i*7919)%12_000)
		hll.Add(visitor)
		exact[visitor] = true
	}
	est := hll.Count()
	fmt.Printf("estimated unique visitors: %d\n", est)
	fmt.Printf("exact: %d, error %.2f%% (expected about %.2f%%)\n",
		len(exact), 100*float64(est-len(exact))/float64(len(exact)), 100*hll.StandardError())
}
