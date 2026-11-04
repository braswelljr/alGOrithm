package design

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestEstimate(t *testing.T) {
	r, err := Estimate(EstimateConfig{DailyUsers: 864_000, ReadsPerUser: 10, WritesPerUser: 1, ObjectBytes: 1000, PeakFactor: 3, Years: 5, Replicas: 3})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{
		"reads per second":       100,
		"peak reads per second":  300,
		"writes per second":      10,
		"peak writes per second": 30,
		"storage per day (GB)":   0.9, // 864,000 x 1,000 B = 0.864 GB
		"total storage (TB)":     4.7, // 0.864 GB x 365 x 5 x 3 = 4.73 TB
	}
	for name, v := range want {
		if got := metric(t, r, name); got != v {
			t.Errorf("%s = %v, want %v", name, got, v)
		}
	}
	text := r.String()
	for _, s := range []string{"10:1", "read heavy", "864.0 MB a day", "4.7 TB over 5 years", "10.0 KB/s in", "100.0 KB/s out"} {
		if !strings.Contains(text, s) {
			t.Errorf("report lacks %q:\n%s", s, text)
		}
	}

	verdicts := []struct {
		reads, writes int
		want          string
	}{
		{5, 0, "all reads"},
		{1, 10, "write heavy"},
		{2, 1, "balanced"},
	}
	for _, v := range verdicts {
		r, err := Estimate(EstimateConfig{DailyUsers: 86_400, ReadsPerUser: v.reads, WritesPerUser: v.writes, ObjectBytes: 1, PeakFactor: 1, Years: 1, Replicas: 1})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.String(), v.want) {
			t.Errorf("%d reads, %d writes: want %q in\n%s", v.reads, v.writes, v.want, r)
		}
	}

	for name, c := range map[string]EstimateConfig{
		"no users":          {ObjectBytes: 1, PeakFactor: 1, Years: 1, Replicas: 1},
		"negative reads":    {DailyUsers: 1, ReadsPerUser: -1, ObjectBytes: 1, PeakFactor: 1, Years: 1, Replicas: 1},
		"zero peak":         {DailyUsers: 1, ObjectBytes: 1, Years: 1, Replicas: 1},
		"zero replicas":     {DailyUsers: 1, ObjectBytes: 1, PeakFactor: 1, Years: 1},
		"zero object bytes": {DailyUsers: 1, PeakFactor: 1, Years: 1, Replicas: 1},
	} {
		if _, err := Estimate(c); !errors.Is(err, ErrInput) {
			t.Errorf("%s: err %v", name, err)
		}
	}
}

func TestBytesAndRounding(t *testing.T) {
	cases := map[float64]string{0: "0.0 B", 999: "999.0 B", 1000: "1.0 KB", 1.5e6: "1.5 MB", 2e18: "2.0 EB", 3e21: "3000.0 EB"}
	for n, want := range cases {
		if got := bytes(n); got != want {
			t.Errorf("bytes(%v) = %q, want %q", n, got, want)
		}
	}
	if round1(0.04) != 0 || round1(0.05) != 0.1 || round1(12.34) != 12.3 {
		t.Error("round1")
	}
}

func TestLatencyNumbers(t *testing.T) {
	r := LatencyNumbers()
	if r.Table == nil || len(r.Table.Rows) != 14 || len(r.Lines) != 2 {
		t.Fatalf("table %v, %d lines", r.Table, len(r.Lines))
	}
	prev := 0
	for _, row := range r.Table.Rows {
		ns, rel := row.Cells[0], row.Cells[1]
		if ns < prev {
			t.Errorf("%s: %d ns is below the row before (%d): the table is sorted", row.Label, ns, prev)
		}
		if rel != ns/100 {
			t.Errorf("%s: %d x memory, want %d", row.Label, rel, ns/100)
		}
		prev = ns
	}
	for _, row := range r.Table.Rows {
		if row.Label == "main memory reference" && row.Cells[1] != 1 {
			t.Errorf("a memory reference is %d memory references", row.Cells[1])
		}
	}
}

func TestPowersOfTwo(t *testing.T) {
	r := PowersOfTwo()
	if r.Table == nil || len(r.Table.Rows) != 6 {
		t.Fatalf("table %v", r.Table)
	}
	for i, row := range r.Table.Rows {
		p := 10 * (i + 1)
		if row.Cells[0] != p || row.Cells[1] != 3*(i+1) {
			t.Errorf("row %d cells %v", i, row.Cells)
		}
		// 2^(10k) is within 16% of 10^(3k) for k up to 6
		if ratio := math.Pow(2, float64(p)) / math.Pow(10, float64(3*(i+1))); ratio < 1 || ratio > 1.16 {
			t.Errorf("2^%d / 10^%d = %v", p, 3*(i+1), ratio)
		}
	}
	if !strings.Contains(r.Table.Rows[0].Label, "1024 bytes: 1 kilobyte") || !strings.Contains(r.Table.Rows[5].Label, "1152921504606846976") {
		t.Errorf("labels %q, %q", r.Table.Rows[0].Label, r.Table.Rows[5].Label)
	}
}

func TestAvailability(t *testing.T) {
	r, err := Availability(AvailabilityConfig{BasisPoints: 9900, Components: 3})
	if err != nil {
		t.Fatal(err)
	}
	if got := metric(t, r, "series availability (%)"); !near(got, 97.0299, 1e-9) {
		t.Errorf("series %v, want 0.99^3 = 97.0299%%", got)
	}
	if got := metric(t, r, "parallel availability (%)"); !near(got, 99.9999, 1e-9) {
		t.Errorf("parallel %v, want 1 - 0.01^3", got)
	}
	if got := metric(t, r, "series nines"); got != 1.53 {
		t.Errorf("series nines %v, want -log10(0.029701) = 1.53", got)
	}
	if len(r.Table.Rows) != 3 || len(r.Lines) != 3 {
		t.Fatalf("%d rows, %d lines", len(r.Table.Rows), len(r.Lines))
	}
	for i := 1; i < len(r.Table.Rows); i++ {
		prev, cur := r.Table.Rows[i-1].Cells, r.Table.Rows[i].Cells
		if cur[0] >= prev[0] || cur[1] <= prev[1] {
			t.Errorf("series should lose availability with each component: %v then %v", prev, cur)
		}
		if cur[2] <= prev[2] || cur[3] >= prev[3] {
			t.Errorf("parallel should gain availability with each replica: %v then %v", prev, cur)
		}
	}
	if !math.IsInf(Nines(1), 1) {
		t.Error("perfect availability has infinitely many nines")
	}
	for _, c := range []AvailabilityConfig{{0, 1}, {10_000, 1}, {9900, 0}, {9900, 11}} {
		if _, err := Availability(c); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", c, err)
		}
	}
}

func TestErrorBudgetVerdicts(t *testing.T) {
	cases := []struct {
		failed int
		want   string
	}{
		{0, "within budget"},
		{500, "within budget"},
		{5000, "slow releases down"},
		{20_000, "page someone"},
	}
	for _, c := range cases {
		r, err := ErrorBudget(ErrorBudgetConfig{SLOBasisPoints: 9990, WindowDays: 30, ElapsedDays: 30, Requests: 1_000_000, Failed: c.failed})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.String(), c.want) {
			t.Errorf("%d failed: want %q in\n%s", c.failed, c.want, r)
		}
		_, has := r.Metric("days until spent at this rate")
		if has != (c.failed > 0) {
			t.Errorf("%d failed: days-until metric present %v", c.failed, has)
		}
	}
	if _, err := ErrorBudget(ErrorBudgetConfig{SLOBasisPoints: 9990, WindowDays: 30, ElapsedDays: 31, Requests: 1}); !errors.Is(err, ErrInput) {
		t.Errorf("elapsed past the window: %v", err)
	}
}

func TestLittlesLaw(t *testing.T) {
	r, err := LittlesLaw(LittleConfig{PerSecond: 100, LatencyMs: 200})
	if err != nil {
		t.Fatal(err)
	}
	if got := metric(t, r, "in flight"); got != 20 {
		t.Errorf("in flight %v, want 100/s x 0.2 s = 20", got)
	}
	var cells []int
	for _, row := range r.Table.Rows {
		cells = append(cells, row.Cells[0])
	}
	if want := []int{20, 40, 100, 200}; !reflect.DeepEqual(cells, want) {
		t.Errorf("in flight as latency grows %v, want %v", cells, want)
	}
	if !strings.Contains(r.Lines[1], "about 20 workers") {
		t.Errorf("line %q", r.Lines[1])
	}
	for _, c := range []LittleConfig{{0, 1}, {1, 0}} {
		if _, err := LittlesLaw(c); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", c, err)
		}
	}
}

func TestIDOrdering(t *testing.T) {
	r, err := IDOrdering(IDOrderingConfig{Inserts: 2000, PageSize: 50, PerMs: 10, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Table.Rows) != 3 {
		t.Fatalf("%d rows", len(r.Table.Rows))
	}
	for _, row := range r.Table.Rows {
		if row.Cells[0]+row.Cells[1] != 2000 {
			t.Errorf("%s: %d appends + %d random != 2000 inserts", row.Label, row.Cells[0], row.Cells[1])
		}
	}
	v4 := metric(t, r, "UUID v4 (random) random writes (%)")
	v7 := metric(t, r, "UUID v7 (48-bit ms + random) random writes (%)")
	ulid := metric(t, r, "ULID (48-bit ms + 80 random bits) random writes (%)")
	if v4 < 80 || v7 > 5 || ulid > 5 {
		t.Errorf("random writes: v4 %v%%, v7 %v%%, ulid %v%%", v4, v7, ulid)
	}
	again, _ := IDOrdering(IDOrderingConfig{Inserts: 2000, PageSize: 50, PerMs: 10, Seed: 1})
	if !reflect.DeepEqual(r, again) {
		t.Error("the same seed gave different runs")
	}
	one, _ := IDOrdering(IDOrderingConfig{Inserts: 1, PageSize: 1, PerMs: 1})
	for _, row := range one.Table.Rows {
		if row.Cells[0] != 1 {
			t.Errorf("a single insert into an empty index is an append: %v", row)
		}
	}
	for _, c := range []IDOrderingConfig{{0, 1, 1, 0}, {1, 0, 1, 0}, {1, 1, 0, 0}, {20_001, 1, 1, 0}} {
		if _, err := IDOrdering(c); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", c, err)
		}
	}
}

func TestTicketServer(t *testing.T) {
	t.Run("two servers interleave odd and even ids", func(t *testing.T) {
		r, err := TicketServer(TicketConfig{Servers: 2, IDs: 6})
		if err != nil {
			t.Fatal(err)
		}
		if r.Lines[0] != "ids handed out: [1 2 3 4 5 6]" {
			t.Errorf("line %q", r.Lines[0])
		}
		if metric(t, r, "highest id") != 6 || metric(t, r, "unused numbers below the highest") != 0 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	t.Run("a server down skips its numbers but stays unique", func(t *testing.T) {
		r, _ := TicketServer(TicketConfig{Servers: 2, IDs: 6, DownAt: 2})
		if r.Lines[0] != "ids handed out: [1 2 4 6 8 10]" {
			t.Errorf("line %q", r.Lines[0])
		}
		if metric(t, r, "highest id") != 10 || metric(t, r, "unused numbers below the highest") != 4 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	t.Run("one server cannot fail over", func(t *testing.T) {
		r, _ := TicketServer(TicketConfig{Servers: 1, IDs: 3, DownAt: 1})
		if r.Lines[0] != "ids handed out: [1 2 3]" {
			t.Errorf("line %q", r.Lines[0])
		}
	})
	t.Run("a long run shows the first twenty", func(t *testing.T) {
		r, _ := TicketServer(TicketConfig{Servers: 3, IDs: 100})
		if !strings.HasPrefix(r.Lines[0], "first ids handed out: [1 2 3 4") || metric(t, r, "ids") != 100 || metric(t, r, "highest id") != 100 {
			t.Errorf("%v %v", r.Lines[0], r.Metrics)
		}
	})
	for _, c := range []TicketConfig{{0, 1, 0}, {17, 1, 0}, {1, 0, 0}, {1, 1, -1}} {
		if _, err := TicketServer(c); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", c, err)
		}
	}
}

func TestReportHelpers(t *testing.T) {
	r := Report{}
	r.say("line %d", 1)
	r.put("x", 2.5)
	if got := r.String(); got != "line 1\nx: 2.5\n" {
		t.Errorf("String %q", got)
	}
	if _, ok := r.Metric("missing"); ok {
		t.Error("a missing metric was found")
	}
	a := arch([]Component{{ID: "a"}, {ID: "b"}}, "a>b:call")
	if a.link("b", "a") != -1 || a.link("a", "b") != 0 || a.Links[0].Label != "call" {
		t.Errorf("links %v", a.Links)
	}
	f := a.frame("x")
	a.send(f, "b", "a", 1, OK)
	if len(f.Requests) != 0 {
		t.Error("a request on a missing link was recorded")
	}
	if percent(5, 0) != 0 || percent(1, 3) != 33 {
		t.Error("percent")
	}
	next := popular(nil, 1)
	for i := 0; i < 3; i++ {
		if next() != 0 {
			t.Error("one key is always key 0")
		}
	}
	if mustBase62(3844) != "100" {
		t.Errorf("mustBase62(3844) = %q", mustBase62(3844))
	}
	for _, s := range []string{"", "ab-c"} {
		if _, err := FromBase62(s); !errors.Is(err, ErrInput) {
			t.Errorf("FromBase62(%q): err %v", s, err)
		}
	}
	if _, err := Base62(-1); !errors.Is(err, ErrInput) {
		t.Errorf("Base62(-1): %v", err)
	}
	if s := (SagaConfig{FailAt: 2, Orchestration: true}).String(); s != "saga failing at 2 (orchestration true)" {
		t.Errorf("saga %q", s)
	}
	if s := (Models{Linearizable: true, Eventual: true}).String(); s != "linearizable true, sequential false, causal false, eventual true" {
		t.Errorf("models %q", s)
	}
}

func TestLeaderboardUnknownPlayer(t *testing.T) {
	lb := NewLeaderboard()
	if r, ok := lb.Rank("nobody"); ok || r != 0 {
		t.Errorf("rank of an unknown player %d %v", r, ok)
	}
	lb.Add("a", 1)
	if r, ok := lb.Rank("a"); !ok || r != 1 || lb.Score("a") != 1 || len(lb.Top(5)) != 1 {
		t.Errorf("rank %d %v, top %v", r, ok, lb.Top(5))
	}
}
