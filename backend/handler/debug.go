package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG, JPG, and PDF are supported for debug crop."})
		return
	}

	var img image.Image

	if ext == ".pdf" {
		// Create a temporary directory for extraction
		tempDir, err := os.MkdirTemp("", "pdf_images")
		if err != nil {
			fmt.Printf("DebugCrop: Temp Dir Error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp dir"})
			return
		}
		defer os.RemoveAll(tempDir) // Cleanup

		// Extract images from the PDF bytes
		
		// pdfcpu API to extract images
		// ExtractImagesFile signature: (inFile string, outDir string, pageNrs []string, conf *Configuration)
		// We need to write PDF to temp file first, then extract
		tempPDF := filepath.Join(tempDir, "input.pdf")
		err = os.WriteFile(tempPDF, fileBytes, 0644)
		if err != nil {
			fmt.Printf("DebugCrop: Failed to write temp PDF: %v\n", err)
			img = createPlaceholder(800, 600, color.RGBA{255, 0, 0, 255})
		} else {
			err = api.ExtractImagesFile(tempPDF, tempDir, nil, nil)
		}
		if err != nil {
			fmt.Printf("DebugCrop: Extraction Error: %v\n", err)
			// Fallback if extraction fails
			img = createPlaceholder(800, 600, color.RGBA{255, 0, 0, 255}) // Red
		} else {
			// Find the first image in the temp dir (skip the input PDF)
			files, err := os.ReadDir(tempDir)
			if err != nil {
				fmt.Println("DebugCrop: Failed to read temp dir")
				img = createPlaceholder(800, 600, color.RGBA{128, 128, 128, 255})
			} else {
				// Find first non-PDF file (actual extracted image)
				var imageFile string
				for _, f := range files {
					if !f.IsDir() && filepath.Ext(f.Name()) != ".pdf" {
						imageFile = filepath.Join(tempDir, f.Name())
						break
					}
				}
				
				if imageFile == "" {
					fmt.Println("DebugCrop: No images found in PDF")
					img = createPlaceholder(800, 600, color.RGBA{128, 128, 128, 255}) // Grey
				} else {
					fmt.Printf("DebugCrop: Found image %s\n", imageFile)
					
					imgFile, err := os.Open(imageFile)
					if err != nil {
						fmt.Printf("DebugCrop: Failed to open image: %v\n", err)
						img = createPlaceholder(800, 600, color.RGBA{255, 0, 0, 255})
					} else {
						defer imgFile.Close()
						img, _, err = image.Decode(imgFile)
						if err != nil {
							fmt.Printf("DebugCrop: Failed to decode image: %v\n", err)
							img = createPlaceholder(800, 600, color.RGBA{255, 0, 0, 255})
						}
					}
				}
			}
		}
	} else {
		// 3. Decode Image
		img, _, err = image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image"})
			return
		}
	}

	// 4. Crop Logic (Mock "Content Box" - Center 50%)
	bounds := img.Bounds()
	W, H := bounds.Dx(), bounds.Dy()

	// Let's pretend the content is in the middle 60%
	marginW := int(float64(W) * 0.2)
	marginH := int(float64(H) * 0.2)
	
	cropRect := image.Rect(marginW, marginH, W-marginW, H-marginH)
	
	// Crop
	croppedImg := imaging.Crop(img, cropRect)

	// 5. Encode Results
	originalBase64, err := encodeImageToBase64(img)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode original"})
		return
	}

	croppedBase64, err := encodeImageToBase64(croppedImg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode crop"})
		return
	}

	// Return both for comparison
	c.JSON(http.StatusOK, gin.H{
		"original": originalBase64,
		"cropped":  croppedBase64,
		"info":     fmt.Sprintf("Original: %dx%d, Cropped: %dx%d", W, H, croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy()),
	})
}

func createPlaceholder(w, h int, c color.Color) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, c)
		}
	}
	return m
}

func encodeImageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encodedString, nil
}
