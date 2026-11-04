package concurrency_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
)

func TestWorkerPool(t *testing.T) {
	jobs := make([]int, 100)
	for i := range jobs {
		jobs[i] = i
	}
	var running, peak int32
	out, err := concurrency.WorkerPool(context.Background(), 4, jobs, func(_ context.Context, n int) (int, error) {
		now := atomic.AddInt32(&running, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if now <= p || atomic.CompareAndSwapInt32(&peak, p, now) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		atomic.AddInt32(&running, -1)
		return n * n, nil
	})
	if err != nil || out[9] != 81 || out[99] != 9801 {
		t.Fatalf("results %v, %v", out[:10], err)
	}
	if peak > 4 {
		t.Errorf("%d workers ran at once, want at most 4", peak)
	}
	boom := errors.New("boom")
	if _, err := concurrency.WorkerPool(context.Background(), 2, jobs, func(_ context.Context, n int) (int, error) {
		if n == 5 {
			return 0, boom
		}
		return n, nil
	}); !errors.Is(err, boom) {
		t.Errorf("error %v, want boom", err)
	}
}

func TestPipe(t *testing.T) {
	in := make(chan int)
	go func() {
		for i := 1; i <= 50; i++ {
			in <- i
		}
		close(in)
	}()
	var got []int
	for v := range concurrency.Pipe(context.Background(), in, 3, func(n int) int { return n * 2 }) {
		got = append(got, v)
	}
	sort.Ints(got)
	if len(got) != 50 || got[0] != 2 || got[49] != 100 {
		t.Errorf("got %d values, %v...", len(got), got[:3])
	}
}

func TestSemaphoreAndBuffer(t *testing.T) {
	s := concurrency.NewSemaphore(2)
	first, second, third := s.TryAcquire(), s.TryAcquire(), s.TryAcquire()
	if !first || !second || third {
		t.Fatal("a semaphore of 2 admitted the wrong number")
	}
	s.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Acquire(ctx); err == nil {
		t.Error("Acquire on a full semaphore did not time out")
	}

	b := concurrency.NewBoundedBuffer[int](3)
	var wg sync.WaitGroup
	var sum int64
	for c := 0; c < 3; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				v, ok := b.Take()
				if !ok {
					return
				}
				atomic.AddInt64(&sum, int64(v))
			}
		}()
	}
	for i := 1; i <= 100; i++ {
		if err := b.Put(i); err != nil {
			t.Fatal(err)
		}
		if b.Len() > 3 {
			t.Fatal("buffer over capacity")
		}
	}
	b.Close()
	wg.Wait()
	if sum != 5050 {
		t.Errorf("consumed sum %d, want 5050", sum)
	}
	if err := b.Put(1); !errors.Is(err, concurrency.ErrClosed) {
		t.Error("Put after Close succeeded")
	}
}

func TestRace(t *testing.T) {
	lostSomewhere := false
	for seed := int64(1); seed <= 20; seed++ {
		r, err := concurrency.Race(4, 5, false, seed)
		if err != nil {
			t.Fatal(err)
		}
		lostSomewhere = lostSomewhere || r.Lost > 0
		safe, _ := concurrency.Race(4, 5, true, seed)
		if safe.Final != 20 || safe.Lost != 0 {
			t.Fatalf("seed %d: with the lock the counter is %d, want 20", seed, safe.Final)
		}
		again, _ := concurrency.Race(4, 5, false, seed)
		if again.Final != r.Final || len(again.Timeline) != len(r.Timeline) {
			t.Fatal("a simulation is not reproducible from its seed")
		}
	}
	if !lostSomewhere {
		t.Error("no seed lost an update without the lock; the race is not shown")
	}
}

func TestProducerConsumer(t *testing.T) {
	r, err := concurrency.ProducerConsumer(2, 3, 2, 30, 7)
	if err != nil {
		t.Fatal(err)
	}
	if r.Produced != 30 || r.Consumed != 30 || r.MaxBuffered > 2 {
		t.Errorf("produced %d consumed %d max buffered %d", r.Produced, r.Consumed, r.MaxBuffered)
	}
	waited := false
	for _, a := range r.Timeline {
		waited = waited || a.State == "blocked"
	}
	if !waited {
		t.Error("nobody ever waited on a buffer of 2")
	}
}

func TestDiningPhilosophers(t *testing.T) {
	deadlocked := false
	for seed := int64(1); seed <= 200 && !deadlocked; seed++ {
		r, _ := concurrency.DiningPhilosophers(3, 3, false, seed)
		deadlocked = r.Deadlocked
	}
	if !deadlocked {
		t.Error("the naive table never deadlocked in 200 seeds")
	}
	for seed := int64(1); seed <= 100; seed++ {
		r, _ := concurrency.DiningPhilosophers(5, 3, true, seed)
		if r.Deadlocked {
			t.Fatalf("seed %d: the ordered table deadlocked", seed)
		}
		for i, m := range r.Meals {
			if m != 3 {
				t.Fatalf("seed %d: philosopher %d ate %d meals", seed, i, m)
			}
		}
	}
}
