package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"goFrame/src/utils"

	"github.com/btcsuite/btcutil/bech32"
)

var (
	CLN_REST_URL string
	RUNE_TOKEN   string
)

func init() {
	config, err := utils.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	CLN_REST_URL = config.CLNRestURL
	RUNE_TOKEN = config.RuneToken
}

type InvoiceRequest struct {
	Amount int64  `json:"amount_msat"`
	Label  string `json:"label"`
	Desc   string `json:"description"`
}

type InvoiceResponse struct {
	PaymentHash   string `json:"payment_hash"`
	ExpiresAt     int64  `json:"expires_at"`
	Bolt11        string `json:"bolt11"`
	PaymentSecret string `json:"payment_secret"`
	CreatedIndex  int    `json:"created_index"`
}

// PaymentLog represents a payment entry
type PaymentLog struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func generateUniqueLabel() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("nostr-%d", r.Intn(1000000))
}

func CLNInvoice(sats int) (string, string, error) {
	label := generateUniqueLabel()
	url := fmt.Sprintf("%s/v1/invoice", CLN_REST_URL)

	requestData := InvoiceRequest{
		Amount: int64(sats * 1000), // Convert sats to msats
		Label:  label,
		Desc:   "Payment for service",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling request data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Rune", RUNE_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response InvoiceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if response.Bolt11 == "" {
		return "", "", fmt.Errorf("empty bolt11 in response: %s", string(body))
	}

	return response.Bolt11, label, nil
}

func WaitForNostrInvoice(label, name, key string) {
	fmt.Println("Waiting for invoice with label:", label)
	url := fmt.Sprintf("%s/v1/waitinvoice", CLN_REST_URL)

	payload := map[string]string{"label": label}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error creating JSON payload:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Rune", RUNE_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error waiting for invoice:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Println("Invoice not paid or other error. Status code:", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Invoice paid! Details:", string(body))

	logPayment(name, key)

	// Notify SSE clients
	if ch, exists := sseClients[label]; exists {
		ch <- `{"status": "paid"}`
	}
}

// DecodeNpub decodes a Bech32 encoded npub to its corresponding pubkey
func DecodeNpub(npub string) (string, error) {
	hrp, data, err := bech32.Decode(npub)
	if err != nil {
		return "", err
	}
	if hrp != "npub" {
		return "", errors.New("invalid hrp")
	}

	decodedData, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", err
	}

	return strings.ToLower(hex.EncodeToString(decodedData)), nil
}

// logPayment logs the payment with name as the key and pubkey as the value under "names"
func logPayment(name, npub string) {
	pubkey, err := DecodeNpub(npub)
	if err != nil {
		fmt.Println("Error decoding npub:", err)
		return
	}

	logFile := "web/logs/nostr.json"

	// Read existing file content
	var logs map[string]map[string]string // Structure: {"names": { "name": "key", ... }}
	file, err := os.ReadFile(logFile)
	if err == nil && len(file) > 0 {
		_ = json.Unmarshal(file, &logs) // Ignore errors; assume empty if invalid
	}

	// Ensure the "names" key exists
	if logs == nil {
		logs = make(map[string]map[string]string)
	}
	if logs["names"] == nil {
		logs["names"] = make(map[string]string)
	}

	// Add or update entry
	logs["names"][name] = pubkey

	// Write back to file
	jsonData, _ := json.MarshalIndent(logs, "", "  ")
	os.WriteFile(logFile, jsonData, 0644)

	fmt.Println("Payment logged successfully!")
}

func HandleNostrInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name string `json:"name"`
		Npub string `json:"npub"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		fmt.Println("Error decoding request payload:", err) // Log error
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println("Received input:", input) // Log input data

	bolt11, label, err := CLNInvoice(10000)
	if err != nil {
		fmt.Println("Error creating invoice:", err) // Log error
		http.Error(w, "Error creating invoice", http.StatusInternalServerError)
		return
	}

	fmt.Println("Generated label:", label) // Add logging

	go WaitForNostrInvoice(label, input.Name, input.Npub)

	response := map[string]string{"bolt11": bolt11, "label": label}
	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling JSON response:", err) // Log error
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData) // Write JSON directly to ResponseWriter
}
