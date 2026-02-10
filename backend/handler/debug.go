package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"floorplan-whiteboard/ai"

	"github.com/gin-gonic/gin"
)

// DebugCrop handles manual image cropping for testing
func DebugCrop(c *gin.Context) {
	// 1. Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Println("DebugCrop: No file uploaded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()
	fmt.Printf("DebugCrop: Received file %s\n", header.Filename)

	// 2. Read file content
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	fileBytes := buf.Bytes()
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Validate extension
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG, JPG, and PDF are supported."})
		return
	}

	// Convert images to PDF format if needed
	var analysisBytes []byte
	if ext == ".pdf" {
		analysisBytes = fileBytes
	} else {
		// Decode image first
		img, _, err := image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image"})
			return
		}

		// Convert image to PNG bytes for Gemini
		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, img); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image to PNG"})
			return
		}
		analysisBytes = pngBuf.Bytes()
	}

	// Analyze with Gemini
	fmt.Println("DebugCrop: Analyzing with Gemini...")
	contentBox, err := ai.DetectContentBoundsFromPDF(c.Request.Context(), analysisBytes)
	if err != nil {
		fmt.Printf("DebugCrop: AI detection failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed"})
		return
	}

	// Encode file to base64 for response
	fileB64 := base64.StdEncoding.EncodeToString(fileBytes)
	mimeType := "application/pdf"
	if ext == ".jpg" || ext == ".jpeg" {
		mimeType = "image/jpeg"
	} else if ext == ".png" {
		mimeType = "image/png"
	}

	resp := gin.H{
		"file":        "data:" + mimeType + ";base64," + fileB64,
		"content_box": contentBox,
		"file_type":   ext,
	}
	c.JSON(http.StatusOK, resp)
}
