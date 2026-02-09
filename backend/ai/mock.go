package ai

import "time"

// MockAnalyzeFloorplan returns a sample JSON response for testing without credentials
func MockAnalyzeFloorplan() string {
	// Simulate processing time
	time.Sleep(1 * time.Second)

	return `{
  "content_box": [0, 0, 1000, 1000],
  "rooms": [
    {
      "name": "Lobby",
      "type": "public",
      "rect": [10, 10, 300, 200]
    },
    {
      "name": "Office 101",
      "type": "office",
      "rect": [320, 10, 150, 200]
    },
    {
      "name": "Meeting Room A",
      "type": "meeting",
      "rect": [10, 220, 200, 200]
    },
    {
      "name": "Corridor",
      "type": "corridor",
      "rect": [220, 220, 100, 200]
    },
    {
      "name": "Restroom",
      "type": "restroom",
      "rect": [340, 220, 130, 150]
    }
  ]
}`
}
