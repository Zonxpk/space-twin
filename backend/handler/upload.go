package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // Support JPEG decoding
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"floorplan-whiteboard/ai"
	"floorplan-whiteboard/models"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

// UploadFloorplan godoc
// @Summary Upload a floorplan image
// @Description Upload a floorplan image (PNG, JPG, JPEG) and return the detected rooms
// @ID uploadFloorplan
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Floorplan image file"
// @Success 200 {object} map[string]interface{} "Detection results with rooms"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/upload [post]
func UploadFloorplan(c *gin.Context) {
	// 1. Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 2. Read file content
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	fileBytes := buf.Bytes()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mimeType := header.Header.Get("Content-Type")

	// Validate extension/mime
	isValid := false
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		isValid = true
		if mimeType == "" {
			mimeType = "image/png" // Default fallback
		}
	} else if ext == ".pdf" {
		// DEPRECATED: PDF should be converted on frontend now.
		// But if sent, we reject it as we removed backend rendering.
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF files should be processed by client. Please retry."})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PNG and JPG are supported."})
		return
	}

	// 3. Call AI Service
	jsonResponse, err := ai.AnalyzeFloorplan(c.Request.Context(), fileBytes, mimeType)
	if err != nil {
		fmt.Printf("AI Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze floorplan: " + err.Error()})
		return
	}

	// 4. Process Image and Remap Coordinates
	finalImageBase64, remappedRooms, err := processAndRemap(fileBytes, jsonResponse, mimeType)
	if err != nil {
		// Send a specific error message back to the frontend
		errorMsg := "Failed to process image after analysis: " + err.Error()
		fmt.Println("Processing Error:", errorMsg) // Keep server log
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": remappedRooms,
		"image": finalImageBase64,
	})
}

// GeminiResponse matches the JSON structure returned by Gemini
type GeminiResponse struct {
	Rooms []GeminiRoom `json:"rooms"`
}

func processAndRemap(fileBytes []byte, jsonStr string, mimeType string) (string, []models.Room, error) {
	// A. Parse JSON
	aiData, err := parseGeminiResponse(jsonStr)
	if err != nil {
		return "", nil, fmt.Errorf("json parse error: %w", err)
	}

	// B. Decode Image
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", nil, fmt.Errorf("image decode error: %w", err)
	}

	// C. Calculate Crop and Remap
	bounds := img.Bounds()
	cropRect, remappedRooms, err := CalculateCropAndRemap(bounds.Dx(), bounds.Dy(), aiData.Rooms)
	if err != nil {
		return "", nil, fmt.Errorf("remap error: %w", err)
	}

	// D. Crop Image
	croppedImg := imaging.Crop(img, cropRect)

	// E. Encode Cropped Image to Base64
	var buf bytes.Buffer
	err = png.Encode(&buf, croppedImg)
	if err != nil {
		return "", nil, fmt.Errorf("image encode error: %w", err)
	}
	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + encodedString, remappedRooms, nil
}

func parseGeminiResponse(jsonStr string) (GeminiResponse, error) {
	clean := cleanAIJSON(jsonStr)

	// Pass 1: clean JSON parses directly.
	var aiData GeminiResponse
	if err := json.Unmarshal([]byte(clean), &aiData); err == nil {
		return aiData, nil
	}

	// Pass 2: strip leading/trailing noise then parse.
	if objectJSON, err := extractFirstJSONObject(clean); err == nil {
		if err2 := json.Unmarshal([]byte(objectJSON), &aiData); err2 == nil {
			return aiData, nil
		}
	}

	// Pass 3: repair truncated brackets/braces then re-try passes 1+2.
	if repairedJSON, err := repairTruncatedJSONObject(clean); err == nil {
		if err2 := json.Unmarshal([]byte(repairedJSON), &aiData); err2 == nil {
			return aiData, nil
		}
		if objectJSON, err2 := extractFirstJSONObject(repairedJSON); err2 == nil {
			if err3 := json.Unmarshal([]byte(objectJSON), &aiData); err3 == nil {
				return aiData, nil
			}
		}
	}

	// Pass 4: last resort — scan rooms array and keep each fully-parsed object,
	// discarding only the final truncated entry. Returns a partial result rather
	// than a hard failure so the caller always gets whatever rooms were complete.
	if rooms := extractPartialRooms(clean); len(rooms) > 0 {
		fmt.Printf("[parser] recovered %d partial room(s) from truncated response\n", len(rooms))
		return GeminiResponse{Rooms: rooms}, nil
	}

	return GeminiResponse{}, errors.New("unable to parse model JSON response")
}

// extractPartialRooms scans the raw string for the rooms array and decodes
// each element individually with json.Decoder, stopping at the first error.
// This recovers all complete room objects even when the response is truncated
// mid-way through the last entry.
func extractPartialRooms(value string) []GeminiRoom {
	// Locate the start of the rooms array.
	idx := strings.Index(value, `"rooms"`)
	if idx == -1 {
		return nil
	}
	arrayStart := strings.Index(value[idx:], "[")
	if arrayStart == -1 {
		return nil
	}
	arrayStart += idx

	dec := json.NewDecoder(strings.NewReader(value[arrayStart:]))

	// Consume the opening '['
	if tok, err := dec.Token(); err != nil || tok != json.Delim('[') {
		return nil
	}

	var rooms []GeminiRoom
	for dec.More() {
		var room GeminiRoom
		if err := dec.Decode(&room); err != nil {
			break // truncated object — stop, keep what we have
		}
		if len(room.Rect) == 4 && room.Name != "" {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func cleanAIJSON(jsonStr string) string {
	clean := strings.TrimSpace(jsonStr)
	clean = strings.ReplaceAll(clean, "```json", "")
	clean = strings.ReplaceAll(clean, "```", "")
	return strings.TrimSpace(clean)
}

func extractFirstJSONObject(value string) (string, error) {
	start := strings.Index(value, "{")
	if start == -1 {
		return "", fmt.Errorf("no JSON object found in model response")
	}

	decoder := json.NewDecoder(strings.NewReader(value[start:]))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return "", fmt.Errorf("incomplete JSON object in model response")
	}

	if len(raw) == 0 || raw[0] != '{' {
		return "", fmt.Errorf("first JSON value is not an object")
	}

	return strings.TrimSpace(string(raw)), nil
}

func repairTruncatedJSONObject(value string) (string, error) {
	start := strings.Index(value, "{")
	if start == -1 {
		return "", fmt.Errorf("no JSON object found in model response")
	}

	candidate := strings.TrimSpace(value[start:])
	if candidate == "" {
		return "", fmt.Errorf("empty JSON candidate in model response")
	}

	// stack tracks closing tokens needed in reverse open order.
	stack := make([]byte, 0, 16)
	inString := false
	escaped := false

	for i := 0; i < len(candidate); i++ {
		ch := candidate[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			stack = append(stack, '}')
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '}' {
				return "", fmt.Errorf("invalid JSON object nesting in model response")
			}
			stack = stack[:len(stack)-1]
		case '[':
			stack = append(stack, ']')
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != ']' {
				return "", fmt.Errorf("invalid JSON array nesting in model response")
			}
			stack = stack[:len(stack)-1]
		}
	}

	if inString {
		// Close the truncated string before closing open structures.
		candidate += `"`
	}

	candidate = strings.TrimRight(candidate, " \t\r\n,")

	// Close in reverse stack order so brackets/braces interleave correctly.
	for i := len(stack) - 1; i >= 0; i-- {
		candidate += string(stack[i])
	}

	return candidate, nil
}
