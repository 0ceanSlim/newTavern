package utils

import "net/http"

func ServeWellKnownNostr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, "web/static/.well-known/nostr.json")
}
