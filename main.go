package main

import (
	"fmt"
	"goFrame/src/api"
	"goFrame/src/routes"
	"goFrame/src/utils"
	"net/http"
)

func main() {
	// Load Configurations
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Start monitoring the RTMP stream
	go utils.MonitorStream()

	go api.LogPrice() // Start logging prices

	mux := http.NewServeMux()

	mux.HandleFunc("/api/btc-price", api.FetchBitcoinPrice)
	mux.HandleFunc("/api/price-logs", api.ServePriceLogs)

	// Access-Control-Allow-Origin", "*" for nostr.json
	mux.HandleFunc("/.well-known/nostr.json", utils.ServeWellKnownNostr)

	// Serve .m3u8 file
	mux.HandleFunc("/live/output.m3u8", utils.ServeHLS)

	mux.Handle("/live/", http.StripPrefix("/live/", utils.ServeHLSFolderWithCORS("web/live/")))
	mux.Handle("/.videos/past-streams/", http.StripPrefix("/.videos/past-streams/", utils.ServePastStreamsWithCORS("web/.videos/past-streams/")))

	// Initialize Routes
	routes.InitializeRoutes(mux)

	fmt.Printf("Server is running on http://localhost:%d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
}
