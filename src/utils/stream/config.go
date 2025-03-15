package stream

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"gopkg.in/yaml.v3"
)

type StreamConfig struct {
	RTMPStreamURL string `yaml:"rtmp_stream_url"`
}

type MetadataConfig struct {
	Title        string   `yaml:"title"`         //Title of Stream
	Summery      string   `yaml:"summery"`       //Summery of Stream
	Image        string   `yaml:"image"`         // url of the Stream Thumbnail
	Tags         []string `yaml:"tags"`          // aray of tags [t] in stream event
	Pubkey       string   `yaml:"pubkey"`        // author pubkey of stream
	Dtag         string   `yaml:"dtag"`          //dtag (unique identifier) of the stream, stays the same through updates
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
