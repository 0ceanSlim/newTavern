package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const logFilePath = "web/.logs/btc_price_logs.json"

// LogPrice fetches Bitcoin price from your /api/btc-price endpoint and logs it
func LogPrice() {
	for {
		// Fetch price from your /api/btc-price endpoint
		resp, err := http.Get("http://localhost:8787/api/btc-price") // Adjust if running on a different port or host
		if err != nil {
			fmt.Println("Error fetching price from /api/btc-price:", err)
			time.Sleep(300 * time.Second) // Wait before retrying
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: /api/btc-price returned status code %d\n", resp.StatusCode)
			time.Sleep(300 * time.Second)
			continue
		}

		var priceResponse PriceResponse
		if err := json.NewDecoder(resp.Body).Decode(&priceResponse); err != nil {
			fmt.Println("Error decoding JSON response:", err)
			time.Sleep(300 * time.Second)
			continue
		}

		if priceResponse.Price != "" {
			// Log the price
			entry := map[string]interface{}{
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				"price":     priceResponse.Price,
			}
			file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error opening log file:", err)
				time.Sleep(300 * time.Second)
				continue
			}
			defer file.Close()

			entryJSON, _ := json.Marshal(entry)
			file.Write(append(entryJSON, '\n'))
			fmt.Println("Logged price:", entry)
		} else {
			fmt.Println("No valid price fetched; skipping log.")
		}

		time.Sleep(300 * time.Second) // Log every 5 minutes
	}
}

func ServePriceLogs(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(logFilePath)
	if err != nil {
		http.Error(w, "Unable to read log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var logs []map[string]interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry map[string]interface{}
		json.Unmarshal(scanner.Bytes(), &entry)
		logs = append(logs, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
