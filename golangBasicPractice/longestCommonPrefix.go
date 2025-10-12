package main

func longestCommonPrefix(strs []string) string {
	var first = strs[0]
	if len(strs) == 1 {
		return first
	}

	if len(strs) > 1 {
		for j := range len(first) {
			for i := 1; i < len(strs); i++ {
				if len(strs[i]) == j {
					return first[:j]
				}

				if strs[i][j] != first[j] {
					return first[:j]
				}
			}
			if j == len(first)-1 {
				return first
			}
		}

	}
	return ""

}

func main() {
	longestCommonPrefix([]string{"a"})
}
