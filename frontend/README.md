# Space Twin Frontend

Space Twin converts 2D floorplan inputs into interactive Digital Twin experiences.

Live app: https://space-twin.vercel.app/

![Space Twin Screenshot](./public/space-twin-screenshot.png)

## What this app does

- Upload and process floorplan files through an AI-assisted pipeline.
- Transform extracted structure into renderable digital twin data.
- Visualize and interact with processed outputs in real time.
- Connect to backend APIs and WebSocket updates for processing status.

## Tech stack

- Vue 3 + Vite
- Vue Router
- TanStack Query
- Konva / Vue-Konva
- pdfjs-dist
- OpenAPI client generation (`@hey-api/openapi-ts`)

## Prerequisites

- Node.js 20+
- npm
- Backend API running at `http://localhost:8080` (or your configured URL)

## Local development

```sh
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173` by default.

## Environment variables

Create a `.env` file in `frontend/` if you need custom endpoints:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080/ws
```

Defaults are used automatically if variables are not set.

## Scripts

- `npm run dev` - start Vite dev server
- `npm run build` - production build
- `npm run preview` - preview the production build
- `npm run generate` - regenerate API client from OpenAPI

## Backend quick start (required)

```sh
cd backend
go mod download
go run main.go
```

Main backend routes used by frontend:

- `POST /api/v1/upload`
- `POST /api/v1/process/edges`
- `POST /api/v1/process/edges-json`
- `POST /api/v1/process/crop`
- `GET /ws`

## Project structure (frontend)

- `src/views` - page-level views
- `src/components` - reusable UI components
- `src/services/websocket.js` - realtime channel integration
- `src/client` - generated API client code
- `src/utils/env.js` - endpoint configuration

## Notes

- This README intentionally uses high-level PDF-to-digital-twin wording for product positioning.
- For deployed experience and feature walkthrough, use the live app: https://space-twin.vercel.app/
