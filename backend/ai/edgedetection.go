package ai

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/disintegration/imaging"
)

// EdgeDetectionOptions configures edge detection parameters
type EdgeDetectionOptions struct {
	BlurRadius     float64 // Gaussian blur radius (default: 1.2)
	CannyLow       float64 // Low threshold for Canny edge detection (default: 50)
	CannyHigh      float64 // High threshold for Canny edge detection (default: 150)
	ResizeMaxWidth int     // Max width for resizing before processing (0 = no resize)
}

// DefaultEdgeDetectionOptions returns sensible defaults
func DefaultEdgeDetectionOptions() EdgeDetectionOptions {
	return EdgeDetectionOptions{
		BlurRadius:     1.2,
		CannyLow:       50,
		CannyHigh:      150,
		ResizeMaxWidth: 800,
	}
}

// ApplyGaussianBlur applies Gaussian blur to an image
func ApplyGaussianBlur(img image.Image, radius float64) image.Image {
	return imaging.Blur(img, radius)
}

// DetectEdgesCanny performs Canny edge detection on an image
// Returns a grayscale image with detected edges highlighted
func DetectEdgesCanny(img image.Image, lowThresh, highThresh float64) image.Image {
	// Convert to grayscale
	gray := imaging.Grayscale(img)

	// Get image bounds
	bounds := gray.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate gradients using Sobel operators
	gx := make([][]float64, height)
	gy := make([][]float64, height)
	for i := range gx {
		gx[i] = make([]float64, width)
		gy[i] = make([]float64, width)
	}

	// Sobel kernels
	sobelX := [3][3]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	sobelY := [3][3]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	// Apply Sobel operators
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var sx, sy float64
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					px := x + dx
					py := y + dy
					r, _, _, _ := gray.At(px, py).RGBA()
					val := float64(r >> 8) // Convert to 0-255

					sx += sobelX[dy+1][dx+1] * val
					sy += sobelY[dy+1][dx+1] * val
				}
			}
			gx[y][x] = sx
			gy[y][x] = sy
		}
	}

	// Calculate magnitude and direction
	magnitude := make([][]float64, height)
	direction := make([][]float64, height)
	for i := range magnitude {
		magnitude[i] = make([]float64, width)
		direction[i] = make([]float64, width)
	}

	maxMag := 0.0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mag := math.Sqrt(gx[y][x]*gx[y][x] + gy[y][x]*gy[y][x])
			magnitude[y][x] = mag
			if mag > maxMag {
				maxMag = mag
			}

			// Calculate direction (0-180 degrees mapped to 0-4)
			angle := math.Atan2(gy[y][x], gx[y][x]) * 180 / math.Pi
			if angle < 0 {
				angle += 180
			}
			direction[y][x] = angle
		}
	}

	// Non-maximum suppression
	suppressed := make([][]float64, height)
	for i := range suppressed {
		suppressed[i] = make([]float64, width)
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			mag := magnitude[y][x]
			angle := direction[y][x]

			var q, r float64
			if (angle >= 0 && angle < 22.5) || (angle >= 157.5 && angle <= 180) {
				// Horizontal
				q = magnitude[y][x+1]
				r = magnitude[y][x-1]
			} else if angle >= 22.5 && angle < 67.5 {
				// Diagonal /
				q = magnitude[y+1][x-1]
				r = magnitude[y-1][x+1]
			} else if angle >= 67.5 && angle < 112.5 {
				// Vertical
				q = magnitude[y+1][x]
				r = magnitude[y-1][x]
			} else {
				// Diagonal \
				q = magnitude[y-1][x-1]
				r = magnitude[y+1][x+1]
			}

			if mag >= q && mag >= r {
				suppressed[y][x] = mag
			}
		}
	}

	// Double thresholding and edge tracking
	edges := make([][]bool, height)
	for i := range edges {
		edges[i] = make([]bool, width)
	}

	// Normalize suppressed values to 0-255 range
	normMax := 0.0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if suppressed[y][x] > normMax {
				normMax = suppressed[y][x]
			}
		}
	}

	if normMax == 0 {
		normMax = 1
	}

	// Apply thresholds
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			normalized := (suppressed[y][x] / normMax) * 255
			if normalized > highThresh {
				edges[y][x] = true
			}
		}
	}

	// Convert to image
	result := image.NewGray(bounds)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if edges[y][x] {
				result.SetGray(x, y, color.Gray{255})
			} else {
				result.SetGray(x, y, color.Gray{0})
			}
		}
	}

	return result
}

// ProcessFloorplanForAnalysis applies preprocessing for better AI analysis
// Returns a processed image with edges detected and blur applied
func ProcessFloorplanForAnalysis(img image.Image, opts EdgeDetectionOptions) image.Image {
	// Optionally resize for faster processing
	if opts.ResizeMaxWidth > 0 && img.Bounds().Dx() > opts.ResizeMaxWidth {
		img = imaging.Fit(img, opts.ResizeMaxWidth, opts.ResizeMaxWidth, imaging.Lanczos)
	}

	// Apply Gaussian blur to reduce noise
	blurred := ApplyGaussianBlur(img, opts.BlurRadius)

	// Detect edges
	edges := DetectEdgesCanny(blurred, opts.CannyLow, opts.CannyHigh)

	return edges
}

// GetEdgeDataURL returns the edge-detected image as a base64 PNG data URL
func GetEdgeDataURL(img image.Image, opts EdgeDetectionOptions) (string, error) {
	processed := ProcessFloorplanForAnalysis(img, opts)

	// Encode to PNG
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, processed); err != nil {
		return "", err
	}

	// Convert to base64 data URL
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + b64, nil
}
