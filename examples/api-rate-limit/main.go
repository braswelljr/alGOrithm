// Command api-rate-limit throttles a client hitting an API with a token
// bucket, the limiter behind most API gateways (AWS API Gateway, Stripe,
// nginx's limit_req is its leaky-bucket cousin): bursts up to the bucket
// size pass, and over time at most one request per interval gets through.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/systems/networking"
)

func main() {
	// Arrival times in milliseconds of one client's requests: a burst, then a trickle.
	arrivals := []int{0, 1, 2, 3, 4, 5, 6, 100, 150, 400, 401, 402}
	decisions, err := networking.TokenBucket(arrivals, 4, 100)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("token bucket: 4 tokens, one more every 100 ms")
	allowed := 0
	for _, d := range decisions {
		status := "429 Too Many Requests"
		if d.Allowed {
			status = "200 OK"
			allowed++
		}
		fmt.Printf("t=%3dms %-22s tokens left %d\n", d.Arrival, status, d.Left)
	}
	fmt.Printf("%d of %d requests allowed\n", allowed, len(decisions))
}
