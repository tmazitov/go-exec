package main

import (
	"fmt"
	"slices"
)

func rotate(matrix [][]int) {
	rows := len(matrix)
	cols := len(matrix[0])

	for i := range rows {
		for j := i; j < cols; j++ {
			temp := matrix[i][j]
			matrix[i][j] = matrix[j][i]
			matrix[j][i] = temp
		}
	}

	for k := range rows {
		slices.Reverse(matrix[k])
	}
}

func main() {
	testCase1 := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	testCase2 := [][]int{
		{5, 1, 9, 11},
		{2, 4, 8, 10},
		{13, 3, 6, 7},
		{15, 14, 12, 16},
	}

	rotate(testCase1)
	rotate(testCase2)

	fmt.Println("test 1: ", testCase1)
	fmt.Println("test 2: ", testCase2)
}
