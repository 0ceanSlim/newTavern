package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type StreamConfig struct {
	RTMPStreamURL string `yaml:"rtmp_stream_url"`
}

type MetadataConfig struct {
	Dtag         string   `yaml:"dtag"`          //dtag (unique identifier) of the stream, stays the same through updates
	Pubkey       string   `yaml:"pubkey"`        // author pubkey of stream
	Title        string   `yaml:"title"`         //Title of Stream
	Summery      string   `yaml:"summery"`       //Summery of Stream
	Image        string   `yaml:"image"`         // url of the Stream Thumbnail
	Tags         []string `yaml:"tags"`          // aray of tags [t] in stream event
	StreamURL    string   `yaml:"stream_url"`    // always https://happytavern.co/live/output.m3u8
	RecordingURL string   `yaml:"recording_url"` // url of the stream recording when handle existing files is called
	Starts       string   `yaml:"starts"`        //unix stamp when stream starts
	Ends         string   `yaml:"ends"`          //unix stamp when stream stops
	Status       string   `yaml:"status"`        //planned, live, ended
}

var (
	ffmpegCmd      *exec.Cmd
	streamConfig   StreamConfig
	metadataConfig MetadataConfig
	metadataMutex  sync.Mutex
	lastMetadata   string // Store a hash or string representation of the last applied metadata
)

func LoadStreamConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return yaml.Unmarshal(data, &streamConfig)
}

func LoadMetadataConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %w", err)
	}
	return yaml.Unmarshal(data, &metadataConfig)
}

func GetMetadataConfig() MetadataConfig {
	return metadataConfig
}

func SaveMetadataConfig(path string) error {
	data, err := yaml.Marshal(&metadataConfig)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func generateDtag() string {
	return fmt.Sprintf("%d", rand.Intn(900000)+100000)
}

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
				metadataConfig.Starts = fmt.Sprintf("%d", time.Now().Unix())
				metadataConfig.Status = "live"
				metadataConfig.RecordingURL = fmt.Sprintf("https://happytavern.co/.videos/past-streams/%s-%s",
					time.Now().Format("1-2-2006"), metadataConfig.Dtag)
				SaveMetadataConfig("stream.yml")
			}
			metadataMutex.Unlock()

			// Create a channel to signal the metadata watcher to stop
			stopWatcher := make(chan bool)

			// Start watching metadata changes in a goroutine
			go watchMetadataChanges(stopWatcher)

			// Start encoding the stream
			encodeHLSStream()

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

func isStreamActive(url string) bool {
	//log.Printf("Checking stream status for URL: %s", url)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5-second timeout
	defer cancel()

	// Execute ffprobe with the context
	cmd := exec.CommandContext(ctx, "ffprobe", "-i", url, "-show_streams", "-select_streams", "v", "-show_entries", "stream=codec_name", "-of", "json", "-v", "quiet")
	output, err := cmd.CombinedOutput()

	// Check for timeout or other errors
	if ctx.Err() == context.DeadlineExceeded {
		//log.Printf("ffprobe timed out while checking stream: %s", url)
		return false
	}

	if err != nil {
		log.Printf("ffprobe error: %v", err)
		log.Printf("ffprobe output: %s", string(output))
		return false
	}

	//log.Printf("ffprobe output: %s", string(output))

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
			//log.Printf("Video stream detected with codec: %s", codecName)
			return true
		}
	}

	//log.Println("No active video stream found.")
	return false
}

func encodeHLSStream() {
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

func watchMetadataChanges(stopChan chan bool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Get initial metadata hash
	lastMetadata = getMetadataHash()

	for {
		select {
		case <-ticker.C:
			// Load updated metadata
			metadataMutex.Lock()
			if err := LoadMetadataConfig("stream.yml"); err != nil {
				log.Printf("Error loading metadata config: %v", err)
				metadataMutex.Unlock()
				continue
			}

			// Check if relevant metadata has changed
			newHash := getMetadataHash()
			if newHash != lastMetadata {
				log.Println("Metadata changed, updating stream...")
				lastMetadata = newHash

				// Only if stream is already running
				if ffmpegCmd != nil && ffmpegCmd.Process != nil {
					// Create a new FFmpeg process with updated metadata
					oldCmd := ffmpegCmd

					// Start new process with updated metadata
					encodeHLSStream()

					// Give the new process a moment to start
					time.Sleep(2 * time.Second)

					// Terminate the old process
					if oldCmd != nil && oldCmd.Process != nil {
						log.Println("Stopping old FFmpeg process...")
						if err := oldCmd.Process.Kill(); err != nil {
							log.Printf("Failed to stop old FFmpeg process: %v", err)
						}
						oldCmd.Wait() // Wait for it to fully terminate
					}
				}
			}
			metadataMutex.Unlock()

		case <-stopChan:
			return
		}
	}
}

// Helper function to get a string representation of metadata fields we care about for stream updates
func getMetadataHash() string {
	return fmt.Sprintf("%s:%s:%s:%s:%v",
		metadataConfig.Title,
		metadataConfig.Summery,
		metadataConfig.Image,
		metadataConfig.Tags,
		metadataConfig.Status)
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

// stopHLSStream terminates the FFmpeg process and archives the stream files
func stopHLSStream() {
	log.Println("stopHLSStream: Beginning stream shutdown process...")

	metadataMutex.Lock()
	defer metadataMutex.Unlock()

	// Update metadata for stream end
	log.Println("Updating metadata for stream end...")
	metadataConfig.Status = "ended"
	metadataConfig.Ends = fmt.Sprintf("%d", time.Now().Unix())
	err := SaveMetadataConfig("stream.yml")
	if err != nil {
		log.Printf("Error saving metadata: %v", err)
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
	handleExistingFiles()
}

// handleExistingFiles archives the stream segments to a permanent location
func handleExistingFiles() {
	log.Println("handleExistingFiles: Archiving existing stream files...")

	// Create a timestamped folder for this stream's archive
	archiveFolder := fmt.Sprintf("web/.videos/past-streams/%s-%s",
		time.Now().Format("1-2-2006"), metadataConfig.Dtag)

	log.Printf("Creating archive directory: %s", archiveFolder)
	if err := os.MkdirAll(archiveFolder, os.ModePerm); err != nil {
		log.Fatalf("Failed to create archive folder: %v", err)
		return
	}

	// Find all files in the live directory
	files, err := filepath.Glob("web/live/*")
	if err != nil {
		log.Fatalf("Failed to list files in live directory: %v", err)
		return
	}

	log.Printf("Found %d files to archive", len(files))

	// Move each file to the archive location
	for _, file := range files {
		destPath := filepath.Join(archiveFolder, filepath.Base(file))
		log.Printf("Moving file from %s to %s", file, destPath)

		err := os.Rename(file, destPath)
		if err != nil {
			log.Printf("Failed to move file %s: %v", file, err)

			// Try to copy if move fails
			srcFile, err := os.Open(file)
			if err != nil {
				log.Printf("Failed to open source file %s: %v", file, err)
				continue
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				log.Printf("Failed to create destination file %s: %v", destPath, err)
				continue
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				log.Printf("Failed to copy file data: %v", err)
				continue
			}

			// Try to remove original after successful copy
			os.Remove(file)
		}
	}

	log.Println("Archiving completed successfully.")

	// Update metadata with recording URL
	metadataConfig.RecordingURL = fmt.Sprintf("https://happytavern.co/.videos/past-streams/%s-%s",
		time.Now().Format("1-2-2006"), metadataConfig.Dtag)
	SaveMetadataConfig("stream.yml")
}
