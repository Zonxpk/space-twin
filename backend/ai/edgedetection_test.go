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

func TestDilate(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 11, 11))
	// Set center pixel to white
	img.SetGray(5, 5, color.Gray{Y: 255})

	// Dilate with radius 2
	radius := 2
	dilated := Dilate(img, radius)
	grayDilated, ok := dilated.(*image.Gray)
	if !ok {
		t.Fatal("Dilate should return *image.Gray")
	}

	// Check bounds of dilation
	// Should be square from (5-2, 5-2) to (5+2, 5+2) = (3,3) to (7,7) inclusive
	for y := 0; y < 11; y++ {
		for x := 0; x < 11; x++ {
			expected := uint8(0)
			if x >= 3 && x <= 7 && y >= 3 && y <= 7 {
				expected = 255
			}

			got := grayDilated.GrayAt(x, y).Y
			if got != expected {
				t.Errorf("At (%d,%d): expected %d, got %d", x, y, expected, got)
			}
		}
	}
}

func TestGetMainContentBoundingBox(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 100, 100))

	// Create two 10x10 squares separated by 5px gap
	// Square 1: (20,20) - (30,30)
	// Square 2: (35,20) - (45,30)
	// Gap is from x=30 to x=35 (5px)

	for y := 20; y < 30; y++ {
		for x := 20; x < 30; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
		for x := 35; x < 45; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	// Dilation radius 10 should merge these
	rect := GetMainContentBoundingBox(img)

	// Expect bounding box to cover both squares roughly
	// Original bounds: (20,20) to (45,30)
	// The function returns rect that is dilated then shrunk back.
	// Since the gap is filled, it should be one component.

	if rect.Empty() {
		t.Error("Returned empty rectangle")
	}

	// Check if it includes both squares
	if rect.Min.X > 20 || rect.Min.Y > 20 {
		t.Errorf("Rect min too large: %v (expected <= (20,20))", rect.Min)
	}
	if rect.Max.X < 45 || rect.Max.Y < 30 {
		t.Errorf("Rect max too small: %v (expected >= (45,30))", rect.Max)
	}
}
