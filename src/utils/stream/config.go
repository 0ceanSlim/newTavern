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
	Title        string   `yaml:"title" json:"title"`
	Summary      string   `yaml:"summary" json:"summary"`
	Image        string   `yaml:"image" json:"image"`
	Tags         []string `yaml:"tags" json:"tags"`
	Pubkey       string   `yaml:"pubkey" json:"pubkey"`
	Dtag         string   `yaml:"dtag" json:"dtag"`
	StreamURL    string   `yaml:"stream_url" json:"stream_url"`
	RecordingURL string   `yaml:"recording_url" json:"recording_url"`
	Starts       string   `yaml:"starts" json:"starts"`
	Ends         string   `yaml:"ends" json:"ends"`
	Status       string   `yaml:"status" json:"status"`
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
