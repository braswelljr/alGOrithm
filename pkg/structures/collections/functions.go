package collections

import (
	"sort"
	"sync"
	"time"
)

// Memoize returns f remembering each result by its argument, so a repeated
// call returns the stored result (lodash _.memoize). It is safe for
// concurrent use; the cache grows without bound, as lodash's does.
//
// Time: O(1) per remembered call.
func Memoize[K comparable, V any](f func(K) V) func(K) V {
	return MemoizeBy(f, func(k K) K { return k })
}

// MemoizeBy is Memoize keyed by key(arg) (lodash's resolver argument).
func MemoizeBy[A any, K comparable, V any](f func(A) V, key func(A) K) func(A) V {
	var mu sync.Mutex
	cache := map[K]V{}
	return func(a A) V {
		k := key(a)
		mu.Lock()
		v, ok := cache[k]
		mu.Unlock()
		if ok {
			return v
		}
		v = f(a)
		mu.Lock()
		cache[k] = v
		mu.Unlock()
		return v
	}
}

// Once returns a function that calls f the first time and returns that
// result every time after (lodash _.once), safe for concurrent use.
func Once[T any](f func() T) func() T {
	var (
		once sync.Once
		v    T
	)
	return func() T {
		once.Do(func() { v = f() })
		return v
	}
}

// Curry2 turns f(a, b) into f(a)(b) (lodash _.curry for two arguments).
func Curry2[A, B, R any](f func(A, B) R) func(A) func(B) R {
	return func(a A) func(B) R { return func(b B) R { return f(a, b) } }
}

// Curry3 turns f(a, b, c) into f(a)(b)(c).
func Curry3[A, B, C, R any](f func(A, B, C) R) func(A) func(B) func(C) R {
	return func(a A) func(B) func(C) R {
		return func(b B) func(C) R { return func(c C) R { return f(a, b, c) } }
	}
}

// Pipe returns a function applying fs left to right (lodash _.flow).
func Pipe[T any](fs ...func(T) T) func(T) T {
	return func(v T) T {
		for _, f := range fs {
			v = f(v)
		}
		return v
	}
}

// Compose returns a function applying fs right to left (lodash _.flowRight).
func Compose[T any](fs ...func(T) T) func(T) T {
	return func(v T) T {
		for i := len(fs) - 1; i >= 0; i-- {
			v = fs[i](v)
		}
		return v
	}
}

// Clock is the time Debounce and Throttle use: the real clock by default, a
// ManualClock in tests and demos.
type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) Timer
}

// Timer is a pending call a Clock will make.
type Timer interface{ Stop() bool }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (realClock) AfterFunc(d time.Duration, f func()) Timer { return time.AfterFunc(d, f) } //nolint:ireturn // Clock returns a Timer: real and manual clocks have different ones

// RealClock is the wall clock.
var RealClock Clock = realClock{}

// ManualClock is a Clock that only moves when Advance is called, running due
// timers in order: deterministic time for tests and step-by-step demos.
type ManualClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*manualTimer
	seq    int
}

type manualTimer struct {
	at      time.Time
	f       func()
	seq     int
	stopped bool
}

func (t *manualTimer) Stop() bool {
	was := !t.stopped
	t.stopped = true
	return was
}

// NewManualClock returns a clock reading start.
func NewManualClock(start time.Time) *ManualClock { return &ManualClock{now: start} }

// Now returns the clock's time.
func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// AfterFunc schedules f for d after the clock's time.
func (c *ManualClock) AfterFunc(d time.Duration, f func()) Timer { //nolint:ireturn // Clock returns a Timer: real and manual clocks have different ones
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	t := &manualTimer{at: c.now.Add(d), f: f, seq: c.seq}
	c.timers = append(c.timers, t)
	return t
}

// Advance moves the clock forward by d, running every timer that falls due,
// earliest first, each with the clock set to its time.
func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	end := c.now.Add(d)
	c.mu.Unlock()
	for {
		c.mu.Lock()
		sort.SliceStable(c.timers, func(i, j int) bool {
			if !c.timers[i].at.Equal(c.timers[j].at) {
				return c.timers[i].at.Before(c.timers[j].at)
			}
			return c.timers[i].seq < c.timers[j].seq
		})
		var due *manualTimer
		for len(c.timers) > 0 && c.timers[0].stopped {
			c.timers = c.timers[1:]
		}
		if len(c.timers) > 0 && !c.timers[0].at.After(end) {
			due, c.timers = c.timers[0], c.timers[1:]
			c.now = due.at
		}
		if due == nil {
			c.now = end
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()
		due.f()
	}
}

// DebounceOptions choose when a debounced call runs.
type DebounceOptions struct {
	// Leading runs f at the start of a burst of calls.
	Leading bool
	// Trailing runs f at the end of the burst, wait after the last call.
	Trailing bool
	// Clock is the time to use; nil is RealClock.
	Clock Clock
}

// Debounce returns a function that delays calling f until wait has passed
// since the last call (lodash _.debounce): a burst of calls runs f once.
// Leading runs it at the start of the burst instead or as well; the zero
// options are lodash's default, trailing only. cancel drops a pending call.
// Safe for concurrent use.
func Debounce(f func(), wait time.Duration, opts DebounceOptions) (call, cancel func()) {
	if !opts.Leading && !opts.Trailing {
		opts.Trailing = true
	}
	clk := opts.Clock
	if clk == nil {
		clk = RealClock
	}
	var (
		mu      sync.Mutex
		timer   Timer
		pending bool
	)
	expire := func() {
		mu.Lock()
		run := pending && opts.Trailing
		pending, timer = false, nil
		mu.Unlock()
		if run {
			f()
		}
	}
	call = func() {
		mu.Lock()
		leading := timer == nil && opts.Leading
		if timer != nil {
			timer.Stop()
		}
		pending = !leading
		timer = clk.AfterFunc(wait, expire)
		mu.Unlock()
		if leading {
			f()
		}
	}
	cancel = func() {
		mu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timer, pending = nil, false
		mu.Unlock()
	}
	return call, cancel
}

// Throttle returns a function that calls f at most once every wait (lodash
// _.throttle): the first call of a window runs f at once when Leading, and a
// call during the window runs it at the window's end when Trailing. The zero
// options are lodash's default, leading and trailing. Safe for concurrent
// use.
func Throttle(f func(), wait time.Duration, opts DebounceOptions) (call, cancel func()) {
	if !opts.Leading && !opts.Trailing {
		opts.Leading, opts.Trailing = true, true
	}
	clk := opts.Clock
	if clk == nil {
		clk = RealClock
	}
	var (
		mu      sync.Mutex
		timer   Timer
		pending bool
	)
	var endWindow func()
	endWindow = func() {
		mu.Lock()
		run := pending && opts.Trailing
		pending = false
		if run {
			timer = clk.AfterFunc(wait, endWindow) // the trailing call opens a new window
		} else {
			timer = nil
		}
		mu.Unlock()
		if run {
			f()
		}
	}
	call = func() {
		mu.Lock()
		if timer != nil {
			pending = true
			mu.Unlock()
			return
		}
		timer = clk.AfterFunc(wait, endWindow)
		run := opts.Leading
		pending = !run
		mu.Unlock()
		if run {
			f()
		}
	}
	cancel = func() {
		mu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timer, pending = nil, false
		mu.Unlock()
	}
	return call, cancel
}
