package nostr

import (
	"encoding/json"
	"log"
	"os"
)

// Broadcasts a Nostr event for stream end
func BroadcastNostrEndEvent(metadataFile string) {
	metadata, err := os.ReadFile(metadataFile)
	if err != nil {
		log.Printf("Error reading metadata file: %v", err)
		return
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(metadata, &meta); err != nil {
		log.Printf("Error parsing metadata JSON: %v", err)
		return
	}

	dtag, ok := meta["dtag"].(string)
	if !ok {
		log.Printf("Error: dtag not found or not a string")
		return
	}
	title, ok := meta["title"].(string)
	if !ok {
		log.Printf("Error: title not found or not a string")
		return
	}
	summary, ok := meta["summary"].(string)
	if !ok {
		log.Printf("Error: summary not found or not a string")
		return
	}
	image, ok := meta["image"].(string)
	if !ok {
		log.Printf("Error: image not found or not a string")
		return
	}
	streamURL, ok := meta["stream_url"].(string)
	if !ok {
		log.Printf("Error: stream_url not found or not a string")
		return
	}
	recordingURL, ok := meta["recording_url"].(string)
	if !ok {
		log.Printf("Error: recording_url not found or not a string")
		return
	}
	starts, ok := meta["starts"].(string) // Changed to string
	if !ok {
		log.Printf("Error: starts not found or not a string")
		return
	}
	ends, ok := meta["ends"].(string) // Changed to string
	if !ok {
		log.Printf("Error: ends not found or not a string")
		return
	}

	tags := [][]string{
		{"d", dtag},
		{"title", title},
		{"summary", summary},
		{"image", image},
		{"streaming", streamURL},
		{"recording", recordingURL},
		{"starts", starts},
		{"ends", ends},
		{"status", "ended"},
	}

	// Handle tags safely
	if rawTags, ok := meta["tags"].([]interface{}); ok {
		for _, tag := range rawTags {
			if tagStr, valid := tag.(string); valid {
				tags = append(tags, []string{"t", tagStr})
			}
		}
	}

	event, err := createEvent(30311, "", tags)
	if err != nil {
		log.Printf("Error creating end event: %v", err)
		return
	}

	sendEvent(event)
}
