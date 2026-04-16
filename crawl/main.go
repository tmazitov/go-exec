package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Result struct {
	URL  string
	Body []byte
	Err  error
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(rand.Intn(100)) * time.Millisecond):
		if rand.Float32() < 0.2 { // 20% шанс ошибки
			return nil, fmt.Errorf("failed: %s", url)
		}
		return []byte("body of " + url), nil
	}
}

// Create a function that do request to the server with limited count of parallel questions.
// Topic : Semaphore

func crawl(ctx context.Context, urls []string, maxConcurrent int) []Result {

	var wg sync.WaitGroup
	total := make([]Result, len(urls))

	semaphore := make(chan struct{}, maxConcurrent)
	defer close(semaphore)

	for index, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				// слот получен, работаем
			case <-ctx.Done():
				// контекст отменён пока ждали слот
				total[index] = Result{URL: url, Err: ctx.Err()}
				return
			}
			defer func() { <-semaphore }()
			fmt.Println("start working on", index, "request")
			
			body, err := fetch(ctx, url)
			result := Result{
				URL:  url,
				Body: body,
				Err:  err,
			}

			// total = append(total, result)
			total[index] = result

			fmt.Println("end working on", index, "request")
		}()
	}

	wg.Wait()

	return total
}

func main() {

	ctx := context.Background()
	maxConcurrent := 3
	urls := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"12",
		"13",
		"14",
		"15",
		"16",
		"17",
		"18",
	}
	fmt.Println(crawl(ctx, urls, maxConcurrent))
}
