package stack

type Stack struct {
	elements []interface{}
}

// Init initializes the set
func (stack *Stack) Init() *Stack {
	stack.elements = make([]interface{}, 0)
	return stack
}

func New(elements ...[]interface{}) *Stack {
	if elements == nil || len(elements) < 1 {
		return new(Stack).Init()
	}

	// covert slice  to node
	// Create a new node.
	var element []interface{}

	// Iterate over the list and create a node for each value.
	for _, item := range elements {
		for _, value := range item {
			//append slice to the end of the list
			element = append(element, value)
		}
	}

	// Create a new node.
	return &Stack{element}
}

// Push : Adds a new element to the end of the stack.
func (stack *Stack) Push(element interface{}) {
	// append item to element
	stack.elements = append(stack.elements, element)
}

// Pop : removes the last element from the end of the stack.
func (stack *Stack) Pop() {
	// slice the array to the last item
	stack.elements = stack.elements[:len(stack.elements)-1]
}

// Size returns the size of the stack
func (stack *Stack) Size() int {
	return len(stack.elements)
}

// Reverse : reverses the Stack
func (stack *Stack) Reverse() {
	for i, j := 0, len(stack.elements)-1; i < j; i, j = i+1, j-1 {
		stack.elements[i], stack.elements[j] = stack.elements[j], stack.elements[i]
	}
}

// ToMap : converts the stack to a map
func (stack *Stack) ToMap() map[int]interface{} {
	var mapStack = make(map[int]interface{})
	for i, value := range stack.elements {
		mapStack[i] = value
	}
	return mapStack
}

// ToArray : converts the stack to an array
func (stack *Stack) ToArray() []interface{} {
	return stack.elements
}
