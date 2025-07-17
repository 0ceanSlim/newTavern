// Create new file: src/lightning/nostr.go
package lightning

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

	"goFrame/src/utils"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"golang.org/x/net/websocket"
)

// LightningNostrEvent represents a Nostr event for lightning service
type LightningNostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

var (
	lightningPrivateKey *btcec.PrivateKey
	lightningPublicKey  string
)

// Initialize lightning service keypair
func init() {
	// Generate a random private key for the lightning service
	var err error
	lightningPrivateKey, err = btcec.NewPrivateKey()
	if err != nil {
		log.Fatalf("Failed to generate lightning service private key: %v", err)
	}

	// Derive the public key
	pubKeyBytes := schnorr.SerializePubKey(lightningPrivateKey.PubKey())
	lightningPublicKey = hex.EncodeToString(pubKeyBytes)

	log.Printf("Lightning service initialized with pubkey: %s", lightningPublicKey)
}

// GetLightningPublicKey returns the lightning service public key
func GetLightningPublicKey() string {
	return lightningPublicKey
}

// GetLightningRelays returns the configured relays for lightning zap receipts
func GetLightningRelays() []string {
	if len(utils.AppConfig.Lightning.ZapRelays) > 0 {
		return utils.AppConfig.Lightning.ZapRelays
	}

	// Default relays if none configured
	return []string{
		"wss://wheat.happytavern.co",
		"wss://nos.lol",
		"wss://relay.damus.io",
	}
}

// createLightningEvent creates a new Nostr event signed by the lightning service
func createLightningEvent(kind int, content string, tags [][]string) (*LightningNostrEvent, error) {
	event := LightningNostrEvent{
		PubKey:    lightningPublicKey,
		CreatedAt: time.Now().Unix(),
		Kind:      kind,
		Tags:      tags,
		Content:   content,
	}

	// Create the exact serialization format required by NIP-01
	serializedData := []interface{}{
		0,
		event.PubKey,
		event.CreatedAt,
		event.Kind,
		event.Tags,
		event.Content,
	}

	// Use a custom JSON encoder to ensure proper formatting
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(serializedData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event: %w", err)
	}

	// Remove the trailing newline that Encode adds
	serialized := bytes.TrimSpace(buffer.Bytes())

	// Calculate ID
	hash := sha256.Sum256(serialized)
	event.ID = fmt.Sprintf("%x", hash[:])

	// Sign the event with lightning service key
	event.Sig = signLightningEvent(hash[:])
	if event.Sig == "" {
		return nil, fmt.Errorf("failed to sign event")
	}

	return &event, nil
}

// signLightningEvent signs an event using the lightning service private key
func signLightningEvent(eventID []byte) string {
	sig, err := schnorr.Sign(lightningPrivateKey, eventID)
	if err != nil {
		log.Printf("Failed to sign lightning event: %v", err)
		return ""
	}

	signature := hex.EncodeToString(sig.Serialize())
	return signature
}

// sendLightningEventToRelays sends a lightning event to specified relays
func sendLightningEventToRelays(event *LightningNostrEvent, relayList []string) {
	if event == nil {
		log.Printf("Error: Attempted to send nil lightning event")
		return
	}

	log.Printf("Sending lightning event ID %s to %d relays", event.ID, len(relayList))

	if len(relayList) == 0 {
		log.Printf("Warning: No relays specified for lightning event")
		return
	}

	var wg sync.WaitGroup
	for _, relayURL := range relayList {
		wg.Add(1)
		go func(relay string) {
			defer wg.Done()

			conn, err := connectToRelay(relay)
			if err != nil {
				log.Printf("Error connecting to relay %s: %v", relay, err)
				return
			}
			defer conn.Close()

			msg, err := json.Marshal([]interface{}{"EVENT", event})
			if err != nil {
				log.Printf("Error encoding lightning event for relay %s: %v", relay, err)
				return
			}

			_, err = conn.Write(msg)
			if err != nil {
				log.Printf("Error sending lightning event to relay %s: %v", relay, err)
				return
			}

			log.Printf("Successfully sent lightning event to relay %s", relay)
		}(relayURL)
	}

	wg.Wait()
}

// connectToRelay connects to a relay via WebSocket
func connectToRelay(relay string) (*websocket.Conn, error) {
	u, err := url.Parse(relay)
	if err != nil {
		return nil, fmt.Errorf("invalid relay URL %s: %w", relay, err)
	}

	conn, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", relay, err)
	}

	return conn, nil
}
