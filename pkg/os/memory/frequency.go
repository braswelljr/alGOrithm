package memory

// LFU evicts the page used least often since it was loaded; ties go to the
// page loaded first. It keeps pages that were popular, which is right for
// steady workloads and slow to adapt when the working set moves.
//
// Time: O(n * frames). Space: O(n * frames).
func LFU(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	count := make([]int, frames)
	loaded := make([]int, frames)
	return replace(refs, frames, func(mem []int, t int) int {
		best := 0
		for f := range mem {
			if count[f] < count[best] || (count[f] == count[best] && loaded[f] < loaded[best]) {
				best = f
			}
		}
		return best
	}, func(f, t int, fault bool) {
		if fault {
			count[f], loaded[f] = 0, t
		}
		count[f]++
	}), nil
}

// Aging approximates LRU with an 8-bit counter per frame, as real kernels do
// with the hardware reference bit: on every reference each counter shifts
// right one bit, and the referenced frame's counter gets its top bit set. The
// frame with the smallest counter, used least in the recent past, is evicted;
// ties go to the lowest frame.
//
// Time: O(n * frames). Space: O(n * frames).
func Aging(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	age := make([]uint8, frames)
	return replace(refs, frames, func(mem []int, t int) int {
		best := 0
		for f := range mem {
			if age[f] < age[best] {
				best = f
			}
		}
		return best
	}, func(f, t int, fault bool) {
		for i := range age {
			age[i] >>= 1
		}
		if fault {
			age[f] = 0
		}
		age[f] |= 0x80
	}), nil
}
