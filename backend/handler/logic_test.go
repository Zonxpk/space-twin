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
		geminiRooms  []GeminiRoom
		wantCropRect image.Rectangle // Pixels
		wantRooms    []models.Room   // [x, y, w, h] relative to full image
		wantErr      bool
	}{
		{
			name: "Simple Case: Full size image",
			imgW: 1000,
			imgH: 1000,
			geminiRooms: []GeminiRoom{
				{Name: "R1", Type: "OFFICE", Rect: []int{100, 100, 200, 200}}, // 100,100 -> 200,200
			},
			wantCropRect: image.Rect(0, 0, 1000, 1000),
			wantRooms: []models.Room{
				{Name: "R1", Type: models.RoomTypeOffice, Rect: []int{100, 100, 100, 100}, Status: models.RoomStatusAvailable}, // w=100, h=100
			},
		},
		{
			name: "Bottom Right Quadrant Room",
			imgW: 1000,
			imgH: 1000,
			geminiRooms: []GeminiRoom{
				{Name: "R2", Type: "MEETING", Rect: []int{600, 600, 800, 800}},
			},
			wantCropRect: image.Rect(0, 0, 1000, 1000),
			wantRooms: []models.Room{
				{Name: "R2", Type: models.RoomTypeMeeting, Rect: []int{600, 600, 200, 200}, Status: models.RoomStatusAvailable},
			},
		},
		{
			name: "Non-Square Aspect Ratio: 2000x1000",
			imgW: 2000,
			imgH: 1000,
			geminiRooms: []GeminiRoom{
				{Name: "R3", Type: "HALLWAY", Rect: []int{100, 100, 200, 200}}, // ymin=100, xmin=100 (10% of width 2000 -> 200)
			},
			wantCropRect: image.Rect(0, 0, 2000, 1000),
			wantRooms: []models.Room{
				// ymin=100/1000*1000=100. xmin=100/1000*2000=200.
				// ymax=200/1000*1000=200. xmax=200/1000*2000=400.
				// w=400-200=200. h=200-100=100.
				{Name: "R3", Type: models.RoomTypeHallway, Rect: []int{200, 100, 200, 100}, Status: models.RoomStatusAvailable},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCropRect, gotRooms, err := CalculateCropAndRemap(tt.imgW, tt.imgH, tt.geminiRooms)
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
