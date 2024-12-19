package routes

import "net/http"

func InitializeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/settings", Settings)
	mux.HandleFunc("/what-is-nostr", WhatIsNostr)
	mux.HandleFunc("/bitcoin-works", BitcoinWorks)
	mux.HandleFunc("/mempool", Mempool)
	mux.HandleFunc("/nostr-clients", NostrClients)
	mux.HandleFunc("/nostr-mobile", NostrMobile)
	mux.HandleFunc("/nostr-mobile-android", NostrMobileAndroid)
	mux.HandleFunc("/nostr-mobile-ios", NostrMobileIos)
	mux.HandleFunc("/nostr-desktop", NostrDesktop)
	mux.HandleFunc("/grain", Grain)
	mux.HandleFunc("/gun-blog", GunBlog)
	mux.HandleFunc("/gun-blog/ar15guide", ArGuide)
	mux.HandleFunc("/contact", Contact)
	mux.HandleFunc("/live/view", LiveView)
}
