package design

import (
	"fmt"
	"math/rand"
)

// Overload is what a queue does when work arrives faster than it is served.
type Overload int

// The ways to handle overload.
const (
	Unbounded    Overload = iota // queue everything
	LoadShedding                 // drop arrivals once the queue is full
	PushBack                     // tell the producer to slow down once the queue is full
)

// BackpressureConfig is a burst of work into a queue.
type BackpressureConfig struct {
	ServicePerTick int      // requests the service finishes each tick
	ArrivePerTick  int      // requests arriving each tick normally
	BurstPerTick   int      // requests arriving each tick during the burst
	BurstTicks     int      // how long the burst lasts, from tick 0
	Ticks          int      // how long to run
	Limit          int      // queue length that counts as full
	Mode           Overload // what to do when full
}

// Backpressure models a service hit by a burst bigger than it can serve. An
// unbounded queue accepts everything, and latency (queue length / service
// rate) grows with every tick of the burst until requests time out long
// after the burst is gone. Load shedding drops arrivals beyond Limit, so the
// requests it does take stay fast. Backpressure refuses them at the edge (a
// 429 or a full channel) so the producer waits and retries later, moving the
// queue to where it can be bounded. It reports the queue and latency per tick.
//
// Time: O(Ticks). Space: O(Ticks).
func Backpressure(c BackpressureConfig) (Report, error) {
	if err := check(c.ServicePerTick >= 1 && c.ArrivePerTick >= 0 && c.BurstPerTick >= 0 && c.BurstTicks >= 0 && c.Ticks >= 1 && c.Ticks <= 10_000 &&
		c.Limit >= 1 && c.Mode >= Unbounded && c.Mode <= PushBack, "positive service rate, limit and ticks (up to 10,000), a known mode"); err != nil {
		return Report{}, err
	}
	var r Report
	t := r.table("arrivals", "queue", "dropped", "pushed back", "latency (ticks)")
	queue, dropped, pushed, waiting, worst := 0, 0, 0, 0, 0
	for tick := 0; tick < c.Ticks; tick++ {
		in := c.ArrivePerTick
		if tick < c.BurstTicks {
			in = c.BurstPerTick
		}
		in += waiting // producers retry what was pushed back
		waiting = 0
		d, p := 0, 0
		switch c.Mode {
		case Unbounded:
			queue += in
		case LoadShedding:
			take := min(in, max(0, c.Limit-queue))
			d = in - take
			queue += take
		case PushBack:
			take := min(in, max(0, c.Limit-queue))
			p = in - take
			waiting = p
			queue += take
		}
		dropped += d
		pushed += p
		queue = max(0, queue-c.ServicePerTick)
		lat := queue / c.ServicePerTick
		worst = max(worst, lat)
		if len(t.Rows) < 40 {
			t.add(fmt.Sprintf("tick %d", tick), in, queue, d, p, lat)
		}
	}
	r.say("%s: worst latency %d ticks, %d dropped, %d pushed back to the producers", []string{"unbounded queue", "load shedding", "backpressure"}[c.Mode], worst, dropped, pushed)
	r.put("worst latency (ticks)", float64(worst))
	r.put("dropped", float64(dropped))
	r.put("pushed back", float64(pushed))
	return r, nil
}

// BulkheadConfig is a service calling a slow and a fast dependency.
type BulkheadConfig struct {
	Threads   int   // worker threads in the service
	Requests  int   // requests each tick, half to each dependency
	SlowTicks int   // ticks a call to the slow dependency holds a thread
	Ticks     int   // how long to run
	Bulkheads bool  // give each dependency its own half of the threads
	Seed      int64 // which dependency each request needs
}

// Bulkhead models the bulkhead pattern, named after a ship's watertight
// compartments. When one dependency slows down, requests to it hold threads
// for longer, and in a shared pool they soon hold all of them: requests to
// the healthy dependency fail too, and the outage spreads. Separate pools
// per dependency let the slow one exhaust only its own. It reports each
// dependency's success rate.
//
// Time: O(Ticks * Requests). Space: O(Threads).
func Bulkhead(c BulkheadConfig) (Report, error) {
	if err := check(c.Threads >= 2 && c.Threads <= 10_000 && c.Requests >= 1 && c.Requests <= 10_000 && c.SlowTicks >= 1 && c.Ticks >= 1 && c.Ticks <= 10_000,
		"2 to 10,000 threads, 1 to 10,000 requests a tick, positive slowness and ticks"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	busyUntil := make([]int, c.Threads) // tick each thread frees up
	pool := func(dep int) (lo, hi int) {
		if !c.Bulkheads {
			return 0, c.Threads
		}
		half := c.Threads / 2
		if dep == 0 {
			return 0, half
		}
		return half, c.Threads
	}
	ok, total := [2]int{}, [2]int{}
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "Clients"},
		{ID: "svc", Kind: Service, Label: "Service", Notes: map[bool]string{true: "a thread pool per dependency", false: "one shared pool"}[c.Bulkheads]},
		{ID: "fast", Kind: Service, Label: "Healthy dependency"},
		{ID: "slow", Kind: Service, Label: "Slow dependency"},
	}, "client>svc", "svc>fast", "svc>slow")
	for tick := 0; tick < c.Ticks; tick++ {
		for q := 0; q < c.Requests; q++ {
			dep := rng.Intn(2)
			total[dep]++
			lo, hi := pool(dep)
			got := -1
			for th := lo; th < hi; th++ {
				if busyUntil[th] <= tick {
					got = th
					break
				}
			}
			if got < 0 {
				continue
			}
			ok[dep]++
			hold := 1
			if dep == 1 {
				hold = c.SlowTicks
			}
			busyUntil[got] = tick + hold
		}
		if a.recording() {
			free := 0
			for _, b := range busyUntil {
				if b <= tick {
					free++
				}
			}
			f := a.frame(fmt.Sprintf("tick %d: %d of %d threads free", tick, free, c.Threads))
			a.send(f, "client", "svc", tick, OK)
			a.send(f, "svc", "fast", tick, map[bool]Status{true: OK, false: Failed}[free > 0 || c.Bulkheads])
			a.send(f, "svc", "slow", tick, Queued)
			f.load("svc", c.Threads-free, 0, -1)
		}
	}
	var r Report
	r.Arch = a
	r.say("%s: the healthy dependency's requests succeeded %d%%, the slow one's %d%%",
		map[bool]string{true: "bulkheads", false: "shared pool"}[c.Bulkheads], percent(ok[0], total[0]), percent(ok[1], total[1]))
	r.put("healthy dependency success (%)", float64(percent(ok[0], total[0])))
	r.put("slow dependency success (%)", float64(percent(ok[1], total[1])))
	return r, nil
}

// CoalesceConfig is a hot cache key that expires under load.
type CoalesceConfig struct {
	Clients  int  // requests for the key each tick
	TTL      int  // ticks the cached value lives
	DBTicks  int  // ticks a database read takes
	Ticks    int  // how long to run
	Coalesce bool // let one request refill the key while the others wait for it
}

// RequestCoalescing models the thundering herd (cache stampede): a hot key
// expires, and every request that misses in the next DBTicks ticks, before
// the first refill lands, goes to the database at once. Coalescing (Go's
// singleflight, a lock per key) sends one read and has the rest wait for
// its result. It reports the database reads and the peak concurrent reads.
// Refreshing ahead of expiry, or jittering TTLs, are the other fixes.
//
// Time: O(Ticks * Clients). Space: O(1).
func RequestCoalescing(c CoalesceConfig) (Report, error) {
	if err := check(c.Clients >= 1 && c.Clients <= 100_000 && c.TTL >= 1 && c.DBTicks >= 1 && c.Ticks >= 1 && c.Ticks <= 10_000, "positive clients (up to 100,000), TTL, database time and ticks"); err != nil {
		return Report{}, err
	}
	a := arch([]Component{
		{ID: "clients", Kind: Client, Label: "Clients"},
		{ID: "api", Kind: Service, Label: "Service", Notes: map[bool]string{true: "singleflight per key", false: "no coalescing"}[c.Coalesce]},
		{ID: "cache", Kind: Cache, Label: "Cache"},
		{ID: "db", Kind: Database, Label: "Database"},
	}, "clients>api", "api>cache", "api>db")
	expires, refillAt, inflight := 0, -1, 0
	reads, peak, hits, total := 0, 0, 0, 0
	for tick := 0; tick < c.Ticks; tick++ {
		if refillAt == tick {
			expires = tick + c.TTL
			refillAt, inflight = -1, 0
		}
		cached := tick < expires
		var f *Frame
		if a.recording() {
			f = a.frame("")
			a.send(f, "clients", "api", tick, OK)
		}
		newReads := 0
		for q := 0; q < c.Clients; q++ {
			total++
			if cached {
				hits++
				continue
			}
			if c.Coalesce && refillAt >= 0 {
				continue // waits for the read already in flight
			}
			reads++
			newReads++
			inflight++
			if refillAt < 0 {
				refillAt = tick + c.DBTicks
			}
		}
		peak = max(peak, inflight)
		if f != nil {
			if cached {
				f.Note = fmt.Sprintf("tick %d: the key is cached", tick)
				a.send(f, "api", "cache", tick, Hit)
			} else {
				f.Note = fmt.Sprintf("tick %d: the key expired; %d new database reads (%d in flight)", tick, newReads, inflight)
				a.send(f, "api", "cache", tick, Miss)
				a.send(f, "api", "db", tick, map[bool]Status{true: Queued, false: OK}[newReads > 1])
			}
			f.load("db", inflight, inflight, -1)
			f.load("cache", 0, 0, percent(hits, total))
		}
	}
	var r Report
	r.Arch = a
	r.say("%s: %d database reads, at most %d at once, cache hit rate %d%%", map[bool]string{true: "coalesced", false: "stampede"}[c.Coalesce], reads, peak, percent(hits, total))
	r.put("database reads", float64(reads))
	r.put("peak concurrent reads", float64(peak))
	r.put("cache hit rate (%)", float64(percent(hits, total)))
	return r, nil
}
