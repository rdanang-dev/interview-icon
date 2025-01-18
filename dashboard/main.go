package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"
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
	RoomName            string              `json:"roomName"`
	MonthStamp          string              `json:"monthStamp"`
	BookPercentage      float32             `json:"bookPercentage"`
	BookDetails         string              `json:"bookDetails"` // New field for booked days details
	TotalConsumptionFee int                 `json:"totalConsumptionFee"`
	ConsumptionDetails  []ConsumptionDetail `json:"consumptionDetails"` // New field for total consumption details
	Bookings            []GroupedBy         `json:"bookings"`
}

type GroupedBy struct {
	BookingDate         string              `json:"bookingDate"`
	Participants        int                 `json:"participants"`
	DailyConsumptionFee int                 `json:"dailyConsumptionFee"`
	ConsumptionDetails  []ConsumptionDetail `json:"consumption"`
}

type ConsumptionDetail struct {
	Name         string `json:"consumption_name"`
	Price        int    `json:"price"`
	Participants int    `json:"participants"`
	TotalPrice   int    `json:"total_price"`
}

// Fetch data from an API
func fetchData(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching data: received status code %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// Parse and validate MM-YYYY format
func parseMonthYear(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("Please insert param MM-YYYY")
	}

	match, _ := regexp.MatchString(`^(0[1-9]|1[0-2])-\d{4}$`, input)
	if !match {
		return "", fmt.Errorf("'month' parameter must be in MM-YYYY format")
	}
	return input, nil
}

func groupBookings(bookings []Booking, konsumsiPrices map[string]int, month string) []GroupedBooking {
	grouped := make(map[string]map[string]*GroupedBy)
	// Extract the number of days in the month
	monthTime, _ := time.Parse("01-2006", month)
	daysInMonth := daysInMonth(monthTime)

	// First, group bookings by room and date
	for _, booking := range bookings {
		// Parse BookingDate into time.Time
		bookingTime, err := time.Parse(time.RFC3339, booking.BookingDate)
		if err != nil {
			// fmt.Printf("Skipping booking: %s (Invalid BookingDate format: %s)\n", booking.ID, booking.BookingDate)
			continue
		}

		// Extract booking month in MM-YYYY format
		bookingMonth := bookingTime.Format("01-2006")
		if bookingMonth != month {
			// fmt.Printf("Skipping booking: %s (BookingDate: %s does not match month %s)\n", booking.ID, bookingMonth, month)
			continue
		}

		// Extract the booking date (YYYY-MM-DD)
		formattedDate := bookingTime.Format("2006-01-02")

		if _, exists := grouped[booking.RoomName]; !exists {
			grouped[booking.RoomName] = make(map[string]*GroupedBy)
		}

		if _, exists := grouped[booking.RoomName][formattedDate]; !exists {
			grouped[booking.RoomName][formattedDate] = &GroupedBy{
				BookingDate:         formattedDate,
				Participants:        0,
				DailyConsumptionFee: 0,
				ConsumptionDetails:  []ConsumptionDetail{},
			}
		}

		entry := grouped[booking.RoomName][formattedDate]
		entry.Participants += booking.Participants

		for _, consumption := range booking.ListConsumption {
			price := konsumsiPrices[consumption.Name]
			totalPrice := price * booking.Participants
			entry.DailyConsumptionFee += totalPrice
			entry.ConsumptionDetails = append(entry.ConsumptionDetails, ConsumptionDetail{
				Name:         consumption.Name,
				Price:        price,
				Participants: booking.Participants,
				TotalPrice:   totalPrice,
			})
		}

		// Debug: log each booking that is grouped
		// fmt.Printf("Booking grouped: %s, Date: %s, Room: %s\n", booking.ID, formattedDate, booking.RoomName)
	}

	// Now create the final result with the additional `ConsumptionDetails` field at the room level
	var result []GroupedBooking
	for roomName, dateGroups := range grouped {
		var bookings []GroupedBy
		totalConsumptionFee := 0
		consumptionDetailsMap := make(map[string]ConsumptionDetail)
		var bookedDays int

		for _, entry := range dateGroups {
			// Sort consumption details for each booking
			sort.Slice(entry.ConsumptionDetails, func(i, j int) bool {
				return entry.ConsumptionDetails[i].Name < entry.ConsumptionDetails[j].Name
			})

			// Count the number of booked days
			bookedDays++

			// Sum the daily consumption details
			for _, detail := range entry.ConsumptionDetails {
				if existingDetail, exists := consumptionDetailsMap[detail.Name]; exists {
					// Update existing consumption detail
					existingDetail.TotalPrice += detail.TotalPrice
					existingDetail.Participants += detail.Participants
					consumptionDetailsMap[detail.Name] = existingDetail
				} else {
					// Add new consumption detail
					consumptionDetailsMap[detail.Name] = detail
				}
			}

			bookings = append(bookings, *entry)
			totalConsumptionFee += entry.DailyConsumptionFee
		}

		// Calculate the percentage usage
		bookPercentage := float32(bookedDays) / float32(daysInMonth) * 100
		bookDetails := fmt.Sprintf("Booked %d of %d days", bookedDays, daysInMonth)

		// Convert the map of consumption details to a slice
		var consumptionDetails []ConsumptionDetail
		for _, detail := range consumptionDetailsMap {
			consumptionDetails = append(consumptionDetails, detail)
		}

		// Debug: log the result of grouping for each room
		// fmt.Printf("Room: %s, TotalConsumptionFee: %d, Month: %s\n", roomName, totalConsumptionFee, month)

		// Add the final GroupedBooking for the room
		result = append(result, GroupedBooking{
			RoomName:            roomName,
			TotalConsumptionFee: totalConsumptionFee,
			MonthStamp:          month,
			ConsumptionDetails:  consumptionDetails, // Include total consumption details
			BookPercentage:      bookPercentage,     // Include usage percentage
			BookDetails:         bookDetails,        // Include booked days details
			Bookings:            bookings,
		})
	}

	return result
}

// Function to get the number of days in a month
func daysInMonth(t time.Time) int {
	// Get the first day of the month
	firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Get the first day of the next month
	nextMonth := firstDay.AddDate(0, 1, 0)
	// Subtract 1 day to get the last day of the current month
	lastDay := nextMonth.AddDate(0, 0, -1)
	// Return the day of the month of the last day
	return lastDay.Day()
}

func main() {
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		// Get the 'month' parameter from the URL query string
		month := r.URL.Query().Get("month")

		// println("month", month)
		// Check if the 'month' parameter is empty
		if month == "" {
			// If no 'month' parameter, return an error response
			errorResponse := map[string]interface{}{
				"statusCode": http.StatusBadRequest,
				"message":    "Please provide a valid 'month' parameter in MM-YYYY format ex: (?month=01-2024)",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
				http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If 'month' parameter is provided, validate it
		_, err := parseMonthYear(month)
		if err != nil {
			// If validation fails, return the error message
			errorResponse := map[string]interface{}{
				"statusCode": http.StatusBadRequest,
				"message":    err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
				http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		bookingURL := "https://66876cc30bc7155dc017a662.mockapi.io/api/dummy-data/bookingList"
		konsumsiURL := "https://6686cb5583c983911b03a7f3.mockapi.io/api/dummy-data/masterJenisKonsumsi"

		// Debug: log that we are fetching data
		// fmt.Println("Fetching booking data from:", bookingURL)
		var bookings []Booking
		if err := fetchData(bookingURL, &bookings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Debug: log the number of bookings retrieved
		// fmt.Printf("Bookings fetched: %d\n", len(bookings))

		// Fetch konsumsi prices
		// Debug: log that we are fetching konsumsi data
		// fmt.Println("Fetching konsumsi data from:", konsumsiURL)
		var konsumsiPrices []MasterJenisKonsumsi
		if err := fetchData(konsumsiURL, &konsumsiPrices); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build the map of konsumsi prices by name
		konsumsiMap := make(map[string]int)
		for _, item := range konsumsiPrices {
			konsumsiMap[item.Name] = item.MaxPrice
		}

		// Group the bookings and calculate the additional fields
		groupedBookings := groupBookings(bookings, konsumsiMap, month)

		response := map[string]interface{}{
			"statusCode": 200,
			"message":    "Dashboard data fetched successfully",
			"data":       groupedBookings,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
