package ai

import (
"fmt"
"image"
)

// NOTE: PDF Rendering has been moved to the Frontend to avoid server-side dependency issues (CGO/DLLs) on Windows.
// This file is kept for interface compatibility or future implementations.

// ConvertPDFToImage converts the first page of a PDF to an image.
// DEPRECATED: Use frontend rendering.
func ConvertPDFToImage(pdfBytes []byte) (image.Image, error) {
return nil, fmt.Errorf("backend PDF rendering is disabled. Please convert to image on client side.")
}

// ConvertPDFToImageBytes converts a PDF file to PNG binary data (first page).
// DEPRECATED: Use frontend rendering.
func ConvertPDFToImageBytes(pdfBytes []byte) ([]byte, error) {
return nil, fmt.Errorf("backend PDF rendering is disabled. Please convert to image on client side.")
}
