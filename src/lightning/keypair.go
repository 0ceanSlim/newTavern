package lightning

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
)

var (
	lightningPrivateKey *btcec.PrivateKey
	lightningPublicKey  string
)

// Initialize lightning service keypair when package loads
func init() {
	if err := generateLightningKeypair(); err != nil {
		log.Fatalf("Failed to generate lightning service keypair: %v", err)
	}
}

// generateLightningKeypair creates a new random keypair for the lightning service
func generateLightningKeypair() error {
	// Generate a random private key
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Store the private key
	lightningPrivateKey = privKey

	// Get the public key in hex format (32 bytes)
	pubKeyBytes := privKey.PubKey().SerializeCompressed()[1:] // Remove the 0x02/0x03 prefix
	lightningPublicKey = hex.EncodeToString(pubKeyBytes)

	log.Printf("Lightning service initialized with pubkey: %s", lightningPublicKey)
	return nil
}

// GetLightningPublicKey returns the lightning service's public key in hex format
func GetLightningPublicKey() string {
	return lightningPublicKey
}

// GetLightningPrivateKey returns the lightning service's private key
// This should only be used internally for signing zap receipts
func GetLightningPrivateKey() *btcec.PrivateKey {
	return lightningPrivateKey
}
