package main

import (
	"fmt"
	"sync"
	"time"
)

type Task func()

func RunTasks(tasks []Task) {
	var wg sync.WaitGroup
	results := make(chan string, len(tasks))
	for i, task := range tasks {
		wg.Add(1)
		go func(id int, task Task) {
			defer wg.Done()
			start := time.Now()
			task()
			duration := time.Since(start)
			results <- fmt.Sprintf("%s took %s\n", id, duration)
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}

}

func main() {
	tasks := []Task{
		func() {
			time.Sleep(1 * time.Second)
			fmt.Println("task1 finished")
		},
		func() {
			time.Sleep(2 * time.Second)
			fmt.Println("task2 finished")
		},
		func() {
			time.Sleep(3 * time.Second)
			fmt.Println("task3 finished")
		},
	}
	RunTasks(tasks)
}
