package heapSort

func HeapSort(arr []int) []int {
  // Build heap (rearrange array)
  for i := len(arr)/2 - 1; i >= 0; i-- {
    heapify(arr, len(arr), i)
  }

  // One by one extract an element from heap
  for i := len(arr) - 1; i >= 0; i-- {
    // Move current root to end
    temp := arr[0]
    arr[0] = arr[i]
    arr[i] = temp

    // call max heapify on the reduced heap
    heapify(arr, i, 0)
  }

  // return arr
  return arr
}

// To heapify a subtree rooted with node i which is
func heapify(arr []int, n, i int) {
  largest := i
  l := 2*i + 1
  r := 2*i + 2

  // If left child is larger than root
  if l < n && arr[l] > arr[largest] {
    largest = l
  }

  // If right child is larger than largest so far
  if r < n && arr[r] > arr[largest] {
    largest = r
  }

  // If largest is not root
  if largest != i {
    temp := arr[i]
    arr[i] = arr[largest]
    arr[largest] = temp

    // Recursively heapify the affected sub-tree
    heapify(arr, n, largest)
  }
}
