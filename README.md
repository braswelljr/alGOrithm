# AL[GO](#)RITHM
Algorithms and data structures in Golang

## Data Structures

Data Structures are representations of data sets in a way which makes storage and performing of operations of data very easy.

### Classification of Data Structures

- #### _[Linear](src/data_structures/linear)_
  - [List](src/data_structures/linear/list/list.go) : is an abstract data type that represents a finite number of ordered values, where the same value may occur more than once.
  - [Set](src/data_structures/linear/set/set.go) : is an abstract data type can store unique values without any particular order.
  - [Tuple](src/data_structures/linear/tuple/tuple.go) : is a finite ordered list (sequence) of elements. Tuple is a reference type.
  - [Queue](src/data_structures/linear/queue/queue.go) : is an abstract data type that implements a First-In-First-Out (`FIFO`) queue of generic items.
  - [Stack](src/data_structures/linear/stack/stack.go) : is an abstract data type that operates on the concept of the Last-In-First-Out (`LIFO`) principle.
  - [Heap](src/data_structures/linear/heap/heap.go) : is a specialized tree-based data structure that satisfies the heap property: for any node, the value of that node is greater than or equal to the values of its children.
- #### _[Non - Linear](#)_
  - [Trees](#)
  - [Tables](#)
  - [Containers](#)
- #### _[Homogeneous](#)_
  - [Two Dimensional Arrays](#)
  - [Multi Dimensional Arrays](#)
- #### _[Heterogeneous](#)_
  - [Linked List](#)
  - [Ordered List](#)
  - [Unordered List](#)
- #### _[Dynamic](#)_
  - [Dictionaries](#)
  - [Tree Sets](#)
  - [Sequences](#)

## Algorithms

Algorithm is a set of instructions that describes how to get something done.

### Classification of Algorithms

- #### _[Sort](src/algorithms/sort)_
  - [Bubble Sort](src/algorithms/sort/bubbleSort/bubbleSort.go) : is a simple sorting algorithm that repeatedly steps through the list, compares elements and swaps them if they are in the wrong order. Bubble Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>.
  - [Merge Sort](src/algorithms/sort/mergeSort/mergeSort.go) : is a comparison based sort that recursively splits the list into smaller sub-lists until the sub-lists are small enough to be sorted individually. Merge Sort was invented by [`John Von Neumman`](https://en.wikipedia.org/wiki/John_von_Neumann) in _1945_. It has a **time complexity** of <code>O(n log n)</code>.
  - [Comb Sort](#) : is a sorting algorithm that uses a gap sequence to sort the array. It is a variant of the Bubble Sort algorithm. Comb Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>. The algorithm is a variation of the [bubble sort](src/algorithms/sort/bubbleSort/bubbleSort.go) algorithm. It was originally designed by [`Włodzimierz Dobosiewicz`](#) and [`Artur Borowy`](#) in 1980
  - [Heap Sort](src/algorithms/sort/heapSort/heapSort.go) : is a comparison based sort that uses a heap data structure to sort the list. Heap Sort has a **time complexity** of <code>O(n log n)</code>.
  - [Quick Sort](src/algorithms/sort/quickSort/quickSort.go) : is a comparison based sort that uses a divide and conquer strategy to sort the list. Quick Sort has a **time complexity** of <code>O(n log n)</code>.
  - [Selection Sort](src/algorithms/sort/selectionSort/selectionSort.go) : is a comparison based sort that finds the smallest element in the list and places it at the beginning. Selection Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>.
  - [Insertion Sort](#) : is a comparison based sort that builds the final sorted array one item at a time. Insertion Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>.
  - [Radix Sort](#)
  - [Shell Sort](#)
  - [Tree Sort](#)
  - [Bucket Sort](#)
  - [Counting Sort](#)
  - [Smooth Sort](#)
  - [Bogo Sort](#)
  - [Cycle Sort](#)
  - [Gnome Sort](#)
  - [Stooge Sort](#)

- #### _[Recursion](src/algorithms/recursion)_
  - [Fibonacci](src/algorithms/recursion/fibonacci/fibonacci.go) : is a recursive algorithm that adds the two preceding numbers to produce the next number in the sequence.
  - [Factorial](src/algorithms/recursion/factorial/factorial.go) : is a recursive algorithm that multiplies all the preceding numbers to produce the next number in the sequence.
  - [Euclidean / GCD](src/algorithms/recursion/euclidean/euclidean.go) : is a recursive algorithm that finds the greatest common divisor of two numbers.
  - [LCM](src/algorithms/recursion/lcm/lcm.go) : is a recursive algorithm that finds the least common multiple of two numbers.
  - [Tower of Hanoi](src/algorithms/recursion/tower_of_hanoi/tower_of_hanoi.go) : is a recursive algorithm that moves a stack of disks from one tower to another.
  - [Ackermann Function](src/algorithms/recursion/ackermann/ackermann.go) : is a recursive algorithm that finds the value of the Ackermann function. It is a total recursive function that can be defined in terms of itself.
  - [McCarthy Function](src/algorithms/recursion/mccarthy91/mccarthy91.go) : is a recursive algorithm that finds the value of the McCarthy function. It is a total recursive function that can be defined in terms of itself.

- #### [Search](src/algorithms/search)
  - [Binary Search](src/algorithms/search/binarySearch/binarySearch.go) : is a search algorithm that finds the position of a target value in a sorted array.
  - [Interpolation Search](src/algorithms/search/interpolationSearch/interpolationSearch.go) : is a search algorithm that finds the position of a target value in a sorted array.
  - [Jump to Search](src/algorithms/search/jumpSearch/jumpSearch.go) : is a search algorithm that finds the position of a target value in a sorted array.
  - [Linear Search](src/algorithms/search/linearSearch/linearSearch.go) : is a search algorithm that finds the position of a target value in a sorted array.
  - [Ternary Search](src/algorithms/search/ternarySearch/ternarySearch.go) : is a search algorithm that finds the position of a target value in a sorted array.

- #### _[Graph](src/algorithms/graph)_
