package routes

import (
	"net/http"

	"goFrame/src/utils"
)

// LiveView serves an HTML page to view the HLS stream.
func LiveView(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Live Stream Debug View",
	}

	// Set up custom data for the video player
	data.CustomData = map[string]interface{}{
		"StreamURL": "/live/output.m3u8",
	}

	// Render the template
	utils.RenderTemplate(w, data, "live-view.html", false)
}
