package ai

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

const (
	targetRoomCount      = 12
	targetRoomUpperBound = 16 // "12 more or less" upper tolerance
)

// DetectedRoom represents a detected room in Gemini-compatible 0-1000 coordinates.
type DetectedRoom struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Rect []int  `json:"rect"` // [ymin, xmin, ymax, xmax] in 0-1000 scale
}

type detectedCandidate struct {
	Rect regionBox
	Name string
	Type string
}

// DetectRoomsFromFloorplan performs classical room detection without LLM inference.
// It returns room boxes in [ymin, xmin, ymax, xmax] on 0-1000 relative coordinates.
func DetectRoomsFromFloorplan(data []byte) ([]DetectedRoom, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("image decode error: %w", err)
	}

	grayImg := imaging.Grayscale(img)

	w := grayImg.Bounds().Dx()
	h := grayImg.Bounds().Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}

	// Moderate wall closing — just enough to merge thin line fragments.
	wallMask := closeWallMask(buildWallMask(grayImg), 1)
	freeMask := invertMask(wallMask)
	compMap, components := connectedComponentsFree(freeMask)

	if len(components) == 0 {
		return nil, nil
	}

	// Identify background (exterior) components by border contact.
	bgIDs := findBackgroundComponents(compMap, components, w, h)

	// Distance-transform seeds to locate room centres.
	distance := distanceTransform(freeMask)

	// Adaptive peak threshold: require seeds to be at least 1/3 of the
	// maximum distance value within non-background free space.
	maxDist := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !freeMask[y][x] {
				continue
			}
			cid := compMap[y][x]
			if cid >= 0 && bgIDs[cid] {
				continue // skip background
			}
			if distance[y][x] > maxDist {
				maxDist = distance[y][x]
			}
		}
	}
	minPeak := max(15, maxDist/3)
	seeds := localMaxSeeds(distance, freeMask, minPeak)

	// Additional spatial dedup: seeds closer than min(w,h)/30 are merged.
	dedupDist := max(minPeak, min(w, h)/30)
	seeds = dedupeNearbySeeds(seeds, dedupDist)

	if len(seeds) == 0 {
		return nil, nil
	}

	// Remove seeds in background or tiny components.
	filteredSeeds := make([]point, 0, len(seeds))
	for _, seed := range seeds {
		compID := compMap[seed.Y][seed.X]
		if compID < 0 || compID >= len(components) {
			continue
		}
		if bgIDs[compID] {
			continue
		}
		comp := components[compID]
		if comp.Area < max(220, (w*h)/2600) {
			continue
		}
		filteredSeeds = append(filteredSeeds, seed)
	}
	seeds = filteredSeeds
	if len(seeds) == 0 {
		return nil, nil
	}

	// Grow a wall-bounded rectangle directly from each seed.
	// This naturally aligns rooms to walls instead of Voronoi cell shapes.
	minRoomDim := max(12, min(w, h)/100)
	tesseractCmd := resolveTesseractCmd()
	candidates := make([]detectedCandidate, 0, len(seeds))
	for _, seed := range seeds {
		rect := growRoomFromSeed(seed, wallMask, w, h)
		if rect.MaxX-rect.MinX < minRoomDim || rect.MaxY-rect.MinY < minRoomDim {
			continue
		}
		name := ""
		if tesseractCmd != "" {
			name = detectRoomText(img, rect, tesseractCmd)
		}
		if name == "" {
			name = "Room"
		}
		rType := inferRoomType(name)
		candidates = append(candidates, detectedCandidate{Rect: rect, Name: name, Type: rType})
	}

	candidates = suppressOverlapsAndSort(candidates)

	if len(candidates) == 0 {
		return nil, nil
	}

	rooms := make([]DetectedRoom, 0, len(candidates))
	counter := 1
	for _, c := range candidates {
		name := c.Name
		if strings.HasPrefix(strings.ToLower(name), "room ") {
			name = fmt.Sprintf("Room %d", counter)
			counter++
		}
		rooms = append(rooms, DetectedRoom{
			Name: name,
			Type: c.Type,
			Rect: []int{
				to1000(c.Rect.MinY, h),
				to1000(c.Rect.MinX, w),
				to1000(c.Rect.MaxY, h),
				to1000(c.Rect.MaxX, w),
			},
		})
	}

	return rooms, nil
}

type point struct {
	X int
	Y int
}

type regionBox struct {
	MinX int
	MinY int
	MaxX int
	MaxY int
}

type componentInfo struct {
	MinX          int
	MinY          int
	MaxX          int
	MaxY          int
	Area          int
	TouchesBorder bool
}

// findBackgroundComponents identifies exterior/background components by
// counting how many border pixels each component occupies.  Only the
// large, heavily-border-touching components are marked as background;
// rooms that merely happen to touch one edge of the image are kept.
func findBackgroundComponents(compMap [][]int, components []componentInfo, w, h int) map[int]bool {
	bgIDs := make(map[int]bool)
	if len(components) == 0 {
		return bgIDs
	}

	// Count border pixels per component.
	borderCount := make([]int, len(components))
	for x := 0; x < w; x++ {
		if id := compMap[0][x]; id >= 0 {
			borderCount[id]++
		}
		if id := compMap[h-1][x]; id >= 0 {
			borderCount[id]++
		}
	}
	for y := 1; y < h-1; y++ {
		if id := compMap[y][0]; id >= 0 {
			borderCount[id]++
		}
		if id := compMap[y][w-1]; id >= 0 {
			borderCount[id]++
		}
	}

	totalBorder := 2*(w+h) - 4
	imageArea := w * h

	for id, comp := range components {
		if !comp.TouchesBorder {
			continue
		}
		borderRatio := float64(borderCount[id]) / float64(totalBorder)
		areaRatio := float64(comp.Area) / float64(imageArea)
		// Background: extensive border contact AND large area, or truly huge.
		if (borderRatio > 0.10 && areaRatio > 0.05) || areaRatio > 0.25 {
			bgIDs[id] = true
		}
	}

	// Fallback: use the single largest border-touching component.
	if len(bgIDs) == 0 {
		bestID := -1
		bestArea := 0
		for id, comp := range components {
			if comp.TouchesBorder && comp.Area > bestArea {
				bestArea = comp.Area
				bestID = id
			}
		}
		if bestID >= 0 {
			bgIDs[bestID] = true
		}
	}

	return bgIDs
}

func buildWallMask(gray *image.NRGBA) [][]bool {
	w := gray.Bounds().Dx()
	h := gray.Bounds().Dy()
	mask := make([][]bool, h)
	for y := 0; y < h; y++ {
		mask[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			c := gray.At(x, y)
			r, _, _, _ := c.RGBA()
			v := uint8(r >> 8)
			// Typical floorplans: walls/strokes are dark on light background.
			mask[y][x] = v < 140
		}
	}
	return mask
}

func invertMask(mask [][]bool) [][]bool {
	h := len(mask)
	w := 0
	if h > 0 {
		w = len(mask[0])
	}
	out := make([][]bool, h)
	for y := 0; y < h; y++ {
		out[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			out[y][x] = !mask[y][x]
		}
	}
	return out
}

func closeWallMask(mask [][]bool, radius int) [][]bool {
	if radius <= 0 {
		return mask
	}
	return erodeMask(dilateMask(mask, radius), radius)
}

func dilateMask(mask [][]bool, radius int) [][]bool {
	h := len(mask)
	w := 0
	if h > 0 {
		w = len(mask[0])
	}
	out := make([][]bool, h)
	for y := 0; y < h; y++ {
		out[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			val := false
			for dy := -radius; dy <= radius && !val; dy++ {
				ny := y + dy
				if ny < 0 || ny >= h {
					continue
				}
				for dx := -radius; dx <= radius; dx++ {
					nx := x + dx
					if nx < 0 || nx >= w {
						continue
					}
					if mask[ny][nx] {
						val = true
						break
					}
				}
			}
			out[y][x] = val
		}
	}
	return out
}

func erodeMask(mask [][]bool, radius int) [][]bool {
	h := len(mask)
	w := 0
	if h > 0 {
		w = len(mask[0])
	}
	out := make([][]bool, h)
	for y := 0; y < h; y++ {
		out[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			val := true
			for dy := -radius; dy <= radius && val; dy++ {
				ny := y + dy
				if ny < 0 || ny >= h {
					val = false
					break
				}
				for dx := -radius; dx <= radius; dx++ {
					nx := x + dx
					if nx < 0 || nx >= w || !mask[ny][nx] {
						val = false
						break
					}
				}
			}
			out[y][x] = val
		}
	}
	return out
}

func connectedComponentsFree(free [][]bool) ([][]int, []componentInfo) {
	h := len(free)
	w := 0
	if h > 0 {
		w = len(free[0])
	}
	compMap := make([][]int, h)
	for y := 0; y < h; y++ {
		compMap[y] = make([]int, w)
		for x := 0; x < w; x++ {
			compMap[y][x] = -1
		}
	}

	components := make([]componentInfo, 0)
	dirs := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !free[y][x] || compMap[y][x] >= 0 {
				continue
			}

			id := len(components)
			comp := componentInfo{MinX: x, MinY: y, MaxX: x, MaxY: y, Area: 0, TouchesBorder: false}
			queue := []point{{X: x, Y: y}}
			compMap[y][x] = id

			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]

				comp.Area++
				if p.X < comp.MinX {
					comp.MinX = p.X
				}
				if p.Y < comp.MinY {
					comp.MinY = p.Y
				}
				if p.X > comp.MaxX {
					comp.MaxX = p.X
				}
				if p.Y > comp.MaxY {
					comp.MaxY = p.Y
				}
				if p.X == 0 || p.Y == 0 || p.X == w-1 || p.Y == h-1 {
					comp.TouchesBorder = true
				}

				for _, d := range dirs {
					nx := p.X + d[0]
					ny := p.Y + d[1]
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					if !free[ny][nx] || compMap[ny][nx] >= 0 {
						continue
					}
					compMap[ny][nx] = id
					queue = append(queue, point{X: nx, Y: ny})
				}
			}

			components = append(components, comp)
		}
	}

	return compMap, components
}

func distanceTransform(free [][]bool) [][]int {
	h := len(free)
	w := 0
	if h > 0 {
		w = len(free[0])
	}
	const inf = int(1e9)
	d := make([][]int, h)
	for y := 0; y < h; y++ {
		d[y] = make([]int, w)
		for x := 0; x < w; x++ {
			if free[y][x] {
				d[y][x] = inf
			} else {
				d[y][x] = 0
			}
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if d[y][x] == 0 {
				continue
			}
			if x > 0 {
				d[y][x] = min(d[y][x], d[y][x-1]+1)
			}
			if y > 0 {
				d[y][x] = min(d[y][x], d[y-1][x]+1)
			}
		}
	}

	for y := h - 1; y >= 0; y-- {
		for x := w - 1; x >= 0; x-- {
			if d[y][x] == 0 {
				continue
			}
			if x+1 < w {
				d[y][x] = min(d[y][x], d[y][x+1]+1)
			}
			if y+1 < h {
				d[y][x] = min(d[y][x], d[y+1][x]+1)
			}
		}
	}

	return d
}

func localMaxSeeds(distance [][]int, free [][]bool, minPeak int) []point {
	h := len(distance)
	w := 0
	if h > 0 {
		w = len(distance[0])
	}
	if minPeak < 3 {
		minPeak = 3
	}

	seeds := make([]point, 0)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			if !free[y][x] {
				continue
			}
			v := distance[y][x]
			if v < minPeak {
				continue
			}
			isMax := true
			for dy := -1; dy <= 1 && isMax; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					if distance[y+dy][x+dx] > v {
						isMax = false
						break
					}
				}
			}
			if isMax {
				seeds = append(seeds, point{X: x, Y: y})
			}
		}
	}

	if len(seeds) == 0 {
		return seeds
	}

	return dedupeNearbySeeds(seeds, minPeak)
}

func dedupeNearbySeeds(seeds []point, minDist int) []point {
	if len(seeds) == 0 {
		return nil
	}
	if minDist < 4 {
		minDist = 4
	}
	thresholdSq := minDist * minDist

	result := make([]point, 0, len(seeds))
	for _, s := range seeds {
		keep := true
		for _, r := range result {
			dx := s.X - r.X
			dy := s.Y - r.Y
			if dx*dx+dy*dy <= thresholdSq {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, s)
		}
	}
	return result
}

func mergeUntilMaxRegions(regions []regionBox, maxCount int) []regionBox {
	if maxCount <= 0 || len(regions) <= maxCount {
		return regions
	}

	merged := append([]regionBox(nil), regions...)
	for len(merged) > maxCount {
		bestI, bestJ := -1, -1
		bestGap := math.MaxInt

		for i := 0; i < len(merged); i++ {
			for j := i + 1; j < len(merged); j++ {
				gap := rectGap(merged[i], merged[j])
				if gap < bestGap {
					bestGap = gap
					bestI = i
					bestJ = j
				}
			}
		}

		if bestI < 0 || bestJ < 0 {
			break
		}

		merged[bestI] = unionRegion(merged[bestI], merged[bestJ])
		merged = append(merged[:bestJ], merged[bestJ+1:]...)
	}

	return merged
}

func unionRegion(a, b regionBox) regionBox {
	return regionBox{
		MinX: min(a.MinX, b.MinX),
		MinY: min(a.MinY, b.MinY),
		MaxX: max(a.MaxX, b.MaxX),
		MaxY: max(a.MaxY, b.MaxY),
	}
}

func rectGap(a, b regionBox) int {
	dx := 0
	if a.MaxX < b.MinX {
		dx = b.MinX - a.MaxX
	} else if b.MaxX < a.MinX {
		dx = a.MinX - b.MaxX
	}

	dy := 0
	if a.MaxY < b.MinY {
		dy = b.MinY - a.MaxY
	} else if b.MaxY < a.MinY {
		dy = a.MinY - b.MaxY
	}

	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
	}
	return dx + dy
}

// growRoomFromSeed expands a rectangle from a seed point in 4 directions
// independently.  Each edge grows outward until the cross-section along
// that edge exceeds a wall-density threshold — i.e. it hits a structural
// wall.  This produces wall-aligned rectangles instead of Voronoi shapes.
func growRoomFromSeed(seed point, wallMask [][]bool, w, h int) regionBox {
	const wallThreshold = 0.08 // stop when 8% of cross-section is wall
	maxGrow := max(w, h) / 2

	rect := regionBox{MinX: seed.X, MinY: seed.Y, MaxX: seed.X, MaxY: seed.Y}

	for i := 0; i < maxGrow; i++ {
		grew := false

		// Try expanding left.
		if rect.MinX > 0 {
			nx := rect.MinX - 1
			wallCount := 0
			total := rect.MaxY - rect.MinY + 1
			for y := rect.MinY; y <= rect.MaxY; y++ {
				if wallMask[y][nx] {
					wallCount++
				}
			}
			if total < 3 || float64(wallCount)/float64(total) < wallThreshold {
				rect.MinX = nx
				grew = true
			}
		}

		// Try expanding right.
		if rect.MaxX < w-1 {
			nx := rect.MaxX + 1
			wallCount := 0
			total := rect.MaxY - rect.MinY + 1
			for y := rect.MinY; y <= rect.MaxY; y++ {
				if wallMask[y][nx] {
					wallCount++
				}
			}
			if total < 3 || float64(wallCount)/float64(total) < wallThreshold {
				rect.MaxX = nx
				grew = true
			}
		}

		// Try expanding up.
		if rect.MinY > 0 {
			ny := rect.MinY - 1
			wallCount := 0
			total := rect.MaxX - rect.MinX + 1
			for x := rect.MinX; x <= rect.MaxX; x++ {
				if wallMask[ny][x] {
					wallCount++
				}
			}
			if total < 3 || float64(wallCount)/float64(total) < wallThreshold {
				rect.MinY = ny
				grew = true
			}
		}

		// Try expanding down.
		if rect.MaxY < h-1 {
			ny := rect.MaxY + 1
			wallCount := 0
			total := rect.MaxX - rect.MinX + 1
			for x := rect.MinX; x <= rect.MaxX; x++ {
				if wallMask[ny][x] {
					wallCount++
				}
			}
			if total < 3 || float64(wallCount)/float64(total) < wallThreshold {
				rect.MaxY = ny
				grew = true
			}
		}

		if !grew {
			break
		}
	}

	return rect
}

func inferRoomType(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "meeting"), strings.Contains(n, "conference"), strings.Contains(n, "conf"):
		return "MEETING"
	case strings.Contains(n, "hall"), strings.Contains(n, "corridor"), strings.Contains(n, "lobby"):
		return "HALLWAY"
	case strings.Contains(n, "office"):
		return "OFFICE"
	default:
		return "UNKNOWN"
	}
}

func to1000(v int, total int) int {
	if total <= 0 {
		return 0
	}
	out := int(float64(v) / float64(total) * 1000.0)
	if out < 0 {
		return 0
	}
	if out > 1000 {
		return 1000
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resolveTesseractCmd() string {
	if custom := strings.TrimSpace(os.Getenv("TESSERACT_CMD")); custom != "" {
		if _, err := exec.LookPath(custom); err == nil {
			return custom
		}
	}
	if _, err := exec.LookPath("tesseract"); err == nil {
		return "tesseract"
	}
	return ""
}

func detectRoomText(img image.Image, rect regionBox, tesseractCmd string) string {
	if tesseractCmd == "" {
		return ""
	}

	bounds := img.Bounds()
	padX := max(2, (rect.MaxX-rect.MinX)/20)
	padY := max(2, (rect.MaxY-rect.MinY)/20)
	x0 := max(bounds.Min.X, rect.MinX+bounds.Min.X+padX)
	y0 := max(bounds.Min.Y, rect.MinY+bounds.Min.Y+padY)
	x1 := min(bounds.Max.X, rect.MaxX+bounds.Min.X-padX)
	y1 := min(bounds.Max.Y, rect.MaxY+bounds.Min.Y-padY)
	if x1-x0 < 10 || y1-y0 < 10 {
		return ""
	}

	crop := imaging.Crop(img, image.Rect(x0, y0, x1, y1))
	gray := imaging.Grayscale(crop)
	gray = imaging.AdjustContrast(gray, 20)

	tmp, err := os.CreateTemp("", "room-ocr-*.png")
	if err != nil {
		return ""
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	f, err := os.Create(tmpPath)
	if err != nil {
		return ""
	}
	if err := png.Encode(f, gray); err != nil {
		_ = f.Close()
		return ""
	}
	_ = f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, tesseractCmd, tmpPath, "stdout", "--psm", "6", "-l", "eng")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return cleanOCRText(string(out))
}

func cleanOCRText(raw string) string {
	line := strings.TrimSpace(raw)
	if line == "" {
		return ""
	}
	line = strings.ReplaceAll(line, "\n", " ")
	line = strings.ReplaceAll(line, "\r", " ")
	line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
	line = regexp.MustCompile(`[^a-zA-Z0-9\-\s]`).ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	if len(line) < 2 {
		return ""
	}
	if len(line) > 50 {
		line = line[:50]
	}
	return line
}

func suppressOverlapsAndSort(candidates []detectedCandidate) []detectedCandidate {
	if len(candidates) < 2 {
		return candidates
	}

	sort.Slice(candidates, func(i, j int) bool {
		areaI := rectArea(candidates[i].Rect)
		areaJ := rectArea(candidates[j].Rect)
		if areaI == areaJ {
			if candidates[i].Rect.MinY == candidates[j].Rect.MinY {
				return candidates[i].Rect.MinX < candidates[j].Rect.MinX
			}
			return candidates[i].Rect.MinY < candidates[j].Rect.MinY
		}
		return areaI > areaJ
	})

	kept := make([]detectedCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		drop := false
		for _, existing := range kept {
			iou := rectIoU(candidate.Rect, existing.Rect)
			overSmaller := overlapOverSmaller(candidate.Rect, existing.Rect)
			// Seeds in the same room grow to similar rects — merge at IoU ≥ 0.50
			// or if the smaller is mostly contained (≥ 0.70).
			if iou >= 0.50 || overSmaller >= 0.70 {
				drop = true
				break
			}
		}
		if !drop {
			kept = append(kept, candidate)
		}
	}

	sort.Slice(kept, func(i, j int) bool {
		if kept[i].Rect.MinY == kept[j].Rect.MinY {
			return kept[i].Rect.MinX < kept[j].Rect.MinX
		}
		return kept[i].Rect.MinY < kept[j].Rect.MinY
	})

	return kept
}

func rectArea(rect regionBox) int {
	w := max(0, rect.MaxX-rect.MinX+1)
	h := max(0, rect.MaxY-rect.MinY+1)
	return w * h
}

func rectIntersectionArea(a, b regionBox) int {
	x0 := max(a.MinX, b.MinX)
	y0 := max(a.MinY, b.MinY)
	x1 := min(a.MaxX, b.MaxX)
	y1 := min(a.MaxY, b.MaxY)
	if x1 < x0 || y1 < y0 {
		return 0
	}
	return (x1 - x0 + 1) * (y1 - y0 + 1)
}

func rectIoU(a, b regionBox) float64 {
	inter := rectIntersectionArea(a, b)
	if inter == 0 {
		return 0
	}
	union := rectArea(a) + rectArea(b) - inter
	if union <= 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func overlapOverSmaller(a, b regionBox) float64 {
	inter := rectIntersectionArea(a, b)
	if inter == 0 {
		return 0
	}
	den := min(rectArea(a), rectArea(b))
	if den <= 0 {
		return 0
	}
	return float64(inter) / float64(den)
}
