package handler

import (
	"floorplan-whiteboard/models"
	"image"
)

// GeminiRoom represents the room structure returned by Gemini (0-1000 coordinates).
type GeminiRoom struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Rect []int  `json:"rect"` // [ymin, xmin, ymax, xmax] 0-1000
}

// CalculateCropAndRemap computes the crop rectangle and remaps room coordinates.
func CalculateCropAndRemap(imgW, imgH int, geminiRooms []GeminiRoom) (image.Rectangle, []models.Room, error) {
	// Helper to scale 0-1000 to pixels
	scaleY := func(v int) int { return int(float64(v) / 1000.0 * float64(imgH)) }
	scaleX := func(v int) int { return int(float64(v) / 1000.0 * float64(imgW)) }

	var cYMin, cXMin, cYMax, cXMax int

	// Default to full image
	cYMin = 0
	cXMin = 0
	cYMax = imgH
	cXMax = imgW

	// Validate bounds
	if cXMin < 0 {
		cXMin = 0
	}
	if cYMin < 0 {
		cYMin = 0
	}
	if cXMax > imgW {
		cXMax = imgW
	}
	if cYMax > imgH {
		cYMax = imgH
	}

	// Create crop rect
	cropRect := image.Rect(cXMin, cYMin, cXMax, cYMax)

	var remappedRooms []models.Room
	for _, room := range geminiRooms {
		if len(room.Rect) != 4 {
			continue
		}

		// Original Room Coords: [ymin, xmin, ymax, xmax]
		rYMin := scaleY(room.Rect[0])
		rXMin := scaleX(room.Rect[1])
		rYMax := scaleY(room.Rect[2])
		rXMax := scaleX(room.Rect[3])

		// Remap relative to crop
		newX := rXMin - cXMin
		newY := rYMin - cYMin
		newW := rXMax - rXMin
		newH := rYMax - rYMin

		// Basic validation: ignore if width/height <= 0
		if newW <= 0 || newH <= 0 {
			continue
		}

		// Convert Type string to Enum
		roomType := models.RoomTypeUnknown
		switch room.Type {
		case "OFFICE":
			roomType = models.RoomTypeOffice
		case "MEETING":
			roomType = models.RoomTypeMeeting
		case "HALLWAY":
			roomType = models.RoomTypeHallway
		}

		remappedRooms = append(remappedRooms, models.Room{
			Name:   room.Name,
			Type:   roomType,
			Rect:   []int{newX, newY, newW, newH},
			Status: models.RoomStatusAvailable, // Default status
		})
	}

	return cropRect, remappedRooms, nil
}
