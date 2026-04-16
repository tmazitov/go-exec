package solutions

import (
	"errors"
	"time"
)

func TimeoutWrapper(timeoutDuration int, action func() (string, error)) (string, error) {

	var (
		resultChan chan string = make(chan string)
		errChan    chan error  = make(chan error)
	)

	// ctx
	// wg (maybe)

	// ctx, cancel := context.WithCancel(context.TODO())
	// defer cancel()

	timer := time.NewTimer(time.Second * time.Duration(timeoutDuration))
	defer timer.Stop()

	// goroutine (action)
	go func() {
		result, err := action()
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return "", err
	case <-timer.C:
		return "", errors.New("timeout error")
	}
}
