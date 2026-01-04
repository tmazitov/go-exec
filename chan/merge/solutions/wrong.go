package solutions

// It's wrong because of reasons below:

// 1. Imagine you have 100 channels and you just write everything to the result channel by one goroutine.
// It will take to much time to handle everything for one goroutine.

// 2. Imagine one of the input channels receive values faster than others.
// After sometime it will cause a block for the goroutine that tries to write to this blocked channel.

// 3. This solution blocking on empty channels.
// It might cause some issues in real projects.

// WrongMerge is an unreflective but executable solution for this task.
// Works only by one inner goroutine and blocks on empty channels.
func WrongMerge(channels ...<-chan int) <-chan int {

	var result chan int = make(chan int)

	go func() {
		defer close(result)

		var closed int

		for {
			closed = 0

			for _, ch := range channels {
				value, ok := <-ch
				if !ok {
					closed += 1
					continue
				}
				result <- value
			}

			if closed == len(channels) {
				break
			}
		}
	}()

	return result
}
