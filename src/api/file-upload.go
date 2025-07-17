package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
)

type UploadResponse struct {
	URL   string `json:"url"`
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// HandleFileUpload proxies file uploads to 0x0.st
func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form (32MB max memory)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeErrorResponse(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeErrorResponse(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size (512MB limit)
	maxSize := int64(512 * 1024 * 1024) // 512MB
	if fileHeader.Size > maxSize {
		writeErrorResponse(w, "File too large. Maximum size is 512MB", http.StatusBadRequest)
		return
	}

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		writeErrorResponse(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Create a new form for 0x0.st
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file to the form
	fileWriter, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		writeErrorResponse(w, "Failed to create form file", http.StatusInternalServerError)
		return
	}

	_, err = fileWriter.Write(fileContent)
	if err != nil {
		writeErrorResponse(w, "Failed to write file data", http.StatusInternalServerError)
		return
	}

	// Add secret field for harder-to-guess URLs
	err = writer.WriteField("secret", "")
	if err != nil {
		writeErrorResponse(w, "Failed to add secret field", http.StatusInternalServerError)
		return
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		writeErrorResponse(w, "Failed to finalize form", http.StatusInternalServerError)
		return
	}

	// Create request to 0x0.st
	req, err := http.NewRequest("POST", "https://0x0.st", &requestBody)
	if err != nil {
		writeErrorResponse(w, "Failed to create upload request", http.StatusInternalServerError)
		return
	}

	// Set the content type with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "HappyTavern/1.0 (File Upload Service)")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeErrorResponse(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeErrorResponse(w, "Failed to read upload response", http.StatusInternalServerError)
		return
	}

	// Check if upload was successful
	if resp.StatusCode != http.StatusOK {
		writeErrorResponse(w, "Upload failed: "+string(responseBody), resp.StatusCode)
		return
	}

	// Get the management token from response headers
	token := resp.Header.Get("X-Token")

	// Prepare success response
	uploadResponse := UploadResponse{
		URL:   string(responseBody),
		Token: token,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uploadResponse)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := UploadResponse{
		Error: message,
	}
	json.NewEncoder(w).Encode(errorResponse)
}
