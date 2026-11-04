// Package circuits builds digital circuits out of logic gates and simulates
// them, one idea per file.
//
// circuit.go is the model: a Circuit is gates wired by name, and Simulate
// propagates the inputs with one unit of delay per gate in three-valued logic
// (0, 1, and X for a wire nothing has driven yet), so the frames show a
// signal moving through the circuit layer by layer, and a circuit's depth is
// how long it takes to settle. Feedback is allowed, which is how latches
// remember.
//
// combinational.go builds, from gates: everything from NAND and from NOR,
// half and full adders, ripple-carry and carry-lookahead adders, a
// subtractor, a multiplexer and demultiplexer, a decoder, a priority encoder,
// a magnitude comparator, the SR latch and the gated D latch.
//
// sequential.go models clocked parts by behavior over clock cycles, as timing
// diagrams: D, JK and T flip-flops, a register, a shift register, ripple and
// synchronous counters, a traffic light state machine, plus an ALU and a
// seven-segment decoder.
package circuits
