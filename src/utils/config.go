package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
    Port        int    `yaml:"port"`
    CLNRestURL  string `yaml:"cln_rest_url"`
    RuneToken   string `yaml:"rune_token"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &config, nil
}
