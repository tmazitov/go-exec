package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func writerProcess(index int, writerChan chan int) {

	delayTime := time.Microsecond * (time.Duration(rand.Int() % 5)) * 100
	time.Sleep(delayTime)
	writerChan <- index
}

func main() {

	var (
		slice       []int    = []int{}
		writerChan  chan int = make(chan int)
		writerCount int      = 10
		readWg      sync.WaitGroup
		writeWg     sync.WaitGroup
	)

	for index := range writerCount {
		writeWg.Add(1)
		go func() {
			defer writeWg.Done()
			writerProcess(index, writerChan)
		}()
	}
	readWg.Add(1)
	go func() {
		defer readWg.Done()
		for value := range writerChan {
			slice = append(slice, value)
		}
	}()
	writeWg.Wait()
	close(writerChan)
	readWg.Wait()

	fmt.Println("slice:", slice)
}
