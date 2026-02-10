# Data Model: PDF to Digital Twin

## 1. Entities

### Floorplan
Represents the processed digital twin of an uploaded document.
- **Attributes**:
  - `id` (UUID): Unique identifier.
  - `filename` (String): Original uploaded file name.
  - `image_url` (String): URL/Data URI to the cropped image.
  - `width` (Int): Width of the cropped image (pixels).
  - `height` (Int): Height of the cropped image (pixels).
  - `created_at` (DateTime): Upload timestamp.

### Room
A distinct functional space within a floorplan.
- **Attributes**:
  - `id` (UUID): Unique identifier (or name-based ID).
  - `floorplan_id` (UUID): Reference to parent floorplan.
  - `name` (String): Display name (e.g., "Meeting Room A").
  - `type` (Enum): `OFFICE`, `MEETING`, `HALLWAY`, `UNKNOWN`.
  - `rect` (Struct): Coordinate structure `[x, y, w, h]` relative to cropped image.
  - `status` (Enum): `AVAILABLE`, `BUSY`, `OFFLINE`.

### ContentBox (Internal)
Used during processing to crop the original image.
- **Attributes**:
  - `original_width` (Int): Width of source image.
  - `original_height` (Int): Height of source image.
  - `bounds` (List<Int>): `[ymin, xmin, ymax, xmax]` (0-1000 relative).

## 2. Relationships

- **Floorplan** `1:N` **Room**
  - A floorplan contains many rooms.
  - A room belongs to exactly one floorplan.

## 3. WebSocket Messages

### Client -> Server
- `subscribe`: `{ "type": "subscribe", "floorplan_id": "..." }`

### Server -> Client
- `status_update`: `{ "type": "status_update", "room_id": "...", "status": "BUSY", "timestamp": "..." }`
