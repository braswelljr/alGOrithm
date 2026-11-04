package hashing

// RobinHood is Robin Hood hashing (Celis, 1986): linear probing where an
// inserted key that has traveled further from its home slot than the slot's
// occupant takes the slot, and the occupant moves on. Taking from the rich
// (keys near home) to give to the poor keeps every key's probe distance close
// to the average, so the worst search is short. Deletion shifts the following
// keys back (backward-shift deletion), so no tombstones are needed. The zero
// RobinHood is not ready; use NewRobinHood.
type RobinHood struct {
	keys []int
	dist []int // probe distance from home; -1 for an empty slot
	size int
}

// NewRobinHood returns an empty table of capacity slots (at least 4).
func NewRobinHood(capacity int) *RobinHood {
	r := &RobinHood{}
	r.alloc(max(capacity, 4))
	return r
}

func (r *RobinHood) alloc(m int) {
	r.keys, r.dist = make([]int, m), make([]int, m)
	for i := range r.dist {
		r.dist[i] = -1
	}
}

func (r *RobinHood) home(key int) int { return hash1(key, len(r.keys)) }

// Len is the number of keys; Cap the slots. Time: O(1).
func (r *RobinHood) Len() int { return r.size }

// Cap is the number of slots.
func (r *RobinHood) Cap() int { return len(r.keys) }

// Slot returns the key in slot i and its probe distance, or -1 when empty.
func (r *RobinHood) Slot(i int) (key, dist int) { return r.keys[i], r.dist[i] }

// MaxDistance is the longest probe distance in the table.
func (r *RobinHood) MaxDistance() int {
	m := 0
	for _, d := range r.dist {
		m = max(m, d)
	}
	return m
}

// Search reports whether key is in the table and the slots probed. It stops
// as soon as it meets a key closer to home than the search has traveled.
// Time: O(1) expected.
func (r *RobinHood) Search(key int) (bool, []int) {
	var probed []int
	m := len(r.keys)
	for d, s := 0, r.home(key); d < m; d, s = d+1, (s+1)%m {
		probed = append(probed, s)
		if r.dist[s] < d { // empty, or a richer key: key would have taken this slot
			return false, probed
		}
		if r.keys[s] == key {
			return true, probed
		}
	}
	return false, probed
}

// Insert adds key and returns the slots it touched; added is false when key
// was already there. Time: O(1) expected, O(n) for a resize.
func (r *RobinHood) Insert(key int) (added bool, probed []int) {
	if ok, p := r.Search(key); ok {
		return false, p
	}
	if r.size+1 > len(r.keys)*9/10 {
		old := r.Keys()
		r.alloc(2 * len(r.keys))
		r.size = 0
		for _, k := range old {
			r.Insert(k)
		}
	}
	m := len(r.keys)
	cur, d := key, 0
	for s := r.home(key); ; s = (s + 1) % m {
		probed = append(probed, s)
		if r.dist[s] < 0 {
			r.keys[s], r.dist[s] = cur, d
			r.size++
			return true, probed
		}
		if r.dist[s] < d { // the occupant is richer: swap and carry it on
			r.keys[s], cur = cur, r.keys[s]
			r.dist[s], d = d, r.dist[s]
		}
		d++
	}
}

// Delete removes key and reports whether it was there, shifting later keys of
// the same run back one slot. Time: O(1) expected.
func (r *RobinHood) Delete(key int) (bool, []int) {
	ok, probed := r.Search(key)
	if !ok {
		return false, probed
	}
	m := len(r.keys)
	s := probed[len(probed)-1]
	for {
		next := (s + 1) % m
		if r.dist[next] <= 0 {
			r.dist[s] = -1
			break
		}
		r.keys[s], r.dist[s] = r.keys[next], r.dist[next]-1
		s = next
	}
	r.size--
	return true, probed
}

// Keys returns the keys in slot order.
func (r *RobinHood) Keys() []int {
	var out []int
	for i, d := range r.dist {
		if d >= 0 {
			out = append(out, r.keys[i])
		}
	}
	return out
}
