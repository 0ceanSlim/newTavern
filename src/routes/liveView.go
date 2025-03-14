package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goFrame/src/utils"
)

// LiveView serves an HTML page to view the HLS stream.
func LiveView(w http.ResponseWriter, r *http.Request) {
	// Load the configuration to get the port
	if err := utils.LoadConfig("config.yml"); err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	// Construct API URL dynamically using the configured port
	apiURL := fmt.Sprintf("http://localhost:%d/api/stream-data", utils.AppConfig.Server.Port)

	// Fetch stream metadata from the API
	resp, err := http.Get(apiURL)
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
