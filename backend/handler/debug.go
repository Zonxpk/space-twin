package handler

import (
"bytes"
"encoding/base64"
"encoding/json"
"fmt"
"io"
"net/http"
"path/filepath"
"strings"

"floorplan-whiteboard/ai"

"github.com/gin-gonic/gin"
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
mimeType := header.Header.Get("Content-Type")

// Validate extension
isValid := false
if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
isValid = true
if mimeType == "" {
mimeType = "image/png"
}
} else if ext == ".pdf" {
isValid = true
mimeType = "application/pdf"
}

if !isValid {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG, JPG, and PDF are supported."})
return
}

// Analyze with Gemini
fmt.Println("DebugCrop: Analyzing with Gemini...")
jsonResp, err := ai.AnalyzeFloorplan(c.Request.Context(), fileBytes, mimeType)
if err != nil {
fmt.Printf("DebugCrop: AI detection failed: %v\n", err)
c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
return
}

// Parse JSON response
var aiData GeminiResponse
cleanJson := strings.ReplaceAll(jsonResp, "```json", "")
cleanJson = strings.ReplaceAll(cleanJson, "```", "")

if err := json.Unmarshal([]byte(cleanJson), &aiData); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
return
}

// Encode file to base64 for response
fileB64 := base64.StdEncoding.EncodeToString(fileBytes)

resp := gin.H{
"file":        "data:" + mimeType + ";base64," + fileB64,
"content_box": aiData.ContentBox,
"file_type":   ext,
}
c.JSON(http.StatusOK, resp)
}
