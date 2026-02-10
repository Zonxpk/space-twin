package models

import (
	"time"
)

// RoomType defines the type of a room (e.g., Office, Meeting Room).
type RoomType string

const (
	RoomTypeOffice  RoomType = "OFFICE"
	RoomTypeMeeting RoomType = "MEETING"
	RoomTypeHallway RoomType = "HALLWAY"
	RoomTypeUnknown RoomType = "UNKNOWN"
)

// RoomStatus defines the current availability of a room.
type RoomStatus string

const (
	RoomStatusAvailable RoomStatus = "AVAILABLE"
	RoomStatusBusy      RoomStatus = "BUSY"
	RoomStatusOffline   RoomStatus = "OFFLINE"
)

// Rect represents a rectangle with integer coordinates [x, y, w, h].
// Used for pixel coordinates relative to the cropped image.
type Rect []int

// Room represents a detected functional space within a floorplan.
type Room struct {
	ID          string     `json:"id"`
	FloorplanID string     `json:"floorplan_id"`
	Name        string     `json:"name"`
	Type        RoomType   `json:"type"`
	Rect        Rect       `json:"rect"` // [x, y, w, h]
	Status      RoomStatus `json:"status"`
}

// Floorplan represents the processed digital twin.
type Floorplan struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	ImageURL  string    `json:"image_url"` // Data URI or URL
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Rooms     []Room    `json:"rooms"`
	CreatedAt time.Time `json:"created_at"`
}

// ContentBox represents the detected content area in the original image.
// Bounds are [ymin, xmin, ymax, xmax] in relative 0-1000 coordinates (Gemini format).
type ContentBox struct {
	OriginalWidth  int   `json:"original_width"`
	OriginalHeight int   `json:"original_height"`
	Bounds         []int `json:"bounds"`
}
