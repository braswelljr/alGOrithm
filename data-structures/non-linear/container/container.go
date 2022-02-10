package container

// Contain interface
type Contain interface {
	// Add an element to the container
	Add(element interface{})

	// Remove an element from the container
	Remove(element interface{})

	// Contains Check if the container contains an element
	Contains(element interface{}) bool

	// Size Get the size of the container
	Size() int

	// Clear the container
	Clear()
}

// Container struct
type Container struct {
	elements []interface{}
}

// Init initializes a new Table with a root node
func (container *Container) Init() *Container {
	// Initialize the container
	return container
}

// New creates a new table instance
func New() *Container {
	// Create a new Container
	return new(Container).Init()
}

// Add an element to the @Container
func (container *Container) Add(element interface{}) {
	// Append the element to the container
	container.elements = append(container.elements, element)
}

// Remove an element from the container
func (container *Container) Remove(element interface{}) {
	for i, e := range container.elements {
		if e == element {
			// Slice rest of container elements and append with the sliced part with the index of element removed
			container.elements = append(container.elements[:i], container.elements[i+1:]...)
			// break from loop
			break
		}
	}
}

// Size returns the number of element in the @Container
func (container *Container) Size() int {
	// return the length of container
	return len(container.elements)
}

// Contains returns true if element is found and false if not
func (container *Container) Contains(element interface{}) bool {
	// loop the container
	for _, e := range container.elements {
		// if the element is found return true
		if e == element {
			return true
		}
	}
	// return false if otherwise
	return false
}

// Clear the container
func (container *Container) Clear() {
	// clear the container by returning an empty container
	container.elements = []interface{}{}
}
