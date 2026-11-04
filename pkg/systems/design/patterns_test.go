package design

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// checkArch fails when an architecture names a missing component, puts a
// request on a missing link, or records more frames than maxFrames.
func checkArch(t *testing.T, r Report) {
	t.Helper()
	if r.Arch == nil || len(r.Arch.Frames) == 0 || len(r.Arch.Frames) > maxFrames {
		t.Fatalf("architecture %v", r.Arch)
	}
	ids := map[string]bool{}
	for _, c := range r.Arch.Components {
		ids[c.ID] = true
	}
	for _, l := range r.Arch.Links {
		if !ids[l.From] || !ids[l.To] {
			t.Errorf("link %v names a missing component", l)
		}
	}
	for _, f := range r.Arch.Frames {
		for _, q := range f.Requests {
			if q.Link < 0 || q.Link >= len(r.Arch.Links) {
				t.Errorf("request on link %d", q.Link)
			}
		}
	}
}

func TestCAP(t *testing.T) {
	t.Run("a partition costs CP writes and AP consistency", func(t *testing.T) {
		r, err := CAP(CAPConfig{Ops: 10, PartitionAt: 2, HealAt: 6, SyncMs: 5, Seed: 3})
		if err != nil {
			t.Fatal(err)
		}
		checkArch(t, r)
		if got := metric(t, r, "AP divergent writes"); got != 4 {
			t.Errorf("AP divergent writes %v, want the 4 ops inside the partition", got)
		}
		if got := metric(t, r, "CP extra latency (ms)"); got != 30 {
			t.Errorf("CP waited %v ms, want 6 ops outside the partition x 5 ms", got)
		}
		refused := metric(t, r, "CP refused writes")
		if refused < 0 || refused > 4 {
			t.Errorf("CP refused %v of 4 partitioned writes", refused)
		}
		if c := metric(t, r, "AP conflicts"); c != 0 && c != 1 {
			t.Errorf("conflicts %v: one heal gives at most one", c)
		}
	})
	t.Run("both sides writing during the partition conflict on healing", func(t *testing.T) {
		found := false
		for seed := int64(0); seed < 20 && !found; seed++ {
			r, _ := CAP(CAPConfig{Ops: 20, PartitionAt: 0, HealAt: 10, Seed: seed})
			found = metric(t, r, "AP conflicts") == 1
		}
		if !found {
			t.Error("ten random partitioned writes never hit both sides")
		}
	})
	t.Run("no partition refuses nothing", func(t *testing.T) {
		r, _ := CAP(CAPConfig{Ops: 50, SyncMs: 2})
		if metric(t, r, "CP refused writes") != 0 || metric(t, r, "AP divergent writes") != 0 || metric(t, r, "CP extra latency (ms)") != 100 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	t.Run("a partition that never heals never reconciles", func(t *testing.T) {
		r, _ := CAP(CAPConfig{Ops: 30, PartitionAt: 0, HealAt: 30})
		if metric(t, r, "AP conflicts") != 0 || metric(t, r, "AP divergent writes") != 30 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	for _, c := range []CAPConfig{{Ops: 0}, {Ops: 10, PartitionAt: 5, HealAt: 4}, {Ops: 10, SyncMs: -1}, {Ops: 10, PartitionAt: -1}} {
		if _, err := CAP(c); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", c, err)
		}
	}
}

func TestCQRS(t *testing.T) {
	c := CQRSConfig{Commands: 500, Queries: 200, LagMax: 5, Seed: 2}
	r, err := CQRS(c)
	if err != nil {
		t.Fatal(err)
	}
	checkArch(t, r)
	if got := metric(t, r, "events"); got != 500 {
		t.Errorf("events %v, want one per command", got)
	}
	stale := metric(t, r, "stale queries")
	if stale <= 0 || stale > 200 {
		t.Errorf("stale queries %v of 200 with a lag of up to 5", stale)
	}
	orders := 0
	for _, row := range r.Table.Rows {
		orders += row.Cells[0]
	}
	if orders < 1 || orders > 500 {
		t.Errorf("%d orders from 500 commands", orders)
	}
	if !strings.Contains(r.Lines[0], "counts "+strconv.Itoa(orders)+" orders") {
		t.Errorf("line %q, want %d orders", r.Lines[0], orders)
	}
	again, _ := CQRS(c)
	if !reflect.DeepEqual(r, again) {
		t.Error("the same seed gave different runs")
	}
	if zero, _ := CQRS(CQRSConfig{Commands: 100, Queries: 100}); metric(t, zero, "stale queries") != 0 {
		t.Error("a projection with no lag is never stale")
	}
	for _, bad := range []CQRSConfig{{}, {Commands: 1, Queries: -1}, {Commands: 1, LagMax: -1}} {
		if _, err := CQRS(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestTransactionalOutbox(t *testing.T) {
	cases := []struct {
		name                   string
		crash                  int
		outbox                 bool
		lost, dupes, delivered float64
	}{
		{"dual write with no crash", 0, false, 0, 0, 100},
		{"dual write always crashing loses everything", 100, false, 100, 0, 0},
		{"outbox always crashing duplicates everything", 100, true, 0, 100, 100},
		{"outbox with no crash", 0, true, 0, 0, 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := TransactionalOutbox(OutboxConfig{Orders: 100, CrashPercent: c.crash, Outbox: c.outbox, Seed: 1})
			if err != nil {
				t.Fatal(err)
			}
			checkArch(t, r)
			if metric(t, r, "lost events") != c.lost || metric(t, r, "duplicates") != c.dupes || metric(t, r, "delivered") != c.delivered {
				t.Errorf("metrics %v", r.Metrics)
			}
		})
	}
	t.Run("every order is lost or delivered", func(t *testing.T) {
		for _, outbox := range []bool{false, true} {
			r, _ := TransactionalOutbox(OutboxConfig{Orders: 300, CrashPercent: 30, Outbox: outbox, Seed: 4})
			if metric(t, r, "lost events")+metric(t, r, "delivered") != 300 {
				t.Errorf("outbox %v: %v", outbox, r.Metrics)
			}
			if outbox && metric(t, r, "lost events") != 0 {
				t.Error("the outbox lost an event")
			}
			if !outbox && metric(t, r, "lost events") == 0 {
				t.Error("a 30% crash rate lost nothing without the outbox")
			}
		}
	})
	for _, bad := range []OutboxConfig{{}, {Orders: 1, CrashPercent: 101}, {Orders: 1, CrashPercent: -1}} {
		if _, err := TransactionalOutbox(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestChangeDataCapture(t *testing.T) {
	r, err := ChangeDataCapture(CDCConfig{Changes: 400, Rows: 20, LagTicks: 7, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	checkArch(t, r)
	if metric(t, r, "changes") != 400 || metric(t, r, "consumer lag") != 7 || metric(t, r, "out of order") != 0 {
		t.Errorf("metrics %v", r.Metrics)
	}
	if !strings.Contains(r.Lines[1], "at offset 393, 7 behind") {
		t.Errorf("line %q", r.Lines[1])
	}
	short, _ := ChangeDataCapture(CDCConfig{Changes: 3, Rows: 1, LagTicks: 50})
	if metric(t, short, "consumer lag") != 3 {
		t.Errorf("the lag cannot exceed the changes: %v", short.Metrics)
	}
	for _, bad := range []CDCConfig{{}, {Changes: 1}, {Changes: 1, Rows: 1, LagTicks: -1}} {
		if _, err := ChangeDataCapture(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestLeaderFollower(t *testing.T) {
	t.Run("async followers serve stale reads", func(t *testing.T) {
		r, err := LeaderFollower(ReplicationConfig{Followers: 3, MaxLag: 10, Writes: 500, Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		checkArch(t, r)
		if s := metric(t, r, "stale reads"); s == 0 || s > 500 {
			t.Errorf("stale reads %v", s)
		}
		if metric(t, r, "reads on the leader") != 0 {
			t.Error("reads went to the leader without read-your-writes")
		}
	})
	t.Run("read-your-writes is never stale", func(t *testing.T) {
		r, _ := LeaderFollower(ReplicationConfig{Followers: 3, MaxLag: 10, Writes: 500, ReadYourWrite: true, Seed: 1})
		checkArch(t, r)
		if metric(t, r, "stale reads") != 0 || metric(t, r, "reads on the leader") != 500 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	t.Run("no lag is never stale", func(t *testing.T) {
		r, _ := LeaderFollower(ReplicationConfig{Followers: 1, Writes: 50})
		if metric(t, r, "stale reads") != 0 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	for _, bad := range []ReplicationConfig{{Writes: 1}, {Followers: 51, Writes: 1}, {Followers: 1}, {Followers: 1, Writes: 1, MaxLag: -1}} {
		if _, err := LeaderFollower(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestMultiLeader(t *testing.T) {
	lww, err := MultiLeader(MultiLeaderConfig{Keys: 20, Writes: 5, Rounds: 30, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	vc, _ := MultiLeader(MultiLeaderConfig{Keys: 20, Writes: 5, Rounds: 30, VectorClock: true, Seed: 1})
	checkArch(t, lww)
	checkArch(t, vc)
	c := metric(t, lww, "conflicts")
	if c == 0 || metric(t, vc, "conflicts") != c {
		t.Errorf("conflicts %v and %v: the same seed writes the same keys", c, metric(t, vc, "conflicts"))
	}
	if metric(t, lww, "lost updates") != c || metric(t, lww, "siblings") != 0 {
		t.Errorf("last write wins: %v", lww.Metrics)
	}
	if metric(t, vc, "lost updates") != 0 || metric(t, vc, "siblings") != c {
		t.Errorf("version vectors: %v", vc.Metrics)
	}
	if quiet, _ := MultiLeader(MultiLeaderConfig{Keys: 5, Rounds: 3}); metric(t, quiet, "conflicts") != 0 {
		t.Error("no writes, no conflicts")
	}
	for _, bad := range []MultiLeaderConfig{{Rounds: 1}, {Keys: 1}, {Keys: 1, Rounds: 1, Writes: -1}} {
		if _, err := MultiLeader(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestSharding(t *testing.T) {
	r, err := Sharding(ShardingConfig{Keys: 10_000, Shards: 8, Writes: 1000, HotPercent: 80, Reads: 5000, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	rng := metric(t, r, "range busiest (%)")
	hash := metric(t, r, "hash busiest (%)")
	dir := metric(t, r, "directory (ranges rebalanced by load) busiest (%)")
	if dir != 12 {
		t.Errorf("a directory splits 6,000 touches into 8 equal ranges: busiest %v%%, want 12", dir)
	}
	if rng < 50 || hash > 20 || rng <= hash {
		t.Errorf("busiest: range %v%%, hash %v%%", rng, hash)
	}
	for _, row := range r.Table.Rows {
		if row.Cells[0] < row.Cells[1] {
			t.Errorf("%s: busiest %d below quietest %d", row.Label, row.Cells[0], row.Cells[1])
		}
	}
	for _, bad := range []ShardingConfig{{Keys: 1, Shards: 1}, {Shards: 2}, {Keys: 1, Shards: 2, HotPercent: 101}} {
		if _, err := Sharding(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestBackpressure(t *testing.T) {
	base := BackpressureConfig{ServicePerTick: 10, ArrivePerTick: 5, BurstPerTick: 30, BurstTicks: 5, Ticks: 40, Limit: 20}
	run := func(m Overload) Report {
		t.Helper()
		c := base
		c.Mode = m
		r, err := Backpressure(c)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	unb, shed, push := run(Unbounded), run(LoadShedding), run(PushBack)
	// the burst adds 20 a tick for 5 ticks: the unbounded queue peaks at 100, 10 ticks of latency
	if metric(t, unb, "worst latency (ticks)") != 10 || metric(t, unb, "dropped") != 0 || metric(t, unb, "pushed back") != 0 {
		t.Errorf("unbounded %v", unb.Metrics)
	}
	if w := metric(t, shed, "worst latency (ticks)"); w > 2 || metric(t, shed, "dropped") == 0 || metric(t, shed, "pushed back") != 0 {
		t.Errorf("load shedding %v", shed.Metrics)
	}
	if w := metric(t, push, "worst latency (ticks)"); w > 2 || metric(t, push, "dropped") != 0 || metric(t, push, "pushed back") == 0 {
		t.Errorf("backpressure %v", push.Metrics)
	}
	// conservation: arrivals = served + queued + dropped, with pushed-back work retried
	for _, r := range []Report{unb, shed} {
		in, last := 0, r.Table.Rows[len(r.Table.Rows)-1]
		for _, row := range r.Table.Rows {
			in += row.Cells[0]
		}
		if in != 5*30+35*5 {
			t.Errorf("arrivals %d, want %d", in, 5*30+35*5)
		}
		if last.Cells[1] != 0 {
			t.Errorf("the queue should drain by the end: %v", last)
		}
	}
	for _, m := range []Overload{-1, 3} {
		c := base
		c.Mode = m
		if _, err := Backpressure(c); !errors.Is(err, ErrInput) {
			t.Errorf("mode %d: err %v", m, err)
		}
	}
}

func TestAutocomplete(t *testing.T) {
	terms, err := ParseTerms("Pizza 10\npizza hut 5\n\n  pasta 7  \npho 2\nPIZZA 3\nthe 42 club")
	if err != nil {
		t.Fatal(err)
	}
	want := []Term{{"pizza", 10}, {"pizza hut", 5}, {"pasta", 7}, {"pho", 2}, {"pizza", 3}, {"the 42 club", 1}}
	if !reflect.DeepEqual(terms, want) {
		t.Errorf("terms %v, want %v", terms, want)
	}
	r, err := Autocomplete(terms, "Pi", 2)
	if err != nil {
		t.Fatal(err)
	}
	checkArch(t, r)
	if r.Lines[0] != `"p": pizza (13), pasta (7)` || r.Lines[1] != `"pi": pizza (13), pizza hut (5)` {
		t.Errorf("lines %q", r.Lines)
	}
	// prefixes: pizza 5, "pizza hut" 4 more, pasta 4 more, pho 2 more, "the 42 club" 11
	if metric(t, r, "prefixes stored") != 26 || metric(t, r, "lookups") != 2 || metric(t, r, "terms a scan would read") != 4+2 {
		t.Errorf("metrics %v", r.Metrics)
	}
	none, _ := Autocomplete(terms, "zz", 3)
	if none.Lines[0] != `"z": no suggestions` {
		t.Errorf("line %q", none.Lines[0])
	}

	for name, text := range map[string]string{"empty": "", "zero count": "pizza 0", "negative count": "pizza -2"} {
		if _, err := ParseTerms(text); !errors.Is(err, ErrInput) {
			t.Errorf("%s: err %v", name, err)
		}
	}
	for _, k := range []int{0, 21} {
		if _, err := Autocomplete(terms, "p", k); !errors.Is(err, ErrInput) {
			t.Errorf("top %d: err %v", k, err)
		}
	}
	if _, err := Autocomplete(terms, "", 3); !errors.Is(err, ErrInput) {
		t.Errorf("nothing typed: %v", err)
	}
}

func TestProximity(t *testing.T) {
	t.Run("a radius inside a cell misses nothing", func(t *testing.T) {
		r, err := Proximity(ProximityConfig{Businesses: 3000, Precision: 5, Lat: 40.7, Lon: -74, RadiusKm: 1, Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		checkArch(t, r)
		cand, found := metric(t, r, "candidates read"), metric(t, r, "found")
		if metric(t, r, "missed (radius larger than the cells)") != 0 {
			t.Errorf("missed %v", r.Metrics)
		}
		if found == 0 || found > cand || cand > 3000 || metric(t, r, "full scan reads") != 3000 {
			t.Errorf("found %v of %v candidates", found, cand)
		}
		g, _ := Geohash(40.7, -74, 5)
		if !strings.Contains(r.Lines[0], "cell "+g) {
			t.Errorf("line %q, want cell %s", r.Lines[0], g)
		}
	})
	t.Run("a radius larger than the cells misses some", func(t *testing.T) {
		r, _ := Proximity(ProximityConfig{Businesses: 3000, Precision: 8, Lat: 40.7, Lon: -74, RadiusKm: 5, Seed: 1})
		if metric(t, r, "missed (radius larger than the cells)") == 0 || !strings.Contains(strings.Join(r.Lines, "\n"), "use a shorter geohash") {
			t.Errorf("%v\n%s", r.Metrics, r)
		}
	})
	t.Run("near the pole the neighbors are clamped", func(t *testing.T) {
		cells := neighbors(90, 180, 3)
		if len(cells) == 0 || len(cells) > 9 {
			t.Errorf("cells %v", cells)
		}
	})
	if d := distanceKm(0, 0, 0, 1); !near(d, 111.19, 0.01) {
		t.Errorf("one degree of longitude at the equator is %v km", d)
	}
	if dLat, dLon := geohashCell(1); dLat != 45 || dLon != 45 {
		t.Errorf("a one-character cell is %v x %v degrees, want 45 x 45", dLat, dLon)
	}
	for _, bad := range []ProximityConfig{{Precision: 5, RadiusKm: 1}, {Businesses: 1, Precision: 2, RadiusKm: 1}, {Businesses: 1, Precision: 5}, {Businesses: 1, Precision: 5, RadiusKm: 1, Lat: 91}} {
		if _, err := Proximity(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestAdaptiveBitrate(t *testing.T) {
	fast, err := AdaptiveBitrate(ABRConfig{Seconds: 120, StartKbps: 20_000, BufferTarget: 10, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	slow, _ := AdaptiveBitrate(ABRConfig{Seconds: 120, StartKbps: 100, BufferTarget: 10, Seed: 1})
	onLadder := map[int]bool{}
	for _, b := range Ladder {
		onLadder[b] = true
	}
	for _, r := range []Report{fast, slow} {
		for _, row := range r.Table.Rows {
			if !onLadder[row.Cells[1]] {
				t.Errorf("%s: bitrate %d is not on the ladder", row.Label, row.Cells[1])
			}
			if row.Cells[0] < 150 || row.Cells[0] > 25_000 || row.Cells[2] < 0 || row.Cells[2] > 100 {
				t.Errorf("%s: cells %v out of bounds", row.Label, row.Cells)
			}
		}
		if len(r.Table.Rows) != 60 {
			t.Errorf("%d rows, want the first 60 seconds", len(r.Table.Rows))
		}
	}
	if metric(t, fast, "average bitrate (kbps)") <= metric(t, slow, "average bitrate (kbps)") {
		t.Errorf("fast %v, slow %v", fast.Metrics, slow.Metrics)
	}
	if metric(t, slow, "stalls") == 0 {
		t.Errorf("150 kbps cannot carry the 400 kbps step without stalling: %v", slow.Metrics)
	}
	again, _ := AdaptiveBitrate(ABRConfig{Seconds: 120, StartKbps: 20_000, BufferTarget: 10, Seed: 1})
	if !reflect.DeepEqual(fast, again) {
		t.Error("the same seed gave different runs")
	}
	for _, bad := range []ABRConfig{{StartKbps: 100, BufferTarget: 1}, {Seconds: 1, StartKbps: 99, BufferTarget: 1}, {Seconds: 1, StartKbps: 100}} {
		if _, err := AdaptiveBitrate(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%+v: err %v", bad, err)
		}
	}
}

func TestStudies(t *testing.T) {
	s := Studies()
	if len(s) == 0 {
		t.Fatal("no studies")
	}
	ids := map[string]bool{}
	for _, st := range s {
		if st.ID == "" || st.Title == "" || st.Pitch == "" || len(st.Functional) == 0 {
			t.Errorf("incomplete study %+v", st)
		}
		if ids[st.ID] {
			t.Errorf("duplicate study %s", st.ID)
		}
		ids[st.ID] = true
		switch st.Level {
		case "beginner", "intermediate", "advanced":
		default:
			t.Errorf("%s: level %q", st.ID, st.Level)
		}
	}
	s[0].ID = "changed"
	if Studies()[0].ID == "changed" {
		t.Error("Studies returned the shared table, not a copy")
	}
}

func TestKeyValueStoreQuorums(t *testing.T) {
	t.Run("R + W > N with every node up is never stale", func(t *testing.T) {
		r, err := KeyValueStore(KVConfig{Nodes: 5, Replicas: 3, R: 2, W: 2, Keys: 30, Ops: 400, DownNode: -1, Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		checkArch(t, r)
		for _, name := range []string{"stale reads", "failed writes", "hinted writes", "hints handed off"} {
			if v := metric(t, r, name); v != 0 {
				t.Errorf("%s = %v with every node up", name, v)
			}
		}
		if !strings.Contains(r.Lines[0], "R + W = 4 > N = 3") || !strings.Contains(r.Lines[2], "0 failed") {
			t.Errorf("lines %q", r.Lines)
		}
	})
	t.Run("a sloppy quorum hands every hint back", func(t *testing.T) {
		r, _ := KeyValueStore(KVConfig{Nodes: 5, Replicas: 3, R: 2, W: 2, Keys: 30, Ops: 600, DownNode: 1, Seed: 2})
		h := metric(t, r, "hinted writes")
		if h == 0 || metric(t, r, "hints handed off") != h || metric(t, r, "failed writes") != 0 {
			t.Errorf("metrics %v", r.Metrics)
		}
	})
	t.Run("no stand-in node fails writes and reads while a node is down", func(t *testing.T) {
		r, err := KeyValueStore(KVConfig{Nodes: 3, Replicas: 3, R: 3, W: 3, Keys: 10, Ops: 300, DownNode: 0, Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		if metric(t, r, "failed writes") == 0 || strings.Contains(r.Lines[2], " 0 failed") {
			t.Errorf("R = W = N with a node down should fail: %v %q", r.Metrics, r.Lines[2])
		}
		if !strings.Contains(r.Lines[0], "R + W = 6 > N = 3") {
			t.Errorf("line %q", r.Lines[0])
		}
	})
}
