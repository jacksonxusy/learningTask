package main

import (
	"fmt"
	"strconv"
	"time"
)

func printOddNumber() {
	for i := 0; i < 10; i++ {
		if i%2 != 0 {
			fmt.Println("odd: " + strconv.Itoa(i))
		}
	}
}

func printEvenNumber() {
	for i := 0; i < 10; i++ {
		if i%2 != 0 {
			fmt.Println("even: " + strconv.Itoa(i))
		}
	}
}

func main() {
	go printOddNumber()
	go printEvenNumber()
	time.Sleep(1 * time.Second)
	fmt.Println("end")
}
