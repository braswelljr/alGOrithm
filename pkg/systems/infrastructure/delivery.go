package infrastructure

import (
	"fmt"
	"math/rand"
)

// Guarantee is a message delivery guarantee.
type Guarantee string

// The three guarantees.
const (
	AtMostOnce  Guarantee = "at-most-once"  // send once, never retry: may lose
	AtLeastOnce Guarantee = "at-least-once" // retry until acknowledged: may duplicate
	ExactlyOnce Guarantee = "exactly-once"  // at least once, plus idempotency keys at the consumer
)

// DeliveryResult counts what a Delivery run did with each message.
type DeliveryResult struct {
	Sent       int      `json:"sent"`       // transmissions, retries included
	Processed  int      `json:"processed"`  // times the consumer ran its side effect
	Lost       int      `json:"lost"`       // messages never processed
	Duplicates int      `json:"duplicates"` // extra processings of a message
	Log        []string `json:"log"`
}

// Delivery sends the given number of messages from a producer to a consumer over a
// link that drops each message, and each acknowledgment, with probability
// loss percent (seeded). At most once sends each once; a loss loses it. At
// least once resends until the acknowledgment arrives, so a lost
// acknowledgment means the consumer processes the message twice. Exactly
// once is at least once with an idempotency key: the consumer remembers
// which message ids it has processed and acknowledges a repeat without
// processing it again. ("Exactly-once delivery" is impossible on a lossy
// link; exactly-once processing is what this builds.)
//
// Time: O(messages * retries). Space: O(messages).
func Delivery(g Guarantee, messages, loss int, seed int64) (DeliveryResult, error) {
	if g != AtMostOnce && g != AtLeastOnce && g != ExactlyOnce {
		return DeliveryResult{}, inputErr("unknown guarantee %q", g)
	}
	if messages < 1 || messages > 1000 || loss < 0 || loss > 80 {
		return DeliveryResult{}, inputErr("1-1000 messages, loss 0-80 percent")
	}
	rng := rand.New(rand.NewSource(seed))
	lost := func() bool { return rng.Intn(100) < loss }
	var res DeliveryResult
	done := map[int]bool{}
	for id := 0; id < messages; id++ {
		processed := 0
		for try := 0; try < 50; try++ {
			res.Sent++
			if lost() {
				if g == AtMostOnce {
					break
				}
				continue // the producer times out and resends
			}
			if g != ExactlyOnce || !done[id] { // a repeat under exactly once is only acknowledged
				processed++
				done[id] = true
			}
			if g == AtMostOnce || !lost() {
				break // acknowledged (at most once does not wait for one)
			}
		}
		res.Processed += processed
		switch {
		case processed == 0:
			res.Lost++
		case processed > 1:
			res.Duplicates += processed - 1
		}
	}
	res.Log = append(res.Log, fmt.Sprintf("%s: %d messages, %d transmissions, processed %d times: %d lost, %d duplicates", g, messages, res.Sent, res.Processed, res.Lost, res.Duplicates))
	return res, nil
}
