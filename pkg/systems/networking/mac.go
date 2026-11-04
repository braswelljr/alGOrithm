package networking

import (
	"fmt"
	"math/rand"
	"strconv"
)

// MACResult is what a medium-access simulation did.
type MACResult struct {
	Timeline   Timeline `json:"timeline"`
	Slots      int      `json:"slots"`
	Delivered  int      `json:"delivered"`
	Collisions int      `json:"collisions"`
	Dropped    int      `json:"dropped"`
}

// CSMACD simulates classic shared Ethernet in slots: stations each with
// frames to send, all ready at slot 0. A station whose backoff has run out
// transmits; alone, its frame gets through (one slot here), but two or more
// at once collide, every one detects it, and each waits a random number of
// slots in [0, 2^k - 1] after its k-th collision (binary exponential
// backoff, k capped at 10), dropping the frame after 16 tries. The randomness
// is seeded. Propagation delay, the jam signal and frame lengths are left
// out, so a collision costs one slot.
//
// Time: O(slots * stations). Space: O(stations + slots).
func CSMACD(stations, frames int, seed int64) (MACResult, error) {
	if stations < 1 || stations > 20 || frames < 1 || frames > 100 {
		return MACResult{}, fmt.Errorf("%w: 1 to 20 stations with 1 to 100 frames each", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	left := make([]int, stations)
	wait := make([]int, stations)
	tries := make([]int, stations)
	for i := range left {
		left[i] = frames
	}
	var res MACResult
	name := func(i int) string { return "station " + strconv.Itoa(i+1) }
	for t := 0; t < MaxTicks; t++ {
		var ready []int
		pending := false
		for i := range left {
			if left[i] == 0 {
				continue
			}
			pending = true
			if wait[i] == 0 {
				ready = append(ready, i)
			} else {
				wait[i]--
			}
		}
		if !pending {
			res.Slots = t
			return res, nil
		}
		switch len(ready) {
		case 0:
		case 1:
			i := ready[0]
			left[i]--
			tries[i] = 0
			res.Delivered++
			res.Timeline.add(t, name(i), "send", "done")
		default:
			res.Collisions++
			for _, i := range ready {
				tries[i]++
				if tries[i] > 16 {
					left[i]--
					tries[i] = 0
					res.Dropped++
					res.Timeline.add(t, name(i), "collide: drop after 16 tries", "blocked")
					continue
				}
				w := r.Intn(1 << min(tries[i], 10))
				wait[i] = w
				res.Timeline.add(t, name(i), fmt.Sprintf("collide %d: back off %d", tries[i], w), "blocked")
			}
		}
	}
	return res, fmt.Errorf("%w: not finished in %d slots", ErrInput, MaxTicks)
}

// CSMACA simulates Wi-Fi's collision avoidance with the hidden-terminal
// problem: stations A and C both send frames (length slots each) to the
// access point B, but are too far apart to hear each other, so carrier
// sense tells each the air is free while the other is sending, and their
// frames collide at B, which then sends no ACK. A sender that gets no ACK
// doubles its contention window (starting at 4 slots) and backs off a
// random number of slots in it (seeded). With rtscts, a sender first sends
// a one-slot RTS; B answers with a CTS that both stations hear, reserving
// the air (the NAV) for the frame and its ACK, so only the short RTS can
// collide. Interframe spaces are folded into the slots.
//
// Time: O(slots). Space: O(slots).
func CSMACA(frames, length int, rtscts bool, seed int64) (MACResult, error) {
	if frames < 1 || frames > 50 || length < 1 || length > 20 {
		return MACResult{}, fmt.Errorf("%w: 1 to 50 frames each of 1 to 20 slots", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	type station struct {
		name                      string
		left, cw, backoff         int
		sendLeft, kind, nav, hold int // kind: 1 data, 2 RTS; hold: slots to wait before the data
		corrupt                   bool
	}
	st := []*station{{name: "A", left: frames, cw: 4}, {name: "C", left: frames, cw: 4}}
	for _, s := range st {
		s.backoff = r.Intn(s.cw)
	}
	var res MACResult
	apBusyUntil := -1 // B is sending a CTS or an ACK until this slot
	fail := func(t int, s *station, what string) {
		res.Collisions++
		s.cw = min(2*s.cw, 1024)
		s.backoff = r.Intn(s.cw)
		res.Timeline.add(t, s.name, what+fmt.Sprintf(": no answer, window %d, back off %d", s.cw, s.backoff), "blocked")
	}
	for t := 0; t < MaxTicks; t++ {
		if st[0].left == 0 && st[1].left == 0 && st[0].sendLeft == 0 && st[1].sendLeft == 0 {
			res.Slots = t
			return res, nil
		}
		// who starts sending this slot
		for _, s := range st {
			if s.sendLeft > 0 || s.left == 0 {
				continue
			}
			if t <= apBusyUntil || t < s.nav { // only B can be heard; NAV set by a CTS
				continue
			}
			if s.backoff > 0 {
				s.backoff--
				continue
			}
			s.corrupt = false
			if rtscts {
				s.kind, s.sendLeft = 2, 1
				res.Timeline.add(t, s.name, "RTS", "running")
			} else {
				s.kind, s.sendLeft = 1, length
			}
		}
		// collisions at B: both on the air in the same slot
		onAir := func(s *station) bool { return s.sendLeft > 0 && s.hold == 0 }
		if onAir(st[0]) && onAir(st[1]) {
			st[0].corrupt, st[1].corrupt = true, true
		}
		for _, s := range st {
			if s.sendLeft == 0 {
				continue
			}
			if s.hold > 0 {
				s.hold--
				continue
			}
			if s.kind == 1 {
				res.Timeline.add(t, s.name, "data", "running")
			}
			s.sendLeft--
			if s.sendLeft > 0 {
				continue
			}
			switch {
			case s.corrupt && s.kind == 1:
				res.Timeline.add(t, "B (access point)", "garbled: hidden terminals collided", "blocked")
				fail(t, s, "data lost")
			case s.corrupt:
				fail(t, s, "RTS collided")
			case s.kind == 2: // CTS: everyone holds off for the frame and its ACK
				res.Timeline.add(t+1, "B (access point)", "CTS to "+s.name, "running")
				apBusyUntil = t + 1
				for _, o := range st {
					if o != s {
						o.nav = t + 2 + length + 1
					}
				}
				s.kind, s.sendLeft, s.hold, s.corrupt = 1, length, 1, false // the data follows the CTS
			default:
				res.Timeline.add(t+1, "B (access point)", "ACK to "+s.name, "done")
				apBusyUntil = t + 1
				s.left--
				res.Delivered++
				s.cw = 4
				s.backoff = r.Intn(s.cw)
			}
		}
	}
	return res, fmt.Errorf("%w: not finished in %d slots", ErrInput, MaxTicks)
}
