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

const rsgCacheFile = "web/logs/last-rsg-price.json"
const rsgLogFile = "web/logs/rsg-price-log.csv"

// Bond costs $6.99 USD from Jagex
const bondPriceUSD = 6.99

// OSRS Bond item ID
const bondItemID = 13190

// RSGPriceCache structure for caching
type RSGPriceCache struct {
	BondPriceGP   int64     `json:"bondPriceGP"`
	USDPerMillion float64   `json:"usdPerMillion"`
	GPPerUSD      float64   `json:"gpPerUSD"`
	Timestamp     time.Time `json:"timestamp"`
}

// RuneScape Wiki API response
type RSWikiPriceResponse struct {
	Data map[string]struct {
		High     int64 `json:"high"`
		HighTime int64 `json:"highTime"`
		Low      int64 `json:"low"`
		LowTime  int64 `json:"lowTime"`
	} `json:"data"`
}

// Fetch bond price from RuneScape Wiki API
func fetchBondPrice() (int64, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://prices.runescape.wiki/api/v1/osrs/latest?id=%d", bondItemID), nil)
	if err != nil {
		return 0, err
	}

	// Required: Set a descriptive User-Agent
	req.Header.Set("User-Agent", "HappyTavern-PriceTracker - https://happytavern.co")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var data RSWikiPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	itemIDStr := strconv.Itoa(bondItemID)
	itemData, exists := data.Data[itemIDStr]
	if !exists {
		return 0, fmt.Errorf("bond price not found in response")
	}

	// Use high price as the current market price (what you'd pay to buy instantly)
	if itemData.High > 0 {
		return itemData.High, nil
	}

	// Fallback to low price if high isn't available
	if itemData.Low > 0 {
		return itemData.Low, nil
	}

	return 0, fmt.Errorf("no valid price data")
}

// Calculate USD to RSG conversion rates
func calculateRSGRates(bondPriceGP int64) (usdPerMillion float64, gpPerUSD float64) {
	// Bond costs $6.99 and gives you bondPriceGP gold
	// So $6.99 = bondPriceGP GP
	// $1 USD = bondPriceGP / 6.99 GP
	gpPerUSD = float64(bondPriceGP) / bondPriceUSD

	// USD per million GP
	// If bondPriceGP = 10,000,000 and bond costs $6.99
	// Then 1M GP = $6.99 / (bondPriceGP / 1,000,000)
	usdPerMillion = bondPriceUSD / (float64(bondPriceGP) / 1_000_000)

	return usdPerMillion, gpPerUSD
}

// Save RSG price to cache
func saveRSGPriceToCache(bondPriceGP int64, usdPerMillion, gpPerUSD float64) {
	data := RSGPriceCache{
		BondPriceGP:   bondPriceGP,
		USDPerMillion: usdPerMillion,
		GPPerUSD:      gpPerUSD,
		Timestamp:     time.Now(),
	}

	file, err := os.Create(rsgCacheFile)
	if err != nil {
		fmt.Println("Failed to save RSG price cache:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

// Log RSG price to CSV
func logRSGPriceToCSV(bondPriceGP int64, usdPerMillion, gpPerUSD float64) error {
	// Ensure directory exists
	dir := filepath.Dir(rsgLogFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Open or create CSV file
	file, err := os.OpenFile(rsgLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check if file is empty (write header)
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() == 0 {
		header := []string{"Unix Timestamp", "Human Readable Timestamp", "Bond Price GP", "USD Per Million", "GP Per USD"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write data row
	unixTimestamp := time.Now().Unix()
	humanTimestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	row := []string{
		strconv.FormatInt(unixTimestamp, 10),
		humanTimestamp,
		strconv.FormatInt(bondPriceGP, 10),
		fmt.Sprintf("%.4f", usdPerMillion),
		fmt.Sprintf("%.2f", gpPerUSD),
	}

	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	return nil
}

// LogRSGPrice fetches and logs the RSG price
func LogRSGPrice() {
	bondPriceGP, err := fetchBondPrice()
	if err != nil {
		fmt.Println("Error fetching bond price:", err)
		return
	}

	usdPerMillion, gpPerUSD := calculateRSGRates(bondPriceGP)

	// Save to cache
	saveRSGPriceToCache(bondPriceGP, usdPerMillion, gpPerUSD)

	// Log to CSV
	if err := logRSGPriceToCSV(bondPriceGP, usdPerMillion, gpPerUSD); err != nil {
		fmt.Println("Error logging RSG price to CSV:", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fmt.Printf("[%s] RSG price logged: Bond=%d GP, $%.4f/M, %.2f GP/$\n",
		timestamp, bondPriceGP, usdPerMillion, gpPerUSD)
}
