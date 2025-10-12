package main

import "fmt"

func isValid(s string) bool {
	var result = map[string]string{
		"(": ")",
		"[": "]",
		"{": "}",
	}
	closeBrackets := []string{}

	for _, ch := range s {
		if string(ch) == "(" || string(ch) == "[" || string(ch) == "{" {
			closeBrackets = append(closeBrackets, string(ch))

		}
		if string(ch) == ")" || string(ch) == "]" || string(ch) == "}" {
			if len(closeBrackets) == 0 {
				return false
			}
			if string(ch) != result[closeBrackets[len(closeBrackets)-1]] {
				return false
			}
			closeBrackets = closeBrackets[:len(closeBrackets)-1]
		}
	}
	if len(closeBrackets) > 0 {
		return false
	}

	return true
}

func main() {
	fmt.Println(isValid("["))
}
