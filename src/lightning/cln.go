package lightning

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"time"

	"goFrame/src/utils"
)

// InvoiceRequest represents the JSON request sent to CLN REST API
type InvoiceRequest struct {
	AmountMsat  int64  `json:"amount_msat"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// InvoiceResponse represents the JSON response from CLN REST API
type InvoiceResponse struct {
	Bolt11        string `json:"bolt11"`
	PaymentHash   string `json:"payment_hash"`
	PaymentSecret string `json:"payment_secret"`
}

// WaitInvoiceRequest represents the request to wait for an invoice payment
type WaitInvoiceRequest struct {
	Label string `json:"label"`
}

// WaitInvoiceResponse represents the response when an invoice is paid
type WaitInvoiceResponse struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	PaymentHash string `json:"payment_hash"`
	Status      string `json:"status"`
	PaidAt      int64  `json:"paid_at"`
	AmountMsat  int64  `json:"amount_received_msat"`
	Preimage    string `json:"payment_preimage"`
}

// Invoice storage for zap receipts
var invoiceStore = make(map[string]string) // label -> bolt11

// FetchInvoice requests an invoice from CLN REST (for regular payments)
func FetchInvoice(amountMsats int64, description string) (string, error) {
	return fetchInvoiceInternal(amountMsats, description, false)
}

// Add this new function to your existing file
func FetchInvoiceWithDescription(amountMsats int64, description string) (string, error) {
	restURL := utils.AppConfig.Lightning.CLNRestURL
	runeToken := utils.AppConfig.Lightning.Rune

	log.Printf("Creating zap invoice: amount=%d msats, CLN_URL=%s", amountMsats, restURL)

	// Construct API URL
	apiURL := fmt.Sprintf("%s/v1/invoice", restURL)

	// Generate a unique label using timestamp
	label := fmt.Sprintf("zap-%d-%d", amountMsats, time.Now().UnixNano())

	// For zaps, use description_hash
	descHash := sha256.Sum256([]byte(description))
	request := map[string]interface{}{
		"amount_msat":      amountMsats,
		"label":            label,
		"description_hash": hex.EncodeToString(descHash[:]),
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request JSON: %v", err)
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (authorization uses Rune)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Rune", runeToken)

	// Send request
	client := &http.Client{}
	log.Printf("Sending zap invoice request to CLN: %s", apiURL)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request to CLN: %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("CLN response status: %d", resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read CLN response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("CLN response body: %s", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("CLN returned error status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("CLN returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response - use the same response struct as your existing FetchInvoice
	var response InvoiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Failed to parse CLN response JSON: %v", err)
		return "", fmt.Errorf("failed to parse response: %w\nResponse: %s", err, string(body))
	}

	if response.Bolt11 == "" {
		log.Printf("CLN returned empty bolt11 invoice")
		return "", fmt.Errorf("CLN returned empty bolt11 invoice")
	}

	log.Printf("Successfully created zap invoice: %s...", response.Bolt11[:50])
	return response.Bolt11, nil
}

// fetchInvoiceInternal handles the actual invoice creation
func fetchInvoiceInternal(amountMsats int64, description string, useDescriptionHash bool) (string, error) {
	restURL := utils.AppConfig.Lightning.CLNRestURL
	runeToken := utils.AppConfig.Lightning.Rune

	// Construct API URL
	apiURL := fmt.Sprintf("%s/v1/invoice", restURL)

	// Generate a unique label using timestamp
	label := fmt.Sprintf("inv-%d-%d", amountMsats, time.Now().UnixNano())

	var requestBody []byte
	var err error

	if useDescriptionHash {
		// For zaps, use description_hash
		descHash := sha256.Sum256([]byte(description))
		request := map[string]interface{}{
			"amount_msat":      amountMsats,
			"label":            label,
			"description_hash": hex.EncodeToString(descHash[:]),
		}
		requestBody, err = json.Marshal(request)
	} else {
		// For regular payments, use description
		request := InvoiceRequest{
			AmountMsat:  amountMsats,
			Label:       label,
			Description: description,
		}
		requestBody, err = json.Marshal(request)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (authorization uses Rune)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Rune", runeToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var response InvoiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w\nResponse: %s", err, string(body))
	}

	// Store the bolt11 for later use in zap receipts
	invoiceStore[label] = response.Bolt11

	// If this is a zap invoice, start monitoring for payment
	if useDescriptionHash {
		go monitorZapInvoice(label, description)
	}

	return response.Bolt11, nil
}

// monitorZapInvoice waits for a zap invoice to be paid and creates a zap receipt
func monitorZapInvoice(label, zapRequestJSON string) {
	restURL := utils.AppConfig.Lightning.CLNRestURL
	runeToken := utils.AppConfig.Lightning.Rune

	// Wait for invoice payment
	apiURL := fmt.Sprintf("%s/v1/waitinvoice", restURL)

	requestBody, err := json.Marshal(WaitInvoiceRequest{Label: label})
	if err != nil {
		fmt.Printf("Error marshaling wait request: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error creating wait request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Rune", runeToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error waiting for invoice: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Wait invoice failed with status: %d\n", resp.StatusCode)
		return
	}

	// Parse the wait response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading wait response: %v\n", err)
		return
	}

	var waitResp WaitInvoiceResponse
	if err := json.Unmarshal(body, &waitResp); err != nil {
		fmt.Printf("Error parsing wait response: %v\n", err)
		return
	}

	// Create and publish zap receipt
	if err := createZapReceipt(zapRequestJSON, waitResp); err != nil {
		fmt.Printf("Error creating zap receipt: %v\n", err)
	}
}

// createZapReceipt creates and publishes a Nostr zap receipt (kind 9735)
func createZapReceipt(zapRequestJSON string, payment WaitInvoiceResponse) error {
	// Parse the original zap request
	var zapRequest map[string]interface{}
	if err := json.Unmarshal([]byte(zapRequestJSON), &zapRequest); err != nil {
		return fmt.Errorf("failed to parse zap request: %w", err)
	}

	// Extract relevant information from zap request
	tags := [][]string{}

	// Add bolt11 tag
	tags = append(tags, []string{"bolt11", getBolt11FromPayment(payment)})

	// Add description tag (the original zap request)
	tags = append(tags, []string{"description", zapRequestJSON})

	// Add preimage if available
	if payment.Preimage != "" {
		tags = append(tags, []string{"preimage", payment.Preimage})
	}

	// Copy relevant tags from the original zap request
	if zapTags, ok := zapRequest["tags"].([]interface{}); ok {
		for _, tag := range zapTags {
			if tagArray, ok := tag.([]interface{}); ok && len(tagArray) >= 2 {
				tagName, _ := tagArray[0].(string)

				// Copy p, e, a tags from the zap request
				if tagName == "p" || tagName == "e" || tagName == "a" {
					stringTag := make([]string, len(tagArray))
					for i, v := range tagArray {
						stringTag[i], _ = v.(string)
					}
					tags = append(tags, stringTag)
				}
			}
		}
	}

	// Add P tag for the zap sender (if we can extract it)
	if senderPubkey, ok := zapRequest["pubkey"].(string); ok {
		tags = append(tags, []string{"P", senderPubkey})
	}

	// Create the zap receipt event using lightning service keys
	zapReceipt, err := createLightningEvent(9735, "", tags)
	if err != nil {
		return fmt.Errorf("failed to create zap receipt event: %w", err)
	}

	// Extract relays from the zap request to know where to publish
	relays := extractRelaysFromZapRequest(zapRequest)

	// Add configured lightning relays
	lightningRelays := GetLightningRelays()
	relays = append(relays, lightningRelays...)

	// Remove duplicates
	relays = removeDuplicateRelays(relays)

	// Publish the zap receipt to the specified relays
	sendLightningEventToRelays(zapReceipt, relays)

	fmt.Printf("Zap receipt created and published: %s\n", zapReceipt.ID)
	return nil
}

// getBolt11FromPayment extracts the bolt11 invoice from payment info
func getBolt11FromPayment(payment WaitInvoiceResponse) string {
	// Return the stored bolt11 for this label
	if bolt11, exists := invoiceStore[payment.Label]; exists {
		// Clean up the stored invoice
		delete(invoiceStore, payment.Label)
		return bolt11
	}
	return ""
}

// extractRelaysFromZapRequest extracts relay list from the zap request
func extractRelaysFromZapRequest(zapRequest map[string]interface{}) []string {
	relays := []string{}

	if zapTags, ok := zapRequest["tags"].([]interface{}); ok {
		for _, tag := range zapTags {
			if tagArray, ok := tag.([]interface{}); ok && len(tagArray) >= 2 {
				if tagName, _ := tagArray[0].(string); tagName == "relays" {
					// Extract all relay URLs from the relays tag
					for i := 1; i < len(tagArray); i++ {
						if relay, ok := tagArray[i].(string); ok {
							relays = append(relays, relay)
						}
					}
					break
				}
			}
		}
	}

	return relays
}

// removeDuplicateRelays removes duplicate relays from a slice
func removeDuplicateRelays(relays []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, relay := range relays {
		if !keys[relay] {
			keys[relay] = true
			result = append(result, relay)
		}
	}

	return result
}
