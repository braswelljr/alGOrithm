package hashing

// CuckooMaxKicks bounds how many keys one insert may displace before the
// tables are rehashed.
const CuckooMaxKicks = 32

// Kick is one displacement during a cuckoo insert: Key was placed in Table at
// Slot, evicting Evicted (or nothing, when EvictedOK is false).
type Kick struct {
	Table, Slot, Key int
	Evicted          int
	EvictedOK        bool
}

// Cuckoo is cuckoo hashing (Pagh and Rodler, 2001): two tables and two hash
// functions, every key in one of its two slots, so a search reads at most two
// slots. An insert that finds both taken evicts the occupant, which moves to
// its other slot, possibly evicting another, like a cuckoo chick pushing eggs
// out of the nest; a cycle grows the tables and rehashes everything. The zero
// Cuckoo is not ready; use NewCuckoo.
type Cuckoo struct {
	tables [2][]int
	used   [2][]bool
	seed   uint64
	size   int
	// Rehashes counts the rehashes so far.
	Rehashes int
}

// NewCuckoo returns an empty cuckoo table with two tables of capacity slots.
func NewCuckoo(capacity int) *Cuckoo {
	c := &Cuckoo{seed: 1}
	c.alloc(max(capacity, 2))
	return c
}

func (c *Cuckoo) alloc(m int) {
	for t := 0; t < 2; t++ {
		c.tables[t], c.used[t] = make([]int, m), make([]bool, m)
	}
}

// Len is the number of keys. Time: O(1).
func (c *Cuckoo) Len() int { return c.size }

// Cap is the slots in each table.
func (c *Cuckoo) Cap() int { return len(c.tables[0]) }

// Slot returns the key in table t, slot i, and whether there is one.
func (c *Cuckoo) Slot(t, i int) (int, bool) { return c.tables[t][i], c.used[t][i] }

func (c *Cuckoo) pos(t, key int) int {
	h := (uint64(key) + c.seed*0x9E3779B97F4A7C15) * [2]uint64{0xff51afd7ed558ccd, 0xc4ceb9fe1a85ec53}[t]
	return int((h >> 29) % uint64(len(c.tables[t])))
}

// Contains reports whether key is in either of its two slots. Time: O(1).
func (c *Cuckoo) Contains(key int) bool {
	for t := 0; t < 2; t++ {
		if p := c.pos(t, key); c.used[t][p] && c.tables[t][p] == key {
			return true
		}
	}
	return false
}

// Insert adds key and returns the displacements it caused; added is false
// when key was already there. When the kicks run into a cycle the tables grow
// and every key is placed again with new hash functions (counted in Rehashes).
// Time: O(1) amortized expected.
func (c *Cuckoo) Insert(key int) (added bool, kicks []Kick) {
	if c.Contains(key) {
		return false, nil
	}
	kicks, homeless, ok := c.place(key)
	if !ok {
		c.rehash(homeless)
	}
	c.size++
	return true, kicks
}

// place puts key in, kicking occupants to their other table; when it gives up
// it returns the key left without a slot.
func (c *Cuckoo) place(key int) (kicks []Kick, homeless int, ok bool) {
	cur, t := key, 0
	for n := 0; n < CuckooMaxKicks; n++ {
		p := c.pos(t, cur)
		k := Kick{Table: t, Slot: p, Key: cur}
		if c.used[t][p] {
			k.Evicted, k.EvictedOK = c.tables[t][p], true
		}
		c.tables[t][p], c.used[t][p] = cur, true
		kicks = append(kicks, k)
		if !k.EvictedOK {
			return kicks, 0, true
		}
		cur, t = k.Evicted, 1-t
	}
	return kicks, cur, false
}

// rehash grows both tables and places every key, plus the homeless one,
// again, with new hash functions, until they all fit.
func (c *Cuckoo) rehash(homeless int) {
	keys := append(c.Keys(), homeless)
	for {
		c.Rehashes++
		c.seed++
		c.alloc(2 * len(c.tables[0]))
		all := true
		for _, k := range keys {
			if _, _, ok := c.place(k); !ok {
				all = false
				break
			}
		}
		if all {
			return
		}
	}
}

// Delete removes key and reports whether it was there. Time: O(1).
func (c *Cuckoo) Delete(key int) bool {
	for t := 0; t < 2; t++ {
		if p := c.pos(t, key); c.used[t][p] && c.tables[t][p] == key {
			c.used[t][p] = false
			c.size--
			return true
		}
	}
	return false
}

// Keys returns every key, table 0 first.
func (c *Cuckoo) Keys() []int {
	var out []int
	for t := 0; t < 2; t++ {
		for i, u := range c.used[t] {
			if u {
				out = append(out, c.tables[t][i])
			}
		}
	}
	return out
}
