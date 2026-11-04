// Package virtualmemory goes beyond single-level paging (in the memory
// package) to how real systems translate and share memory, one file per
// idea:
//
//	pagewalk.go  multi-level page tables: splitting an address and walking it
//	tlb.go       the TLB, its hit rate, and the effective access time
//	inverted.go  an inverted page table: one entry per frame, found by hash
//	workingset.go the working set, and thrashing when it does not fit
//	cow.go       copy-on-write after fork
//	mmap.go      memory-mapped files against read()
package virtualmemory

import (
	"errors"
	"fmt"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("virtualmemory: invalid input")

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}
