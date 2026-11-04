package networking

import (
	"fmt"
	"strconv"
	"strings"
)

// Addr is an IPv4 address as a 32-bit number.
type Addr uint32

// ParseAddr reads a dotted-quad address ("192.168.1.10").
func ParseAddr(s string) (Addr, error) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("%w: %q is not a dotted-quad address", ErrInput, s)
	}
	var a Addr
	for _, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > 255 {
			return 0, fmt.Errorf("%w: %q: each part is 0 to 255", ErrInput, s)
		}
		a = a<<8 | Addr(v)
	}
	return a, nil
}

// String writes a in dotted-quad form.
func (a Addr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}

// Prefix is a CIDR block: the addresses whose first Bits bits match Addr.
type Prefix struct {
	Addr Addr `json:"addr"`
	Bits int  `json:"bits"`
}

// ParsePrefix reads "10.1.0.0/16"; the host bits are cleared, so
// "10.1.2.3/16" is 10.1.0.0/16.
func ParsePrefix(s string) (Prefix, error) {
	addr, bits, ok := strings.Cut(strings.TrimSpace(s), "/")
	if !ok {
		return Prefix{}, fmt.Errorf("%w: %q has no /length", ErrInput, s)
	}
	a, err := ParseAddr(addr)
	if err != nil {
		return Prefix{}, err
	}
	b, err := strconv.Atoi(bits)
	if err != nil || b < 0 || b > 32 {
		return Prefix{}, fmt.Errorf("%w: %q: the length is 0 to 32", ErrInput, s)
	}
	p := Prefix{Bits: b}
	p.Addr = a & p.Mask()
	return p, nil
}

// Mask is the prefix's network mask (255.255.255.0 for /24).
func (p Prefix) Mask() Addr {
	if p.Bits == 0 {
		return 0
	}
	return ^Addr(0) << (32 - p.Bits)
}

// Network is the first address of the block, which names the block.
func (p Prefix) Network() Addr { return p.Addr }

// Broadcast is the last address of the block, which reaches every host on it.
func (p Prefix) Broadcast() Addr { return p.Addr | ^p.Mask() }

// Hosts is how many addresses can be given to hosts: all but the network
// and broadcast addresses, except that /31 (a point-to-point link, RFC 3021)
// and /32 (one host) use every address.
func (p Prefix) Hosts() int {
	size := 1 << (32 - p.Bits)
	if p.Bits >= 31 {
		return size
	}
	return size - 2
}

// Contains reports whether a is in the block.
func (p Prefix) Contains(a Addr) bool { return a&p.Mask() == p.Addr }

// String writes the prefix in CIDR form.
func (p Prefix) String() string { return fmt.Sprintf("%s/%d", p.Addr, p.Bits) }

// Split divides p into n equal subnets; n must be a power of two that leaves
// at least one address in each.
//
// Time: O(n). Space: O(n).
func Split(p Prefix, n int) ([]Prefix, error) {
	extra := 0
	for 1<<extra < n {
		extra++
	}
	if n < 1 || 1<<extra != n || p.Bits+extra > 32 || n > 4096 {
		return nil, fmt.Errorf("%w: split /%d into a power of two subnets, at most /32 and 4096 of them", ErrInput, p.Bits)
	}
	out := make([]Prefix, n)
	step := Addr(1) << (32 - p.Bits - extra)
	for i := range out {
		out[i] = Prefix{Addr: p.Addr + Addr(i)*step, Bits: p.Bits + extra}
	}
	return out, nil
}

// RouteEntry is one line of a routing table: a prefix and where to send it.
type RouteEntry struct {
	Prefix Prefix `json:"prefix"`
	Hop    string `json:"hop"`
}

// RoutingTable finds routes by longest-prefix match with a binary trie: one
// level per address bit, so a lookup reads at most 32 nodes whatever the
// table's size.
type RoutingTable struct {
	root *trieNode
	n    int
}

type trieNode struct {
	kids  [2]*trieNode
	route *RouteEntry
}

// ParseRoutes reads one route a line, "10.0.0.0/8 hop"; "0.0.0.0/0 hop" is
// the default route.
func ParseRoutes(text string) (*RoutingTable, error) {
	t := &RoutingTable{root: &trieNode{}}
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 2 {
			return nil, fmt.Errorf("%w: line %d: want \"prefix/len hop\"", ErrInput, i+1)
		}
		p, err := ParsePrefix(f[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		t.Add(RouteEntry{Prefix: p, Hop: f[1]})
	}
	if t.n == 0 {
		return nil, fmt.Errorf("%w: the table has no routes", ErrInput)
	}
	return t, nil
}

// Add puts a route in the table, replacing one for the same prefix.
//
// Time: O(bits). Space: O(bits) new nodes.
func (t *RoutingTable) Add(r RouteEntry) {
	n := t.root
	for i := 0; i < r.Prefix.Bits; i++ {
		b := r.Prefix.Addr >> (31 - i) & 1
		if n.kids[b] == nil {
			n.kids[b] = &trieNode{}
		}
		n = n.kids[b]
	}
	if n.route == nil {
		t.n++
	}
	n.route = &r
}

// Lookup returns the route with the longest prefix containing a, the
// prefixes it passed on the way (shortest first), and false when no route
// matches (no default route).
//
// Time: O(32). Space: O(32).
func (t *RoutingTable) Lookup(a Addr) (best RouteEntry, matched []Prefix, ok bool) {
	n := t.root
	for i := 0; n != nil; i++ {
		if n.route != nil {
			best, ok = *n.route, true
			matched = append(matched, n.route.Prefix)
		}
		if i == 32 {
			break
		}
		n = n.kids[a>>(31-i)&1]
	}
	return best, matched, ok
}
