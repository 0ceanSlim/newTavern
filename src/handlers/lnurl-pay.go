package handlers

import (
	"encoding/json"
	"fmt"
	"goFrame/src/lightning"
	"net/http"
	"strconv"
	"strings"
)

// InvoiceRequest handles LNURL-Pay invoice generation
func InvoiceRequest(w http.ResponseWriter, r *http.Request) {
	// Extract username
	username := r.URL.Query().Get("username")
	
	// Handle clients that might append ?amount instead of &amount
	amountStr := r.URL.Query().Get("amount")
	
	// Check if username contains an embedded amount parameter
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

	// Fetch invoice via CLN REST
	invoice, err := lightning.FetchInvoice(amountMsats, fmt.Sprintf("Payment to %s", username))
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
// This can be used for endpoint like /lnurl/pay/{username}
func PathInvoiceRequest(w http.ResponseWriter, r *http.Request) {
	// Extract username from path
	path := r.URL.Path
	pathParts := strings.Split(path, "/")
	
	// The username should be the last part of the path
	if len(pathParts) < 1 {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	
	username := pathParts[len(pathParts)-1]
	
	// Get amount from query parameter
	amountStr := r.URL.Query().Get("amount")
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

	// Fetch invoice via CLN REST
	invoice, err := lightning.FetchInvoice(amountMsats, fmt.Sprintf("Payment to %s", username))
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