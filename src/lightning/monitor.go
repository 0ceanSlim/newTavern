package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"goFrame/src/utils"
)

// WaitInvoiceResponse represents the response from CLN's waitinvoice
type WaitInvoiceResponse struct {
	Label       string `json:"label"`
	Status      string `json:"status"`
	PaymentHash string `json:"payment_hash"`
	AmountMsat  int64  `json:"amount_msat"`
	PaidAt      int64  `json:"paid_at"`
	Preimage    string `json:"payment_preimage"` // Note: CLN uses "payment_preimage" not "preimage"
}

// MonitorZapPayment monitors a zap invoice for payment and creates zap receipt when paid
func MonitorZapPayment(label, zapRequestJSON, bolt11 string) {
	log.Printf("Starting to monitor zap payment for label: %s", label)

	// Wait for the invoice to be paid
	paymentInfo, err := waitForInvoicePayment(label)
	if err != nil {
		log.Printf("Error waiting for zap payment %s: %v", label, err)
		return
	}

	if paymentInfo.Status != "paid" {
		log.Printf("Zap invoice %s was not paid (status: %s)", label, paymentInfo.Status)
		return
	}

	log.Printf("Zap invoice %s was paid! Creating zap receipt...", label)

	// Create and publish zap receipt
	err = CreateAndPublishZapReceipt(zapRequestJSON, bolt11, paymentInfo)
	if err != nil {
		log.Printf("Error creating zap receipt for %s: %v", label, err)
		return
	}

	log.Printf("Zap receipt created and published for %s", label)
}

// waitForInvoicePayment waits for an invoice to be paid using CLN's waitinvoice
func waitForInvoicePayment(label string) (*WaitInvoiceResponse, error) {
	restURL := utils.AppConfig.Lightning.CLNRestURL
	runeToken := utils.AppConfig.Lightning.Rune

	// Construct API URL
	apiURL := fmt.Sprintf("%s/v1/waitinvoice", restURL)

	// Create request payload
	requestData := map[string]string{
		"label": label,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Rune", runeToken)

	// Create client with longer timeout for waitinvoice
	client := &http.Client{
		Timeout: 5 * time.Minute, // waitinvoice can take a while
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code - Accept both 200 and 201 as success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("waitinvoice failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var waitResp WaitInvoiceResponse
	if err := json.Unmarshal(body, &waitResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nResponse: %s", err, string(body))
	}

	log.Printf("Invoice %s status: %s, paid_at: %d", label, waitResp.Status, waitResp.PaidAt)
	return &waitResp, nil
}
