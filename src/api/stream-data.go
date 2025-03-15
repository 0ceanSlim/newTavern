package api

import (
	"encoding/json"
	"net/http"

	"goFrame/src/utils/stream"
)

func GetStreamMetadata(w http.ResponseWriter, r *http.Request) {
	// Load latest metadata
	if err := stream.LoadMetadataConfig("stream.yml"); err != nil {
		http.Error(w, "Failed to load stream metadata", http.StatusInternalServerError)
		return
	}

	// Retrieve metadata using the getter function
	metadata := stream.GetMetadataConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
