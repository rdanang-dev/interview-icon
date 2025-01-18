package main

import "fmt"

func fibonacci(n int) []int {
	sequence := []int{0, 1}
	for i := 2; i < n; i++ {
		next := sequence[i-1] + sequence[i-2]
		sequence = append(sequence, next)
	}
	return sequence
}

func main() {
	n := 10 // Change to desired length
	result := fibonacci(n)
	fmt.Println(result)
}
