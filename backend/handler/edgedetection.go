package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
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

// CropFloorplanRequest request body for cropping
type CropFloorplanRequest struct {
	Image   string                `json:"image" binding:"required"` // base64 or data:image/...
	Options *EdgeDetectionRequest `json:"options"`
}

// CropFloorplanResponse returns cropped floorplan
type CropFloorplanResponse struct {
	CroppedImage string `json:"cropped_image"`
	Message      string `json:"message"`
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
	return base64.StdEncoding.DecodeString(data)
}

// detectFloorplanEdges performs edge detection on an image using AI algorithms
// This is the core edge detection logic shared by multiple handlers
func detectFloorplanEdges(img image.Image, opts ai.EdgeDetectionOptions) (image.Image, error) {
	// ========== AI PROCESSING: EDGE DETECTION ==========
	// ProcessFloorplanForAnalysis applies:
	// 1. Gaussian blur to reduce noise and smooth the image
	// 2. Canny edge detection to identify sharp edges (boundaries of floorplan and paper)
	// This returns a binary edge map showing where edges are detected
	edgeImg := ai.ProcessFloorplanForAnalysis(img, opts)
	return edgeImg, nil
}

// CropFloorplanHandler godoc
// @Summary Automatic crop floorplan from paper
// @Description Detect and crop the floorplan area from a paper document image
// @ID cropFloorplan
// @Tags edge-detection
// @Accept json
// @Produce json
// @Param request body CropFloorplanRequest true "Image and edge detection options"
// @Success 200 {object} CropFloorplanResponse "Cropped floorplan"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/process/crop [post]
func CropFloorplanHandler(c *gin.Context) {
	var req CropFloorplanRequest

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

	// Build edge detection options
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

	// ========== AI PROCESSING STEP 1: EDGE DETECTION ==========
	// Call ProcessFloorplanEdges logic through helper function
	// This applies Gaussian blur and Canny edge detection
	edgeImg, err := detectFloorplanEdges(img, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Edge detection error: %v", err)})
		return
	}

	// ========== AI PROCESSING STEP 2: CROP CONTENT ==========
	// Find the bounding box of the content in the edge image
	// Use GetMainContentBoundingBox which uses dilation to group nearby objects (walls)
	// This helps detect the full floorplan area rather than just a single wall
	cropRect := ai.GetMainContentBoundingBox(edgeImg)

	// Scale cropRect to original image size
	// edgeImg might be resized if ResizeMaxWidth was set
	origBounds := img.Bounds()
	edgeBounds := edgeImg.Bounds()

	// Calculate scale factors
	scaleX := float64(origBounds.Dx()) / float64(edgeBounds.Dx())
	scaleY := float64(origBounds.Dy()) / float64(edgeBounds.Dy())

	// Scale the rectangle
	finalRect := image.Rect(
		int(float64(cropRect.Min.X)*scaleX),
		int(float64(cropRect.Min.Y)*scaleY),
		int(float64(cropRect.Max.X)*scaleX),
		int(float64(cropRect.Max.Y)*scaleY),
	)

	// Ensure the rectangle is within the original image bounds
	finalRect = finalRect.Intersect(origBounds)

	// Crop the original image
	cropped := ai.CropImage(img, finalRect)

	// Encode cropped image to data URL for sending to client
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, cropped); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
		return
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := "data:image/png;base64," + b64

	c.JSON(http.StatusOK, CropFloorplanResponse{
		CroppedImage: dataURL,
		Message:      "Floorplan cropped successfully",
	})
}
