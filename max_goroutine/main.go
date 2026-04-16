package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var max atomic.Int32
	var wg sync.WaitGroup

	for i := 1000; i >= 0; i-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// if i%2 == 0 && i > int(max.Load()) {
			// 	time.Sleep(time.Second)
			// 	max.Store(int32(i))
			// }
			if i%2 != 0 {
				return
			}
			for {
				cur := max.Load()
				if int32(i) <= cur {
					break
				}

				if max.CompareAndSwap(cur, int32(i)) {
					break
				}
			}
		}()
	}

	wg.Wait()

	fmt.Println("max:", max.Load())
}
