package ai

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/genai"
)

const (
	ProjectID = "floorplan-digital-twin"
	Location  = "global"
	ModelName = "gemini-3-flash-preview"
)

var (
	clientOnce sync.Once
	clientInst *genai.Client
	clientErr  error
)

// AnalyzeFloorplan sends image/PDF data to Vertex AI and returns the JSON analysis
func AnalyzeFloorplan(ctx context.Context, data []byte, mimeType string) (string, error) {
	client, err := getClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %w", err)
	}

	systemInstruction := `You are an expert floorplan analysis assistant extract structured data from architectural floorplan images.
Return machine-parseable JSON only.`

	promptText := `OBJECTIVE:
Detect all functional rooms/spaces in the floorplan image.

LABEL RULES:
- For each room, set "name" to the visible room label text when readable.
- If no readable label exists, use a short descriptive name.

TYPE RULES:
- Allowed values for "type": "OFFICE", "MEETING", "UNKNOWN".
- Use "UNKNOWN" for corridors, hallways, unlabeled open areas, or ambiguous room types.

COORDINATE RULES:
- "rect" must be [ymin, xmin, ymax, xmax].
- Use integer coordinates only.
- Use relative scale 0..1000 where [0,0] is top-left and [1000,1000] is bottom-right.
- Enforce ymin < ymax and xmin < xmax.

OUTPUT CONTRACT:
- Return strictly valid JSON with exactly one top-level key: "rooms".
- "rooms" is an array of objects with exactly: {"name", "type", "rect"}.
- Do not include markdown, prose, code fences, comments, or extra keys.
- If no valid rooms are detectable, return {"rooms":[]}.

EXAMPLE OUTPUT:
{"rooms":[{"name":"Office 101","type":"OFFICE","rect":[120,80,280,300]},{"name":"Unlabeled corridor","type":"UNKNOWN","rect":[300,40,420,960]}]}`

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
		SystemInstruction: &genai.Content{
			Role: "system",
			Parts: []*genai.Part{
				{Text: systemInstruction},
			},
		},
		Temperature:     float32Ptr(1),
		TopP:            float32Ptr(0.2),
		MaxOutputTokens: 65535,
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingLevel: genai.ThinkingLevelMinimal,
		},
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdOff,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdOff,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdOff,
			},
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdOff,
			},
		},
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type:     genai.TypeObject,
			Required: []string{"rooms"},
			Properties: map[string]*genai.Schema{
				"rooms": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type:     genai.TypeObject,
						Required: []string{"name", "type", "rect"},
						Properties: map[string]*genai.Schema{
							"name": {Type: genai.TypeString},
							"type": {
								Type: genai.TypeString,
								Enum: []string{"OFFICE", "MEETING", "UNKNOWN"},
							},
							"rect": {
								Type:     genai.TypeArray,
								MinItems: int64Ptr(4),
								MaxItems: int64Ptr(4),
								Items:    &genai.Schema{Type: genai.TypeInteger},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		if len(resp.Candidates) > 0 {
			fmt.Printf("[AI] no content — finish reason: %v\n", resp.Candidates[0].FinishReason)
		}
		return "", fmt.Errorf("no content returned")
	}

	candidate := resp.Candidates[0]
	fmt.Printf("[AI] finish reason: %v | token count: %v\n", candidate.FinishReason, candidate.TokenCount)

	part := candidate.Content.Parts[0]
	if part.Text != "" {
		fmt.Printf("[AI] raw response (first 500 chars): %.500s\n", part.Text)
		return part.Text, nil
	}

	return "", fmt.Errorf("unexpected response format")
}

func float32Ptr(value float32) *float32 {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

// getClient creates a new Vertex AI client (shared helper)
func getClient(ctx context.Context) (*genai.Client, error) {
	clientOnce.Do(func() {
		clientInst, clientErr = genai.NewClient(ctx, &genai.ClientConfig{
			Project:  ProjectID,
			Location: Location,
			Backend:  genai.BackendVertexAI,
		})
	})

	return clientInst, clientErr
}
