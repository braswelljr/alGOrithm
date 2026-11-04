package networking

import (
	"fmt"
	"slices"
)

// Decision is what a rate limiter did with one request.
type Decision struct {
	Arrival int  `json:"arrival"`
	Allowed bool `json:"allowed"`
	// Left is the tokens left after a token bucket's decision, or the queue
	// length after a leaky bucket's.
	Left int `json:"left"`
	// Served is the tick a leaky bucket sent the request on, -1 if dropped;
	// for a token bucket it is the arrival when allowed.
	Served int `json:"served"`
}

// TokenBucket decides each request of a trace of arrival ticks: the bucket
// holds up to capacity tokens, starts full, and gains one every interval
// ticks; a request takes a token or, with none left, is rejected. Bursts up
// to capacity pass at once; over time no more than one request per interval
// gets through.
//
// Time: O(n log n) to sort the arrivals. Space: O(n).
func TokenBucket(arrivals []int, capacity, interval int) ([]Decision, error) {
	if err := limiter(arrivals, capacity, interval); err != nil {
		return nil, err
	}
	a := slices.Clone(arrivals)
	slices.Sort(a)
	tokens, last := capacity, a[0]
	out := make([]Decision, len(a))
	for i, t := range a {
		// refill for the whole intervals since the last refill
		gained := (t - last) / interval
		tokens = min(capacity, tokens+gained)
		last += gained * interval
		if tokens == capacity {
			last = t // a full bucket does not bank time
		}
		d := Decision{Arrival: t, Served: -1}
		if tokens > 0 {
			tokens--
			d.Allowed, d.Served = true, t
		}
		d.Left = tokens
		out[i] = d
	}
	return out, nil
}

// LeakyBucket queues each request of a trace of arrival ticks, holding up to
// capacity waiting (the one being sent does not wait), and sends one every
// interval ticks, so what leaves is smooth
// however bursty what arrives; a request that finds the queue full is
// dropped. Unlike the token bucket it delays bursts instead of passing them.
//
// Time: O(n log n) to sort the arrivals. Space: O(n).
func LeakyBucket(arrivals []int, capacity, interval int) ([]Decision, error) {
	if err := limiter(arrivals, capacity, interval); err != nil {
		return nil, err
	}
	a := slices.Clone(arrivals)
	slices.Sort(a)
	out := make([]Decision, len(a))
	var queue []int // the ticks queued requests will be served on
	nextFree := a[0]
	for i, t := range a {
		for len(queue) > 0 && queue[0] <= t {
			queue = queue[1:]
		}
		d := Decision{Arrival: t, Served: -1}
		if len(queue) < capacity {
			at := max(t, nextFree)
			nextFree = at + interval
			queue = append(queue, at)
			d.Allowed, d.Served = true, at
		}
		d.Left = len(queue)
		out[i] = d
	}
	return out, nil
}

func limiter(arrivals []int, capacity, interval int) error {
	if len(arrivals) == 0 || len(arrivals) > 10_000 || capacity < 1 || interval < 1 {
		return fmt.Errorf("%w: give 1 to 10000 arrival ticks, a capacity and an interval of at least 1", ErrInput)
	}
	for _, t := range arrivals {
		if t < 0 {
			return fmt.Errorf("%w: arrival ticks are not negative", ErrInput)
		}
	}
	return nil
}
