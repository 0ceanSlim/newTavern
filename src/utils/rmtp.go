package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type StreamConfig struct {
	RTMPStreamURL string `yaml:"rtmp_stream_url"`
}

type MetadataConfig struct {
	Dtag 		string `yaml:"dtag"` //dtag (unique identifier) of the stream, stays the same through updates
	Pubkey 	string `yaml:"pubkey"`// author pubkey of stream
	Title       string `yaml:"title"` //Title of Stream
	Summery string `yaml:"summery"` //Summery of Stream
	Image 	 string `yaml:"image"` // url of the Stream Thumbnail
	Tags 	 []string `yaml:"tags"` // aray of tags [t] in stream event
	StreamURL string `yaml:"stream_url"` // always https://happytavern.co/live/output.m3u8
	RecordingURL string `yaml:"recording_url"`// url of the stream recording when handle existing files is called
	Starts string `yaml:"starts"` //unix stamp when stream starts
	Ends string `yaml:"ends"` //unix stamp when stream stops
	Status string `yaml:"status"` //planned, live, ended
}

var (
	ffmpegCmd       *exec.Cmd
	streamConfig    StreamConfig
	metadataConfig  MetadataConfig
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
	return fmt.Sprintf("%d", rand.Intn(900000) + 100000)
}

func MonitorStream() {
	if err := LoadStreamConfig("config.yml"); err != nil {
		log.Fatalf("Error loading stream config: %v", err)
	}
	if err := LoadMetadataConfig("stream.yml"); err != nil {
		log.Fatalf("Error loading metadata config: %v", err)
	}

	metadataConfig.Dtag = generateDtag()
	metadataConfig.Starts = fmt.Sprintf("%d", time.Now().Unix())
	metadataConfig.Status = "live"
	SaveMetadataConfig("stream.yml")

	for {
		if isStreamActive(streamConfig.RTMPStreamURL) {
			log.Println("Stream detected, starting HLS process...")
			startHLSStream()
			watchMetadataChanges()
			waitForStreamToStop()
			handleExistingFiles()
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
		//log.Printf("ffprobe output: %s", string(output))
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

func watchMetadataChanges() {
	for {
		time.Sleep(10 * time.Second)
		LoadMetadataConfig("stream.yml")
	}
}

func startHLSStream() {
	log.Println("Starting HLS stream with metadata...")

	// Update metadata before starting the stream
	metadataConfig.Dtag = generateDtag()
	metadataConfig.Starts = fmt.Sprintf("%d", time.Now().Unix())
	metadataConfig.Status = "live"
	metadataConfig.RecordingURL = fmt.Sprintf("https://happytavern.co/.videos/past-streams/%s-%s", time.Now().Format("1-2-2006"), metadataConfig.Dtag)
	SaveMetadataConfig("stream.yml")

	// Start watching metadata changes in a goroutine
	go watchMetadataChanges()

	encodeHLSStream()
}

func encodeHLSStream() {
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
		"-metadata", fmt.Sprintf("tags=%s", metadataConfig.Tags),
		"-metadata", fmt.Sprintf("stream_url=%s", metadataConfig.StreamURL),
		"-metadata", fmt.Sprintf("recording_url=%s", metadataConfig.RecordingURL),
		"-metadata", fmt.Sprintf("starts=%s", metadataConfig.Starts),
		"-metadata", fmt.Sprintf("ends=%s", metadataConfig.Ends),
		"-metadata", fmt.Sprintf("status=%s", metadataConfig.Status),
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"web/live/output.m3u8",
	)
	if err := ffmpegCmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}
	log.Println("HLS stream started.")
}


func waitForStreamToStop() {
	inactiveChecks := 0
	const maxInactiveChecks = 3

	for {
		if !isStreamActive(streamConfig.RTMPStreamURL) {
			inactiveChecks++
			if inactiveChecks >= maxInactiveChecks {
				log.Println("Stream stopped. Stopping HLS stream...")
				stopHLSStream()
				break
			}
		} else {
			inactiveChecks = 0
		}
		time.Sleep(5 * time.Second)
	}
}

func stopHLSStream() {
	if ffmpegCmd != nil && ffmpegCmd.Process != nil {
		log.Println("Stopping HLS stream...")
		metadataConfig.Status = "ended"
		metadataConfig.Ends = fmt.Sprintf("%d", time.Now().Unix())
		SaveMetadataConfig("stream.yml")

		encodeHLSStream()

		err := ffmpegCmd.Process.Kill()
		if err != nil {
			log.Printf("Failed to stop FFmpeg process: %v", err)
		} else {
			log.Println("FFmpeg process stopped.")
		}
		ffmpegCmd = nil
	}

	handleExistingFiles()
}

func handleExistingFiles() {
	log.Println("Archiving existing files...")
	archiveFolder := fmt.Sprintf("web/.videos/past-streams/%s-%s", time.Now().Format("1-2-2006"), metadataConfig.Dtag)

	if err := os.MkdirAll(archiveFolder, os.ModePerm); err != nil {
		log.Fatalf("Failed to create archive folder: %v", err)
	}

	files, err := filepath.Glob("web/live/*")
	if err != nil {
		log.Fatalf("Failed to list files in live directory: %v", err)
	}

	for _, file := range files {
		err := os.Rename(file, filepath.Join(archiveFolder, filepath.Base(file)))
		if err != nil {
			log.Printf("Failed to move file %s: %v", file, err)
		} else {
			log.Printf("Moved file %s to %s", file, archiveFolder)
		}
	}

	log.Println("Archiving completed.")
}
