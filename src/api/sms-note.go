package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// SMSMessage represents the expected incoming SMS format
type SMSMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

// smsLogFile is the file where messages will be logged
const smsLogFile = "sms_log.txt"

// SMSHandler logs incoming SMS messages
func SMSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var sms SMSMessage
	if err := json.NewDecoder(r.Body).Decode(&sms); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logEntry := fmt.Sprintf("%s - From: %s | Message: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), sms.From, sms.Message)

	// Append log to file
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
	w.Write([]byte("SMS logged successfully"))
}
