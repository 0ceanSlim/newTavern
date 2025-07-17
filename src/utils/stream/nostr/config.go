package nostr

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"goFrame/src/utils"

	"github.com/btcsuite/btcd/btcec/v2"
	"gopkg.in/yaml.v3"
)

// Event represents a Nostr event structure
type Event struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// Config represents the structure of nostr.yml (fallback)
type Config struct {
	PrivateKey string   `yaml:"private_key"`
	PublicKey  string   `yaml:"public_key"`
	Relays     []string `yaml:"relays"`
}

var (
	privateKey *btcec.PrivateKey
	publicKey  string
	relays     []string
)

// Initialize nostr configuration from main config or fallback to nostr.yml
func init() {
	// Try to load from main config first
	if err := LoadConfigFromMain(); err != nil {
		// Fallback to nostr.yml
		if err := LoadConfig("nostr.yml"); err != nil {
			log.Printf("Warning: Failed to load Nostr config from both sources: %v", err)
		}
	}
}

// LoadConfigFromMain loads Nostr config from the main application config
func LoadConfigFromMain() error {
	// Check if main config is loaded and has Nostr settings
	if utils.AppConfig.Nostr.PublicKey == "" || utils.AppConfig.Nostr.PrivateKey == "" {
		return fmt.Errorf("nostr configuration not found in main config")
	}

	// Parse private key
	keyBytes, err := hex.DecodeString(utils.AppConfig.Nostr.PrivateKey)
	if err != nil {
		return fmt.Errorf("error decoding private key from main config: %w", err)
	}

	privateKey, _ = btcec.PrivKeyFromBytes(keyBytes)

	// Parse public key
	publicKeyBytes, err := hex.DecodeString(utils.AppConfig.Nostr.PublicKey)
	if err != nil {
		return fmt.Errorf("error decoding public key from main config: %w", err)
	}
	publicKey = fmt.Sprintf("%x", publicKeyBytes)

	// Set relays
	relays = utils.AppConfig.Nostr.Relays

	log.Println("Nostr configuration loaded from main config")
	return nil
}

// LoadConfig reads nostr.yml and loads the configuration (fallback)
func LoadConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	keyBytes, err := hex.DecodeString(cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("error decoding private key: %w", err)
	}

	privateKey, _ = btcec.PrivKeyFromBytes(keyBytes)
	publicKeyBytes, err := hex.DecodeString(cfg.PublicKey)
	if err != nil {
		return fmt.Errorf("error decoding public key: %w", err)
	}
	publicKey = fmt.Sprintf("%x", publicKeyBytes)

	relays = cfg.Relays
	log.Println("Nostr configuration loaded from nostr.yml")
	return nil
}

// GetPublicKey returns the configured public key
func GetPublicKey() string {
	return publicKey
}

// GetRelays returns the configured relays
func GetRelays() []string {
	return relays
}
