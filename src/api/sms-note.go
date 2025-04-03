package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// WebhookPayload matches Textbelt's JSON structure
type WebhookPayload struct {
	TextID     string `json:"textId"`
	FromNumber string `json:"fromNumber"`
	Text       string `json:"text"`
}

const smsLogFile = "sms_log.txt"

func SMSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	logEntry := fmt.Sprintf("%s - TextID: %s | From: %s | Message: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), payload.TextID, payload.FromNumber, payload.Text)

	// Append to log file
	file, err := os.OpenFile(smsLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Could not write to log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		http.Error(w, "Failed to write log", http.StatusInternalServerError)
		return
	}

	log.Println("Logged SMS:", logEntry)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SMS reply logged successfully"))
}
