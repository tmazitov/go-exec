package main

import (
	"fmt"
	"math/big"
)

// Memory concept (140 bit)
// 1. Square bits part: (30)
// -	[0-9] - zero temp square
// -	[0-9] - first temp square
// -	[0-9] - second temp square
// 2. Row bits part (10)
// -	[0-9] - just one row
// 2. Column bits part (100)
// -	[0-9] - zero temp column
// -	[0-9] - second temp column
// -	....
// -	[0-9] - ninth temp column

var squareBitsCount = 30

func isValidSudoku(board [][]byte) bool {
	var (
		memory big.Int
	)

	for rowNumber, row := range board {

		
		rowIndex := rowNumber % 3
		if rowIndex == 0 {
			for i := range squareBitsCount {
				memory.SetBit(&memory, i, 0)
			}
		}
		
		for i := range 10 {
			memory.SetBit(&memory, squareBitsCount+i, 0)
		}

		for itemNumber, item := range row {

			if item == '.' {
				continue
			}

			index := int(item - '0')
			squareIndex := itemNumber / 3
			// Row and Column check
			if index < 0 || index > 9 || // valid check
			memory.Bit(squareBitsCount+index) != 0 || // row check
			memory.Bit(squareIndex*10+index) != 0 || // square check
			memory.Bit((itemNumber+4)*10+index) != 0 { // column check
				fmt.Println(rowIndex, itemNumber, squareIndex, memory.Bit(squareIndex*10+index) != 0, memory.Bit((itemNumber+3)*10+index) != 0)
				return false
			}

			memory.SetBit(&memory, squareBitsCount+index, 1)
			memory.SetBit(&memory, squareIndex*10+index, 1)
			memory.SetBit(&memory, (itemNumber+4)*10+index, 1)
		}

	}
	return true
}

func main() {

	testCase1 := [][]byte{
		{'5', '3', '.', '.', '7', '.', '.', '.', '.'},
		{'6', '.', '.', '1', '9', '5', '.', '.', '.'},
		{'.', '9', '8', '.', '.', '.', '.', '6', '.'},
		{'8', '.', '.', '.', '6', '.', '.', '.', '3'},
		{'4', '.', '.', '8', '.', '3', '.', '.', '1'},
		{'7', '.', '.', '.', '2', '.', '.', '.', '6'},
		{'.', '6', '.', '.', '.', '.', '2', '8', '.'},
		{'.', '.', '.', '4', '1', '9', '.', '.', '5'},
		{'.', '.', '.', '.', '8', '.', '.', '7', '9'}}

	testCase2 := [][]byte{
		{'.', '.', '.', '.', '5', '.', '.', '1', '.'},
		{'.', '4', '.', '3', '.', '.', '.', '.', '.'},
		{'.', '.', '.', '.', '.', '2', '.', '.', '1'},
		{'8', '.', '.', '.', '.', '.', '.', '2', '.'},
		{'.', '.', '2', '.', '7', '.', '.', '.', '.'},
		{'.', '1', '5', '.', '.', '.', '.', '.', '.'},
		{'.', '.', '.', '.', '.', '2', '.', '.', '.'},
		{'.', '2', '.', '9', '.', '.', '.', '.', '.'},
		{'.', '.', '4', '.', '.', '.', '.', '.', '.'},
	}

	result1 := isValidSudoku(testCase1)
	result2 := isValidSudoku(testCase2)
	fmt.Println("valid1:", result1)
	fmt.Println("valid2:", result2)
}
