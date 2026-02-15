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
	prompt := genai.Text(`Analyze this architectural floor plan image.
				Identify all functional 'rooms' and spaces (Offices, Meeting Rooms, Hallways, etc.).
				For each room, provide:
				- 'name': The text label found in the room (e.g., "Office 101", "Conf Room A"). If no label, use a descriptive name.
				- 'type': Categorize as "OFFICE", "MEETING", "HALLWAY", or "UNKNOWN".
				- 'rect': The bounding box of the room's interior space.

				Return a strictly valid JSON object (no markdown formatting) with this schema:
				{
					"rooms": [
						{ "name": "string", "type": "string", "rect": [ymin, xmin, ymax, xmax] }
					]
				}
				All coordinates MUST be integers on a relative scale of 0 to 1000, where [0,0] is top-left and [1000,1000] is bottom-right of the original image.`)

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
