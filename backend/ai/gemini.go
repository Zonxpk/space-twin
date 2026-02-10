package ai

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
)

const (
	ProjectID = "floorplan-digital-twin"
	Location  = "us-central1"
	ModelName = "gemini-2.5-flash"
)

// AnalyzeFloorplan sends image/PDF data to Vertex AI and returns the JSON analysis
func AnalyzeFloorplan(ctx context.Context, data []byte, mimeType string) (string, error) {
	client, err := genai.NewClient(ctx, ProjectID, Location)
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(ModelName)
	model.ResponseMIMEType = "application/json"

	// Prompt to force JSON structure for rooms and content_box
	prompt := genai.Text("Analyze this floor plan. " +
		"1. Identify the 'content_box': the bounding box that contains the actual building layout, strictly EXCLUDING the page margins, title blocks, legends, and empty whitespace. " +
		"2. Identify all 'rooms': typical room detection. " +
		"Return a JSON object with: " +
		"'content_box': [ymin, xmin, ymax, xmax] (relative 0-1000), " +
		"'rooms': list of objects with 'name', 'type', and 'rect' [ymin, xmin, ymax, xmax] (relative 0-1000). " +
		"Ensure the output is valid JSON.")

	// Create data part based on mimeType
	var part genai.Part
	if mimeType == "application/pdf" {
		part = genai.Blob{MIMEType: mimeType, Data: data}
	} else {
		part = genai.ImageData(mimeType, data)
	}

	resp, err := model.GenerateContent(ctx, prompt, part)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(txt), nil
	}

	return "", fmt.Errorf("unexpected response format")
}

// getClient creates a new Vertex AI client (shared helper)
func getClient(ctx context.Context) (*genai.Client, error) {
	return genai.NewClient(ctx, ProjectID, Location)
}
