# AL[GO](#)RITHM
Algorithms and data structures in Golang

## Data Structures

Data Structures are representations of data sets in a way which makes storage and performing of operations of data very easy.

### Classification of Data Structures

- #### _[Linear](data-structures/linear)_
  - [List](data-structures/linear/list/list.go) : is an abstract data type that represents a finite number of ordered values, where the same value may occur more than once.
  - [Set](data-structures/linear/set/set.go) : is an abstract data type can store unique values without any particular order.
  - [Tuple](data-structures/linear/tuple/tuple.go) : is a finite ordered list (sequence) of elements. Tuple is a reference type.
  - [Queue](data-structures/linear/queue/queue.go) : is an abstract data type that implements a First-In-First-Out (`FIFO`) queue of generic items.
  - [Stack](data-structures/linear/stack/stack.go) : is an abstract data type that operates on the concept of the Last-In-First-Out (`LIFO`) principle.
  - [Heap](data-structures/linear/heap/heap.go) : is a specialized tree-based data structure that satisfies the heap property: for any node, the value of that node is greater than or equal to the values of its children.
- #### _[Non - Linear](#)_
  - [Trees](data-structures/non-linear/trees/trees.go) : is a data structure that consists of a set of nested nodes, each node having a value and a set of child nodes.
  - [Tables](data-structures/non-linear/table/table.go) : is a data structure that consists of a set of rows and columns, each row having a set of columns. The table is a two-dimensional array. The table is a reference type.
  - [Containers](data-structures/non-linear/container/container.go) : is a data structure that consists of a set of elements, each element having a key and a value. The container is a reference type. The container is a reference type.
- #### _[Homogeneous](#)_
  - [Two Dimensional Arrays](#)
  - [Multi Dimensional Arrays](#)
- #### _[Heterogeneous](#)_
  - [Linked List](data-structures/heterogeneous/linkedList): is a data structure that consists of a set of nodes, each node having a value and a reference to the next node.
    - [Single Linked List](data-structures/heterogeneous/linkedList/singleLinkedList/singleLinkedList.go) : is a data structure that consists of a set of nodes, each node having a value and a reference to the next node. The list is a reference type.
    - [Double Linked List](data-structures/heterogeneous/linkedList/doubleLinkedList/doubleLinkedList.go) : is a data structure that consists of a set of nodes, each node having a value and a reference to the previous and next node. The list is a reference type.
    - [Circular Linked List](data-structures/heterogeneous/linkedList/circularLinkedList/circularLinkedList.go) : is a data structure which has the last item contains link of the first element as next and the first element has a link to the last element as previous. The list is a reference type.
      - [Circular Single Linked List](data-structures/heterogeneous/linkedList/circularLinkedList/circularSingleLinkedList/circularSingleLinkedList.go) : is a data structure which has the last item contains link of the first element as next and the first element has a link to the last element as previous. The list is a reference type.
      - [Circular Double Linked List](data-structures/heterogeneous/linkedList/circularLinkedList/circularDoubleLinkedList/circularDoubleLinkedList.go) : is a data structure which has the last item contains link of the first element as next and the first element has a link to the last element as previous. The list is a reference type.
  - [Ordered List](data-structures/heterogeneous/orderedList/orderedList.go) : is a data structure that consists of a set of nodes, each node having a value and a reference to the next node. The list is a reference type.
  - [Unordered List](data-structures/heterogeneous/unorderedList/unorderedList.go) : is a data structure that consists of a set of nodes, each node having a value and a reference to the next node. The list is a reference type.
- #### _[Dynamic](#)_
  - [Dictionaries](#)
  - [Tree Sets](#)
  - [Sequences](#)

## Algorithms

Algorithm is a set of instructions that describes how to get something done.

### Classification of Algorithms

- #### _[Sort](algorithms/sort)_
  - [Bubble Sort](algorithms/sort/bubbleSort/bubbleSort.go) : is a simple sorting algorithm that repeatedly steps through the list, compares elements and swaps them if they are in the wrong order. Bubble Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>.
  - [Merge Sort](algorithms/sort/mergeSort/mergeSort.go) : is a comparison based sort that recursively splits the list into smaller sub-lists until the sub-lists are small enough to be sorted individually. Merge Sort was invented by [`John Von Neumman`](https://en.wikipedia.org/wiki/John_von_Neumann) in _1945_. It has a **time complexity** of <code>O(n log n)</code>.
  - [Comb Sort](#) : is a sorting algorithm that uses a gap sequence to sort the array. It is a variant of the Bubble Sort algorithm. Comb Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>. The algorithm is a variation of the [bubble sort](algorithms/sort/bubbleSort/bubbleSort.go) algorithm. It was originally designed by [`Włodzimierz Dobosiewicz`](#) and [`Artur Borowy`](#) in 1980
  - [Heap Sort](algorithms/sort/heapSort/heapSort.go) : is a comparison based sort that uses a heap data structure to sort the list. Heap Sort has a **time complexity** of <code>O(n log n)</code>.
  - [Quick Sort](algorithms/sort/quickSort/quickSort.go) : is a comparison based sort that uses a divide and conquer strategy to sort the list. Quick Sort has a **time complexity** of <code>O(n log n)</code>.
  - [Selection Sort](algorithms/sort/selectionSort/selectionSort.go) : is a comparison based sort that finds the smallest element in the list and places it at the beginning. Selection Sort has a **time complexity** of <code>O(n<sup>2</sup>)</code>.
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

- #### _[Recursion](algorithms/recursion)_
  - [Fibonacci](algorithms/recursion/fibonacci/fibonacci.go) : is a recursive algorithm that adds the two preceding numbers to produce the next number in the sequence.
  - [Factorial](algorithms/recursion/factorial/factorial.go) : is a recursive algorithm that multiplies all the preceding numbers to produce the next number in the sequence.
  - [Euclidean / GCD](algorithms/recursion/euclidean/euclidean.go) : is a recursive algorithm that finds the greatest common divisor of two numbers.
  - [LCM](algorithms/recursion/lcm/lcm.go) : is a recursive algorithm that finds the least common multiple of two numbers.
  - [Tower of Hanoi](algorithms/recursion/towerOfHanoi/tower_of_hanoi.go) : is a recursive algorithm that moves a stack of disks from one tower to another.
  - [Ackermann Function](algorithms/recursion/ackermann/ackermann.go) : is a recursive algorithm that finds the value of the Ackermann function. It is a total recursive function that can be defined in terms of itself.
  - [McCarthy Function](algorithms/recursion/mccarthy91/mccarthy91.go) : is a recursive algorithm that finds the value of the McCarthy function. It is a total recursive function that can be defined in terms of itself.
  - [Palindrome](algorithms/recursion/palindrome/palindrome.go) : is an algorithm that checks if a word is equal to its reversed.

- #### [Search](algorithms/search)
  - [Binary Search](algorithms/search/binarySearch/binarySearch.go) : is a search algorithm that finds the position of a target value in a sorted array. It has a **time complexity** of <code>O(log n)</code> and a **space complexity** of <code>O(1)</code>.
  - [Interpolation Search](algorithms/search/interpolationSearch/interpolationSearch.go) : is a search algorithm that finds the position of a target value in a sorted array.
  - [Jump to Search](algorithms/search/jumpSearch/jumpSearch.go) : is a search algorithm that finds the position of a target value in a sorted array. It is a variation of the [Binary Search](algorithms/search/binarySearch/binarySearch.go) algorithm. It has a **time complexity** of <code>O(log n)</code> and a **space complexity** of <code>O(1)</code>.
  - [Linear Search](algorithms/search/linearSearch/linearSearch.go) : is a search algorithm that finds the position of a target value in a sorted array. It has a **time complexity** of <code>O(n)</code> and a **space complexity** of <code>O(1)</code>.
  - [Ternary Search](algorithms/search/ternarySearch/ternarySearch.go) : is a search algorithm that finds the position of a target value in a sorted array.

- #### [Cipher](algorithms/cipher)
  - [Caesar Cipher](algorithms/cipher/caesarCipher/caesarCipher.go) : is a cipher that shifts each letter in a message by a certain number of places. It has a time complexity of <code>O(n)</code>.
  - [Vigenere Cipher](algorithms/cipher/vigenereCipher/vigenereCipher.go) : is a cipher that shifts each letter in a message by a certain number of places. It has a time complexity of <code>O(n)</code>.
  - [Hill Cipher](#) : is a cipher that shifts each letter in a message by a certain number of places. It has a time complexity of <code>O(n)</code>.
  - [Rot13 Cipher](#) : is a cipher that shifts each letter in a message by 13 number of places. It has a time complexity of <code>O(n)</code>.
- #### _[Graph](algorithms/graph)_
