package api

import (
	"fmt"
	"net/http"
	"time"
)

// Store active SSE clients
var sseClients = make(map[string]chan string)

// SSE handler for invoice status updates
func InvoiceEventsHandler(w http.ResponseWriter, r *http.Request) {
	label := r.URL.Query().Get("label")
	if label == "" {
		http.Error(w, "Missing label", http.StatusBadRequest)
		return
	}

	// Set SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create a channel for this client
	messageChan := make(chan string)
	sseClients[label] = messageChan
	defer delete(sseClients, label)

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
			return // Close connection after sending "paid"
		case <-time.After(30 * time.Second): // Prevent infinite hanging
			fmt.Fprintf(w, "data: %s\n\n", `{"status": "pending"}`)
			flusher.Flush()
		}
	}
}
