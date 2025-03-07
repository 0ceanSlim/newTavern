package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type PriceResponse struct {
	Price string `json:"Price"`
}

func LogBitcoinPrice() {
	logFilePath := "web/logs/btc-price-log.csv"
	apiURL := "http://localhost:8787/api/btc-price"
	blockHeightURL := "https://mempool.happytavern.co/api/blocks/tip/height"

	// Ensure directory and file exist
	if err := ensureFileExists(logFilePath); err != nil {
		fmt.Println("Error ensuring file exists:", err)
		return
	}

	// Get Bitcoin price
	priceResponse, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error fetching price:", err)
		return
	}
	defer priceResponse.Body.Close()

	if priceResponse.StatusCode != http.StatusOK {
		fmt.Println("Error fetching price, status code:", priceResponse.StatusCode)
		return
	}

	var priceData PriceResponse
	if err := json.NewDecoder(priceResponse.Body).Decode(&priceData); err != nil {
		fmt.Println("Error decoding price JSON:", err)
		return
	}
	price := priceData.Price

	// Get block height
	blockHeightResponse, err := http.Get(blockHeightURL)
	if err != nil {
		fmt.Println("Error fetching block height:", err)
		return
	}
	defer blockHeightResponse.Body.Close()

	if blockHeightResponse.StatusCode != http.StatusOK {
		fmt.Println("Error fetching block height, status code:", blockHeightResponse.StatusCode)
		return
	}

	var blockHeight int
	if err := json.NewDecoder(blockHeightResponse.Body).Decode(&blockHeight); err != nil {
		fmt.Println("Error decoding block height JSON:", err)
		return
	}

	// Get timestamps
	unixTimestamp := time.Now().Unix()
	humanReadableTimestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	// Create or append to CSV
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening/creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file stats", err)
		return
	}

	if fileInfo.Size() == 0 {
		if err := writer.Write([]string{"Unix Timestamp", "Human Readable Timestamp", "Block Height", "Price"}); err != nil {
			fmt.Println("Error writing CSV header:", err)
			return
		}
	}

	if err := writer.Write([]string{strconv.FormatInt(unixTimestamp, 10), humanReadableTimestamp, strconv.Itoa(blockHeight), price}); err != nil {
		fmt.Println("Error writing CSV row:", err)
		return
	}

	fmt.Println("Bitcoin price logged successfully to", logFilePath)
}

func ensureFileExists(filePath string) error {
	dir := filepath.Dir(filePath)

	// Create directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
	}

	return nil
}
