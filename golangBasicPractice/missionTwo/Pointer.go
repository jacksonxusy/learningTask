package main

import "fmt"

func main() {
	//var x int = 10
	//fmt.Println(increaseInt(&x))
	var arr = []int{1, 3, 5, 7, 9}
	fmt.Println(multiElement(&arr))

}

func increaseInt(x *int) int {
	*x += 10
	return *x
}

func multiElement(arr *[]int) []int {
	for index, element := range *arr {
		(*arr)[index] = element * 2
	}
	return *arr
}
