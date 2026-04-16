// Given an array of integers, write a function that returns
// True if the array is sorted in non-decreasing order, and False otherwise.
// Non-decreasing means each element is greater than or equal to the one before it

// [4, 3, 2, 1] -> False
// [1, 2, 2, 3, 4] -> True
// [1, 5, 2, 3] -> False

package main

import "fmt"

func nonDecreaseCheck(array []int) bool {
	// for
	// previous

	if len(array) == 0 {
		return false
	}

	var previousValue int = array[0]

	for _, item := range array[1:] {
		// if index == 0 {
		// 	previousValue = item
		// 	continue
		// }
		if item < previousValue {
			return false
		}
		previousValue = item
	}

	return true
}

func main() {
	test1 := []int{4, 3, 2, 1}
	test2 := []int{1, 2, 2, 3, 4}
	test3 := []int{1, 2, 3, 1, 5, 6}

	fmt.Println("test 1:", nonDecreaseCheck(test1))
	fmt.Println("test 2:", nonDecreaseCheck(test2))
	fmt.Println("test 3:", nonDecreaseCheck(test3))
}

// Monolith
// 1. Auth (Customers, Worker) 
// 2. Users (Customers, Workers, Stocks)
// 3. Delivery
// Database: Postgresql

// 1. reliability  
// 2. easy to scale
// 3. efficient 

// User --gRPC--> Delivery (for statistics)
// 50 000 -->
//   	  -->