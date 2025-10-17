package main

import (
	"fmt"
	"time"
)

func main() {
	results := make(chan int)
	go func() {
		for i := 1; i <= 10; i++ {
			results <- i
		}
		close(results)
	}()
	go func() {
		for result := range results {
			fmt.Println(result)
		}
	}()
	time.Sleep(1 * time.Second)
	fmt.Println("end")
}
