package concurrency

import (
	"context"
	"errors"
	"sync"
)

// WorkerPool runs fn on every job with at most workers goroutines at a time
// and returns the results in the jobs' order. The first error cancels the
// context the remaining calls see and is returned; results of jobs that did
// not run are zero.
//
// Time: about O(len(jobs) / workers) calls deep. Space: O(len(jobs)).
func WorkerPool[T, R any](ctx context.Context, workers int, jobs []T, fn func(context.Context, T) (R, error)) ([]R, error) {
	if workers < 1 {
		workers = 1
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make([]R, len(jobs))
	next := make(chan int)
	var (
		wg      sync.WaitGroup
		errOnce sync.Once
		first   error
	)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				r, err := fn(ctx, jobs[i])
				if err != nil {
					errOnce.Do(func() { first = err; cancel() })
					continue
				}
				results[i] = r
			}
		}()
	}
feed:
	for i := range jobs {
		select {
		case next <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(next)
	wg.Wait()
	if first == nil && ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		first = ctx.Err()
	}
	return results, first
}

// Pipe starts workers goroutines that apply fn to each value from in and send
// the results on the returned channel, which is closed when in is drained or
// ctx is done. Results arrive in completion order, not input order.
func Pipe[T, R any](ctx context.Context, in <-chan T, workers int, fn func(T) R) <-chan R {
	if workers < 1 {
		workers = 1
	}
	out := make(chan R)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range in {
				select {
				case out <- fn(v):
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Semaphore lets at most n holders in at once: a counting semaphore on a
// buffered channel.
type Semaphore struct{ slots chan struct{} }

// NewSemaphore returns a semaphore with n slots (at least 1).
func NewSemaphore(n int) *Semaphore { return &Semaphore{slots: make(chan struct{}, max(1, n))} }

// Acquire takes a slot, waiting until one is free or ctx is done.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire takes a slot if one is free now, and reports whether it did.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release frees a slot taken by Acquire.
func (s *Semaphore) Release() { <-s.slots }

// ErrClosed is returned by BoundedBuffer.Put after Close.
var ErrClosed = errors.New("concurrency: buffer closed")

// BoundedBuffer is a fixed-size FIFO shared by producers and consumers: Put
// waits while it is full and Take while it is empty, each waking the other
// through a sync.Cond. It is the classic monitor solution to the
// producer-consumer problem.
type BoundedBuffer[T any] struct {
	mu       sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
	items    []T
	cap      int
	closed   bool
}

// NewBoundedBuffer returns a buffer holding at most capacity values (at least 1).
func NewBoundedBuffer[T any](capacity int) *BoundedBuffer[T] {
	b := &BoundedBuffer[T]{cap: max(1, capacity)}
	b.notFull = sync.NewCond(&b.mu)
	b.notEmpty = sync.NewCond(&b.mu)
	return b
}

// Put adds v, waiting while the buffer is full; ErrClosed after Close.
func (b *BoundedBuffer[T]) Put(v T) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for len(b.items) == b.cap && !b.closed {
		b.notFull.Wait()
	}
	if b.closed {
		return ErrClosed
	}
	b.items = append(b.items, v)
	b.notEmpty.Signal()
	return nil
}

// Take removes the oldest value, waiting while the buffer is empty. ok is
// false once the buffer is closed and drained.
func (b *BoundedBuffer[T]) Take() (v T, ok bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for len(b.items) == 0 && !b.closed {
		b.notEmpty.Wait()
	}
	if len(b.items) == 0 {
		return v, false
	}
	v = b.items[0]
	b.items = b.items[1:]
	b.notFull.Signal()
	return v, true
}

// Len is the number of values waiting.
func (b *BoundedBuffer[T]) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}

// Close wakes everyone: Put fails from now on, and Take drains what is left.
func (b *BoundedBuffer[T]) Close() {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
	b.notFull.Broadcast()
	b.notEmpty.Broadcast()
}
