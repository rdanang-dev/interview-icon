package main

import "fmt"

// findBestBuyPrice finds the buy price that leads to the maximum profit
func findBestBuyPrice(prices []int) int {
	if len(prices) < 2 {
		return 0 // Not enough data to calculate profit
	}

	minPrice := prices[0]     // Track the minimum price (buy price)
	maxProfit := 0            // Track the maximum profit
	bestBuyPrice := prices[0] // Track the buy price that leads to max profit

	for _, price := range prices {
		// Calculate profit if we sell at the current price
		profit := price - minPrice

		// Update maxProfit and bestBuyPrice if this profit is higher
		if profit > maxProfit {
			maxProfit = profit
			bestBuyPrice = minPrice
		}

		// Update minPrice if a lower price is found
		if price < minPrice {
			minPrice = price
		}
	}

	return bestBuyPrice
}

func main() {
	case1 := []int{10, 9, 6, 5, 15}    // Buy at 5, sell at 15 → Best buy price = 5
	case2 := []int{7, 8, 3, 10, 8}     // Buy at 3, sell at 10 → Best buy price = 3
	case3 := []int{5, 12, 11, 12, 10}  // Buy at 5, sell at 12 → Best buy price = 5
	case4 := []int{7, 18, 27, 10, 29}  // Buy at 7, sell at 29 → Best buy price = 7
	case5 := []int{20, 17, 15, 14, 10} // No profit possible → Best buy price = 20

	fmt.Println(findBestBuyPrice(case1)) // Output: 5
	fmt.Println(findBestBuyPrice(case2)) // Output: 3
	fmt.Println(findBestBuyPrice(case3)) // Output: 5
	fmt.Println(findBestBuyPrice(case4)) // Output: 7
	fmt.Println(findBestBuyPrice(case5)) // Output: 20
}
