package handler

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"floorplan-whiteboard/ai"

	"github.com/gin-gonic/gin"
)

// EdgeDetectionRequest contains parameters for edge detection
type EdgeDetectionRequest struct {
	BlurRadius     float64 `json:"blur_radius" default:"1.2"`
	CannyLow       float64 `json:"canny_low" default:"50"`
	CannyHigh      float64 `json:"canny_high" default:"150"`
	ResizeMaxWidth int     `json:"resize_max_width" default:"800"`
}

// EdgeDetectionResponse returns the processed image as a data URL
type EdgeDetectionResponse struct {
	ProcessedImage string `json:"processed_image"` // data:image/png;base64,...
	Message        string `json:"message"`
}

// ProcessFloorplanEdges godoc
// @Summary Detect edges in a floorplan image
// @Description Process a floorplan image to detect edges using Canny edge detection
// @ID detectEdges
// @Tags edge-detection
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Floorplan image file"
// @Param blur_radius formData number false "Blur radius for preprocessing" default(1.2)
// @Param canny_low formData number false "Canny low threshold" default(50)
// @Param canny_high formData number false "Canny high threshold" default(150)
// @Param resize_max_width formData integer false "Maximum width for resizing" default(800)
// @Success 200 {object} EdgeDetectionResponse "Edge detection result"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/process/edges [post]
func ProcessFloorplanEdges(c *gin.Context) {
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

	// Validate extension
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PNG, JPG, JPEG files are supported"})
		return
	}

	// 3. Decode image
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
		return
	}

	// 4. Parse edge detection options from query params
	opts := ai.DefaultEdgeDetectionOptions()

	// Allow optional override of parameters via query string
	if blurStr := c.Query("blur_radius"); blurStr != "" {
		var blur float64
		if _, err := fmt.Sscanf(blurStr, "%f", &blur); err == nil {
			opts.BlurRadius = blur
		}
	}

	if lowStr := c.Query("canny_low"); lowStr != "" {
		var low float64
		if _, err := fmt.Sscanf(lowStr, "%f", &low); err == nil {
			opts.CannyLow = low
		}
	}

	if highStr := c.Query("canny_high"); highStr != "" {
		var high float64
		if _, err := fmt.Sscanf(highStr, "%f", &high); err == nil {
			opts.CannyHigh = high
		}
	}

	if resizeStr := c.Query("resize_max_width"); resizeStr != "" {
		var resize int
		if _, err := fmt.Sscanf(resizeStr, "%d", &resize); err == nil {
			opts.ResizeMaxWidth = resize
		}
	}

	// 5. Process floorplan with edge detection
	dataURL, err := ai.GetEdgeDataURL(img, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Processing error: %v", err)})
		return
	}

	// 6. Return response
	response := EdgeDetectionResponse{
		ProcessedImage: dataURL,
		Message:        "Edge detection completed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// ProcessFloorplanWithJSON godoc
// @Summary Detect edges using JSON request with base64 image
// @Description Process a floorplan image (provided as base64) to detect edges
// @ID detectEdgesJSON
// @Tags edge-detection
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "JSON request with image and options"
// @Success 200 {object} map[string]interface{} "Edge detection result with processed image"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/process/edges-json [post]
func ProcessFloorplanWithJSON(c *gin.Context) {
	// Parse JSON body for base64 image
	var req struct {
		Image   string                `json:"image" binding:"required"` // base64 or data:image/...
		Options *EdgeDetectionRequest `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Decode base64 image
	imageBytes, err := DecodeBase64Image(req.Image)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 image"})
		return
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image data"})
		return
	}

	// Build options
	opts := ai.DefaultEdgeDetectionOptions()
	if req.Options != nil {
		if req.Options.BlurRadius > 0 {
			opts.BlurRadius = req.Options.BlurRadius
		}
		if req.Options.CannyLow >= 0 {
			opts.CannyLow = req.Options.CannyLow
		}
		if req.Options.CannyHigh >= 0 {
			opts.CannyHigh = req.Options.CannyHigh
		}
		if req.Options.ResizeMaxWidth > 0 {
			opts.ResizeMaxWidth = req.Options.ResizeMaxWidth
		}
	}

	// Process image
	dataURL, err := ai.GetEdgeDataURL(img, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Processing error: %v", err)})
		return
	}

	response := map[string]interface{}{
		"processed_image": dataURL,
		"message":         "Edge detection completed successfully",
		"options_used":    opts,
	}

	c.JSON(http.StatusOK, response)
}

// DecodeBase64Image decode base64 image data, handling both plain base64 and data URLs
func DecodeBase64Image(data string) ([]byte, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(data, "data:") {
		// Format: data:image/png;base64,<base64data>
		parts := strings.Split(data, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		data = parts[1]
	}

	// Decode base64
	return []byte(data), nil
}
