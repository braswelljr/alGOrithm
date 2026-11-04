package core

// Path is a learning path: an ordered walk through entries that builds one
// idea on the last, with a note on what to notice at each step. Paths turn the
// catalog into a curriculum; they name entries by ID and never copy them.
type Path struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Level    string `json:"level"`    // Beginner, Intermediate or Advanced
	Audience string `json:"audience"` // who it is for, in a few words
	// Prerequisites are the IDs of paths worth walking first.
	Prerequisites []string `json:"prerequisites,omitempty"`
	// Guide is a repository-relative link to a longer read, if there is one.
	Guide string     `json:"guide,omitempty"`
	Steps []PathStep `json:"steps"`
}

// PathStep is one entry to run, what to watch for while it runs, and the input
// that shows it best; a nil Input means the entry's own Example.
type PathStep struct {
	Entry string `json:"entry"`
	Note  string `json:"note"`
	Input *Input `json:"input,omitempty"`
}

// step is a PathStep on the entry's own example.
func step(id, note string) PathStep { return PathStep{Entry: id, Note: note} }

// stepOn is a PathStep on a chosen input.
func stepOn(id, note string, in Input) PathStep { return PathStep{Entry: id, Note: note, Input: &in} }

// reversed12 and nearlySorted are the inputs the sorting steps compare.
var (
	reversed12   = Input{Values: []int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}}
	nearlySorted = Input{Values: []int{1, 2, 3, 5, 4, 6, 7, 8, 10, 9, 11, 12}}
	belady       = []int{1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5}
)

// PathByID returns the path called id.
func PathByID(id string) (Path, bool) {
	for _, p := range Paths() {
		if p.ID == id {
			return p, true
		}
	}
	return Path{}, false
}

// Paths returns every learning path, easiest first. It is the only list of
// paths; the command, the site and docs/learn.md all read it.
func Paths() []Path {
	return []Path{
		{
			ID: "big-o-by-example", Title: "Big O by example", Level: Beginner, Audience: "anyone new to algorithms",
			Summary: "What O(1), O(log n), O(n), O(n log n), O(n^2) and O(2^n) feel like, seen in the counts of real runs.",
			Guide:   "docs/guides/big-o.md",
			Steps: []PathStep{
				step("hash-table", "A lookup costs about the same however many keys there are: O(1) on average."),
				step("binary-search", "Count the comparisons: halving the range each time means at most about log2(n) of them: 3 to find 51 among 14 values."),
				step("linear-search", "The same data, one value at a time: the comparisons grow with n, O(n)."),
				stepOn("merge-sort", "Halving and merging: about n log2(n) comparisons, the best a comparison sort can do.", reversed12),
				stepOn("bubble-sort", "Every pair compared: on 12 reversed values that is 66 comparisons and 66 swaps, O(n^2).", reversed12),
				stepOn("fibonacci", "Plain recursion recomputes the same values: the calls grow like 1.6^n, so 25 is already slow.", Input{Values: []int{25}}),
				stepOn("fibonacci-memo", "Remember each answer once and the same question takes n steps: O(n) instead of O(1.6^n).", Input{Values: []int{25}}),
				step("tower-of-hanoi", "Each extra disk doubles the moves: 2^n - 1, the cost of a truly exponential problem."),
			},
		},
		{
			ID: "sorting-from-zero", Title: "Sorting from zero", Level: Beginner, Audience: "first algorithms course",
			Summary:       "From the simplest sorts to the ones in standard libraries, and why input order changes the cost.",
			Prerequisites: []string{"big-o-by-example"},
			Steps: []PathStep{
				stepOn("bubble-sort", "Watch how many swaps bubble sort makes on reversed input: every pair is out of order.", reversed12),
				stepOn("selection-sort", "The same 66 comparisons as bubble sort, but at most one swap per pass: only 6 here.", reversed12),
				stepOn("insertion-sort", "On nearly sorted input almost nothing moves: insertion sort is close to O(n) here.", nearlySorted),
				step("shell-sort", "Insertion sort on far-apart elements first, so the last pass has little left to do."),
				step("merge-sort", "Split, sort the halves, merge: O(n log n) whatever the input, at the cost of extra space."),
				step("quick-sort", "Partition around a pivot: fast on average, and the swaps happen in place."),
				step("heap-sort", "Build a heap, then pop the largest again and again: O(n log n) with O(1) extra space."),
				step("counting-sort", "No comparisons at all: count each value, then write them out in order."),
				step("radix-sort", "Sort by one digit at a time with a stable pass: linear in n for fixed-width keys."),
				step("tim-sort", "What Python and Java use: find the runs already sorted and merge them."),
			},
		},
		{
			ID: "searching-and-sorted-data", Title: "Searching, and why sorted data matters", Level: Beginner, Audience: "first algorithms course",
			Summary:       "What sorting buys you: from scanning every value to finding one in a handful of steps.",
			Prerequisites: []string{"sorting-from-zero"},
			Steps: []PathStep{
				step("linear-search", "Works on any data, and looks at every value in the worst case."),
				step("binary-search", "Needs sorted data, and halves the range with each comparison."),
				step("lower-bound", "Binary search for the first value not less than the target: where it would be inserted."),
				step("jump-search", "Jump ahead sqrt(n) at a time, then step back: between linear and binary."),
				step("interpolation-search", "Guess where the target should be from its value: very fast on evenly spread data."),
				step("exponential-search", "Double the range until it passes the target: useful when the end is unknown."),
				step("quickselect", "Find the k-th smallest without sorting everything: O(n) on average."),
				step("median-of-medians", "The same answer with a guaranteed O(n): a careful pivot choice."),
			},
		},
		{
			ID: "arrays-to-trees", Title: "Data structures: from arrays to trees", Level: Beginner, Audience: "first data structures course",
			Summary:       "How the way data is held decides which operations are fast.",
			Prerequisites: []string{"big-o-by-example"},
			Steps: []PathStep{
				step("dynamic-array", "Appends are cheap until the array is full; then it doubles, so they stay O(1) on average."),
				step("linked-list", "Inserting at the front is O(1), but finding the k-th value means walking k links."),
				step("stack", "Last in, first out: the shape of function calls and undo."),
				step("queue", "First in, first out: the shape of waiting lines and breadth-first search."),
				step("access-orders", "The same values through a stack, a queue and a heap: three orders back."),
				step("hash-table", "Keys spread over buckets: lookups in O(1) on average."),
				step("binary-search-tree", "Smaller to the left, larger to the right: fast while the tree stays bushy."),
				step("avl-tree", "Rotations keep the tree balanced, so every operation stays O(log n)."),
				step("min-heap", "Only the smallest value is easy to find, and that is all a priority queue needs."),
				step("trie", "A tree of characters: every word with a prefix is below one node."),
				step("union-find", "Which group is each element in? Nearly O(1) per question with path compression."),
			},
		},
		{
			ID: "javascript-and-lodash", Title: "JavaScript arrays and lodash, inside", Level: Beginner, Audience: "web developers",
			Summary: "The array methods and utility functions you use every day, implemented and traced.",
			Steps: []PathStep{
				step("js-array", "push, pop, shift and unshift: why shift and unshift move every element."),
				step("js-map", "A new array of the same length, one call per element."),
				step("js-filter", "A new array of only the elements that pass."),
				step("js-reduce", "Everything above can be written as a reduce."),
				step("js-splice", "Removing from the middle shifts everything after it: O(n)."),
				step("js-sort", "The comparator decides the order, and sort changes the array in place."),
				step("deep-clone", "A copy with no shared nested objects, so changing one never changes the other."),
				step("deep-merge", "Nested objects merged key by key, with the later value winning."),
				step("group-by", "One pass, one bucket per key."),
				step("memoize-demo", "Cache results by argument: the second call with the same input is free."),
				step("debounce-demo", "Only the last call in a burst runs."),
				step("throttle-demo", "At most one call per interval, however many arrive."),
			},
		},
		{
			ID: "recursion-and-dynamic-programming", Title: "Thinking recursively, then dynamic programming", Level: Intermediate, Audience: "second algorithms course",
			Summary:       "Solve a problem in terms of smaller ones, then stop solving the same smaller ones twice.",
			Prerequisites: []string{"big-o-by-example"},
			Steps: []PathStep{
				step("factorial", "The smallest recursion: a base case and one smaller call."),
				step("tower-of-hanoi", "Trust the recursion: move n - 1 disks, move one, move n - 1 again."),
				step("permutations", "Choose each first element, then permute the rest."),
				step("n-queens", "Recursion that undoes a choice when it leads nowhere: backtracking."),
				stepOn("fibonacci", "Overlapping subproblems: the same calls again and again.", Input{Values: []int{20}}),
				stepOn("fibonacci-memo", "Memoization: the same recursion, each answer computed once.", Input{Values: []int{20}}),
				step("coin-change", "The table fills from small amounts up: each cell uses ones already filled."),
				step("longest-common-subsequence", "A two-dimensional table: each cell looks left, up and diagonally."),
				step("edit-distance", "The same table shape, counting insertions, deletions and substitutions."),
				step("knapsack-01", "Take an item or leave it: the table remembers the best of both."),
			},
		},
		{
			ID: "greedy-versus-optimal", Title: "Greedy choices, and when they fail", Level: Intermediate, Audience: "second algorithms course",
			Summary:       "Taking the best-looking choice is fast, and sometimes exactly right; these pairs show when.",
			Prerequisites: []string{"recursion-and-dynamic-programming"},
			Steps: []PathStep{
				step("greedy-coin-change", "Coins 1, 3 and 4 for 6: greedy takes 4 + 1 + 1, three coins."),
				stepOn("coin-change", "Dynamic programming finds 3 + 3, two coins: greedy was not optimal.", Input{Values: []int{1, 3, 4}, Target: 6}),
				step("activity-selection", "Here greedy is optimal: always take the meeting that ends first."),
				step("fractional-knapsack", "When items can be cut, the best value per weight first is optimal."),
				step("knapsack-01", "When they cannot, greedy can be beaten: this needs the table."),
				step("huffman-coding", "Merge the two rarest symbols first: greedy, and it gives the shortest code."),
				step("tsp-nearest-neighbor", "Always visit the closest city next: quick, and often a poor tour."),
				step("tsp-held-karp", "The exact tour, in O(2^n n^2): fine for ten cities, hopeless for a hundred."),
			},
		},
		{
			ID: "pathfinding-on-grids", Title: "Pathfinding on grids", Level: Beginner, Audience: "game and robotics beginners",
			Summary:       "How a search decides which cell to explore next, and why that changes everything.",
			Prerequisites: []string{"arrays-to-trees"},
			Steps: []PathStep{
				step("breadth-first-search", "A queue explores in rings: the first path to the end is a shortest one."),
				step("depth-first-search", "A stack dives deep: it finds a path, rarely the shortest."),
				step("dijkstra", "A priority queue by cost so far: shortest paths when cells cost different amounts."),
				step("greedy-best-first", "By distance to the goal only: fast, and fooled by walls."),
				step("a-star", "Cost so far plus distance to go: as fast as greedy, as right as Dijkstra."),
			},
		},
		{
			ID: "graphs-step-by-step", Title: "Graphs step by step", Level: Intermediate, Audience: "second algorithms course",
			Summary:       "Traversals, orders, shortest paths, spanning trees, components and flows on weighted graphs.",
			Prerequisites: []string{"pathfinding-on-grids"},
			Steps: []PathStep{
				step("graph-bfs", "Visit by distance in edges: the frontier is a queue."),
				step("graph-dfs", "Visit as deep as possible first: the frontier is a stack."),
				step("topological-sort", "An order where every edge points forward: only a graph without cycles has one."),
				step("graph-dijkstra", "Shortest paths with weights, as long as none is negative."),
				step("bellman-ford", "Negative weights allowed: relax every edge n - 1 times, and detect negative cycles."),
				step("floyd-warshall", "Every pair at once, in a table: O(n^3)."),
				step("kruskal", "Cheapest edges first, skipping any that would close a cycle: a minimum spanning tree."),
				step("prim", "The same tree grown from one node, always by its cheapest edge out."),
				step("tarjan-scc", "Groups where every node reaches every other, in one depth-first pass."),
				step("max-flow", "Push flow along paths with room left until none is left."),
				step("pagerank", "Importance as the chance a random walk is at a node."),
			},
		},
		{
			ID: "text-and-patterns", Title: "Searching text", Level: Intermediate, Audience: "anyone who has used grep",
			Summary:       "Finding patterns in text without reading it twice, and the structures that make it fast.",
			Prerequisites: []string{"searching-and-sorted-data"},
			Steps: []PathStep{
				step("naive-search", "Try every position: O(n m) when the text is full of near matches."),
				step("knuth-morris-pratt", "Never move back in the text: the pattern's own structure says how far to jump."),
				step("z-algorithm", "For each position, how long a match with the start: another linear-time view."),
				step("boyer-moore-horspool", "Compare from the end of the pattern and skip ahead on a mismatch."),
				step("rabin-karp", "Compare hashes of windows instead of the windows themselves."),
				step("aho-corasick", "Many patterns at once in a single pass over the text."),
				step("suffix-array", "Every suffix, sorted: any substring search becomes a binary search."),
				step("longest-repeated-substring", "With the suffixes sorted, repeats sit side by side: compare neighbors only."),
				step("longest-common-substring", "The same trick across two texts."),
				step("manacher", "Every longest palindrome around every center, in linear time."),
			},
		},
		{
			ID: "words-and-documents", Title: "Words and documents", Level: Intermediate, Audience: "anyone building search or text tools",
			Summary:       "From counting words to ranking documents, suggesting spellings and showing what changed.",
			Prerequisites: []string{"text-and-patterns"},
			Steps: []PathStep{
				step("word-frequency", "The first thing any text tool does: count the words."),
				step("n-grams", "Runs of n words or characters: the context a single word lacks."),
				step("tf-idf", "A word matters in a document if it is frequent there and rare elsewhere."),
				step("soundex", "Names that sound alike get the same code, however they are spelled."),
				step("fuzzy-search", "Find matches that are close, not exact: edit distance at work."),
				step("spelling-suggestions", "The nearest dictionary words to a misspelling."),
				step("myers-diff", "The fewest insertions and deletions between two texts: what diff and git show."),
				step("text-justification", "Break lines and spread spaces so both edges line up."),
			},
		},
		{
			ID: "making-data-smaller", Title: "Making data smaller", Level: Intermediate, Audience: "anyone curious about zip files",
			Summary:       "Lossless compression: repeated runs, frequent symbols, repeated phrases, and transforms that help.",
			Prerequisites: []string{"greedy-versus-optimal"},
			Steps: []PathStep{
				step("run-length-encoding", "Runs of the same symbol become a count: great for some data, worse for most."),
				step("huffman-coding", "Frequent symbols get short codes."),
				step("lz77", "Repeated phrases become a pointer back to where they appeared."),
				step("lzw", "A dictionary of phrases built as the text is read."),
				step("burrows-wheeler", "A reversible shuffle that groups similar characters together."),
				step("move-to-front", "After that shuffle, recent symbols become small numbers that compress well."),
			},
		},
		{
			ID: "numbers-and-math", Title: "Numbers and the math under algorithms", Level: Intermediate, Audience: "learners who like math",
			Summary:       "Divisibility, primes, modular arithmetic and fast arithmetic: the toolkit behind cryptography.",
			Prerequisites: []string{"big-o-by-example"},
			Steps: []PathStep{
				step("gcd", "Euclid's algorithm: the remainder shrinks quickly, about log steps."),
				step("extended-gcd", "The same steps also give the coefficients of the gcd."),
				step("sieve-of-eratosthenes", "Every prime up to n by crossing out multiples."),
				step("fast-power", "Square and multiply: x^n in log2(n) multiplications."),
				step("mod-inverse", "Division in modular arithmetic, from the extended gcd."),
				step("miller-rabin", "Is a huge number prime? A few random witnesses make an error very unlikely."),
				step("chinese-remainder", "One number from its remainders by several moduli."),
				step("karatsuba", "Multiply big numbers with three half-size products instead of four."),
				step("newton-raphson", "Follow the tangent to a root: the correct digits roughly double each step."),
				step("fft", "Multiply polynomials in O(n log n) by working with their values instead."),
			},
		},
		{
			ID: "classical-ciphers", Title: "Classical ciphers and how they break", Level: Beginner, Audience: "puzzle lovers and security beginners",
			Summary: "Two thousand years of secret writing, and the attacks that make each one obsolete.",
			Steps: []PathStep{
				step("caesar-cipher", "Shift every letter by the same amount."),
				step("caesar-brute-force", "Only 25 keys: try them all and read the one that makes sense."),
				step("substitution-cipher", "A shuffled alphabet: too many keys to try."),
				step("frequency-analysis", "But letter frequencies survive, and give the key away."),
				step("vigenere-cipher", "Several shifts in turn hide the frequencies."),
				step("vigenere-key-length", "Repeated fragments reveal the length of the key."),
				step("vigenere-crack", "Then each position is a Caesar cipher again."),
				step("one-time-pad", "A random key as long as the message: unbreakable, if never reused."),
				step("enigma", "Rotors that turn with every letter: a machine that changes its cipher each keypress."),
			},
		},
		{
			ID: "crypto-internals", Title: "Crypto internals", Level: Advanced, Audience: "developers who use cryptography libraries",
			Summary:       "What happens inside hashes, ciphers and key exchange. For learning only: never protect data with these.",
			Prerequisites: []string{"classical-ciphers", "numbers-and-math"},
			Steps: []PathStep{
				step("base64", "Encoding is not encryption: anyone can reverse it."),
				step("sha-256", "A fixed-size fingerprint: change one bit of input and half the output bits change."),
				step("hmac-sha256", "A hash with a key: proves who wrote the message."),
				step("pbkdf2", "A hash run thousands of times, so guessing passwords is slow."),
				step("constant-time-compare", "Compare secrets without leaking how many bytes matched."),
				step("feistel-network", "The round structure behind DES: encryption and decryption share the code."),
				step("aes-128", "Substitute, shift, mix and add the key, ten rounds over."),
				step("chacha20", "A stream cipher built only from add, rotate and xor."),
				step("diffie-hellman", "Two parties agree on a secret over a public channel."),
				step("rsa", "Easy to multiply two primes, hard to factor the product."),
				step("elliptic-curve", "The same idea on a curve, with much smaller keys."),
				step("merkle-tree", "Hashes of hashes: prove one block is in a large set with a few hashes."),
				step("shamir-secret-sharing", "Split a secret so any k of n shares recover it, and fewer reveal nothing."),
			},
		},
		{
			ID: "hashing-inside-out", Title: "Hashing inside out", Level: Intermediate, Audience: "developers who use maps every day",
			Summary:       "Hash functions, and what a hash table does when two keys land in the same place.",
			Prerequisites: []string{"arrays-to-trees"},
			Steps: []PathStep{
				step("djb2", "A classic string hash: multiply by 33 and add each character."),
				step("fnv1a", "Xor then multiply: simple, fast and well spread."),
				step("hash-distribution", "How evenly a hash spreads keys over buckets."),
				step("linear-probing", "On a collision, take the next free slot: fast until clusters form."),
				step("quadratic-probing", "Jump further each time to break up clusters."),
				step("double-hashing", "A second hash decides the jump."),
				step("robin-hood-hashing", "Rich keys give their slot to poor ones: probe lengths even out."),
				step("cuckoo-hashing", "Two places per key, and kicking out whoever is there: O(1) worst-case lookups."),
				step("bloom-filter", "A few bits per key: never a false no, sometimes a false yes."),
				step("consistent-hashing", "Add a server and only a few keys move."),
			},
		},
		{
			ID: "machine-learning-from-scratch", Title: "Machine learning from scratch", Level: Intermediate, Audience: "developers curious about ML",
			Summary:       "Fit a line, then classify, cluster and train a tiny neural network, with every number visible.",
			Prerequisites: []string{"numbers-and-math"},
			Steps: []PathStep{
				step("linear-regression", "The line that makes the squared errors smallest, solved in one formula."),
				step("gradient-descent", "The same line found by stepping downhill on the error: how most models learn."),
				step("polynomial-regression", "More terms fit the data better, and past a point fit the noise too."),
				step("regression-metrics", "How wrong is the model? Mean squared error, mean absolute error and R squared."),
				step("train-test-split", "Judge a model on data it has not seen, or the score means little."),
				step("k-nearest-neighbors", "No training at all: a new point takes the class of its nearest neighbors."),
				step("logistic-regression", "A line that outputs a probability, for yes or no questions."),
				step("confusion-matrix", "Which classes the model mixes up, not just how often it is right."),
				step("precision-recall-f1", "When one class is rare, accuracy hides the misses; precision and recall do not."),
				step("decision-tree", "A sequence of yes or no questions, each chosen to split the classes best."),
				step("naive-bayes-text", "Spam or not, by multiplying how likely each word is in each class."),
				step("k-means", "Move each center to the mean of its points until nothing changes."),
				step("k-means-plus-plus", "Spread the first centers out, and k-means settles faster and better."),
				step("silhouette-score", "How well each point sits in its cluster: a way to choose k."),
				step("dbscan", "Clusters as dense regions, of any shape, with noise left out."),
				step("perceptron", "One neuron: a weighted sum and a threshold, nudged after every mistake."),
				step("activation-functions", "The bends that let a network be more than one big linear function."),
				step("xor-network", "XOR is not linearly separable: one hidden layer and backpropagation learn it."),
			},
		},
		{
			ID: "how-an-os-works", Title: "How an operating system works", Level: Intermediate, Audience: "operating systems course",
			Summary:       "Processes and the CPU, then memory, disks, deadlock and threads: the decisions an OS makes every millisecond.",
			Prerequisites: []string{"arrays-to-trees"},
			Steps: []PathStep{
				step("fcfs-scheduling", "Processes run in arrival order: one long job makes everyone behind it wait."),
				step("sjf-scheduling", "Shortest job first gives the lowest average waiting time, if the lengths are known."),
				stepOn("round-robin", "Each process gets a slice of time in turn: try quantum 1 and quantum 10 and compare the waiting times.", Input{Text: "name arrival burst priority\nP1 0 5 1\nP2 1 3 2\nP3 2 8 1\nP4 3 6 3", Options: map[string]int{"quantum": 1}}),
				step("mlfq-scheduling", "Jobs that use their whole slice drop to a lower queue: interactive ones stay responsive."),
				step("first-fit", "Place a process in the first hole big enough: holes appear between allocations."),
				step("address-translation", "A virtual address becomes a page number and an offset, then a physical frame."),
				stepOn("fifo-page-replacement", "Evict the oldest page: with 3 frames this string faults 9 times.", Input{Values: belady, Options: map[string]int{"frames": 3}}),
				stepOn("fifo-page-replacement", "Belady's anomaly: with 4 frames the same string faults 10 times, more memory and more faults.", Input{Values: belady, Options: map[string]int{"frames": 4}}),
				step("lru-page-replacement", "Evict the page used least recently: no anomaly."),
				step("optimal-page-replacement", "Evict the page used furthest in the future: the best possible, and impossible in practice."),
				step("fcfs-disk", "Serve disk requests in order: the head zigzags across the platter."),
				step("scan-disk", "Sweep one way, then back, like an elevator: far less head movement."),
				step("bankers-algorithm", "Grant a request only if every process can still finish: no deadlock."),
				step("race-condition", "Two threads add to a counter without a lock: some updates are lost."),
				step("producer-consumer", "A bounded buffer: producers wait when it is full, consumers when it is empty."),
				step("dining-philosophers", "Everyone takes their left fork and waits for the right: deadlock, then the fix."),
				step("cache-simulator", "Addresses map to cache lines: watch hits and misses as the program walks memory."),
			},
		},
		{
			ID: "how-systems-talk-and-agree", Title: "How systems talk and agree", Level: Intermediate, Audience: "backend developers",
			Summary:       "From one lossy link to a cluster that agrees: retransmission, congestion, clocks, quorums, consensus, and the infrastructure that keeps it up.",
			Prerequisites: []string{"how-an-os-works"},
			Steps: []PathStep{
				step("stop-and-wait", "Send one frame, wait for its ack, resend on a timeout: reliable, but one frame per round trip."),
				step("go-back-n", "A window of frames in flight: faster, but one loss resends everything after it."),
				step("tcp-congestion-control", "Slow start doubles the window, a loss halves it: TCP's sawtooth."),
				step("dns-resolution", "A name walks from the root to the authoritative server, then sits in caches."),
				step("lamport-clock", "No shared clock: order events by counters that every message carries."),
				step("vector-clock", "One counter per node tells concurrent events apart from ordered ones."),
				step("quorum-consistency", "With R + W > N every read meets the latest write."),
				step("two-phase-commit", "Everyone votes, then everyone commits; one no aborts all, and a crashed coordinator blocks them."),
				step("raft", "Elect a leader, replicate its log to a majority: at most one leader per term."),
				step("round-robin-load-balancer", "Spread requests across servers; then compare least connections."),
				step("circuit-breaker", "Stop calling a failing dependency, then test it again after a timeout."),
			},
		},
		{
			ID: "system-design", Title: "System design from estimates to architectures", Level: Intermediate, Audience: "engineers preparing a design or an interview",
			Summary:       "Size a system, give it IDs, keep it up, then design real ones: a URL shortener, a key-value store, a feed, payments, and the data patterns behind them.",
			Prerequisites: []string{"how-systems-talk-and-agree"},
			Guide:         "docs/guides/system-design.md",
			Steps: []PathStep{
				step("back-of-the-envelope", "Start with numbers: 10 million users reading 20 times a day is about 2,300 reads a second, peaking near 7,000."),
				step("latency-numbers", "Memory is 100 times a cache hit, a datacenter round trip 5,000 times: keep the hot path in memory."),
				step("availability-nines", "Each dependency in series costs availability; each replica in parallel adds a nine."),
				step("snowflake-id", "IDs that sort by time with no coordination, and what a clock going back does."),
				step("url-shortener", "A counter or a hash for codes, and a cache that serves most redirects."),
				step("leader-follower-replication", "Followers lag; a user can miss their own write unless reads follow the leader for a moment."),
				step("sharding-strategies", "Range sharding piles hot keys on one shard; hashing spreads them."),
				step("key-value-store", "A ring, N replicas, R and W quorums, and sloppy quorums while a node is down."),
				step("news-feed", "Push or pull, and the celebrity whose post would be millions of writes."),
				step("cap-pacelc", "During a partition, refuse writes or diverge; without one, pay latency or risk staleness."),
				step("transactional-outbox", "Two writes without a transaction lose events; one transaction with an outbox does not."),
				step("saga-pattern", "A failure undoes the finished steps in reverse order."),
				step("payment-system", "Idempotency keys stop double charges; the ledger always sums to zero."),
				step("request-coalescing", "A hot key expires: one database read instead of a hundred."),
			},
		},
	}
}

// PathView is a path as the command, the browser and the docs show it: each
// step with its entry's name and the command that runs it.
type PathView struct {
	Path
	Steps []StepView `json:"steps"`
}

// StepView is a PathStep with its entry's name and the command that runs it.
type StepView struct {
	PathStep
	Name    string `json:"name"`
	Command string `json:"command"`
}

// View returns p with every step's entry name and command filled in.
func (p Path) View() PathView {
	v := PathView{Path: p, Steps: make([]StepView, 0, len(p.Steps))}
	for _, st := range p.Steps {
		sv := StepView{PathStep: st, Name: st.Entry, Command: "algorithm run " + st.Entry + " --example"}
		if e, ok := Lookup(st.Entry); ok {
			sv.Name = e.Name
			if st.Input != nil {
				sv.Command = MakeSnippet(e, *st.Input).CLI
			}
		}
		v.Steps = append(v.Steps, sv)
	}
	return v
}
