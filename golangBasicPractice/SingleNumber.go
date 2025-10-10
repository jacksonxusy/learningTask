package golangBasicPractice

import "sort"

func singleNumber(nums []int) int {
	sort.Ints(nums)
	for i := 0; i < len(nums); i = i + 2 {
		if i == len(nums)-1 {
			return nums[i]
		}
		if nums[i] != nums[i+1] {
			return nums[i]
		}
	}
	return 1
}
