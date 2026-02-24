package handler

import "testing"

func TestParseGeminiResponse_CleanJSON(t *testing.T) {
	input := "```json\n{\"rooms\":[{\"name\":\"Office 101\",\"type\":\"OFFICE\",\"rect\":[10,20,100,200]}]}\n```"

	got, err := parseGeminiResponse(input)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}

	if len(got.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(got.Rooms))
	}

	if got.Rooms[0].Name != "Office 101" {
		t.Fatalf("expected room name Office 101, got %q", got.Rooms[0].Name)
	}
}

func TestParseGeminiResponse_ExtractFirstObject(t *testing.T) {
	input := "Result:\n{\"rooms\":[]}\nThanks"

	got, err := parseGeminiResponse(input)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}

	if len(got.Rooms) != 0 {
		t.Fatalf("expected empty rooms, got %d", len(got.Rooms))
	}
}

func TestParseGeminiResponse_TruncatedObjectRecovered(t *testing.T) {
	input := "{\"rooms\":[{\"name\":\"Office 7\",\"type\":\"OFFICE\",\"rect\":[10,20,50,70]}]"

	got, err := parseGeminiResponse(input)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}

	if len(got.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(got.Rooms))
	}

	if got.Rooms[0].Name != "Office 7" {
		t.Fatalf("expected room name Office 7, got %q", got.Rooms[0].Name)
	}
}

func TestParseGeminiResponse_TruncatedMidString(t *testing.T) {
	// Second room is truncated mid-name; only first room should be recovered.
	input := `{"rooms":[{"name":"Office 1","type":"OFFICE","rect":[10,20,50,70]},{"name":"Off`

	got, err := parseGeminiResponse(input)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}
	if len(got.Rooms) < 1 {
		t.Fatalf("expected at least 1 room, got %d", len(got.Rooms))
	}
	if got.Rooms[0].Name != "Office 1" {
		t.Fatalf("expected room name Office 1, got %q", got.Rooms[0].Name)
	}
}

func TestParseGeminiResponse_TruncatedMidRect(t *testing.T) {
	// Second room's rect is incomplete; only first room should survive.
	input := `{"rooms":[{"name":"Office 1","type":"OFFICE","rect":[10,20,50,70]},{"name":"Office 2","type":"MEETING","rect":[80,`

	got, err := parseGeminiResponse(input)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}
	if len(got.Rooms) < 1 {
		t.Fatalf("expected at least 1 room, got %d", len(got.Rooms))
	}
	if got.Rooms[0].Name != "Office 1" {
		t.Fatalf("expected room name Office 1, got %q", got.Rooms[0].Name)
	}
}

func TestParseGeminiResponse_Invalid(t *testing.T) {
	input := "no json here"

	_, err := parseGeminiResponse(input)
	if err == nil {
		t.Fatalf("expected error for invalid input")
	}
}
