package handler

import (
	"floorplan-whiteboard/models"
	"image"
	"reflect"
	"testing"
)

func TestCalculateCropAndRemap(t *testing.T) {
	tests := []struct {
		name         string
		imgW, imgH   int
		contentBox   []int // [ymin, xmin, ymax, xmax] 0-1000
		geminiRooms  []GeminiRoom
		wantCropRect image.Rectangle // Pixels
		wantRooms    []models.Room   // [x, y, w, h] relative to crop
		wantErr      bool
	}{
		{
			name:       "Simple Case: Full size content box",
			imgW:       1000,
			imgH:       1000,
			contentBox: []int{0, 0, 1000, 1000}, // Full image
			geminiRooms: []GeminiRoom{
				{Name: "R1", Type: "OFFICE", Rect: []int{100, 100, 200, 200}}, // 100,100 -> 200,200
			},
			wantCropRect: image.Rect(0, 0, 1000, 1000),
			wantRooms: []models.Room{
				{Name: "R1", Type: models.RoomTypeOffice, Rect: []int{100, 100, 100, 100}, Status: models.RoomStatusAvailable}, // w=100, h=100
			},
		},
		{
			name:       "Simple Crop: Bottom Right Quadrant",
			imgW:       1000,
			imgH:       1000,
			contentBox: []int{500, 500, 1000, 1000}, // ymin, xmin, ymax, xmax
			geminiRooms: []GeminiRoom{
				{Name: "R2", Type: "MEETING", Rect: []int{600, 600, 800, 800}}, // Inside crop
			},
			wantCropRect: image.Rect(500, 500, 1000, 1000),
			wantRooms: []models.Room{
				{Name: "R2", Type: models.RoomTypeMeeting, Rect: []int{100, 100, 200, 200}, Status: models.RoomStatusAvailable}, // relative: 600-500=100
			},
		},
		{
			name:       "Non-Square Aspect Ratio: 2000x1000",
			imgW:       2000,
			imgH:       1000,
			contentBox: []int{0, 0, 1000, 500}, // Top Left Half (vertical split in image)
			// Wait, xmax=500 means 50% width -> 1000px
			// ymax=1000 means 100% height -> 1000px
			geminiRooms: []GeminiRoom{
				{Name: "R3", Type: "HALLWAY", Rect: []int{100, 100, 200, 200}}, // 100 means 10% -> y=100, x=200
			},
			wantCropRect: image.Rect(0, 0, 1000, 1000), // xmax=500/1000 * 2000 = 1000px
			wantRooms: []models.Room{
				{Name: "R3", Type: models.RoomTypeHallway, Rect: []int{200, 100, 200, 100}, Status: models.RoomStatusAvailable}, // x=10%*2000=200, y=10%*1000=100. w=200, h=100
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCropRect, gotRooms, err := CalculateCropAndRemap(tt.imgW, tt.imgH, tt.contentBox, tt.geminiRooms)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateCropAndRemap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCropRect, tt.wantCropRect) {
				t.Errorf("CalculateCropAndRemap() gotCropRect = %v, want %v", gotCropRect, tt.wantCropRect)
			}
			// Only check essential fields of rooms
			if len(gotRooms) != len(tt.wantRooms) {
				t.Errorf("CalculateCropAndRemap() gotRooms len = %d, want %d", len(gotRooms), len(tt.wantRooms))
				return
			}
			for i := range gotRooms {
				if gotRooms[i].Name != tt.wantRooms[i].Name {
					t.Errorf("Room[%d].Name = %v, want %v", i, gotRooms[i].Name, tt.wantRooms[i].Name)
				}
				if !reflect.DeepEqual(gotRooms[i].Rect, tt.wantRooms[i].Rect) {
					t.Errorf("Room[%d].Rect = %v, want %v", i, gotRooms[i].Rect, tt.wantRooms[i].Rect)
				}
			}
		})
	}
}
