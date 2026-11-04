package networking

import "fmt"

// Segment is one segment a sender put on the wire.
type Segment struct {
	Tick  int `json:"tick"`
	Bytes int `json:"bytes"`
}

// NagleResult is what a run of Nagle's algorithm sent.
type NagleResult struct {
	Timeline Timeline  `json:"timeline"`
	Segments []Segment `json:"segments"`
	// Delay is the total ticks bytes waited in the buffer, summed per byte.
	Delay int `json:"delay"`
}

// Nagle runs Nagle's algorithm (RFC 896) on writes[t], the bytes the
// application writes at tick t. With nagle on, the sender sends a full
// segment of mss bytes whenever it has one, but sends a smaller one only
// when no data is waiting for an ack; meanwhile small writes collect in the
// buffer. Each segment's ack comes back rtt ticks later. With nagle off,
// every write goes out at once, however small: many tiny segments, each
// paying a 40-byte header (the "small packet problem" of a remote terminal).
// After the last write the run continues until the buffer is empty. Delayed
// acks, which interact badly with Nagle, are left out.
//
// Time: O(ticks + bytes / mss). Space: O(segments).
func Nagle(writes []int, mss, rtt int, nagle bool) (NagleResult, error) {
	if len(writes) == 0 || len(writes) > 1000 || mss < 1 || mss > 65535 || rtt < 1 || rtt > 100 {
		return NagleResult{}, fmt.Errorf("%w: 1 to 1000 writes, an MSS of 1 to 65535 and a round trip of 1 to 100 ticks", ErrInput)
	}
	var res NagleResult
	buffered, unacked := 0, 0 // bytes waiting; segments in flight
	var acks []int            // ticks acks arrive
	var waiting []int         // tick each buffered byte group was written, for the delay
	var sizes []int
	send := func(t, n int) {
		res.Segments = append(res.Segments, Segment{Tick: t, Bytes: n})
		res.Timeline.add(t, "wire", fmt.Sprintf("segment %d bytes", n), "running")
		unacked++
		acks = append(acks, t+rtt)
		buffered -= n
		for n > 0 && len(sizes) > 0 { // charge the waiting time of the bytes sent
			take := min(n, sizes[0])
			res.Delay += take * (t - waiting[0])
			sizes[0] -= take
			n -= take
			if sizes[0] == 0 {
				sizes, waiting = sizes[1:], waiting[1:]
			}
		}
	}
	for t := 0; t < MaxTicks; t++ {
		for len(acks) > 0 && acks[0] <= t {
			acks = acks[1:]
			unacked--
			res.Timeline.add(t, "acks", "ack", "done")
		}
		if t < len(writes) && writes[t] > 0 {
			if writes[t] > 1<<20 {
				return NagleResult{}, fmt.Errorf("%w: a write is at most 1 MiB", ErrInput)
			}
			buffered += writes[t]
			sizes, waiting = append(sizes, writes[t]), append(waiting, t)
			res.Timeline.add(t, "app", fmt.Sprintf("write %d", writes[t]), "idle")
		}
		for buffered >= mss {
			send(t, mss)
		}
		if buffered > 0 && (!nagle || unacked == 0) {
			send(t, buffered)
		}
		if t >= len(writes) && buffered == 0 {
			return res, nil
		}
	}
	return res, fmt.Errorf("%w: not finished in %d ticks", ErrInput, MaxTicks)
}
