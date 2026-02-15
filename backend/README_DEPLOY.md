# GCP Backend Deployment

## Prerequisites

1.  **Google Cloud SDK**: Ensure `gcloud` CLI is installed and initialized.
    - Run `gcloud init` to set up your account and project.
2.  **Docker** (Optional for local testing, not required for Cloud Build).
3.  **Permissions**: You need permissions to use Cloud Build and Cloud Run.
4.  **APIs Enabled**: Ensure the following Google Cloud APIs are enabled:
    -   Cloud Build API
    -   Cloud Run Admin API
    -   Vertex AI API
    -   Artifact Registry API (if using Artifact Registry instead of Container Registry)

## Configuration

Edit `deploy.ps1` (Windows) or `deploy.sh` (Linux/Mac) to set your GCP Project ID and Region.

Default values:
- Project ID: `floorplan-digital-twin`
- Region: `asia-southeast3`
- Service Name: `floorplan-backend`

## Deployment Steps

### Windows (PowerShell)

1.  Open PowerShell in the `backend` directory.
2.  Run the deployment script:
    ```powershell
    .\deploy.ps1
    ```

### Linux / Mac (Bash)

1.  Open terminal in the `backend` directory.
2.  Make the script executable:
    ```bash
    chmod +x deploy.sh
    ```
3.  Run the script:
    ```bash
    ./deploy.sh
    ```

## Environment Variables

The deployment script sets the following environment variables on the Cloud Run service:

- `GCP_PROJECT_ID`: The Google Cloud Project ID.
- `GCP_LOCATION`: The region for Vertex AI calls.
- `PORT`: (Automatically set by Cloud Run) Port to listen on.

## Verification

After deployment, the script will output the Service URL.
Visit `https://<SERVICE_URL>/` to verify the backend is running.
Visit `https://<SERVICE_URL>/swagger/index.html` for API documentation.
