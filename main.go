package main

import (
	"fmt"
	"github.com/braswelljr/alGOrithm/src/data-structures/linear/list"
	"github.com/braswelljr/alGOrithm/src/utils"
)

func main() {
	array, _, err := utils.GetArray()
	if err != nil {
		panic(err)
	}

	array2 := []string{"Happy", "New", "Year"}

	arrAy, _ := utils.TypeToInterface(array)
	arrAY, _ := utils.TypeToInterface(array2)
	arr := list.New(arrAy)
	arr.Push(1)
	arr.Pop()
	arr.Push(2)
	fmt.Printf("%v", arr)
	fmt.Printf("%v", arrAY)
}
