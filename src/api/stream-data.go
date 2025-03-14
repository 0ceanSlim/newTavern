package api

import (
	"encoding/json"
	"net/http"

	"goFrame/src/utils"
)

func GetStreamMetadata(w http.ResponseWriter, r *http.Request) {
	// Load latest metadata
	if err := utils.LoadMetadataConfig("stream.yml"); err != nil {
		http.Error(w, "Failed to load stream metadata", http.StatusInternalServerError)
		return
	}

	// Retrieve metadata using the getter function
	metadata := utils.GetMetadataConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
