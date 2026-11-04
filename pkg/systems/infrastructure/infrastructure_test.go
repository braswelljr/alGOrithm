package infrastructure

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestCircuitBreaker(t *testing.T) {
	r, err := CircuitBreaker("fail fail fail ok ok ok ok ok fail ok ok", 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []BreakerState{Closed, Closed, Closed, Open, Open, HalfOpen, Closed}
	for i, s := range want {
		if r.States[i] != s {
			t.Fatalf("state %d is %s, want %s: %v", i, r.States[i], s, r.States)
		}
	}
	if r.Rejected != 2 {
		t.Errorf("rejected %d, want 2", r.Rejected)
	}
}

func TestBalancingAdaptsToSpeed(t *testing.T) {
	rr, _ := Balance(RoundRobin, []int{1, 4}, 100, 1)
	lc, _ := Balance(LeastConnections, []int{1, 4}, 100, 1)
	if lc.Latency >= rr.Latency {
		t.Errorf("least connections %.1f should beat round robin %.1f on uneven servers", lc.Latency, rr.Latency)
	}
}

func TestJitterSpreadsRetries(t *testing.T) {
	none, _ := Backoff(100, 4, 2, 64, NoJitter, 1)
	full, _ := Backoff(100, 4, 2, 64, FullJitter, 1)
	if none.Peak != 100 || full.Peak >= none.Peak {
		t.Errorf("peaks %d and %d", none.Peak, full.Peak)
	}
}

func TestDeliveryGuarantees(t *testing.T) {
	most, _ := Delivery(AtMostOnce, 200, 20, 1)
	least, _ := Delivery(AtLeastOnce, 200, 20, 1)
	exact, _ := Delivery(ExactlyOnce, 200, 20, 1)
	if most.Duplicates != 0 || most.Lost == 0 || least.Lost != 0 || least.Duplicates == 0 || exact.Lost != 0 || exact.Duplicates != 0 {
		t.Errorf("%+v\n%+v\n%+v", most, least, exact)
	}
}

func TestStampedeFixes(t *testing.T) {
	none, _ := Stampede(20, 60, 20, 3, NoFix, 1)
	lock, _ := Stampede(20, 60, 20, 3, LockFix, 1)
	if none.Peak != 20 || lock.Peak != 1 {
		t.Errorf("peaks %d and %d", none.Peak, lock.Peak)
	}
}

func TestCgroups(t *testing.T) {
	cs, err := Cgroups(2, "web 1024 2 512 300\nbatch 512 2 256 100\nleak 1024 1 128 400")
	if err != nil {
		t.Fatal(err)
	}
	if !cs[2].Killed || cs[2].Got != 0 {
		t.Errorf("leak should be OOM-killed: %v", cs[2])
	}
	if d := cs[0].Got + cs[1].Got; d < 1.999 || cs[0].Got < 1.33 || cs[0].Got > 1.34 {
		t.Errorf("web got %.3f, batch %.3f", cs[0].Got, cs[1].Got)
	}
	if !strings.Contains(cs[2].String(), "OOM") {
		t.Error(cs[2].String())
	}
}

func TestCanaryLimitsTheDamage(t *testing.T) {
	bg, _ := Deploy(BlueGreen, 10, 1000, 30, true)
	cn, _ := Deploy(Canary, 10, 1000, 30, true)
	if !bg.RolledBack || !cn.RolledBack || cn.Errors >= bg.Errors {
		t.Errorf("blue/green %d errors, canary %d", bg.Errors, cn.Errors)
	}
}

func TestAutoscale(t *testing.T) {
	t.Run("never exceeds the maximum or goes under the minimum", func(t *testing.T) {
		load := []int{10, 50, 400, 900, 900, 900, 200, 10, 10, 10, 10, 10, 10}
		r, err := Autoscale(load, 100, 70, 2, 3, 1, 8)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Steps) != len(load) {
			t.Fatalf("%d steps for %d load points", len(r.Steps), len(load))
		}
		dropped, cost := 0, 0
		for i, s := range r.Steps {
			if s.Instances+s.Booting > 8 {
				t.Errorf("tick %d: %d serving + %d booting is over the maximum 8", i, s.Instances, s.Booting)
			}
			if s.Instances < 1 {
				t.Errorf("tick %d: %d serving is under the minimum", i, s.Instances)
			}
			if s.Load != load[i] {
				t.Errorf("tick %d: load %d, want %d", i, s.Load, load[i])
			}
			if want := max(0, s.Load-s.Instances*100); s.Dropped != want {
				t.Errorf("tick %d: dropped %d, want %d", i, s.Dropped, want)
			}
			dropped += s.Dropped
			cost += s.Instances + s.Booting
		}
		if dropped != r.Dropped || cost != r.Cost {
			t.Errorf("totals %d dropped, %d cost; steps sum to %d and %d", r.Dropped, r.Cost, dropped, cost)
		}
		if r.Dropped == 0 {
			t.Error("a spike that outruns the boot time should drop requests")
		}
		if last := r.Steps[len(r.Steps)-1]; last.Instances != 1 {
			t.Errorf("after the load falls back it should scale in to 1, has %d", last.Instances)
		}
		if len(r.Log) == 0 {
			t.Error("no log")
		}
	})

	t.Run("drops nothing when the minimum capacity suffices", func(t *testing.T) {
		r, err := Autoscale([]int{100, 200, 150, 300}, 100, 100, 5, 0, 3, 10)
		if err != nil {
			t.Fatal(err)
		}
		if r.Dropped != 0 {
			t.Errorf("dropped %d with 3 instances of 100 for loads up to 300", r.Dropped)
		}
		if r.Cost != 12 {
			t.Errorf("cost %d, want 3 instances for 4 ticks = 12", r.Cost)
		}
	})

	t.Run("instant boot serves a spike at once", func(t *testing.T) {
		r, err := Autoscale([]int{0, 1000, 1000}, 100, 100, 0, 0, 1, 20)
		if err != nil {
			t.Fatal(err)
		}
		if r.Dropped != 0 {
			t.Errorf("dropped %d with boot 0", r.Dropped)
		}
		if r.Steps[1].Instances != 10 || r.Steps[1].Booting != 0 {
			t.Errorf("tick 1: %+v, want 10 serving and none booting", r.Steps[1])
		}
	})

	t.Run("cooldown delays scaling in", func(t *testing.T) {
		load := []int{500, 100, 100, 100, 100}
		r, err := Autoscale(load, 100, 100, 0, 3, 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		got := []int{}
		for _, s := range r.Steps {
			got = append(got, s.Instances)
		}
		if want := []int{5, 5, 5, 1, 1}; !reflect.DeepEqual(got, want) {
			t.Errorf("instances %v, want %v", got, want)
		}
	})

	t.Run("booting instances start serving after boot ticks", func(t *testing.T) {
		r, err := Autoscale([]int{300, 300, 300, 300}, 100, 100, 2, 0, 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if s := r.Steps[0]; s.Instances != 1 || s.Booting != 2 || s.Dropped != 200 {
			t.Errorf("tick 0: %+v", s)
		}
		if s := r.Steps[2]; s.Instances != 3 || s.Booting != 0 || s.Dropped != 0 {
			t.Errorf("tick 2: %+v", s)
		}
	})

	bad := []struct {
		name string
		load []int
		args [6]int
	}{
		{"no load", nil, [6]int{100, 70, 1, 1, 1, 5}},
		{"zero capacity", []int{1}, [6]int{0, 70, 1, 1, 1, 5}},
		{"target too low", []int{1}, [6]int{100, 5, 1, 1, 1, 5}},
		{"negative boot", []int{1}, [6]int{100, 70, -1, 1, 1, 5}},
		{"maximum under minimum", []int{1}, [6]int{100, 70, 1, 1, 5, 2}},
		{"negative load", []int{5, -1}, [6]int{100, 70, 1, 1, 1, 5}},
	}
	for _, c := range bad {
		t.Run(c.name, func(t *testing.T) {
			a := c.args
			if _, err := Autoscale(c.load, a[0], a[1], a[2], a[3], a[4], a[5]); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestCacheStrategy(t *testing.T) {
	// After every write, a read must see the new value, never a stale one,
	// for cache-aside and write-through.
	script := "write a 1\nread a\nread a\nwrite a 2\nread a\nread b\n\nwrite b x y\nread b"
	for _, p := range []CachePolicy{CacheAside, WriteThrough} {
		t.Run(string(p)+" never serves a stale read", func(t *testing.T) {
			r, err := CacheStrategy(p, script)
			if err != nil {
				t.Fatal(err)
			}
			// The read after "write a 2" must return "2".
			readAfter := ""
			seenSecondWrite := false
			writes := 0
			for _, l := range r.Log {
				if strings.HasPrefix(l, "write a") {
					writes++
					seenSecondWrite = writes == 2
					continue
				}
				if seenSecondWrite && strings.HasPrefix(l, "read a") {
					readAfter = l
					break
				}
			}
			if !strings.Contains(readAfter, `"2"`) {
				t.Errorf("read after the second write: %q", readAfter)
			}
			if !strings.Contains(strings.Join(r.Log, "\n"), `"x y"`) {
				t.Error("a multi-word value should be joined with spaces")
			}
			if r.DBWrites != 3 {
				t.Errorf("%d database writes, want 3", r.DBWrites)
			}
			if r.Lost != 0 {
				t.Errorf("lost %d", r.Lost)
			}
		})
	}

	t.Run("cache-aside misses the first read after a write", func(t *testing.T) {
		r, _ := CacheStrategy(CacheAside, "write a 1\nread a\nread a")
		if r.Hits != 1 || r.DBReads != 1 {
			t.Errorf("hits %d, db reads %d; want 1 and 1", r.Hits, r.DBReads)
		}
	})

	t.Run("write-through keeps the cache warm", func(t *testing.T) {
		r, _ := CacheStrategy(WriteThrough, "write a 1\nread a\nread a")
		if r.Hits != 2 || r.DBReads != 0 {
			t.Errorf("hits %d, db reads %d; want 2 and 0", r.Hits, r.DBReads)
		}
	})

	t.Run("write-behind batches writes and loses them in a crash", func(t *testing.T) {
		r, err := CacheStrategy(WriteBehind, "write a 1\nwrite a 2\nwrite a 3\nwrite b 1\nflush\nwrite c 1\nwrite d 1\ncrash\nread c\nread a")
		if err != nil {
			t.Fatal(err)
		}
		if r.DBWrites != 2 {
			t.Errorf("%d database writes, want 2 (a and b once each)", r.DBWrites)
		}
		if r.Lost != 2 {
			t.Errorf("lost %d, want 2 (c and d)", r.Lost)
		}
		if r.DBReads != 2 || r.Hits != 0 {
			t.Errorf("db reads %d, hits %d; want 2 and 0 after the crash", r.DBReads, r.Hits)
		}
		log := strings.Join(r.Log, "\n")
		if !strings.Contains(log, "read c: miss, and not in the database") || !strings.Contains(log, `read "3"`) {
			t.Errorf("log:\n%s", log)
		}
	})

	t.Run("unknown policy", func(t *testing.T) {
		if _, err := CacheStrategy("lazy", "read a"); !errors.Is(err, ErrInput) {
			t.Errorf("err %v", err)
		}
	})
	for _, line := range []string{"read", "write a", "evict a", "read a b"} {
		t.Run("bad line "+line, func(t *testing.T) {
			_, err := CacheStrategy(CacheAside, "read x\n"+line)
			if !errors.Is(err, ErrInput) || !strings.Contains(err.Error(), "line 2") {
				t.Errorf("err %v, want ErrInput naming line 2", err)
			}
		})
	}
}

func TestCDN(t *testing.T) {
	t.Run("origin requests plus edge hits equal requests", func(t *testing.T) {
		for _, ttl := range []int{1, 10, 100, 5000} {
			r, err := CDN(4, 200, 3000, ttl, 7)
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Regions) != 4 {
				t.Fatalf("%d regions", len(r.Regions))
			}
			req, hits, misses := 0, 0, 0
			for _, s := range r.Regions {
				if s.Hits+s.Misses != s.Requests {
					t.Errorf("%s: %d hits + %d misses != %d requests", s.Region, s.Hits, s.Misses, s.Requests)
				}
				req += s.Requests
				hits += s.Hits
				misses += s.Misses
			}
			if req != 3000 || r.Origin+hits != 3000 || r.Origin != misses {
				t.Errorf("ttl %d: requests %d, origin %d, hits %d, misses %d", ttl, req, r.Origin, hits, misses)
			}
			if len(r.Log) != 5 {
				t.Errorf("%d log lines, want one per region and a total", len(r.Log))
			}
		}
	})

	t.Run("ttl 1 never hits", func(t *testing.T) {
		r, _ := CDN(2, 5, 500, 1, 1)
		if r.Origin != 500 {
			t.Errorf("origin %d, want every request", r.Origin)
		}
	})

	t.Run("a longer ttl never hits less", func(t *testing.T) {
		short, _ := CDN(3, 100, 2000, 5, 3)
		long, _ := CDN(3, 100, 2000, 5000, 3)
		if long.Origin > short.Origin {
			t.Errorf("ttl 5000 origin %d > ttl 5 origin %d", long.Origin, short.Origin)
		}
		// With a ttl longer than the run, each region misses each object once at most.
		if long.Origin > 3*100 {
			t.Errorf("origin %d is more than regions * objects", long.Origin)
		}
	})

	t.Run("one object is a miss per region, then hits", func(t *testing.T) {
		r, _ := CDN(1, 1, 100, 5000, 9)
		if r.Origin != 1 || r.Regions[0].Hits != 99 || r.Regions[0].Region != "us-east" {
			t.Errorf("%+v", r)
		}
	})

	t.Run("seeded runs repeat", func(t *testing.T) {
		a, _ := CDN(6, 50, 1000, 20, 42)
		b, _ := CDN(6, 50, 1000, 20, 42)
		if !reflect.DeepEqual(a, b) {
			t.Error("the same seed gave different runs")
		}
	})

	t.Run("an unused region reports zero", func(t *testing.T) {
		r, _ := CDN(6, 1, 1, 1, 1)
		zero := 0
		for _, s := range r.Regions {
			if s.Requests == 0 {
				zero++
			}
		}
		if zero != 5 {
			t.Errorf("%d idle regions, want 5", zero)
		}
	})

	for _, c := range [][5]int{{0, 1, 1, 1}, {7, 1, 1, 1}, {1, 0, 1, 1}, {1, 1, 0, 1}, {1, 1, 1, 0}, {1, 1001, 1, 1}} {
		if _, err := CDN(c[0], c[1], c[2], c[3], 1); !errors.Is(err, ErrInput) {
			t.Errorf("CDN%v: err %v, want ErrInput", c[:4], err)
		}
	}
}

func TestCgroupsSharing(t *testing.T) {
	t.Run("a container that wants little gives the rest away", func(t *testing.T) {
		cs, err := Cgroups(4, "small 1024 0.5 100 10\nbig 1024 8 100 10\nidle 1024 0 100 10")
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(cs[0].Got-0.5) > 1e-9 || math.Abs(cs[1].Got-3.5) > 1e-9 || cs[2].Got != 0 {
			t.Errorf("got %.3f %.3f %.3f, want 0.5 3.5 0", cs[0].Got, cs[1].Got, cs[2].Got)
		}
		if !strings.Contains(cs[1].String(), "3.50 of 8.00") {
			t.Error(cs[1].String())
		}
	})

	t.Run("no contention gives everyone what it wants", func(t *testing.T) {
		cs, _ := Cgroups(8, "a 2 1 10 1\nb 262144 2 10 1")
		if cs[0].Got != 1 || cs[1].Got != 2 {
			t.Errorf("got %v %v", cs[0].Got, cs[1].Got)
		}
	})

	for name, spec := range map[string]string{
		"too few fields": "a 1024 1 10",
		"low shares":     "a 1 1 10 1",
		"bad number":     "a x 1 10 1",
		"no containers":  "\n\n",
		"zero limit":     "a 1024 1 0 1",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Cgroups(2, spec); !errors.Is(err, ErrInput) {
				t.Errorf("err %v", err)
			}
		})
	}
	if _, err := Cgroups(0, "a 1024 1 10 1"); !errors.Is(err, ErrInput) {
		t.Errorf("0 CPUs: %v", err)
	}
}

func TestBalanceStrategies(t *testing.T) {
	for _, s := range []Strategy{RoundRobin, WeightedRobin, LeastConnections, PowerOfTwo, Random} {
		t.Run(string(s)+" serves every request", func(t *testing.T) {
			r, err := Balance(s, []int{1, 2, 3}, 120, 5)
			if err != nil {
				t.Fatal(err)
			}
			total := 0
			for _, n := range r.Served {
				total += n
			}
			if total != 120 {
				t.Errorf("served %d of 120", total)
			}
			if r.Latency < 1 || r.Peak < 1 || len(r.Log) != 4 {
				t.Errorf("latency %.2f, peak %d, %d log lines", r.Latency, r.Peak, len(r.Log))
			}
			again, _ := Balance(s, []int{1, 2, 3}, 120, 5)
			if !reflect.DeepEqual(r, again) {
				t.Error("the same seed gave different runs")
			}
		})
	}
	for name, f := range map[string]func() error{
		"unknown strategy": func() error { _, err := Balance("fastest", []int{1, 2}, 10, 1); return err },
		"one server":       func() error { _, err := Balance(RoundRobin, []int{1}, 10, 1); return err },
		"speed zero":       func() error { _, err := Balance(RoundRobin, []int{1, 0}, 10, 1); return err },
		"no requests":      func() error { _, err := Balance(RoundRobin, []int{1, 2}, 0, 1); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := f(); !errors.Is(err, ErrInput) {
				t.Errorf("err %v", err)
			}
		})
	}
}

func TestDeployStrategies(t *testing.T) {
	t.Run("rolling a good build reaches every instance with no errors", func(t *testing.T) {
		r, err := Deploy(Rolling, 4, 1000, 30, false)
		if err != nil {
			t.Fatal(err)
		}
		var shares []int
		for _, s := range r.Steps {
			shares = append(shares, s.NewShare)
		}
		if want := []int{25, 50, 75, 100, 100}; !reflect.DeepEqual(shares, want) || r.Errors != 0 || r.RolledBack {
			t.Errorf("shares %v errors %d rolled back %v", shares, r.Errors, r.RolledBack)
		}
	})
	t.Run("rolling a bad build rolls back between canary and blue-green", func(t *testing.T) {
		bg, _ := Deploy(BlueGreen, 10, 1000, 30, true)
		ro, _ := Deploy(Rolling, 10, 1000, 30, true)
		cn, _ := Deploy(Canary, 10, 1000, 30, true)
		if !ro.RolledBack || ro.Errors > bg.Errors || ro.Errors < cn.Errors {
			t.Errorf("blue/green %d, rolling %d, canary %d", bg.Errors, ro.Errors, cn.Errors)
		}
		if last := ro.Steps[len(ro.Steps)-1]; last.NewShare != 0 {
			t.Errorf("the rollback step has %d%% on the new version", last.NewShare)
		}
	})
	for name, f := range map[string]func() error{
		"unknown strategy": func() error { _, err := Deploy("big-bang", 1, 1, 1, true); return err },
		"no instances":     func() error { _, err := Deploy(Canary, 0, 1, 1, true); return err },
		"error rate 101":   func() error { _, err := Deploy(Canary, 1, 1, 101, true); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := f(); !errors.Is(err, ErrInput) {
				t.Errorf("err %v", err)
			}
		})
	}
}

func TestStampedeEarlyExpiry(t *testing.T) {
	early, err := Stampede(20, 100, 20, 3, EarlyExpiry, 1)
	if err != nil {
		t.Fatal(err)
	}
	none, _ := Stampede(20, 100, 20, 3, NoFix, 1)
	if early.Peak != 1 || early.Total == 0 || early.Total >= none.Total {
		t.Errorf("early peak %d total %d; none total %d", early.Peak, early.Total, none.Total)
	}
	again, _ := Stampede(20, 100, 20, 3, EarlyExpiry, 1)
	if !reflect.DeepEqual(early, again) {
		t.Error("the same seed gave different runs")
	}
	if _, err := Stampede(1, 1, 2, 1, "magic", 1); !errors.Is(err, ErrInput) {
		t.Errorf("unknown fix: %v", err)
	}
	if _, err := Stampede(1, 1, 5, 5, NoFix, 1); !errors.Is(err, ErrInput) {
		t.Errorf("recompute equal to ttl: %v", err)
	}
}

func TestBackoffEqualJitter(t *testing.T) {
	r, err := Backoff(50, 3, 2, 8, EqualJitter, 3)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 150 {
		t.Errorf("total %d, want clients * attempts = 150", r.Total)
	}
	sum := 0
	for _, n := range r.PerTick {
		sum += n
	}
	if sum != r.Total {
		t.Errorf("per tick sums to %d, total %d", sum, r.Total)
	}
	if _, err := Backoff(1, 1, 1, 1, "half", 1); !errors.Is(err, ErrInput) {
		t.Errorf("unknown jitter: %v", err)
	}
	if _, err := Backoff(1, 1, 5, 4, NoJitter, 1); !errors.Is(err, ErrInput) {
		t.Errorf("cap under base: %v", err)
	}
}
