package set

// Sets are abstract data type can store unique values without any particular order.
// Sets are abstract data that can perform the following operations:
// 		- Add an element
// 		- Check if an element is present
// 		- Remove an element
// 		- Get the size of the set
// 		- Get the union of two set
// 		- Get the intersection of two set
// 		- Get the difference of two set
// 		- Get the symmetric difference of two set
// 		- Check if two set are equal
// 		- Check if one set is a subset of another
// 		- Check if one set is a superset of another
// 		- Check if one set is a proper subset of another
// 		- Check if one set is a proper superset of another
// 		- Check if two set are disjoint
// 		- Check if two set are equal

type Set struct {
	values map[int]interface{}
}

// Init initializes the set
func (set *Set) Init() *Set {
	set.values = make(map[int]interface{})
	return set
}

// New creates a new set
func New() *Set {
	return new(Set).Init()
}

// Value returns the value of the set as a slice
func (set *Set) Value() []interface{} {
	// initialize a slice
	var value []interface{}
	// iterate the set and extract the values
	for _, v := range set.values {
		// append value to the slice
		value = append(value, v)
	}
	// return value
	return value
}

// Add adds a value to the set.
func (set *Set) Add(value interface{}) {
	// check if value is already present
	// loop and set key to value
	for _, v := range set.values {
		if v == value {
			return
		}
	}
	// set key to value
	set.values[len(set.values)+1] = value
}

// Size returns the size of the set.
func (set *Set) Size() int {
	// return the size of the set
	return len(set.values)
}

// Contains checks if the set contains a value.
func (set *Set) Contains(value interface{}) bool {
	// loop and check if value is present
	for _, v := range set.values {
		if v == value {
			return true
		}
	}
	return false
}

// Remove removes a value from the set.
func (set *Set) Remove(value interface{}) {
	// loop and remove value
	for k, v := range set.values {
		// value is present
		if v == value {
			// delete key
			delete(set.values, k)
		}
	}
}

// Union returns the union of two sets.
func (set *Set) Union(other *Set) *Set {
	// create a new set
	newSet := New()
	// loop and add all values to new set
	for _, v := range set.values {
		newSet.Add(v)
	}
	// loop and add all values to new set
	for _, v := range other.values {
		newSet.Add(v)
	}
	// return new set
	return newSet
}

// Intersection returns the intersection of two sets.
func (set *Set) Intersection(other *Set) *Set {
	// create a new set
	newSet := New()
	// loop and add all values to new set
	for _, v := range set.values {
		if other.Contains(v) {
			newSet.Add(v)
		}
	}
	// return new set
	return newSet
}

// Difference returns the difference of two sets.
func (set *Set) Difference(other *Set) *Set {
	// create a new set
	newSet := New()
	// loop and add all values to new set
	for _, v := range set.values {
		if !other.Contains(v) {
			newSet.Add(v)
		}
	}
	// return new set
	return newSet
}

// SymmetricDifference returns the symmetric difference of two sets.
func (set *Set) SymmetricDifference(other *Set) *Set {
	// create a new set
	newSet := New()
	// loop and add all values to new set
	for _, v := range set.values {
		if !other.Contains(v) {
			newSet.Add(v)
		}
	}
	// loop and add all values to new set
	for _, v := range other.values {
		if !set.Contains(v) {
			newSet.Add(v)
		}
	}
	// return new set
	return newSet
}

// IsSubset checks if the set is a subset of another set.
func (set *Set) IsSubset(other *Set) bool {
	// loop and check if other set contains all values
	for _, v := range set.values {
		if !other.Contains(v) {
			return false
		}
	}
	return true
}

// IsSuperset checks if the set is a superset of another set.
func (set *Set) IsSuperset(other *Set) bool {
	// loop and check if other set contains all values
	for _, v := range other.values {
		if !set.Contains(v) {
			return false
		}
	}
	return true
}

// IsProperSubset checks if the set is a proper subset of another set.
func (set *Set) IsProperSubset(other *Set) bool {
	// check if other set contains all values
	if set.IsSubset(other) {
		// check if other set is not equal to set
		if !set.IsEqual(other) {
			return true
		}
	}
	return false
}

// IsEqual checks if the set is equal to another set.
func (set *Set) IsEqual(other *Set) bool {
	// check if other set contains all values
	if set.IsSubset(other) && set.IsSuperset(other) {
		return true
	}
	return false
}

// IsProperSuperset checks if the set is a proper superset of another set.
func (set *Set) IsProperSuperset(other *Set) bool {
	// check if other set contains all values
	if set.IsSuperset(other) {
		// check if other set is not equal to set
		if !set.IsEqual(other) {
			return true
		}
	}
	return false
}

// IsDisjoint checks if the set is disjoint with another set.
func (set *Set) IsDisjoint(other *Set) bool {
	// loop and check if other set contains any values
	for _, v := range set.values {
		if other.Contains(v) {
			return false
		}
	}
	return true
}
