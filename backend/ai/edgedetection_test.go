package ai

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage creates a simple test image with a black square on white background
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Draw black square in center
	sqSize := width / 3
	for y := height/3 - sqSize/2; y < height/3+sqSize/2; y++ {
		for x := width/3 - sqSize/2; x < width/3+sqSize/2; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, color.Black)
			}
		}
	}

	return img
}

func TestApplyGaussianBlur(t *testing.T) {
	img := createTestImage(100, 100)
	blurred := ApplyGaussianBlur(img, 1.2)

	if blurred == nil {
		t.Error("ApplyGaussianBlur returned nil")
	}

	if blurred.Bounds() != img.Bounds() {
		t.Error("Blurred image bounds don't match original")
	}
}

func TestDetectEdgesCanny(t *testing.T) {
	img := createTestImage(100, 100)
	edges := DetectEdgesCanny(img, 50, 150)

	if edges == nil {
		t.Error("DetectEdgesCanny returned nil")
	}

	if edges.Bounds() != img.Bounds() {
		t.Error("Edge image bounds don't match original")
	}
}

func TestProcessFloorplanForAnalysis(t *testing.T) {
	img := createTestImage(100, 100)
	opts := DefaultEdgeDetectionOptions()
	opts.ResizeMaxWidth = 0 // Disable resizing for test

	processed := ProcessFloorplanForAnalysis(img, opts)

	if processed == nil {
		t.Error("ProcessFloorplanForAnalysis returned nil")
	}

	if processed.Bounds() != img.Bounds() {
		t.Error("Processed image bounds don't match original")
	}
}

func TestGetEdgeDataURL(t *testing.T) {
	img := createTestImage(100, 100)
	opts := DefaultEdgeDetectionOptions()
	opts.ResizeMaxWidth = 0

	dataURL, err := GetEdgeDataURL(img, opts)

	if err != nil {
		t.Fatalf("GetEdgeDataURL returned error: %v", err)
	}

	if dataURL == "" {
		t.Error("GetEdgeDataURL returned empty string")
	}

	// Check that it's a valid data URL
	if len(dataURL) < 30 || dataURL[:22] != "data:image/png;base64," {
		t.Error("GetEdgeDataURL returned invalid data URL format")
	}
}

func TestDefaultEdgeDetectionOptions(t *testing.T) {
	opts := DefaultEdgeDetectionOptions()

	if opts.BlurRadius != 1.2 {
		t.Errorf("Expected BlurRadius 1.2, got %f", opts.BlurRadius)
	}

	if opts.CannyLow != 50 {
		t.Errorf("Expected CannyLow 50, got %f", opts.CannyLow)
	}

	if opts.CannyHigh != 150 {
		t.Errorf("Expected CannyHigh 150, got %f", opts.CannyHigh)
	}

	if opts.ResizeMaxWidth != 800 {
		t.Errorf("Expected ResizeMaxWidth 800, got %d", opts.ResizeMaxWidth)
	}
}
