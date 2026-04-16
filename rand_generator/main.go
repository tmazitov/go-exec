package main

import (
	"bufio"
	"fmt"
	"os"
)

func randNumsGenerator(n int) <-chan int {

	file, err := os.Open("/dev/urandom")
	if err != nil {
		panic("no /dev/urandom file in system")
	}
	reader := bufio.NewReader(file)

	generator := make(chan int)

	go func() {
		defer file.Close()
		defer close(generator)

		var temp []byte = make([]byte, 4)
		var number int
		for range n {

			if _, err := reader.Read(temp); err != nil {
				return
			}

			number = 0

			for index := len(temp) - 1; index >= 0; index-- {
				number |= int(temp[index]) << (8 * (3 - index))
			}

			generator <- number
		}
	}()

	return generator
}

func main() {
	for i := range randNumsGenerator(10) {
		fmt.Println(i)
	}
}
