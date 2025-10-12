package main

import (
	"sort"
)

func merge(intervals [][]int) [][]int {
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	if len(intervals) == 1 {
		return intervals
	}
	var resultTemp []int
	resultTemp = append(resultTemp, intervals[0][0])
	resultTemp = append(resultTemp, intervals[0][1])
	for i := 1; i < len(intervals); i++ {
		if intervals[i][0] <= resultTemp[len(resultTemp)-1] {
			if intervals[i][1] >= resultTemp[len(resultTemp)-1] {
				resultTemp = resultTemp[:len(resultTemp)-1]
				resultTemp = append(resultTemp, intervals[i][1])
			}

		} else {
			resultTemp = append(resultTemp, intervals[i][0])
			resultTemp = append(resultTemp, intervals[i][1])
		}

	}
	var result [][]int
	for i := 0; i < len(resultTemp); i = i + 2 {
		result = append(result, []int{resultTemp[i], resultTemp[i+1]})
	}
	return result

}

func main() {
	var temp = [][]int{
		{2, 3},
		{4, 5},
		{1, 10},
	}
	merge(temp)
}
