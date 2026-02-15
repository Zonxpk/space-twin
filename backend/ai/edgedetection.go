package ai

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
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

// GetConnectedComponentBoundingBoxes finds bounding boxes of all connected components in the image
func GetConnectedComponentBoundingBoxes(img image.Image) []image.Rectangle {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	var rects []image.Rectangle

	// Helper to check if pixel is "edge" (white-ish)
	isEdge := func(x, y int) bool {
		r, _, _, _ := img.At(x, y).RGBA()
		return r > 0x7FFF // > 50% brightness
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !visited[y][x] && isEdge(x, y) {
				// BFS to find component
				minX, minY, maxX, maxY := x, y, x, y
				q := []image.Point{{x, y}}
				visited[y][x] = true

				for len(q) > 0 {
					p := q[0]
					q = q[1:]

					if p.X < minX {
						minX = p.X
					}
					if p.X > maxX {
						maxX = p.X
					}
					if p.Y < minY {
						minY = p.Y
					}
					if p.Y > maxY {
						maxY = p.Y
					}

					// Check neighbors
					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							if dx == 0 && dy == 0 {
								continue
							}
							nx, ny := p.X+dx, p.Y+dy
							if nx >= 0 && nx < width && ny >= 0 && ny < height {
								if !visited[ny][nx] && isEdge(nx, ny) {
									visited[ny][nx] = true
									q = append(q, image.Point{nx, ny})
								}
							}
						}
					}
				}

				rects = append(rects, image.Rect(minX, minY, maxX+1, maxY+1))
			}
		}
	}

	return rects
}

// GetLargestComponentBoundingBox finds the bounding box of the largest connected component in the edge image
func GetLargestComponentBoundingBox(img image.Image) image.Rectangle {
	rects := GetConnectedComponentBoundingBoxes(img)
	if len(rects) == 0 {
		return img.Bounds()
	}

	maxArea := 0
	maxRect := img.Bounds()

	for _, r := range rects {
		area := r.Dx() * r.Dy()
		if area > maxArea {
			maxArea = area
			maxRect = r
		}
	}
	return maxRect
}

// CropImage crops the image to the specified rectangle
func CropImage(img image.Image, rect image.Rectangle) image.Image {
	return imaging.Crop(img, rect)
}

// Dilate performs morphological dilation on the image
// radius: the size of the dilation kernel radius (kernel width = 2*radius + 1)
func Dilate(img image.Image, radius int) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert to gray if needed
	var gray *image.Gray
	switch v := img.(type) {
	case *image.Gray:
		gray = v
	default:
		// Convert to *image.Gray
		gray = image.NewGray(bounds)
		draw.Draw(gray, bounds, img, bounds.Min, draw.Src)
	}

	if radius <= 0 {
		return gray
	}

	// Two-pass dilation (separable approximation with max filter)
	// Note: Rectangular dilation is separable.

	temp := image.NewGray(bounds)
	dst := image.NewGray(bounds)

	// Pass 1: Horizontal
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var maxVal uint8 = 0
			// Check neighbors from x-radius to x+radius
			for dx := -radius; dx <= radius; dx++ {
				nx := x + dx
				if nx >= 0 && nx < width {
					val := gray.GrayAt(nx, y).Y
					if val > maxVal {
						maxVal = val
					}
				}
			}
			temp.SetGray(x, y, color.Gray{Y: maxVal})
		}
	}

	// Pass 2: Vertical
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			var maxVal uint8 = 0
			// Check neighbors from y-radius to y+radius
			for dy := -radius; dy <= radius; dy++ {
				ny := y + dy
				if ny >= 0 && ny < height {
					val := temp.GrayAt(x, ny).Y
					if val > maxVal {
						maxVal = val
					}
				}
			}
			dst.SetGray(x, y, color.Gray{Y: maxVal})
		}
	}

	return dst
}

// GetMainContentBoundingBox finds the bounding box of the main content (floorplan)
// It uses dilation to connect disjoint edges, then finds all significant components
// and merges their bounding boxes to capture the full floorplan area.
func GetMainContentBoundingBox(img image.Image) image.Rectangle {
	// 1. Dilate to merge nearby edges and walls into a single blob
	// Radius 10 creates a 21x21 pixel kernel, bridging gaps up to ~20 pixels
	const dilationRadius = 10
	dilated := Dilate(img, dilationRadius)

	// 2. Find all connected components in the dilated image
	rects := GetConnectedComponentBoundingBoxes(dilated)

	if len(rects) == 0 {
		return img.Bounds()
	}

	// 3. Find the largest component area to use as a baseline
	maxArea := 0
	for _, r := range rects {
		area := r.Dx() * r.Dy()
		if area > maxArea {
			maxArea = area
		}
	}

	// 4. Filter components that are significant enough
	// Threshold: 5% of the largest component's area
	// This helps include other parts of the floorplan (e.g. detached garage)
	// while ignoring small noise (speckles).
	const areaThresholdRatio = 0.05
	threshold := int(float64(maxArea) * areaThresholdRatio)

	var significantRects []image.Rectangle
	for _, r := range rects {
		area := r.Dx() * r.Dy()
		if area >= threshold {
			significantRects = append(significantRects, r)
		}
	}

	// 5. Merge significant bounding boxes
	if len(significantRects) == 0 {
		return img.Bounds()
	}

	finalRect := significantRects[0]
	for i := 1; i < len(significantRects); i++ {
		finalRect = finalRect.Union(significantRects[i])
	}

	// 6. Compensate for dilation expansion to get closer to original bounds
	// Dilation expands the object by 'radius' pixels in all directions
	// We shrink it back, but be careful not to over-shrink if gaps were bridged.

	finalRect.Min.X += dilationRadius
	finalRect.Min.Y += dilationRadius
	finalRect.Max.X -= dilationRadius
	finalRect.Max.Y -= dilationRadius

	// Ensure valid rectangle
	if finalRect.Min.X >= finalRect.Max.X || finalRect.Min.Y >= finalRect.Max.Y {
		// If shrinking made it invalid, try without shrinking (or less shrinking)
		// Or return the union as is (slightly padded is better than empty)

		// Let's recompute union without shrinking for fallback
		finalRect = significantRects[0]
		for i := 1; i < len(significantRects); i++ {
			finalRect = finalRect.Union(significantRects[i])
		}
	}

	// Intersect with original bounds to be safe
	return finalRect.Intersect(img.Bounds())
}
