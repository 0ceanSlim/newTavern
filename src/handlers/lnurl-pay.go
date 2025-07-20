package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"goFrame/src/lightning"
)

// ZapRequest represents a stored zap request
type ZapRequest struct {
	JSON   string // Original zap request JSON
	Label  string // Invoice label
	Bolt11 string // Invoice when created
}

// Global storage for zap requests (in-memory)
var (
	zapRequests = make(map[string]*ZapRequest) // label -> ZapRequest
	zapMutex    sync.RWMutex
)

// StoreZapRequest stores a zap request for later processing
func StoreZapRequest(label, zapJSON, bolt11 string) {
	zapMutex.Lock()
	defer zapMutex.Unlock()

	zapRequests[label] = &ZapRequest{
		JSON:   zapJSON,
		Label:  label,
		Bolt11: bolt11,
	}
}

// GetZapRequest retrieves a stored zap request
func GetZapRequest(label string) (*ZapRequest, bool) {
	zapMutex.RLock()
	defer zapMutex.RUnlock()

	zap, exists := zapRequests[label]
	return zap, exists
}

// RemoveZapRequest removes a zap request from storage
func RemoveZapRequest(label string) {
	zapMutex.Lock()
	defer zapMutex.Unlock()

	delete(zapRequests, label)
}

// CleanupZapRequest is a public function that can be called from other packages
func CleanupZapRequest(label string) {
	RemoveZapRequest(label)
}

// GetZapRequestCount returns the number of stored zap requests (for debugging)
func GetZapRequestCount() int {
	zapMutex.RLock()
	defer zapMutex.RUnlock()
	return len(zapRequests)
}

// monitorZapPaymentWithCleanup wraps the lightning monitor with cleanup
func monitorZapPaymentWithCleanup(label, zapRequestJSON, bolt11 string) {
	defer func() {
		// Clean up the stored zap request when monitoring is done
		CleanupZapRequest(label)
	}()

	// Call the lightning package monitor function
	lightning.MonitorZapPayment(label, zapRequestJSON, bolt11)
}

// InvoiceRequest handles LNURL-Pay invoice generation with zap support
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

	// Check for zap request (nostr parameter)
	nostrParam := r.URL.Query().Get("nostr")
	var zapRequestJSON string

	if nostrParam != "" {
		// Decode the nostr parameter (it's URI encoded)
		decodedNostr, err := url.QueryUnescape(nostrParam)
		if err != nil {
			http.Error(w, "Invalid nostr parameter encoding", http.StatusBadRequest)
			return
		}

		// Validate it's proper JSON and looks like a zap request
		var zapReq map[string]interface{}
		if err := json.Unmarshal([]byte(decodedNostr), &zapReq); err != nil {
			http.Error(w, "Invalid zap request JSON", http.StatusBadRequest)
			return
		}

		// Basic validation: should be kind 9734
		if kind, ok := zapReq["kind"].(float64); !ok || int(kind) != 9734 {
			http.Error(w, "Not a valid zap request (kind should be 9734)", http.StatusBadRequest)
			return
		}

		zapRequestJSON = decodedNostr
	}

	// Fetch invoice via CLN REST
	invoiceResult, err := lightning.FetchInvoiceWithLabel(amountMsats, fmt.Sprintf("Payment to %s", username))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create invoice: %v", err), http.StatusInternalServerError)
		return
	}

	// If this is a zap request, store it for processing when paid
	if zapRequestJSON != "" {
		StoreZapRequest(invoiceResult.Label, zapRequestJSON, invoiceResult.Bolt11)

		// Start monitoring this zap payment with cleanup
		go monitorZapPaymentWithCleanup(invoiceResult.Label, zapRequestJSON, invoiceResult.Bolt11)
	}

	// Send LNURL response
	response := map[string]interface{}{
		"pr": invoiceResult.Bolt11,
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

	// Check for zap request (nostr parameter)
	nostrParam := r.URL.Query().Get("nostr")
	var zapRequestJSON string

	if nostrParam != "" {
		// Decode the nostr parameter (it's URI encoded)
		decodedNostr, err := url.QueryUnescape(nostrParam)
		if err != nil {
			http.Error(w, "Invalid nostr parameter encoding", http.StatusBadRequest)
			return
		}

		// Validate it's proper JSON and looks like a zap request
		var zapReq map[string]interface{}
		if err := json.Unmarshal([]byte(decodedNostr), &zapReq); err != nil {
			http.Error(w, "Invalid zap request JSON", http.StatusBadRequest)
			return
		}

		// Basic validation: should be kind 9734
		if kind, ok := zapReq["kind"].(float64); !ok || int(kind) != 9734 {
			http.Error(w, "Not a valid zap request (kind should be 9734)", http.StatusBadRequest)
			return
		}

		zapRequestJSON = decodedNostr
	}

	// Fetch invoice via CLN REST
	invoiceResult, err := lightning.FetchInvoiceWithLabel(amountMsats, fmt.Sprintf("Payment to %s", username))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create invoice: %v", err), http.StatusInternalServerError)
		return
	}

	// If this is a zap request, store it for processing when paid
	if zapRequestJSON != "" {
		StoreZapRequest(invoiceResult.Label, zapRequestJSON, invoiceResult.Bolt11)

		// Start monitoring this zap payment with cleanup
		go monitorZapPaymentWithCleanup(invoiceResult.Label, zapRequestJSON, invoiceResult.Bolt11)
	}

	// Send LNURL response
	response := map[string]interface{}{
		"pr": invoiceResult.Bolt11,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
