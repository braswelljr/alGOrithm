// Command moving-average computes a 3-hour moving average and the peak of
// each 3-hour window over request counts with prefix sums and a monotonic
// queue, the sliding-window math behind monitoring dashboards and alerts.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
)

func main() {
	hourly := []int{120, 135, 160, 410, 390, 150, 140, 130, 520, 145}
	const w = 3
	prefix := arrays.PrefixSums(hourly) // prefix[i] = sum of the first i values
	peaks := arrays.WindowMax(hourly, w)
	fmt.Println("hours  avg   peak")
	for i := 0; i+w <= len(hourly); i++ {
		avg := float64(prefix[i+w]-prefix[i]) / w
		alert := ""
		if peaks[i] > 2*int(avg) {
			alert = "  spike"
		}
		fmt.Printf("%2d-%2d  %5.1f %4d%s\n", i, i+w-1, avg, peaks[i], alert)
	}
}
