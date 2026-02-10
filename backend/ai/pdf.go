package ai

import (
"bytes"
"fmt"
"image"
"image/png"

"github.com/gen2brain/go-fitz"
)

// FitzRenderer implements PDFRenderer using go-fitz (MuPDF wrapper).
type FitzRenderer struct{}

// RenderPage converts the specified page (1-based index) of the PDF data to an image.
func (r *FitzRenderer) RenderPage(pdfBytes []byte, pageNum int) (image.Image, error) {
doc, err := fitz.NewFromMemory(pdfBytes)
if err != nil {
return nil, fmt.Errorf("fitz.NewFromMemory: %w", err)
}
defer doc.Close()

// fitz uses 0-based index
// Assuming 1-based pageNum input
if pageNum < 1 {
return nil, fmt.Errorf("invalid page number: %d", pageNum)
}
idx := pageNum - 1

if idx >= doc.NumPage() {
return nil, fmt.Errorf("page number %d out of range (total %d)", pageNum, doc.NumPage())
}

img, err := doc.Image(idx)
if err != nil {
return nil, fmt.Errorf("doc.Image(%d): %w", idx, err)
}

return img, nil
}

// ConvertPDFToImage converts the first page of a PDF to an image.
// This is a convenience function for backward compatibility or simple usage.
func ConvertPDFToImage(pdfBytes []byte) (image.Image, error) {
renderer := &FitzRenderer{}
return renderer.RenderPage(pdfBytes, 1)
}

// ConvertPDFToImageBytes converts a PDF file to PNG binary data (first page).
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
