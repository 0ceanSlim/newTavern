package lightning

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// publishZapReceipt publishes a zap receipt to multiple Nostr relays
func publishZapReceipt(event *NostrEvent, relays []string) error {
	if len(relays) == 0 {
		return fmt.Errorf("no relays specified for publishing zap receipt")
	}

	var wg sync.WaitGroup
	var lastError error
	successCount := 0

	for _, relayURL := range relays {
		wg.Add(1)
		go func(relay string) {
			defer wg.Done()

			if err := publishToRelay(event, relay); err != nil {
				log.Printf("Failed to publish zap receipt to %s: %v", relay, err)
				lastError = err
			} else {
				log.Printf("Successfully published zap receipt to %s", relay)
				successCount++
			}
		}(relayURL)
	}

	wg.Wait()

	if successCount == 0 {
		return fmt.Errorf("failed to publish to any relay, last error: %v", lastError)
	}

	log.Printf("Published zap receipt to %d/%d relays", successCount, len(relays))
	return nil
}

// publishToRelay publishes a single event to a specific relay
func publishToRelay(event *NostrEvent, relayURL string) error {
	// Parse and validate URL
	u, err := url.Parse(relayURL)
	if err != nil {
		return fmt.Errorf("invalid relay URL %s: %w", relayURL, err)
	}

	// Connect to relay
	conn, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		return fmt.Errorf("failed to connect to relay %s: %w", relayURL, err)
	}
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Create EVENT message
	message := []interface{}{"EVENT", event}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal event message: %w", err)
	}

	// Send event
	_, err = conn.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	// Read response (optional, but good for debugging)
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		// Don't treat read errors as fatal, some relays might not respond
		log.Printf("No response from relay %s (this is often normal)", relayURL)
	} else {
		log.Printf("Response from %s: %s", relayURL, string(response[:n]))
	}

	return nil
}
