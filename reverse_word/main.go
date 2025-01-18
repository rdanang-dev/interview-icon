package main

import (
	"fmt"
	"strings"
)

func reverseWords(input string) string {
	words := strings.Split(input, " ")
	for i, word := range words {
		runes := []rune(word)
		for j, k := 0, len(runes)-1; j < k; j, k = j+1, k-1 {
			runes[j], runes[k] = runes[k], runes[j]
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func main() {
	input := "italem irad irigayaj iadab itsap ulalreb nalub kusutret gnalali"
	output := reverseWords(input)
	fmt.Println(output)
}
