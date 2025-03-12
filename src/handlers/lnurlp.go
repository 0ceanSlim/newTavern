package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// LNURLpResponse represents the metadata returned from .well-known/lnurlp/{username}
type LNURLpResponse struct {
	Tag            string `json:"tag"`
	Callback       string `json:"callback"`
	Metadata       string `json:"metadata"`
	MinSendable    int64  `json:"minSendable"`
	MaxSendable    int64  `json:"maxSendable"`
	CommentAllowed int    `json:"commentAllowed"`
}

// LNURLpHandler serves metadata for a user at .well-known/lnurlp/{username}
func LNURLpHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/.well-known/lnurlp/")

	// Validate username
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// Construct callback URL where the wallet will request an invoice
	callback := fmt.Sprintf("https://%s/lnurl/pay?username=%s", r.Host, username)

	// Define LNURLp metadata response
	response := LNURLpResponse{
		Tag:            "payRequest",
		Callback:       callback,
		Metadata:       fmt.Sprintf("[[\"text/plain\", \"Pay %s\"]]", username),
		MinSendable:    1000,     // 1 sat (1000 msats)
		MaxSendable:    10000000, // 10,000 sats (10,000,000 msats)
		CommentAllowed: 120,      // Allow comments up to 120 chars
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
