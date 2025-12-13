package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func CheckNameHandler(w http.ResponseWriter, r *http.Request) {
	// Get the name from the query parameters
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	logFile := "web/.well-known/nostr.json"

	// Read the JSON file
	file, err := os.ReadFile(logFile)
	if err != nil {
		http.Error(w, "Could not read log file", http.StatusInternalServerError)
		return
	}

	// Parse JSON into a struct
	var logs struct {
		Names map[string]string `json:"names"`
	}

	if err := json.Unmarshal(file, &logs); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusInternalServerError)
		return
	}

	// Case-insensitive name lookup
	nameLower := strings.ToLower(name)
	for existingName := range logs.Names {
		if strings.ToLower(existingName) == nameLower {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `<p class="text-red-500">❌ Name is already taken</p>`)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<p class="text-green-500">✅ Name is available</p>`)
}
