package stream

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func startHLSStream() {
	log.Println("Starting HLS stream...")

	outputDir := "web/live"
	outputFile := filepath.Join(outputDir, "output.m3u8")
	metadataFile := filepath.Join(outputDir, "metadata.json")

	// Ensure the directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Save metadata as JSON
	if err := saveMetadata(metadataFile); err != nil {
		log.Fatalf("Failed to save metadata: %v", err)
	}

	// Start FFmpeg without metadata embedding
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", streamConfig.RTMPStreamURL,
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-b:a", "160k",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		outputFile,
	)

	if err := ffmpegCmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("HLS stream started.")
}
