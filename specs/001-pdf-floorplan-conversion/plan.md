# Implementation Plan: Automated PDF to Digital Twin Pipeline

**Branch**: `001-pdf-floorplan-conversion` | **Date**: February 10, 2026 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-pdf-floorplan-conversion/spec.md`

## Summary

Implement an automated pipeline that uploads a PDF floorplan, converts it to an image (resolving rasterization challenges), sends it to Google Gemini to identify the content box and rooms, crops the image to remove margins, and displays it on an interactive map using Vue and Konva. Real-time status updates are pushed via WebSockets.

## Technical Context

**Language/Version**: Go 1.25.7, JavaScript (Vue 3.5.27)
**Primary Dependencies**: Gin (Web), Vertex AI (LLM), Imaging (Image Proc), Gorilla WebSocket (Realtime), Konva (Canvas UI), Unidoc (PDF Metadata)
**Storage**: In-memory (MVP) or File System for images. No DB required yet.
**Testing**: Go `testing` package, standard Vue testing.
**Target Platform**: Web (Linux/Windows server backend, Browser frontend)
**Project Type**: Web Application (Monorepo: backend + frontend)
**Performance Goals**: <5s processing time (excluding LLM latency), <100ms WebSocket latency.
**Constraints**: PDF rasterization must work in the deployment environment (Windows dev, Linux prod).
**Scale/Scope**: MVP focuses on single-user upload flow and multi-user viewing.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Library-First**: Core logic (AI analysis, Image processing) is separated in `backend/ai` and `backend/handler` packages.
- **CLI Interface**: Not strictly applicable for web endpoints but `main.go` accepts flags/env vars.
- **Test-First**: Unit tests required for coordinate math and image cropping logic.
- **Integration Testing**: End-to-end test of the upload -> analyze -> response flow.
- **Observability**: Structured logging for AI interactions.

**Status**: PASSED

## Project Structure

### Documentation (this feature)

```text
specs/001-pdf-floorplan-conversion/
 plan.md              # This file
 research.md          # Strategy for PDF rasterization and AI prompting
 data-model.md        # Entities: Floorplan, Room, ContentBox
 quickstart.md        # Setup guide
 contracts/           # API definitions
    api.yaml
 tasks.md             # To be generated
```

### Source Code (repository root)

```text
backend/
 go.mod
 main.go
 ai/
    gemini.go       # Vertex AI client and prompting
    pdf.go          # PDF processing (Rasterization logic)
 handler/
    upload.go       # HTTP handler for file upload & coordination
    debug.go
 realtime/
     hub.go          # WebSocket hub

frontend/
 package.json
 vite.config.js
 src/
     App.vue
     services/
        websocket.js
     components/
         FloorplanMap.vue # Konva map visualization
```

**Structure Decision**: Monorepo with separated backend (Go) and frontend (Vue).

## Complexity Tracking

N/A - Standard Architecture.
