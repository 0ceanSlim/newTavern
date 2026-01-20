package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

const goldCacheFile = "web/logs/last-gold-price.json"

// GoldPriceCache structure
type GoldPriceCache struct {
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Fetch from gold-api.com (primary source)
func fetchGoldAPIPrice() (float64, error) {
	resp, err := http.Get("https://api.gold-api.com/price/XAU")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var data struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.Price, nil
}

// Fetch from goldprice.org (backup source)
func fetchGoldPriceOrgPrice() (float64, error) {
	resp, err := http.Get("https://data-asg.goldprice.org/dbXRates/USD")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var data struct {
		Items []struct {
			XAUPrice float64 `json:"xauPrice"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if len(data.Items) == 0 {
		return 0, fmt.Errorf("no price data")
	}

	return data.Items[0].XAUPrice, nil
}

// Fetch from xe.com (fallback source)
func fetchXEPrice() (float64, error) {
	resp, err := http.Get("https://www.xe.com/currencyconverter/convert/?Amount=1&From=XAU&To=USD")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Extract XAU rate from embedded JSON
	re := regexp.MustCompile(`"XAU":([0-9.]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return 0, fmt.Errorf("XAU rate not found")
	}

	var xauRate float64
	if _, err := fmt.Sscanf(matches[1], "%f", &xauRate); err != nil {
		return 0, err
	}

	// Convert rate to price (1 / rate)
	return 1 / xauRate, nil
}

// FetchGoldPrice tries multiple sources with fallback
func fetchGoldPrice() (string, error) {
	var price float64
	var err error

	// Try gold-api.com first
	price, err = fetchGoldAPIPrice()
	if err != nil {
		fmt.Println("gold-api.com failed:", err)

		// Try goldprice.org
		price, err = fetchGoldPriceOrgPrice()
		if err != nil {
			fmt.Println("goldprice.org failed:", err)

			// Try xe.com as last resort
			price, err = fetchXEPrice()
			if err != nil {
				fmt.Println("xe.com failed:", err)
				return "", fmt.Errorf("all gold price sources failed")
			}
		}
	}

	return fmt.Sprintf("%.2f", price), nil
}

// Save price to cache
func saveGoldPriceToCache(price string) {
	data := GoldPriceCache{
		Price:     price,
		Timestamp: time.Now(),
	}

	file, err := os.Create(goldCacheFile)
	if err != nil {
		fmt.Println("Failed to save gold price cache:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

func LogGoldPrice() {
	// Fetch gold price
	goldPrice, err := fetchGoldPrice()
	if err != nil {
		fmt.Println("Error fetching gold price:", err)
		return
	}

	// Save to cache
	saveGoldPriceToCache(goldPrice)

	// Log timestamp for debugging
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fmt.Printf("[%s] Gold price logged successfully: $%s\n", timestamp, goldPrice)
}
