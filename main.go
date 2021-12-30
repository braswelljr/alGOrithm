package main

import (
	"fmt"
	"github.com/braswelljr/alGOrithm/src/data-structures/linear/list"
)

func main() {
	//array := []int{2, 3, 5}
	arr := list.New()
	arr.Push(1)
	//arr.Pop()
	//arr.Join(arrAY)
	//arr.Push(2)
	arr.Insert(6, 0)
	arr.Remove(6)
	//fmt.Printf("%v\n", arr.Next())
	fmt.Printf("%v", arr)
}
