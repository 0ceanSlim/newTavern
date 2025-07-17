package handlers

import (
	"encoding/json"
	"fmt"
	"goFrame/src/lightning"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// InvoiceRequest handles LNURL-Pay invoice generation with zap support
func InvoiceRequest(w http.ResponseWriter, r *http.Request) {
	// Extract username
	username := r.URL.Query().Get("username")
	amountStr := r.URL.Query().Get("amount")
	nostrParam := r.URL.Query().Get("nostr")

	// Handle clients that might append ?amount instead of &amount
	if strings.Contains(username, "?amount=") {
		parts := strings.Split(username, "?amount=")
		if len(parts) > 1 {
			username = parts[0]
			amountStr = parts[1]
		}
	}

	// Validate amount (must be in msats)
	amountMsats, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountMsats < 1000 {
		http.Error(w, "Invalid or missing amount", http.StatusBadRequest)
		return
	}

	// Validate username
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	var zapRequest *ZapRequest
	var description string

	// If nostr parameter is present, this is a zap request
	if nostrParam != "" {
		// Decode the nostr parameter
		decodedNostr, err := url.QueryUnescape(nostrParam)
		if err != nil {
			http.Error(w, "Invalid nostr parameter encoding", http.StatusBadRequest)
			return
		}

		// Parse the zap request
		if err := json.Unmarshal([]byte(decodedNostr), &zapRequest); err != nil {
			http.Error(w, "Invalid zap request JSON", http.StatusBadRequest)
			return
		}

		// Validate the zap request
		if !validateZapRequest(zapRequest, amountMsats, username) {
			http.Error(w, "Invalid zap request", http.StatusBadRequest)
			return
		}

		// Use the zap request as the description
		description = decodedNostr
	} else {
		// Regular LNURL payment
		description = fmt.Sprintf("Payment to %s", username)
	}

	// Create invoice with description
	invoice, err := lightning.FetchInvoiceWithDescription(amountMsats, description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create invoice: %v", err), http.StatusInternalServerError)
		return
	}

	// Send LNURL response
	response := map[string]interface{}{
		"pr": invoice,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PathInvoiceRequest handles LNURL-Pay invoice generation with username in the path
func PathInvoiceRequest(w http.ResponseWriter, r *http.Request) {
	// Extract username from path
	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 4 {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	username := pathParts[len(pathParts)-1]

	// Get parameters
	amountStr := r.URL.Query().Get("amount")
	nostrParam := r.URL.Query().Get("nostr")

	if amountStr == "" {
		http.Error(w, "Amount required", http.StatusBadRequest)
		return
	}

	// Validate amount (must be in msats)
	amountMsats, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amountMsats < 1000 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	var zapRequest *ZapRequest
	var description string

	// If nostr parameter is present, this is a zap request
	if nostrParam != "" {
		// Decode the nostr parameter
		decodedNostr, err := url.QueryUnescape(nostrParam)
		if err != nil {
			http.Error(w, "Invalid nostr parameter encoding", http.StatusBadRequest)
			return
		}

		// Parse the zap request
		if err := json.Unmarshal([]byte(decodedNostr), &zapRequest); err != nil {
			http.Error(w, "Invalid zap request JSON", http.StatusBadRequest)
			return
		}

		// Validate the zap request
		if !validateZapRequest(zapRequest, amountMsats, username) {
			http.Error(w, "Invalid zap request", http.StatusBadRequest)
			return
		}

		// Use the zap request as the description
		description = decodedNostr
	} else {
		// Regular LNURL payment
		description = fmt.Sprintf("Payment to %s", username)
	}

	// Create invoice with description
	invoice, err := lightning.FetchInvoiceWithDescription(amountMsats, description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create invoice: %v", err), http.StatusInternalServerError)
		return
	}

	// Send LNURL response
	response := map[string]interface{}{
		"pr": invoice,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validateZapRequest validates a zap request according to NIP-57
func validateZapRequest(zr *ZapRequest, amountMsats int64, username string) bool {
	// Must be kind 9734
	if zr.Kind != 9734 {
		return false
	}

	// Must have tags
	if len(zr.Tags) == 0 {
		return false
	}

	// Check for required tags
	var hasP, hasRelays bool
	var zapAmount int64

	for _, tag := range zr.Tags {
		if len(tag) < 2 {
			continue
		}

		switch tag[0] {
		case "p":
			hasP = true
		case "relays":
			hasRelays = true
		case "amount":
			if len(tag) > 1 {
				if amt, err := strconv.ParseInt(tag[1], 10, 64); err == nil {
					zapAmount = amt
				}
			}
		}
	}

	// Must have exactly one 'p' tag and relays
	if !hasP || !hasRelays {
		return false
	}

	// If amount is specified in zap request, it must match
	if zapAmount > 0 && zapAmount != amountMsats {
		return false
	}

	return true
}
