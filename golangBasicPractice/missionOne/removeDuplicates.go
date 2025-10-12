package main

func removeDuplicates(nums []int) int {
	var sortIndex = 0
	if len(nums) == 1 {
		return 1
	}
	for i := 0; i < len(nums)-1; i++ {
		if nums[i] != nums[i+1] {
			sortIndex++
			nums[sortIndex] = nums[i+1]
		}
	}
	return sortIndex
}
