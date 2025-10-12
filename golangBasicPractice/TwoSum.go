package main

func twoSum(nums []int, target int) []int {
	var temp = make(map[int]int)
	for i, num := range nums {
		temp[num] = i
	}
	for i := 0; i < len(nums); i++ {
		secondIndex, ok := temp[target-nums[i]]
		if ok && secondIndex != i {
			return []int{secondIndex, i}
		}
	}
	return []int{}
}

func main() {
	twoSum([]int{3, 2, 4}, 6)
}
