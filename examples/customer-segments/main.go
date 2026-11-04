// Command customer-segments groups shoppers by yearly spend and visits with
// k-means, a common first step in marketing analytics: each customer joins
// the nearest segment center, and the centers move to their members' mean.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/ml/clustering"
)

func main() {
	// spend in hundreds of dollars, visits per year
	customers := [][]float64{
		{2, 3}, {3, 4}, {2, 5}, {3, 3}, // occasional
		{20, 40}, {22, 38}, {19, 45}, {24, 41}, // loyal
		{60, 8}, {58, 10}, {65, 7}, // big spenders, rare visits
	}
	res, err := clustering.KMeans(customers, 3, 7)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("3 segments after %d iterations:\n", res.Iterations)
	for c, center := range res.Centroids {
		n := 0
		for _, l := range res.Labels {
			if l == c {
				n++
			}
		}
		fmt.Printf("  segment %d: %d customers, avg spend $%.0f, %.0f visits\n", c, n, center[0]*100, center[1])
	}
}
