package routes

import (
	"encoding/json"
	"net/http"

	"goFrame/src/utils"
)

// LiveView serves an HTML page to view the HLS stream.
func LiveView(w http.ResponseWriter, r *http.Request) {
	// Fetch stream metadata from the API
	resp, err := http.Get("http://localhost:8787/api/stream-data")
	if err != nil {
		http.Error(w, "Failed to fetch stream metadata", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var streamData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&streamData); err != nil {
		http.Error(w, "Failed to decode stream metadata", http.StatusInternalServerError)
		return
	}

	// Populate the page data
	data := utils.PageData{
		Title: "Live Stream Debug View",
		CustomData: map[string]interface{}{
			"StreamURL":    streamData["StreamURL"],
			"Title":        streamData["Title"],
			"Summery":      streamData["Summery"],
			"Image":        streamData["Image"],
			"Tags":         streamData["Tags"],
			"Status":       streamData["Status"],
			"Starts":       streamData["Starts"],
			"RecordingURL": streamData["RecordingURL"],
		},
	}

	// Render the template
	utils.RenderTemplate(w, data, "live-view.html", false)
}
