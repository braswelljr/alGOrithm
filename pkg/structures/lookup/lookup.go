// Package lookup holds structures built to find something by its key: a
// HashTable with separate chaining, a hash Set, a Trie for prefixes, UnionFind
// for which group an element is in, and the caches that look up recent values
// and evict old ones: LRU, LFUCache and TTLCache.
package lookup
