# Tasks: Automated PDF to Digital Twin Pipeline

**Branch**: `001-pdf-floorplan-conversion`
**Spec**: [spec.md](spec.md)
**Plan**: [plan.md](plan.md)

## Phase 1: Setup & Configuration

- [X] T001 Verify backend environment and Google Cloud credentials in backend/
- [X] T002 Verify frontend environment and dependencies in frontend/
- [X] T003 Create backend/ai/render.go interface for PDF rasterization strategy
- [ ] T004 Create backend/models package for shared data structures (Room, Floorplan)

## Phase 2: Foundational Components (Blocking)

- [ ] T005 Implement go-fitz or fallback PDF rendering logic in backend/ai/pdf.go
- [ ] T006 Update backend/ai/gemini.go prompt to strictly exclude legends/margins
- [ ] T007 Implement strict coordinate mapping and cropping logic unit tests in backend/handler/upload_test.go
- [ ] T008 Refactor backend/handler/upload.go to use new models and render packages

## Phase 3: User Story 1 - Upload and Process Floorplan

**Goal**: Users can upload a PDF and see a clean, cropped image with rooms detected.

- [ ] T009 [US1] Update backend/handler/upload.go to handle multipart uploads and trigger pipeline
- [ ] T010 [P] [US1] Implement frontend Upload.vue component with drag-and-drop
- [ ] T011 [US1] Connect frontend upload to backend API in frontend/src/views/Home.vue
- [ ] T012 [US1] Display raw processing results (JSON) in frontend for debugging (temporary)
- [ ] T013 [US1] Implement "Loading" state in frontend while processing

## Phase 4: User Story 2 - View Interactive Map

**Goal**: Users see the spotlight effect and interactive rooms.

- [ ] T014 [US2] Update FloorplanMap.vue to accept image and rooms props from API response
- [ ] T015 [P] [US2] Implement Konva "spotlight" layer (dimmed background with clear holes)
- [ ] T016 [P] [US2] Add hover interactions (tooltip/highlight) for rooms in FloorplanMap.vue
- [ ] T017 [US2] Integrate FloorplanMap.vue into Home.vue with real data

## Phase 5: User Story 3 - Real-Time Status Monitoring

**Goal**: Live updates via WebSocket.

- [ ] T018 [US3] Implement WebSocket Hub in backend/realtime/hub.go to manage clients
- [ ] T019 [US3] Add subscribe message handling in backend
- [ ] T020 [US3] Create endpoint/method to trigger status updates (simulated for now) in backend/main.go
- [ ] T021 [P] [US3] Implement WebSocket client service in frontend/services/websocket.js
- [ ] T022 [US3] Wire up WebSocket events to update rooms state in Home.vue

## Phase 6: Polish & Cross-Cutting

- [ ] T023 Add error handling for invalid PDFs or failed AI processing
- [ ] T024 Add "Retry" mechanism in frontend
- [ ] T025 Finalize UI styling (Tailwind/CSS)

## Dependencies

- Phase 2 (Foundational) blocks all Story phases.
- Story 1 (Upload) blocks Story 2 (Map View).
- Story 2 (Map View) blocks Story 3 (Real-time) - *Visuals needed before status updates*.

## Parallel Execution

- T010 (Frontend Upload) can run parallel to T009 (Backend Handler).
- T015/T016 (Konva Visuals) can run parallel to T014 (Data wiring).
- T021 (Frontend WS) can run parallel to T018 (Backend WS).

## Implementation Strategy

1. **MVP (Story 1 & 2)**: Get the "Magic" working first. Upload -> Crop -> Interactive Map.
2. **Real-time (Story 3)**: Add the "Live" layer on top.
