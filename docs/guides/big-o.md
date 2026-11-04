# Understanding Big O

Big O notation says how the work an algorithm does grows as its input grows. This guide explains what that means, why it leaves things out on purpose, and how to check a stated cost yourself with `algorithm bench`. Every claim below is tied to an entry you can run; the numbers are what this repository's code reports.

The site has the same ideas with an interactive chart and a measuring panel on its Big O page, and every entry's docs page says where its own cost comes from (the `why` line of `algorithm info ID`).

## What it measures

An algorithm takes steps: comparisons, swaps, writes, visits. Big O describes how the number of steps grows with the input's size, called `n`, ignoring everything that does not change with `n`.

```bash
algorithm bench selection-sort
```

```text
    n     ops  ops / prev
  100    5540           -
  200   21248       x3.84
  400   82790       x3.90
  800  326010       x3.94
```

Each time `n` doubles, the steps multiply by about four. That is what O(n^2) means: twice the input, four times the work, on any machine, in any language. The time in seconds would change with the computer; the factor of four would not.

## Why it drops constants and smaller terms

Selection sort compares each position with every later one: `(n - 1) + (n - 2) + ... + 1 = n^2/2 - n/2` comparisons. Big O keeps only `n^2`:

```text
     n     n^2/2 - n/2    n^2/2      share of the n^2 term
    10              45       50      90%
   100           4,950    5,000      99%
 1,000         499,500  500,000      99.9%
```

- **Smaller terms stop mattering.** As `n` grows, the biggest term is almost all of the work.
- **Constant factors do not change the growth.** `n^2/2` and `5n^2` both quadruple when `n` doubles. Constants matter for speed on a given machine, but they cannot turn O(n^2) into O(n log n).

So `O(f(n))` reads "grows no faster than `f(n)`, give or take a constant factor, once `n` is large enough".

## The classes, slowest growing first

```text
class        doubling n            example to run                       why
-----------  --------------------  -----------------------------------  -------------------------------------------
O(1)         no change             algorithm run stack --ops "push 1"   push touches only the top
O(log n)     adds one step         algorithm bench binary-search        each comparison halves what is left
O(sqrt n)    about x1.4            algorithm bench jump-search          jumps of sqrt n, then a walk of sqrt n
O(n)         x2                    algorithm bench linear-search        looks at each value once
O(n log n)   a bit over x2         algorithm bench merge-sort           log n levels, n work per level
O(n^2)       x4                    algorithm bench selection-sort       a loop inside a loop
O(n^3)       x8                    algorithm info matrix-chain          every split of every sub-chain
O(2^n)       squares               algorithm info tower-of-hanoi        each extra disk doubles the moves
```

At a billion steps a second, n = 64 takes: O(log n) 6 steps, O(n) 64, O(n log n) 384, O(n^2) 4,096, O(n^3) 262,144, and O(2^n) about 585 years.

### O(log n): halving

```bash
algorithm run binary-search --generate sorted --size 16 --target 30 --trace
```

Binary search compares the target with the middle value and throws away the half that cannot hold it:

```text
16 values -> 8 -> 4 -> 2 -> 1      4 comparisons
1,000,000 values                   at most 20 comparisons
```

You can halve `n` only `log2 n` times before one value is left. Doubling the input adds one comparison:

```text
algorithm bench binary-search --sizes 1000,2000,4000,8000,16000

      n  ops (7 searches)
   1000  129
   2000  155
   4000  165
   8000  169
  16000  201
```

Sixteen times the values, less than twice the work.

### O(n log n): levels of work

Merge sort halves the list until each piece has one value, then merges pieces back together:

```text
level 0   [8 values]                    merge touches 8
level 1   [4] [4]                       merge touches 8
level 2   [2] [2] [2] [2]               merge touches 8
level 3   [1] [1] [1] [1] [1] [1] [1] [1]
          log2 8 = 3 levels  x  8 per level  =  24
```

`log n` levels of `n` work each give `n log n`. `algorithm bench merge-sort` shows each doubling costing a little more than twice (x2.30, x2.26, x2.24): the extra is the one more level.

### O(n^2): nested loops

A loop over `n` values with another loop over `n` inside it does `n * n` steps. Bubble, selection and insertion sort all do this on random values. `algorithm source selection-sort` shows the two loops.

### O(2^n) and how to escape it

Plain recursive Fibonacci computes `F(n-1)` and `F(n-2)`, and each of those does the same again, so the same values are computed over and over:

```text
F(5)
+-- F(4)
|   +-- F(3)
|   |   +-- F(2)
|   |   +-- F(1)
|   +-- F(2)
+-- F(3)            <- computed again
    +-- F(2)        <- and again
    +-- F(1)
```

The number of calls grows like `F(n)` itself, about `1.618^n`: F(25) makes 242,785 calls, F(30) makes 2,692,537. On one machine:

```text
algorithm run fibonacci 20     19 microseconds
algorithm run fibonacci 25    194 microseconds
algorithm run fibonacci 30    2.0 milliseconds      (+5 to n, x10 the time: 1.618^5 = 11)
```

Remembering each answer the first time (memoization) computes every `F(k)` once, so `algorithm run fibonacci-memo 90` is n additions: O(n) instead of O(1.618^n). Changing the algorithm beat any speedup a faster computer could give.

## Best, average and worst case

One algorithm can have several costs depending on the input:

```bash
algorithm bench insertion-sort --input sorted   # O(n): nothing to shift
algorithm bench insertion-sort                  # O(n^2) on random values
algorithm bench tree-sort                       # O(n log n): a bushy tree
algorithm bench tree-sort --input sorted        # O(n^2): the tree is a line
```

When a cost says "O(n log n) average, O(n^2) worst", the worst case is a promise about every input and the average is a promise about typical ones. A stated cost in this repository gives the worst case, and the average or best when it differs.

## Amortized cost

Some operations are usually cheap and occasionally expensive. A dynamic array appends into a spare slot, and when it is full it doubles its size and copies everything:

```bash
algorithm run dynamic-array --ops "append 1, append 2, append 3, append 4, append 5, append 6, append 7, append 8, append 9"
```

```text
append 2  full, so grew 1 to 2 slots and copied 1 values
append 3  full, so grew 2 to 4 slots and copied 2 values
append 5  full, so grew 4 to 8 slots and copied 4 values
append 9  full, so grew 8 to 16 slots and copied 8 values
```

Nine appends copied 1 + 2 + 4 + 8 = 15 values, less than two copies per append. Because each growth doubles the size, the copies add up to less than `2n` over `n` appends, so each append is **O(1) amortized**: on average over any sequence, even though one append costs O(n). Growing by a fixed 10 slots instead would make the copies add up to `n^2/20`, so each append would cost O(n).

## Space

The same notation counts memory. Merge sort needs O(n) extra space for merging; heap sort sorts in place with O(1) extra; plain recursive Fibonacci uses O(n) stack space although it takes O(1.618^n) time. `algorithm info ID` shows both.

## Measuring: `algorithm bench`

```bash
algorithm bench quick-sort
algorithm bench bubble-sort --input sorted
algorithm bench linear-search --sizes 1000,2000,4000,8000 --json
```

How it works:

1. It runs the entry at each size (100, 200, 400, 800 unless you give `--sizes`) on generated input with a fixed seed: values in the order `--input` asks for (random, sorted, reversed, nearly-sorted, few-unique, negative), a text, a grid, a graph or points, depending on what the entry takes.
2. It counts the events the entry reports (comparisons, swaps, writes, visits). Counts are exact and the same on every run, unlike time.
3. For each class it finds the constant `c` that makes `c * f(n)` closest to the counts, by least squares on the relative error, and keeps the class with the smallest error.
4. It compares that class with the entry's stated cost.

How to read the result:

```text
fit      O(n log n) on ops, off by 9.1% (good)
next     O(n), off by 19.3%
stated   O(n log n) average, O(n^2) worst
verdict  agrees: it grows no faster than stated
```

- **off by**: how far the fitted curve is from the counts, on average. Under 10% is a good fit, under 25% rough.
- **next**: the runner-up. When the two are close, measure at more and larger sizes.
- **verdict**: whether the fit grows no faster than the worst stated cost. A cost written in another size, such as O(n * m) with the pattern's length fixed, cannot be compared and says so.
- Data structures cannot be measured this way, because they report one event per operation whatever that operation costs inside. Entries that report no events are fitted on time, which is noisy, and the result says so.

`TestStatedCostsAreHonest` in `internal/core` measures every traced sort and search, and more, this way and fails when one grows faster than its entry says.

## Common traps

- **The worst case of a hash table is O(n).** Lookups are O(1) on average only while keys spread over the buckets. If every key lands in one bucket, a lookup walks a list of all of them. See `algorithm run hash-table --example` and the collision entries, such as `algorithm info linear-probing`.
- **Quick sort's worst case depends on its pivot rule.** This quick sort takes the middle value as its pivot, so sorted input is its best case (`algorithm bench quick-sort --input sorted` fits O(n log n)). A pivot that is always the first value would make sorted input the O(n^2) worst case. The worst case is an input built against the rule, which is why intro sort (`algorithm info intro-sort`) switches to heap sort when the recursion gets too deep.
- **Constant factors win on small inputs.** For 10 values, insertion sort's few simple steps beat merge sort's bookkeeping, which is why tim sort (`algorithm info tim-sort`) sorts short runs by insertion before merging. Big O tells you what happens as `n` grows, not which is faster at `n = 10`.
- **O(n) with a large constant can lose to O(n log n).** `log2 n` is under 30 for any input that fits in memory, so a large constant can matter more than the log.
- **Input size is not always the number of values.** Counting sort is O(n + k) for values in a range of `k`; knapsack is O(n * capacity), where the capacity is a number, so doubling the capacity's digits squares the work.

## Related

- [Getting started](getting-started.md): the command, the library and the visualizer.
- [Catalog](../catalog.md): every entry with its stated cost.
- `algorithm info ID`: the cost, and the `why` line that explains it.
