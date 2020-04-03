package main

import (
	"fmt"
)

func B(ch chan<- int) {
	ch <- 2
}

func BB(ch <-chan int) {
	for x := range ch {
		_ = x
	}
}

func main() {
	fmt.Printf("s")
}
