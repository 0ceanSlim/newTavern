package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const cacheFile = "web/logs/last-gold-price.json"

// GoldPriceCache structure
type GoldPriceCache struct {
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Load price from cache
func loadGoldPriceFromCache() (string, error) {
	file, err := os.Open(cacheFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var data GoldPriceCache
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return "", err
	}

	return data.Price, nil
}

// Response structure for the API
type GoldPriceResponse struct {
	Price string `json:"price"`
	Error string `json:"error,omitempty"`
}

// GoldPriceHandler serves the cached gold price (updated by background scraper)
func GoldPriceHandler(w http.ResponseWriter, r *http.Request) {
	// Load and serve the cached price
	goldPrice, err := loadGoldPriceFromCache()
	if err != nil {
		fmt.Println("Failed to load gold price from cache:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GoldPriceResponse{Error: "Gold price not available"})
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GoldPriceResponse{Price: goldPrice})
}
