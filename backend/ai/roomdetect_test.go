package ai

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabiante/gopop/pdftoppm"
)

func buildTwoRoomPlan(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Outer walls
	for x := 0; x < width; x++ {
		for t := 0; t < 4; t++ {
			img.Set(x, t, color.Black)
			img.Set(x, height-1-t, color.Black)
		}
	}
	for y := 0; y < height; y++ {
		for t := 0; t < 4; t++ {
			img.Set(t, y, color.Black)
			img.Set(width-1-t, y, color.Black)
		}
	}

	// Middle dividing wall
	for y := 4; y < height-4; y++ {
		for t := -2; t <= 2; t++ {
			img.Set(width/2+t, y, color.Black)
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestDetectRoomsFromFloorplan_TwoRooms(t *testing.T) {
	data := buildTwoRoomPlan(240, 140)
	rooms, err := DetectRoomsFromFloorplan(data)
	if err != nil {
		t.Fatalf("DetectRoomsFromFloorplan returned error: %v", err)
	}

	if len(rooms) < 2 {
		t.Fatalf("expected at least 2 rooms, got %d", len(rooms))
	}

	for i, room := range rooms {
		if len(room.Rect) != 4 {
			t.Fatalf("room[%d] rect length = %d, want 4", i, len(room.Rect))
		}
		ymin, xmin, ymax, xmax := room.Rect[0], room.Rect[1], room.Rect[2], room.Rect[3]
		if ymin < 0 || xmin < 0 || ymax > 1000 || xmax > 1000 {
			t.Fatalf("room[%d] rect out of 0-1000 range: %v", i, room.Rect)
		}
		if ymax <= ymin || xmax <= xmin {
			t.Fatalf("room[%d] invalid rect ordering: %v", i, room.Rect)
		}
	}
}

func TestDetectRoomsFromFloorplan_InvalidImage(t *testing.T) {
	_, err := DetectRoomsFromFloorplan([]byte("not-an-image"))
	if err == nil {
		t.Fatal("expected decode error for invalid image")
	}
}

func TestCleanOCRText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "normal", in: "Office 101\n", want: "Office 101"},
		{name: "symbols", in: "@@Conf-Room##", want: "Conf-Room"},
		{name: "too short", in: "-", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanOCRText(tt.in)
			if got != tt.want {
				t.Fatalf("cleanOCRText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMergeUntilMaxRegionsCapsCount(t *testing.T) {
	regions := make([]regionBox, 0, 25)
	for i := 0; i < 25; i++ {
		x := i * 5
		regions = append(regions, regionBox{MinX: x, MinY: x, MaxX: x + 3, MaxY: x + 3})
	}

	merged := mergeUntilMaxRegions(regions, targetRoomUpperBound)
	if len(merged) > targetRoomUpperBound {
		t.Fatalf("expected <= %d regions after merge, got %d", targetRoomUpperBound, len(merged))
	}
}

func TestDetectRoomsFromFloorplan_MockSitePDF(t *testing.T) {
	pdfPath := filepath.Clean(filepath.Join("..", "..", "frontend", "public", "mock-site-floorplan.pdf"))
	if _, err := os.Stat(pdfPath); err != nil {
		t.Fatalf("missing test asset %s: %v", pdfPath, err)
	}

	imgBytes, converted := convertFirstPagePDFToPNG(t, pdfPath)
	if !converted {
		t.Skip("No local PDF rasterizer found (pdftoppm or magick); skipping PDF integration test")
	}

	rooms, err := DetectRoomsFromFloorplan(imgBytes)
	if err != nil {
		t.Fatalf("DetectRoomsFromFloorplan on mock-site-floorplan.pdf failed: %v", err)
	}
	t.Logf("mock-site-floorplan detected rooms: %d", len(rooms))

	if len(rooms) < 8 || len(rooms) > 16 {
		t.Fatalf("mock-site-floorplan room count = %d, expected around 12 (range 8..16)", len(rooms))
	}
}

func convertFirstPagePDFToPNG(t *testing.T, pdfPath string) ([]byte, bool) {
	t.Helper()
	tmpDir := t.TempDir()
	prefix := filepath.Join(tmpDir, "page")

	cmd, err := pdftoppm.NewCommand(
		pdfPath,
		prefix,
		pdftoppm.PNG(),
		pdftoppm.First(1),
		pdftoppm.Last(1),
		pdftoppm.Resolution(200),
	)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if runErr := cmd.Run(ctx); runErr == nil {
			if files, globErr := filepath.Glob(prefix + "*.png"); globErr == nil && len(files) > 0 {
				if data, readErr := os.ReadFile(files[0]); readErr == nil {
					return data, true
				}
			}
		}
	}

	if _, err := exec.LookPath("pdftoppm"); err == nil {
		cmd := exec.Command("pdftoppm", "-f", "1", "-singlefile", "-png", pdfPath, prefix)
		if err := cmd.Run(); err == nil {
			if data, readErr := os.ReadFile(prefix + ".png"); readErr == nil {
				return data, true
			}
		}
	}

	if _, err := exec.LookPath("magick"); err == nil {
		outPath := filepath.Join(tmpDir, "page.png")
		cmd := exec.Command("magick", "-density", "200", pdfPath+"[0]", outPath)
		if err := cmd.Run(); err == nil {
			if data, readErr := os.ReadFile(outPath); readErr == nil {
				return data, true
			}
		}
	}

	return nil, false
}
