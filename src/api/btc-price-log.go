package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type PriceLogEntry struct {
	UnixTimestamp          int64  `json:"unixTimestamp"`
	HumanReadableTimestamp string `json:"humanReadableTimestamp"`
	BlockHeight            int    `json:"blockHeight"`
	Price                  string `json:"price"`
}

func ServePriceLogs(w http.ResponseWriter, r *http.Request) {
	logFilePath := "web/logs/btc-price-log.csv"

	// Open the CSV file
	file, err := os.Open(logFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening log file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read CSV data
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading log file: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert CSV records to JSON
	var priceLogs []PriceLogEntry
	if len(records) > 1 { // Skip header row
		for _, record := range records[1:] {
			unixTimestamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing unix timestamp: %v", err), http.StatusInternalServerError)
				return
			}
			blockHeight, err := strconv.Atoi(record[2])
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing block height: %v", err), http.StatusInternalServerError)
				return
			}

			priceLogs = append(priceLogs, PriceLogEntry{
				UnixTimestamp:          unixTimestamp,
				HumanReadableTimestamp: record[1],
				BlockHeight:            blockHeight,
				Price:                  record[3],
			})
		}
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(priceLogs); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}
