package design

import (
	"fmt"
	"math"
	"math/rand"
)

// Series is the availability of components that must all be up for the
// system to be up: the product of their availabilities. Each added
// dependency lowers it.
//
// Time: O(n). Space: O(1).
func Series(availabilities ...float64) float64 {
	a := 1.0
	for _, x := range availabilities {
		a *= x
	}
	return a
}

// Parallel is the availability of redundant components where one being up
// is enough: one minus the chance that all are down at once. It assumes they
// fail independently, which shared power, deploys and bugs often break.
//
// Time: O(n). Space: O(1).
func Parallel(availabilities ...float64) float64 {
	down := 1.0
	for _, x := range availabilities {
		down *= 1 - x
	}
	return 1 - down
}

// DowntimePerYear is the minutes a year a system with availability a is down.
//
// Time: O(1). Space: O(1).
func DowntimePerYear(a float64) float64 { return (1 - a) * 365 * 24 * 60 }

// Nines is how many nines availability a has (0.999 is 3).
//
// Time: O(1). Space: O(1).
func Nines(a float64) float64 {
	if a >= 1 {
		return math.Inf(1)
	}
	return -math.Log10(1 - a)
}

// AvailabilityConfig is a chain of components, each with the same availability.
type AvailabilityConfig struct {
	BasisPoints int // each component's availability in hundredths of a percent (9990 is 99.90%)
	Components  int // components in the chain, or replicas side by side
}

// Availability compares Components components in series (a request needs all
// of them: load balancer, service, database) with the same number side by
// side (any one will do), as availability, nines and downtime a year, for 1
// up to Components. Series loses a little with every hop; parallel gains a
// nine with every replica, if failures are independent.
//
// Time: O(Components). Space: O(Components).
func Availability(c AvailabilityConfig) (Report, error) {
	if err := check(c.BasisPoints > 0 && c.BasisPoints < 10_000 && c.Components >= 1 && c.Components <= 10, "availability between 0 and 10,000 basis points, 1 to 10 components"); err != nil {
		return Report{}, err
	}
	a := float64(c.BasisPoints) / 10_000
	var r Report
	t := r.table("series (bp x100)", "series downtime (min/yr)", "parallel (bp x100)", "parallel downtime (min/yr)")
	for n := 1; n <= c.Components; n++ {
		as := make([]float64, n)
		for i := range as {
			as[i] = a
		}
		s, p := Series(as...), Parallel(as...)
		t.add(fmt.Sprintf("%d component(s)", n), int(s*1_000_000), int(DowntimePerYear(s)), int(p*1_000_000), int(DowntimePerYear(p)))
		r.say("%d: in series %.4f%% (%.0f min down a year), in parallel %.6f%% (%.1f min)", n, s*100, DowntimePerYear(s), p*100, DowntimePerYear(p))
	}
	as := make([]float64, c.Components)
	for i := range as {
		as[i] = a
	}
	r.put("series availability (%)", Series(as...)*100)
	r.put("parallel availability (%)", Parallel(as...)*100)
	r.put("series nines", math.Round(Nines(Series(as...))*100)/100)
	return r, nil
}

// ErrorBudgetConfig is a service level objective and what happened so far.
type ErrorBudgetConfig struct {
	SLOBasisPoints int // the objective in hundredths of a percent (9990 is 99.9% good requests)
	WindowDays     int // the SLO window
	ElapsedDays    int // days of the window gone
	Requests       int // requests so far in the window
	Failed         int // of which failed
}

// ErrorBudget is the error budget of an SLO: the failures it allows over the
// window (1 - SLO of all requests), how much of it is spent, and the burn
// rate, the error rate divided by the allowed rate. A burn rate of 1 spends
// the budget exactly by the window's end; 14.4 spends a 30-day budget in two
// days, the usual threshold for paging someone. It assumes the traffic so
// far continues at the same rate.
//
// Time: O(1). Space: O(1).
func ErrorBudget(c ErrorBudgetConfig) (Report, error) {
	if err := check(c.SLOBasisPoints > 0 && c.SLOBasisPoints < 10_000 && c.WindowDays >= 1 && c.ElapsedDays >= 1 &&
		c.ElapsedDays <= c.WindowDays && c.Requests >= 1 && c.Failed >= 0 && c.Failed <= c.Requests, "an SLO under 100%%, elapsed days within the window, failed at most requests"); err != nil {
		return Report{}, err
	}
	slo := float64(c.SLOBasisPoints) / 10_000
	expected := float64(c.Requests) * float64(c.WindowDays) / float64(c.ElapsedDays)
	budget := expected * (1 - slo)
	errRate := float64(c.Failed) / float64(c.Requests)
	burn := errRate / (1 - slo)
	var r Report
	r.say("SLO %.2f%% over %d days allows %.0f failed requests of about %.0f", slo*100, c.WindowDays, budget, expected)
	r.say("%d failed so far: %.0f%% of the budget spent after %d of %d days", c.Failed, 100*float64(c.Failed)/budget, c.ElapsedDays, c.WindowDays)
	r.say("burn rate %.2f: %s", burn, burnVerdict(burn, c.WindowDays))
	r.put("budget (requests)", math.Round(budget))
	r.put("budget spent (%)", math.Round(1000*float64(c.Failed)/budget)/10)
	r.put("burn rate", math.Round(burn*100)/100)
	if burn > 0 {
		r.put("days until spent at this rate", math.Round(float64(c.WindowDays)/burn*10)/10)
	}
	return r, nil
}

func burnVerdict(burn float64, window int) string {
	switch {
	case burn >= 14.4:
		return fmt.Sprintf("the %d-day budget goes in under %.1f days: page someone", window, float64(window)/burn)
	case burn > 1:
		return "spending faster than the budget allows: slow releases down"
	default:
		return "within budget: room to ship"
	}
}

// Little is Little's law, L = lambda x W: the average number of requests in a
// system is the arrival rate times the average time each spends in it. It
// holds for any stable system whatever the distributions, which is why it
// sizes thread pools and connection pools.
//
// Time: O(1). Space: O(1).
func Little(arrivalsPerSecond, secondsInSystem float64) float64 {
	return arrivalsPerSecond * secondsInSystem
}

// LittleConfig is an arrival rate and a latency.
type LittleConfig struct {
	PerSecond int // requests arriving a second
	LatencyMs int // average time a request spends inside, in milliseconds
}

// LittlesLaw applies Little's law to size a service: with PerSecond requests
// a second each taking LatencyMs, about lambda x W are in flight at once,
// which is how many threads or connections it needs, and it shows how that
// grows as latency grows.
//
// Time: O(1). Space: O(1).
func LittlesLaw(c LittleConfig) (Report, error) {
	if err := check(c.PerSecond >= 1 && c.LatencyMs >= 1, "a positive rate and latency"); err != nil {
		return Report{}, err
	}
	var r Report
	l := Little(float64(c.PerSecond), float64(c.LatencyMs)/1000)
	r.say("L = lambda x W = %d/s x %d ms = %.1f requests in flight on average", c.PerSecond, c.LatencyMs, l)
	r.say("so a pool of about %.0f workers keeps up; if latency doubles, so does the pool needed", math.Ceil(l))
	t := r.table("in flight")
	for _, f := range []int{1, 2, 5, 10} {
		t.add(fmt.Sprintf("latency %d ms", c.LatencyMs*f), int(math.Ceil(Little(float64(c.PerSecond), float64(c.LatencyMs*f)/1000))))
	}
	r.put("in flight", math.Round(l*10)/10)
	return r, nil
}

// MM1 is an M/M/1 queue (random arrivals at rate lambda, one server with
// random service times at rate mu): the average time a request spends in the
// system, W = 1 / (mu - lambda), and the average number waiting, Lq = rho^2 /
// (1 - rho) with rho = lambda / mu. It is an error unless lambda < mu.
//
// Time: O(1). Space: O(1).
func MM1(lambda, mu float64) (w, lq float64, err error) {
	if err := check(lambda >= 0 && mu > 0 && lambda < mu, "arrivals must be slower than service"); err != nil {
		return 0, 0, err
	}
	rho := lambda / mu
	return 1 / (mu - lambda), rho * rho / (1 - rho), nil
}

// QueueConfig is a single server queue to model.
type QueueConfig struct {
	ServicePerSecond int   // requests the server finishes a second
	Customers        int   // requests simulated at each utilization
	Seed             int64 // the random arrivals and services
}

// Queueing shows why latency explodes near full utilization: for an M/M/1
// queue at 10% to 95% utilization it gives the formula's average latency and
// queue length, and a seeded simulation's measured latency beside it. At 50%
// a request waits as long as it is served; at 90%, nine times as long; past
// that, a little more load means a lot more delay. That is why services run
// well below 100%.
//
// Time: O(Customers) per utilization. Space: O(1).
func Queueing(c QueueConfig) (Report, error) {
	if err := check(c.ServicePerSecond >= 1 && c.Customers >= 1 && c.Customers <= 200_000, "a positive service rate, 1 to 200,000 customers"); err != nil {
		return Report{}, err
	}
	mu := float64(c.ServicePerSecond)
	rng := rand.New(rand.NewSource(c.Seed))
	var r Report
	t := r.table("formula latency (us)", "simulated latency (us)", "queue length x100")
	for _, u := range []int{10, 30, 50, 70, 80, 90, 95} {
		lambda := mu * float64(u) / 100
		w, lq, _ := MM1(lambda, mu)
		t.add(fmt.Sprintf("%d%% utilization", u), int(w*1e6), int(simulateMM1(rng, lambda, mu, c.Customers)*1e6), int(lq*100))
	}
	r.say("W = 1/(mu - lambda): latency is 2x the service time at 50%% busy, 10x at 90%%, 20x at 95%%")
	w, _, _ := MM1(0.9*mu, mu)
	r.put("latency at 90% (ms)", math.Round(w*1e4)/10)
	return r, nil
}

// simulateMM1 is the measured average time in an M/M/1 system over n customers.
func simulateMM1(rng *rand.Rand, lambda, mu float64, n int) float64 {
	arrive, free, total := 0.0, 0.0, 0.0
	for i := 0; i < n; i++ {
		arrive += rng.ExpFloat64() / lambda
		start := math.Max(arrive, free)
		free = start + rng.ExpFloat64()/mu
		total += free - arrive
	}
	return total / float64(n)
}
