package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds server-related configurations
type ServerConfig struct {
	Port int  `yaml:"port"`
	TLS  bool `yaml:"tls"`
}

// LightningConfig holds settings for the Lightning backend (LND, CLN, or Eclair)
type LightningConfig struct {
	Type       string `yaml:"type"`         // "lnd", "cln", or "eclair"
	PeerID     string `yaml:"peer_id"`      // Node address
	Rune       string `yaml:"rune"`         // CLN Runes (if applicable)
	CLNRestURL string `yaml:"cln_rest_url"` // REST API URL
}

// NostrConfig holds Nostr-related settings
type NostrConfig struct {
	PublicKey  string   `yaml:"public_key"`  // Your Nostr public key (hex)
	PrivateKey string   `yaml:"private_key"` // Your Nostr private key (hex)
	Relays     []string `yaml:"relays"`      // Default relays to publish to
}

// Config holds the full application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Lightning LightningConfig `yaml:"lightning"`
	Nostr     NostrConfig     `yaml:"nostr"`
}

// Global variable to hold the config after loading
var AppConfig Config

// LoadConfig reads the YAML config file into the AppConfig struct
func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}
