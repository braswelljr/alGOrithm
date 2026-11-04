// Package hashing holds hash functions and the ways a hash table resolves
// collisions, one file each.
//
// The functions are FNV-1a (32 and 64 bit), DJB2, SDBM, Murmur3 (32 bit),
// CRC-32 (IEEE, table-driven), Adler-32 and a polynomial rolling hash, written
// out so each step can be read; Distribution measures how evenly a function
// spreads keys over buckets. The tables are open addressing with linear,
// quadratic and double-hashing probes (OpenTable), cuckoo hashing, Robin Hood
// hashing, and a consistent-hashing ring. Every table operation returns the
// slots it probed, which is what the visualizer draws.
//
// None of these hashes is cryptographic: use crypto/sha256 to protect data.
//
//	hashing.FNV1a32([]byte("hello")) // 0x4f9f2cab
package hashing
