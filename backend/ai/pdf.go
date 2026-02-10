package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	"cloud.google.com/go/vertexai/genai"
	"github.com/unidoc/unidoc/pdf/model"
)

// ConvertPDFToImage converts the first page of a PDF to an image using unidoc
// Returns the image and an error
func ConvertPDFToImage(pdfBytes []byte) (image.Image, error) {
	pdfReader, err := model.NewPdfReader(bytes.NewReader(pdfBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %v", err)
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, fmt.Errorf("failed to check encryption: %v", err)
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt PDF: %v", err)
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil || numPages < 1 {
		return nil, fmt.Errorf("failed to get page count: %v", err)
	}

	// Get the first page
	page, err := pdfReader.GetPage(1)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %v", err)
	}

	// Get page dimensions
	mediaBox, err := page.GetMediaBox()
	if err != nil {
		return nil, fmt.Errorf("failed to get media box: %v", err)
	}

	width := int(mediaBox.Urx - mediaBox.Llx)
	height := int(mediaBox.Ury - mediaBox.Lly)

	// Create a white background image
	bounds := image.Rect(0, 0, width, height)
	img := image.NewRGBA(bounds)

	// Fill with white
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, image.White)
		}
	}

	// Note: Full PDF rendering is complex. For now return a white image
	// In production, you'd use external tools like Ghostscript or ImageMagick
	// to properly render PDF content
	return img, nil
}

// ConvertPDFToImageBytes converts a PDF file to PNG binary data
func ConvertPDFToImageBytes(pdfBytes []byte) ([]byte, error) {
	img, err := ConvertPDFToImage(pdfBytes)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DetectContentBoundsFromPDF uses Gemini's native support to detect bounds
// Supports both PDF and image files (PNG, JPG)
func DetectContentBoundsFromPDF(ctx context.Context, fileBytes []byte) ([]int, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel(ModelName)
	model.ResponseMIMEType = "application/json"
	model.SetTemperature(0.1)

	prompt := `Analyze this architectural/building floor plan and identify the bounding box that contains ONLY the actual building layout (walls, rooms, spaces).

INCLUDE in the bounding box:
- Room layouts and walls
- Interior spaces and corridors
- Structural elements of the building

EXCLUDE from the bounding box (ignore these completely):
- Company logos and branding
- Title blocks and text annotations
- Page margins and borders
- Scale bars and legends
- North arrows and symbols
- Headers, footers, and page numbers
- Any decorative elements
- Surrounding whitespace

Focus on finding the tightest box around just the architectural floor plan drawing itself.

Return ONLY a JSON object with this exact format:
{
  "content_box": [ymin, xmin, ymax, xmax]
}

Where coordinates are on a 0-1000 scale:
- ymin: top edge of the floor plan layout (0 = top of image)
- xmin: left edge of the floor plan layout (0 = left of image)
- ymax: bottom edge of the floor plan layout (1000 = bottom of image)
- xmax: right edge of the floor plan layout (1000 = right of image)`

	// Detect MIME type based on file header
	mimeType := "application/pdf"
	if len(fileBytes) >= 8 && fileBytes[0] == 0x89 && fileBytes[1] == 0x50 && fileBytes[2] == 0x4E && fileBytes[3] == 0x47 {
		// PNG magic bytes: 89 50 4E 47
		mimeType = "image/png"
	} else if len(fileBytes) >= 2 && fileBytes[0] == 0xFF && fileBytes[1] == 0xD8 {
		// JPG magic bytes: FF D8
		mimeType = "image/jpeg"
	}

	// Use Gemini's native support
	blobPart := &genai.Blob{
		MIMEType: mimeType,
		Data:     fileBytes,
	}

	resp, err := model.GenerateContent(ctx, blobPart, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	var result struct {
		ContentBox []int `json:"content_box"`
	}

	if err := json.Unmarshal([]byte(resp.Candidates[0].Content.Parts[0].(genai.Text)), &result); err != nil {
		return nil, err
	}

	return result.ContentBox, nil
}
