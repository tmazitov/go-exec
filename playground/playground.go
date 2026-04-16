package main

import "fmt"

func main() {
	var i interface{} = "a string"

	valueOf, ok := i.(string)

	fmt.Println(valueOf, ok)
}