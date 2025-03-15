package api

import (
	"encoding/json"
	"net/http"
	"os"

	"goFrame/src/utils/stream"
)

// GetStreamMetadata fetches metadata from JSON instead of YAML
func GetStreamMetadata(w http.ResponseWriter, r *http.Request) {
	metadataFile := "web/live/metadata.json"

	// Check if metadata file exists
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		// Return default "offline" response
		defaultMetadata := stream.MetadataConfig{
			Status: "offline",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(defaultMetadata)
		return
	}

	// Load metadata from JSON
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		http.Error(w, "Failed to read stream metadata", http.StatusInternalServerError)
		return
	}

	// Parse JSON into MetadataConfig
	var metadata stream.MetadataConfig
	if err := json.Unmarshal(data, &metadata); err != nil {
		http.Error(w, "Failed to parse stream metadata", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
