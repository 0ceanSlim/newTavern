package handlers

import (
	"encoding/json"
	"fmt"
	"goFrame/src/lightning"
	"log"
	"net/http"
	"strconv"
)

// InvoiceRequest handles LNURL-Pay invoice generation
func InvoiceRequest(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	amountStr := r.URL.Query().Get("amount")

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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
