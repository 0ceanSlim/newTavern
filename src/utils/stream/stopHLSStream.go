package stream

import (
	"fmt"
	"log"
	"time"
)

// stopHLSStream terminates the FFmpeg process and archives the stream files
func stopHLSStream() {
	log.Println("stopHLSStream: Beginning stream shutdown process...")

	metadataMutex.Lock()
	defer metadataMutex.Unlock()

	// Update metadata for stream end
	log.Println("Updating metadata for stream end...")
	metadataConfig.Status = "ended"
	metadataConfig.Ends = fmt.Sprintf("%d", time.Now().Unix())

	// Save metadata to YAML
	err := SaveMetadataConfig("stream.yml")
	if err != nil {
		log.Printf("Error saving metadata to YAML: %v", err)
	}

	// Save metadata to JSON before archiving
	metadataFile := "web/live/metadata.json"
	err = saveMetadata(metadataFile)
	if err != nil {
		log.Printf("Error saving metadata to JSON: %v", err)
	}

	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
		log.Println("Stopping FFmpeg process...")

		// Kill the FFmpeg process
		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Printf("Failed to stop FFmpeg process: %v", err)
		} else {
			log.Println("FFmpeg process kill signal sent successfully.")
		}

		// Wait for the process to fully terminate
		log.Println("Waiting for FFmpeg process to fully terminate...")
		err = ffmpegCmd.Wait()
		if err != nil {
			log.Printf("FFmpeg wait error: %v", err)
		}
		log.Println("FFmpeg process fully terminated.")
		ffmpegCmd = nil
	} else {
		log.Println("No active FFmpeg process found to terminate.")
	}

	// Archive the stream files
	log.Println("Beginning archive process for stream files...")
	archiveStream()
}
