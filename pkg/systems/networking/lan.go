package networking

import (
	"fmt"
	"strconv"
	"strings"
)

// lanHosts is the simulated LAN 10.0.0.0/24: the gateway at .1 and hosts at
// .2 to .9, each with a made-up MAC address.
func lanMAC(a Addr) (string, bool) {
	last := int(a & 0xFF)
	if a&0xFFFFFF00 != 0x0A000000 || last < 1 || last > 9 {
		return "", false
	}
	return fmt.Sprintf("02:00:00:00:00:%02x", last), true
}

// ARPResult is what an ARP run did.
type ARPResult struct {
	Steps      []string `json:"steps"`
	Timeline   Timeline `json:"timeline"`
	Broadcasts int      `json:"broadcasts"`
	Hits       int      `json:"hits"`
	Failures   int      `json:"failures"`
}

// ARP has host 10.0.0.2 send to each address in order on the LAN
// 10.0.0.0/24 (gateway 10.0.0.1, hosts .2 to .9). To put an IP packet in an
// Ethernet frame it needs the next hop's MAC address: for an address on the
// LAN that is the address itself, otherwise the gateway's. It looks in its
// ARP cache first; on a miss it broadcasts "who has X?" to every host, and
// the owner replies "X is at MAC" directly. Entries expire ttl sends after
// they were learned. An address nobody owns gets no reply.
//
// Time: O(sends). Space: O(cache).
func ARP(dests []string, ttl int) (ARPResult, error) {
	if len(dests) == 0 || len(dests) > 200 || ttl < 1 {
		return ARPResult{}, fmt.Errorf("%w: 1 to 200 destinations and a time to live of at least 1", ErrInput)
	}
	var res ARPResult
	type entry struct {
		mac     string
		learned int
	}
	cache := map[Addr]entry{}
	gateway := Addr(0x0A000001)
	for t, d := range dests {
		a, err := ParseAddr(d)
		if err != nil {
			return ARPResult{}, err
		}
		hop, via := a, ""
		if a&0xFFFFFF00 != 0x0A000000 {
			hop, via = gateway, " (off the LAN: via the gateway)"
		}
		if e, ok := cache[hop]; ok && t-e.learned < ttl {
			res.Hits++
			res.Steps = append(res.Steps, fmt.Sprintf("send to %s%s: cache has %s at %s", a, via, hop, e.mac))
			res.Timeline.add(t, "10.0.0.2", "cache hit "+hop.String(), "done")
			continue
		}
		res.Broadcasts++
		res.Timeline.add(t, "10.0.0.2", "who has "+hop.String()+"?", "running")
		mac, ok := lanMAC(hop)
		if !ok {
			res.Failures++
			res.Steps = append(res.Steps, fmt.Sprintf("send to %s: broadcast who has %s? no reply: the packet is dropped", a, hop))
			res.Timeline.add(t, "broadcast", "no reply", "blocked")
			continue
		}
		cache[hop] = entry{mac, t}
		res.Steps = append(res.Steps, fmt.Sprintf("send to %s%s: broadcast who has %s? reply: %s is at %s (cached)", a, via, hop, hop, mac))
		res.Timeline.add(t, hop.String(), "is at "+mac, "done")
	}
	return res, nil
}

// DHCPResult is what a DHCP run handed out.
type DHCPResult struct {
	Steps    []string          `json:"steps"`
	Timeline Timeline          `json:"timeline"`
	Leases   map[string]string `json:"leases"` // client -> address
}

// DHCP has clients join one after another and get an address from a server
// whose pool holds pool addresses from 192.168.1.100, by DORA: the client
// broadcasts DISCOVER (it has no address yet), the server OFFERs a free
// address, the client broadcasts REQUEST for it (so other servers know),
// and the server ACKs, starting a lease of lease ticks. With the pool
// empty the server stays silent and the client, after its timeout, gives
// itself a link-local address (169.254.x.y). Clients listed in release give
// their address back halfway, so a later client can get it. Renewal at half
// the lease is left out.
//
// Time: O(clients * pool). Space: O(pool).
func DHCP(clients, pool, lease int, release []int) (DHCPResult, error) {
	if clients < 1 || clients > 50 || pool < 0 || pool > 100 || lease < 1 {
		return DHCPResult{}, fmt.Errorf("%w: 1 to 50 clients, a pool of 0 to 100, a lease of at least 1", ErrInput)
	}
	res := DHCPResult{Leases: map[string]string{}}
	free := make([]bool, pool)
	for i := range free {
		free[i] = true
	}
	held := map[int]int{} // client -> pool index
	t := 0
	for c := 1; c <= clients; c++ {
		name := "client " + strconv.Itoa(c)
		if c == clients/2+1 {
			for _, r := range release {
				if i, ok := held[r]; ok {
					free[i] = true
					delete(held, r)
					res.Steps = append(res.Steps, fmt.Sprintf("t%d: client %d sends RELEASE for %s", t, r, poolAddr(i)))
					res.Timeline.add(t, "client "+strconv.Itoa(r), "RELEASE", "idle")
					delete(res.Leases, "client "+strconv.Itoa(r))
					t++
				}
			}
		}
		res.Timeline.add(t, name, "DISCOVER (broadcast)", "running")
		idx := -1
		for i, f := range free {
			if f {
				idx = i
				break
			}
		}
		if idx < 0 {
			res.Steps = append(res.Steps, fmt.Sprintf("t%d: %s DISCOVER: the pool is empty, no offer", t, name))
			ll := fmt.Sprintf("169.254.%d.%d", 1+c/250, 1+c%250)
			res.Timeline.add(t+2, name, "timeout: self-assigns "+ll, "blocked")
			res.Leases[name] = ll
			t += 3
			continue
		}
		a := poolAddr(idx)
		res.Timeline.add(t+1, "server", "OFFER "+a+" to "+name, "running")
		res.Timeline.add(t+2, name, "REQUEST "+a+" (broadcast)", "running")
		res.Timeline.add(t+3, "server", "ACK "+a+fmt.Sprintf(", lease %d", lease), "done")
		res.Steps = append(res.Steps, fmt.Sprintf("t%d: %s DISCOVER, OFFER %s, REQUEST, ACK (lease %d)", t, name, a, lease))
		free[idx], held[c] = false, idx
		res.Leases[name] = a
		t += 4
	}
	return res, nil
}

func poolAddr(i int) string { return "192.168.1." + strconv.Itoa(100+i) }

// NATMapping is one line of a NAT's translation table.
type NATMapping struct {
	Private string `json:"private"` // inside address:port
	Public  int    `json:"public"`  // the router's port standing for it
	Remote  string `json:"remote"`  // the outside address:port it talks to
}

// NAT translates packets through a home router with port address
// translation (PAT, "NAT overload"): many inside hosts share the public
// address 203.0.113.1. Each line of the script is a packet:
// "out 192.168.1.10:5000 93.184.215.14:443" (inside to outside) gets its
// source rewritten to 203.0.113.1 and a public port from 40000 up, recorded
// in the table; "in 93.184.215.14:443 40000" (outside to the public port)
// is let in and rewritten to the inside host only if the table has that
// port and the packet comes from the remote it was opened to; anything else
// is dropped. Timeouts of idle mappings are left out.
//
// Time: O(packets). Space: O(mappings).
func NAT(script string) ([]string, []NATMapping, error) {
	var steps []string
	var table []NATMapping
	byPrivate := map[string]int{} // private+remote -> index
	next := 40000
	n := 0
	for i, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if n++; n > 500 {
			return nil, nil, fmt.Errorf("%w: at most 500 packets", ErrInput)
		}
		bad := fmt.Errorf("%w: line %d: want \"out IP:PORT IP:PORT\" or \"in IP:PORT PUBLICPORT\"", ErrInput, i+1)
		if len(f) != 3 {
			return nil, nil, bad
		}
		switch f[0] {
		case "out":
			if !hostPort(f[1]) || !hostPort(f[2]) {
				return nil, nil, bad
			}
			key := f[1] + " " + f[2]
			idx, ok := byPrivate[key]
			if !ok {
				idx = len(table)
				table = append(table, NATMapping{Private: f[1], Public: next, Remote: f[2]})
				byPrivate[key] = idx
				next++
			}
			m := table[idx]
			note := "reuses"
			if !ok {
				note = "new mapping"
			}
			steps = append(steps, fmt.Sprintf("out %s -> %s becomes 203.0.113.1:%d -> %s (%s)", f[1], f[2], m.Public, f[2], note))
		case "in":
			port, err := strconv.Atoi(f[2])
			if err != nil || !hostPort(f[1]) {
				return nil, nil, bad
			}
			found := -1
			for j, m := range table {
				if m.Public == port {
					found = j
				}
			}
			switch {
			case found < 0:
				steps = append(steps, fmt.Sprintf("in %s -> 203.0.113.1:%d: no mapping, dropped", f[1], port))
			case table[found].Remote != f[1]:
				steps = append(steps, fmt.Sprintf("in %s -> 203.0.113.1:%d: mapped to %s, not %s, dropped", f[1], port, table[found].Remote, f[1]))
			default:
				steps = append(steps, fmt.Sprintf("in %s -> 203.0.113.1:%d becomes %s -> %s", f[1], port, f[1], table[found].Private))
			}
		default:
			return nil, nil, bad
		}
	}
	if n == 0 {
		return nil, nil, fmt.Errorf("%w: the script is empty", ErrInput)
	}
	return steps, table, nil
}

func hostPort(s string) bool {
	h, p, ok := strings.Cut(s, ":")
	if !ok {
		return false
	}
	port, err := strconv.Atoi(p)
	if err != nil || port < 1 || port > 65535 {
		return false
	}
	_, err = ParseAddr(h)
	return err == nil
}
