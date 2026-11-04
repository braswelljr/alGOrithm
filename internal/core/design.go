package core

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/design"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SystemArchitecture is a system design's components, connections and the
// requests in flight at each step, as pkg/systems/design draws it.
type SystemArchitecture = design.Architecture

// designOutput turns a design report into an Output: its lines, its metrics
// as the summary, its table and its architecture.
func designOutput(r design.Report) Output {
	out := Output{Lines: r.Lines, Architecture: r.Arch}
	if len(r.Metrics) > 0 {
		out.Summary = map[string]float64{}
		for _, m := range r.Metrics {
			out.Summary[m.Name] = m.Value
		}
	}
	if t := r.Table; t != nil && len(t.Rows) > 0 {
		tab := &TableOutput{Rows: len(t.Rows), Cols: len(t.Columns), ColLabels: t.Columns}
		for _, row := range t.Rows {
			tab.RowLabels = append(tab.RowLabels, row.Label)
			cells := make([]int, len(t.Columns))
			copy(cells, row.Cells)
			tab.Cells = append(tab.Cells, cells...)
		}
		out.Table = tab
	}
	return out
}

// designSpec is one system-design entry: its options and how to build the
// model's config from an input.
type designSpec struct {
	id, name, file, fn, summary, time, space string
	options                                  []Option
	example                                  map[string]int
	// config builds the model's config from the input; run runs it.
	config func(in Input) any
	run    func(cfg any) (design.Report, error)
}

func (s designSpec) entry() Entry {
	e := entry(s.id, s.name, "systems/design/"+s.file, s.fn)
	e.Summary, e.Time, e.Space, e.Input = s.summary, s.time, s.space, Simulation
	e.Options = s.options
	e.Example = &Input{Options: s.example}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		r, err := s.run(s.config(in))
		if err != nil {
			return Output{}, wrap(err)
		}
		return designOutput(r), nil
	}
	e.Call = func(in Input) (string, string) {
		cfg := s.config(in)
		if cfg == nil {
			return "", fmt.Sprintf("design.%s()", s.fn)
		}
		return "", fmt.Sprintf("design.%s(%#v)", s.fn, cfg)
	}
	return e
}

// flag reads an on/off option: any value but 0 is on.
func flag(in Input, name string, def int) bool { return in.opt(name, def) != 0 }

func designEntries() []Entry {
	seed := Option{"seed", "random seed", 1, 0, 1 << 30}
	specs := []designSpec{
		{
			id: "back-of-the-envelope", name: "Back-of-the-Envelope Estimation", file: "estimate.go", fn: "Estimate",
			summary: "Size a service before designing it: reads and writes a second, peak traffic, storage over the years, bandwidth.",
			time:    "O(1)", space: "O(1)",
			options: []Option{{"users", "daily active users", 10_000_000, 1, 2_000_000_000}, {"reads", "reads per user a day", 20, 0, 100_000},
				{"writes", "writes per user a day", 2, 0, 100_000}, {"bytes", "bytes per write", 1000, 1, 1 << 30}, {"peak", "peak / average", 3, 1, 100},
				{"years", "years kept", 5, 1, 100}, {"replicas", "copies of the data", 3, 1, 10}},
			example: map[string]int{"users": 10_000_000, "reads": 20, "writes": 2, "bytes": 1000, "peak": 3, "years": 5, "replicas": 3},
			config: func(in Input) any {
				return design.EstimateConfig{DailyUsers: in.opt("users", 10_000_000), ReadsPerUser: in.opt("reads", 20), WritesPerUser: in.opt("writes", 2),
					ObjectBytes: in.opt("bytes", 1000), PeakFactor: in.opt("peak", 3), Years: in.opt("years", 5), Replicas: in.opt("replicas", 3)}
			},
			run: func(c any) (design.Report, error) { return design.Estimate(c.(design.EstimateConfig)) },
		},
		{
			id: "latency-numbers", name: "Latency Numbers Every Programmer Should Know", file: "estimate.go", fn: "LatencyNumbers",
			summary: "From an L1 cache hit (half a nanosecond) to a packet across the Atlantic (150 ms): the ratios behind every design choice.",
			time:    "O(1)", space: "O(1)", example: map[string]int{},
			config: func(Input) any { return nil },
			run:    func(any) (design.Report, error) { return design.LatencyNumbers(), nil },
		},
		{
			id: "powers-of-two", name: "Powers of Two", file: "estimate.go", fn: "PowersOfTwo",
			summary: "2^10 is about a thousand, 2^20 a million, 2^30 a billion: the table for sizing data in your head.",
			time:    "O(1)", space: "O(1)", example: map[string]int{},
			config: func(Input) any { return nil },
			run:    func(any) (design.Report, error) { return design.PowersOfTwo(), nil },
		},
		{
			id: "snowflake-id", name: "Snowflake IDs", file: "ids.go", fn: "SnowflakeRun",
			summary: "64-bit IDs that sort by time, unique across workers with no coordination: milliseconds, worker and sequence bits, and what a clock going back does.",
			time:    "O(workers x ids)", space: "O(ids)",
			options: []Option{{"workers", "workers", 3, 1, 1024}, {"perms", "ids a worker makes each ms", 5, 1, 10_000}, {"millis", "milliseconds", 20, 1, 1000},
				{"skewat", "ms when worker 0's clock jumps back (0 never)", 10, 0, 1000}, {"skew", "ms it jumps back", 3, 0, 100}},
			example: map[string]int{"workers": 3, "perms": 5, "millis": 20, "skewat": 10, "skew": 3},
			config: func(in Input) any {
				return design.SnowflakeConfig{Workers: in.opt("workers", 3), PerMs: in.opt("perms", 5), Millis: in.opt("millis", 20),
					SkewAt: in.opt("skewat", 10), SkewMs: in.opt("skew", 3), FirstMs: 1_700_000_000_000}
			},
			run: func(c any) (design.Report, error) { return design.SnowflakeRun(c.(design.SnowflakeConfig)) },
		},
		{
			id: "uuid-ordering", name: "UUID v4 vs v7 vs ULID", file: "ids.go", fn: "IDOrdering",
			summary: "Random UUIDs scatter inserts across an index; time-ordered UUID v7 and ULIDs append at the end like an auto-increment key.",
			time:    "O(n^2)", space: "O(n)",
			options: []Option{{"inserts", "ids inserted", 2000, 1, 20_000}, {"page", "index entries per page", 100, 1, 10_000}, {"perms", "ids a millisecond", 4, 1, 1000}, seed},
			example: map[string]int{"inserts": 2000, "page": 100, "perms": 4, "seed": 1},
			config: func(in Input) any {
				return design.IDOrderingConfig{Inserts: in.opt("inserts", 2000), PageSize: in.opt("page", 100), PerMs: in.opt("perms", 4), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.IDOrdering(c.(design.IDOrderingConfig)) },
		},
		{
			id: "ticket-server", name: "Ticket Servers", file: "ids.go", fn: "TicketServer",
			summary: "Flickr's unique IDs: n database servers each counting by n from their own offset, so they never collide, even when one is down.",
			time:    "O(ids)", space: "O(ids)",
			options: []Option{{"servers", "ticket servers", 2, 1, 16}, {"ids", "ids handed out", 12, 1, 10_000}, {"down", "request when server 1 goes down (0 never)", 6, 0, 10_000}},
			example: map[string]int{"servers": 2, "ids": 12, "down": 6},
			config: func(in Input) any {
				return design.TicketConfig{Servers: in.opt("servers", 2), IDs: in.opt("ids", 12), DownAt: in.opt("down", 6)}
			},
			run: func(c any) (design.Report, error) { return design.TicketServer(c.(design.TicketConfig)) },
		},
		{
			id: "availability-nines", name: "Availability in Series and Parallel", file: "reliability.go", fn: "Availability",
			summary: "Every dependency in series multiplies availability down; every redundant copy in parallel adds a nine. Nines and downtime a year.",
			time:    "O(n)", space: "O(n)",
			options: []Option{{"bp", "each component's availability (basis points: 9900 is 99%)", 9900, 1, 9999}, {"components", "components", 3, 1, 10}},
			example: map[string]int{"bp": 9900, "components": 3},
			config: func(in Input) any {
				return design.AvailabilityConfig{BasisPoints: in.opt("bp", 9900), Components: in.opt("components", 3)}
			},
			run: func(c any) (design.Report, error) { return design.Availability(c.(design.AvailabilityConfig)) },
		},
		{
			id: "error-budget", name: "SLO Error Budget and Burn Rate", file: "reliability.go", fn: "ErrorBudget",
			summary: "A 99.9% SLO allows 0.1% failures: the budget, how much is spent, and the burn rate that says when to stop shipping or page someone.",
			time:    "O(1)", space: "O(1)",
			options: []Option{{"slo", "SLO (basis points: 9990 is 99.9%)", 9990, 1, 9999}, {"window", "window (days)", 30, 1, 366}, {"elapsed", "days so far", 10, 1, 366},
				{"requests", "requests so far", 1_000_000, 1, 1 << 40}, {"failed", "failed so far", 800, 0, 1 << 40}},
			example: map[string]int{"slo": 9990, "window": 30, "elapsed": 10, "requests": 1_000_000, "failed": 800},
			config: func(in Input) any {
				return design.ErrorBudgetConfig{SLOBasisPoints: in.opt("slo", 9990), WindowDays: in.opt("window", 30), ElapsedDays: in.opt("elapsed", 10),
					Requests: in.opt("requests", 1_000_000), Failed: in.opt("failed", 800)}
			},
			run: func(c any) (design.Report, error) { return design.ErrorBudget(c.(design.ErrorBudgetConfig)) },
		},
		{
			id: "littles-law", name: "Little's Law", file: "reliability.go", fn: "LittlesLaw",
			summary: "Requests in flight = arrival rate x time in the system: how many threads or connections a service needs, and why latency costs capacity.",
			time:    "O(1)", space: "O(1)",
			options: []Option{{"rate", "requests a second", 500, 1, 10_000_000}, {"latency", "latency (ms)", 80, 1, 600_000}},
			example: map[string]int{"rate": 500, "latency": 80},
			config: func(in Input) any {
				return design.LittleConfig{PerSecond: in.opt("rate", 500), LatencyMs: in.opt("latency", 80)}
			},
			run: func(c any) (design.Report, error) { return design.LittlesLaw(c.(design.LittleConfig)) },
		},
		{
			id: "queueing-latency", name: "Utilization vs Latency (M/M/1)", file: "reliability.go", fn: "Queueing",
			summary: "Why services run well below 100%: an M/M/1 queue's latency is 2x at half load, 10x at 90%, with a simulation beside the formula.",
			time:    "O(customers)", space: "O(1)",
			options: []Option{{"service", "requests served a second", 100, 1, 1_000_000}, {"customers", "requests simulated", 20_000, 1, 200_000}, seed},
			example: map[string]int{"service": 100, "customers": 20_000, "seed": 1},
			config: func(in Input) any {
				return design.QueueConfig{ServicePerSecond: in.opt("service", 100), Customers: in.opt("customers", 20_000), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Queueing(c.(design.QueueConfig)) },
		},
		{
			id: "url-shortener", name: "Design a URL Shortener", file: "shortener.go", fn: "URLShortener",
			summary: "Short codes from a base-62 counter or a hash (with collisions), and a cache in front of the database for the read-heavy redirects.",
			time:    "O(links + reads)", space: "O(links)",
			options: []Option{{"links", "links created", 1000, 1, 1_000_000}, {"reads", "redirects", 5000, 0, 1_000_000}, {"cache", "links cached", 100, 0, 1_000_000},
				{"hashed", "hash the URL (1) or count (0)", 0, 0, 1}, {"length", "characters of a hashed code", 4, 1, 11}, seed},
			example: map[string]int{"links": 1000, "reads": 5000, "cache": 100, "hashed": 0, "length": 4, "seed": 1},
			config: func(in Input) any {
				return design.ShortenerConfig{URLs: in.opt("links", 1000), Reads: in.opt("reads", 5000), CacheSize: in.opt("cache", 100),
					CodeLen: in.opt("length", 4), Hashed: flag(in, "hashed", 0), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.URLShortener(c.(design.ShortenerConfig)) },
		},
		{
			id: "base62", name: "Base 62 Encoding", file: "shortener.go", fn: "Base62",
			summary: "Write a number with 0-9, a-z and A-Z: seven characters cover 3.5 trillion short links.",
			time:    "O(log n)", space: "O(log n)",
			options: []Option{{"n", "number", 125_000_000_000, 0, 1 << 52}},
			example: map[string]int{"n": 125_000_000_000},
			config:  func(in Input) any { return int64(in.opt("n", 125_000_000_000)) },
			run: func(c any) (design.Report, error) {
				n := c.(int64)
				s, err := design.Base62(n)
				if err != nil {
					return design.Report{}, err
				}
				back, _ := design.FromBase62(s)
				return design.Report{Lines: []string{fmt.Sprintf("%d in base 62 is %q (%d characters)", n, s, len(s)), fmt.Sprintf("and %q reads back as %d", s, back)}}, nil
			},
		},
		{
			id: "key-value-store", name: "Design a Key-Value Store (Dynamo)", file: "kvstore.go", fn: "KeyValueStore",
			summary: "A consistent-hashing ring with N replicas, R and W quorums, sloppy quorums with hinted handoff while a node is down, and read repair.",
			time:    "O(ops x nodes)", space: "O(keys x nodes)",
			options: []Option{{"nodes", "nodes", 5, 2, 32}, {"n", "replicas N", 3, 1, 32}, {"r", "read quorum R", 2, 1, 32}, {"w", "write quorum W", 2, 1, 32},
				{"keys", "keys", 50, 1, 100_000}, {"ops", "operations", 400, 1, 100_000}, {"down", "node down mid-run (-1 none)", 2, -1, 31}, seed},
			example: map[string]int{"nodes": 5, "n": 3, "r": 2, "w": 2, "keys": 50, "ops": 400, "down": 2, "seed": 1},
			config: func(in Input) any {
				return design.KVConfig{Nodes: in.opt("nodes", 5), Replicas: in.opt("n", 3), R: in.opt("r", 2), W: in.opt("w", 2), Keys: in.opt("keys", 50),
					Ops: in.opt("ops", 400), DownNode: in.opt("down", 2), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.KeyValueStore(c.(design.KVConfig)) },
		},
		{
			id: "news-feed", name: "Design a News Feed", file: "feed.go", fn: "NewsFeed",
			summary: "Fan-out on write, fan-out on read, or a hybrid that pulls only celebrities: the work each does, and the celebrity problem.",
			time:    "O(posts x followers + reads x follows)", space: "O(users x follows)",
			options: []Option{{"users", "users", 1000, 2, 100_000}, {"follows", "follows each", 20, 1, 10_000}, {"celebrities", "celebrities", 3, 0, 1000},
				{"posts", "posts", 200, 0, 100_000}, {"reads", "feed reads", 500, 0, 100_000}, {"mode", "push 0, pull 1, hybrid 2", 2, 0, 2},
				{"threshold", "followers above which hybrid pulls", 100, 0, 1_000_000}, seed},
			example: map[string]int{"users": 1000, "follows": 20, "celebrities": 3, "posts": 200, "reads": 500, "mode": 2, "threshold": 100, "seed": 1},
			config: func(in Input) any {
				return design.FeedConfig{Users: in.opt("users", 1000), Follows: in.opt("follows", 20), Celebrities: in.opt("celebrities", 3), Posts: in.opt("posts", 200),
					Reads: in.opt("reads", 500), Mode: design.FeedMode(in.opt("mode", 2)), Threshold: in.opt("threshold", 100), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.NewsFeed(c.(design.FeedConfig)) },
		},
		{
			id: "chat-system", name: "Design a Chat System", file: "chat.go", fn: "Chat",
			summary: "WebSocket gateways, a sequence number per conversation to keep order through network reordering, presence, and the offline inbox.",
			time:    "O(m log m)", space: "O(m)",
			options: []Option{{"users", "users", 20, 2, 10_000}, {"messages", "messages", 200, 1, 100_000}, {"offline", "percent of recipients offline", 15, 0, 100},
				{"delay", "most ticks of network delay", 3, 0, 100}, seed},
			example: map[string]int{"users": 20, "messages": 200, "offline": 15, "delay": 3, "seed": 1},
			config: func(in Input) any {
				return design.ChatConfig{Users: in.opt("users", 20), Messages: in.opt("messages", 200), OfflinePercent: in.opt("offline", 15), MaxDelay: in.opt("delay", 3), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Chat(c.(design.ChatConfig)) },
		},
		{
			id: "notification-system", name: "Design a Notification System", file: "notify.go", fn: "Notifications",
			summary: "A queue, idempotency keys that drop the producer's retries, exponential backoff on provider failures, and a dead-letter queue.",
			time:    "O(n x retries)", space: "O(n)",
			options: []Option{{"notifications", "notifications", 200, 1, 100_000}, {"duplicates", "percent sent twice", 15, 0, 100}, {"fail", "percent of provider calls failing", 25, 0, 100},
				{"retries", "retries", 3, 0, 10}, seed},
			example: map[string]int{"notifications": 200, "duplicates": 15, "fail": 25, "retries": 3, "seed": 1},
			config: func(in Input) any {
				return design.NotifyConfig{Notifications: in.opt("notifications", 200), DuplicatePercent: in.opt("duplicates", 15), FailPercent: in.opt("fail", 25),
					MaxRetries: in.opt("retries", 3), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Notifications(c.(design.NotifyConfig)) },
		},
		{
			id: "web-crawler", name: "Design a Web Crawler", file: "crawler.go", fn: "WebCrawler",
			summary: "A URL frontier with a queue per host for politeness, fetchers, a parser, and a seen set so no page is fetched twice.",
			time:    "O(ticks x fetchers x links)", space: "O(pages)",
			options: []Option{{"hosts", "sites", 8, 1, 1000}, {"pages", "pages per site", 50, 1, 1000}, {"links", "links per page", 5, 0, 50},
				{"politeness", "ticks between fetches of one site", 3, 1, 1000}, {"fetchers", "fetches a tick", 4, 1, 1000}, {"ticks", "ticks", 60, 1, 10_000}, seed},
			example: map[string]int{"hosts": 8, "pages": 50, "links": 5, "politeness": 3, "fetchers": 4, "ticks": 60, "seed": 1},
			config: func(in Input) any {
				return design.CrawlerConfig{Hosts: in.opt("hosts", 8), PagesPerHost: in.opt("pages", 50), LinksPerPage: in.opt("links", 5), Politeness: in.opt("politeness", 3),
					Fetchers: in.opt("fetchers", 4), Ticks: in.opt("ticks", 60), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.WebCrawler(c.(design.CrawlerConfig)) },
		},
		{
			id: "geohash", name: "Geohash", file: "geo.go", fn: "Geohash",
			summary: "Interleave the bits of latitude and longitude and write them in base 32: nearby places share a prefix.",
			time:    "O(precision)", space: "O(precision)",
			options: []Option{{"lat", "latitude (1/10,000 degree)", 377_749, -900_000, 900_000}, {"lon", "longitude (1/10,000 degree)", -1_224_194, -1_800_000, 1_800_000},
				{"precision", "characters", 7, 1, 12}},
			example: map[string]int{"lat": 377_749, "lon": -1_224_194, "precision": 7},
			config: func(in Input) any {
				return [3]float64{float64(in.opt("lat", 377_749)) / 10_000, float64(in.opt("lon", -1_224_194)) / 10_000, float64(in.opt("precision", 7))}
			},
			run: func(c any) (design.Report, error) {
				p := c.([3]float64)
				var r design.Report
				for n := 1; n <= int(p[2]); n++ {
					g, err := design.Geohash(p[0], p[1], n)
					if err != nil {
						return design.Report{}, err
					}
					r.Lines = append(r.Lines, fmt.Sprintf("%2d characters: %s", n, g))
				}
				return r, nil
			},
		},
		{
			id: "proximity-service", name: "Design a Proximity Service", file: "geo.go", fn: "Proximity",
			summary: "Index businesses by geohash and search the user's cell and its eight neighbors: a few candidates instead of a city-wide scan.",
			time:    "O(candidates) per search", space: "O(businesses)",
			options: []Option{{"businesses", "businesses", 5000, 1, 200_000}, {"precision", "geohash characters", 5, 3, 8}, {"radius", "radius (100 m)", 20, 1, 1000}, seed},
			example: map[string]int{"businesses": 5000, "precision": 5, "radius": 20, "seed": 1},
			config: func(in Input) any {
				return design.ProximityConfig{Businesses: in.opt("businesses", 5000), Precision: in.opt("precision", 5), Lat: 37.7749, Lon: -122.4194,
					RadiusKm: float64(in.opt("radius", 20)) / 10, Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Proximity(c.(design.ProximityConfig)) },
		},
		{
			id: "ride-matching", name: "Design Ride Matching", file: "geo.go", fn: "RideMatching",
			summary: "Drivers in a grid of cells, and each rider matched to the nearest free driver by searching rings of cells outward.",
			time:    "O(riders x rings^2)", space: "O(drivers)",
			options: []Option{{"drivers", "drivers", 200, 1, 100_000}, {"riders", "riders", 150, 1, 100_000}, {"size", "grid size", 30, 2, 1000}, {"rings", "most rings searched", 4, 0, 1000}, seed},
			example: map[string]int{"drivers": 200, "riders": 150, "size": 30, "rings": 4, "seed": 1},
			config: func(in Input) any {
				return design.RideConfig{Drivers: in.opt("drivers", 200), Riders: in.opt("riders", 150), Size: in.opt("size", 30), MaxRing: in.opt("rings", 4), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.RideMatching(c.(design.RideConfig)) },
		},
		{
			id: "payment-system", name: "Design a Payment System", file: "payments.go", fn: "Payments",
			summary: "Idempotency keys so a retried payment is never charged twice, and a double-entry ledger that always sums to zero.",
			time:    "O(payments)", space: "O(payments)",
			options: []Option{{"payments", "payments", 300, 1, 100_000}, {"retry", "percent retried by the client", 25, 0, 100}, {"fail", "percent declined", 10, 0, 100},
				{"idempotent", "honor idempotency keys (1) or not (0)", 1, 0, 1}, {"cents", "amount (cents)", 1999, 1, 1 << 30}, seed},
			example: map[string]int{"payments": 300, "retry": 25, "fail": 10, "idempotent": 1, "cents": 1999, "seed": 1},
			config: func(in Input) any {
				return design.PaymentConfig{Payments: in.opt("payments", 300), RetryPercent: in.opt("retry", 25), FailPercent: in.opt("fail", 10),
					Idempotent: flag(in, "idempotent", 1), AmountCents: in.opt("cents", 1999), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Payments(c.(design.PaymentConfig)) },
		},
		{
			id: "distributed-rate-limiter", name: "Design a Distributed Rate Limiter", file: "ratelimit.go", fn: "DistributedRateLimiter",
			summary: "One limit across many API nodes: a central counter, local shares, or synced local counters, and what each lets through.",
			time:    "O(requests x nodes)", space: "O(nodes)",
			options: []Option{{"nodes", "API nodes", 4, 1, 100}, {"limit", "requests allowed", 1000, 1, 1 << 30}, {"requests", "requests in the window", 2000, 1, 1_000_000},
				{"sync", "requests between syncs", 50, 1, 1_000_000}, seed},
			example: map[string]int{"nodes": 4, "limit": 1000, "requests": 2000, "sync": 50, "seed": 1},
			config: func(in Input) any {
				return design.LimiterConfig{Nodes: in.opt("nodes", 4), Limit: in.opt("limit", 1000), Requests: in.opt("requests", 2000), SyncEvery: in.opt("sync", 50), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.DistributedRateLimiter(c.(design.LimiterConfig)) },
		},
		{
			id: "video-pipeline", name: "Design a Video Upload Pipeline", file: "video.go", fn: "VideoPipeline",
			summary: "Split uploads into chunks, transcode every chunk to every resolution in parallel on a pool of workers, then serve from a CDN.",
			time:    "O(jobs x workers)", space: "O(jobs)",
			options: []Option{{"videos", "videos", 6, 1, 1000}, {"chunks", "most chunks a video", 10, 1, 1000}, {"renditions", "resolutions", 4, 1, 8}, {"workers", "transcoders", 8, 1, 1000}, seed},
			example: map[string]int{"videos": 6, "chunks": 10, "renditions": 4, "workers": 8, "seed": 1},
			config: func(in Input) any {
				return design.VideoConfig{Videos: in.opt("videos", 6), Chunks: in.opt("chunks", 10), Renditions: in.opt("renditions", 4), Workers: in.opt("workers", 8), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.VideoPipeline(c.(design.VideoConfig)) },
		},
		{
			id: "adaptive-bitrate", name: "Adaptive Bitrate Streaming", file: "video.go", fn: "AdaptiveBitrate",
			summary: "Each second the player picks the best bitrate the network allows, dropping a step when the buffer runs thin, to avoid stalls.",
			time:    "O(seconds)", space: "O(seconds)",
			options: []Option{{"seconds", "seconds of playback", 60, 1, 3600}, {"kbps", "starting bandwidth (kbps)", 4000, 100, 100_000}, {"buffer", "buffer target (s)", 10, 1, 60}, seed},
			example: map[string]int{"seconds": 60, "kbps": 4000, "buffer": 10, "seed": 1},
			config: func(in Input) any {
				return design.ABRConfig{Seconds: in.opt("seconds", 60), StartKbps: in.opt("kbps", 4000), BufferTarget: in.opt("buffer", 10), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.AdaptiveBitrate(c.(design.ABRConfig)) },
		},
		{
			id: "leaderboard", name: "Design a Real-Time Leaderboard", file: "leaderboard.go", fn: "LeaderboardDemo",
			summary: "A sorted set (Redis ZSET) in front of the database: add points, the top k and any player's rank without scanning every player.",
			time:    "O(updates x players)", space: "O(players)",
			options: []Option{{"players", "players", 100, 1, 100_000}, {"updates", "score updates", 500, 0, 1_000_000}, {"top", "top k", 5, 1, 1000}, seed},
			example: map[string]int{"players": 100, "updates": 500, "top": 5, "seed": 1},
			config: func(in Input) any {
				return design.LeaderboardConfig{Players: in.opt("players", 100), Updates: in.opt("updates", 500), TopK: in.opt("top", 5), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.LeaderboardDemo(c.(design.LeaderboardConfig)) },
		},
		{
			id: "leader-follower-replication", name: "Leader-Follower Replication", file: "replication.go", fn: "LeaderFollower",
			summary: "Writes to the leader, reads from lagging followers: how often a user misses their own write, and how read-your-writes routing fixes it.",
			time:    "O(writes)", space: "O(followers)",
			options: []Option{{"followers", "followers", 3, 1, 50}, {"lag", "most ticks of lag", 5, 0, 1000}, {"writes", "writes read back", 200, 1, 100_000},
				{"ryw", "read-your-writes (1) or not (0)", 0, 0, 1}, seed},
			example: map[string]int{"followers": 3, "lag": 5, "writes": 200, "ryw": 0, "seed": 1},
			config: func(in Input) any {
				return design.ReplicationConfig{Followers: in.opt("followers", 3), MaxLag: in.opt("lag", 5), Writes: in.opt("writes", 200),
					ReadYourWrite: flag(in, "ryw", 0), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.LeaderFollower(c.(design.ReplicationConfig)) },
		},
		{
			id: "multi-leader-replication", name: "Multi-Leader Conflict Resolution", file: "replication.go", fn: "MultiLeader",
			summary: "Two datacenters each taking writes: last write wins silently drops conflicting updates; version vectors keep both for a merge.",
			time:    "O(rounds x writes)", space: "O(keys)",
			options: []Option{{"keys", "keys", 100, 1, 10_000}, {"writes", "writes a side a round", 10, 0, 10_000}, {"rounds", "replication rounds", 20, 1, 1000},
				{"vectors", "version vectors (1) or last write wins (0)", 0, 0, 1}, seed},
			example: map[string]int{"keys": 100, "writes": 10, "rounds": 20, "vectors": 0, "seed": 1},
			config: func(in Input) any {
				return design.MultiLeaderConfig{Keys: in.opt("keys", 100), Writes: in.opt("writes", 10), Rounds: in.opt("rounds", 20), VectorClock: flag(in, "vectors", 0), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.MultiLeader(c.(design.MultiLeaderConfig)) },
		},
		{
			id: "sharding-strategies", name: "Range vs Hash vs Directory Sharding", file: "sharding.go", fn: "Sharding",
			summary: "Range sharding piles new and hot keys on one shard; hashing spreads them but scatters scans; a directory moves ranges where load is.",
			time:    "O(keys + reads)", space: "O(shards)",
			options: []Option{{"keys", "rows", 100_000, 1, 1_000_000}, {"shards", "shards", 8, 2, 256}, {"writes", "new rows", 20_000, 0, 1_000_000},
				{"hot", "percent of reads to the newest tenth", 70, 0, 100}, {"reads", "reads", 50_000, 0, 1_000_000}, seed},
			example: map[string]int{"keys": 100_000, "shards": 8, "writes": 20_000, "hot": 70, "reads": 50_000, "seed": 1},
			config: func(in Input) any {
				return design.ShardingConfig{Keys: in.opt("keys", 100_000), Shards: in.opt("shards", 8), Writes: in.opt("writes", 20_000), HotPercent: in.opt("hot", 70),
					Reads: in.opt("reads", 50_000), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Sharding(c.(design.ShardingConfig)) },
		},
		{
			id: "resharding", name: "Resharding Cost", file: "sharding.go", fn: "Resharding",
			summary: "Grow from 4 to 5 shards: hash mod n moves 80% of the keys, consistent hashing about a third, close to the least possible 20%.",
			time:    "O(keys log shards)", space: "O(shards)",
			options: []Option{{"keys", "keys", 20_000, 1, 200_000}, {"from", "shards before", 4, 1, 255}, {"to", "shards after", 5, 2, 256}},
			example: map[string]int{"keys": 20_000, "from": 4, "to": 5},
			config: func(in Input) any {
				return design.ReshardingConfig{Keys: in.opt("keys", 20_000), From: in.opt("from", 4), To: in.opt("to", 5)}
			},
			run: func(c any) (design.Report, error) { return design.Resharding(c.(design.ReshardingConfig)) },
		},
		{
			id: "cap-pacelc", name: "CAP and PACELC", file: "consistency.go", fn: "CAP",
			summary: "During a partition, choose consistency (refuse writes) or availability (diverge, reconcile later); else, choose latency or consistency.",
			time:    "O(ops)", space: "O(1)",
			options: []Option{{"ops", "operations", 100, 1, 100_000}, {"partition", "op the network splits at", 30, 0, 100_000}, {"heal", "op it heals at", 60, 0, 100_000},
				{"sync", "ms a synchronous write waits", 5, 0, 10_000}, seed},
			example: map[string]int{"ops": 100, "partition": 30, "heal": 60, "sync": 5, "seed": 1},
			config: func(in Input) any {
				return design.CAPConfig{Ops: in.opt("ops", 100), PartitionAt: in.opt("partition", 30), HealAt: in.opt("heal", 60), SyncMs: in.opt("sync", 5), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.CAP(c.(design.CAPConfig)) },
		},
		{
			id: "cqrs", name: "CQRS", file: "events.go", fn: "CQRS",
			summary: "Separate the write model from a read model shaped for queries, kept up to date by events: scalable, but reads can be stale.",
			time:    "O(commands + queries)", space: "O(commands)",
			options: []Option{{"commands", "commands", 300, 1, 100_000}, {"queries", "queries", 100, 0, 100_000}, {"lag", "most events the projection trails", 5, 0, 10_000}, seed},
			example: map[string]int{"commands": 300, "queries": 100, "lag": 5, "seed": 1},
			config: func(in Input) any {
				return design.CQRSConfig{Commands: in.opt("commands", 300), Queries: in.opt("queries", 100), LagMax: in.opt("lag", 5), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.CQRS(c.(design.CQRSConfig)) },
		},
		{
			id: "event-sourcing", name: "Event Sourcing", file: "events.go", fn: "EventSourcing",
			summary: "Store every change as an event and rebuild state by replaying them; snapshots keep the replay short.",
			time:    "O(events)", space: "O(events)",
			options: []Option{{"events", "events", 1000, 1, 1_000_000}, {"snapshot", "events between snapshots (0 none)", 100, 0, 1_000_000}, seed},
			example: map[string]int{"events": 1000, "snapshot": 100, "seed": 1},
			config: func(in Input) any {
				return design.EventSourcingConfig{Events: in.opt("events", 1000), SnapshotEvery: in.opt("snapshot", 100), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) {
				_, r, err := design.EventSourcing(c.(design.EventSourcingConfig))
				return r, err
			},
		},
		{
			id: "saga-pattern", name: "The Saga Pattern", file: "saga.go", fn: "Saga",
			summary: "A transaction across services as local steps, each with a compensation; a failure undoes the finished steps in reverse order.",
			time:    "O(steps)", space: "O(steps)",
			options: []Option{{"fail", "step that fails (0 none)", 4, 0, len(design.SagaSteps)}, {"orchestration", "orchestration (1) or choreography (0)", 1, 0, 1}},
			example: map[string]int{"fail": 4, "orchestration": 1},
			config: func(in Input) any {
				return design.SagaConfig{FailAt: in.opt("fail", 4), Orchestration: flag(in, "orchestration", 1)}
			},
			run: func(c any) (design.Report, error) {
				_, r, err := design.Saga(c.(design.SagaConfig))
				return r, err
			},
		},
		{
			id: "transactional-outbox", name: "Transactional Outbox", file: "events.go", fn: "TransactionalOutbox",
			summary: "Write and publish in two steps and a crash loses the event; write the event to an outbox in the same transaction and nothing is lost.",
			time:    "O(orders)", space: "O(orders)",
			options: []Option{{"orders", "orders", 200, 1, 100_000}, {"crash", "percent crashing mid-way", 10, 0, 100}, {"outbox", "use an outbox (1) or dual write (0)", 0, 0, 1}, seed},
			example: map[string]int{"orders": 200, "crash": 10, "outbox": 0, "seed": 1},
			config: func(in Input) any {
				return design.OutboxConfig{Orders: in.opt("orders", 200), CrashPercent: in.opt("crash", 10), Outbox: flag(in, "outbox", 0), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.TransactionalOutbox(c.(design.OutboxConfig)) },
		},
		{
			id: "change-data-capture", name: "Change Data Capture", file: "events.go", fn: "ChangeDataCapture",
			summary: "Tail the database's log and stream every change, in order per key, to search indexes and caches, so the app writes once.",
			time:    "O(changes)", space: "O(rows)",
			options: []Option{{"changes", "changes", 200, 1, 100_000}, {"rows", "rows", 20, 1, 1_000_000}, {"lag", "consumer lag", 4, 0, 100_000}, seed},
			example: map[string]int{"changes": 200, "rows": 20, "lag": 4, "seed": 1},
			config: func(in Input) any {
				return design.CDCConfig{Changes: in.opt("changes", 200), Rows: in.opt("rows", 20), LagTicks: in.opt("lag", 4), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.ChangeDataCapture(c.(design.CDCConfig)) },
		},
		{
			id: "pagination-strategies", name: "Offset vs Cursor Pagination", file: "pagination.go", fn: "Pagination",
			summary: "Offset pages repeat items when new ones arrive and read every skipped row; cursor pages never repeat and read only the page.",
			time:    "O(pages x offset)", space: "O(rows)",
			options: []Option{{"rows", "rows", 1000, 1, 1_000_000}, {"size", "page size", 20, 1, 10_000}, {"pages", "pages read", 5, 1, 1000}, {"new", "new rows before each page", 3, 0, 10_000}},
			example: map[string]int{"rows": 1000, "size": 20, "pages": 5, "new": 3},
			config: func(in Input) any {
				return design.PaginationConfig{Rows: in.opt("rows", 1000), PageSize: in.opt("size", 20), Pages: in.opt("pages", 5), NewPerPage: in.opt("new", 3)}
			},
			run: func(c any) (design.Report, error) { return design.Pagination(c.(design.PaginationConfig)) },
		},
		{
			id: "backpressure", name: "Backpressure and Load Shedding", file: "resilience.go", fn: "Backpressure",
			summary: "A burst bigger than the service: an unbounded queue's latency grows without limit, shedding drops the excess, backpressure pushes it back.",
			time:    "O(ticks)", space: "O(ticks)",
			options: []Option{{"service", "served a tick", 10, 1, 1_000_000}, {"arrive", "arrivals a tick", 8, 0, 1_000_000}, {"burst", "arrivals a tick in the burst", 20, 0, 1_000_000},
				{"burstticks", "burst length", 10, 0, 10_000}, {"ticks", "ticks", 30, 1, 10_000}, {"limit", "queue limit", 30, 1, 1_000_000}, {"mode", "unbounded 0, shed 1, backpressure 2", 1, 0, 2}},
			example: map[string]int{"service": 10, "arrive": 8, "burst": 20, "burstticks": 10, "ticks": 30, "limit": 30, "mode": 1},
			config: func(in Input) any {
				return design.BackpressureConfig{ServicePerTick: in.opt("service", 10), ArrivePerTick: in.opt("arrive", 8), BurstPerTick: in.opt("burst", 20),
					BurstTicks: in.opt("burstticks", 10), Ticks: in.opt("ticks", 30), Limit: in.opt("limit", 30), Mode: design.Overload(in.opt("mode", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Backpressure(c.(design.BackpressureConfig)) },
		},
		{
			id: "bulkhead-pattern", name: "The Bulkhead Pattern", file: "resilience.go", fn: "Bulkhead",
			summary: "One slow dependency takes every thread of a shared pool and the healthy one fails too; a pool per dependency contains it.",
			time:    "O(ticks x requests)", space: "O(threads)",
			options: []Option{{"threads", "threads", 10, 2, 10_000}, {"requests", "requests a tick", 6, 1, 10_000}, {"slow", "ticks a slow call holds a thread", 8, 1, 10_000},
				{"ticks", "ticks", 40, 1, 10_000}, {"bulkheads", "pool per dependency (1) or shared (0)", 0, 0, 1}, seed},
			example: map[string]int{"threads": 10, "requests": 6, "slow": 8, "ticks": 40, "bulkheads": 0, "seed": 1},
			config: func(in Input) any {
				return design.BulkheadConfig{Threads: in.opt("threads", 10), Requests: in.opt("requests", 6), SlowTicks: in.opt("slow", 8), Ticks: in.opt("ticks", 40),
					Bulkheads: flag(in, "bulkheads", 0), Seed: int64(in.opt("seed", 1))}
			},
			run: func(c any) (design.Report, error) { return design.Bulkhead(c.(design.BulkheadConfig)) },
		},
		{
			id: "request-coalescing", name: "Thundering Herd and Request Coalescing", file: "resilience.go", fn: "RequestCoalescing",
			summary: "A hot key expires and every request hits the database at once; coalescing sends one read and the rest wait for it.",
			time:    "O(ticks x clients)", space: "O(1)",
			options: []Option{{"clients", "requests a tick", 100, 1, 100_000}, {"ttl", "ticks the value lives", 5, 1, 10_000}, {"db", "ticks a database read takes", 3, 1, 10_000},
				{"ticks", "ticks", 30, 1, 10_000}, {"coalesce", "coalesce (1) or not (0)", 0, 0, 1}},
			example: map[string]int{"clients": 100, "ttl": 5, "db": 3, "ticks": 30, "coalesce": 0},
			config: func(in Input) any {
				return design.CoalesceConfig{Clients: in.opt("clients", 100), TTL: in.opt("ttl", 5), DBTicks: in.opt("db", 3), Ticks: in.opt("ticks", 30), Coalesce: flag(in, "coalesce", 0)}
			},
			run: func(c any) (design.Report, error) { return design.RequestCoalescing(c.(design.CoalesceConfig)) },
		},
	}
	out := make([]Entry, 0, len(specs)+2)
	for _, s := range specs {
		out = append(out, s.entry())
	}
	// base62 and geohash take plain arguments, not a config struct
	for i := range out {
		switch out[i].ID {
		case "base62":
			out[i].Call = func(in Input) (string, string) {
				return "", fmt.Sprintf("design.Base62(%d)", in.opt("n", 125_000_000_000))
			}
		case "geohash":
			out[i].Call = func(in Input) (string, string) {
				return "", fmt.Sprintf("design.Geohash(%v, %v, %d)", float64(in.opt("lat", 377_749))/10_000, float64(in.opt("lon", -1_224_194))/10_000, in.opt("precision", 7))
			}
		}
	}
	return append(out, consistencyEntry(), autocompleteEntry())
}

// consistencyEntry checks a history typed as text against the consistency models.
func consistencyEntry() Entry {
	e := entry("consistency-models", "Consistency Models", "systems/design/consistency.go", "CheckConsistency")
	e.Summary = "Is a history of reads and writes linearizable, sequential, causal or only eventual? A search for an order each model allows."
	e.Time, e.Space, e.Input = "O(n!)", "O(n)", Text
	e.Limits = "at most 10 operations"
	e.Example = &Input{Text: "A w x 1 0 1\nB r x 0 2 3\nB r x 1 4 5"}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ops, err := design.ParseHistory(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		_, r, err := design.CheckConsistency(ops)
		if err != nil {
			return Output{}, wrap(err)
		}
		return designOutput(r), nil
	}
	e.Call = func(in Input) (string, string) {
		return fmt.Sprintf("ops, err := design.ParseHistory(%q)\n\t", in.Text) + guard, "design.CheckConsistency(ops)"
	}
	return e
}

// autocompleteEntry suggests completions for a typed query from search terms.
func autocompleteEntry() Entry {
	e := entry("search-autocomplete", "Design Search Autocomplete", "systems/design/autocomplete.go", "Autocomplete")
	e.Summary = "Precompute the top k searches for every prefix, so each keystroke is one lookup instead of a scan of every matching term."
	e.Time, e.Space, e.Input = "O(1) per keystroke", "O(total length x k)", TwoTexts
	e.Options = []Option{{"k", "suggestions shown", 3, 1, 20}}
	e.Example = &Input{Text: "golang 90\ngo modules 40\ngoogle 120\ngopher 15\ngo generics 60\ngraph algorithms 30", Other: "go", Options: map[string]int{"k": 3}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		terms, err := design.ParseTerms(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		r, err := design.Autocomplete(terms, strings.TrimSpace(in.Other), in.opt("k", 3))
		if err != nil {
			return Output{}, wrap(err)
		}
		return designOutput(r), nil
	}
	e.Call = func(in Input) (string, string) {
		return fmt.Sprintf("terms, err := design.ParseTerms(%q)\n\t", in.Text) + guard, fmt.Sprintf("design.Autocomplete(terms, %q, %d)", strings.TrimSpace(in.Other), in.opt("k", 3))
	}
	return e
}
