package main

import "time"

func longOperation() (string, error) {
	time.Sleep(3 * time.Second)
	return "OK", nil
}

func main() {

}
