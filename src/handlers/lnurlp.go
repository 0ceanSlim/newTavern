package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
	// Extract username from path segments
	pathSegments := strings.Split(r.URL.Path, "/")
	
	// Path format: /.well-known/lnurlp/{username}
	if len(pathSegments) < 5 || pathSegments[3] != "lnurlp" {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	
	username := pathSegments[4]
	
	// Validate username format (basic check)
	if username == "" || strings.ContainsAny(username, "/\\") {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	// Rest of your existing code...
	callback := fmt.Sprintf("https://%s/lnurl/pay?username=%s", r.Host, username)
	response := LNURLpResponse{
		Tag:            "payRequest",
		Callback:       callback,
		Metadata:       fmt.Sprintf("[[\"text/plain\", \"Pay %s\"]]", username),
		MinSendable:    1000,
		MaxSendable:    10000000,
		CommentAllowed: 120,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}