package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// Function to fetch the gold price
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

// Handler for the /gold-price route
func GoldPriceHandler(w http.ResponseWriter, r *http.Request) {
	goldPrice, err := FetchGoldPrice()
	response := GoldPriceResponse{}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Error = err.Error()
	} else {
		response.Price = goldPrice
	}

	// Set Content-Type and send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
