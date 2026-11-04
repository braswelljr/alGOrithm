package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/systems/infrastructure"
)

// infrastructureEntries is every entry of the infrastructure category.
func infrastructureEntries() []Entry {
	return infrastructureSimEntries()
}

func infrastructureSimEntries() []Entry {
	var out []Entry
	for _, s := range infraSims() {
		out = append(out, s.entry())
	}
	return out
}

func infraSimBalancer(st infrastructure.Strategy, id, name, summary string) sim {
	return sim{
		id: id, name: name, file: "systems/infrastructure/loadbalance.go", fn: "Balance", summary: summary,
		time: "O(T * (servers + open requests))", room: "O(servers * T)", input: Ints,
		example: Input{Values: []int{1, 2, 4}, Options: map[string]int{"requests": 40, "seed": 1}},
		options: []Option{{"requests", "requests (one a tick)", 40, 1, 300}, seedOption},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Balance(st, in.Values, in.opt("requests", 40), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Rows: len(r.Active)}
			for s, row := range r.Active {
				t.Cols = len(row)
				t.Cells = append(t.Cells, row...)
				t.RowLabels = append(t.RowLabels, fmt.Sprintf("server %d (speed %d)", s, in.Values[s]))
			}
			return Output{Lines: r.Log, Table: t, Values: r.Served, Summary: map[string]float64{"average latency": r.Latency, "peak open requests": float64(r.Peak)}}, nil
		},
		args: func(in Input) []any { return []any{string(st), in.Values, in.opt("requests", 40), in.opt("seed", 1)} },
	}
}

var infraSimJitters = []infrastructure.Jitter{infrastructure.NoJitter, infrastructure.FullJitter, infrastructure.EqualJitter}

var infraSimPolicies = []infrastructure.CachePolicy{infrastructure.CacheAside, infrastructure.WriteThrough, infrastructure.WriteBehind}

var infraSimFixes = []infrastructure.StampedeFix{infrastructure.NoFix, infrastructure.LockFix, infrastructure.EarlyExpiry}

var infraSimGuarantees = []infrastructure.Guarantee{infrastructure.AtMostOnce, infrastructure.AtLeastOnce, infrastructure.ExactlyOnce}

var infraSimRollouts = []infrastructure.Rollout{infrastructure.BlueGreen, infrastructure.Canary, infrastructure.Rolling}

//nolint:dupl // rows of one table share their shape; each is a different simulation
//nolint:dupl // rows of one table share their shape; each is a different simulation
var infraSimTable = []sim{
	infraSimBalancer(infrastructure.RoundRobin, "round-robin-load-balancer", "Round-Robin Load Balancing", "Send each request to the next server in turn, fast or slow: simple, and unfair to slow servers."),
	infraSimBalancer(infrastructure.WeightedRobin, "weighted-load-balancer", "Weighted Round-Robin Load Balancing", "Send requests in proportion to each server's speed."),
	infraSimBalancer(infrastructure.LeastConnections, "least-connections", "Least-Connections Load Balancing", "Send each request to the server with the fewest open requests for its speed."),
	infraSimBalancer(infrastructure.PowerOfTwo, "power-of-two-choices", "Power of Two Choices", "Sample two servers at random and take the less loaded: nearly as good as checking them all."),
	{
		id: "circuit-breaker", name: "Circuit Breaker", file: "systems/infrastructure/breaker.go", fn: "CircuitBreaker",
		summary: "After repeated failures stop calling a sick service for a while, then let one trial call decide whether to resume.",
		time:    "O(n)", room: "O(n)", input: Text,
		example: Input{Text: "ok fail fail fail ok ok ok ok fail ok ok ok", Options: map[string]int{"threshold": 3, "cooldown": 3}},
		options: []Option{{"threshold", "failures in a row to open", 3, 1, 20}, {"cooldown", "ticks open before a trial", 3, 1, 50}},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.CircuitBreaker(in.Text, in.opt("threshold", 3), in.opt("cooldown", 3))
			if err != nil {
				return Output{}, err
			}
			out := Output{Lines: r.Log, Summary: map[string]float64{"rejected fast": float64(r.Rejected), "failed": float64(r.Failed)}}
			for _, e := range r.Timeline {
				out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
			}
			return out, nil
		},
		args: func(in Input) []any { return []any{in.Text, in.opt("threshold", 3), in.opt("cooldown", 3)} },
	},
	{
		id: "exponential-backoff", name: "Exponential Backoff and Jitter", file: "systems/infrastructure/backoff.go", fn: "Backoff",
		summary: "Many clients retry after an outage: without jitter their retries arrive in synchronized spikes, with jitter they spread out.",
		time:    "O(clients * attempts + T)", room: "O(T)", input: Simulation,
		example: Input{Options: map[string]int{"clients": 50, "attempts": 4, "base": 2, "cap": 32, "jitter": 1, "seed": 1}},
		options: []Option{{"clients", "clients", 50, 1, 500}, {"attempts", "retries each", 4, 1, 10}, {"base", "base delay", 2, 1, 100}, {"cap", "largest delay", 32, 1, 1000}, {"jitter", "0 none, 1 full, 2 equal", 1, 0, 2}, seedOption},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Backoff(in.opt("clients", 50), in.opt("attempts", 4), in.opt("base", 2), in.opt("cap", 32), infraSimJitters[in.opt("jitter", 1)], int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Values: r.PerTick, Summary: map[string]float64{"peak retries in a tick": float64(r.Peak), "retries": float64(r.Total)}}, nil
		},
		args: func(in Input) []any {
			return []any{in.opt("clients", 50), in.opt("attempts", 4), in.opt("base", 2), in.opt("cap", 32), string(infraSimJitters[in.opt("jitter", 1)]), in.opt("seed", 1)}
		},
	},
	{
		id: "cache-strategies", name: "Cache-Aside, Write-Through, Write-Behind", file: "systems/infrastructure/caching.go", fn: "CacheStrategy",
		summary: "Three ways to keep a cache in front of a database, traded off in database writes, misses and lost writes.",
		time:    "O(n)", room: "O(keys)", input: Text,
		example: Input{Text: "write user:1 ada\nread user:1\nread user:1\nwrite user:1 grace\nwrite user:1 linus\nread user:1\ncrash\nread user:1", Options: map[string]int{"policy": 0}},
		options: []Option{{"policy", "0 cache-aside, 1 write-through, 2 write-behind", 0, 0, 2}},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.CacheStrategy(infraSimPolicies[in.opt("policy", 0)], in.Text)
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Summary: map[string]float64{"cache hits": float64(r.Hits), "database reads": float64(r.DBReads), "database writes": float64(r.DBWrites), "writes lost": float64(r.Lost)}}, nil
		},
		args: func(in Input) []any { return []any{string(infraSimPolicies[in.opt("policy", 0)]), in.Text} },
	},
	{
		id: "cache-stampede", name: "Cache Stampede", file: "systems/infrastructure/stampede.go", fn: "Stampede",
		summary: "A hot key expires and every request hits the database at once; a lock or early refresh keeps it to one.",
		time:    "O(ticks * rate)", room: "O(ticks)", input: Simulation,
		example: Input{Options: map[string]int{"rate": 20, "ticks": 60, "ttl": 20, "recompute": 4, "fix": 0, "seed": 1}},
		options: []Option{{"rate", "requests a tick", 20, 1, 100}, {"ticks", "ticks", 60, 1, 300}, {"ttl", "cache TTL", 20, 2, 100}, {"recompute", "ticks to recompute", 4, 1, 99}, {"fix", "0 none, 1 lock, 2 early expiry", 0, 0, 2}, seedOption},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Stampede(in.opt("rate", 20), in.opt("ticks", 60), in.opt("ttl", 20), in.opt("recompute", 4), infraSimFixes[in.opt("fix", 0)], int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Values: r.DBPerTick, Summary: map[string]float64{"peak queries in a tick": float64(r.Peak), "queries": float64(r.Total)}}, nil
		},
		args: func(in Input) []any {
			return []any{in.opt("rate", 20), in.opt("ticks", 60), in.opt("ttl", 20), in.opt("recompute", 4), string(infraSimFixes[in.opt("fix", 0)]), in.opt("seed", 1)}
		},
	},
	{
		id: "cdn-edge-cache", name: "CDN Edge Caches", file: "systems/infrastructure/cdn.go", fn: "CDN",
		summary: "Requests from several regions for popular and rare objects: how much each edge serves and how much reaches the origin.",
		time:    "O(requests)", room: "O(regions * objects)", input: Simulation,
		example: Input{Options: map[string]int{"regions": 3, "objects": 50, "requests": 600, "ttl": 100, "seed": 1}},
		options: []Option{{"regions", "regions", 3, 1, 6}, {"objects", "objects", 50, 1, 1000}, {"requests", "requests", 600, 1, 5000}, {"ttl", "edge TTL (ticks)", 100, 1, 5000}, seedOption},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.CDN(in.opt("regions", 3), in.opt("objects", 50), in.opt("requests", 600), in.opt("ttl", 100), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Cols: 3, ColLabels: []string{"requests", "edge hits", "origin fetches"}}
			for _, st := range r.Regions {
				t.Rows++
				t.Cells = append(t.Cells, st.Requests, st.Hits, st.Misses)
				t.RowLabels = append(t.RowLabels, st.Region)
			}
			return Output{Lines: r.Log, Table: t, Number: ptr(r.Origin)}, nil
		},
		args: func(in Input) []any {
			return []any{in.opt("regions", 3), in.opt("objects", 50), in.opt("requests", 600), in.opt("ttl", 100), in.opt("seed", 1)}
		},
	},
	{
		id: "delivery-guarantees", name: "At-Most-Once, At-Least-Once, Exactly-Once", file: "systems/infrastructure/delivery.go", fn: "Delivery",
		summary: "Send messages over a lossy link: lose some, duplicate some, or dedupe with idempotency keys.",
		time:    "O(messages * retries)", room: "O(messages)", input: Simulation,
		example: Input{Options: map[string]int{"guarantee": 1, "messages": 100, "loss": 20, "seed": 1}},
		options: []Option{{"guarantee", "0 at most once, 1 at least once, 2 exactly once", 1, 0, 2}, {"messages", "messages", 100, 1, 1000}, {"loss", "loss percent", 20, 0, 80}, seedOption},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Delivery(infraSimGuarantees[in.opt("guarantee", 1)], in.opt("messages", 100), in.opt("loss", 20), int64(in.opt("seed", 1)))
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Summary: map[string]float64{"transmissions": float64(r.Sent), "processed": float64(r.Processed), "lost": float64(r.Lost), "duplicates": float64(r.Duplicates)}}, nil
		},
		args: func(in Input) []any {
			return []any{string(infraSimGuarantees[in.opt("guarantee", 1)]), in.opt("messages", 100), in.opt("loss", 20), in.opt("seed", 1)}
		},
	},
	{
		id: "autoscaling", name: "Autoscaling", file: "systems/infrastructure/autoscale.go", fn: "Autoscale",
		summary: "Add and remove instances to follow a load curve, with boot time and a cooldown, and count the requests dropped.",
		time:    "O(n)", room: "O(n)", input: Ints,
		example: Input{Values: []int{20, 20, 40, 90, 160, 160, 150, 80, 40, 30, 30, 20}, Options: map[string]int{"capacity": 25, "target": 70, "boot": 2, "cooldown": 3, "min": 1, "max": 20}},
		options: []Option{{"capacity", "requests per instance", 25, 1, 10000}, {"target", "target utilization %", 70, 10, 100}, {"boot", "boot ticks", 2, 0, 50}, {"cooldown", "cooldown ticks", 3, 0, 100}, {"min", "fewest instances", 1, 1, 1000}, {"max", "most instances", 20, 1, 1000}},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Autoscale(in.Values, in.opt("capacity", 25), in.opt("target", 70), in.opt("boot", 2), in.opt("cooldown", 3), in.opt("min", 1), in.opt("max", 20))
			if err != nil {
				return Output{}, err
			}
			n := len(r.Steps)
			t := &TableOutput{Rows: 4, Cols: n, Cells: make([]int, 4*n), RowLabels: []string{"load", "serving", "booting", "dropped"}}
			for i, s := range r.Steps {
				t.Cells[i], t.Cells[n+i], t.Cells[2*n+i], t.Cells[3*n+i] = s.Load, s.Instances, s.Booting, s.Dropped
				t.ColLabels = append(t.ColLabels, strconv.Itoa(i))
			}
			return Output{Lines: r.Log, Table: t, Summary: map[string]float64{"dropped": float64(r.Dropped), "instance ticks": float64(r.Cost)}}, nil
		},
		args: func(in Input) []any {
			return []any{in.Values, in.opt("capacity", 25), in.opt("target", 70), in.opt("boot", 2), in.opt("cooldown", 3), in.opt("min", 1), in.opt("max", 20)}
		},
	},
	{
		id: "deployment-strategies", name: "Blue/Green, Canary and Rolling Deployments", file: "systems/infrastructure/rollout.go", fn: "Deploy",
		summary: "Roll out a bad build three ways and count the failed requests before the rollback.",
		time:    "O(steps)", room: "O(steps)", input: Simulation,
		example: Input{Options: map[string]int{"strategy": 1, "instances": 10, "rps": 1000, "errors": 30, "bad": 1}},
		options: []Option{{"strategy", "0 blue/green, 1 canary, 2 rolling", 1, 0, 2}, {"instances", "instances", 10, 1, 50}, {"rps", "requests a step", 1000, 1, 100000}, {"errors", "error rate of the bad build %", 30, 0, 100}, {"bad", "the build is bad (1) or good (0)", 1, 0, 1}},
		run: func(in Input) (Output, error) {
			r, err := infrastructure.Deploy(infraSimRollouts[in.opt("strategy", 1)], in.opt("instances", 10), in.opt("rps", 1000), in.opt("errors", 30), on(in, "bad", 1))
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Cols: 2, ColLabels: []string{"new version %", "errors"}}
			var lines []string
			for i, s := range r.Steps {
				t.Rows++
				t.Cells = append(t.Cells, s.NewShare, s.Errors)
				t.RowLabels = append(t.RowLabels, "step "+strconv.Itoa(i+1))
				lines = append(lines, fmt.Sprintf("step %d: %s", i+1, s.Note))
			}
			return Output{Lines: lines, Table: t, Number: ptr(r.Errors), Bool: ptr(r.RolledBack)}, nil
		},
		args: func(in Input) []any {
			return []any{string(infraSimRollouts[in.opt("strategy", 1)]), in.opt("instances", 10), in.opt("rps", 1000), in.opt("errors", 30), on(in, "bad", 1)}
		},
	},
	{
		id: "cgroups-limits", name: "Container CPU Shares and Memory Limits", file: "systems/infrastructure/cgroups.go", fn: "Cgroups",
		summary: "Containers split the CPUs by shares only under contention, and one over its memory limit meets the OOM killer.",
		time:    "O(n^2)", room: "O(n)", input: Text,
		example: Input{Text: "web 1024 2 512 300\nbatch 512 2 256 100\nlogger 256 0.2 64 30\nleaky 1024 1 128 400", Options: map[string]int{"cpus": 2}},
		options: []Option{{"cpus", "CPUs", 2, 1, 64}},
		run: func(in Input) (Output, error) {
			cs, err := infrastructure.Cgroups(in.opt("cpus", 2), in.Text)
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Cols: 5, ColLabels: []string{"shares", "wanted (1/100 CPU)", "got (1/100 CPU)", "memory limit MiB", "memory used MiB"}}
			var lines []string
			for _, c := range cs {
				t.Rows++
				t.Cells = append(t.Cells, c.Shares, int(c.Want*100+0.5), int(c.Got*100+0.5), c.MemLimit, c.MemUse)
				t.RowLabels = append(t.RowLabels, c.Name)
				lines = append(lines, c.String())
			}
			return Output{Lines: lines, Table: t}, nil
		},
		args: func(in Input) []any { return []any{in.opt("cpus", 2), in.Text} },
	},
}

func infraSims() []sim {
	return infraSimTable
}
