package networking

import (
	"fmt"
	"slices"
	"strings"
)

// PortRole is a switch port's role in the spanning tree.
type PortRole struct {
	Bridge string `json:"bridge"`
	Peer   string `json:"peer"` // the bridge at the other end of the link
	Role   string `json:"role"` // "root", "designated" or "blocked"
}

// SpanningTree runs the spanning tree protocol (Radia Perlman, 1985; IEEE
// 802.1D) on bridges joined by links with costs (t's routers are the
// bridges; a lower name stands for a lower bridge ID). The bridge with the
// lowest ID becomes the root. Every other bridge keeps the port with the
// cheapest path to the root as its root port (ties broken by the lower
// neighbor ID). On each link, the end with the cheaper path to the root
// (then the lower ID) is designated and forwards; a port that is neither
// root nor designated is blocked, which breaks every loop while keeping
// every bridge reachable. The exchange of BPDUs that reaches this result is
// summarized, not simulated message by message.
//
// Time: O(V^2 + E). Space: O(V + E).
func SpanningTree(t *Topology) (root string, cost map[string]int, ports []PortRole) {
	n := len(t.Routers)
	order := make([]int, n) // bridges by ID
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int { return strings.Compare(t.Routers[a], t.Routers[b]) })
	r := order[0]
	id := func(i int) string { return t.Routers[i] }
	dist := make([]int, n)
	done := make([]bool, n)
	for i := range dist {
		dist[i] = 1 << 30
	}
	dist[r] = 0
	for {
		u := -1
		for i := range dist {
			if !done[i] && dist[i] < 1<<30 && (u < 0 || dist[i] < dist[u]) {
				u = i
			}
		}
		if u < 0 {
			break
		}
		done[u] = true
		for v := 0; v < n; v++ {
			if c, ok := t.Cost(u, v); ok && dist[u]+c < dist[v] {
				dist[v] = dist[u] + c
			}
		}
	}
	rootPort := make([]int, n) // the neighbor each bridge's root port faces
	for b := 0; b < n; b++ {
		rootPort[b] = -1
		if b == r || dist[b] >= 1<<30 {
			continue
		}
		for _, nb := range order { // lower ID first, so ties go to it
			if c, ok := t.Cost(b, nb); ok && dist[nb]+c == dist[b] {
				rootPort[b] = nb
				break
			}
		}
	}
	cost = map[string]int{}
	for i := 0; i < n; i++ {
		cost[id(i)] = dist[i]
	}
	for _, a := range order {
		for _, b := range order {
			if _, ok := t.Cost(a, b); !ok {
				continue
			}
			role := "blocked"
			switch {
			case rootPort[a] == b:
				role = "root"
			case dist[a] < dist[b] || (dist[a] == dist[b] && strings.Compare(id(a), id(b)) < 0):
				role = "designated"
			}
			ports = append(ports, PortRole{Bridge: id(a), Peer: id(b), Role: role})
		}
	}
	return id(r), cost, ports
}

// BGPRound is the routes every autonomous system holds after one round.
type BGPRound struct {
	Paths map[string][]string `json:"paths"` // AS -> its AS path to the origin (nil when none)
	Notes []string            `json:"notes"`
}

// BGP runs path-vector routing as BGP does between autonomous systems (t's
// routers are the ASes): the origin announces its prefix; each round every
// AS tells its neighbors its best AS path with itself prepended; an AS
// rejects any path that already contains it, which is how a path vector
// prevents loops without knowing the whole map, and otherwise prefers the
// shortest AS path, then the lowest neighbor name. Real BGP chooses by
// local preference and policy first, which is left out, as are link costs.
//
// Time: O(rounds * E * V). Space: O(V^2).
func BGP(t *Topology, origin string) ([]BGPRound, error) {
	o := slices.Index(t.Routers, origin)
	if o < 0 {
		return nil, fmt.Errorf("%w: no AS named %q", ErrInput, origin)
	}
	n := len(t.Routers)
	paths := make([][]string, n)
	paths[o] = []string{origin}
	snap := func(notes []string) BGPRound {
		m := map[string][]string{}
		for i, p := range paths {
			m[t.Routers[i]] = slices.Clone(p)
		}
		return BGPRound{Paths: m, Notes: notes}
	}
	rounds := []BGPRound{snap([]string{origin + " announces its prefix"})}
	for round := 0; round < n+1; round++ {
		next := make([][]string, n)
		copy(next, paths)
		var notes []string
		changed := false
		for x := 0; x < n; x++ {
			if x == o {
				continue
			}
			var best []string
			for nb := 0; nb < n; nb++ {
				if _, ok := t.Cost(x, nb); !ok || paths[nb] == nil {
					continue
				}
				offer := paths[nb] // nb's path, which starts with nb itself
				if slices.Contains(offer, t.Routers[x]) {
					notes = append(notes, fmt.Sprintf("%s rejects %s from %s: it is already on the path (loop)", t.Routers[x], strings.Join(offer, " "), t.Routers[nb]))
					continue
				}
				cand := append([]string{t.Routers[x]}, offer...)
				if best == nil || len(cand) < len(best) || (len(cand) == len(best) && cand[1] < best[1]) {
					best = cand
				}
			}
			if !slices.Equal(best, paths[x]) {
				changed = true
				notes = append(notes, fmt.Sprintf("%s takes %s", t.Routers[x], strings.Join(best, " ")))
			}
			next[x] = best
		}
		paths = next
		rounds = append(rounds, snap(notes))
		if !changed {
			break // stable: this last round only shows the loops refused
		}
	}
	return rounds, nil
}
