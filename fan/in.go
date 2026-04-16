package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
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

func joinChannels(chs ...<-chan int) <-chan int {

	var result chan int = make(chan int)

	var wg sync.WaitGroup

	for _, channel := range chs {
		wg.Add(1)
		go func(ch <-chan int) {
			defer wg.Done()
			for number := range ch {
				result <- number
			}
		}(channel)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func tester(numberChan chan<- int) {
	go func() {
		for number := range randNumsGenerator(10) {
			numberChan <- number
		}
		close(numberChan)
	}()
}

func main() {

	testerCount := 10

	var channels []chan int = make([]chan int, 10)
	for i := range testerCount {
		channels[i] = make(chan int)
		go tester(channels[i])
	}

	readOnlyChannels := make([]<-chan int, len(channels))
	for i, ch := range channels {
		readOnlyChannels[i] = ch
	}

	counter := 0
	for element := range joinChannels(readOnlyChannels...) {
		counter++
		fmt.Println(element)
	}

	fmt.Println("total:", counter)
}
