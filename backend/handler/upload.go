package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	_ "image/jpeg" // Support JPEG decoding
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"floorplan-whiteboard/ai"

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
		isValid = true
		mimeType = "application/pdf"
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG, JPG, and PDF are supported."})
		return
	}

	// 3. Call AI Service
	// Context from Gin request
	jsonResponse, err := ai.AnalyzeFloorplan(c.Request.Context(), fileBytes, mimeType)
	if err != nil {
		fmt.Printf("AI Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze floorplan: " + err.Error()})
		return
	}

	// 4. Process Image and Remap Coordinates
	finalImageBase64, remappedRooms, err := processImageAndCoordinates(fileBytes, jsonResponse, mimeType)
	if err != nil {
		fmt.Printf("Processing Error: %v\n", err)
		// Fallback to original if processing fails (e.g. PDF rasterization issues or invalid JSON)
		// For now, let's just return the raw JSON and no image update
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, jsonResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": remappedRooms,
		"image": finalImageBase64,
	})
}

// Data structures for JSON parsing
type GeminiResponse struct {
	ContentBox []int  `json:"content_box"` // [ymin, xmin, ymax, xmax] 0-1000
	Rooms      []Room `json:"rooms"`
}

type Room struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Rect []int  `json:"rect"` // [ymin, xmin, ymax, xmax] 0-1000
}

func processImageAndCoordinates(fileBytes []byte, jsonStr string, mimeType string) (string, []Room, error) {
	// A. Parse JSON
	var aiData GeminiResponse
	// Clean JSON string if needed (remove markdown)
	cleanJson := strings.ReplaceAll(jsonStr, "```json", "")
	cleanJson = strings.ReplaceAll(cleanJson, "```", "")
	
	if err := json.Unmarshal([]byte(cleanJson), &aiData); err != nil {
		return "", nil, fmt.Errorf("json parse error: %w", err)
	}

	// If no content_box, return original
	if len(aiData.ContentBox) != 4 {
		return "", nil, fmt.Errorf("no valid content_box found")
	}

	// B. Decode Image
	var img image.Image
	var err error
	if mimeType == "application/pdf" {
		// PDF processing complexity - for now, if PDF, we skip cropping or need a rasterizer.
		// Since we didn't add a PDF rasterizer yet (like pdfcpu's rasterize which requires dependencies or external tools),
		// we will SKIP image processing for PDFs and rely on frontend to render PDF.
		// BUT, the prompt returns coordinates relative to the PDF page 0-1000.
		// If we don't crop, the coordinates are fine as is (relative).
		// So we return raw rooms and no image.
		return "", aiData.Rooms, nil 
	} else {
		img, _, err = image.Decode(bytes.NewReader(fileBytes))
	}
	
	if err != nil {
		return "", nil, fmt.Errorf("image decode error: %w", err)
	}

	// C. Calculate Crop in Pixels
	bounds := img.Bounds()
	W, H := bounds.Dx(), bounds.Dy()
	
	// Gemini uses 0-1000 relative scale
	scaleY := func(v int) int { return int(float64(v) / 1000.0 * float64(H)) }
	scaleX := func(v int) int { return int(float64(v) / 1000.0 * float64(W)) }

	// ContentBox: [ymin, xmin, ymax, xmax]
	cYMin := scaleY(aiData.ContentBox[0])
	cXMin := scaleX(aiData.ContentBox[1])
	cYMax := scaleY(aiData.ContentBox[2])
	cXMax := scaleX(aiData.ContentBox[3])

	// Validate bounds
	if cXMin < 0 { cXMin = 0 }
	if cYMin < 0 { cYMin = 0 }
	if cXMax > W { cXMax = W }
	if cYMax > H { cYMax = H }
	
	cropRect := image.Rect(cXMin, cYMin, cXMax, cYMax)
	
	// D. Crop Image
	croppedImg := imaging.Crop(img, cropRect)

	// E. Remap Rooms
	var cleanRooms []Room
	for _, room := range aiData.Rooms {
		if len(room.Rect) != 4 {
			continue
		}
		// Original: [ymin, xmin, ymax, xmax] 0-1000
		rYMin := scaleY(room.Rect[0])
		rXMin := scaleX(room.Rect[1])
		rYMax := scaleY(room.Rect[2])
		rXMax := scaleX(room.Rect[3])

		// Apply Crop Shift
		newX := rXMin - cXMin
		newY := rYMin - cYMin
		newW := rXMax - rXMin
		newH := rYMax - rYMin
		
		// If negative (room outside content box?), clamp or ignore
		// Return in format frontend expects: [x, y, w, h]
		newRoom := Room{
			Name: room.Name,
			Type: room.Type,
			Rect: []int{newX, newY, newW, newH},
		}
		cleanRooms = append(cleanRooms, newRoom)
	}

	// F. Encode Cropped Image to Base64
	var buf bytes.Buffer
	err = png.Encode(&buf, croppedImg) // Use PNG for lossless
	if err != nil {
		return "", nil, fmt.Errorf("image encode error: %w", err)
	}
	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
	
	return "data:image/png;base64," + encodedString, cleanRooms, nil
}
