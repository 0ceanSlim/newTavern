package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// API endpoints and API keys
const (
	coingeckoURL      = "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"
	coinmarketcapURL  = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=1&convert=USD"
	blockchainInfoURL = "https://blockchain.info/ticker"
	coinbaseURL       = "https://api.coinbase.com/v2/prices/BTC-USD/spot"
	coinmarketcapKey  = "YOUR_COINMARKETCAP_API_KEY" // Replace with your API key
)

// PriceResponse represents the JSON response for the endpoint
type PriceResponse struct {
	Price string `json:"Price"`
	Error string `json:"error,omitempty"`
}

// FetchBitcoinPrice handles the /api/btc-price endpoint
func FetchBitcoinPrice(w http.ResponseWriter, r *http.Request) {
	var (
		coingeckoPrice      float64
		coinmarketcapPrice  float64
		blockchain15mPrice  float64
		coinbasePrice       float64
		validPrices         []float64
	)

	// Fetch prices from the different APIs
	coingeckoPrice = fetchCoingeckoPrice()
	coinmarketcapPrice = fetchCoinmarketcapPrice()
	blockchain15mPrice = fetchBlockchainInfoPrice()
	coinbasePrice = fetchCoinbasePrice()

	// Collect valid prices
	if coingeckoPrice > 0 {
		validPrices = append(validPrices, coingeckoPrice)
	}
	if coinmarketcapPrice > 0 {
		validPrices = append(validPrices, coinmarketcapPrice)
	}
	if blockchain15mPrice > 0 {
		validPrices = append(validPrices, blockchain15mPrice)
	}
	if coinbasePrice > 0 {
		validPrices = append(validPrices, coinbasePrice)
	}

	// Calculate average price
	var averagePrice float64
	if len(validPrices) > 0 {
		for _, price := range validPrices {
			averagePrice += price
		}
		averagePrice /= float64(len(validPrices))
	}

	// Prepare the response
	response := PriceResponse{}
	if len(validPrices) > 0 {
		response.Price = fmt.Sprintf("%.2f", averagePrice)
	} else {
		response.Error = "Unable to fetch Bitcoin price from all sources."
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions to fetch prices
func fetchCoingeckoPrice() float64 {
	resp, err := http.Get(coingeckoURL)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0
	}
	return data["bitcoin"]["usd"]
}

func fetchCoinmarketcapPrice() float64 {
	req, err := http.NewRequest("GET", coinmarketcapURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}
	req.Header.Set("X-CMC_PRO_API_KEY", coinmarketcapKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//fmt.Printf("CoinMarketCap API returned status: %d\n", resp.StatusCode)
		return 0
	}

	var data map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0
	}

	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return 0
	}

	// Validate nested data structure
	quoteData, ok := data["data"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Invalid 'data' field in CoinMarketCap response")
		return 0
	}

	coinData, ok := quoteData["1"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Invalid '1' field in CoinMarketCap response")
		return 0
	}

	quote, ok := coinData["quote"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Invalid 'quote' field in CoinMarketCap response")
		return 0
	}

	usdData, ok := quote["USD"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Invalid 'USD' field in CoinMarketCap response")
		return 0
	}

	price, ok := usdData["price"].(float64)
	if !ok {
		fmt.Println("Error: Invalid 'price' field in CoinMarketCap response")
		return 0
	}

	return price
}


func fetchBlockchainInfoPrice() float64 {
	resp, err := http.Get(blockchainInfoURL)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0
	}
	return data["USD"]["15m"]
}

func fetchCoinbasePrice() float64 {
	resp, err := http.Get(coinbaseURL)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0
	}
	price, _ := strconv.ParseFloat(data["data"].(map[string]interface{})["amount"].(string), 64)
	return price
}
