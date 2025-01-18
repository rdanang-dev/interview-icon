package main

import (
	"fmt"
	"unicode"
)

func countNumbers(input []string) int {
	count := 0
	for _, item := range input {
		isNumber := true
		for _, char := range item {
			if !unicode.IsDigit(char) {
				isNumber = false
				break
			}
		}
		if isNumber {
			count++
		}
	}
	return count
}

func main() {
	case1 := []string{"2", "h", "6", "u", "y", "t", "7", "j", "y", "h", "8"}
	case2 := []string{"b", "7", "h", "6", "h", "k", "i", "5", "g", "7", "8"}
	case3 := []string{"7", "b", "8", "5", "6", "9", "n", "f", "y", "6", "9"}
	case4 := []string{"u", "h", "b", "n", "7", "6", "5", "1", "g", "7", "9"}

	fmt.Println(countNumbers(case1))
	fmt.Println(countNumbers(case2))
	fmt.Println(countNumbers(case3))
	fmt.Println(countNumbers(case4))
}
