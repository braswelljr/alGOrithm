package design

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func metric(t *testing.T, r Report, name string) float64 {
	t.Helper()
	v, ok := r.Metric(name)
	if !ok {
		t.Fatalf("no metric %q in %v", name, r.Metrics)
	}
	return v
}

func TestSnowflakeIncreasesAndDecodes(t *testing.T) {
	g, err := NewSnowflake(7)
	if err != nil {
		t.Fatal(err)
	}
	var last int64 = -1
	for ms := int64(1000); ms < 1003; ms++ {
		for k := 0; k < 5000; k++ {
			id, ok, err := g.Next(ms)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				if k < 4096 {
					t.Fatalf("ms %d: ran out after %d ids", ms, k)
				}
				break
			}
			if id <= last {
				t.Fatalf("id %d after %d does not increase", id, last)
			}
			last = id
			if gotMs, w, seq := DecodeSnowflake(id); gotMs != ms || w != 7 || seq != k {
				t.Fatalf("decode %d = %d %d %d, want %d 7 %d", id, gotMs, w, seq, ms, k)
			}
		}
	}
	if _, _, err := g.Next(999); !errors.Is(err, ErrClockBackwards) {
		t.Errorf("clock going back: err %v", err)
	}
	if _, err := NewSnowflake(1024); !errors.Is(err, ErrInput) {
		t.Errorf("worker 1024: err %v", err)
	}
	r, err := SnowflakeRun(SnowflakeConfig{Workers: 3, PerMs: 5, Millis: 20, SkewAt: 10, SkewMs: 3, FirstMs: 100})
	if err != nil {
		t.Fatal(err)
	}
	// the clock reads 107, 108, 109 against a last id at 109: the first two
	// milliseconds are refused, the third continues the sequence
	if metric(t, r, "refused (clock went back)") != 10 {
		t.Errorf("refused %v, want 2 ms x 5", metric(t, r, "refused (clock went back)"))
	}
}

func TestAvailabilityMath(t *testing.T) {
	if got := Parallel(0.99, 0.99); !near(got, 0.9999, 1e-12) {
		t.Errorf("two parallel 99%%: %v, want 0.9999", got)
	}
	if got := Series(0.99, 0.99); !near(got, 0.9801, 1e-12) {
		t.Errorf("two in series 99%%: %v", got)
	}
	if got := DowntimePerYear(0.999); !near(got, 525.6, 0.01) {
		t.Errorf("three nines downtime %v min, want 525.6", got)
	}
	if got := Nines(0.999); !near(got, 3, 1e-9) {
		t.Errorf("nines %v", got)
	}
}

func TestLittleAndMM1(t *testing.T) {
	if got := Little(100, 0.2); got != 20 {
		t.Errorf("Little(100/s, 200 ms) = %v, want 20", got)
	}
	w, lq, err := MM1(9, 10)
	if err != nil || !near(w, 1, 1e-12) || !near(lq, 8.1, 1e-9) {
		t.Errorf("MM1(9, 10) = %v %v %v, want 1 s and 8.1", w, lq, err)
	}
	if _, _, err := MM1(10, 10); !errors.Is(err, ErrInput) {
		t.Errorf("saturated queue: err %v", err)
	}
	r, err := Queueing(QueueConfig{ServicePerSecond: 100, Customers: 50_000, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range r.Table.Rows {
		formula, simulated := float64(row.Cells[0]), float64(row.Cells[1])
		if math.Abs(simulated-formula) > 0.25*formula {
			t.Errorf("%s: simulated %v far from formula %v", row.Label, simulated, formula)
		}
	}
}

func TestErrorBudget(t *testing.T) {
	r, err := ErrorBudget(ErrorBudgetConfig{SLOBasisPoints: 9990, WindowDays: 30, ElapsedDays: 30, Requests: 1_000_000, Failed: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if metric(t, r, "budget (requests)") != 1000 || metric(t, r, "burn rate") != 1 {
		t.Errorf("budget %v burn %v, want 1000 and 1", metric(t, r, "budget (requests)"), metric(t, r, "burn rate"))
	}
}

func TestSagaCompensatesInReverse(t *testing.T) {
	for _, orch := range []bool{true, false} {
		got, _, err := Saga(SagaConfig{FailAt: 4, Orchestration: orch})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"create order", "reserve inventory", "charge payment", "schedule shipping FAILED", "refund payment", "release inventory", "cancel order"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("orchestration %v: %v", orch, got)
		}
	}
	got, _, _ := Saga(SagaConfig{})
	if len(got) != len(SagaSteps) {
		t.Errorf("no failure: %v", got)
	}
}

func TestConsistencyModels(t *testing.T) {
	cases := []struct {
		history string
		want    Models
	}{
		// B reads 0 after A's write finished: sequential (B's read can go first) but not linearizable
		{"A w x 1 0 1\nB r x 0 2 3", Models{Linearizable: false, Sequential: true, Causal: true, Eventual: true}},
		{"A w x 1 0 1\nB r x 1 2 3", Models{Linearizable: true, Sequential: true, Causal: true, Eventual: true}},
		// A writes 1 then 2; B reads 2 then 1: breaks A's program order for everyone
		{"A w x 1 0 1\nA w x 2 2 3\nB r x 2 4 5\nB r x 1 6 7", Models{Eventual: true}},
		// two clients see two concurrent writes in opposite orders: causal (PRAM) but not sequential
		{"A w x 1 0 1\nB w y 1 0 1\nC r x 1 2 3\nC r y 0 4 5\nD r y 1 2 3\nD r x 0 4 5", Models{Causal: true, Eventual: true}},
	}
	for _, c := range cases {
		ops, err := ParseHistory(c.history)
		if err != nil {
			t.Fatal(err)
		}
		got, _, err := CheckConsistency(ops)
		if err != nil || got != c.want {
			t.Errorf("%q: %v, want %v (%v)", c.history, got, c.want, err)
		}
	}
	if _, err := ParseHistory("A x x 1 0 1"); !errors.Is(err, ErrInput) {
		t.Errorf("bad op: err %v", err)
	}
}

func TestNewsFeedCostMatchesFormula(t *testing.T) {
	c := FeedConfig{Users: 200, Follows: 10, Celebrities: 2, Posts: 100, Reads: 50, Seed: 1}
	push, err := NewsFeed(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Mode = FanOutOnRead
	pull, _ := NewsFeed(c)
	if metric(t, pull, "feed writes") != 0 || metric(t, pull, "timelines pulled") != 50*10 {
		t.Errorf("pull: writes %v, pulls %v, want 0 and reads x follows", metric(t, pull, "feed writes"), metric(t, pull, "timelines pulled"))
	}
	if metric(t, push, "timelines pulled") != 0 {
		t.Errorf("push pulled %v", metric(t, push, "timelines pulled"))
	}
	// every user follows both celebrities, so a celebrity post costs Users-1 writes
	if metric(t, push, "worst post (writes)") != 199 {
		t.Errorf("worst push %v, want 199", metric(t, push, "worst post (writes)"))
	}
	c.Mode, c.Threshold = Hybrid, 50
	hy, _ := NewsFeed(c)
	if metric(t, hy, "worst post (writes)") >= 199 || metric(t, hy, "work") >= metric(t, push, "work") {
		t.Errorf("hybrid should avoid celebrity fan-out: %v", hy.Metrics)
	}
}

func TestBase62RoundTrip(t *testing.T) {
	for _, n := range []int64{0, 1, 61, 62, 3843, 3844, 1 << 40} {
		s, err := Base62(n)
		if err != nil {
			t.Fatal(err)
		}
		back, err := FromBase62(s)
		if err != nil || back != n {
			t.Errorf("%d -> %q -> %d %v", n, s, back, err)
		}
	}
	if s, _ := Base62(62); s != "10" {
		t.Errorf("62 is %q", s)
	}
}

func TestPaymentsNeverDoubleCharge(t *testing.T) {
	c := PaymentConfig{Payments: 500, RetryPercent: 30, FailPercent: 10, Idempotent: true, AmountCents: 1999, Seed: 4}
	r, err := Payments(c)
	if err != nil {
		t.Fatal(err)
	}
	if metric(t, r, "double charges") != 0 || metric(t, r, "ledger balance") != 0 || metric(t, r, "retries deduplicated") == 0 {
		t.Errorf("idempotent: %v", r.Metrics)
	}
	c.Idempotent = false
	r, _ = Payments(c)
	if metric(t, r, "double charges") == 0 {
		t.Errorf("without keys retries should double charge: %v", r.Metrics)
	}
}

func TestEventSourcingSnapshotsAgree(t *testing.T) {
	log, r, err := EventSourcing(EventSourcingConfig{Events: 1000, SnapshotEvery: 100, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	if Replay(0, log) != int(metric(t, r, "balance")) || metric(t, r, "events replayed (from snapshot)") != 0 {
		t.Errorf("%v", r.Metrics)
	}
	b := 0
	for _, e := range log {
		if b = Apply(b, e); b < 0 {
			t.Fatal("an overdraft was recorded")
		}
	}
}

func TestPaginationCursorNeverRepeats(t *testing.T) {
	r, err := Pagination(PaginationConfig{Rows: 1000, PageSize: 20, Pages: 5, NewPerPage: 3})
	if err != nil {
		t.Fatal(err)
	}
	if metric(t, r, "offset repeats") != 12 {
		t.Errorf("offset repeats %v, want 4 pages x 3", metric(t, r, "offset repeats"))
	}
	if metric(t, r, "cursor rows read") != 100 {
		t.Errorf("cursor read %v", metric(t, r, "cursor rows read"))
	}
}

func TestResharding(t *testing.T) {
	r, err := Resharding(ReshardingConfig{Keys: 20_000, From: 4, To: 5})
	if err != nil {
		t.Fatal(err)
	}
	if m, ring := metric(t, r, "mod n moved (%)"), metric(t, r, "consistent hashing moved (%)"); m < 70 || ring > 35 {
		t.Errorf("mod n %v%%, ring %v%%", m, ring)
	}
}

func TestCaseStudiesEmitArchitectures(t *testing.T) {
	runs := map[string]func() (Report, error){
		"shortener": func() (Report, error) {
			return URLShortener(ShortenerConfig{URLs: 100, Reads: 500, CacheSize: 10, CodeLen: 3, Hashed: true, Seed: 1})
		},
		"kv": func() (Report, error) {
			return KeyValueStore(KVConfig{Nodes: 5, Replicas: 3, R: 2, W: 2, Keys: 50, Ops: 300, DownNode: 2, Seed: 1})
		},
		"chat": func() (Report, error) {
			return Chat(ChatConfig{Users: 10, Messages: 50, OfflinePercent: 20, MaxDelay: 3, Seed: 1})
		},
		"notify": func() (Report, error) {
			return Notifications(NotifyConfig{Notifications: 50, DuplicatePercent: 20, FailPercent: 30, MaxRetries: 2, Seed: 1})
		},
		"crawler": func() (Report, error) {
			return WebCrawler(CrawlerConfig{Hosts: 5, PagesPerHost: 20, LinksPerPage: 4, Politeness: 2, Fetchers: 3, Ticks: 30, Seed: 1})
		},
		"ride": func() (Report, error) {
			return RideMatching(RideConfig{Drivers: 50, Riders: 30, Size: 20, MaxRing: 5, Seed: 1})
		},
		"limiter": func() (Report, error) {
			return DistributedRateLimiter(LimiterConfig{Nodes: 4, Limit: 100, Requests: 300, SyncEvery: 20, Seed: 1})
		},
		"video": func() (Report, error) {
			return VideoPipeline(VideoConfig{Videos: 4, Chunks: 6, Renditions: 3, Workers: 5, Seed: 1})
		},
		"board": func() (Report, error) {
			return LeaderboardDemo(LeaderboardConfig{Players: 20, Updates: 100, TopK: 3, Seed: 1})
		},
		"bulkhead": func() (Report, error) {
			return Bulkhead(BulkheadConfig{Threads: 10, Requests: 6, SlowTicks: 8, Ticks: 20, Bulkheads: true, Seed: 1})
		},
	}
	for name, run := range runs {
		r, err := run()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if r.Arch == nil || len(r.Arch.Frames) == 0 || len(r.Arch.Frames) > maxFrames {
			t.Fatalf("%s: architecture %v", name, r.Arch)
		}
		ids := map[string]bool{}
		for _, c := range r.Arch.Components {
			ids[c.ID] = true
		}
		for _, l := range r.Arch.Links {
			if !ids[l.From] || !ids[l.To] {
				t.Errorf("%s: link %v names a missing component", name, l)
			}
		}
		for _, f := range r.Arch.Frames {
			for _, q := range f.Requests {
				if q.Link < 0 || q.Link >= len(r.Arch.Links) {
					t.Errorf("%s: request on link %d", name, q.Link)
				}
			}
		}
	}
}

func TestBulkheadsIsolateTheSlowDependency(t *testing.T) {
	c := BulkheadConfig{Threads: 10, Requests: 6, SlowTicks: 8, Ticks: 50, Seed: 3}
	shared, _ := Bulkhead(c)
	c.Bulkheads = true
	split, _ := Bulkhead(c)
	if metric(t, split, "healthy dependency success (%)") <= metric(t, shared, "healthy dependency success (%)") {
		t.Errorf("bulkheads %v vs shared %v", split.Metrics, shared.Metrics)
	}
}

func TestCoalescingCutsDatabaseReads(t *testing.T) {
	c := CoalesceConfig{Clients: 100, TTL: 5, DBTicks: 3, Ticks: 40}
	herd, _ := RequestCoalescing(c)
	c.Coalesce = true
	one, _ := RequestCoalescing(c)
	if metric(t, one, "peak concurrent reads") != 1 || metric(t, herd, "peak concurrent reads") < 100 {
		t.Errorf("coalesced %v, stampede %v", one.Metrics, herd.Metrics)
	}
}

func TestGeohashKnownValue(t *testing.T) {
	// the classic example: 57.64911, 10.40744 is u4pruydqqvj
	g, err := Geohash(57.64911, 10.40744, 11)
	if err != nil || g != "u4pruydqqvj" {
		t.Errorf("geohash %q %v", g, err)
	}
	if !strings.HasPrefix(g, "u4pr") {
		t.Error("prefix")
	}
}

func TestLeaderboardRanks(t *testing.T) {
	lb := NewLeaderboard()
	lb.Add("ann", 10)
	lb.Add("bob", 30)
	lb.Add("cat", 20)
	lb.Add("ann", 25) // 35
	if got := lb.Top(3); !reflect.DeepEqual(got, []string{"ann", "bob", "cat"}) {
		t.Errorf("top %v", got)
	}
	if r, ok := lb.Rank("cat"); !ok || r != 3 {
		t.Errorf("rank cat %d %v", r, ok)
	}
}

func TestEveryModelRejectsBadInput(t *testing.T) {
	bad := []error{}
	_, err := Estimate(EstimateConfig{})
	bad = append(bad, err)
	_, err = URLShortener(ShortenerConfig{})
	bad = append(bad, err)
	_, err = KeyValueStore(KVConfig{Nodes: 3, Replicas: 4})
	bad = append(bad, err)
	_, err = Pagination(PaginationConfig{})
	bad = append(bad, err)
	_, _, err = Saga(SagaConfig{FailAt: 9})
	bad = append(bad, err)
	for i, e := range bad {
		if !errors.Is(e, ErrInput) {
			t.Errorf("case %d: err %v", i, e)
		}
	}
}
