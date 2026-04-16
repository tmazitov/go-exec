package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Make a function with timeout that replies with string with delay from goroutine

// 1) Using time.After chanel
// ! Don't use an atomic flag with closed chanel meaning.
// ! Problem : goroutine leak.
// ! Problem : timers take a CPU from runtime.
// func fetchWithTimeout(timeout time.Duration) (string, error) {

// 	ch := make(chan string, 1)

// 	go func() {
// 		time.Sleep(time.Second * 3)
// 		ch <- "test" // can't write to the closed channel
// 	}()

// 	select {
// 	case v := <-ch:
// 		return v, nil
// 	case <-time.After(timeout):
// 		return "", errors.New("timeout error")
// 	}
// }

// 2) Using context.WithTimeout
func fetchWithTimeout(timeout time.Duration) (string, error) {
	ch := make(chan string, 1)

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	go func() {
		select {
		// Mock task execution process that takes 3 seconds
		case <-time.After(time.Second * 3):
			ch <- "test"
		case <-ctx.Done():
			return
		}
	}()

	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		return "", errors.New("timeout error")
	}
}

func main() {

	result, err := fetchWithTimeout(time.Second * 4)

	time.Sleep(time.Second * 2)

	fmt.Println("result:", result, err)

}
