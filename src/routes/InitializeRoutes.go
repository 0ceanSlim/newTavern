package routes

import "net/http"

func InitializeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/index-mockup", IndexMockup)
	mux.HandleFunc("/settings", Settings)
	mux.HandleFunc("/what-is-nostr", WhatIsNostr)
	mux.HandleFunc("/bitcoin-works", BitcoinWorks)
	mux.HandleFunc("/mempool", Mempool)
	mux.HandleFunc("/grain", Grain)
	mux.HandleFunc("/nostr-hero", NostrHero)
	mux.HandleFunc("/gnostream", Gnostream)
	mux.HandleFunc("/nostr-software", NostrSoftware)
	mux.HandleFunc("/gun-blog", GunBlog)
	mux.HandleFunc("/gun-blog/ar15guide", ArGuide)
	mux.HandleFunc("/file-upload", FileUpload)
	mux.HandleFunc("/contact", Contact)
	mux.HandleFunc("/btc-dash", BitcoinDashboard)
	mux.HandleFunc("/nostr-verified", NostrVerified)
	mux.HandleFunc("/core-lnurl", CoreLnurl)
}
