package nostr

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

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

// Config represents the structure of nostr.yml
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

// Automatically loads config when the package is initialized
func init() {
	if err := LoadConfig("nostr.yml"); err != nil {
		log.Fatalf("Failed to load Nostr config: %v", err)
	}
}

// LoadConfig reads nostr.yml and loads the configuration
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
	return nil
}
