package solutions

import "sync"

func Merge(channels ...<-chan int) <-chan int {
	var result chan int = make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func() {
			for {
				value, ok := <-ch
				if !ok {
					break
				}

				result <- value
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
