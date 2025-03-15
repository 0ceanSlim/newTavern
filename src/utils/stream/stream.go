package stream

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
					startHLSStream()

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
