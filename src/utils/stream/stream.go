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
				metadataConfig.Summery = updatedMetadata.Summery
				metadataConfig.Image = updatedMetadata.Image
				metadataConfig.Tags = updatedMetadata.Tags
				metadataMutex.Unlock()

				// Save the updated metadata to JSON
				if err := saveMetadata(metadataFile); err != nil {
					log.Printf("Failed to save updated metadata: %v", err)
				}

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
