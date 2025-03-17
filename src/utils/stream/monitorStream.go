package stream

import (
	"fmt"
	"log"
	"time"

	"goFrame/src/utils/stream/nostr"
)

// MonitorStream is the main function that handles the streaming process
func MonitorStream() {
	if err := LoadStreamConfig("config.yml"); err != nil {
		log.Fatalf("Error loading stream config: %v", err)
	}
	if err := LoadMetadataConfig("stream.yml"); err != nil {
		log.Fatalf("Error loading metadata config: %v", err)
	}

	for {
		rtmpURL := streamConfig.RTMPStreamURL

		if isStreamActive(rtmpURL) {
			log.Println("Stream detected, starting HLS process...")

			metadataMutex.Lock()
			// Initialize stream metadata if this is a new stream
			if metadataConfig.Status != "live" {
				metadataConfig.Dtag = generateDtag()
				metadataConfig.Ends = ""
				metadataConfig.Starts = fmt.Sprintf("%d", time.Now().Unix())
				metadataConfig.Status = "live"
				metadataConfig.StreamURL = ("https://happytavern.co/live/output.m3u8")
				metadataConfig.RecordingURL = fmt.Sprintf("https://happytavern.co/.videos/past-streams/%s-%s",
					time.Now().Format("1-2-2006"), metadataConfig.Dtag)

				// Save to JSON immediately
				saveMetadata("web/live/metadata.json")
				log.Printf("MonitorStream: Initial StreamURL: %s", metadataConfig.StreamURL) //add log
				nostr.BroadcastNostrStartEvent("web/live/metadata.json")
			}
			metadataMutex.Unlock()

			// Create a channel to signal the metadata watcher to stop
			stopWatcher := make(chan bool)

			// Start watching metadata changes in a goroutine
			go watchMetadata(stopWatcher) //add the nostr broadcast update to this function

			// Start encoding the stream
			startHLSStream()

			// Wait for the stream to stop
			log.Println("Beginning to monitor stream for inactivity...")
			waitForStreamToStop(rtmpURL)

			log.Println("Stream has been detected as inactive, beginning shutdown sequence...")

			// Signal metadata watcher to stop
			stopWatcher <- true

			// Stop the HLS stream and perform cleanup
			log.Println("Calling stopHLSStream function...")
			stopHLSStream()

			log.Println("Stream shutdown sequence completed")

		}

		time.Sleep(5 * time.Second)
	}
}

// waitForStreamToStop monitors the RTMP stream and returns when it detects the stream has stopped
func waitForStreamToStop(rtmpStreamURL string) {
	consecutiveInactiveChecks := 0
	const maxInactiveChecks = 3 // Require 3 consecutive inactive checks to confirm stop

	for {
		active := isStreamActive(rtmpStreamURL)

		if !active {
			consecutiveInactiveChecks++
			log.Printf("Inactive check %d/%d", consecutiveInactiveChecks, maxInactiveChecks)

			if consecutiveInactiveChecks >= maxInactiveChecks {
				log.Println("Stream stopped after multiple inactive checks. Proceeding to cleanup...")
				return // Exit the function to handle stream stop
			}
		} else {
			if consecutiveInactiveChecks > 0 {
				log.Println("Stream is active again, resetting inactive checks.")
			}
			consecutiveInactiveChecks = 0 // Reset if the stream becomes active again
		}

		time.Sleep(5 * time.Second)
	}
}
