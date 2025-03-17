package nostr

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// Broadcasts a Nostr event for stream start
func BroadcastNostrStartEvent(metadataFile string) {
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
	startsFloat, ok := meta["starts"].(float64)
	if !ok {
		log.Printf("Error: starts not found or not a number")
		return
	}
	endsFloat, ok := meta["ends"].(float64)
	if !ok {
		log.Printf("Error: ends not found or not a number")
		return
	}
	status, ok := meta["status"].(string)
	if !ok {
		log.Printf("Error: status not found or not a string")
		return
	}

	starts := strconv.FormatInt(int64(startsFloat), 10)
	ends := strconv.FormatInt(int64(endsFloat), 10)

	tags := [][]string{
		{"d", dtag},
		{"title", title},
		{"summary", summary},
		{"image", image},
		{"streaming", streamURL},
		{"recording", recordingURL},
		{"starts", starts},
		{"ends", ends},
		{"status", status},
	}

	// Add tags dynamically from metadata
	if tagList, ok := meta["tags"].([]interface{}); ok {
		for _, tag := range tagList {
			if tagStr, valid := tag.(string); valid {
				tags = append(tags, []string{"t", tagStr})
			}
		}
	}

	event, err := createEvent(30311, "", tags)
	if err != nil {
		log.Printf("Error creating live event: %v", err)
		return
	}

	sendEvent(event)
}
