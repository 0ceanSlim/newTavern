package stream

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"time"
)

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
