package lightning

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"goFrame/src/utils"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// NostrEvent represents a Nostr event
type NostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// ZapRequestData represents parsed zap request
type ZapRequestData struct {
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
}

// CreateAndPublishZapReceipt creates a zap receipt and publishes it to relays
func CreateAndPublishZapReceipt(zapRequestJSON, bolt11 string, paymentInfo *WaitInvoiceResponse) error {
	// Parse the original zap request
	var zapRequest ZapRequestData
	if err := json.Unmarshal([]byte(zapRequestJSON), &zapRequest); err != nil {
		return fmt.Errorf("failed to parse zap request: %w", err)
	}

	// Create zap receipt event
	zapReceipt, err := createZapReceiptEvent(&zapRequest, zapRequestJSON, bolt11, paymentInfo)
	if err != nil {
		return fmt.Errorf("failed to create zap receipt event: %w", err)
	}

	// Get relays to publish to
	relays := getZapReceiptRelays(&zapRequest)

	// Publish to relays
	err = publishZapReceipt(zapReceipt, relays)
	if err != nil {
		return fmt.Errorf("failed to publish zap receipt: %w", err)
	}

	return nil
}

// createZapReceiptEvent creates a kind 9735 zap receipt event
func createZapReceiptEvent(zapRequest *ZapRequestData, zapRequestJSON, bolt11 string, paymentInfo *WaitInvoiceResponse) (*NostrEvent, error) {
	// Build tags for zap receipt
	tags := [][]string{
		{"bolt11", bolt11},
		{"description", zapRequestJSON},
	}

	// Extract required tags from zap request
	var recipientPubkey string

	for _, tag := range zapRequest.Tags {
		if len(tag) < 2 {
			continue
		}

		switch tag[0] {
		case "p":
			// Recipient pubkey (required)
			recipientPubkey = tag[1]
			tags = append(tags, []string{"p", tag[1]})
		case "e":
			// Event being zapped (optional)
			tags = append(tags, []string{"e", tag[1]})
		case "a":
			// Event coordinate (optional)
			tags = append(tags, []string{"a", tag[1]})
		}
	}

	// Add sender pubkey (P tag)
	if zapRequest.PubKey != "" {
		tags = append(tags, []string{"P", zapRequest.PubKey})
	}

	// Add preimage if available
	if paymentInfo.Preimage != "" {
		tags = append(tags, []string{"preimage", paymentInfo.Preimage})
	}

	// Validate required fields
	if recipientPubkey == "" {
		return nil, fmt.Errorf("zap request missing required 'p' tag (recipient pubkey)")
	}

	// Create the event
	event := &NostrEvent{
		PubKey:    GetLightningPublicKey(),
		CreatedAt: paymentInfo.PaidAt,
		Kind:      9735, // Zap receipt
		Tags:      tags,
		Content:   "", // Should be empty for zap receipts
	}

	// Calculate event ID and signature
	if err := signNostrEvent(event); err != nil {
		return nil, fmt.Errorf("failed to sign zap receipt: %w", err)
	}

	return event, nil
}

// signNostrEvent calculates the ID and signature for a Nostr event
func signNostrEvent(event *NostrEvent) error {
	// Create the serialization for ID calculation
	serializedData := []interface{}{
		0,
		event.PubKey,
		event.CreatedAt,
		event.Kind,
		event.Tags,
		event.Content,
	}

	// Serialize to JSON
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(serializedData); err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Remove trailing newline
	serialized := bytes.TrimSpace(buffer.Bytes())

	// Calculate ID (SHA256 hash)
	hash := sha256.Sum256(serialized)
	event.ID = hex.EncodeToString(hash[:])

	// Sign the event
	privKey := GetLightningPrivateKey()
	if privKey == nil {
		return fmt.Errorf("lightning private key not available")
	}

	sig, err := schnorr.Sign(privKey, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign event: %w", err)
	}

	event.Sig = hex.EncodeToString(sig.Serialize())

	log.Printf("Created zap receipt event with ID: %s", event.ID)
	return nil
}

// getZapReceiptRelays extracts relay list from zap request and combines with config relays
func getZapReceiptRelays(zapRequest *ZapRequestData) []string {
	var relays []string

	// Extract relays from zap request
	for _, tag := range zapRequest.Tags {
		if len(tag) >= 2 && tag[0] == "relays" {
			// The relays tag contains multiple relay URLs
			for i := 1; i < len(tag); i++ {
				relays = append(relays, tag[i])
			}
			break
		}
	}

	// Add configured zap relays
	configRelays := utils.AppConfig.Lightning.ZapRelays
	relays = append(relays, configRelays...)

	// Remove duplicates
	uniqueRelays := make(map[string]bool)
	var result []string
	for _, relay := range relays {
		if relay != "" && !uniqueRelays[relay] {
			uniqueRelays[relay] = true
			result = append(result, relay)
		}
	}

	log.Printf("Publishing zap receipt to %d relays: %v", len(result), result)
	return result
}
