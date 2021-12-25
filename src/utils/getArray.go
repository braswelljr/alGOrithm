package utils

import "fmt"

func GetArray() ([]int, int, error) {
	// get array of integers
	var arrayLen int
	fmt.Printf("Please enter the length of the Array : ")
	// read the length of the array by input
	_, err := fmt.Scan(&arrayLen)
	// handle input error
	if err != nil {
		fmt.Printf("\nError : %v", err)
	}

	// create a new slice by using the array length
	array := make([]int, arrayLen)

	// add elements to the slice using length of the array
	for i := 0; i < arrayLen; i++ {
		fmt.Printf("Please enter the element %d : \n", i+1)
		// read the length of the array by input
		_, err := fmt.Scan(&array[i])
		// handle error
		if err != nil {
			fmt.Printf("\nError : %v", err)
		}
	}

	return array, arrayLen, err
}
