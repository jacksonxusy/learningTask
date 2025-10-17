package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var res = 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			for j := 0; j < 1000; j++ {
				res++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(res)
}
