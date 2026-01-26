package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const rsgCacheFile = "web/logs/last-rsg-price.json"

// Bond costs $6.99 USD from Jagex
const bondPriceUSD = 6.99

// OSRS Bond item ID
const bondItemID = 13190

// RSGPriceCache structure
type RSGPriceCache struct {
	BondPriceGP    int64     `json:"bondPriceGP"`    // Bond price in GP
	USDPerMillion  float64   `json:"usdPerMillion"`  // USD per million GP
	GPPerUSD       float64   `json:"gpPerUSD"`       // GP you get per $1 USD
	Timestamp      time.Time `json:"timestamp"`
}

// RSGPriceResponse for the API
type RSGPriceResponse struct {
	BondPriceGP   int64   `json:"bondPriceGP"`
	USDPerMillion float64 `json:"usdPerMillion"`
	GPPerUSD      float64 `json:"gpPerUSD"`
	Error         string  `json:"error,omitempty"`
}

// Load RSG price from cache
func loadRSGPriceFromCache() (*RSGPriceCache, error) {
	file, err := os.Open(rsgCacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data RSGPriceCache
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

// RSGPriceHandler serves the cached RSG price
func RSGPriceHandler(w http.ResponseWriter, r *http.Request) {
	data, err := loadRSGPriceFromCache()
	if err != nil {
		fmt.Println("Failed to load RSG price from cache:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RSGPriceResponse{Error: "RSG price not available"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RSGPriceResponse{
		BondPriceGP:   data.BondPriceGP,
		USDPerMillion: data.USDPerMillion,
		GPPerUSD:      data.GPPerUSD,
	})
}
