package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

const (
	ProjectID = "floorplan-digital-twin"
	Location  = "global"
	ModelName = "gemini-3-flash-preview"
)

// AnalyzeFloorplan sends image/PDF data to Vertex AI and returns the JSON analysis
func AnalyzeFloorplan(ctx context.Context, data []byte, mimeType string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  ProjectID,
		Location: Location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %w", err)
	}
	// defer client.Close() // Client struct doesn't have Close method in new SDK, or it's not needed/shown in doc.
	// Wait, let me check if Client has Close.

	// Prompt to force JSON structure for rooms and content_box
	promptText := `Analyze this architectural floor plan image.
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
				All coordinates MUST be integers on a relative scale of 0 to 1000, where [0,0] is top-left and [1000,1000] is bottom-right of the original image.`

	// Create data part based on mimeType
	dataPart := genai.NewPartFromBytes(data, mimeType)

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: promptText},
				dataPart,
			},
			Role: "user",
		},
	}

	resp, err := client.Models.GenerateContent(ctx, ModelName, contents, &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		return part.Text, nil
	}

	return "", fmt.Errorf("unexpected response format")
}

// getClient creates a new Vertex AI client (shared helper)
func getClient(ctx context.Context) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		Project:  ProjectID,
		Location: Location,
		Backend:  genai.BackendVertexAI,
	})
}
