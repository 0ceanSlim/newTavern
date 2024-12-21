package utils

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ffmpegCmd *exec.Cmd

func MonitorStream() {
	rtmpStreamURL := "rtmp://10.1.10.7/live" // Replace with your RTMP stream URL
	for {
		if isStreamActive(rtmpStreamURL) {
			log.Println("Stream detected, starting HLS process...")
			startHLSStream(rtmpStreamURL)
			waitForStreamToStop(rtmpStreamURL)
			stopHLSStream()
			handleExistingFiles()
		} else {
			log.Println("No active stream detected. Retrying...")
		}
		time.Sleep(5 * time.Second) // Check every 5 seconds
	}
}

func isStreamActive(url string) bool {
	log.Printf("Checking stream status for URL: %s", url)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5-second timeout
	defer cancel()

	// Execute ffprobe with the context
	cmd := exec.CommandContext(ctx, "ffprobe", "-i", url, "-show_streams", "-select_streams", "v", "-show_entries", "stream=codec_name", "-of", "json", "-v", "quiet")
	output, err := cmd.CombinedOutput()

	// Check for timeout or other errors
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("ffprobe timed out while checking stream: %s", url)
		return false
	}

	if err != nil {
		log.Printf("ffprobe error: %v", err)
		log.Printf("ffprobe output: %s", string(output))
		return false
	}

	log.Printf("ffprobe output: %s", string(output))

	// Check for active video stream
	return containsVideoStream(output)
}

// Helper function to parse ffprobe JSON output and check for video streams
func containsVideoStream(output []byte) bool {
	var result map[string]interface{}

	// Parse JSON output
	err := json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Failed to parse ffprobe JSON output: %v", err)
		return false
	}

	// Check if "streams" key exists and contains video streams
	streams, ok := result["streams"].([]interface{})
	if !ok || len(streams) == 0 {
		log.Println("No streams found in ffprobe output.")
		return false
	}

	// Look for a video stream
	for _, stream := range streams {
		streamMap, ok := stream.(map[string]interface{})
		if !ok {
			continue
		}
		codecName, ok := streamMap["codec_name"].(string)
		if ok && codecName != "" {
			log.Printf("Video stream detected with codec: %s", codecName)
			return true
		}
	}

	log.Println("No active video stream found.")
	return false
}

func handleExistingFiles() {
	log.Println("Archiving existing files...")
	dateFolder := time.Now().Format("2006-01-02_15-04-05")
	targetDir := filepath.Join("web/.videos/past-streams", dateFolder)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create archive folder: %v", err)
	}

	files, err := filepath.Glob("web/live/*")
	if err != nil {
		log.Fatalf("Failed to list files in live directory: %v", err)
	}

	for _, file := range files {
		err := os.Rename(file, filepath.Join(targetDir, filepath.Base(file)))
		if err != nil {
			log.Printf("Failed to move file %s: %v", file, err)
		} else {
			log.Printf("Moved file %s to %s", file, targetDir)
		}
	}
	log.Println("Archiving completed.")
}

func startHLSStream(rtmpStreamURL string) {
	log.Println("Starting HLS stream...")
	ffmpegCmd = exec.Command("ffmpeg",
		"-i", rtmpStreamURL,
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-b:a", "160k",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_flags", "delete_segments",
		"web/live/output.m3u8",
	)
	err := ffmpegCmd.Start()
	if err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("HLS stream started.")
}

func waitForStreamToStop(rtmpStreamURL string) {
	consecutiveInactiveChecks := 0
	const maxInactiveChecks = 3 // Require 5 consecutive inactive checks to confirm stop

	for {
		if !isStreamActive(rtmpStreamURL) {
			consecutiveInactiveChecks++
			log.Printf("Inactive check %d/%d", consecutiveInactiveChecks, maxInactiveChecks)
			if consecutiveInactiveChecks >= maxInactiveChecks {
				log.Println("Stream stopped after multiple inactive checks. Stopping HLS stream...")
				stopHLSStream()
				break
			}
		} else {
			consecutiveInactiveChecks = 0 // Reset if the stream becomes active again
			log.Println("Stream is active, resetting inactive checks.")
		}
		time.Sleep(5 * time.Second)
	}
}

func stopHLSStream() {
	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
		log.Println("Stopping HLS stream...")
		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Printf("Failed to stop FFmpeg process: %v", err)
		} else {
			log.Println("FFmpeg process stopped.")
		}
		ffmpegCmd = nil
	}
}

func ServeHLS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

	// Serve the HLS file
	http.ServeFile(w, r, "web/live/output.m3u8")
}

func ServeHLSFolderWithCORS(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
	})
}
