package main

import (
	//"goFrame/src/handlers"
	"goFrame/src/routes"
	"goFrame/src/utils"

	"fmt"
	"net/http"
)

func main() {
	// Load Configurations
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	mux := http.NewServeMux()

	// Initialize Routes
	mux.HandleFunc("/", routes.Index)
	mux.HandleFunc("/settings", routes.Settings)
	mux.HandleFunc("/what-is-nostr", routes.WhatIsNostr)
	mux.HandleFunc("/bitcoin-works", routes.BitcoinWorks)
	mux.HandleFunc("/mempool", routes.Mempool)
	mux.HandleFunc("/nostr-clients", routes.NostrClients)
	mux.HandleFunc("/nostr-mobile", routes.NostrMobile)
	mux.HandleFunc("/nostr-mobile-android", routes.NostrMobileAndroid)
	mux.HandleFunc("/nostr-mobile-ios", routes.NostrMobileIos)
	mux.HandleFunc("/nostr-desktop", routes.NostrDesktop)

	fmt.Printf("Server is running on http://localhost:%d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
}
