package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Check if the decoded pubkey exists in nostr.json
func CheckNpubHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, `<p class="text-red-500">❌ npub is required</p>`, http.StatusBadRequest)
		return
	}

	// Decode npub to public key
	pubkey, err := DecodeNpub(npub)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `<p class="text-red-500">❌ %s</p>`, err.Error())
		return
	}

	logFile := "web/.well-known/nostr.json"

	// Read the JSON file
	file, err := os.ReadFile(logFile)
	if err != nil {
		http.Error(w, `<p class="text-red-500">❌ Could not read log file</p>`, http.StatusInternalServerError)
		return
	}

	// Parse JSON into a struct
	var logs struct {
		Names map[string]string `json:"names"`
	}

	if err := json.Unmarshal(file, &logs); err != nil {
		http.Error(w, `<p class="text-red-500">❌ Invalid JSON format</p>`, http.StatusInternalServerError)
		return
	}

	// Check if pubkey exists
	for _, existingKey := range logs.Names {
		if existingKey == pubkey {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			fmt.Fprint(w, `<p class="text-yellow-500">⚡ You're already verified!</p>`)
			return
		}
	}

	// If the key does not exist
	w.WriteHeader(http.StatusOK) // 200 OK
	fmt.Fprint(w, `<p class="text-green-500">✅ This npub is available</p>`)
}
