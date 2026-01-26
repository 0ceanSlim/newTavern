package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type RSGPriceLogEntry struct {
	UnixTimestamp          int64   `json:"unixTimestamp"`
	HumanReadableTimestamp string  `json:"humanReadableTimestamp"`
	BondPriceGP            int64   `json:"bondPriceGP"`
	USDPerMillion          float64 `json:"usdPerMillion"`
	GPPerUSD               float64 `json:"gpPerUSD"`
}

func ServeRSGPriceLogs(w http.ResponseWriter, r *http.Request) {
	logFilePath := "web/logs/rsg-price-log.csv"

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
	var priceLogs []RSGPriceLogEntry
	if len(records) > 1 { // Skip header row
		for _, record := range records[1:] {
			unixTimestamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				continue // Skip invalid entries
			}
			bondPriceGP, err := strconv.ParseInt(record[2], 10, 64)
			if err != nil {
				continue
			}
			usdPerMillion, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				continue
			}
			gpPerUSD, err := strconv.ParseFloat(record[4], 64)
			if err != nil {
				continue
			}

			priceLogs = append(priceLogs, RSGPriceLogEntry{
				UnixTimestamp:          unixTimestamp,
				HumanReadableTimestamp: record[1],
				BondPriceGP:            bondPriceGP,
				USDPerMillion:          usdPerMillion,
				GPPerUSD:               gpPerUSD,
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
