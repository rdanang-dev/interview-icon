package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// Define structures to match the JSON response
type Booking struct {
	ID              string        `json:"id"`
	BookingDate     string        `json:"bookingDate"`
	OfficeName      string        `json:"officeName"`
	StartTime       string        `json:"startTime"`
	EndTime         string        `json:"endTime"`
	ListConsumption []Consumption `json:"listConsumption"`
	Participants    int           `json:"participants"`
	RoomName        string        `json:"roomName"`
}

type Consumption struct {
	Name string `json:"name"`
}

type MasterJenisKonsumsi struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MaxPrice int    `json:"maxPrice"`
}

type GroupedBooking struct {
	RoomName string      `json:"roomName"`
	Bookings []GroupedBy `json:"bookings"`
}

type GroupedBy struct {
	DateRange           string              `json:"dateRange"`
	Participants        int                 `json:"participants"`
	TotalConsumptionFee int                 `json:"total_consumption_fee"`
	ConsumptionDetails  []ConsumptionDetail `json:"consumption"`
}

type ConsumptionDetail struct {
	Name         string `json:"consumption_name"`
	Price        int    `json:"price"`
	Participants int    `json:"participants"`
	TotalPrice   int    `json:"total_price"`
}

// Reusable function to fetch data from an API
func fetchData(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// Group bookings by office name and date range
func groupBookings(bookings []Booking, konsumsiPrices map[string]int) []GroupedBooking {
	grouped := make(map[string]map[string]*GroupedBy)

	for _, booking := range bookings {
		if _, exists := grouped[booking.RoomName]; !exists {
			grouped[booking.RoomName] = make(map[string]*GroupedBy)
		}

		dateRange := fmt.Sprintf("%s - %s", booking.BookingDate, booking.BookingDate)
		if _, exists := grouped[booking.RoomName][dateRange]; !exists {
			grouped[booking.RoomName][dateRange] = &GroupedBy{
				DateRange:           dateRange,
				Participants:        0,
				TotalConsumptionFee: 0,
				ConsumptionDetails:  []ConsumptionDetail{},
			}
		}

		entry := grouped[booking.RoomName][dateRange]
		entry.Participants += booking.Participants

		for _, consumption := range booking.ListConsumption {
			price := konsumsiPrices[consumption.Name]
			totalPrice := price * booking.Participants
			entry.TotalConsumptionFee += totalPrice
			entry.ConsumptionDetails = append(entry.ConsumptionDetails, ConsumptionDetail{
				Name:         consumption.Name,
				Price:        price,
				Participants: booking.Participants,
				TotalPrice:   totalPrice,
			})
		}
	}

	// Convert map to slice
	var result []GroupedBooking
	for roomName, dateGroups := range grouped {
		var bookings []GroupedBy
		for _, entry := range dateGroups {
			sort.Slice(entry.ConsumptionDetails, func(i, j int) bool {
				return entry.ConsumptionDetails[i].Name < entry.ConsumptionDetails[j].Name
			})
			bookings = append(bookings, *entry)
		}
		result = append(result, GroupedBooking{
			RoomName: roomName, // Use RoomName here
			Bookings: bookings,
		})
	}

	return result
}

func main() {
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		bookingURL := "https://66876cc30bc7155dc017a662.mockapi.io/api/dummy-data/bookingList"
		konsumsiURL := "https://6686cb5583c983911b03a7f3.mockapi.io/api/dummy-data/masterJenisKonsumsi"

		// Fetch bookings
		var bookings []Booking
		if err := fetchData(bookingURL, &bookings); err != nil {
			http.Error(w, "Error fetching booking data", http.StatusInternalServerError)
			return
		}

		// Fetch konsumsi data
		var konsumsi []MasterJenisKonsumsi
		if err := fetchData(konsumsiURL, &konsumsi); err != nil {
			http.Error(w, "Error fetching konsumsi data", http.StatusInternalServerError)
			return
		}

		// Create a map for konsumsi prices
		konsumsiPrices := make(map[string]int)
		for _, item := range konsumsi {
			konsumsiPrices[item.Name] = item.MaxPrice
		}

		// Group bookings
		groupedBookings := groupBookings(bookings, konsumsiPrices)

		// Prepare and send the response
		response := map[string]interface{}{
			"statusCode": 200,
			"message":    "dashboard data fetched successfully",
			"data":       groupedBookings,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
