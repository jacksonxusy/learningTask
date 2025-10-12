package main

import "strconv"

func isPalindrome(x int) bool {
	var initial = strconv.Itoa(x)
	var arr = []string{}
	for _, ch := range initial {
		arr = append(arr, string(ch))
	}
	for i := range len(arr) / 2 {
		if arr[i] != arr[len(arr)-i-1] {
			return false
		}
	}
	return true
}
