# Space Twin

Space Twin converts floorplan inputs into interactive Digital Twin experiences using a Vue frontend and a Go backend.

Space Twin is built for teams that need to move from static plans to operational insight faster. Instead of spending days manually interpreting layouts, stakeholders can turn floorplan documents into navigable digital spaces that support planning, communication, and execution.

Live demo: https://space-twin.vercel.app/

![Space Twin Screenshot](./frontend/public/space-twin-screenshot.png)

## Business Value

- Accelerate project delivery by reducing manual floorplan interpretation time.
- Improve decision quality with a shared, visual source of truth across teams.
- Increase operational readiness through faster space understanding and walkthrough.
- Scale portfolio analysis by standardizing how floorplans become digital assets.

## Ideal Use Cases

- Real estate and proptech teams building property intelligence workflows.
- Architecture, engineering, and construction teams coordinating design intent.
- Facilities and operations teams planning maintenance, occupancy, and navigation.
- Smart building initiatives that require structured geometry for downstream systems.

## Why Space Twin

- Converts 2D plan inputs into digital twin-ready structure.
- Supports fast iteration from upload to interactive visualization.
- Combines AI-assisted extraction with real-time frontend interaction.
- Provides a foundation for analytics, simulation, and automation extensions.

## Repository Structure

- `frontend/` - Vue 3 + Vite app (UI, upload flow, visualization, generated API client)
- `backend/` - Go API, AI processing pipeline, WebSocket hub, Swagger docs
- `specs/` - product spec, plan, tasks, and API contract references

## Quick Start

### 1) Start backend

```sh
cd backend
go mod download
go run main.go
```

Backend runs on `http://localhost:8080`.

### 2) Start frontend

```sh
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173`.

## Environment Variables (Frontend)

Create `frontend/.env` if needed:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080/ws
```

## Main Backend Endpoints

- `POST /api/v1/upload`
- `POST /api/v1/process/edges`
- `POST /api/v1/process/edges-json`
- `POST /api/v1/process/crop`
- `GET /ws`

## Detailed Docs

- Frontend documentation: [frontend/README.md](./frontend/README.md)
- API Swagger UI (local): `http://localhost:8080/swagger/index.html`
- API contract source: [specs/001-pdf-floorplan-conversion/contracts/api.yaml](./specs/001-pdf-floorplan-conversion/contracts/api.yaml)
