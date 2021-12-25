package main

import (
	"fmt"
	"github.com/braswelljr/alGOrithm/src/utils"
)

func main() {
	array, length, err := utils.GetArray()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Array: %v\n -> length of %v", array, length)
}
