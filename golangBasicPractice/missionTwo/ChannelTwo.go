package main

import (
	"fmt"
	"sync"
)

func main() {
	var results = make(chan int, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			results <- i
			fmt.Println("send: ", i)
		}
		close(results)
	}()

	go func() {
		defer wg.Done()
		for result := range results {
			fmt.Println("receive: ", result)
		}
	}()

	wg.Wait()
	fmt.Println("end")
}
