package core

// Notes are the shortcuts to understanding an entry: what it is in one line,
// the idea that makes it work, when to reach for it, and the mistake people
// make with it. They are written for a student meeting the entry for the
// first time, and shown as the TL;DR card on the site and in the cheat sheets.
type Notes struct {
	// TLDR is the entry in one sentence.
	TLDR string `json:"tldr"`
	// KeyIdea is the one insight that makes it work.
	KeyIdea string `json:"keyIdea"`
	// WhenToUse says where it is the right tool.
	WhenToUse string `json:"whenToUse"`
	// Pitfall is the mistake or surprise to watch for.
	Pitfall string `json:"pitfall"`
}

// notes holds the Notes by entry ID; Entries copies them in. Each field is a
// sentence ending in a period (TestNotesNameRealEntries checks it).
var notes = map[string]Notes{
	// sorting
	"bubble-sort": {
		"Swap neighbors that are out of order, pass after pass, until a pass swaps nothing.",
		"Each pass carries the largest remaining value to the end, like a bubble rising.",
		"Teaching, or an almost-sorted list where one early-exit pass is all it takes.",
		"It is O(n^2) on random data; forgetting the early exit makes even sorted input quadratic.",
	},
	"selection-sort": {
		"Find the smallest remaining value and swap it into the next slot.",
		"After pass i the first i slots hold the i smallest values, in order, for good.",
		"When writes are expensive: it makes at most n - 1 swaps.",
		"It always makes about n^2/2 comparisons, even on sorted input, and it is not stable.",
	},
	"insertion-sort": {
		"Take each value and slide it left into its place among the sorted ones before it.",
		"The prefix is always sorted, so each new value only moves past the larger ones.",
		"Small or nearly sorted inputs; real libraries switch to it below about 16 values.",
		"Reversed input is its worst case: every value travels the whole prefix.",
	},
	"merge-sort": {
		"Split in half, sort each half, then merge the two sorted halves.",
		"Merging two sorted lists takes one pass, and halving gives log n levels of passes.",
		"When you need a stable O(n log n) sort, or are sorting linked lists or data on disk.",
		"The array version needs O(n) extra memory for merging.",
	},
	"quick-sort": {
		"Pick a pivot, put smaller values left and larger right, then sort each side.",
		"After partitioning the pivot is in its final place, and the two sides never mix again.",
		"General in-memory sorting where average speed matters more than stability.",
		"A bad pivot every time (such as always the first value on sorted input) makes it O(n^2).",
	},
	"heap-sort": {
		"Build a max-heap, then repeatedly move the root to the end and repair the heap.",
		"A heap finds the maximum in O(1) and restores itself in O(log n).",
		"When you need a guaranteed O(n log n) with O(1) extra memory.",
		"It is not stable and jumps around memory, so it is often slower than quick sort in practice.",
	},
	"shell-sort": {
		"Insertion sort over values far apart, with the gap shrinking to 1.",
		"Long jumps move values close to their place early, so the final pass has little left to do.",
		"Embedded code that wants something faster than O(n^2) without recursion or extra memory.",
		"Its cost depends heavily on the gap sequence; halving gaps is one of the weaker choices.",
	},
	"counting-sort": {
		"Count how many times each value occurs, then write them back in order.",
		"With values in a small range, counting replaces comparing.",
		"Integers or keys in a known small range, and as the step inside radix sort.",
		"Memory grows with the range of values, not the number of them; a range of a billion is a problem.",
	},
	"radix-sort": {
		"Sort by the last digit, then the next, with a stable sort each time.",
		"Stability keeps earlier digit orders intact, so after the last digit everything is in order.",
		"Many fixed-width integers or strings, where it beats comparison sorts.",
		"It only works on keys that split into digits, and each pass costs O(n + base).",
	},
	"bucket-sort": {
		"Spread values into buckets by range, sort each bucket, and join them.",
		"If values are spread evenly, each bucket holds about one value.",
		"Uniformly distributed numbers, such as floats in [0, 1).",
		"Clustered data piles into one bucket and the cost falls back to the inner sort's worst case.",
	},
	"tim-sort": {
		"Find runs that are already sorted, extend short ones with insertion sort, then merge runs.",
		"Real data is often partly sorted, and merging existing runs is cheap.",
		"It is the default stable sort in Python and Java for objects, a safe general choice.",
		"Its many details (galloping, run stack rules) are easy to get wrong when writing your own.",
	},
	"intro-sort": {
		"Quick sort that switches to heap sort when recursion gets too deep.",
		"Watching the depth catches the bad-pivot case and caps the worst case at O(n log n).",
		"General sorting with quick sort's speed and a worst-case guarantee; C++ std::sort works this way.",
		"It is not stable; use merge sort or tim sort when equal keys must keep their order.",
	},
	"cocktail-shaker-sort": {
		"Bubble sort that passes left to right, then right to left.",
		"Small values near the end move left quickly instead of one step per pass.",
		"Teaching the turtle problem of bubble sort.",
		"It is still O(n^2); the improvement over bubble sort is only a constant factor.",
	},

	// searching
	"linear-search": {
		"Check each value in turn until you find the target.",
		"It needs no order, only a way to compare for equality.",
		"Unsorted or tiny lists, or when you search once and sorting would cost more.",
		"On n values it can take n comparisons; repeated searches deserve sorting or a hash table.",
	},
	"binary-search": {
		"Compare with the middle and throw away the half that cannot hold the target.",
		"Each comparison halves what is left, so n values need about log2 n steps.",
		"Any sorted list, and any monotone yes/no question such as the first day a value passes a limit.",
		"Off-by-one errors in the bounds, and middle = (lo + hi) / 2 overflowing in fixed-width integers.",
	},
	"lower-bound": {
		"The first position whose value is not less than the target.",
		"It is binary search that keeps going after a match, so it handles duplicates and misses.",
		"Inserting into a sorted list, counting values in a range, and finding where a value would go.",
		"It returns a position even when the target is missing; check the value there before using it.",
	},
	"interpolation-search": {
		"Guess where the target sits from its value, like opening a dictionary near the right letter.",
		"On evenly spread values the guess lands close, giving about log log n probes.",
		"Large sorted arrays of uniformly distributed numbers.",
		"Skewed data can make it probe one slot at a time: O(n) in the worst case.",
	},
	"exponential-search": {
		"Double a bound until it passes the target, then binary search inside it.",
		"The cost depends on where the target is, not on the length of the list.",
		"Unbounded or very long sorted sequences where the target is near the front.",
		"It still needs sorted data, and it is no faster than binary search when the target is far away.",
	},
	"jump-search": {
		"Jump ahead in blocks of sqrt(n), then search the one block that can hold the target.",
		"sqrt(n) jumps plus sqrt(n) steps balance out at O(sqrt n).",
		"Sorted data where jumping back is costly, such as tape or a linked structure with skips.",
		"Binary search is faster on arrays; jump search only wins when moving backward is expensive.",
	},
	"quickselect": {
		"Find the k-th smallest value by partitioning like quick sort but recursing into one side only.",
		"Only the side holding position k matters, so the work shrinks geometrically: O(n) on average.",
		"Medians, percentiles and top-k without sorting everything.",
		"Bad pivots make it O(n^2); median of medians guarantees O(n) at a higher constant.",
	},
	"ternary-search": {
		"Find the peak of a function that rises then falls by cutting the range into thirds.",
		"Comparing two inner points tells you which outer third cannot hold the peak.",
		"Maximizing a unimodal function, such as the best price on a smooth demand curve.",
		"For searching a sorted list it is slower than binary search; it is for unimodal functions.",
	},

	// pathfinding
	"breadth-first-search": {
		"Explore the grid ring by ring from the start, nearest cells first.",
		"A queue visits cells in order of distance, so the first time you reach a cell is the shortest way.",
		"Shortest paths when every step costs the same: mazes, grids, social hops.",
		"With different step costs it finds the fewest steps, not the cheapest path.",
	},
	"depth-first-search": {
		"Follow one path as far as it goes, then back up and try the next.",
		"A stack (or recursion) remembers where to back up to, using memory only for the current path.",
		"Exploring everything reachable, detecting cycles, generating mazes, and topological order.",
		"The first path it finds is usually not the shortest.",
	},
	"dijkstra": {
		"Always expand the cheapest known cell next, so each cell's cost is final when it is expanded.",
		"With no negative costs, nothing found later can be cheaper than the current cheapest.",
		"Shortest paths with different non-negative costs: roads, network routing.",
		"Negative edge costs break it; use Bellman-Ford instead.",
	},
	"a-star": {
		"Dijkstra steered toward the goal: expand the cell with the lowest cost so far plus estimated cost left.",
		"A heuristic that never overestimates keeps the path optimal while exploring far less.",
		"Games and maps, wherever you have one goal and a good distance estimate.",
		"An overestimating heuristic finds paths faster but they may not be the shortest.",
	},
	"greedy-best-first": {
		"Always expand the cell that looks closest to the goal, ignoring the cost so far.",
		"Trusting only the estimate heads straight for the goal.",
		"Quick, good-enough paths where optimality does not matter.",
		"Walls in the way can lead it into long detours; the path it returns may be far from shortest.",
	},

	// dynamic programming
	"coin-change": {
		"The fewest coins that make an amount, built up from smaller amounts.",
		"The best way to make x uses one coin c plus the best way to make x - c.",
		"Making change with arbitrary coin systems, and as the model for any unbounded-choice DP.",
		"Greedy (always the largest coin) fails for coins like 1, 3, 4 and amount 6.",
	},
	"edit-distance": {
		"The fewest insertions, deletions and substitutions that turn one word into another.",
		"The answer for two prefixes depends only on the answers for slightly shorter prefixes.",
		"Spell checking, DNA alignment, fuzzy matching and diff tools.",
		"The full table takes O(nm) memory; two rows are enough when you only need the distance.",
	},
	"longest-common-subsequence": {
		"The longest sequence of characters that appears in order in both texts.",
		"If the last characters match they extend the answer; otherwise drop one and take the better.",
		"Diffing files, comparing DNA, and version control merges.",
		"A subsequence can skip characters; it is not the same as the longest common substring.",
	},
	"longest-increasing-subsequence": {
		"The longest run of values, not necessarily adjacent, that only goes up.",
		"For each value, the best chain ending at it extends the best smaller one before it.",
		"Scheduling, box stacking, and as the classic O(n^2) DP with an O(n log n) improvement.",
		"The O(n log n) version's tails array is not itself a valid subsequence.",
	},
	"knapsack-01": {
		"Pick items to fit a weight limit with the greatest total value, each item at most once.",
		"For each item and capacity, either skip it or take it and solve the smaller capacity.",
		"Budget and resource allocation where items cannot be split.",
		"It is pseudo-polynomial: O(n * W) grows with the capacity's value, not its number of digits.",
	},
	"matrix-chain": {
		"The cheapest order to multiply a chain of matrices.",
		"The best split of a chain is the cheapest of every split point plus the best of each side.",
		"Query planning, and as the model for interval DP problems.",
		"The order changes the cost, never the result, so it is easy to forget it matters.",
	},
	"subset-sum": {
		"Can some of the numbers add up exactly to a target?",
		"A table of reachable sums grows one number at a time.",
		"Partitioning work evenly, and as the gateway to NP-complete problems.",
		"Like knapsack it is pseudo-polynomial: fast for small targets, hopeless for huge ones.",
	},
	"unique-paths": {
		"Count the paths from the top-left corner to the bottom-right moving only right or down.",
		"Paths to a cell are the paths to the cell above plus the cell to the left.",
		"A first DP problem; grid counting with obstacles.",
		"Without obstacles the answer is a binomial coefficient, and it overflows quickly.",
	},

	// string matching
	"naive-search": {
		"Try the pattern at every position of the text.",
		"It is the baseline every faster matcher is measured against.",
		"Short patterns and texts, where its simplicity wins.",
		"Patterns like aaab in aaaa...a make it O(nm).",
	},
	"knuth-morris-pratt": {
		"Search without ever moving backward in the text, using what the pattern already matched.",
		"A prefix table says how far the pattern can shift after a mismatch.",
		"Streaming text, or patterns with repeated parts.",
		"Building the prefix table is the subtle part; an off-by-one there breaks every search.",
	},
	"rabin-karp": {
		"Compare rolling hashes of the pattern and each window, and check characters only on a hash match.",
		"Sliding the window updates the hash in O(1).",
		"Searching for many patterns of the same length, and plagiarism detection.",
		"Hash collisions force character checks; a poor hash makes it O(nm).",
	},
	"boyer-moore-horspool": {
		"Compare from the end of the pattern and skip ahead by a table on a mismatch.",
		"A mismatch on a character not in the pattern skips the pattern's whole length.",
		"Long patterns over large alphabets, like searching English text in an editor.",
		"Short patterns or tiny alphabets (DNA) give it little to skip.",
	},
	"z-algorithm": {
		"For every position, the length of the longest prefix of the text that starts there.",
		"A window of the rightmost match lets most values be copied instead of recomputed.",
		"Pattern search (on pattern + separator + text) and string periodicity.",
		"The separator must not occur in either string, or matches run across it.",
	},

	// recursion
	"factorial": {
		"n! = n * (n - 1)!, with 0! = 1.",
		"Every recursion needs a base case that stops it and a step that gets closer to it.",
		"The first recursive function to learn, and counting arrangements.",
		"It overflows a 64-bit integer at 21!.",
	},
	"fibonacci": {
		"fib(n) = fib(n - 1) + fib(n - 2), computed by plain recursion.",
		"The same subproblems are solved again and again, which is why it is exponential.",
		"Showing why memoization and dynamic programming exist.",
		"fib(50) this way takes billions of calls; never use it for real work.",
	},
	"fibonacci-memo": {
		"Fibonacci with a memo: each value is computed once and remembered.",
		"Caching subproblem answers turns exponential recursion into linear time.",
		"Any recursion whose subproblems overlap.",
		"The memo costs O(n) memory; an iterative loop with two variables needs O(1).",
	},
	"gcd": {
		"The greatest common divisor by Euclid's rule: gcd(a, b) = gcd(b, a mod b).",
		"Any common divisor of a and b also divides a mod b, and the numbers shrink fast.",
		"Simplifying fractions, LCM, and modular inverses in cryptography.",
		"Taking a mod b with b = 0 crashes; gcd(a, 0) = a is the base case.",
	},
	"tower-of-hanoi": {
		"Move n disks between pegs, never putting a larger disk on a smaller one.",
		"To move n disks, move n - 1 out of the way, move the largest, then move the n - 1 back on top.",
		"Learning to trust a recursive definition without tracing every call.",
		"It takes 2^n - 1 moves; 64 disks would take longer than the age of the universe.",
	},
	"permutations": {
		"Every ordering of a list, generated by choosing each value first and permuting the rest.",
		"Choosing, recursing and un-choosing is the shape of every backtracking search.",
		"Brute-force search over small orderings, and testing.",
		"There are n! of them; 12 values already give about 479 million.",
	},

	// graphs
	"graph-bfs": {
		"Visit a graph's nodes in order of hops from the start.",
		"A queue keeps the frontier in distance order.",
		"Shortest paths in unweighted graphs, levels, and connectivity.",
		"Mark nodes visited when you enqueue them, not when you dequeue, or nodes get queued twice.",
	},
	"graph-dfs": {
		"Go deep along one edge before trying the next, backing up at dead ends.",
		"Discovery and finish times reveal cycles, components and topological order.",
		"Cycle detection, topological sort, strongly connected components, articulation points.",
		"Deep recursion can overflow the stack on long graphs; use an explicit stack.",
	},
	"graph-dijkstra": {
		"Shortest paths from one source with a priority queue of tentative distances.",
		"The closest unfinished node can never get closer later when edges are non-negative.",
		"Routing, maps, and any shortest path with non-negative weights.",
		"One negative edge can make it return wrong answers silently.",
	},
	"bellman-ford": {
		"Relax every edge n - 1 times to find shortest paths, even with negative edges.",
		"A shortest path has at most n - 1 edges, so n - 1 rounds are enough.",
		"Negative weights, and detecting negative cycles (as in currency arbitrage).",
		"O(VE) is slow on big graphs; use Dijkstra when weights are non-negative.",
	},
	"floyd-warshall": {
		"Shortest paths between every pair, by allowing one more intermediate node at a time.",
		"The best path from i to j through nodes 1..k either uses k or it does not.",
		"Dense graphs and small node counts, transitive closure, all-pairs queries.",
		"O(V^3) time and O(V^2) memory rule out large graphs.",
	},
	"topological-sort": {
		"Order a DAG's nodes so every edge points forward.",
		"A node can go next once everything pointing to it has gone.",
		"Build systems, course prerequisites, task scheduling, spreadsheet recalculation.",
		"A cycle makes it impossible; check that every node was output.",
	},
	"kruskal": {
		"Build a minimum spanning tree by adding the cheapest edges that join two different trees.",
		"The cheapest edge across any cut is always safe to add.",
		"Network design: the cheapest set of cables connecting every site.",
		"Without union-find, checking whether an edge makes a cycle is slow.",
	},
	"prim": {
		"Grow a minimum spanning tree from one node, always adding the cheapest edge leaving it.",
		"The same cut property as Kruskal, applied to the tree grown so far.",
		"Dense graphs, where it pairs well with an adjacency matrix.",
		"It only spans the start node's component; a disconnected graph needs one run per component.",
	},
	"tarjan-scc": {
		"Find strongly connected components in one depth-first pass.",
		"A node's low-link value finds the root of the component it belongs to.",
		"Finding cycles of dependencies, and condensing a graph into a DAG.",
		"Updating low-link from nodes already finished in another component is the usual bug.",
	},
	"max-flow": {
		"The most that can flow from source to sink, by repeatedly pushing along augmenting paths.",
		"Residual edges let the algorithm undo earlier choices.",
		"Matching, scheduling, network capacity and image segmentation.",
		"Picking augmenting paths by DFS can be very slow; BFS (Edmonds-Karp) bounds it.",
	},
	"pagerank": {
		"Rank pages by how likely a random surfer is to be on them.",
		"A page is important if important pages link to it, solved by repeated averaging.",
		"Ranking web pages, citations, or influence in any link graph.",
		"Pages with no outgoing links leak rank unless their rank is spread over all pages.",
	},
	"cycle-detection": {
		"Does the graph contain a cycle?",
		"In a directed graph, an edge back to a node still on the DFS path closes a cycle.",
		"Deadlock detection, dependency checks, validating DAGs.",
		"Undirected and directed graphs need different tests; the parent edge is not a cycle.",
	},
	"euler-path": {
		"A walk that uses every edge exactly once.",
		"It exists exactly when zero or two nodes have odd degree, and the graph is connected.",
		"Route planning such as street sweeping and drawing a figure without lifting the pen.",
		"Do not confuse it with a Hamiltonian path, which visits every node and is NP-hard.",
	},
	"bipartite-check": {
		"Can the nodes be split into two groups with every edge between groups?",
		"Two-color by BFS; an edge between same colors means an odd cycle.",
		"Matching problems, scheduling two teams, detecting odd cycles.",
		"Check every component, not just the one holding the start node.",
	},

	// greedy and backtracking
	"activity-selection": {
		"Choose the most non-overlapping activities by always taking the one that ends first.",
		"Ending first leaves the most room for everything after it.",
		"Room booking, scheduling jobs on one machine.",
		"Sorting by start time or by length gives wrong answers.",
	},
	"huffman-coding": {
		"Give frequent symbols short codes by repeatedly merging the two rarest.",
		"Merging the two least frequent symbols is always part of an optimal code.",
		"Compression inside ZIP, PNG and JPEG.",
		"The code table must travel with the data, which can outweigh the savings on small inputs.",
	},
	"fractional-knapsack": {
		"Fill a bag by value per weight, taking a fraction of the last item.",
		"When items can be split, the best ratio first is optimal.",
		"Cutting up resources that can be divided, like bandwidth or ingredients.",
		"The same greedy rule fails for 0/1 knapsack, where items cannot be split.",
	},
	"n-queens": {
		"Place n queens so that none attack each other, one row at a time.",
		"Undo a choice as soon as it conflicts, so dead ends are cut off early.",
		"Learning backtracking, and constraint satisfaction.",
		"Without pruning by columns and diagonals it explores n^n placements.",
	},
	"graph-coloring": {
		"Give each node a color so no edge joins two nodes of the same color.",
		"Try a color, recurse, and back up when a neighbor already has it.",
		"Register allocation in compilers, exam timetabling, map coloring.",
		"Finding the fewest colors is NP-hard; greedy coloring is fast but not optimal.",
	},

	// numbers
	"sieve-of-eratosthenes": {
		"Find every prime up to n by crossing out multiples of each prime.",
		"Start crossing out at p * p, because smaller multiples were crossed out by smaller primes.",
		"Many primality questions up to a few hundred million.",
		"It needs O(n) memory; a segmented sieve handles larger ranges.",
	},
	"prime": {
		"Is n prime? Try dividing by every number up to sqrt(n).",
		"If n = a * b, one of a and b is at most sqrt(n).",
		"Checking one modest number.",
		"For cryptographic sizes use a probabilistic test such as Miller-Rabin.",
	},
	"fast-power": {
		"Compute x^n by squaring, using the bits of n.",
		"x^n is (x^(n/2))^2 for even n, so it takes about log2 n multiplications.",
		"Modular exponentiation in RSA and Diffie-Hellman, and matrix powers.",
		"Reduce modulo m after every multiplication, or the numbers overflow.",
	},
	"extended-gcd": {
		"Find x and y with a*x + b*y = gcd(a, b).",
		"Run Euclid's algorithm backward to express the gcd as a combination.",
		"Modular inverses, the core of RSA key generation.",
		"An inverse exists only when gcd(a, m) = 1.",
	},
	"karatsuba": {
		"Multiply large numbers with three half-size products instead of four.",
		"(a + b)(c + d) - ac - bd gives the middle term with one multiplication.",
		"Big integer arithmetic, where it beats schoolbook multiplication beyond a few hundred digits.",
		"For small numbers its overhead makes it slower than the schoolbook method.",
	},

	// hashing
	"linear-probing": {
		"On a collision, try the next slot, then the next.",
		"Neighboring slots share cache lines, so probing is fast in practice.",
		"Open addressing tables with a low load factor.",
		"Clusters form and grow; keep the load factor under about 0.7.",
	},
	"consistent-hashing": {
		"Place servers and keys on a ring; a key belongs to the next server clockwise.",
		"Adding or removing a server moves only the keys next to it.",
		"Distributed caches and databases that add and remove nodes.",
		"Without virtual nodes the load is uneven.",
	},
	"cuckoo-hashing": {
		"Each key has two possible slots; a newcomer kicks the old key to its other slot.",
		"Lookups check at most two places, so they are O(1) in the worst case.",
		"Tables where lookups must be predictably fast.",
		"Inserts can loop forever; detect the loop and rehash.",
	},

	// arrays
	"dynamic-array": {
		"An array that doubles its capacity when it runs out of room.",
		"Doubling makes the copying cost O(1) amortized per append.",
		"Go slices, Python lists, JavaScript arrays: the default list.",
		"An append that triggers growth copies everything; references into the old array go stale.",
	},
	"two-pointers": {
		"Walk two indices toward each other over a sorted array to find a pair.",
		"If the sum is too small move the left pointer up; if too large move the right down.",
		"Pair sums, removing duplicates, and merging sorted arrays.",
		"It relies on sorted order; on unsorted data use a hash set.",
	},
	"prefix-sums": {
		"Precompute running totals so any range sum is one subtraction.",
		"sum(i..j) = prefix[j + 1] - prefix[i].",
		"Many range-sum queries on data that does not change.",
		"Updates force an O(n) rebuild; use a Fenwick tree when values change.",
	},
	"max-subarray": {
		"Kadane's algorithm: the largest sum of any contiguous run.",
		"A run worth extending never has a negative prefix, so restart when the sum drops below zero.",
		"Best profit windows, signal peaks, and a classic interview DP.",
		"All-negative input: decide whether the answer is the largest value or an empty run.",
	},
	"sliding-window-max": {
		"The maximum of every window of k values, in O(n) total.",
		"A deque keeps only values that could still be a maximum.",
		"Rolling maxima over time series and stream monitoring.",
		"Store indices, not values, so you know when the front has left the window.",
	},
	"rotate-array": {
		"Rotate by k with three reversals: the whole array, then each part.",
		"Reversing twice puts each part back in order at its new place.",
		"Rotating in place with O(1) extra memory.",
		"Take k modulo n first; k larger than the length is common.",
	},
	"js-map": {
		"Array.prototype.map: a new array with a function applied to each element.",
		"It transforms without changing the original.",
		"Turning one list into another of the same length.",
		"Using map only for its side effects (where forEach is meant) builds an array nobody uses.",
	},
	"js-reduce": {
		"Array.prototype.reduce: fold an array into one value with an accumulator.",
		"Each step combines the accumulator with the next element.",
		"Sums, grouping, and building objects from lists.",
		"Without an initial value, an empty array throws.",
	},

	// sets and trees
	"bitset": {
		"A set of small integers stored one bit per possible member.",
		"Union and intersection become single machine instructions on whole words.",
		"Dense sets of small integers: flags, sieves, graph adjacency.",
		"Memory follows the largest possible member, not the number of members.",
	},
	"power-set": {
		"Every subset of a set: 2^n of them.",
		"Each subset is a binary number saying which members are in.",
		"Brute force over small sets, and test generation.",
		"n = 30 is already a billion subsets.",
	},
	"tree-inorder": {
		"Visit the left subtree, the node, then the right subtree.",
		"On a binary search tree it yields the keys in sorted order.",
		"Printing a BST in order, and validating one.",
		"Recursion depth is the tree's height, which is n for a degenerate tree.",
	},
	"tree-level-order": {
		"Visit a tree level by level with a queue.",
		"The queue holds one level while the next is being added.",
		"Printing by depth, finding the shallowest leaf, serializing trees.",
		"A wide tree's last level can hold half the nodes in the queue at once.",
	},
	"validate-bst": {
		"Check that every key is between the bounds its ancestors set.",
		"Each node must respect all of its ancestors, not only its parent.",
		"Checking a tree built by hand or received from outside.",
		"Comparing a node only with its children misses violations deeper down.",
	},
	"lowest-common-ancestor": {
		"The deepest node that has both given nodes below it.",
		"Lift both nodes to the same depth, then lift them together until they meet.",
		"Family trees, file system paths, and distances in trees.",
		"Repeated queries deserve binary lifting rather than walking up each time.",
	},
	"tree-height-diameter": {
		"The height of a tree and its longest path between any two nodes.",
		"The longest path through a node is its two deepest subtrees joined.",
		"Measuring trees, and the core of many tree DP problems.",
		"The diameter does not always pass through the root.",
	},

	// structures
	"stack": {
		"Last in, first out: push on top, pop from the top.",
		"Only the top is reachable, which matches nested work such as function calls.",
		"Undo, matching brackets, depth-first search, expression evaluation.",
		"Popping an empty stack: check, do not assume.",
	},
	"queue": {
		"First in, first out: add at the back, remove from the front.",
		"Things are served in the order they arrived.",
		"Breadth-first search, task queues, buffering.",
		"A slice-backed queue that only reslices the front leaks memory; use a ring buffer.",
	},
	"deque": {
		"A double-ended queue: add and remove at both ends in O(1).",
		"It is a stack and a queue at once.",
		"Sliding-window problems, work stealing, palindromes.",
		"Indexing into the middle is not what it is for.",
	},
	"linked-list": {
		"Nodes that each point to the next.",
		"Inserting or deleting next to a node you hold is O(1), with no shifting.",
		"Queues, LRU caches, and splicing without copying.",
		"Finding the k-th element is O(k), and scattered nodes are cache-unfriendly.",
	},
	"doubly-linked-list": {
		"A linked list whose nodes also point back.",
		"Removing a node you hold is O(1) because you can reach its predecessor.",
		"LRU caches and editors' undo lists.",
		"Every insert and delete updates four pointers; missing one corrupts the list.",
	},
	"circular-buffer": {
		"A fixed-size array where the end wraps around to the start.",
		"Head and tail indices modulo the size make a queue with no shifting.",
		"Audio buffers, logs of the last n events, producer-consumer queues.",
		"Full and empty look the same unless you keep a count or leave one slot free.",
	},
	"min-stack": {
		"A stack that also answers its minimum in O(1).",
		"Each entry remembers the minimum at the time it was pushed.",
		"Tracking the lowest price or the best score while undoing.",
		"It doubles memory unless you store a minimum only when it changes.",
	},
	"monotonic-queue": {
		"A deque kept in sorted order by dropping values that can never be the answer.",
		"A newer, larger value makes every older smaller one useless for maxima.",
		"Sliding-window maximum or minimum in O(n).",
		"Pop from the front only when the front's index leaves the window.",
	},
	"min-heap": {
		"A tree in an array where every parent is smaller than its children.",
		"Only the root is guaranteed smallest, so keeping that property costs just O(log n).",
		"Priority queues: Dijkstra, schedulers, merging k sorted lists, top-k.",
		"It is not sorted; iterating it does not give the values in order.",
	},
	"binary-search-tree": {
		"A tree where everything left is smaller and everything right is larger.",
		"Each comparison chooses one subtree, so search follows one path.",
		"Ordered maps with range queries and in-order iteration.",
		"Inserting sorted keys makes it a linked list with O(n) operations.",
	},
	"avl-tree": {
		"A binary search tree that rotates to keep subtree heights within one of each other.",
		"Balance keeps the height at about 1.44 log n, so every operation is O(log n).",
		"Lookup-heavy ordered maps.",
		"Getting the four rotation cases right is the hard part.",
	},
	"red-black-tree": {
		"A self-balancing BST that colors nodes red or black to bound its height.",
		"No red node has a red child, so the longest path is at most twice the shortest.",
		"Ordered maps in standard libraries (C++ std::map, Java TreeMap).",
		"Deletion has many cases; it is easier to use a tested library than to write one.",
	},
	"b-tree": {
		"A balanced tree whose nodes hold many keys, built for disks.",
		"Wide nodes make the tree shallow, so a lookup reads only a few disk pages.",
		"Database indexes and file systems.",
		"Splitting and merging nodes on insert and delete is fiddly.",
	},
	"skip-list": {
		"A sorted linked list with express lanes chosen by coin flips.",
		"Each level skips about half the nodes below, giving O(log n) expected search.",
		"Concurrent ordered maps (Redis sorted sets, LevelDB's memtable).",
		"Its bounds are expected, not guaranteed; bad luck makes it slow.",
	},
	"fenwick-tree": {
		"Prefix sums with updates, both in O(log n).",
		"The lowest set bit of an index says how many values that slot covers.",
		"Running totals that change: leaderboards, inversion counting.",
		"It is 1-indexed; translating from 0-indexed code is the classic bug.",
	},
	"segment-tree": {
		"A tree over ranges that answers range queries and updates in O(log n).",
		"Each node stores the answer for its range, combined from its two halves.",
		"Range minimum, maximum or sum with updates.",
		"It takes up to 4n slots; sizing it at 2n overflows.",
	},
	"hash-table": {
		"Keys map to slots by a hash function, for O(1) average lookups.",
		"A good hash spreads keys evenly, so each slot holds about one key.",
		"Dictionaries, caches, counting, deduplication.",
		"Its worst case is O(n), and iteration order is not stable.",
	},
	"trie": {
		"A tree of characters where each path from the root spells a key.",
		"Keys with the same prefix share a path, so prefix queries are fast.",
		"Autocomplete, spell checking, IP routing tables.",
		"A node per character can use much more memory than a hash set.",
	},
	"union-find": {
		"Track which items are in the same group, and merge groups.",
		"Path compression and union by rank make operations nearly O(1).",
		"Kruskal's algorithm, connected components, image labeling.",
		"It cannot split groups once they are merged.",
	},
	"lru-cache": {
		"A fixed-size cache that evicts the least recently used item.",
		"A hash map finds items and a linked list keeps them in use order, both O(1).",
		"Caching web pages, database rows, function results.",
		"Updating recency on reads, not only writes, is easy to forget.",
	},
	"bloom-filter": {
		"A compact set that may say yes wrongly but never says no wrongly.",
		"Several hashes set bits; a member has all of them set.",
		"Skipping disk reads for keys that are certainly absent (databases, caches).",
		"You cannot remove items, and the false-positive rate climbs as it fills.",
	},

	// operating systems
	"fcfs-scheduling": {
		"Run processes in the order they arrive, each to completion.",
		"Simple and fair in arrival order, with no preemption.",
		"Batch systems, and as the baseline for comparing schedulers.",
		"One long job makes everyone behind it wait: the convoy effect.",
	},
	"sjf-scheduling": {
		"Run the shortest job next.",
		"Short jobs first minimizes the average waiting time.",
		"Batch work where job lengths are known.",
		"Long jobs can starve, and real burst lengths have to be guessed.",
	},
	"srtf-scheduling": {
		"Preemptive SJF: switch whenever a job arrives with less time left.",
		"Always running the job closest to finishing keeps average waiting lowest.",
		"Showing the best possible average waiting time.",
		"Frequent switching costs time, and long jobs starve.",
	},
	"round-robin": {
		"Give each process a fixed time slice in turn.",
		"No process waits more than (n - 1) quanta for its next turn.",
		"Interactive, time-sharing systems.",
		"A tiny quantum wastes time switching; a huge one behaves like FCFS.",
	},
	"priority-scheduling": {
		"Run the highest-priority process first.",
		"Importance, not arrival, decides who runs.",
		"Real-time and system tasks that must preempt user work.",
		"Low-priority processes can starve; aging raises their priority over time.",
	},
	"fifo-page-replacement": {
		"Evict the page that has been in memory longest.",
		"A queue of pages is all it needs.",
		"The simplest replacement policy, and the baseline.",
		"Belady's anomaly: more frames can mean more faults.",
	},
	"lru-page-replacement": {
		"Evict the page not used for the longest time.",
		"The recent past predicts the near future.",
		"General-purpose memory and caches.",
		"Exact LRU needs a timestamp or list update on every access; hardware approximates it.",
	},
	"optimal-page-replacement": {
		"Evict the page that will not be needed for the longest time.",
		"It sees the future, so no policy can do better.",
		"Measuring how close real policies come to the best possible.",
		"It cannot be implemented for real, because the future is unknown.",
	},
	"clock-page-replacement": {
		"A circular list with a use bit: skip pages used recently, evict the first that was not.",
		"One bit per page approximates LRU cheaply.",
		"Real operating systems' page replacement.",
		"If every page is in use the hand goes all the way around, and it degrades to FIFO.",
	},
	"first-fit": {
		"Put each process in the first free block big enough.",
		"Stopping at the first fit keeps allocation fast.",
		"Simple memory allocators.",
		"Small leftover holes collect at the start of memory.",
	},
	"best-fit": {
		"Put each process in the smallest free block big enough.",
		"The tightest fit leaves the least waste in that block.",
		"When memory is tight and requests vary in size.",
		"It leaves many tiny unusable holes, and it scans every block.",
	},
	"sstf-disk": {
		"Serve the disk request nearest to the head next.",
		"Minimizing each seek reduces total head movement.",
		"Lowering average seek time under steady load.",
		"Requests far from the head can starve.",
	},
	"scan-disk": {
		"The elevator: sweep one way serving requests, then reverse.",
		"Every request is reached within two sweeps, so none starve.",
		"Disks under heavy load.",
		"Requests just behind the head wait for the round trip.",
	},
	"c-look-disk": {
		"Sweep one way up to the last request, then jump back to the lowest one.",
		"Serving in one direction only gives more uniform waiting times.",
		"Fair disk scheduling under load.",
		"The jump back still costs a seek, even though nothing is served on it.",
	},

	// beyond the classics
	"k-means": {
		"Group points into k clusters by moving each center to the mean of its points.",
		"Assigning points and moving centers each lower the total distance, so it converges.",
		"Customer segmentation, color quantization, a first look at unlabeled data.",
		"It needs k up front and can settle in a poor local minimum; k-means++ seeding helps.",
	},
	"linear-regression": {
		"Fit the straight line that minimizes the squared vertical distances to the points.",
		"The best slope and intercept have a closed form from means and covariances.",
		"Predicting a number from one feature, and as the baseline for any regression.",
		"A good fit does not mean cause and effect, and outliers pull the line hard.",
	},
	"sha-256": {
		"A hash that turns any input into 256 bits that look random.",
		"Each round mixes bits with additions, rotations and constants until every output bit depends on every input bit.",
		"Checksums, content addressing (Git), blockchains, and inside HMAC and signatures.",
		"It is not for passwords on its own; use a slow, salted function such as PBKDF2.",
	},
	"rsa": {
		"Public-key encryption: anyone can encrypt with (e, n), only the owner of d can decrypt.",
		"Exponents e and d undo each other modulo n, and finding d requires factoring n.",
		"Key exchange and signatures; learning how public-key cryptography works.",
		"Textbook RSA without padding is insecure; real systems use OAEP or PSS.",
	},
	"diffie-hellman": {
		"Two parties agree on a shared secret over a public channel.",
		"g^(ab) is easy for both sides to compute and hard for an eavesdropper.",
		"The key exchange inside TLS and SSH.",
		"Without authentication a man in the middle can do the exchange with each side.",
	},
	"raft": {
		"A consensus protocol: elect a leader, which replicates a log to followers.",
		"An entry is committed once a majority stores it, and any two majorities overlap.",
		"Replicated state machines in etcd, Consul and many databases.",
		"Network partitions and timeouts make the edge cases hard; test with randomized failures.",
	},
	"two-phase-commit": {
		"A coordinator asks everyone to prepare, then tells everyone to commit or abort.",
		"Nobody commits until everyone has promised they can.",
		"Atomic transactions across several databases.",
		"If the coordinator crashes after prepare, participants are stuck waiting.",
	},
	"tcp-congestion-control": {
		"TCP grows its window until loss, then backs off.",
		"Additive increase and multiplicative decrease share bandwidth fairly.",
		"Understanding why downloads ramp up and why loss slows everything.",
		"Treating every loss as congestion hurts on wireless links where loss is noise.",
	},
	"subnetting": {
		"Split an IP network into smaller networks by moving the prefix boundary.",
		"The prefix length sets how many bits name the network and how many name hosts.",
		"Planning networks, firewall rules, and reading routing tables.",
		"The network and broadcast addresses are not usable host addresses.",
	},
	"hash-join": {
		"Join two tables by hashing the smaller one and probing it with each row of the other.",
		"One pass to build and one to probe beats comparing every pair.",
		"Equality joins on large tables, the default in most databases.",
		"It only works for equality, and a build side larger than memory must be partitioned.",
	},
	"b-plus-tree": {
		"A B-tree whose leaves hold all the keys and are linked in order.",
		"Inner nodes only guide; linked leaves make range scans sequential.",
		"Almost every database index and many file systems.",
		"Splits and merges must keep the leaf links right.",
	},
	"write-ahead-log": {
		"Write each change to a log before changing the data itself.",
		"After a crash, replaying the log redoes committed work and undoes the rest.",
		"Durability in databases and file systems.",
		"The log must actually reach disk (fsync) before the commit is acknowledged.",
	},
}
