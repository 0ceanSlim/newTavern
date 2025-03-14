package main

import (
	"fmt"
	"goFrame/src/api"
	"goFrame/src/routes"
	"goFrame/src/utils"
	"log"
	"net/http"
	"time"
)

func main() {

	// Default config path
	configPath := "config.yml"

	// Load config
	if err := utils.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start monitoring the RTMP stream
	go utils.MonitorStream()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/btc-price", api.FetchBitcoinPrice)
	mux.HandleFunc("/api/btc-price-log", api.ServePriceLogs)
	mux.HandleFunc("/api/gold-price", api.GoldPriceHandler)
	mux.HandleFunc("/create-invoice", api.HandleNostrInvoice)
	mux.HandleFunc("/invoice-events", api.InvoiceEventsHandler)
	mux.HandleFunc("/check-name", api.CheckNameHandler)
	mux.HandleFunc("/check-npub", api.CheckNpubHandler)
	mux.HandleFunc("/api/stream-data", api.GetStreamMetadata)

	// Access-Control-Allow-Origin", "*" for nostr.json
	mux.HandleFunc("/.well-known/nostr.json", utils.ServeWellKnownNostr)

	// Serve .m3u8 file
	mux.HandleFunc("/live/output.m3u8", utils.ServeHLS)

	mux.Handle("/live/", http.StripPrefix("/live/", utils.ServeHLSFolderWithCORS("web/live/")))
	mux.Handle("/.videos/past-streams/", http.StripPrefix("/.videos/past-streams/", utils.ServePastStreamsWithCORS("web/.videos/past-streams/")))

	// Initialize Routes
	routes.InitializeRoutes(mux)

	// Start logging prices as a goroutine with 5 minute interval
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			utils.LogBitcoinPrice()
			<-ticker.C
		}
	}()

	fmt.Printf("Server is running on http://localhost:%d\n", utils.AppConfig.Server.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", utils.AppConfig.Server.Port), mux)
}
