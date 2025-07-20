package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	Bolt11 string `json:"bolt11"`
}

// InvoiceResult contains both the invoice and the label used
type InvoiceResult struct {
	Bolt11 string
	Label  string
}

// FetchInvoice requests an invoice from CLN REST and returns both invoice and label
func FetchInvoice(amountMsats int64, description string) (string, error) {
	result, err := FetchInvoiceWithLabel(amountMsats, description)
	if err != nil {
		return "", err
	}
	return result.Bolt11, nil
}

// FetchInvoiceWithLabel requests an invoice from CLN REST and returns both invoice and label
func FetchInvoiceWithLabel(amountMsats int64, description string) (*InvoiceResult, error) {
	restURL := utils.AppConfig.Lightning.CLNRestURL
	runeToken := utils.AppConfig.Lightning.Rune

	// Construct API URL
	apiURL := fmt.Sprintf("%s/v1/invoice", restURL)

	// Generate a unique label using timestamp
	label := fmt.Sprintf("lnurl-%d-%d", amountMsats, time.Now().UnixNano())

	// Construct request payload
	requestBody, err := json.Marshal(InvoiceRequest{
		AmountMsat:  amountMsats,
		Label:       label,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (authorization uses Rune)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Rune", runeToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var response InvoiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nResponse: %s", err, string(body))
	}

	return &InvoiceResult{
		Bolt11: response.Bolt11,
		Label:  label,
	}, nil
}
