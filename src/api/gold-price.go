package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const cacheFile = "web/logs/last-gold-price.json"

// GoldPriceCache structure
type GoldPriceCache struct {
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Save price to cache
func saveGoldPriceToCache(price string) {
	data := GoldPriceCache{
		Price:     price,
		Timestamp: time.Now(),
	}

	file, err := os.Create(cacheFile)
	if err != nil {
		fmt.Println("Failed to save cache:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
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

// Function to fetch the gold price
func FetchGoldPrice() (string, error) {
	url := "https://www.moneymetals.com/gold-price"

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers to simulate a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.moneymetals.com/")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Execute the request
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch the website: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch the website, status code: %d", resp.StatusCode)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse the HTML: %v", err)
	}

	// Extract gold price using the updated selector
	var goldPrice string
	doc.Find("div.mt-4.md\\:mt-0.md\\:order-last div.text-slate-700.text-lg").Each(func(i int, s *goquery.Selection) {
		// Extract the text content of the div
		text := s.Text()
		if text != "" {
			goldPrice = text
		}
	})

	// Check if the price was found
	if goldPrice == "" {
		return "", fmt.Errorf("gold price not found in the HTML")
	}

	// Remove the dollar sign
	goldPrice = strings.ReplaceAll(goldPrice, "$", "")

	return goldPrice, nil
}

func FetchGoldPriceWithBrowser() (string, error) {
	// Create a context with a timeout for chromedp
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Launch chromedp browser context
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var price string

	// Run chromedp tasks to navigate and extract data
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.apmex.com/gold-price"),
		chromedp.Text(`#metal-priceask`, &price, chromedp.ByID), // Extract text by ID
	)
	if err != nil {
		return "", err
	}

	// Return extracted price
	return price, nil
}

// Response structure for the API
type GoldPriceResponse struct {
	Price string `json:"price"`
	Error string `json:"error,omitempty"`
}

// GoldPriceHandler handles the /gold-price route
func GoldPriceHandler(w http.ResponseWriter, r *http.Request) {
	var goldPrice string
	var err error

	// Try FetchGoldPrice first
	goldPrice, err = FetchGoldPrice()
	if err != nil {
		fmt.Println("FetchGoldPrice failed:", err)

		// Try FetchGoldPriceWithBrowser
		goldPrice, err = FetchGoldPriceWithBrowser()
		if err != nil {
			fmt.Println("FetchGoldPriceWithBrowser failed:", err)

			// Try loading from cache as the last resort
			goldPrice, err = loadGoldPriceFromCache()
			if err != nil {
				fmt.Println("Failed to load price from cache:", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(GoldPriceResponse{Error: "Failed to fetch gold price"})
				return
			}
		}
	}

	// Cache the price if successfully fetched from any method except cache
	if err == nil {
		saveGoldPriceToCache(goldPrice)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GoldPriceResponse{Price: goldPrice})
}
