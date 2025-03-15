package stream

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func startHLSStream() {
	log.Println("Starting HLS stream with metadata...")

	metadataMutex.Lock()
	defer metadataMutex.Unlock()

	// Update the last applied metadata hash
	lastMetadata = getMetadataHash()

	// Create a new output directory with timestamp to avoid conflicts
	outputDir := "web/live"
	outputFile := filepath.Join(outputDir, "output.m3u8")

	// Ensure the directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Convert tags slice to comma-separated string for metadata
	tagString := ""
	if len(metadataConfig.Tags) > 0 {
		tagBytes, err := json.Marshal(metadataConfig.Tags)
		if err == nil {
			tagString = string(tagBytes)
		}
	}

	ffmpegCmd = exec.Command("ffmpeg",
		"-i", streamConfig.RTMPStreamURL,
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-b:a", "160k",
		"-metadata", fmt.Sprintf("dtag=%s", metadataConfig.Dtag),
		"-metadata", fmt.Sprintf("pubkey=%s", metadataConfig.Pubkey),
		"-metadata", fmt.Sprintf("title=%s", metadataConfig.Title),
		"-metadata", fmt.Sprintf("summery=%s", metadataConfig.Summery),
		"-metadata", fmt.Sprintf("image=%s", metadataConfig.Image),
		"-metadata", fmt.Sprintf("tags=%s", tagString),
		"-metadata", fmt.Sprintf("stream_url=%s", metadataConfig.StreamURL),
		"-metadata", fmt.Sprintf("recording_url=%s", metadataConfig.RecordingURL),
		"-metadata", fmt.Sprintf("starts=%s", metadataConfig.Starts),
		"-metadata", fmt.Sprintf("ends=%s", metadataConfig.Ends),
		"-metadata", fmt.Sprintf("status=%s", metadataConfig.Status),
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		outputFile,
	)

	// Start the FFmpeg process
	if err := ffmpegCmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("HLS stream started with updated metadata.")
}
