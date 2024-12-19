package utils

import (
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
			handleExistingFiles()
			startHLSStream(rtmpStreamURL)
			waitForStreamToStop(rtmpStreamURL)
			stopHLSStream()
		}
		time.Sleep(5 * time.Second) // Check every 5 seconds
	}
}

func isStreamActive(url string) bool {
	cmd := exec.Command("ffprobe", "-i", url, "-show_streams", "-select_streams", "v", "-show_entries", "stream=codec_name", "-of", "default=nw=1", "-v", "quiet")
	err := cmd.Run()
	return err == nil // If ffprobe succeeds, the stream is live
}

func handleExistingFiles() {
	dateFolder := time.Now().Format("2006-01-02_15-04-05")
	targetDir := filepath.Join("web/live", dateFolder)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create folder for old streams: %v", err)
	}

	files, err := filepath.Glob("web/live/*")
	if err != nil {
		log.Fatalf("Failed to list files in live directory: %v", err)
	}

	for _, file := range files {
		if !isFile(file) {
			continue
		}
		err := os.Rename(file, filepath.Join(targetDir, filepath.Base(file)))
		if err != nil {
			log.Printf("Failed to move file %s: %v", file, err)
		}
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func startHLSStream(rtmpStreamURL string) {
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
		"web/live/output.m3u8",
	)
	err := ffmpegCmd.Start()
	if err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("HLS stream started.")
}

func waitForStreamToStop(rtmpStreamURL string) {
	for {
		if !isStreamActive(rtmpStreamURL) {
			log.Println("Stream stopped.")
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func stopHLSStream() {
	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
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

// ServeHLSFolderWithCORS serves the HLS files with CORS headers.
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
