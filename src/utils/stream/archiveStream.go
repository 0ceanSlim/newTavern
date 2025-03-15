package stream

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// archiveStream archives the stream segments to a permanent location
func archiveStream() {
	log.Println("archiveStream: Archiving existing stream files...")

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
