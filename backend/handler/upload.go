package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Support JPEG decoding
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"floorplan-whiteboard/ai"
	"floorplan-whiteboard/models"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

// UploadFloorplan handles the file upload and calls the AI service
func UploadFloorplan(c *gin.Context) {
	// 1. Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 2. Read file content
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	fileBytes := buf.Bytes()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mimeType := header.Header.Get("Content-Type")

	// Validate extension/mime
	isValid := false
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		isValid = true
		if mimeType == "" {
			mimeType = "image/png" // Default fallback
		}
	} else if ext == ".pdf" {
		// DEPRECATED: PDF should be converted on frontend now.
		// But if sent, we reject it as we removed backend rendering.
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF files should be processed by client. Please retry."})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG and JPG are supported."})
		return
	}

	// 3. Call AI Service
	jsonResponse, err := ai.AnalyzeFloorplan(c.Request.Context(), fileBytes, mimeType)
	if err != nil {
		fmt.Printf("AI Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze floorplan: " + err.Error()})
		return
	}

	// 4. Process Image and Remap Coordinates
	finalImageBase64, remappedRooms, err := processAndRemap(fileBytes, jsonResponse, mimeType)
	if err != nil {
		// Send a specific error message back to the frontend
		errorMsg := "Failed to process image after analysis: " + err.Error()
		fmt.Println("Processing Error:", errorMsg) // Keep server log
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": remappedRooms,
		"image": finalImageBase64,
	})
}

// GeminiResponse matches the JSON structure returned by Gemini
type GeminiResponse struct {
	ContentBox []int        `json:"content_box"` // [ymin, xmin, ymax, xmax] 0-1000
	Rooms      []GeminiRoom `json:"rooms"`
}

func processAndRemap(fileBytes []byte, jsonStr string, mimeType string) (string, []models.Room, error) {
	// A. Parse JSON
	var aiData GeminiResponse
	cleanJson := strings.ReplaceAll(jsonStr, "```json", "")
	cleanJson = strings.ReplaceAll(cleanJson, "```", "")

	if err := json.Unmarshal([]byte(cleanJson), &aiData); err != nil {
		return "", nil, fmt.Errorf("json parse error: %w", err)
	}

	if len(aiData.ContentBox) != 4 {
		return "", nil, fmt.Errorf("no valid content_box found")
	}

	// B. Decode Image
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", nil, fmt.Errorf("image decode error: %w", err)
	}

	// C. Calculate Crop and Remap
	bounds := img.Bounds()
	cropRect, remappedRooms, err := CalculateCropAndRemap(bounds.Dx(), bounds.Dy(), aiData.ContentBox, aiData.Rooms)
	if err != nil {
		return "", nil, fmt.Errorf("remap error: %w", err)
	}

	// D. Crop Image
	croppedImg := imaging.Crop(img, cropRect)

	// E. Encode Cropped Image to Base64
	var buf bytes.Buffer
	err = png.Encode(&buf, croppedImg)
	if err != nil {
		return "", nil, fmt.Errorf("image encode error: %w", err)
	}
	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + encodedString, remappedRooms, nil
}
