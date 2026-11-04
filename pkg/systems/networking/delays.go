package networking

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// Switching compares the time to send a message of messageKB kilobytes
// across hops links of bandwidth Mbit/s, each with prop milliseconds of
// propagation delay, three ways. Circuit switching first sets up a path
// (setup ms), then streams the message straight through: setup + L/R +
// hops*prop. Message switching stores and forwards the whole message at
// each hop: hops*(L/R + prop). Packet switching cuts it into packets of
// packetKB and pipelines them, so while one packet crosses link 2 the next
// crosses link 1: (hops + packets - 1) * P/R + hops*prop. Headers and
// queueing are left out. Times are in milliseconds.
//
// Time: O(1). Space: O(1).
func Switching(messageKB, packetKB, hops int, bandwidth, prop, setup float64) (circuit, message, packet float64, packets int, err error) {
	if messageKB < 1 || packetKB < 1 || packetKB > messageKB || hops < 1 || hops > 100 || bandwidth <= 0 || prop < 0 || setup < 0 {
		return 0, 0, 0, 0, fmt.Errorf("%w: positive sizes, a packet no larger than the message, 1 to 100 hops", ErrInput)
	}
	ms := func(kb int) float64 { return float64(kb) * 8 * 1000 / (bandwidth * 1000) } // kilobits / (kbit per ms)
	packets = (messageKB + packetKB - 1) / packetKB
	h := float64(hops)
	circuit = setup + ms(messageKB) + h*prop
	message = h * (ms(messageKB) + prop)
	packet = float64(hops+packets-1)*ms(packetKB) + h*prop
	return circuit, message, packet, packets, nil
}

// BDP is the bandwidth-delay product of a link and what a window gets from it.
type BDP struct {
	Bytes      float64 `json:"bytes"`      // bits in flight to fill the pipe, in bytes
	Throughput float64 `json:"throughput"` // Mbit/s a window of the given size achieves
	Use        float64 `json:"use"`        // throughput / bandwidth
}

// BandwidthDelay computes the bandwidth-delay product of a link of
// bandwidth Mbit/s and a round trip of rtt milliseconds: the bytes that
// must be in flight to keep it busy. A sender limited to a window of
// windowKB kilobytes sends at most one window per round trip, so it gets
// min(bandwidth, window/rtt): on a long fat network the default 64 KB
// window fills a fraction of the link, which is why TCP has window
// scaling.
//
// Time: O(1). Space: O(1).
func BandwidthDelay(bandwidth, rtt float64, windowKB int) (BDP, error) {
	if bandwidth <= 0 || rtt <= 0 || windowKB < 1 {
		return BDP{}, fmt.Errorf("%w: a positive bandwidth, round trip and window", ErrInput)
	}
	b := BDP{Bytes: bandwidth * 1e6 / 8 * rtt / 1000}
	b.Throughput = math.Min(bandwidth, float64(windowKB)*1024*8/(rtt/1000)/1e6)
	b.Use = b.Throughput / bandwidth
	return b, nil
}

// Hop is one router on a path: its name and its one-way delay from the
// previous hop in milliseconds; Silent routers do not answer probes.
type Hop struct {
	Name   string
	Delay  int
	Silent bool
}

// ParseHops reads one hop a line, "name delay", or "name *" for a router
// that drops probes without answering; the last line is the destination.
func ParseHops(text string) ([]Hop, error) {
	var hops []Hop
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 2 {
			return nil, fmt.Errorf("%w: line %d: want \"name delay\" or \"name *\"", ErrInput, i+1)
		}
		h := Hop{Name: f[0]}
		if f[1] == "*" {
			h.Silent, h.Delay = true, 1
		} else {
			d, err := strconv.Atoi(f[1])
			if err != nil || d < 0 || d > 10_000 {
				return nil, fmt.Errorf("%w: line %d: a delay of 0 to 10000 ms", ErrInput, i+1)
			}
			h.Delay = d
		}
		hops = append(hops, h)
	}
	if len(hops) == 0 || len(hops) > 64 {
		return nil, fmt.Errorf("%w: 1 to 64 hops", ErrInput)
	}
	return hops, nil
}

// Traceroute finds the routers on a path by sending probes with time to
// live 1, 2, 3, ...: each router decrements the TTL and, when it reaches
// zero, drops the probe and answers ICMP Time Exceeded, revealing itself;
// the destination answers Port Unreachable instead. Each TTL sends probes
// probes; a probe's round trip is twice the delay to that hop plus a little
// seeded jitter. A silent router shows as "*". Real paths can change
// between probes and differ per direction, which is left out.
//
// Time: O(hops * probes). Space: O(hops).
func Traceroute(hops []Hop, probes int, seed int64) ([]string, error) {
	if probes < 1 || probes > 5 {
		return nil, fmt.Errorf("%w: 1 to 5 probes per hop", ErrInput)
	}
	r := rand.New(rand.NewSource(seed))
	var out []string
	oneWay := 0
	for ttl, h := range hops {
		oneWay += h.Delay
		line := fmt.Sprintf("%2d  %-14s", ttl+1, h.Name)
		if h.Silent {
			line = fmt.Sprintf("%2d  %-14s", ttl+1, "*")
		}
		for p := 0; p < probes; p++ {
			if h.Silent {
				line += "  *"
				continue
			}
			line += fmt.Sprintf("  %d ms", 2*oneWay+r.Intn(3))
		}
		why := "ICMP time exceeded"
		if ttl == len(hops)-1 {
			why = "ICMP port unreachable: the destination"
		}
		if h.Silent {
			why = "no answer"
		}
		out = append(out, line+"  ("+why+")")
	}
	return out, nil
}

// Fragment is one IPv4 fragment as it arrives.
type Fragment struct {
	Offset int  `json:"offset"` // payload offset in bytes (the header stores it divided by 8)
	Length int  `json:"length"` // payload bytes
	More   bool `json:"more"`   // the more-fragments flag
}

// Fragments sends a datagram with payload bytes across links whose MTUs are
// mtus, in order. A packet bigger than a link's MTU is cut into fragments
// whose payloads are multiples of 8 bytes (the offset field counts 8-byte
// units), each with its own 20-byte header; a fragment can be fragmented
// again on a smaller link. Only the destination reassembles. With df (don't
// fragment) set, the first link that is too small drops the packet and
// answers ICMP "fragmentation needed" with its MTU, which path MTU
// discovery uses. It returns the fragments that arrive and the steps.
//
// Time: O(fragments * links). Space: O(fragments).
func Fragments(payload int, mtus []int, df bool) ([]Fragment, []string, error) {
	if payload < 1 || payload > 65515 || len(mtus) == 0 || len(mtus) > 20 {
		return nil, nil, fmt.Errorf("%w: a payload of 1 to 65515 bytes over 1 to 20 links", ErrInput)
	}
	const header = 20
	frags := []Fragment{{Offset: 0, Length: payload, More: false}}
	var steps []string
	for i, mtu := range mtus {
		if mtu < 68 || mtu > 65535 {
			return nil, nil, fmt.Errorf("%w: an MTU is 68 to 65535", ErrInput)
		}
		var next []Fragment
		for _, f := range frags {
			if f.Length+header <= mtu {
				next = append(next, f)
				continue
			}
			if df {
				steps = append(steps, fmt.Sprintf("link %d (MTU %d): %d bytes is too big and DF is set: dropped, ICMP fragmentation needed (MTU %d)", i+1, mtu, f.Length+header, mtu))
				return nil, steps, nil
			}
			chunk := (mtu - header) / 8 * 8
			for off := 0; off < f.Length; off += chunk {
				n := min(chunk, f.Length-off)
				next = append(next, Fragment{Offset: f.Offset + off, Length: n, More: f.More || off+n < f.Length})
			}
		}
		steps = append(steps, fmt.Sprintf("link %d (MTU %d): %d fragments", i+1, mtu, len(next)))
		frags = next
	}
	return frags, steps, nil
}

// RTOSample is the retransmission timer after one round-trip sample.
type RTOSample struct {
	Sample int     `json:"sample"` // ms; negative marks a retransmitted segment's sample
	SRTT   float64 `json:"srtt"`
	RTTVar float64 `json:"rttvar"`
	RTO    float64 `json:"rto"`
}

// RTO estimates TCP's retransmission timeout from round-trip samples (ms)
// as RFC 6298 does (Van Jacobson and Michael Karels): the first sample sets
// SRTT = R and RTTVAR = R/2; each later one sets RTTVAR = 3/4 RTTVAR +
// 1/4 |SRTT - R| and SRTT = 7/8 SRTT + 1/8 R; RTO = SRTT + 4 RTTVAR, at
// least minimum ms. A negative sample stands for a timeout of a
// retransmitted segment: by Karn's algorithm its round trip is ambiguous, so
// it is not used, and the RTO doubles (exponential backoff) instead.
//
// Time: O(samples). Space: O(samples).
func RTO(samples []int, minimum int) ([]RTOSample, error) {
	if len(samples) == 0 || len(samples) > 1000 || minimum < 0 {
		return nil, fmt.Errorf("%w: 1 to 1000 samples and a minimum RTO", ErrInput)
	}
	var out []RTOSample
	var srtt, rttvar, rto float64
	first := true
	for _, s := range samples {
		if s < 0 {
			if first {
				rto = 1000 // RFC 6298's initial RTO
			}
			rto *= 2
			out = append(out, RTOSample{Sample: s, SRTT: srtt, RTTVar: rttvar, RTO: rto})
			continue
		}
		r := float64(s)
		if first {
			srtt, rttvar, first = r, r/2, false
		} else {
			rttvar = 0.75*rttvar + 0.25*math.Abs(srtt-r)
			srtt = 0.875*srtt + 0.125*r
		}
		rto = math.Max(srtt+4*rttvar, float64(minimum))
		out = append(out, RTOSample{Sample: s, SRTT: srtt, RTTVar: rttvar, RTO: rto})
	}
	return out, nil
}
