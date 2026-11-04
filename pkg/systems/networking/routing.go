package networking

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Infinity is the distance that means unreachable, as in RIP: 16 hops.
const Infinity = 16

// Topology is a network of routers joined by links with a cost.
type Topology struct {
	Routers []string
	cost    map[[2]int]int
}

// ParseTopology reads one link a line, "A B [cost]" (cost 1 when left out);
// links work both ways. Blank lines and lines starting with # are skipped.
func ParseTopology(text string) (*Topology, error) {
	t := &Topology{cost: map[[2]int]int{}}
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) < 2 || len(f) > 3 {
			return nil, fmt.Errorf("%w: line %d: want \"A B [cost]\"", ErrInput, i+1)
		}
		c := 1
		if len(f) == 3 {
			v, err := strconv.Atoi(f[2])
			if err != nil || v < 1 || v >= Infinity {
				return nil, fmt.Errorf("%w: line %d: a cost is 1 to %d", ErrInput, i+1, Infinity-1)
			}
			c = v
		}
		if f[0] == f[1] {
			return nil, fmt.Errorf("%w: line %d: a link joins two routers", ErrInput, i+1)
		}
		a, b := t.router(f[0]), t.router(f[1])
		t.cost[[2]int{a, b}], t.cost[[2]int{b, a}] = c, c
	}
	if len(t.Routers) < 2 || len(t.Routers) > 26 {
		return nil, fmt.Errorf("%w: 2 to 26 routers", ErrInput)
	}
	return t, nil
}

func (t *Topology) router(name string) int {
	if i := slices.Index(t.Routers, name); i >= 0 {
		return i
	}
	t.Routers = append(t.Routers, name)
	return len(t.Routers) - 1
}

// Cost returns the cost of the link a-b and whether there is one.
func (t *Topology) Cost(a, b int) (int, bool) {
	c, ok := t.cost[[2]int{a, b}]
	return c, ok
}

// Cut removes the link between the routers named a and b.
func (t *Topology) Cut(a, b string) error {
	i, j := slices.Index(t.Routers, a), slices.Index(t.Routers, b)
	if _, ok := t.Cost(i, j); i < 0 || j < 0 || !ok {
		return fmt.Errorf("%w: there is no link %s-%s", ErrInput, a, b)
	}
	delete(t.cost, [2]int{i, j})
	delete(t.cost, [2]int{j, i})
	return nil
}

// DVRound is every router's distance vector after one exchange.
type DVRound struct {
	Dist    [][]int  `json:"dist"`    // Dist[i][j]: router i's distance to j, Infinity if unreachable
	Next    [][]int  `json:"next"`    // Next[i][j]: the neighbor i forwards to j through, -1 for none
	Changes []string `json:"changes"` // what changed, in words
}

// DistanceVector runs distance-vector routing (distributed Bellman-Ford, as
// RIP does): in each round every router sends its vector to its neighbors
// and takes, for each destination, the neighbor that offers the cheapest
// link cost plus distance. It runs until no vector changes, then, if cut
// names a link ("A B"), cuts it and runs again: without poisoned reverse two
// routers can each believe the other reaches the lost destination, and count
// up to Infinity together (count to infinity). With poison true, a router
// tells the neighbor it routes through that the destination is unreachable
// through it, which stops loops of two. Rounds are synchronous and at most
// maxRounds run in each phase.
//
// Time: O(rounds * V^2 * degree). Space: O(rounds * V^2).
func DistanceVector(t *Topology, cut string, poison bool, maxRounds int) (rounds []DVRound, cutAt int, err error) {
	var link []string
	if cut = strings.TrimSpace(cut); cut != "" {
		if link = strings.Fields(cut); len(link) != 2 {
			return nil, 0, fmt.Errorf("%w: name the link to cut as \"A B\"", ErrInput)
		}
	}
	dv := newDV(t, poison)
	rounds = append(rounds, dv.snapshot([]string{"every router knows its own links"}))
	rounds = dv.converge(rounds, maxRounds)
	if link == nil {
		return rounds, -1, nil
	}
	if err := t.Cut(link[0], link[1]); err != nil {
		return nil, 0, err
	}
	dv.lose(slices.Index(t.Routers, link[0]), slices.Index(t.Routers, link[1]))
	cutAt = len(rounds)
	rounds = append(rounds, dv.snapshot([]string{"link " + link[0] + "-" + link[1] + " is cut"}))
	return dv.converge(rounds, maxRounds), cutAt, nil
}

// distanceVector is every router's vector: its distance to each destination
// and the neighbor it goes through (-1 for none).
type distanceVector struct {
	t          *Topology
	poison     bool
	dist, next [][]int
}

// newDV starts each router knowing only itself and its own links.
func newDV(t *Topology, poison bool) *distanceVector {
	n := len(t.Routers)
	dv := &distanceVector{t: t, poison: poison, dist: make([][]int, n), next: make([][]int, n)}
	for i := range dv.dist {
		dv.dist[i], dv.next[i] = make([]int, n), make([]int, n)
		for j := range dv.dist[i] {
			dv.dist[i][j], dv.next[i][j] = Infinity, -1
			if i == j {
				dv.dist[i][j], dv.next[i][j] = 0, i
			} else if c, ok := t.Cost(i, j); ok {
				dv.dist[i][j], dv.next[i][j] = c, j
			}
		}
	}
	return dv
}

func (dv *distanceVector) snapshot(changes []string) DVRound {
	return DVRound{Dist: copyGrid(dv.dist), Next: copyGrid(dv.next), Changes: changes}
}

// converge runs synchronous rounds until no vector changes, at most max.
func (dv *distanceVector) converge(rounds []DVRound, maxRounds int) []DVRound {
	for r := 0; r < maxRounds; r++ {
		changes := dv.round()
		if len(changes) == 0 {
			break
		}
		rounds = append(rounds, dv.snapshot(changes))
	}
	return rounds
}

// round has every router take the best offer for every destination from the
// vectors its neighbors sent, and returns what changed.
func (dv *distanceVector) round() []string {
	nd, nn := copyGrid(dv.dist), copyGrid(dv.next)
	var changes []string
	for i := range dv.dist {
		for j := range dv.dist {
			if i == j {
				continue
			}
			best, via := dv.best(i, j)
			nd[i][j], nn[i][j] = best, via
			if best != dv.dist[i][j] || via != dv.next[i][j] {
				changes = append(changes, fmt.Sprintf("%s to %s: %s", dv.t.Routers[i], dv.t.Routers[j], distWord(best, via, dv.t.Routers)))
			}
		}
	}
	if len(changes) > 0 {
		dv.dist, dv.next = nd, nn
	}
	return changes
}

// best is router i's cheapest route to j through a neighbor: the link cost
// plus what that neighbor offers. With poison, a neighbor that routes to j
// through i offers Infinity.
func (dv *distanceVector) best(i, j int) (dist, via int) {
	dist, via = Infinity, -1
	for k := range dv.dist {
		c, ok := dv.t.Cost(i, k)
		if !ok {
			continue
		}
		offer := dv.dist[k][j]
		if dv.poison && dv.next[k][j] == i {
			offer = Infinity // k routes to j through i, so it says it cannot
		}
		if d := min(c+offer, Infinity); d < dist {
			dist, via = d, k
		}
	}
	if dist >= Infinity {
		via = -1
	}
	return dist, via
}

// lose drops every route over the cut link a-b at both ends at once.
func (dv *distanceVector) lose(a, b int) {
	for _, end := range [][2]int{{a, b}, {b, a}} {
		for j := range dv.dist {
			if dv.next[end[0]][j] == end[1] {
				dv.dist[end[0]][j], dv.next[end[0]][j] = Infinity, -1
			}
		}
	}
}

func distWord(d, via int, names []string) string {
	if via < 0 {
		return "unreachable"
	}
	return fmt.Sprintf("%d via %s", d, names[via])
}

func copyGrid(g [][]int) [][]int {
	out := make([][]int, len(g))
	for i := range g {
		out[i] = slices.Clone(g[i])
	}
	return out
}

// Route is one line of a forwarding table.
type Route struct {
	To   string `json:"to"`
	Next string `json:"next"` // the neighbor to forward to, "" when unreachable or local
	Cost int    `json:"cost"`
}

// LinkState runs link-state routing (as OSPF does): every router floods its
// links, so each has the whole map, and each runs Dijkstra's algorithm from
// itself to build its forwarding table. It returns one table per router, in
// the order of t.Routers. The flooding itself is assumed done.
//
// Time: O(V * V^2) with the simple Dijkstra used here. Space: O(V^2).
func LinkState(t *Topology) [][]Route {
	n := len(t.Routers)
	tables := make([][]Route, n)
	for s := 0; s < n; s++ {
		dist, first := make([]int, n), make([]int, n)
		done := make([]bool, n)
		for i := range dist {
			dist[i], first[i] = 1<<30, -1
		}
		dist[s] = 0
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
					first[v] = first[u]
					if u == s {
						first[v] = v
					}
				}
			}
		}
		for j := 0; j < n; j++ {
			r := Route{To: t.Routers[j], Cost: dist[j]}
			if dist[j] >= 1<<30 {
				r.Cost = -1
			} else if first[j] >= 0 {
				r.Next = t.Routers[first[j]]
			}
			tables[s] = append(tables[s], r)
		}
	}
	return tables
}
