# Research: PDF to Digital Twin

## Unknowns & Decisions

### 1. PDF Rasterization Strategy
**Context**: We need to convert a PDF page to an image to (1) send to Gemini for analysis and (2) crop and display to the user.
**Problem**: The current `unidoc` v2 library does not support full page rendering. System tools like `pdftoppm` are not available in the current Windows environment.
**Options**:
- **Option A**: `gen2brain/go-fitz` (Wraps MuPDF). High quality, fast. Requires CGO and GCC/MinGW on Windows.
- **Option B**: Extract embedded images. If the PDF is a scan, this works. If it's a vector CAD drawing, this fails.
- **Option C**: Use an external API (e.g., Google Cloud Document AI) just for rendering? Overkill/Cost.
- **Option D**: Use `mattn/go-sixel` or similar hacks? No.
- **Decision**: **Try Option A (go-fitz)** if environment allows. **Fallback to Option B (Image Extraction)** if CGO is an issue.
- **Implementation Note**: The current `pdf.go` returns a white image. We will attempt to implement `go-fitz` or a similar CGO binding. If that fails during implementation, we might need to ask the user to install `poppler` (pdftoppm).

### 2. Gemini Prompting
**Context**: We need `content_box` and `rooms` with coordinates.
**Findings**:
- `backend/ai/gemini.go` already implements a prompt requesting `content_box` and `rooms` in 0-1000 relative coordinates.
- **Validation**: 0-1000 is a standard Gemini output format.
- **Decision**: Stick with the current prompt structure but refine it to ensure it strictly ignores "legends" and "title blocks".

### 3. Coordinate Mapping
**Context**: We crop the image to `content_box`. We need to shift `room` coordinates.
**Math**:
- `ScaleX = ImageWidth / 1000`
- `CropX = ContentBox.XMin * ScaleX`
- `RoomX_New = (Room.XMin * ScaleX) - CropX`
- **Decision**: The logic in `backend/handler/upload.go` implements this correctly. We just need to ensure the `image.Image` used for cropping matches the dimensions used for scaling.

### 4. Real-time Updates (WebSocket)
**Context**: Updates need to be pushed to frontend.
**Status**: `backend/realtime/hub.go` exists (implied). `frontend/services/websocket.js` exists (implied).
**Decision**: Use standard Gorilla WebSocket hub pattern.
- **Event format**: `{ "type": "status_update", "roomId": "...", "status": "busy" }`

### 5. Frontend Visualization (Konva)
**Context**: "Spotlight" effect.
**Findings**: `FloorplanMap.vue` already implements `destination-out` composite operation to create "holes" in a dark overlay.
**Decision**: Proceed with this implementation.

## Tech Stack Confirmation
- **Backend**: Go + Gin + Gorilla WebSocket + Google Vertex AI
- **PDF**: `go-fitz` (Tentative) or `unidoc` (for metadata)
- **Frontend**: Vue 3 + Vite + Konva
- **State**: Vue Reactivity (ref/reactive)

## Risk Assessment
- **PDF Rendering**: High risk on Windows without pre-installed tools.
- **Gemini Accuracy**: Spatial accuracy can be hit-or-miss. We accept "good enough" for MVP (95% target).
