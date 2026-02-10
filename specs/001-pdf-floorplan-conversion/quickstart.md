# Quickstart: Automated PDF to Digital Twin

## Prerequisites
- **Go**: 1.25.7+
- **Node.js**: 20+
- **Google Cloud**: Valid `google-credentials.json` with Vertex AI enabled.
- **Git**: Installed.

## Environment Setup
1. **Google Cloud Credentials**:
   - Place your service account JSON key at `backend/google-credentials.json`.
   - Set env var: `export GOOGLE_APPLICATION_CREDENTIALS="./google-credentials.json"` (or Windows equivalent).

2. **Backend**:
   ```bash
   cd backend
   go mod download
   go run main.go
   ```
   - Server starts on `http://localhost:8080`.

3. **Frontend**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   - App opens on `http://localhost:5173`.

## Usage
1. Open frontend in browser.
2. Drag & Drop a PDF floorplan onto the upload area.
3. Wait for processing (PDF -> Image -> Gemini -> Crop -> Display).
4. See interactive map.
5. Hover over rooms to see status.

## Troubleshooting
- **"Failed to create genai client"**: Ensure `GOOGLE_APPLICATION_CREDENTIALS` is set correctly.
- **"PDF not supported"**: If running on Windows without `go-fitz` build constraints, PDF rendering might fail. Try converting PDF to PNG manually first.
