// Package architecture simulates how a CPU runs programs, on a tiny
// educational instruction set with eight registers:
//
//	isa.go        the instructions and a two-pass assembler with labels
//	cpu.go        the fetch, decode, execute, write-back cycle
//	pipeline.go   a five-stage pipeline with data and control hazards,
//	              stalls, forwarding and flushes
//	predict.go    branch predictors: 1-bit, 2-bit saturating, gshare
//	perf.go       CPI, Amdahl's law, the memory hierarchy, von Neumann and
//	              Harvard buses
//	system.go     interrupts, context switches and the call stack
//	layout.go     struct alignment and padding
//
// Every simulation returns its steps as plain data, so they can be printed
// or animated. The caches themselves are simulated in pkg/os/cache.
package architecture
