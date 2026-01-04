package main

import (
	"fan_in_merge/solutions"
	"fmt"
	"time"
)

// Merge Channels

// Implement the Merge function that takes a variable number of input channels and returns a single output channel.
// The output channel should receive all values from all input channels.
// Once all input channels are closed and all values have been sent to the output channel, the output channel should be closed.

// Requirements:

//     The function should handle any number of input channels, including zero.
//     All values from all input channels must be sent to the output channel.
//     The output channel should be closed only after all input channels are closed and all values have been forwarded.
//     Your implementation must not leak goroutines.
//     Values should be forwarded as soon as they are available from any input channel.

func generate(id, amount int, result chan int) {
	var generatedValue int

	for index := range amount {
		generatedValue = id*10 + index
		result <- generatedValue
	}

	close(result)
}

func process(id, amount int) chan int {
	var result chan int = make(chan int)

	go generate(id, amount, result)

	return result
}

func fanPrint(outer <-chan int) {

	for {
		item, ok := <-outer
		if !ok {
			break
		}
		fmt.Printf("out : %d\n", item)
	}
}

func main() {

	var generators []<-chan int = make([]<-chan int, 5)

	for index := range generators {
		generators[index] = process(index, len(generators)+index)
	}

	var blocker chan int = make(chan int)
	generators = append(generators, blocker)

	go func() {
		time.Sleep(time.Second * 10)
		close(blocker)
	}()

	fanPrint(solutions.Merge(generators...))
}
