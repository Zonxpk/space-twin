package ai

import (
	"image"
)

// PDFRenderer defines the strategy for rasterizing PDF pages into images.
// Implementations can wrap different libraries (e.g., go-fitz, unidoc, or external CLI tools).
type PDFRenderer interface {
	// RenderPage converts the specified page (1-based index) of the PDF data to an image.
	// It returns the rasterized image or an error if rendering fails.
	RenderPage(pdfBytes []byte, pageNum int) (image.Image, error)
}
