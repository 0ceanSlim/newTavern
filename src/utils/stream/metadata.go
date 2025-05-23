package stream

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"goFrame/src/utils/stream/nostr"
)

// Watch metadata file and update JSON when changes occur
func watchMetadata(stopWatcher chan bool) {
	lastModified := time.Time{}
	metadataFile := "web/live/metadata.json"
	yamlFile := "stream.yml"

	for {
		select {
		case <-stopWatcher:
			log.Println("Stopping metadata watcher...")
			return
		default:
			info, err := os.Stat(yamlFile)
			if err != nil {
				log.Printf("Error watching metadata file: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			modTime := info.ModTime()
			if modTime.After(lastModified) {
				log.Println("Metadata file changed, updating JSON...")

				// Load updated metadata from YAML
				var updatedMetadata MetadataConfig
				if err := loadMetadata(yamlFile, &updatedMetadata); err != nil {
					log.Printf("Failed to reload metadata: %v", err)
					continue
				}

				// Only update allowed fields
				metadataMutex.Lock()
				metadataConfig.Title = updatedMetadata.Title
				metadataConfig.Summary = updatedMetadata.Summary
				metadataConfig.Image = updatedMetadata.Image
				metadataConfig.Tags = updatedMetadata.Tags
				metadataMutex.Unlock()

				// Save the updated metadata to JSON
				if err := saveMetadata(metadataFile); err != nil {
					log.Printf("Failed to save updated metadata: %v", err)
				}

				// Broadcast Nostr event for metadata update
				nostr.BroadcastNostrUpdateEvent(metadataFile)

				lastModified = modTime
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// Load metadata from YAML into a provided struct
func loadMetadata(filename string, dest *MetadataConfig) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(file, dest)
}

// Save metadata as a JSON file
func saveMetadata(filename string) error {
	metadataMap := structToMap(metadataConfig)

	lowercaseMetadata := make(map[string]interface{})
	for key, value := range metadataMap {
		lowercaseKey := strings.ToLower(key)
		lowercaseMetadata[lowercaseKey] = value
	}

	data, err := json.MarshalIndent(lowercaseMetadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
