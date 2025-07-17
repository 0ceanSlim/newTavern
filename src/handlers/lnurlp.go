package handlers

import (
	"encoding/json"
	"fmt"
	"goFrame/src/lightning"
	"net/http"
	"strings"
)

// LNURLpResponse represents the metadata returned from .well-known/lnurlp/{username}
type LNURLpResponse struct {
	Tag            string `json:"tag"`
	Callback       string `json:"callback"`
	Metadata       string `json:"metadata"`
	MinSendable    int64  `json:"minSendable"`
	MaxSendable    int64  `json:"maxSendable"`
	CommentAllowed int    `json:"commentAllowed"`
	AllowsNostr    bool   `json:"allowsNostr"`
	NostrPubkey    string `json:"nostrPubkey"`
}

// ZapRequest represents a Nostr zap request (kind 9734)
type ZapRequest struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// LNURLpHandler serves metadata for a user at .well-known/lnurlp/{username}
func LNURLpHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/.well-known/lnurlp/")

	// Validate username
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// Get the lightning service public key (auto-generated)
	nostrPubkey := lightning.GetLightningPublicKey()

	// Construct callback URL where the wallet will request an invoice
	callback := fmt.Sprintf("https://%s/lnurl/pay?username=%s", r.Host, username)

	// Define LNURLp metadata response with Nostr support
	response := LNURLpResponse{
		Tag:            "payRequest",
		Callback:       callback,
		Metadata:       fmt.Sprintf("[[\"text/plain\", \"Pay %s\"]]", username),
		MinSendable:    1000,      // 1 sat (1000 msats)
		MaxSendable:    100000000, // 100,000 sats (100,000,000 msats)
		CommentAllowed: 255,       // Allow comments up to 255 chars
		AllowsNostr:    true,      // Enable Nostr zaps
		NostrPubkey:    nostrPubkey,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
