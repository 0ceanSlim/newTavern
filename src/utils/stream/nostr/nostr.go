package nostr

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"golang.org/x/net/websocket"
)

// Connects to a relay via WebSocket with improved debugging
func connectToRelay(relay string) (*websocket.Conn, error) {
	log.Printf("Attempting to connect to relay: %s", relay)
	u, err := url.Parse(relay)
	if err != nil {
		return nil, fmt.Errorf("invalid relay URL %s: %w", relay, err)
	}

	conn, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", relay, err)
	}

	log.Printf("Successfully connected to relay: %s", relay)
	return conn, nil
}

func createEvent(kind int, content string, tags [][]string) (*Event, error) {
	event := Event{
		PubKey:    publicKey,
		CreatedAt: time.Now().Unix(),
		Kind:      kind,
		Tags:      tags,
		Content:   content,
	}

	// Create the exact serialization format required by NIP-01
	serializedData := []interface{}{
		0,
		event.PubKey, // Make sure this is already a lowercase hex string
		event.CreatedAt,
		event.Kind,
		event.Tags,
		event.Content,
	}

	// Use a custom JSON encoder to ensure proper formatting
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false) // Important for content with HTML characters
	err := encoder.Encode(serializedData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event: %w", err)
	}

	// Remove the trailing newline that Encode adds
	serialized := bytes.TrimSpace(buffer.Bytes())

	log.Printf("Serialization for ID calculation: %s", serialized)

	// Calculate ID
	hash := sha256.Sum256(serialized)
	event.ID = fmt.Sprintf("%x", hash[:])

	log.Printf("Calculated event ID: %s", event.ID)

	// Sign the event
	event.Sig = signEvent(hash[:])
	if event.Sig == "" {
		return nil, fmt.Errorf("failed to sign event")
	}

	return &event, nil
}

// Signs an event using Schnorr signatures with debugging
func signEvent(eventID []byte) string {
	log.Printf("Signing event ID: %x", eventID)

	sig, err := schnorr.Sign(privateKey, eventID)
	if err != nil {
		log.Printf("Failed to sign event: %v", err)
		return ""
	}

	signature := hex.EncodeToString(sig.Serialize())
	log.Printf("Event signed successfully with signature: %s", signature)
	return signature
}

// Sends the Nostr event to all relays concurrently with improved debugging
func sendEvent(event *Event) {
	if event == nil {
		log.Printf("Error: Attempted to send nil event")
		return
	}

	log.Printf("Starting to send event ID %s to %d relays", event.ID, len(relays))

	if len(relays) == 0 {
		log.Printf("Warning: No relays configured to send to")
		return
	}

	var wg sync.WaitGroup
	for _, relayURL := range relays {
		wg.Add(1)
		go func(relay string) {
			defer wg.Done()
			log.Printf("Connecting to relay: %s", relay)

			conn, err := connectToRelay(relay)
			if err != nil {
				log.Printf("Error connecting to relay %s: %v", relay, err)
				return
			}
			defer conn.Close()

			msg, err := json.Marshal([]interface{}{"EVENT", event})
			if err != nil {
				log.Printf("Error encoding event for relay %s: %v", relay, err)
				return
			}

			log.Printf("Sending to %s: %s", relay, string(msg))

			n, err := conn.Write(msg)
			if err != nil {
				log.Printf("Error sending event to relay %s: %v", relay, err)
				return
			}

			log.Printf("Successfully sent %d bytes to relay %s", n, relay)

			// Add a response listener for confirmation
			var response = make([]byte, 1024)
			n, err = conn.Read(response)
			if err != nil {
				log.Printf("Error reading response from relay %s: %v", relay, err)
				return
			}

			log.Printf("Received response from %s: %s", relay, string(response[:n]))
		}(relayURL)
	}

	log.Printf("Waiting for all relay operations to complete...")
	wg.Wait()
	log.Printf("Event %s sent to all available relays", event.ID)
}
