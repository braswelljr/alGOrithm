// Package databases shows how a database works inside, one idea per file,
// small enough to read top to bottom:
//
//	storage       slotted pages, a buffer pool (clock and LRU-K), a
//	              write-ahead log with crash recovery
//	indexes       a B+ tree, an LSM tree with a Bloom filter per run,
//	              extendible hashing, linear hashing
//	queries       nested-loop, block nested-loop, index nested-loop, hash and
//	              sort-merge joins; external merge sort; relational algebra
//	              (select, project, group by, sort, limit); a tiny planner
//	              choosing an index scan or a full scan by estimated cost
//	transactions  strict two-phase locking with deadlock detection,
//	              conflict serializability, timestamp ordering, MVCC snapshot
//	              isolation, and the anomalies each isolation level allows
//
// Everything runs in memory on small inputs, and each file says what it
// simplifies: there are no disks, no concurrency and no SQL parser. What is
// kept is the algorithm a real database runs, so its steps can be traced and
// its costs counted.
package databases

import "errors"

// ErrInput is wrapped by every error about malformed input: a table, a
// schedule, a query or an operation that cannot be read.
var ErrInput = errors.New("databases: invalid input")
