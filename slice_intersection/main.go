package main

import (
	"fmt"
	"slice_intersection/solutions"
)

func main() {
	var (
		sliceX = []int{1, 3, 4, 8}
		sliceY = []int{2, 6, 1, 2, 4, 8}
	)

	fmt.Println("intersection: ", solutions.Intersection(sliceX, sliceY))
}
